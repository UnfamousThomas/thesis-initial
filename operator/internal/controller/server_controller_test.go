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
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	networkv1alpha1 "github.com/unfamousthomas/thesis-operator/api/v1alpha1"
)

var basicServerSpec = networkv1alpha1.ServerSpec{
	Pod: corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:  "nginx",
				Image: "nginx:1.7.9",
			},
		},
	},
	AllowForceDelete: false,
	TimeOut: &metav1.Duration{
		Duration: 5 * time.Minute,
	},
}

type TestChecker struct {
	deleteAllowed map[string]bool
}

func (p TestChecker) IsDeletionAllowed(server *networkv1alpha1.Server, pod *corev1.Pod) (bool, error) {
	//Basically for mocking deletion allowing behaviour, we just use a map
	return p.deleteAllowed[server.Name], nil
}

var checker = TestChecker{
	deleteAllowed: make(map[string]bool),
}

var _ = Describe("ServerReconciler", func() {
	Context("Reconcile logic", func() {
		const (
			ServerName      = "test-server"
			ServerNamespace = "default"
		)

		var (
			ctx            context.Context
			namespacedName types.NamespacedName
		)

		BeforeEach(func() {
			ctx = context.Background()
			namespacedName = types.NamespacedName{
				Name:      ServerName,
				Namespace: ServerNamespace,
			}

			By("Creating a Server resource")
			server := &networkv1alpha1.Server{
				ObjectMeta: metav1.ObjectMeta{
					Name:      ServerName,
					Namespace: ServerNamespace,
				},
				Spec: basicServerSpec,
			}
			//Do not allow deletions
			checker.deleteAllowed[server.Name] = false
			err := k8sClient.Create(ctx, server)
			Expect(err).To(Succeed())
		})

		AfterEach(func() {
			By("Deleting the Server resource if it exists")
			server := &networkv1alpha1.Server{
				ObjectMeta: metav1.ObjectMeta{
					Name:      ServerName,
					Namespace: ServerNamespace,
				},
			}
			reconciler := &ServerReconciler{
				Client:          k8sClient,
				Scheme:          k8sClient.Scheme(),
				DeletionAllowed: checker,
				Recorder:        NewFakeRecorder(),
			}
			_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
			Expect(err).To(BeNil())
			_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
			Expect(err).To(BeNil())
			err = k8sClient.Get(ctx, namespacedName, server)
			if err == nil {
				//Try to delete
				Expect(k8sClient.Delete(ctx, server)).To(Succeed())
				//This should allow due to delete not being allowed
				_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
				Expect(err).To(Not(BeNil()))
				checker.deleteAllowed[server.Name] = true
				_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
				Expect(err).To(BeNil())

				// Wait until the resource is deleted
				Eventually(func() bool {
					err := k8sClient.Get(ctx, namespacedName, server)
					return errors.IsNotFound(err)
				}).Should(BeTrue())
			}
		})

		It("should add a finalizer if not present", func() {
			reconciler := &ServerReconciler{
				Client:          k8sClient,
				Scheme:          k8sClient.Scheme(),
				DeletionAllowed: checker,
				Recorder:        NewFakeRecorder(),
			}
			By("Reconciling the resource")
			_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
			Expect(err).NotTo(HaveOccurred())

			By("Validating the finalizer is added")
			server := &networkv1alpha1.Server{}
			_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
			Expect(err).To(BeNil())
			Expect(k8sClient.Get(ctx, namespacedName, server)).To(Succeed())
			Expect(controllerutil.ContainsFinalizer(server, SERVER_FINALIZER)).To(BeTrue())
		})

		It("should create a Pod for the Server", func() {
			reconciler := &ServerReconciler{
				Client:          k8sClient,
				Scheme:          k8sClient.Scheme(),
				DeletionAllowed: checker,
				Recorder:        NewFakeRecorder(),
			}
			By("Reconciling the resource")
			_, err := reconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: namespacedName})
			Expect(err).NotTo(HaveOccurred())
			_, err = reconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: namespacedName})
			Expect(err).NotTo(HaveOccurred())
			By("Validating a Pod is created")
			pod := &corev1.Pod{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      ServerName + "-pod",
				Namespace: ServerNamespace,
			}, pod)).To(Succeed())
		})

		It("should handle deletion of the Server and remove the Pod", func() {
			reconciler := &ServerReconciler{
				Client:          k8sClient,
				Scheme:          k8sClient.Scheme(),
				DeletionAllowed: checker,
				Recorder:        NewFakeRecorder(),
			}
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

		It("Should return error on get fail", func() {
			fakeClient := FakeFailClient{
				client:     k8sClient,
				FailUpdate: false,
				FailCreate: false,
				FailDelete: false,
				FailGet:    true,
				FailList:   false,
				FailPatch:  false,
			}

			reconciler := &ServerReconciler{
				Client:          fakeClient,
				Scheme:          k8sClient.Scheme(),
				DeletionAllowed: checker,
				Recorder:        NewFakeRecorder(),
			}

			_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
			Expect(err).ToNot(BeNil())

		})

		It("Should return error on update fail", func() {
			By("Basic update fail")
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

			reconciler := &ServerReconciler{
				Client:          fakeClient,
				Scheme:          k8sClient.Scheme(),
				DeletionAllowed: checker,
				Recorder:        NewFakeRecorder(),
			}

			_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
			Expect(err).ToNot(BeNil())

			By("Second update fail")
			fakeClient.FailUpdate = false
			reconciler.Client = fakeClient
			_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
			Expect(err).To(BeNil())
			fakeClient.FailUpdate = true
			reconciler.Client = fakeClient

			_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
			Expect(err).To(BeNil())
			_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
			Expect(err).ToNot(BeNil())

			fakeClient.FailUpdate = false
			fakeClient.FailGetOnPod = true
			reconciler.Client = fakeClient

			_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
			Expect(err).ToNot(BeNil())
		})

	})
})
