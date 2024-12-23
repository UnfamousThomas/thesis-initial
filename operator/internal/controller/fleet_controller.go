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
	"github.com/unfamousthomas/thesis-operator/internal/utils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
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
		if err := r.handleDeletion(ctx, fleet, logger); err != nil { //todo
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
	if fleet.Spec.Replicas != fleet.Status.CurrentReplicas {
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
	if fleet.Status.CurrentReplicas < fleet.Spec.Replicas {
		//Scale up
		serversNeeded := fleet.Spec.Replicas - fleet.Status.CurrentReplicas
		logger.Info(fmt.Sprintf("Scaling servers needed: %d", serversNeeded))
		for range serversNeeded {
			server := utils.CreateServerForFleet(*fleet, namespace)
			err := r.Create(ctx, server)
			if err != nil {
				return err
			}
		}
	}
	if fleet.Status.CurrentReplicas > fleet.Spec.Replicas {
		serversToDelete := fleet.Status.CurrentReplicas - fleet.Spec.Replicas
		logger.Info(fmt.Sprintf("Deleting servers needed: %d", serversToDelete))
		servers, err := r.getServers(ctx, fleet, logger)
		if err != nil {
			return err
		}
		if err := r.deleteOneServer(ctx, fleet, servers, logger); err != nil {
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

func (r *FleetReconciler) deleteOneServer(ctx context.Context, fleet *networkv1alpha1.Fleet, servers *networkv1alpha1.ServerList, logger logr.Logger) error {
	selectedServer := &networkv1alpha1.Server{}
	var oldestServer *networkv1alpha1.Server

	for _, server := range servers.Items {
		podName := server.Name + "-pod"
		pod := &v1.Pod{}
		err := r.Client.Get(ctx, types.NamespacedName{Namespace: server.Namespace, Name: podName}, pod)
		if err != nil {
			logger.Error(err, "Failed to fetch Pod for server", "server", server.Name)
			continue
		}

		// Check if deletion is allowed for this server's pod
		allowed, err := utils.IsDeleteAllowed(pod)
		if err != nil {
			logger.Error(err, "Failed to determine if deletion is allowed", "pod", podName)
			continue
		}

		if allowed {
			// Find the oldest server among those allowed to be deleted
			if oldestServer == nil || server.CreationTimestamp.Before(&oldestServer.CreationTimestamp) {
				oldestServer = &server
			}
		}
	}

	// Default to the oldest server if no eligible server is found
	if oldestServer == nil && len(servers.Items) > 0 {
		logger.Info("No server eligible for deletion, defaulting to the oldest server")
		oldestServer = &servers.Items[0]
		for _, server := range servers.Items {
			if server.CreationTimestamp.Before(&oldestServer.CreationTimestamp) && server.GetDeletionTimestamp().IsZero() {
				oldestServer = &server
			}
		}
	}

	if oldestServer == nil {
		logger.Info("No servers available for deletion")
		return nil
	}

	selectedServer = oldestServer
	logger.Info("Deleting server", "server", selectedServer.Name)

	// Delete the selected server
	if err := r.Client.Delete(ctx, selectedServer); err != nil {
		logger.Error(err, "Failed to delete server", "server", selectedServer.Name)
		return err
	}

	logger.Info("Server deleted successfully", "server", selectedServer.Name)
	return nil
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
