package kube

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

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
	Metadata Metadata           `json:"metadata"`
	Spec     GameAutoscalerSpec `json:"spec"`
}

func CreateScaler(ctx context.Context, scaler GameAutoscaler, client *dynamic.DynamicClient) (error, map[string]interface{}) {
	resource := client.Resource(ServerGCR).Namespace(scaler.Metadata.Namespace)
	scalerStruct := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": crdGroup + "/" + crdVersion,
			"kind":       scalerResourceName,
			"metadata":   map[string]interface{}{"name": scaler.Metadata.Name, "namespace": scaler.Metadata.Namespace},
			"spec":       scaler.Spec,
		},
	}
	_, err := resource.Create(ctx, scalerStruct, metav1.CreateOptions{})
	if err != nil {
		return err, nil
	}
	return nil, scalerStruct.Object
}

func DeleteScaler(ctx context.Context, metadata Metadata, client *dynamic.DynamicClient, clientset *kubernetes.Clientset) error {
	resource := client.Resource(ServerGCR).Namespace(metadata.Namespace)
	err := resource.Delete(ctx, metadata.Name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}
