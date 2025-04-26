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
	"errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"time"
)

// log is for logging in this package.
var gametypelog = logf.Log.WithName("gametype-resource")

// SetupWebhookWithManager will setup the manager to manage the webhooks
func (r *GameType) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-network-unfamousthomas-me-v1alpha1-gametype,mutating=true,failurePolicy=fail,sideEffects=None,groups=network.unfamousthomas.me,resources=gametypes,verbs=create;update,versions=v1alpha1,name=mgametype.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &GameType{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *GameType) Default() {
	if r.Spec.FleetSpec.ServerSpec.TimeOut == nil {
		r.Spec.FleetSpec.ServerSpec.TimeOut = &metav1.Duration{Duration: time.Minute * 40}
	}
}

// NOTE: The 'path' attribute must follow a specific pattern and should not be modified directly here.
// Modifying the path for an invalid path can cause API server errors; failing to locate the webhook.
// +kubebuilder:webhook:path=/validate-network-unfamousthomas-me-v1alpha1-gametype,mutating=false,failurePolicy=fail,sideEffects=None,groups=network.unfamousthomas.me,resources=gametypes,verbs=create;update;delete,versions=v1alpha1,name=vgametype.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &GameType{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *GameType) ValidateCreate() (admission.Warnings, error) {

	for _, container := range r.Spec.FleetSpec.ServerSpec.Pod.Containers {
		if container.Image == "" {
			return nil, errors.New("image is required for every container")
		}
	}
	if len(r.Spec.FleetSpec.ServerSpec.Pod.Containers) == 0 {
		return nil, errors.New("at least one container is required")
	}
	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *GameType) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *GameType) ValidateDelete() (admission.Warnings, error) {

	return nil, nil
}
