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

var _ = Describe("Server Webhook", func() {

	Context("When creating Server under Defaulting Webhook", func() {
		It("Should fill in the default value if a required field is empty", func() {
			server := Server{
				Spec: ServerSpec{
					Pod: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Image: "image",
							},
						},
					},
				},
			}

			server.Default()
			Expect(server.Spec.TimeOut).To(Equal(&metav1.Duration{Duration: time.Minute * 40}))

		})
	})

	Context("When creating Server under Validating Webhook", func() {
		It("Should deny if a required field is empty for creation", func() {

			server := Server{
				Spec: ServerSpec{
					Pod: corev1.PodSpec{},
				},
			}

			By("Check if containers are required")
			_, err := server.ValidateCreate()
			Expect(err).To(HaveOccurred())

			By("Check if image is required")
			server.Spec.Pod.Containers = []corev1.Container{
				{
					Name: "test",
				},
			}
			_, err = server.ValidateCreate()
			Expect(err).To(HaveOccurred())

			By("Check if suceeds with image")
			server.Spec.Pod.Containers[0].Image = "image"
			_, err = server.ValidateCreate()
			Expect(err).To(Succeed())
		})
	})

	Context("When updating Server under Validating Webhook", func() {
		It("Should deny if a required field is invalid for update", func() {
			server := Server{
				Spec: ServerSpec{
					Pod: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "test",
								Image: "image",
							},
						},
					},
				},
			}
			By("Check if it errors with invalid type")
			_, err := server.ValidateUpdate(&Fleet{})
			Expect(err).To(HaveOccurred())

			By("Check if fails if different image")
			newServer := Server{
				Spec: ServerSpec{
					Pod: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "test",
								Image: "image2",
							},
						},
					},
				},
			}
			_, err = server.ValidateUpdate(&newServer)
			Expect(err).To(HaveOccurred())

			By("Check if suceeds when same pod spec")
			newServer.Spec.Pod.Containers[0].Image = "image"
			newServer.Spec.TimeOut = &metav1.Duration{Duration: time.Minute * 20}
			_, err = server.ValidateUpdate(&newServer)
			Expect(err).To(Succeed())
		})
	})
	Context("When deleting Server under Validating Webhook", func() {
		It("Should deny if a required field is invalid for deletion", func() {
			server := Server{
				Spec: ServerSpec{},
			}

			_, err := server.ValidateDelete()
			Expect(err).To(Succeed())
		})
	})

})
