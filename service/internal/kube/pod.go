package kube

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"maps"
)

// AddPodLabels adds label to the pod object, note that all labels from metadata are copied.
// Meaning it overwrites.
func AddPodLabels(context context.Context, metadata Metadata, client *kubernetes.Clientset) error {
	resource := client.CoreV1().Pods(metadata.Namespace)
	pod, err := resource.Get(context, metadata.Name+"-pod", metav1.GetOptions{})
	if err != nil {
		return err
	}

	labels := pod.GetLabels()
	maps.Copy(labels, metadata.Labels)
	pod.SetLabels(labels)

	_, err = resource.Update(context, pod, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func RemovePodLabel(context context.Context, metadata Metadata, label string, client *kubernetes.Clientset) error {
	resource := client.CoreV1().Pods(metadata.Namespace)
	pod, err := resource.Get(context, metadata.Name+"-pod", metav1.GetOptions{})
	if err != nil {
		return err
	}

	labels := pod.GetLabels()
	delete(labels, label)

	pod.SetLabels(labels)
	_, err = resource.Update(context, pod, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}
