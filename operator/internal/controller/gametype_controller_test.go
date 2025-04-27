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
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"
)

var basicGametypeSpec = networkv1alpha1.GameTypeSpec{
	FleetSpec: basicFleetSpec,
}
var _ = Describe("GameType Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-gametype"
		const namespace = "default"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: namespace,
		}

		BeforeEach(func() {
			gametype := &networkv1alpha1.GameType{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: namespace,
				},
				Spec: basicGametypeSpec,
			}
			By("Making sure server does not exist")
			err := k8sClient.Get(ctx, typeNamespacedName, gametype)
			if err == nil {
				return
			}
			if !errors.IsNotFound(err) {
				Expect(err).To(Succeed())
			}
			By("creating the custom resource for the Kind GameType")
			err = k8sClient.Create(ctx, gametype)
			Expect(err).To(BeNil())

			Eventually(func() error {
				var gt networkv1alpha1.GameType
				err := k8sClient.Get(ctx, typeNamespacedName, &gt)
				return err
			}, time.Second*10, time.Millisecond*500).Should(Succeed())
		})

		AfterEach(func() {
			var gametype networkv1alpha1.GameType
			err := k8sClient.Get(ctx, typeNamespacedName, &gametype)
			if err != nil && errors.IsNotFound(err) {
				return
			}
			Expect(err).NotTo(HaveOccurred())

			reconciler := &GameTypeReconciler{
				Client:   k8sClient,
				Scheme:   k8sClient.Scheme(),
				Recorder: NewFakeRecorder(),
			}
			_, err = reconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			if err != nil && errors.IsNotFound(err) {
				return
			}
			Expect(err).To(BeNil())

			By("Cleanup the specific resource instance GameType")
			err = k8sClient.Get(ctx, typeNamespacedName, &gametype)
			if err != nil && errors.IsNotFound(err) {
				return
			}
			Expect(err).NotTo(HaveOccurred())
			Expect(k8sClient.Delete(ctx, &gametype)).To(Succeed())
			_, err = reconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).To(BeNil())

			Eventually(func() error {
				var gt networkv1alpha1.GameType
				err := k8sClient.Get(ctx, typeNamespacedName, &gt)
				if err != nil && errors.IsNotFound(err) {
					return nil
				}
				if err != nil {
					return err
				}

				return fmt.Errorf("gametype still exists")
			}, time.Second*10, time.Millisecond*500).Should(Succeed())
		})

		It("Has correct finalizer", func() {
			reconciler := &GameTypeReconciler{
				Client:   k8sClient,
				Scheme:   k8sClient.Scheme(),
				Recorder: NewFakeRecorder(),
			}
			_, err := reconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).To(BeNil())

			gametype := &networkv1alpha1.GameType{}
			err = k8sClient.Get(ctx, typeNamespacedName, gametype)
			Expect(err).To(BeNil())

			hasFinalizer := controllerutil.ContainsFinalizer(gametype, TypeFinalizer)
			Expect(hasFinalizer).To(BeTrue())
		})

		It("Fails on get", func() {
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
			reconciler := &GameTypeReconciler{
				Client:   fakeClient,
				Scheme:   k8sClient.Scheme(),
				Recorder: NewFakeRecorder(),
			}
			By("Fail get")
			_, err := reconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).To(Not(BeNil()))
		})

		It("Fails on update", func() {
			By("Create failing client")
			fakeClient := FakeFailClient{
				client:     k8sClient,
				FailUpdate: true,
				FailCreate: false,
				FailDelete: false,
				FailGet:    false,
				FailList:   false,
				FailPatch:  false,
			}
			reconciler := &GameTypeReconciler{
				Client:   fakeClient,
				Scheme:   k8sClient.Scheme(),
				Recorder: NewFakeRecorder(),
			}
			By("Fail update")
			_, err := reconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).To(Not(BeNil()))
		})

		It("Fails on delete", func() {
			By("Create failing client")
			fakeClient := FakeFailClient{
				client:     k8sClient,
				FailUpdate: false,
				FailCreate: false,
				FailDelete: false,
				FailGet:    false,
				FailList:   false,
				FailPatch:  false,
			}
			reconciler := &GameTypeReconciler{
				Client:   fakeClient,
				Scheme:   k8sClient.Scheme(),
				Recorder: NewFakeRecorder(),
			}
			By("initial setup")
			_, err := reconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).To(BeNil())
			_, err = reconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).To(BeNil())
			By("Get gametype")
			gametype := &networkv1alpha1.GameType{}
			err = k8sClient.Get(ctx, typeNamespacedName, gametype)
			Expect(err).To(BeNil())
			By("Try to delete gametype")
			err = k8sClient.Delete(ctx, gametype)
			Expect(err).To(BeNil())
			By("Fail delete")
			fakeClient.FailDelete = true
			reconciler.Client = fakeClient
			_, err = reconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).To(Not(BeNil()))
		})

		It("Fails on create", func() {
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
			reconciler := &GameTypeReconciler{
				Client:   fakeClient,
				Scheme:   k8sClient.Scheme(),
				Recorder: NewFakeRecorder(),
			}
			By("Fail create")
			_, err := reconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).To(BeNil())
			_, err = reconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).To(Not(BeNil()))
		})

		It("Should create a fleet with the correct label", func() {
			By("Setup reconciler")
			reconciler := &GameTypeReconciler{
				Client:   k8sClient,
				Scheme:   k8sClient.Scheme(),
				Recorder: NewFakeRecorder(),
			}

			By("Trigger initial reconciliations")
			_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: typeNamespacedName})
			Expect(err).To(BeNil())
			_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: typeNamespacedName})
			Expect(err).To(BeNil())

			By("Expect one fleet created")
			var fleetList networkv1alpha1.FleetList
			err = k8sClient.List(ctx, &fleetList, kclient.MatchingLabels{"type": resourceName})
			Expect(err).To(BeNil())
			Expect(fleetList.Items).To(HaveLen(1))
		})

		It("Replaces the fleet when FleetSpec changes", func() {
			By("Initial reconciliation")
			reconciler := &GameTypeReconciler{
				Client:   k8sClient,
				Scheme:   k8sClient.Scheme(),
				Recorder: NewFakeRecorder(),
			}
			_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: typeNamespacedName})
			Expect(err).To(BeNil())
			_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: typeNamespacedName})
			Expect(err).To(BeNil())

			var initialFleets networkv1alpha1.FleetList
			err = k8sClient.List(ctx, &initialFleets, kclient.MatchingLabels{"type": resourceName})
			Expect(err).To(BeNil())
			Expect(initialFleets.Items).To(HaveLen(1))
			oldFleetName := initialFleets.Items[0].Name

			By("Update the GameType spec to force replacement")
			var gt networkv1alpha1.GameType
			Expect(k8sClient.Get(ctx, typeNamespacedName, &gt)).To(Succeed())
			gt.Spec.FleetSpec.ServerSpec.Pod.Containers[0].Image = "changed-image"
			Expect(k8sClient.Update(ctx, &gt)).To(Succeed())

			By("Reconcile again to trigger replacement")
			_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: typeNamespacedName})
			Expect(err).To(BeNil())

			var updatedFleets networkv1alpha1.FleetList
			err = k8sClient.List(ctx, &updatedFleets, kclient.MatchingLabels{"type": resourceName})
			Expect(err).To(BeNil())
			Expect(updatedFleets.Items).To(HaveLen(2))

			names := []string{updatedFleets.Items[0].Name, updatedFleets.Items[1].Name}
			Expect(names).To(ContainElement(oldFleetName))
		})

		It("Updates the replica count when changed", func() {
			By("Initial reconciliation")
			reconciler := &GameTypeReconciler{
				Client:   k8sClient,
				Scheme:   k8sClient.Scheme(),
				Recorder: NewFakeRecorder(),
			}
			_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: typeNamespacedName})
			Expect(err).To(BeNil())
			_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: typeNamespacedName})
			Expect(err).To(BeNil())

			var gt networkv1alpha1.GameType
			Expect(k8sClient.Get(ctx, typeNamespacedName, &gt)).To(Succeed())
			gt.Spec.FleetSpec.Scaling.Replicas = 5
			Expect(k8sClient.Update(ctx, &gt)).To(Succeed())

			By("Trigger reconcile to update replicas")
			_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: typeNamespacedName})
			Expect(err).To(BeNil())

			var fleetList networkv1alpha1.FleetList
			Expect(k8sClient.List(ctx, &fleetList, kclient.MatchingLabels{"type": resourceName})).To(Succeed())
			Expect(fleetList.Items).To(HaveLen(1))
			Expect(fleetList.Items[0].Spec.Scaling.Replicas).To(Equal(int32(5)))
		})

		It("Deletes the oldest fleet if multiple fleets exist", func() {
			By("Initial reconciliation")
			reconciler := &GameTypeReconciler{
				Client:   k8sClient,
				Scheme:   k8sClient.Scheme(),
				Recorder: NewFakeRecorder(),
			}
			_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: typeNamespacedName})
			Expect(err).To(BeNil())
			_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: typeNamespacedName})
			Expect(err).To(BeNil())

			var gt networkv1alpha1.GameType
			Expect(k8sClient.Get(ctx, typeNamespacedName, &gt)).To(Succeed())

			By("Force fleet spec change to trigger new fleet creation")
			gt.Spec.FleetSpec.ServerSpec.Pod.Containers[0].Image = "another-image"
			Expect(k8sClient.Update(ctx, &gt)).To(Succeed())

			_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: typeNamespacedName})
			Expect(err).To(BeNil())

			var fleetList networkv1alpha1.FleetList
			Expect(k8sClient.List(ctx, &fleetList, kclient.MatchingLabels{"type": resourceName})).To(Succeed())
			Expect(len(fleetList.Items)).To(BeNumerically(">", 1))

			By("Trigger cleanup of oldest fleet")
			_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: typeNamespacedName})
			Expect(err).To(BeNil())

			Eventually(func() int {
				_ = k8sClient.List(ctx, &fleetList, kclient.MatchingLabels{"type": resourceName})
				return len(fleetList.Items)
			}, time.Second*5, time.Millisecond*500).Should(BeNumerically("<=", 2))
		})

		It("Should emit the correct events", func() {
			recorder := NewFakeRecorder()
			fakeClient := FakeFailClient{
				client:       k8sClient,
				FailUpdate:   false,
				FailCreate:   false,
				FailDelete:   false,
				FailGet:      false,
				FailList:     false,
				FailPatch:    false,
				FailGetOnPod: false,
			}
			reconciler := &GameTypeReconciler{
				Client:   fakeClient,
				Scheme:   fakeClient.Scheme(),
				Recorder: recorder,
			}

			By("Initial reconciliations")
			_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: typeNamespacedName})
			Expect(err).To(BeNil())
			_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: typeNamespacedName})
			Expect(err).To(BeNil())
			_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: typeNamespacedName})
			Expect(err).To(BeNil())

			var gt networkv1alpha1.GameType
			Expect(k8sClient.Get(ctx, typeNamespacedName, &gt)).To(Succeed())

			hasFinalizerAddingEvent := false
			hasInitialFleetEvent := false

			for _, event := range recorder.Events {
				if event.Message == "Created initial fleet" {
					hasInitialFleetEvent = true
				}
				if event.Message == "Added finalizers to game" {
					hasFinalizerAddingEvent = true
				}
			}

			Expect(hasFinalizerAddingEvent).To(BeTrue())
			Expect(hasInitialFleetEvent).To(BeTrue())

			By("Check if scaling event is emitted")
			gt.Spec.FleetSpec.Scaling.Replicas = gt.Spec.FleetSpec.Scaling.Replicas + 1
			Expect(k8sClient.Update(ctx, &gt)).To(Succeed())
			_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: typeNamespacedName})
			Expect(err).To(BeNil())
			_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: typeNamespacedName})
			Expect(err).To(BeNil())
			_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: typeNamespacedName})
			Expect(err).To(BeNil())
			Expect(k8sClient.Get(ctx, typeNamespacedName, &gt)).To(Succeed())

			hasScalingEvent := false
			requiredMsg := fmt.Sprintf("Scaling gametype to %d", gt.Spec.FleetSpec.Scaling.Replicas)
			for _, event := range recorder.Events {
				if event.Message == requiredMsg {
					hasScalingEvent = true
					break
				}
			}
			Expect(hasScalingEvent).To(BeTrue())

			By("Check if new fleet event is emitted")
			gt.Spec.FleetSpec.ServerSpec.Pod.Containers[0].Image = "another-image"
			Expect(k8sClient.Update(ctx, &gt)).To(Succeed())
			_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: typeNamespacedName})
			Expect(err).To(BeNil())
			hasNewFleetEvent := false
			for _, event := range recorder.Events {
				if event.Message == "Creating new fleet" {
					hasNewFleetEvent = true
					break
				}
			}
			Expect(hasNewFleetEvent).To(BeTrue())

			By("Check if old fleet was deleted")
			_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: typeNamespacedName})
			Expect(err).To(BeNil())
			hasOldFleetDeleteEvent := false
			for _, event := range recorder.Events {
				if event.Message == "Deleting extra fleet" {
					hasOldFleetDeleteEvent = true
					break
				}
			}
			Expect(hasOldFleetDeleteEvent).To(BeTrue())

			By("Delete gametype")
			Expect(k8sClient.Delete(ctx, &gt)).To(Succeed())
			_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: typeNamespacedName})
			Expect(err).To(BeNil())

			By("Check if deletion has correct events")
			hasFinalizersRemovedEvent := false
			for _, event := range recorder.Events {
				if event.Message == "Removed finalizer" {
					hasFinalizersRemovedEvent = true
					break
				}
			}
			Expect(hasFinalizersRemovedEvent).To(BeTrue())
		})

	})
})
