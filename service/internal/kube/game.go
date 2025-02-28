package kube

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

var GameGCR = schema.GroupVersionResource{
	Group:    crdGroup,
	Version:  crdVersion,
	Resource: gameResourceName,
}

type GameTypeSpec struct {
	Scaling   TypeScaling `json:"scaling"`
	FleetSpec FleetSpec   `json:"fleetSpec"`
}

type TypeScaling struct {
	CurrentReplicas int `json:"replicas"`
}

type GameType struct {
	Metadata Metadata     `json:"metadata"`
	Spec     GameTypeSpec `json:"spec"`
}

func CreateGame(context context.Context, game GameType, client *dynamic.DynamicClient) (error, map[string]interface{}) {
	resource := client.Resource(GameGCR).Namespace(game.Metadata.Namespace)
	gameStruct := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": crdGroup + "/" + crdVersion,
			"kind":       gameResourceName,
			"metadata":   map[string]interface{}{"name": game.Metadata.Name, "namespace": game.Metadata.Namespace},
			"spec":       game.Spec,
		},
	}
	_, err := resource.Create(context, gameStruct, metav1.CreateOptions{})
	if err != nil {
		return err, nil
	}
	return nil, gameStruct.Object
}

func DeleteGame(ctx context.Context, metadata Metadata, client *dynamic.DynamicClient, clientset *kubernetes.Clientset, force bool) error {
	resource := client.Resource(GameGCR).Namespace(metadata.Namespace)
	err := resource.Delete(ctx, metadata.Name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	if force {
		err := removeFleetsForGame(ctx, metadata, client, clientset, force)
		if err != nil {
			return err
		}
	}

	return nil
}

func removeFleetsForGame(ctx context.Context, metadata Metadata, client *dynamic.DynamicClient, clientset *kubernetes.Clientset, force bool) error {
	fleets, err := client.Resource(FleetGCR).Namespace(metadata.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, fleet := range fleets.Items {
		gamename, exists := fleet.GetLabels()["game"]
		if !exists || gamename != metadata.Name {
			continue
		}
		err := DeleteFleet(ctx, metadata, client, clientset, force)
		if err != nil {
			return err
		}
	}
	return nil
}
