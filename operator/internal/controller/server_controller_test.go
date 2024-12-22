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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	networkv1alpha1 "github.com/unfamousthomas/thesis-operator/api/v1alpha1"
)

var allowed bool = false

type TestChecker struct{}

func (p TestChecker) GetPlayerCount(server *networkv1alpha1.Server) (int32, error) {
	return 0, nil
}

func (p TestChecker) IsDeletionAllowed(server *networkv1alpha1.Server) (bool, error) {
	return allowed, nil
}

var _ = Describe("ServerReconciler", func() {
	Context("Reconcile logic", func() {
		const (
			ServerName      = "test-server"
			ServerNamespace = "default"
		)

		var (
			ctx            context.Context
			reconciler     *ServerReconciler
			namespacedName types.NamespacedName
		)

		BeforeEach(func() {
			ctx = context.Background()
			namespacedName = types.NamespacedName{
				Name:      ServerName,
				Namespace: ServerNamespace,
			}
			checker := TestChecker{}
			reconciler = &ServerReconciler{
				Client:          k8sClient,
				Scheme:          k8sClient.Scheme(),
				DeletionAllowed: checker,
				PlayerCount:     checker,
			}

			By("Creating a Server resource")
			server := &networkv1alpha1.Server{
				ObjectMeta: metav1.ObjectMeta{
					Name:      ServerName,
					Namespace: ServerNamespace,
				},
				Spec: networkv1alpha1.ServerSpec{
					Pod: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name:  "nginx-test",
								Image: "nginx:latest",
							},
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, server)).To(Succeed())
		})

		AfterEach(func() {
			By("Deleting the Server resource if it exists")
			server := &networkv1alpha1.Server{}
			_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
			Expect(err).To(BeNil())
			_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
			Expect(err).To(BeNil())
			err = k8sClient.Get(ctx, namespacedName, server)
			if err == nil {
				Expect(k8sClient.Delete(ctx, server)).To(Succeed())
				_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
				Expect(err).To(BeNil())
				// Wait until the resource is deleted
				Eventually(func() bool {
					err := k8sClient.Get(ctx, namespacedName, server)
					return errors.IsNotFound(err)
				}).Should(BeTrue())
			}
		})

		It("should add a finalizer if not present", func() {
			By("Reconciling the resource")
			_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
			Expect(err).NotTo(HaveOccurred())

			By("Validating the finalizer is added")
			server := &networkv1alpha1.Server{}
			_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
			Expect(err).To(BeNil())
			Expect(k8sClient.Get(ctx, namespacedName, server)).To(Succeed())
			//Expect(controllerutil.ContainsFinalizer(server, FINALIZER)).To(BeTrue())
		})

		It("should create a Pod for the Server", func() {
			By("Reconciling the resource")
			_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
			Expect(err).NotTo(HaveOccurred())
			_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
			Expect(err).NotTo(HaveOccurred())
			By("Validating a Pod is created")
			pod := &corev1.Pod{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      ServerName + "-pod",
				Namespace: ServerNamespace,
			}, pod)).To(Succeed())
		})

		It("should handle deletion of the Server and remove the Pod", func() {
			By("Deleting the Server resource")
			server := &networkv1alpha1.Server{}
			Expect(k8sClient.Get(ctx, namespacedName, server)).To(Succeed())
			Expect(k8sClient.Delete(ctx, server)).To(Succeed())

			By("Reconciling the deletion")
			_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
			Expect(err).NotTo(HaveOccurred())

			By("Validating the Pod is deleted")
			pod := &corev1.Pod{}
			err = k8sClient.Get(ctx, types.NamespacedName{
				Name:      ServerName + "-pod",
				Namespace: ServerNamespace,
			}, pod)
			Expect(errors.IsNotFound(err)).To(BeTrue())

			By("Validating the finalizer is removed")
			err = k8sClient.Get(ctx, namespacedName, server)
			Expect(errors.IsNotFound(err)).To(BeTrue())
		})

	})
})
