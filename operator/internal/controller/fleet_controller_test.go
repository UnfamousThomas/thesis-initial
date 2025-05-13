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
	networkv1alpha1 "github.com/unfamousthomas/thesis-operator/api/v1alpha1"
	"github.com/unfamousthomas/thesis-operator/internal/utils"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"
)

var prodChecker = utils.ProdDeletionChecker{}
var basicFleetSpec = networkv1alpha1.FleetSpec{
	Scaling: networkv1alpha1.FleetScaling{
		Replicas:          2,
		PrioritizeAllowed: false,
		AgePriority:       networkv1alpha1.OldestFirst,
	},
	ServerSpec: basicServerSpec,
}

var _ = Describe("Fleet Controller", func() {
	Context("When reconciling a resource", func() {
		const (
			FleetName      = "test-fleet"
			FleetNamespace = "default"
		)

		var (
			ctx            context.Context
			namespacedName types.NamespacedName
		)

		BeforeEach(func() {
			ctx = context.Background()
			namespacedName = types.NamespacedName{
				Name:      FleetName,
				Namespace: FleetNamespace,
			}

			By("Creating a Fleet resource")
			fleet := &networkv1alpha1.Fleet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      FleetName,
					Namespace: FleetNamespace,
				},
				Spec: basicFleetSpec,
			}
			Expect(k8sClient.Create(ctx, fleet)).To(Succeed())
		})

		AfterEach(func() {
			fleet := &networkv1alpha1.Fleet{}
			err := k8sClient.Get(ctx, namespacedName, fleet)
			if err != nil && errors.IsNotFound(err) {
				return
			}
			Expect(err).To(Succeed())
			Expect(k8sClient.Delete(ctx, fleet)).To(Succeed())
			recorder := NewFakeRecorder()
			reconciler := &FleetReconciler{
				Client:          k8sClient,
				Scheme:          k8sClient.Scheme(),
				Recorder:        recorder,
				DeletionChecker: prodChecker,
			}
			_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
			Expect(err).ToNot(HaveOccurred())
			Expect(err).To(Succeed())
			Eventually(func() bool {
				err := k8sClient.Get(ctx, namespacedName, fleet)
				return errors.IsNotFound(err)
			}, time.Second*10, time.Millisecond*500).Should(BeTrue())

			hasFinalizerRemoveEvent := false
			for _, event := range recorder.Events {
				if event.Message == "Fleet finalizers removed" {
					hasFinalizerRemoveEvent = true
				}
				break
			}
			Expect(hasFinalizerRemoveEvent).To(BeTrue())
		})

		It("Should emit the correct events", func() {
			By("Setting up reconciler")
			recorder := NewFakeRecorder()
			reconciler := &FleetReconciler{
				Client:          k8sClient,
				Scheme:          k8sClient.Scheme(),
				Recorder:        recorder,
				DeletionChecker: prodChecker,
			}

			By("Initial reconcile")
			_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
			Expect(err).ToNot(HaveOccurred())

			By("Checking for finalizer event")
			hasFleetFinalizersEvent := false

			for _, event := range recorder.Events {
				if event.Message == "Fleet finalizers added" {
					hasFleetFinalizersEvent = true
					break
				}
			}
			Expect(hasFleetFinalizersEvent).To(BeTrue())

			By("Checking for scaling up event")
			_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
			Expect(err).ToNot(HaveOccurred())
			hasScaleEvent := false
			for _, event := range recorder.Events {
				required := fmt.Sprintf("Scaled servers up to %d", basicFleetSpec.Scaling.Replicas)
				if event.Message == required {
					hasScaleEvent = true
					break
				}
			}
			Expect(hasScaleEvent).To(BeTrue())

			By("Updating replica count")
			var fleet networkv1alpha1.Fleet
			err = k8sClient.Get(ctx, namespacedName, &fleet)
			Expect(err).ToNot(HaveOccurred())
			fleet.Spec.Scaling.Replicas = fleet.Spec.Scaling.Replicas - 1
			err = k8sClient.Update(ctx, &fleet)
			Expect(err).ToNot(HaveOccurred())
			_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
			Expect(err).ToNot(HaveOccurred())

			By("Checking for scaling down evenet")
			hasScaleEvent = false
			for _, event := range recorder.Events {
				required := fmt.Sprintf("Scaled servers down to %d", fleet.Spec.Scaling.Replicas)
				if event.Message == required {
					hasScaleEvent = true
					break
				}
			}
			Expect(hasScaleEvent).To(BeTrue())
		})
		It("should delete all servers and remove the finalizer on fleet deletion", func() {
			reconciler := &FleetReconciler{
				Client:          k8sClient,
				Scheme:          k8sClient.Scheme(),
				Recorder:        NewFakeRecorder(),
				DeletionChecker: prodChecker,
			}

			// Initial reconciles to create servers
			_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
			Expect(err).NotTo(HaveOccurred())
			_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() int {
				serverList := &networkv1alpha1.ServerList{}
				_ = k8sClient.List(ctx, serverList)
				return len(serverList.Items)
			}, time.Second*10, time.Millisecond*500).Should(Equal(int(basicFleetSpec.Scaling.Replicas)))

			// Delete the fleet
			fleet := &networkv1alpha1.Fleet{}
			Expect(k8sClient.Get(ctx, namespacedName, fleet)).To(Succeed())
			Expect(k8sClient.Delete(ctx, fleet)).To(Succeed())

			// Reconcile deletion (to handle cleanup + finalizer logic)
			Eventually(func() bool {
				_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
				return err == nil
			}, time.Second*5, time.Millisecond*200).Should(BeTrue())

			// Finalizer should eventually be removed after cleanup
			Eventually(func() bool {
				err := k8sClient.Get(ctx, namespacedName, fleet)
				if errors.IsNotFound(err) {
					return true // Finalizer removed and object is gone
				}
				if err != nil {
					return false
				}
				return !controllerutil.ContainsFinalizer(fleet, FLEET_FINALIZER)
			}, time.Second*10, time.Millisecond*500).Should(BeTrue())
		})

		It("should add a finalizer if not present", func() {
			reconciler := &FleetReconciler{
				Client:          k8sClient,
				Scheme:          k8sClient.Scheme(),
				Recorder:        NewFakeRecorder(),
				DeletionChecker: prodChecker,
			}

			_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
			Expect(err).NotTo(HaveOccurred())

			fleet := &networkv1alpha1.Fleet{}
			Expect(k8sClient.Get(ctx, namespacedName, fleet)).To(Succeed())
			Expect(controllerutil.ContainsFinalizer(fleet, FLEET_FINALIZER)).To(BeTrue())
		})

		It("should scale up servers to match the desired replicas", func() {
			reconciler := &FleetReconciler{
				Client:          k8sClient,
				Scheme:          k8sClient.Scheme(),
				Recorder:        NewFakeRecorder(),
				DeletionChecker: prodChecker,
			}

			_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
			Expect(err).NotTo(HaveOccurred())
			_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() int {
				serverList := &networkv1alpha1.ServerList{}
				err := k8sClient.List(ctx, serverList)
				if err != nil {
					return -1
				}
				return len(serverList.Items)
			}, time.Second*10, time.Millisecond*500).Should(Equal(int(basicFleetSpec.Scaling.Replicas)))

			lowerReplicas := basicFleetSpec.Scaling.Replicas - 1
			var fleet networkv1alpha1.Fleet
			err = k8sClient.Get(ctx, namespacedName, &fleet)
			Expect(err).NotTo(HaveOccurred())
			fleet.Spec.Scaling.Replicas = lowerReplicas
			err = k8sClient.Update(ctx, &fleet)
			Expect(err).NotTo(HaveOccurred())

			_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() int {
				serverList := &networkv1alpha1.ServerList{}
				err := k8sClient.List(ctx, serverList)
				if err != nil {
					return -1
				}
				return len(serverList.Items)
			}, time.Second*10, time.Millisecond*500).Should(Equal(int(lowerReplicas)))

		})

		It("should delete all servers when fleet is deleted", func() {
			reconciler := &FleetReconciler{
				Client:          k8sClient,
				Scheme:          k8sClient.Scheme(),
				Recorder:        NewFakeRecorder(),
				DeletionChecker: prodChecker,
			}

			// Ensure the fleet is scaled first
			_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
			Expect(err).NotTo(HaveOccurred())
			_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() int {
				serverList := &networkv1alpha1.ServerList{}
				_ = k8sClient.List(ctx, serverList)
				return len(serverList.Items)
			}, time.Second*10, time.Millisecond*500).Should(Equal(int(basicFleetSpec.Scaling.Replicas)))

			// Delete the fleet
			fleet := &networkv1alpha1.Fleet{}
			Expect(k8sClient.Get(ctx, namespacedName, fleet)).To(Succeed())
			Expect(k8sClient.Delete(ctx, fleet)).To(Succeed())

			// Reconcile deletion
			_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() bool {
				serverList := &networkv1alpha1.ServerList{}
				_ = k8sClient.List(ctx, serverList)
				return len(serverList.Items) == 0
			}, time.Second*10, time.Millisecond*500).Should(BeTrue())
		})

		It("Should return errors based on failures", func() {
			By("Create failing client")
			fakeClient := FakeFailClient{
				client:     k8sClient,
				FailUpdate: false,
				FailCreate: false,
				FailDelete: false,
				FailGet:    true,
				FailList:   false,
				FailPatch:  false,
			}
			reconciler := &FleetReconciler{
				Client:          fakeClient,
				Scheme:          k8sClient.Scheme(),
				Recorder:        NewFakeRecorder(),
				DeletionChecker: prodChecker,
			}

			By("Fail Get")
			_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
			Expect(err).To(HaveOccurred())
			fakeClient.FailGet = false
			fakeClient.FailUpdate = true
			reconciler.Client = fakeClient

			By("Fail Update")
			_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
			Expect(err).To(HaveOccurred())

			By("Fail list")
			fakeClient.FailUpdate = false
			fakeClient.FailList = true
			reconciler.Client = fakeClient
			_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
			fmt.Println(err)
			Expect(err).To(Not(HaveOccurred()))
			_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
			Expect(err).To(HaveOccurred())

		})

		It("Fail create and delete", func() {
			By("Create failing client")
			fakeClient := FakeFailClient{
				client:     k8sClient,
				FailUpdate: false,
				FailCreate: true,
				FailDelete: false,
				FailGet:    false,
				FailList:   false,
				FailPatch:  false,
			}
			reconciler := &FleetReconciler{
				Client:          fakeClient,
				Scheme:          k8sClient.Scheme(),
				Recorder:        NewFakeRecorder(),
				DeletionChecker: prodChecker,
			}

			By("Fail Create")
			_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
			Expect(err).To(Not(HaveOccurred()))
			_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
			Expect(err).To(HaveOccurred())

			By("Succeed in create")
			fakeClient.FailCreate = false
			reconciler.Client = fakeClient
			_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
			Expect(err).To(Not(HaveOccurred()))

			By("Fail Delete")
			var fleet networkv1alpha1.Fleet
			err = fakeClient.Get(ctx, namespacedName, &fleet)
			Expect(err).To(Not(HaveOccurred()))
			fleet.Spec.Scaling.Replicas = fleet.Spec.Scaling.Replicas - 1
			err = fakeClient.Update(ctx, &fleet)
			fakeClient.FailDelete = true
			reconciler.Client = fakeClient
			_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
			Expect(err).To(HaveOccurred())
		})
	})
})
