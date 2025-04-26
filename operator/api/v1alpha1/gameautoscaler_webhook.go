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

package v1alpha1

import (
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var gameautoscalerlog = logf.Log.WithName("gameautoscaler-resource")

// SetupWebhookWithManager will setup the manager to manage the webhooks
func (r *GameAutoscaler) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-network-unfamousthomas-me-v1alpha1-gameautoscaler,mutating=true,failurePolicy=fail,sideEffects=None,groups=network.unfamousthomas.me,resources=gameautoscalers,verbs=create;update,versions=v1alpha1,name=mgameautoscaler.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &GameAutoscaler{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *GameAutoscaler) Default() {
	//I don't think defaulting currently needed
}

// NOTE: The 'path' attribute must follow a specific pattern and should not be modified directly here.
// Modifying the path for an invalid path can cause API server errors; failing to locate the webhook.
// +kubebuilder:webhook:path=/validate-network-unfamousthomas-me-v1alpha1-gameautoscaler,mutating=false,failurePolicy=fail,sideEffects=None,groups=network.unfamousthomas.me,resources=gameautoscalers,verbs=create;update;delete,versions=v1alpha1,name=vgameautoscaler.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &GameAutoscaler{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *GameAutoscaler) ValidateCreate() (admission.Warnings, error) {
	if r.Spec.GameName == "" {
		return admission.Warnings{}, fmt.Errorf("GameAutoscaler must specify GameName")
	}
	if (r.Spec.Sync == Sync{}) {
		return nil, fmt.Errorf("cannot create GameAutoscaler without Sync")
	}
	if (r.Spec.AutoscalePolicy == AutoscalePolicy{}) {
		return nil, fmt.Errorf("cannot create GameAutoscaler without AutoscalePolicy")
	}

	autoscalePolicy := r.Spec.AutoscalePolicy
	webhookautoscaler := autoscalePolicy.WebhookAutoscalerSpec
	if autoscalePolicy.Type == Webhook && webhookautoscaler.Service == nil && webhookautoscaler.Url == nil {
		return nil, fmt.Errorf("cannot create GameAutoscaler without url or service specified")
	}

	if r.Spec.Sync.Time == nil || r.Spec.Sync.Time.Milliseconds() <= 0 {
		return nil, fmt.Errorf("cannot create GameAutoscaler without proper time")
	}
	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *GameAutoscaler) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	//Same logic as creation
	return r.ValidateCreate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *GameAutoscaler) ValidateDelete() (admission.Warnings, error) {
	//No limitations
	return nil, nil
}
