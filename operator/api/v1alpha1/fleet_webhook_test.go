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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

var _ = Describe("Fleet Webhook", func() {

	Context("When creating Fleet under Defaulting Webhook", func() {
		It("Should fill in the default value if a required field is empty", func() {

			fleet := &Fleet{
				Spec: FleetSpec{
					ServerSpec: ServerSpec{
						Pod: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Image: "foo:bar",
								},
							},
						},
					},
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "default",
					Namespace: "default",
				},
			}
			fleet.Default()

			Expect(fleet.Spec.ServerSpec.TimeOut).To(Equal(&metav1.Duration{Duration: time.Minute * 40}))
			// TODO(user): Add your logic here

		})
	})

	Context("When creating Fleet under Validating Webhook", func() {
		It("Should deny if a required field is empty", func() {

			// TODO(user): Add your logic here

		})

		It("Should admit if all required fields are provided", func() {

			// TODO(user): Add your logic here

		})
	})

})
