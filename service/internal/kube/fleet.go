package kube

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

var FleetGCR = schema.GroupVersionResource{
	Group:    crdGroup,
	Version:  crdVersion,
	Resource: fleetResourceName,
}

type FleetSpec struct {
	ServerSpec ServerSpec   `json:"spec"`
	Scaling    FleetScaling `json:"scaling"`
}

type Priority string

var validPriorities = map[Priority]struct{}{
	OldestFirst: {},
	NewestFirst: {},
	// Add new priorities here as needed
}

const (
	OldestFirst Priority = "oldest_first"
	NewestFirst Priority = "newest_first"
)

type FleetScaling struct {
	Replicas          int32    `json:"replicas"`
	PrioritizeAllowed bool     `json:"prioritizeAllowed"`
	AgePriority       Priority `json:"agePriority"`
}

type Fleet struct {
	Metadata Metadata  `json:"metadata"`
	Spec     FleetSpec `json:"spec"`
}

func CreateFleet(context context.Context, fleet Fleet, client *dynamic.DynamicClient) (error, map[string]interface{}) {
	resource := client.Resource(FleetGCR).Namespace(fleet.Metadata.Namespace)
	fleetStruct := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": crdGroup + "/" + crdVersion,
			"kind":       fleetResourceName,
			"metadata":   map[string]interface{}{"name": fleet.Metadata.Name, "namespace": fleet.Metadata.Namespace},
			"spec":       fleet.Spec,
		},
	}
	_, err := resource.Create(context, fleetStruct, metav1.CreateOptions{})
	if err != nil {
		return err, nil
	}
	return nil, fleetStruct.Object
}

func DeleteFleet(ctx context.Context, metadata Metadata, client *dynamic.DynamicClient, clientset *kubernetes.Clientset, force bool) error {
	resource := client.Resource(FleetGCR).Namespace(metadata.Namespace)
	err := resource.Delete(ctx, metadata.Name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	if force {
		err := forceDeleteFleet(ctx, metadata, client, clientset)
		if err != nil {
			return err
		}
	}

	return nil
}

func forceDeleteFleet(ctx context.Context, metadata Metadata, client *dynamic.DynamicClient, clientset *kubernetes.Clientset) error {
	resources, err := client.Resource(FleetGCR).Namespace(metadata.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, server := range resources.Items {
		fleetName, exists := server.GetLabels()["fleet"]
		if !exists || fleetName != metadata.Name {
			continue
		}
		err := sendDeleteAllowed(ctx, server.GetName()+"-pod", clientset)
		if err != nil {
			return err
		}
	}
	return nil
}
