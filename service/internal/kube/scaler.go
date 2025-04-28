package kube

import (
	"context"
	"encoding/json"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
)

// The structs and types are copied from the operator types

type PolicyStrategy string
type SyncStrategy string

var validPolicyStrategies = map[PolicyStrategy]struct{}{
	Webhook: {},
	// Add new strategies here as needed
}
var validSyncStrategy = map[SyncStrategy]struct{}{
	FixedInterval: {},
	// Add new strategies here as needed
}

var (
	Webhook       PolicyStrategy = "webhook"
	FixedInterval SyncStrategy   = "fixedinterval"
)

type GameAutoscalerSpec struct {
	GameName        string          `json:"gameName"`
	AutoscalePolicy AutoscalePolicy `json:"policy"`
	Sync            Sync            `json:"sync"`
}

type AutoscalePolicy struct {
	Type                  PolicyStrategy        `json:"type"`
	WebhookAutoscalerSpec WebhookAutoscalerSpec `json:"webhook"`
}

type WebhookAutoscalerSpec struct {
	Url     *string `json:"url"`
	Path    string  `json:"path"`
	Service Service `json:"service"`
}

type Service struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Port      int    `json:"port"`
}
type Sync struct {
	Type          SyncStrategy `json:"type"`
	FixedInterval int          `json:"fixedInterval"`
}

type GameAutoscaler struct {
	ApiVersion APIVersion         `json:"apiVersion"`
	Kind       Kind               `json:"kind"`
	Metadata   Metadata           `json:"metadata"`
	Spec       GameAutoscalerSpec `json:"spec"`
}

// CreateScaler is used to create a new gameautoscaler using the dynamic client
func CreateScaler(ctx context.Context, scaler *GameAutoscaler, client *dynamic.DynamicClient) error {
	resource := client.Resource(ServerGCR).Namespace(scaler.Metadata.Namespace)
	scalerStruct, err := scalerToUnstructured(scaler)
	if err != nil {
		return err
	}
	_, err = resource.Create(ctx, scalerStruct, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

// DeleteScaler is used to delete a gameautoscaler, based on the namespace and name passed to the metadata
func DeleteScaler(ctx context.Context, metadata Metadata, client *dynamic.DynamicClient) error {
	resource := client.Resource(ServerGCR).Namespace(metadata.Namespace)
	err := resource.Delete(ctx, metadata.Name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}

// serverToUnstructured is used to make a GameAutoscaler object into an unstructured object which can interact with dynamic client
func scalerToUnstructured(autoscaler *GameAutoscaler) (*unstructured.Unstructured, error) {
	autoscaler.ApiVersion = crdGroup + "/" + crdVersion
	autoscaler.Kind = scalerResourceName
	bytes, err := json.Marshal(autoscaler)
	if err != nil {
		return nil, err
	}

	obj := &unstructured.Unstructured{}
	if err := obj.UnmarshalJSON(bytes); err != nil {
		return nil, err
	}
	return obj, nil
}
