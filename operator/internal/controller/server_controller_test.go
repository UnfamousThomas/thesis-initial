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
	"strings"
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
			fakeRecorder := NewFakeRecorder()
			checker := TestChecker{
				deleteAllowed: make(map[string]bool),
			}
			reconciler := &ServerReconciler{
				Client:            k8sClient,
				Scheme:            k8sClient.Scheme(),
				ErrorOnNotAllowed: true,
				DeletionAllowed:   checker,
				Recorder:          fakeRecorder,
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

				foundNotAllowed := false
				for _, event := range fakeRecorder.Events {
					if event.Message == "Server did not respond with allowed" {
						foundNotAllowed = true
						break
					}
				}
				Expect(foundNotAllowed).To(BeTrue())

				checker.deleteAllowed[server.Name] = true
				_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
				Expect(err).To(BeNil())

				By("Checking if finalizer event is logged")
				containsEvent := false
				for _, event := range fakeRecorder.Events {
					if event.Message == "Finalizer removed" {
						containsEvent = true
					}
				}

				Expect(containsEvent).To(BeTrue())

				// Wait until the resource is deleted
				Eventually(func() bool {
					err := k8sClient.Get(ctx, namespacedName, server)
					return errors.IsNotFound(err)
				}).Should(BeTrue())
			}
		})

		It("should add a finalizer if not present", func() {
			recorder := NewFakeRecorder()
			checker := TestChecker{
				deleteAllowed: make(map[string]bool),
			}
			reconciler := &ServerReconciler{
				Client:            k8sClient,
				Scheme:            k8sClient.Scheme(),
				ErrorOnNotAllowed: true,
				DeletionAllowed:   checker,
				Recorder:          recorder,
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

			By("Checking if finalizer event was emitted")
			containsEvent := false
			for _, event := range recorder.Events {
				if event.Message == "Finalizer added" {
					containsEvent = true
					break
				}
			}
			Expect(containsEvent).To(BeTrue())
		})

		It("should create a Pod for the Server", func() {
			recorder := NewFakeRecorder()
			checker := TestChecker{
				deleteAllowed: make(map[string]bool),
			}
			reconciler := &ServerReconciler{
				Client:            k8sClient,
				Scheme:            k8sClient.Scheme(),
				ErrorOnNotAllowed: true,
				DeletionAllowed:   checker,
				Recorder:          recorder,
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

			By("Validating that events are correct")
			_, err = reconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: namespacedName})
			Expect(err).NotTo(HaveOccurred())
			containsPodCreated := false
			containsPodFinalizer := false
			containsServerFinalizer := false
			for _, event := range recorder.Events {
				GinkgoLogr.Info(event.Message)
				if event.Message == "Pod created successfully" {
					containsPodCreated = true
				}
				_, podOk := event.Object.(*corev1.Pod)

				if podOk && event.Message == "Pod finalizer added" {
					containsPodFinalizer = true
				}

				_, serverOk := event.Object.(*networkv1alpha1.Server)
				if serverOk && event.Message == "Pod finalizer added" {
					containsServerFinalizer = true
				}
			}

			Expect(containsPodCreated).To(BeTrue())
			Expect(containsPodFinalizer).To(BeTrue())
			Expect(containsServerFinalizer).To(BeTrue())

		})

		It("should handle deletion of the Server and remove the Pod", func() {
			fakeRecorder := NewFakeRecorder()
			checker := TestChecker{
				deleteAllowed: make(map[string]bool),
			}
			reconciler := &ServerReconciler{
				Client:            k8sClient,
				Scheme:            k8sClient.Scheme(),
				ErrorOnNotAllowed: true,
				DeletionAllowed:   checker,
				Recorder:          fakeRecorder,
			}
			By("Reconcile the basic server")
			_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
			Expect(err).NotTo(HaveOccurred())
			_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
			Expect(err).NotTo(HaveOccurred())
			_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
			Expect(err).NotTo(HaveOccurred())

			//Make sure pod exists
			pod := &corev1.Pod{}
			err = k8sClient.Get(ctx, types.NamespacedName{
				Name:      ServerName + "-pod",
				Namespace: ServerNamespace,
			}, pod)
			Expect(err).NotTo(HaveOccurred())
			By("Deleting the Server resource")
			server := &networkv1alpha1.Server{}
			Expect(k8sClient.Get(ctx, namespacedName, server)).To(Succeed())
			Expect(k8sClient.Delete(ctx, server)).To(Succeed())

			By("Reconciling the deletion")
			_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
			Expect(err).To(HaveOccurred())

			By("Reconciling the deletion with allowed")
			checker.deleteAllowed[server.Name] = true
			_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
			Expect(err).NotTo(HaveOccurred())

			By("Validating the Pod is deleted")
			pod = &corev1.Pod{}
			err = k8sClient.Get(ctx, types.NamespacedName{
				Name:      ServerName + "-pod",
				Namespace: ServerNamespace,
			}, pod)
			Expect(errors.IsNotFound(err)).To(BeTrue())

			GinkgoLogr.Info("Events in recorder", "amount", len(fakeRecorder.Events), "events", fakeRecorder.Events)
			for _, event := range fakeRecorder.Events {
				GinkgoLogr.Info("Event", "message", event.Message)
			}

			By("Check if has pod deletion event")
			hasEvent := false
			for _, event := range fakeRecorder.Events {
				if event.Message == "Pod successfully deleted during finalization" {
					hasEvent = true
					break
				}
			}
			Expect(hasEvent).To(BeTrue())

			By("Validating the finalizer is removed")
			err = k8sClient.Get(ctx, namespacedName, server)
			Expect(errors.IsNotFound(err)).To(BeTrue())

			By("Check if server has finalizer removal event")
			hasGlobalFinalizerRemoved := false
			hasPodFinalizerRemoved := false
			hasServerFinalizerRemoved := false
			for _, event := range fakeRecorder.Events {
				if event.Message == "Finalizer removed" {
					hasGlobalFinalizerRemoved = true
				}
				_, isPod := event.Object.(*corev1.Pod)
				if isPod && event.Message == "Pod finalizer removed" {
					hasPodFinalizerRemoved = true
				}

				_, isServer := event.Object.(*networkv1alpha1.Server)
				if isServer && event.Message == "Pod finalizer removed" {
					hasServerFinalizerRemoved = true
				}

			}
			Expect(hasPodFinalizerRemoved).To(BeTrue())
			Expect(hasServerFinalizerRemoved).To(BeTrue())
			Expect(hasGlobalFinalizerRemoved).To(BeTrue())
		})

		It("Should return error on get fail", func() {
			checker := TestChecker{
				deleteAllowed: make(map[string]bool),
			}
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
				Client:            fakeClient,
				Scheme:            k8sClient.Scheme(),
				ErrorOnNotAllowed: true,
				DeletionAllowed:   checker,
				Recorder:          NewFakeRecorder(),
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
			checker := TestChecker{
				deleteAllowed: make(map[string]bool),
			}
			recorder := NewFakeRecorder()
			reconciler := &ServerReconciler{
				Client:            fakeClient,
				Scheme:            k8sClient.Scheme(),
				ErrorOnNotAllowed: true,
				DeletionAllowed:   checker,
				Recorder:          recorder,
			}

			_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
			Expect(err).ToNot(BeNil())
			failUpdateEvent := false
			for _, event := range recorder.Events {
				if strings.HasPrefix(event.Message, "failed to update server") {
					failUpdateEvent = true
					break
				}
			}
			Expect(failUpdateEvent).To(BeTrue())
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
