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

// The structs and types are copied from the operator types

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
	ApiVersion APIVersion `json:"apiVersion"`
	Kind       Kind       `json:"kind"`
	Metadata   Metadata   `json:"metadata"`
	Spec       ServerSpec `json:"spec"`
}

// CreateServer is used to create a new Server resource on the cluster, matching the Server struct
func CreateServer(context context.Context, server *Server, client *dynamic.DynamicClient) error {
	resource := client.Resource(ServerGCR).Namespace(server.Metadata.Namespace)
	serverStruct, err := serverToUnstructured(server)
	if err != nil {
		return err
	}
	_, err = resource.Create(context, serverStruct, metav1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

// DeleteServer is used to delete a Server resource from the cluster, based on Metadata.Name and Metadata.Namespace
func DeleteServer(context context.Context, metadata Metadata, client *dynamic.DynamicClient, clientset *kubernetes.Clientset, force bool) error {
	resource := client.Resource(ServerGCR).Namespace(metadata.Namespace)
	err := resource.Delete(context, metadata.Name, metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	if force {
		err = sendDeleteAllowed(context, metadata.Name, metadata.Namespace, clientset)
		if err != nil {
			return err
		}
	}
	return nil
}

// sendDeleteAllowed is used to tell the pods they can be deleted. This is used when force is true for DeleteServer
func sendDeleteAllowed(context context.Context, name string, namespace string, client *kubernetes.Clientset) error {
	resource := client.CoreV1().Pods(namespace)
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

// serverToUnstructured is used to make a Server object into a unstructured object which can interact with dynamic client
func serverToUnstructured(server *Server) (*unstructured.Unstructured, error) {
	server.ApiVersion = crdGroup + "/" + crdVersion
	server.Kind = gameResourceName
	bodyBytes, err := json.Marshal(server)
	if err != nil {
		return nil, err
	}

	obj := &unstructured.Unstructured{}
	if err := obj.UnmarshalJSON(bodyBytes); err != nil {
		return nil, err
	}
	return obj, nil
}
