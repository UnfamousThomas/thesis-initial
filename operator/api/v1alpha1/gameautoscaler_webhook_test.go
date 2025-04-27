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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

var _ = Describe("GameAutoscaler Webhook", func() {

	Context("Should default the autoscaler", func() {
		gameautoscaler := &GameAutoscaler{
			Spec: GameAutoscalerSpec{
				GameName: "random",
				AutoscalePolicy: AutoscalePolicy{
					Type:                  Webhook,
					WebhookAutoscalerSpec: WebhookAutoscalerSpec{},
				},
			},
		}
		gameautoscaler.Default()
	})
	Context("When creating GameAutoscaler under Validating Webhook", func() {
		It("Should deny if a required field is empty", func() {
			service := Service{
				Name:      "a",
				Namespace: "default",
				Port:      33,
			}
			gameautoscaler := &GameAutoscaler{
				Spec: GameAutoscalerSpec{
					GameName: "random",
					AutoscalePolicy: AutoscalePolicy{
						Type:                  Webhook,
						WebhookAutoscalerSpec: WebhookAutoscalerSpec{},
					},
				},
			}
			By("Fails if no sync")
			_, err := gameautoscaler.ValidateCreate()
			Expect(err).To(HaveOccurred())
			By("Fails if the game name is empty")
			gameautoscaler.Spec.GameName = ""
			_, err = gameautoscaler.ValidateCreate()
			Expect(err).To(HaveOccurred())
			gameautoscaler.Spec.GameName = "game"

			By("Fails if the policy is empty")
			gameautoscaler.Spec.AutoscalePolicy = AutoscalePolicy{}
			gameautoscaler.Spec.Sync = Sync{
				Type: FixedInterval,
				Time: &metav1.Duration{Duration: 5 * time.Second},
			}
			_, err = gameautoscaler.ValidateCreate()
			Expect(err).To(HaveOccurred())

			By("Fails if no service and url")
			gameautoscaler.Spec.AutoscalePolicy = AutoscalePolicy{
				Type:                  Webhook,
				WebhookAutoscalerSpec: WebhookAutoscalerSpec{},
			}
			_, err = gameautoscaler.ValidateCreate()
			Expect(err).To(HaveOccurred())

			By("Fails if time is not specified")
			gameautoscaler.Spec.AutoscalePolicy.WebhookAutoscalerSpec.Service = &service
			gameautoscaler.Spec.Sync.Time = nil
			_, err = gameautoscaler.ValidateCreate()
			Expect(err).To(HaveOccurred())

			By("Succeeds if no issues")
			gameautoscaler.Spec.Sync.Time = &metav1.Duration{Duration: 5 * time.Second}
			_, err = gameautoscaler.ValidateCreate()
			Expect(err).To(Not(HaveOccurred()))
		})

		It("Should work the same for updating", func() {
			service := Service{
				Name:      "a",
				Namespace: "default",
				Port:      33,
			}
			gameautoscaler := &GameAutoscaler{
				Spec: GameAutoscalerSpec{
					GameName: "random",
					AutoscalePolicy: AutoscalePolicy{
						Type:                  Webhook,
						WebhookAutoscalerSpec: WebhookAutoscalerSpec{},
					},
				},
			}
			By("Fails if no sync")
			_, err := gameautoscaler.ValidateUpdate(nil)
			Expect(err).To(HaveOccurred())
			By("Fails if the game name is empty")
			gameautoscaler.Spec.GameName = ""
			_, err = gameautoscaler.ValidateUpdate(nil)
			Expect(err).To(HaveOccurred())
			gameautoscaler.Spec.GameName = "game"

			By("Fails if the policy is empty")
			gameautoscaler.Spec.AutoscalePolicy = AutoscalePolicy{}
			gameautoscaler.Spec.Sync = Sync{
				Type: FixedInterval,
				Time: &metav1.Duration{Duration: 5 * time.Second},
			}
			_, err = gameautoscaler.ValidateUpdate(nil)
			Expect(err).To(HaveOccurred())

			By("Fails if no service and url")
			gameautoscaler.Spec.AutoscalePolicy = AutoscalePolicy{
				Type:                  Webhook,
				WebhookAutoscalerSpec: WebhookAutoscalerSpec{},
			}
			_, err = gameautoscaler.ValidateUpdate(nil)
			Expect(err).To(HaveOccurred())

			By("Fails if time is not specified")
			gameautoscaler.Spec.AutoscalePolicy.WebhookAutoscalerSpec.Service = &service
			gameautoscaler.Spec.Sync.Time = nil
			_, err = gameautoscaler.ValidateUpdate(nil)
			Expect(err).To(HaveOccurred())

			By("Succeeds if no issues")
			gameautoscaler.Spec.Sync.Time = &metav1.Duration{Duration: 5 * time.Second}
			_, err = gameautoscaler.ValidateUpdate(nil)
			Expect(err).To(Not(HaveOccurred()))
		})

		It("Should validate delete", func() {
			gameautoscaler := &GameAutoscaler{
				Spec: GameAutoscalerSpec{
					GameName: "random",
					AutoscalePolicy: AutoscalePolicy{
						Type:                  Webhook,
						WebhookAutoscalerSpec: WebhookAutoscalerSpec{},
					},
				},
			}
			_, err := gameautoscaler.ValidateDelete()
			Expect(err).To(Not(HaveOccurred()))

		})
	})

})
