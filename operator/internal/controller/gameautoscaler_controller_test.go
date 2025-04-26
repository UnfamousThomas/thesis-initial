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
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/unfamousthomas/thesis-operator/internal/utils"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	networkv1alpha1 "github.com/unfamousthomas/thesis-operator/api/v1alpha1"
)

type TestWebhook struct {
	Scale    bool
	Replicas int
	Error    bool
}

func (t *TestWebhook) SendScaleWebhookRequest(autoscaler *networkv1alpha1.GameAutoscaler, gametype *networkv1alpha1.GameType) (utils.AutoscaleResponse, error) {
	if t.Error {
		return utils.AutoscaleResponse{}, fmt.Errorf("random error with webhook")
	}
	return utils.AutoscaleResponse{
		Scale:           t.Scale,
		DesiredReplicas: t.Replicas,
	}, nil
}

var duration = metav1.Duration{Duration: 5 * time.Second}
var path = "/scale"

const resourceName = "test-resource"
const namespace = "default"

var basicGameautoscaler = networkv1alpha1.GameAutoscalerSpec{
	GameName: resourceName,
	AutoscalePolicy: networkv1alpha1.AutoscalePolicy{
		Type: networkv1alpha1.Webhook,
		WebhookAutoscalerSpec: networkv1alpha1.WebhookAutoscalerSpec{
			Path: &path,
			Service: &networkv1alpha1.Service{
				Name:      "some-random-service",
				Namespace: metav1.NamespaceDefault,
				Port:      8080,
			},
		},
	},
	Sync: networkv1alpha1.Sync{
		Type: networkv1alpha1.FixedInterval,
		Time: &duration,
	},
}

var _ = Describe("GameAutoscaler Controller", func() {

	Context("When reconciling a resource", func() {

		ctx := context.Background()

		autoscalerNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: namespace,
		}
		gameTypeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: namespace,
		}

		BeforeEach(func() {
			By("creating a new game to match the autoscaler")
			gametype := &networkv1alpha1.GameType{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: namespace,
				},
			}
			gameautoscaler := &networkv1alpha1.GameAutoscaler{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: namespace,
				}}

			err := k8sClient.Get(ctx, gameTypeNamespacedName, gametype)
			if err != nil && errors.IsNotFound(err) {
				resource := &networkv1alpha1.GameType{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: namespace,
					},
					Spec: basicGametypeSpec,
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
				Eventually(func() error {
					return k8sClient.Get(ctx, gameTypeNamespacedName, &networkv1alpha1.GameType{})
				}, time.Second*5, time.Millisecond*100).Should(Succeed())

			}
			By("creating the custom resource for the Kind GameAutoscaler")
			err = k8sClient.Get(ctx, autoscalerNamespacedName, gameautoscaler)
			if err != nil && errors.IsNotFound(err) {
				resource := &networkv1alpha1.GameAutoscaler{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: basicGameautoscaler,
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}

		})
		AfterEach(func() {
			By("Check if gameautoscaler exists")
			gameautoscaler := &networkv1alpha1.GameAutoscaler{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: namespace,
				},
			}
			err := k8sClient.Get(ctx, autoscalerNamespacedName, gameautoscaler)
			Expect(err).To(BeNil())

			By("Cleanup the specific gameautoscaler instance")
			err = k8sClient.Delete(ctx, gameautoscaler)
			Expect(err).To(BeNil())

			Eventually(func() error {
				err := k8sClient.Get(ctx, autoscalerNamespacedName, gameautoscaler)
				if err != nil && !errors.IsNotFound(err) {
					return fmt.Errorf("error deleting game autoscaler: %w", err)
				}
				return nil
			}, time.Second*5, time.Millisecond*100).Should(Succeed())
			By("Check if autoscaler was deleted")
			err = k8sClient.Get(ctx, autoscalerNamespacedName, gameautoscaler)
			Expect(err).To(Not(BeNil()))

			By("Check if game exists")
			game := &networkv1alpha1.GameType{
				ObjectMeta: metav1.ObjectMeta{
					Name:      gameTypeNamespacedName.Name,
					Namespace: namespace,
				},
			}

			err = k8sClient.Get(ctx, gameTypeNamespacedName, game)
			Expect(err).To(BeNil())

			By("Cleanup the specific game instance")
			err = k8sClient.Delete(ctx, game)
			Expect(err).To(BeNil())

			Eventually(func() error {
				err := k8sClient.Get(ctx, gameTypeNamespacedName, game)
				if !errors.IsNotFound(err) {
					return fmt.Errorf("error deleting game: %w", err)
				}
				return nil
			}, time.Second*5, time.Millisecond*100).Should(Succeed())

			By("Check if game was deleted")
			err = k8sClient.Get(ctx, gameTypeNamespacedName, game)
			Expect(err).To(Not(BeNil()))
		})

		It("should successfully reconcile the gameautoscaler", func() {
			By("create the reconciler for gameautoscaler")
			hook := &TestWebhook{
				Scale:    false,
				Replicas: 1,
				Error:    false,
			}
			controllerReconciler := &GameAutoscalerReconciler{
				Client:   k8sClient,
				Scheme:   k8sClient.Scheme(),
				Webhook:  hook,
				Recorder: NewFakeRecorder(),
			}

			By("first reconciling for gameautoscaler")
			res, err := controllerReconciler.Reconcile(ctx, reconcile.Request{NamespacedName: autoscalerNamespacedName})
			Expect(err).To(BeNil())
			Expect(res.RequeueAfter).To(BeEquivalentTo(5 * time.Second))

			By("second reconciling for gameautoscaler")
			res, err = controllerReconciler.Reconcile(ctx, reconcile.Request{NamespacedName: autoscalerNamespacedName})
			Expect(err).To(BeNil())
			Expect(res.RequeueAfter).To(BeEquivalentTo(5 * time.Second))

			By("reconcile not existing autoscaler")
			res, err = controllerReconciler.Reconcile(ctx, reconcile.Request{NamespacedName: types.NamespacedName{
				Namespace: namespace,
				Name:      "some-random-scaler",
			},
			})

			Expect(err).To(Not(BeNil()))
			Expect(err).To(Not(Succeed()))

			By("Reconcile with webhook error")
			hook.Error = true
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{NamespacedName: autoscalerNamespacedName})
			Expect(err).To(Not(BeNil()))
			hook.Error = false
		})

		It("Reconcile with scaling", func() {
			By("Setup reconciler")
			hook := &TestWebhook{
				Scale:    false,
				Replicas: 1,
				Error:    false,
			}
			controllerReconciler := &GameAutoscalerReconciler{
				Client:   k8sClient,
				Scheme:   k8sClient.Scheme(),
				Webhook:  hook,
				Recorder: NewFakeRecorder(),
			}

			By("Setup hook")
			hook.Scale = true
			hook.Replicas = 10

			By("Reconcile and check time")
			res, err := controllerReconciler.Reconcile(ctx, reconcile.Request{NamespacedName: autoscalerNamespacedName})
			Expect(err).To(BeNil())
			Expect(res.RequeueAfter).To(BeEquivalentTo(5 * time.Second))
			By("Check updated game")
			updatedGameType := networkv1alpha1.GameType{}
			err = k8sClient.Get(ctx, gameTypeNamespacedName, &updatedGameType)
			Expect(err).To(BeNil())
			Expect(updatedGameType.Spec.Scaling.CurrentReplicas).Should(BeEquivalentTo(10))
		})

		It("Reconcile with invalid types", func() {
			By("Setup reconciler")
			hook := &TestWebhook{
				Scale:    false,
				Replicas: 1,
				Error:    false,
			}
			controllerReconciler := &GameAutoscalerReconciler{
				Client:   k8sClient,
				Scheme:   k8sClient.Scheme(),
				Webhook:  hook,
				Recorder: NewFakeRecorder(),
			}

			gameautoscaler := &networkv1alpha1.GameAutoscaler{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: namespace,
				}}

			By("Reconcile")
			err := k8sClient.Get(ctx, autoscalerNamespacedName, gameautoscaler)
			Expect(err).To(BeNil())

			By("Invalid gametype reconcile")
			err = k8sClient.Get(ctx, autoscalerNamespacedName, gameautoscaler)
			Expect(err).To(BeNil())
			gameautoscaler.Spec.AutoscalePolicy.Type = networkv1alpha1.Webhook
			gameautoscaler.Spec.GameName = "thisdoesnotexist"
			err = k8sClient.Update(ctx, gameautoscaler)
			Expect(err).To(BeNil())
			Eventually(func() error {
				autoscaler := networkv1alpha1.GameAutoscaler{}
				err := k8sClient.Get(ctx, autoscalerNamespacedName, &autoscaler)
				if err != nil {
					return err
				}
				if autoscaler.Spec.GameName != "thisdoesnotexist" {
					return fmt.Errorf("still correct game")
				}
				return nil
			}, time.Second*5, time.Millisecond*100).Should(Succeed())
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{NamespacedName: autoscalerNamespacedName})
			Expect(err).ToNot(BeNil())
		})

		It("should fail update", func() {
			fakeClient := FakeFailClient{
				client:       k8sClient,
				FailUpdate:   true,
				FailCreate:   false,
				FailDelete:   false,
				FailGet:      false,
				FailList:     false,
				FailPatch:    false,
				FailGetOnPod: false,
			}
			hook := &TestWebhook{
				Scale:    false,
				Replicas: 1,
				Error:    false,
			}
			controllerReconciler := &GameAutoscalerReconciler{
				Client:   fakeClient,
				Scheme:   fakeClient.Scheme(),
				Webhook:  hook,
				Recorder: NewFakeRecorder(),
			}
			hook.Scale = true
			hook.Replicas = 10
			hook.Error = false

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Namespace: namespace,
					Name:      resourceName,
				},
			})
			Expect(err).To(Not(BeNil()))
		})

		It("Should emit the correct events", func() {
			recorder := NewFakeRecorder()

			gameautoscaler := &networkv1alpha1.GameAutoscaler{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: namespace,
				}}

			hook := &TestWebhook{
				Scale:    false,
				Replicas: 1,
				Error:    false,
			}
			controllerReconciler := &GameAutoscalerReconciler{
				Client:   k8sClient,
				Scheme:   k8sClient.Scheme(),
				Webhook:  hook,
				Recorder: recorder,
			}

			originalGametype := resourceName
			By("Update to invalid gamename")
			err := k8sClient.Get(ctx, autoscalerNamespacedName, gameautoscaler)
			Expect(err).To(BeNil())
			gameautoscaler.Spec.GameName = originalGametype + "-1"
			err = k8sClient.Update(ctx, gameautoscaler)
			Expect(err).To(BeNil())

			By("Reconcile after update")
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{NamespacedName: autoscalerNamespacedName})
			Expect(err).ToNot(BeNil())
			By("Check for fail to find gametype event")
			hasGametypeErrorEvent := false
			for _, event := range recorder.Events {
				if event.Message == "Failed to find the gametype" {
					hasGametypeErrorEvent = true
					break
				}
			}
			Expect(hasGametypeErrorEvent).To(BeTrue())

			By("Reset game name")
			gameautoscaler.Spec.GameName = originalGametype
			err = k8sClient.Update(ctx, gameautoscaler)
			Expect(err).To(BeNil())
			err = k8sClient.Get(ctx, autoscalerNamespacedName, gameautoscaler)
			Expect(err).To(BeNil())

			By("Check if scale event is emitted")
			hook.Scale = true
			hook.Replicas = 5
			hook.Error = false
			controllerReconciler.Webhook = hook
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{NamespacedName: autoscalerNamespacedName})
			Expect(err).To(BeNil())
			hasScaleEvent := false
			for _, event := range recorder.Events {
				if event.Message == "Scaling game to 5" {
					hasScaleEvent = true
					break
				}
			}
			Expect(hasScaleEvent).To(BeTrue())
		})
	})
})
