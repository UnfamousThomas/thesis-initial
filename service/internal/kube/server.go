package kube

import (
	"context"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"maps"
)

var ServerGCR = schema.GroupVersionResource{
	Group:    crdGroup,
	Version:  crdVersion,
	Resource: serverResourceName,
}

type ServerSpec struct {
	Pod              v1.PodSpec       `json:"pod,omitempty"`
	TimeOut          *metav1.Duration `json:"timeout"`
	AllowForceDelete bool             `json:"allowForceDelete,omitempty"`
}

type Server struct {
	Metadata Metadata   `json:"metadata"`
	Spec     ServerSpec `json:"spec"`
}

func CreateServer(context context.Context, server Server, client *dynamic.DynamicClient) (error, map[string]interface{}) {
	resource := client.Resource(ServerGCR).Namespace(server.Metadata.Namespace)
	serverStruct := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": crdGroup + "/" + crdVersion,
			"kind":       serverResourceName,
			"metadata":   map[string]interface{}{"name": server.Metadata.Name, "namespace": server.Metadata.Namespace},
			"spec":       server.Spec,
		},
	}
	_, err := resource.Create(context, serverStruct, metav1.CreateOptions{})
	if err != nil {
		return err, nil
	}
	return nil, serverStruct.Object
}

func DeleteServer(context context.Context, metadata Metadata, client *dynamic.DynamicClient) error {
	resource := client.Resource(ServerGCR).Namespace(metadata.Namespace)

	err := resource.Delete(context, metadata.Name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}

// AddServerLabels adds label to the server object (NOT POD), note that all labels from metadata are copied.
// Meaning it overwrites.
func AddServerLabels(context context.Context, metadata Metadata, client *dynamic.DynamicClient) error {
	resource := client.Resource(ServerGCR).Namespace(metadata.Namespace)
	unstr, err := resource.Get(context, metadata.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	labels := unstr.GetLabels()
	maps.Copy(labels, metadata.Labels)
	unstr.SetLabels(labels)

	_, err = resource.Update(context, unstr, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func RemoveServerLabel(context context.Context, metadata Metadata, label string, client *dynamic.DynamicClient) error {
	resource := client.Resource(ServerGCR).Namespace(metadata.Namespace)
	unstr, err := resource.Get(context, metadata.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	labels := unstr.GetLabels()
	delete(labels, label)

	unstr.SetLabels(labels)
	_, err = resource.Update(context, unstr, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}
