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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"time"
)

// log is for logging in this package.
var fleetlog = logf.Log.WithName("fleet-resource")

// SetupWebhookWithManager will setup the manager to manage the webhooks
func (r *Fleet) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-network-unfamousthomas-me-v1alpha1-fleet,mutating=true,failurePolicy=fail,sideEffects=None,groups=network.unfamousthomas.me,resources=fleets,verbs=create;update,versions=v1alpha1,name=mfleet.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Fleet{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Fleet) Default() {
	if r.Spec.ServerSpec.TimeOut == nil {
		r.Spec.ServerSpec.TimeOut = &metav1.Duration{Duration: time.Minute * 40}
	}
}

// NOTE: The 'path' attribute must follow a specific pattern and should not be modified directly here.
// Modifying the path for an invalid path can cause API server errors; failing to locate the webhook.
// +kubebuilder:webhook:path=/validate-network-unfamousthomas-me-v1alpha1-fleet,mutating=false,failurePolicy=fail,sideEffects=None,groups=network.unfamousthomas.me,resources=fleets,verbs=create;update;delete,versions=v1alpha1,name=vfleet.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Fleet{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Fleet) ValidateCreate() (admission.Warnings, error) {

	err := r.validatePriorities()
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Fleet) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	err := r.validatePriorities()
	if err != nil {
		return nil, err
	}

	oldFleet, ok := old.(*Fleet)
	if !ok {
		return nil, fmt.Errorf("expected old object to be *Fleet, got %T", old)
	}
	warnings := admission.Warnings{}
	if oldFleet.Spec.ServerSpec.TimeOut != r.Spec.ServerSpec.TimeOut {
		warnings = append(warnings, "New timeout will not affect previously created servers")
	}
	if oldFleet.Spec.ServerSpec.AllowForceDelete != r.Spec.ServerSpec.AllowForceDelete {
		warnings = append(warnings, "New allowForceDelete will not affect previously created servers")
	}
	if !arePodSpecsEqual(oldFleet.Spec.ServerSpec.Pod, r.Spec.ServerSpec.Pod) {
		return nil, fmt.Errorf("pod template cannot be updated")
	}

	return warnings, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Fleet) ValidateDelete() (admission.Warnings, error) {

	return nil, nil
}

func (r Fleet) validatePriorities() error {
	if _, exists := validPriorities[r.Spec.Scaling.AgePriority]; !exists {
		return fmt.Errorf("unknown priority %s", r.Spec.Scaling.AgePriority)
	}

	return nil
}
