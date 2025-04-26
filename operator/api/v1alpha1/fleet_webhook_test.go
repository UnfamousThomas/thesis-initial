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
		})
	})

	Context("When creating Fleet under Validating Webhook", func() {
		It("Should deny if a required field is empty", func() {
			fleet := &Fleet{
				Spec: FleetSpec{
					ServerSpec: ServerSpec{
						Pod: corev1.PodSpec{
							Containers: []corev1.Container{
								{},
							},
						},
					},
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "default",
					Namespace: "default",
				},
			}

			_, err := fleet.ValidateCreate()
			Expect(err).To(HaveOccurred())
		})

		It("Should admit if all required fields are provided", func() {
			fleet := &Fleet{
				Spec: FleetSpec{
					ServerSpec: ServerSpec{
						Pod: corev1.PodSpec{
							Containers: []corev1.Container{
								{},
							},
						},
					},
					Scaling: FleetScaling{
						Replicas:          1,
						PrioritizeAllowed: false,
						AgePriority:       OldestFirst,
					},
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "default",
					Namespace: "default",
				},
			}
			By("Fails if image missing")
			_, err := fleet.ValidateCreate()
			Expect(err).To(HaveOccurred())

			By("Fails if timeout missing")
			fleet.Spec.ServerSpec.Pod.Containers[0].Image = "image"
			_, err = fleet.ValidateCreate()
			Expect(err).To(HaveOccurred())

			By("Fails if agepriority missing")
			fleet.Spec.Scaling.AgePriority = ""
			fleet.Default()
			_, err = fleet.ValidateCreate()
			Expect(err).To(HaveOccurred())

			By("Fail if invalid agepriority")
			fleet.Spec.Scaling.AgePriority = "randompriority"
			_, err = fleet.ValidateCreate()
			Expect(err).To(HaveOccurred())

			By("Succeeds if fields are present")
			fleet.Spec.Scaling.AgePriority = OldestFirst
			_, err = fleet.ValidateCreate()
			Expect(err).To(Not(HaveOccurred()))
		})

		It("Should allow update when correct details", func() {
			initialFleet := &Fleet{
				Spec: FleetSpec{
					ServerSpec: ServerSpec{
						Pod: corev1.PodSpec{
							Containers: []corev1.Container{
								{},
							},
						},
						AllowForceDelete: false,
						TimeOut:          &metav1.Duration{Duration: time.Minute * 40},
					},
					Scaling: FleetScaling{
						Replicas:          1,
						PrioritizeAllowed: false,
						AgePriority:       OldestFirst,
					},
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "default",
					Namespace: "default",
				},
			}

			newFleet := &Fleet{
				Spec: FleetSpec{
					ServerSpec: ServerSpec{
						Pod: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Image: "foo:bar",
								},
							},
						},
						TimeOut:          &metav1.Duration{Duration: time.Minute * 20},
						AllowForceDelete: true,
					},
					Scaling: FleetScaling{
						Replicas:          1,
						PrioritizeAllowed: false,
						AgePriority:       OldestFirst,
					},
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "default",
					Namespace: "default",
				},
			}

			By("Fails when not equal pods")
			_, err := newFleet.ValidateUpdate(initialFleet)
			Expect(err).To(HaveOccurred())

			By("Warns when server spec differences")
			newFleet.Spec.ServerSpec.Pod = initialFleet.Spec.ServerSpec.Pod
			warn, err := newFleet.ValidateUpdate(initialFleet)
			Expect(err).NotTo(HaveOccurred())
			Expect(warn).To(Not(BeNil()))
			Expect(len(warn)).To(Equal(2))

			By("Fails when invalid priority")
			newFleet.Spec.Scaling.AgePriority = "randompriority"
			_, err = newFleet.ValidateUpdate(initialFleet)
			Expect(err).To(HaveOccurred())

			By("Fails when invalid old type")
			newFleet.Spec.Scaling.AgePriority = OldestFirst
			_, err = newFleet.ValidateUpdate(&Server{})
			Expect(err).To(HaveOccurred())
		})

		It("Should allow delete", func() {
			fleet := &Fleet{
				Spec: FleetSpec{
					ServerSpec: ServerSpec{
						Pod: corev1.PodSpec{
							Containers: []corev1.Container{
								{},
							},
						},
						AllowForceDelete: false,
						TimeOut:          &metav1.Duration{Duration: time.Minute * 40},
					},
					Scaling: FleetScaling{
						Replicas:          1,
						PrioritizeAllowed: false,
						AgePriority:       OldestFirst,
					},
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "default",
					Namespace: "default",
				},
			}
			_, err := fleet.ValidateDelete()
			Expect(err).To(Not(HaveOccurred()))
		})
	})

})
