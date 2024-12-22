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
	"github.com/unfamousthomas/thesis-operator/internal/scaling"
	"github.com/unfamousthomas/thesis-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	networkv1alpha1 "github.com/unfamousthomas/thesis-operator/api/v1alpha1"
)

var FINALIZER = "servers.unfamousthomas.me/finalizer"

// ServerReconciler reconciles a Server object
type ServerReconciler struct {
	client.Client
	Scheme          *runtime.Scheme
	Recorder        record.EventRecorder
	DeletionAllowed scaling.Deletion
	PlayerCount     scaling.PlayerCount
}

// +kubebuilder:rbac:groups=network.unfamousthomas.me,resources=servers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=network.unfamousthomas.me,resources=servers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=network.unfamousthomas.me,resources=servers/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=events,verbs=create;patch
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patcch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.18.4/pkg/reconcile
func (r *ServerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithValues("server", req.Name, "namespace", req.Namespace)

	// Fetch the Server resource
	server := &networkv1alpha1.Server{}
	if err := r.Get(ctx, req.NamespacedName, server); err != nil {
		if client.IgnoreNotFound(err) != nil {
			logger.Error(err, "Failed to get Server resource")
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Handle finalizer addition
	if server.DeletionTimestamp == nil && !controllerutil.ContainsFinalizer(server, FINALIZER) {
		logger.Info("Adding finalizer to Server")
		controllerutil.AddFinalizer(server, FINALIZER)
		if err := r.Update(ctx, server); err != nil {
			logger.Error(err, "Failed to add finalizer")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	// Handle resource deletion
	if server.DeletionTimestamp != nil || !server.GetDeletionTimestamp().IsZero() {
		logger.Info("Handling deletion of Server")
		if err := r.handleDeletion(ctx, server, logger); err != nil { //todo
			logger.Error(err, "Failed to handle Server deletion")
			return ctrl.Result{}, err
		}
		logger.Info("Successfully finalized Server, removing finalizer")
		controllerutil.RemoveFinalizer(server, FINALIZER)
		if err := r.Update(ctx, server); err != nil {
			logger.Error(err, "Failed to remove finalizer")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil // Return after finalizer removal
	}

	// Ensure Pod exists
	podExists, err := r.ensurePodExists(ctx, server, logger)
	if err != nil {
		logger.Error(err, "Failed to ensure Pod exists for Server")
		return ctrl.Result{}, err
	}
	if !podExists {
		// If a Pod was created, exit early to requeue the reconciliation
		return ctrl.Result{}, nil
	}

	update, err := r.ensurePodFinalizer(ctx, server, logger)
	if err != nil || update {
		return ctrl.Result{}, err
	}

	if err := r.Status().Update(ctx, server); err != nil {
		logger.Error(err, "Failed to update Server resource")
		return ctrl.Result{}, err
	}
	logger.Info("Reconciliation finished")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&networkv1alpha1.Server{}).
		Owns(&corev1.Pod{}).
		WithOptions(controller.Options{MaxConcurrentReconciles: 10}).
		Complete(r)
}

func (r *ServerReconciler) ensurePodExists(ctx context.Context, server *networkv1alpha1.Server, logger logr.Logger) (bool, error) {
	pod := &corev1.Pod{}
	namespacedName := types.NamespacedName{Namespace: server.Namespace, Name: server.Name + "-pod"}
	err := r.Get(ctx, namespacedName, pod)

	if client.IgnoreNotFound(err) != nil {
		logger.Error(err, "Failed to get Pod resource")
		meta.SetStatusCondition(&server.Status.Conditions, metav1.Condition{
			Type:               "PodFailed",
			Status:             metav1.ConditionFalse,
			LastTransitionTime: metav1.Now(),
			Reason:             "GetPodFailed",
			Message:            "Failed to retrieve Pod from the cluster",
		})
		return false, err
	}

	if err != nil { // Pod does not exist
		newPod := utils.GetNewPod(server, server.Namespace)
		if err := r.Create(ctx, newPod); err != nil {
			meta.SetStatusCondition(&server.Status.Conditions, metav1.Condition{
				Type:               "PodFailed",
				Status:             metav1.ConditionFalse,
				LastTransitionTime: metav1.Now(),
				Reason:             "PodCreationFailed",
				Message:            "Failed to create the Pod",
			})
			return false, err
		}

		meta.SetStatusCondition(&server.Status.Conditions, metav1.Condition{
			Type:               "PodCreated",
			Status:             metav1.ConditionTrue,
			LastTransitionTime: metav1.Now(),
			Reason:             "PodCreatedSuccessfully",
			Message:            "Pod has been successfully created",
		})

		meta.SetStatusCondition(&server.Status.Conditions, metav1.Condition{
			Type:               "PodCreated",
			Status:             metav1.ConditionTrue,
			LastTransitionTime: metav1.Now(),
			Reason:             "PodCreatedSuccessfully",
			Message:            "Pod has been successfully created",
		})
		return false, nil
	}

	meta.SetStatusCondition(&server.Status.Conditions, metav1.Condition{
		Type:               "PodCreated",
		Status:             metav1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Reason:             "PodAlreadyExists",
		Message:            "Pod already exists",
	})
	return true, nil
}

func (r *ServerReconciler) handleDeletion(ctx context.Context, server *networkv1alpha1.Server, logger logr.Logger) error {
	pod := &corev1.Pod{}
	namespacedName := types.NamespacedName{Namespace: server.Namespace, Name: server.Name + "-pod"}
	if err := r.Get(ctx, namespacedName, pod); err != nil {
		return err
	}
	allowed, err := r.DeletionAllowed.IsDeletionAllowed(server, pod)
	if err != nil {
		logger.Error(err, "Failed to check if deletion allowed for Server")
		return err
	}
	if !allowed {
		logger.Info("Server deletion not currently allowed")
		return nil
	}

	if pod != nil {
		controllerutil.RemoveFinalizer(pod, FINALIZER)
		if err := r.Update(ctx, pod); err != nil {
			return err
		}

		if err := r.Get(ctx, namespacedName, pod); err != nil {
			return err
		}

		if err := r.Delete(ctx, pod); err != nil {
			meta.SetStatusCondition(&server.Status.Conditions, metav1.Condition{
				Type:               "Finalizing",
				Status:             metav1.ConditionFalse,
				LastTransitionTime: metav1.Now(),
				Reason:             "PodDeletionFailed",
				Message:            "Failed to delete the Pod during finalization",
			})
			return err
		}
	}

	meta.SetStatusCondition(&server.Status.Conditions, metav1.Condition{
		Type:               "Finalizing",
		Status:             metav1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Reason:             "PodDeleted",
		Message:            "Pod successfully deleted during finalization",
	})

	return nil
}

func (r *ServerReconciler) ensurePodFinalizer(ctx context.Context, server *networkv1alpha1.Server, logger logr.Logger) (bool, error) {
	pod := &corev1.Pod{}
	namespacedName := types.NamespacedName{Namespace: server.Namespace, Name: server.Name + "-pod"}
	if err := r.Get(ctx, namespacedName, pod); err != nil {
		return false, err
	}
	if controllerutil.ContainsFinalizer(pod, FINALIZER) {
		return false, nil
	}
	controllerutil.AddFinalizer(pod, FINALIZER)
	if err := r.Update(ctx, pod); err != nil {
		logger.Error(err, "Failed to add finalizer to pod")
		return false, err
	}
	return true, nil
}
