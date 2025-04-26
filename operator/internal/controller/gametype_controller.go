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
	"github.com/go-logr/logr"
	"github.com/unfamousthomas/thesis-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	networkv1alpha1 "github.com/unfamousthomas/thesis-operator/api/v1alpha1"
)

// GameTypeReconciler reconciles a GameType object
type GameTypeReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

const TypeFinalizer = "gametype.unfamousthomas.me/finalizer"

// +kubebuilder:rbac:groups=network.unfamousthomas.me,resources=gametypes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=network.unfamousthomas.me,resources=gametypes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=network.unfamousthomas.me,resources=gametypes/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// the GameType object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.18.4/pkg/reconcile
func (r *GameTypeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithValues("gametype", req.Name, "namespace", req.Namespace)

	logger.Info("Reconciling GameType")
	gametype := &networkv1alpha1.GameType{}
	if err := r.Get(ctx, req.NamespacedName, gametype); err != nil {
		logger.Error(err, "Failed to get gametype resource")
		return ctrl.Result{}, err
	}

	// Handle finalizer addition
	if gametype.DeletionTimestamp == nil && !controllerutil.ContainsFinalizer(gametype, TypeFinalizer) {
		logger.Info("Adding finalizer to gametype")
		controllerutil.AddFinalizer(gametype, TypeFinalizer)
		if err := r.Update(ctx, gametype); err != nil {
			r.emitEventf(gametype, corev1.EventTypeWarning, utils.ReasonGametypeInitialized, "failed to add finalizers: %s", err)
			logger.Error(err, "Failed to add finalizer to gametype")
			return ctrl.Result{Requeue: true}, err
		}
		r.emitEvent(gametype, corev1.EventTypeNormal, utils.ReasonGametypeInitialized, "Added finalizers to game")
		return ctrl.Result{Requeue: true}, nil
	}

	// Handle resource deletion
	if gametype.DeletionTimestamp != nil || !gametype.GetDeletionTimestamp().IsZero() {
		logger.Info("Handling deletion of gametype")
		if err := r.handleDeletion(ctx, gametype, logger); err != nil {
			r.emitEventf(gametype, corev1.EventTypeWarning, utils.ReasonGametypeInitialized, "failed to remove finalizers: %s", err)
			logger.Error(err, "Failed to handle gametype deletion")
			return ctrl.Result{Requeue: true}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	result, err, done := r.handleUpdating(ctx, gametype, logger)
	if done {
		return result, err
	}

	return ctrl.Result{Requeue: true}, nil
}

func (r *GameTypeReconciler) handleUpdating(ctx context.Context, gametype *networkv1alpha1.GameType, logger logr.Logger) (ctrl.Result, error, bool) {
	//Check fleet count, if len(fleets) < 1 then need to create a fleet
	//Check if currentFleet is up to date with spec
	//If its not for whatever reason, create a new fleet with the same amount of replicas
	//If fleet amount is more than 1 then delete the oldest fleet first
	//If 1 fleet, update its replica count

	fleets, err := utils.GetFleetsForType(ctx, r.Client, gametype, logger)
	if err != nil {
		return ctrl.Result{}, err, true
	}
	if len(fleets.Items) == 0 {
		_, err := r.handleCreation(ctx, gametype, logger)
		if err != nil {
			return ctrl.Result{Requeue: true}, err, true
		}
		r.emitEvent(gametype, corev1.EventTypeNormal, utils.ReasonGametypeInitialized, "Created initial fleet")
		return ctrl.Result{Requeue: true}, nil, true
	}
	if len(fleets.Items) == 1 {
		fleet := fleets.Items[0]
		gametype.Status.CurrentFleetName = fleet.Name
		if err := r.Status().Update(ctx, gametype); err != nil {
			return ctrl.Result{Requeue: true}, err, true
		}
		if !networkv1alpha1.AreFleetsPodsEqual(&fleet.Spec, &gametype.Spec.FleetSpec) {
			r.emitEvent(gametype, corev1.EventTypeNormal, utils.ReasonGametypeSpecUpdated, "Creating new fleet")
			res, err := r.handleCreation(ctx, gametype, logger)
			return res, err, true
		} else {
			if fleet.Spec.Scaling.Replicas != int32(gametype.Spec.Scaling.CurrentReplicas) {
				fleet.Spec.Scaling.Replicas = int32(gametype.Spec.Scaling.CurrentReplicas)
				if err := r.Update(ctx, &fleet); err != nil {
					return ctrl.Result{Requeue: true}, err, true
				}
				r.emitEventf(gametype, corev1.EventTypeNormal, utils.ReasonGametypeReplicasUpdated, "Scaling gametype to %d", fleet.Spec.Scaling.Replicas)
			}
		}
	}
	if len(fleets.Items) > 1 {
		var oldestFleet *networkv1alpha1.Fleet
		for _, fleet := range fleets.Items {
			if oldestFleet == nil {
				oldestFleet = &fleet
			} else if fleet.CreationTimestamp.Before(&oldestFleet.CreationTimestamp) {
				oldestFleet = &fleet
			}
		}
		r.emitEvent(gametype, corev1.EventTypeNormal, utils.ReasonGametypeSpecUpdated, "Deleting extra fleet")

		if oldestFleet != nil && oldestFleet.GetDeletionTimestamp() == nil {
			if err := r.Delete(ctx, oldestFleet); err != nil {
				return ctrl.Result{}, err, true
			}
		}
	}
	return ctrl.Result{}, nil, false
}

// SetupWithManager sets up the controller with the Manager.
func (r *GameTypeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&networkv1alpha1.GameType{}).
		Owns(&networkv1alpha1.Fleet{}).
		WithOptions(controller.Options{MaxConcurrentReconciles: 10}).
		Complete(r)
}

func (r *GameTypeReconciler) handleDeletion(ctx context.Context, gametype *networkv1alpha1.GameType, logger logr.Logger) error {
	if controllerutil.ContainsFinalizer(gametype, TypeFinalizer) {
		//Finalizer not yet removed, we can presume that fleet deletion in progress or starting
		fleets, err := utils.GetFleetsForType(ctx, r.Client, gametype, logger)
		if err != nil {
			return err
		}
		for _, fleet := range fleets.Items {
			if err := r.Delete(ctx, &fleet); err != nil {
				r.emitEventf(gametype, corev1.EventTypeWarning, utils.ReasonGametypeServersDeleted, "Failed to delete fleet %s", fleet.Name)
				return err
			}
		}
		fleets, err = utils.GetFleetsForType(ctx, r.Client, gametype, logger)
		if err != nil {
			return err
		}
		if len(fleets.Items) == 0 {
			controllerutil.RemoveFinalizer(gametype, TypeFinalizer)
			if err := r.Update(ctx, gametype); err != nil {
				return err
			}
			r.emitEvent(gametype, corev1.EventTypeNormal, utils.ReasonGametypeServersDeleted, "Removed finalizer")
		}
	}
	return nil
}

func (r *GameTypeReconciler) handleCreation(ctx context.Context, gametype *networkv1alpha1.GameType, logger logr.Logger) (ctrl.Result, error) {
	fleet := utils.GetFleetObjectForType(gametype)
	if err := r.Create(ctx, fleet); err != nil {
		r.emitEventf(gametype, corev1.EventTypeWarning, utils.ReasonGametypeReplicasUpdated, "Failed to create new fleet %s", err)
		logger.Error(err, "failed to create a new fleet for gametype")
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *GameTypeReconciler) emitEvent(object runtime.Object, eventtype string, reason utils.EventReason, message string) {
	r.Recorder.Event(object, eventtype, string(reason), message)
}

func (r *GameTypeReconciler) emitEventf(object runtime.Object, eventtype string, reason utils.EventReason, message string, args ...interface{}) {
	r.Recorder.Eventf(object, eventtype, string(reason), message, args...)
}
