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

var GameGCR = schema.GroupVersionResource{
	Group:    crdGroup,
	Version:  crdVersion,
	Resource: gameResourceName,
}

type GameTypeSpec struct {
	FleetSpec FleetSpec `json:"fleetSpec"`
}

type GameType struct {
	ApiVersion APIVersion   `json:"apiVersion"`
	Kind       Kind         `json:"kind"`
	Metadata   Metadata     `json:"metadata"`
	Spec       GameTypeSpec `json:"spec"`
}

// CreateGame creates a new GameType in the cluster using the dynamic client.
func CreateGame(context context.Context, game *GameType, client *dynamic.DynamicClient) error {
	resource := client.Resource(GameGCR).Namespace(game.Metadata.Namespace)
	gameStruct, err := gametypeToUnstructured(game)
	if err != nil {
		return err
	}
	_, err = resource.Create(context, gameStruct, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return err
}

// DeleteGame triggers the game deletion, it is possible to force it using the force variable. It finds the game using Metadata.Name and Metadata.Namespace
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

// removeFleetsForGame is used for deleting all fleets related to a game. This is used when force is true.
func removeFleetsForGame(ctx context.Context, metadata Metadata, client *dynamic.DynamicClient, clientset *kubernetes.Clientset, force bool) error {
	fleets, err := client.Resource(FleetGCR).Namespace(metadata.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, fleet := range fleets.Items {
		gamename, exists := fleet.GetLabels()["type"]
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

// gametypeToUnstructured is used to make a GameType object into a unstructured object which can interact with dynamic client
func gametypeToUnstructured(gametype *GameType) (*unstructured.Unstructured, error) {
	gametype.ApiVersion = crdGroup + "/" + crdVersion
	gametype.Kind = gameResourceName
	bytes, err := json.Marshal(gametype)
	if err != nil {
		return nil, err
	}

	obj := &unstructured.Unstructured{}
	if err := obj.UnmarshalJSON(bytes); err != nil {
		return nil, err
	}
	return obj, nil
}
