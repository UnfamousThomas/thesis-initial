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
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	"time"
)

// log is for logging in this package.
var serverlog = logf.Log.WithName("server-resource")

// SetupWebhookWithManager will setup the manager to manage the webhooks
func (r *Server) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-network-unfamousthomas-me-v1alpha1-server,mutating=true,failurePolicy=fail,sideEffects=None,groups=network.unfamousthomas.me,resources=servers,verbs=create;update,versions=v1alpha1,name=mserver.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Server{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Server) Default() {
	if r.Spec.TimeOut == nil {
		r.Spec.TimeOut = &metav1.Duration{Duration: time.Minute * 40}
	}
}

// NOTE: The 'path' attribute must follow a specific pattern and should not be modified directly here.
// Modifying the path for an invalid path can cause API server errors; failing to locate the webhook.
// +kubebuilder:webhook:path=/validate-network-unfamousthomas-me-v1alpha1-server,mutating=false,failurePolicy=fail,sideEffects=None,groups=network.unfamousthomas.me,resources=servers,verbs=create;update;delete,versions=v1alpha1,name=vserver.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Server{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Server) ValidateCreate() (admission.Warnings, error) {
	if len(r.Spec.Pod.Containers) < 1 {
		return nil, errors.New("at least 1 container required")
	}
	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Server) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	oldServer, ok := old.(*Server)
	if !ok {
		return nil, fmt.Errorf("expected old object to be *Server, got %T", old)
	}
	if arePodSpecsEqual(oldServer.Spec.Pod, r.Spec.Pod) {
		return nil, errors.New("updating a servers pod spec is not allowed, please remake the server")
	}

	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Server) ValidateDelete() (admission.Warnings, error) {
	return nil, nil
}

func arePodSpecsEqual(a, b corev1.PodSpec) bool {
	return reflect.DeepEqual(a, b)
}
