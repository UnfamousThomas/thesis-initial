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
	"github.com/go-logr/logr"
	"github.com/unfamousthomas/thesis-operator/internal/scaling"
	"github.com/unfamousthomas/thesis-operator/internal/utils"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	networkv1alpha1 "github.com/unfamousthomas/thesis-operator/api/v1alpha1"
)

const FLEET_FINALIZER = "fleets.unfamousthomas.me/finalizer"

// FleetReconciler reconciles a Fleet object
type FleetReconciler struct {
	client.Client
	Scheme *runtime.Scheme
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
	logger := log.FromContext(ctx).WithValues("fleet", req.Name, "namespace", req.Namespace)

	//Fetch the fleet resource from the cluster
	fleet := &networkv1alpha1.Fleet{}
	if err := r.Get(ctx, req.NamespacedName, fleet); err != nil {
		if client.IgnoreNotFound(err) != nil {
			logger.Error(err, "Failed to get Fleet resource")
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Handle finalizer addition
	if fleet.DeletionTimestamp == nil && !controllerutil.ContainsFinalizer(fleet, FLEET_FINALIZER) {
		logger.Info("Adding finalizer to fleet")
		controllerutil.AddFinalizer(fleet, FLEET_FINALIZER)
		if err := r.Update(ctx, fleet); err != nil {
			logger.Error(err, "Failed to add finalizer to fleet")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	// Handle resource deletion
	if fleet.DeletionTimestamp != nil || !fleet.GetDeletionTimestamp().IsZero() {
		logger.Info("Handling deletion of fleet")
		if err := r.handleDeletion(ctx, fleet, logger); err != nil {
			logger.Error(err, "Failed to handle fleet deletion")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil // Return after so we do not accidentally scale again
	}

	servers, err := r.getServers(ctx, fleet, logger)
	if err != nil {
		return ctrl.Result{}, err
	}
	fleet.Status.CurrentReplicas = int32(len(servers.Items))
	if fleet.Spec.Scaling.Replicas != fleet.Status.CurrentReplicas {
		if err := r.scaleServerCount(ctx, fleet, req.Namespace, logger); err != nil {
			return ctrl.Result{}, err
		}
		servers, err := r.getServers(ctx, fleet, logger)
		if err != nil {
			return ctrl.Result{}, err
		}
		fleet.Status.CurrentReplicas = int32(len(servers.Items))
	}

	if err := r.Status().Update(ctx, fleet); err != nil {
		logger.Error(err, "Failed to update Fleet resource")
		return ctrl.Result{}, err
	}
	logger.Info("Reconciliation finished")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *FleetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&networkv1alpha1.Fleet{}).
		Owns(&networkv1alpha1.Server{}).
		WithOptions(controller.Options{MaxConcurrentReconciles: 10}).
		Complete(r)
}

func (r *FleetReconciler) scaleServerCount(ctx context.Context, fleet *networkv1alpha1.Fleet, namespace string, logger logr.Logger) error {
	if fleet.Status.CurrentReplicas < fleet.Spec.Scaling.Replicas {
		//Scale up
		serversNeeded := fleet.Spec.Scaling.Replicas - fleet.Status.CurrentReplicas
		logger.Info(fmt.Sprintf("Scaling servers needed: %d", serversNeeded))
		for range serversNeeded {
			server := utils.CreateServerForFleet(*fleet, namespace)
			err := r.Create(ctx, server)
			if err != nil {
				return err
			}
		}
	}
	//Scale down
	if fleet.Status.CurrentReplicas > fleet.Spec.Scaling.Replicas {
		serversToDelete := fleet.Status.CurrentReplicas - fleet.Spec.Scaling.Replicas
		logger.Info(fmt.Sprintf("Deleting servers needed: %d", serversToDelete))
		servers, err := r.getServers(ctx, fleet, logger)
		if err != nil {
			return err
		}
		server, err := scaling.FindDeleteServer(ctx, fleet, servers, r.Client, logger)
		if err != nil {
			return err
		}
		if err := r.Client.Delete(ctx, server); err != nil {
			return err
		}
	}
	return nil
}

func (r *FleetReconciler) getServers(ctx context.Context, fleet *networkv1alpha1.Fleet, logger logr.Logger) (*networkv1alpha1.ServerList, error) {
	serverList := &networkv1alpha1.ServerList{}
	labelSelector := client.MatchingLabels{"fleet": fleet.Name}
	if err := r.List(ctx, serverList, client.InNamespace(fleet.Namespace), labelSelector); err != nil {
		return nil, err
	}
	return serverList, nil
}

func (r *FleetReconciler) handleDeletion(ctx context.Context, fleet *networkv1alpha1.Fleet, logger logr.Logger) error {
	servers, err := r.getServers(ctx, fleet, logger)
	if err != nil {
		return err
	}
	for _, server := range servers.Items {
		if err := r.Delete(ctx, &server); err != nil {
			return err
		}
	}

	servers, err = r.getServers(ctx, fleet, logger)
	if err != nil {
		return err
	}
	if len(servers.Items) == 0 {
		controllerutil.RemoveFinalizer(fleet, FLEET_FINALIZER)
		if err := r.Update(ctx, fleet); err != nil {
			return err
		}
	}
	return nil
}
