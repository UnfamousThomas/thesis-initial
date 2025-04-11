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
	"errors"
	"github.com/unfamousthomas/thesis-operator/internal/utils"
	"k8s.io/apimachinery/pkg/types"
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
	Scheme  *runtime.Scheme
	Webhook utils.Webhook
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
		logger.Error(err, "Failed to get gametype resource")
		return ctrl.Result{Requeue: true}, err
	}

	if autoscaler.Spec.AutoscalePolicy.Type != networkv1alpha1.Webhook {
		logger.Error(errors.New("unable to handle strategies besides webhook"), "please implement new types")
		return ctrl.Result{}, nil
	}

	result, err := r.Webhook.SendScaleWebhookRequest(autoscaler, gametype)
	if err != nil {
		logger.Error(err, "Failed to send scale webhook")
		return ctrl.Result{RequeueAfter: time.Minute}, err
	}

	if autoscaler.Spec.Sync.Type != networkv1alpha1.FixedInterval {
		logger.Error(errors.New("unable to handle syncs besides fixed interval"), "please implement new types")
		return ctrl.Result{}, nil
	}

	secondsBetween := time.Second * time.Duration(autoscaler.Spec.Sync.FixedInterval)

	if !result.Scale {
		return ctrl.Result{
			RequeueAfter: secondsBetween,
		}, nil
	}

	gametype.Spec.Scaling.CurrentReplicas = result.DesiredReplicas

	if err := r.Client.Update(ctx, gametype); err != nil {
		logger.Error(err, "failed to update gametype with new replica count")
		return ctrl.Result{}, err
	}

	return ctrl.Result{
		RequeueAfter: secondsBetween,
	}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GameAutoscalerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&networkv1alpha1.GameAutoscaler{}).
		Complete(r)
}
