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

var _ = Describe("GameType Webhook", func() {

	Context("When creating GameType under Defaulting Webhook", func() {
		It("Should fill in the default value if a required field is empty", func() {

			gametype := GameType{
				Spec: GameTypeSpec{
					FleetSpec: FleetSpec{
						ServerSpec: ServerSpec{},
					},
				},
			}
			gametype.Default()
			Expect(*gametype.Spec.FleetSpec.ServerSpec.TimeOut).To(Equal(metav1.Duration{Duration: time.Minute * 40}))
		})
	})

	Context("When creating GameType under Validating Webhook", func() {
		It("Should deny if a required field is empty", func() {
			By("Check if fails with no containers")
			gametype := GameType{
				Spec: GameTypeSpec{
					FleetSpec: FleetSpec{
						ServerSpec: ServerSpec{},
					},
				},
			}
			_, err := gametype.ValidateCreate()
			Expect(err).ToNot(Succeed())
			By("Check if fails with no image but container")
			gametype.Spec.FleetSpec.ServerSpec.Pod = corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name: "yes",
					},
				},
			}
			_, err = gametype.ValidateCreate()
			Expect(err).ToNot(Succeed())
			By("Check if suceeds with image")
			gametype.Spec.FleetSpec.ServerSpec.Pod.Containers[0].Image = "someimage"
			_, err = gametype.ValidateCreate()
			Expect(err).To(Succeed())
		})

	})

	Context("When updating GameType under Validating Webhook", func() {
		It("Should deny if a required field is empty", func() {
			initialGameType := GameType{
				Spec: GameTypeSpec{
					FleetSpec: FleetSpec{
						ServerSpec: ServerSpec{
							Pod: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name:  "yes",
										Image: "someimage",
									},
								},
							},
						},
					},
				},
			}

			gametype := GameType{
				Spec: GameTypeSpec{
					FleetSpec: FleetSpec{
						ServerSpec: ServerSpec{
							Pod: corev1.PodSpec{
								Containers: []corev1.Container{
									{
										Name:  "yes",
										Image: "someimage2",
									},
								},
							},
						},
					},
				},
			}

			_, err := gametype.ValidateUpdate(&initialGameType)
			Expect(err).To(Succeed())
		})
	})

	Context("When deleting GameType under Validating Webhook", func() {
		It("Should deny if a required field is empty", func() {
			gametype := GameType{
				Spec: GameTypeSpec{
					FleetSpec: FleetSpec{
						ServerSpec: ServerSpec{},
					},
				},
			}
			_, err := gametype.ValidateDelete()
			Expect(err).To(Succeed())
		})
	})
})
