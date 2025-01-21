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
	"context"
	"github.com/unfamousthomas/thesis-operator/internal/utils"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	networkv1alpha1 "github.com/unfamousthomas/thesis-operator/api/v1alpha1"
)

type TestWebhook struct{}

func (t TestWebhook) SendScaleWebhookRequest(autoscaler *networkv1alpha1.GameAutoscaler, gametype *networkv1alpha1.GameType) (utils.AutoscaleResponse, error) {
	//TODO
	panic("implement me")
}

var _ = Describe("GameAutoscaler Controller", func() {

	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default",
		}
		gameautoscaler := &networkv1alpha1.GameAutoscaler{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind GameAutoscaler")
			err := k8sClient.Get(ctx, typeNamespacedName, gameautoscaler)
			if err != nil && errors.IsNotFound(err) {
				resource := &networkv1alpha1.GameAutoscaler{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: networkv1alpha1.GameAutoscalerSpec{
						GameName: "some-random-game",
						AutoscalePolicy: networkv1alpha1.AutoscalePolicy{
							Type: networkv1alpha1.Webhook,
							WebhookAutoscalerSpec: networkv1alpha1.WebhookAutoscalerSpec{
								Path: "/scale",
								Service: networkv1alpha1.Service{
									Name:      "some-random-service",
									Namespace: metav1.NamespaceDefault,
									Port:      8080,
								},
							},
						},
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			resource := &networkv1alpha1.GameAutoscaler{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance GameAutoscaler")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &GameAutoscalerReconciler{
				Client:  k8sClient,
				Scheme:  k8sClient.Scheme(),
				Webhook: &TestWebhook{},
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			// TODO(user): Add more specific assertions depending on your controller's reconciliation logic.
			// Example: If you expect a certain status condition after reconciliation, verify it here.
		})
	})
})
