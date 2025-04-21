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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
)

// FleetSpec defines the desired state of Fleet
type FleetSpec struct {
	ServerSpec ServerSpec   `json:"spec"`
	Scaling    FleetScaling `json:"scaling"`
}

type Priority string

var validPriorities = map[Priority]struct{}{
	OldestFirst: {},
	NewestFirst: {},
	// Add new priorities here as needed
}

const (
	OldestFirst Priority = "oldest_first"
	NewestFirst Priority = "newest_first"
)

type FleetScaling struct {
	// How many replicas of the servers should exist
	// +kubebuilder:default=1
	Replicas int32 `json:"replicas"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=true
	// If we should first delete the servers where deletion is allowed
	PrioritizeAllowed bool `json:"prioritizeAllowed"`
	// Whether we should first delete the oldest or newest
	// +kubebuilder:default=oldest_first
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum=oldest_first;smallest_first
	AgePriority Priority `json:"agePriority"`
}

// FleetStatus defines the observed state of Fleet
type FleetStatus struct {
	Conditions      []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
	CurrentReplicas int32              `json:"current_replicas,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Desired Replicas",type=integer,JSONPath=`.spec.scaling.replicas`
// +kubebuilder:printcolumn:name="Current Replicas",type=integer,JSONPath=`.status.current_replicas`

// Fleet is the Schema for the fleets API
type Fleet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FleetSpec   `json:"spec,omitempty"`
	Status FleetStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// FleetList contains a list of Fleet
type FleetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Fleet `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Fleet{}, &FleetList{})
}

func AreFleetsPodsEqual(fleet1, fleet2 *FleetSpec) bool {
	return reflect.DeepEqual(fleet1.ServerSpec.Pod, fleet2.ServerSpec.Pod)
}
