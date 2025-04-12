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
	. "github.com/onsi/ginkgo/v2"
	networkv1alpha1 "github.com/unfamousthomas/thesis-operator/api/v1alpha1"
)

var basicFleetSpec = networkv1alpha1.FleetSpec{
	Scaling: networkv1alpha1.FleetScaling{
		Replicas:          3,
		PrioritizeAllowed: false,
		AgePriority:       networkv1alpha1.OldestFirst,
	},
	ServerSpec: basicServerSpec,
}

const testNs = "test-fleet-ns"

var _ = Describe("Fleet Controller", func() {
	Context("When reconciling a resource", func() {

	})
})
