/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	networkv1alpha1 "github.com/unfamousthomas/thesis-operator/api/v1alpha1"
	"github.com/unfamousthomas/thesis-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const FLEET_FINALIZER = "fleets.unfamousthomas.me/finalizer"

// FleetReconciler reconciles a Fleet object
type FleetReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=network.unfamousthomas.me,resources=fleets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=network.unfamousthomas.me,resources=fleets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=network.unfamousthomas.me,resources=fleets/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.18.4/pkg/reconcile
func (r *FleetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	//Fetch the fleet resource from the cluster
	fleet := &networkv1alpha1.Fleet{}
	if err := r.Get(ctx, req.NamespacedName, fleet); err != nil {
		return ctrl.Result{}, err
	}

	// Handle finalizer addition
	if fleet.DeletionTimestamp == nil && !controllerutil.ContainsFinalizer(fleet, FLEET_FINALIZER) {
		controllerutil.AddFinalizer(fleet, FLEET_FINALIZER)
		if err := r.Update(ctx, fleet); err != nil {
			r.emitEventf(fleet, corev1.EventTypeNormal, utils.ReasonFleetUpdateFailed, "Fleet finalizer update failed: %s", err)
			return ctrl.Result{Requeue: true}, fmt.Errorf("failed to add finalizer to fleet: %w", err)
		}
		r.emitEvent(fleet, corev1.EventTypeNormal, utils.ReasonFleetInitialized, "Fleet finalizers added")
		return ctrl.Result{Requeue: true}, nil
	}

	// Handle resource deletion
	if fleet.DeletionTimestamp != nil || !fleet.GetDeletionTimestamp().IsZero() {
		if err := r.handleDeletion(ctx, fleet); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to handle fleet deletion: %w", err)
		}
		return ctrl.Result{}, nil // Return after so we do not accidentally scale again
	}

	servers, err := r.getServers(ctx, fleet)
	if err != nil {
		return ctrl.Result{Requeue: true}, err
	}
	fleet.Status.CurrentReplicas = int32(len(servers.Items))
	if fleet.Spec.Scaling.Replicas != fleet.Status.CurrentReplicas {
		if err := r.scaleServerCount(ctx, fleet, req.Namespace); err != nil {
			return ctrl.Result{}, err
		}
		servers, err := r.getServers(ctx, fleet)
		if err != nil {
			return ctrl.Result{Requeue: true}, err
		}
		fleet.Status.CurrentReplicas = int32(len(servers.Items))
	}

	if err := r.Status().Update(ctx, fleet); err != nil {
		return ctrl.Result{Requeue: true}, fmt.Errorf("failed to update Fleet status resource: %w", err)
	}
	return ctrl.Result{Requeue: true}, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *FleetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&networkv1alpha1.Fleet{}).
		Owns(&networkv1alpha1.Server{}).
		WithOptions(controller.Options{MaxConcurrentReconciles: 10}).
		Complete(r)
}

func (r *FleetReconciler) scaleServerCount(ctx context.Context, fleet *networkv1alpha1.Fleet, namespace string) error {
	if fleet.Status.CurrentReplicas < fleet.Spec.Scaling.Replicas {
		//Scale up
		serversNeeded := fleet.Spec.Scaling.Replicas - fleet.Status.CurrentReplicas
		for range serversNeeded {
			server := utils.CreateServerForFleet(*fleet, namespace)
			err := r.Create(ctx, server)
			if err != nil {
				return err
			}
		}
		r.emitEventf(fleet, corev1.EventTypeNormal, utils.ReasonFleetScaleServers, "Scaled servers up to %d", fleet.Spec.Scaling.Replicas)
	}
	//Scale down
	if fleet.Status.CurrentReplicas > fleet.Spec.Scaling.Replicas {
		servers, err := r.getServers(ctx, fleet)
		if err != nil {
			return err
		}
		server, err := utils.FindDeleteServer(ctx, fleet, servers, r.Client)
		if err != nil {
			return err
		}
		if err := r.Client.Delete(ctx, server); err != nil {
			return err
		}
		r.emitEventf(fleet, corev1.EventTypeNormal, utils.ReasonFleetScaleServers, "Scaled servers down to %d", fleet.Spec.Scaling.Replicas)
	}
	return nil
}

func (r *FleetReconciler) getServers(ctx context.Context, fleet *networkv1alpha1.Fleet) (*networkv1alpha1.ServerList, error) {
	serverList := &networkv1alpha1.ServerList{}
	labelSelector := client.MatchingLabels{"fleet": fleet.Name}
	if err := r.List(ctx, serverList, client.InNamespace(fleet.Namespace), labelSelector); err != nil {
		return nil, err
	}
	return serverList, nil
}

func (r *FleetReconciler) handleDeletion(ctx context.Context, fleet *networkv1alpha1.Fleet) error {
	servers, err := r.getServers(ctx, fleet)
	if err != nil {
		return err
	}
	for _, server := range servers.Items {
		if err := r.Delete(ctx, &server); err != nil {
			return err
		}
	}

	servers, err = r.getServers(ctx, fleet)
	if err != nil {
		return err
	}
	if len(servers.Items) == 0 {
		controllerutil.RemoveFinalizer(fleet, FLEET_FINALIZER)
		if err := r.Update(ctx, fleet); err != nil {
			r.emitEventf(fleet, corev1.EventTypeWarning, utils.ReasonFleetUpdateFailed, "Failed to remvoe finalizer: %s", err)
			return err
		}
		r.emitEvent(fleet, corev1.EventTypeNormal, utils.ReasonFleetServersRemoved, "Fleet finalizers removed correctly")
	}
	return nil
}

func (r *FleetReconciler) emitEvent(object runtime.Object, eventtype string, reason utils.EventReason, message string) {
	r.Recorder.Event(object, eventtype, string(reason), message)
}

func (r *FleetReconciler) emitEventf(object runtime.Object, eventtype string, reason utils.EventReason, message string, args ...interface{}) {
	r.Recorder.Eventf(object, eventtype, string(reason), message, args)
}
