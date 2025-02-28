package kube

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"net/http"
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

func DeleteServer(context context.Context, metadata Metadata, client *dynamic.DynamicClient, clientset *kubernetes.Clientset, force bool) error {
	resource := client.Resource(ServerGCR).Namespace(metadata.Namespace)
	err := resource.Delete(context, metadata.Name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	if force {
		err = sendDeleteAllowed(context, metadata.Name, clientset)
		if err != nil {
			return err
		}
	}
	return nil
}

func sendDeleteAllowed(context context.Context, name string, client *kubernetes.Clientset) error {
	resource := client.CoreV1().Pods(name)
	pod, err := resource.Get(context, name+"-pod", metav1.GetOptions{})
	if err != nil {
		return err
	}

	url := fmt.Sprintf("http://%s:8080/allow_delete", pod.Status.PodIP)
	payload := map[string]any{
		"allowed": true,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	httpclient := &http.Client{}
	resp, err := httpclient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
