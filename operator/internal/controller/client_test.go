package controller

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sync"
)

type FakeFailClient struct {
	client       client.Client
	FailUpdate   bool
	FailCreate   bool
	FailDelete   bool
	FailGet      bool
	FailList     bool
	FailPatch    bool
	FailGetOnPod bool
}

func (client FakeFailClient) Get(ctx context.Context, name types.NamespacedName, obj client.Object, opts ...client.GetOption) error {
	if client.FailGet {
		return fmt.Errorf("fail get")
	}
	if client.FailGetOnPod {
		if isPod(obj) {
			return fmt.Errorf("fail get on pod")
		}
	}
	return client.client.Get(ctx, name, obj, opts...)
}

func (client FakeFailClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	if client.FailCreate {
		return fmt.Errorf("fail create")
	}
	return client.client.Create(ctx, obj, opts...)
}

func (client FakeFailClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	if client.FailUpdate {
		return fmt.Errorf("fail update")
	}
	return client.client.Update(ctx, obj, opts...)
}

func (client FakeFailClient) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	if client.FailDelete {
		return fmt.Errorf("fail delete")
	}
	return client.client.Delete(ctx, obj, opts...)
}

func (client FakeFailClient) DeleteAllOf(ctx context.Context, obj client.Object, opts ...client.DeleteAllOfOption) error {
	if client.FailDelete {
		return fmt.Errorf("fail deleteAllOf")
	}
	return client.client.DeleteAllOf(ctx, obj, opts...)
}
func (client FakeFailClient) List(ctx context.Context, obj client.ObjectList, opts ...client.ListOption) error {
	if client.FailList {
		return fmt.Errorf("fail list")
	}
	return client.client.List(ctx, obj, opts...)
}

func (client FakeFailClient) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	if client.FailPatch {
		return fmt.Errorf("fail patch")
	}
	return client.client.Patch(ctx, obj, patch, opts...)
}

func (client FakeFailClient) Status() client.SubResourceWriter {
	return client.client.Status()
}

func (client FakeFailClient) SubResource(subResource string) client.SubResourceClient {
	return client.client.SubResource(subResource)
}

func (client FakeFailClient) GroupVersionKindFor(obj runtime.Object) (schema.GroupVersionKind, error) {
	return client.client.GroupVersionKindFor(obj)
}

func (client FakeFailClient) Scheme() *runtime.Scheme {
	return client.client.Scheme()
}
func (client FakeFailClient) RESTMapper() meta.RESTMapper {
	return client.client.RESTMapper()
}
func (client FakeFailClient) IsObjectNamespaced(obj runtime.Object) (bool, error) {
	return client.client.IsObjectNamespaced(obj)
}

func isPod(obj client.Object) bool {
	_, ok := obj.(*corev1.Pod)
	return ok
}

type FakeEvent struct {
	Object      runtime.Object
	EventType   string
	Reason      string
	Message     string
	Annotations map[string]string
}
type FakeRecorder struct {
	mu     sync.Mutex
	Events []FakeEvent
}

func NewFakeRecorder() *FakeRecorder {
	return &FakeRecorder{}
}

func (f *FakeRecorder) Event(object runtime.Object, eventtype, reason, message string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Events = append(f.Events, FakeEvent{
		Object:    object,
		EventType: eventtype,
		Reason:    reason,
		Message:   message,
	})
}

func (f *FakeRecorder) Eventf(object runtime.Object, eventtype, reason, messageFmt string, args ...interface{}) {
	f.Event(object, eventtype, reason, fmt.Sprintf(messageFmt, args...))
}

func (f *FakeRecorder) AnnotatedEventf(object runtime.Object, annotations map[string]string, eventtype, reason, messageFmt string, args ...interface{}) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.Events = append(f.Events, FakeEvent{
		Object:      object,
		EventType:   eventtype,
		Reason:      reason,
		Message:     fmt.Sprintf(messageFmt, args...),
		Annotations: annotations,
	})
}
