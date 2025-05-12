package kube

import (
	"context"
	"encoding/json"
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
	ApiVersion APIVersion `json:"apiVersion"`
	Kind       Kind       `json:"kind"`
	Metadata   Metadata   `json:"metadata"`
	Spec       FleetSpec  `json:"spec"`
}

// CreateFleet is used to create a new fleet using the dynamicclient
func CreateFleet(context context.Context, fleet *Fleet, client *dynamic.DynamicClient) error {
	resource := client.Resource(FleetGCR).Namespace(fleet.Metadata.Namespace)
	fleetStruct, err := fleetToUnstructured(fleet)
	if err != nil {
		return err
	}
	_, err = resource.Create(context, fleetStruct, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

// DeleteFleet is used to delete a fleet. It matches the Fleet using Metadata.Name and Metadata.Namespace. If force is true, it will force delete without waiting for server to allow it.
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

// forceDeleteFleet is used to find all the servers related to a fleet and send them a deleteallow request.
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
		err := sendDeleteAllowed(ctx, server.GetName(), server.GetNamespace(), clientset)
		if err != nil {
			return err
		}
	}
	return nil
}

// fleetToUnstructured is used to make a Fleet object into an unstructured object which can interact with dynamic client
func fleetToUnstructured(fleet *Fleet) (*unstructured.Unstructured, error) {
	fleet.ApiVersion = crdGroup + "/" + crdVersion
	fleet.Kind = fleetResourceName
	bytes, err := json.Marshal(fleet)
	if err != nil {
		return nil, err
	}

	obj := &unstructured.Unstructured{}
	if err := obj.UnmarshalJSON(bytes); err != nil {
		return nil, err
	}
	return obj, nil
}
