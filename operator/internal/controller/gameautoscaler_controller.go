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
	"github.com/unfamousthomas/thesis-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	networkv1alpha1 "github.com/unfamousthomas/thesis-operator/api/v1alpha1"
)

// GameAutoscalerReconciler reconciles a GameAutoscaler object
type GameAutoscalerReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Webhook  utils.Webhook
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=network.unfamousthomas.me,resources=gameautoscalers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=network.unfamousthomas.me,resources=gameautoscalers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=network.unfamousthomas.me,resources=gameautoscalers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.18.4/pkg/reconcile
func (r *GameAutoscalerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithValues("autoscaler", req.Name, "namespace", req.Namespace)

	autoscaler := &networkv1alpha1.GameAutoscaler{}
	if err := r.Get(ctx, req.NamespacedName, autoscaler); err != nil {
		logger.Error(err, "Failed to get autoscaler resource")
		return ctrl.Result{Requeue: true}, err
	}

	gametype := &networkv1alpha1.GameType{}
	namespacedGametype := types.NamespacedName{
		Name:      autoscaler.Spec.GameName,
		Namespace: autoscaler.Namespace,
	}
	if err := r.Get(ctx, namespacedGametype, gametype); err != nil {
		r.emitEvent(autoscaler, corev1.EventTypeWarning, utils.ReasonGameAutoscalerInvalidServer, "Failed to find the gametype")
		return ctrl.Result{Requeue: true}, err
	}

	//Make sure the type is fine
	if autoscaler.Spec.AutoscalePolicy.Type != networkv1alpha1.Webhook {
		r.emitEvent(autoscaler, corev1.EventTypeWarning, utils.ReasonGameAutoscalerInvalidAutoscalePolicy,
			"invalid game autoscaler policy type")
		return ctrl.Result{}, fmt.Errorf("%s is not a valid policy type", autoscaler.Spec.AutoscalePolicy.Type)
	}

	//Send request to defined webhook
	result, err := r.Webhook.SendScaleWebhookRequest(autoscaler, gametype)
	if err != nil {
		r.emitEventf(autoscaler, corev1.EventTypeWarning, utils.ReasonGameautoscalerWebhook, "failed to send the webhook request: %v", err)
		return ctrl.Result{RequeueAfter: time.Minute}, fmt.Errorf("failed to send scale webhook request: %w", err)
	}

	//Check that the sync type is fine
	if autoscaler.Spec.Sync.Type != networkv1alpha1.FixedInterval {
		r.emitEventf(autoscaler, corev1.EventTypeWarning, utils.ReasonGameAutoscalerInvalidSyncType, "%s is not a valid sync type", autoscaler.Spec.Sync.Type)
		return ctrl.Result{}, fmt.Errorf("%s is not a valid sync type, currently only fixed interval is supported", autoscaler.Spec.Sync.Type)
	}

	//If scaleing not requested, requeue
	if !result.Scale {
		return ctrl.Result{
			RequeueAfter: autoscaler.Spec.Sync.Time.Duration,
		}, nil
	}

	//Otherwise, scale to new replica count
	gametype.Spec.FleetSpec.Scaling.Replicas = int32(result.DesiredReplicas)
	if err := r.Client.Update(ctx, gametype); err != nil {
		r.emitEvent(autoscaler, corev1.EventTypeWarning, utils.ReasonGameautoscalerScale, "failed to update the gametype")
		return ctrl.Result{}, fmt.Errorf("failed to update gametype with new replica count: %w", err)
	}
	r.emitEventf(autoscaler, corev1.EventTypeNormal, utils.ReasonGameautoscalerScale, "Scaling game to %d", result.DesiredReplicas)

	//Requeue after the defined time
	return ctrl.Result{
		RequeueAfter: autoscaler.Spec.Sync.Time.Duration,
	}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GameAutoscalerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&networkv1alpha1.GameAutoscaler{}).
		Complete(r)
}

// emitEvent is used by the GameAutoscalerReconciler to easily add events to objects
func (r *GameAutoscalerReconciler) emitEvent(object runtime.Object, eventtype string, reason utils.EventReason, message string) {
	r.Recorder.Event(object, eventtype, string(reason), message)
}

// emitEventf is used by the GameAutoscalerReconciler to easily add events to objects with arguments
func (r *GameAutoscalerReconciler) emitEventf(object runtime.Object, eventtype string, reason utils.EventReason, message string, args ...interface{}) {
	r.Recorder.Eventf(object, eventtype, string(reason), message, args...)
}
