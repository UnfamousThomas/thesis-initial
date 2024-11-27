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

var FINALIZER = "servers.finalizers.unfamousthomas.me"

// ServerReconciler reconciles a Server object
type ServerReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
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
	logger := log.FromContext(ctx)

	server := &networkv1alpha1.Server{}
	err := r.Get(ctx, req.NamespacedName, server)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if meta.IsStatusConditionFalse(server.Status.Conditions, "PodReady") {
		pod := createPodObject(server, req.Namespace)

		if err := r.Create(ctx, pod); err != nil {
			logger.Error(err, "Failed to create matching pod for server")
			return ctrl.Result{}, err
		}

		meta.SetStatusCondition(&server.Status.Conditions, metav1.Condition{
			Type:    "PodReady",
			Status:  metav1.ConditionTrue,
			Reason:  "PodCreated",
			Message: "Pod has been successfully created",
		})
		if err := r.Status().Update(ctx, server); err != nil {
			logger.Error(err, "Failed to update server status")
		}
		return ctrl.Result{}, err
	}

	if server.CreationTimestamp.IsZero() && controllerutil.ContainsFinalizer(server, FINALIZER) {
		//Server being deleted...
		if meta.IsStatusConditionFalse(server.Status.Conditions, "PodDeleted") {

			pod := &corev1.Pod{}
			err := r.Get(ctx, types.NamespacedName{Name: server.Name + "-pod", Namespace: server.Namespace}, pod)
			if err != nil {
				logger.Info("Failed to get pod for server, nothing to delete")
				return ctrl.Result{}, nil
			}

			if err := r.Delete(ctx, pod); err != nil {
				logger.Error(err, "Failed to delete matching pod for server")
				return ctrl.Result{}, err
			}
			meta.SetStatusCondition(&server.Status.Conditions, metav1.Condition{
				Type:    "PodDeleted",
				Status:  metav1.ConditionTrue,
				Reason:  "PodDeleted",
				Message: "Pod has been successfully deleted",
			})
			if err := r.Status().Update(ctx, server); err != nil {
				logger.Error(err, "Failed to update server status")
			}
			return ctrl.Result{}, nil
		}
		controllerutil.RemoveFinalizer(server, FINALIZER)
		if err := r.Status().Update(ctx, server); err != nil {
			logger.Error(err, "Failed to update server status")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func createPodObject(server *networkv1alpha1.Server, namespace string) *corev1.Pod {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      server.Name + "-pod",
			Namespace: namespace,
			Labels: map[string]string{
				"server": server.Name,
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "server-container",
					Image: server.Spec.Image,
				},
			},
		},
	}
	return pod

}

// SetupWithManager sets up the controller with the Manager.
func (r *ServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&networkv1alpha1.Server{}).
		WithOptions(controller.Options{MaxConcurrentReconciles: 2}).
		Complete(r)
}
