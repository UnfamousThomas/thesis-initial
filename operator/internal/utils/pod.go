package utils

import (
	networkv1alpha1 "github.com/unfamousthomas/thesis-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
)

func addContainer(spec *corev1.PodSpec, container corev1.Container) *corev1.PodSpec {
	spec.Containers = append(spec.Containers, container)
	return spec
}

func getPodSpec(server *networkv1alpha1.Server) *corev1.PodSpec {
	spec := server.Spec
	pod := addContainer(&spec.Pod, corev1.Container{
		Name:  "loputoo-sidecar",
		Image: "ghcr.io/unfamousthomas/sidecar:latest",
		Ports: []corev1.ContainerPort{
			{
				Name:          "http",
				ContainerPort: 8080,
			},
		},
	})
	for i := range pod.Containers {
		container := &pod.Containers[i]
		container.Env = append(container.Env, corev1.EnvVar{
			Name:  "CONTAINER_IMAGE",
			Value: container.Image,
		})
		container.Env = append(container.Env, corev1.EnvVar{
			Name:  "SERVER_NAME",
			Value: server.Name,
		})
		if fleet, ok := server.Labels["fleet"]; ok {
			container.Env = append(container.Env, corev1.EnvVar{
				Name:  "FLEET_NAME",
				Value: fleet,
			})
		}

		if fleet, ok := server.Labels["type"]; ok {
			container.Env = append(container.Env, corev1.EnvVar{
				Name:  "GAME_NAME",
				Value: fleet,
			})
		}

	}

	pod.ImagePullSecrets = append(pod.ImagePullSecrets, corev1.LocalObjectReference{
		Name: os.Getenv("IMAGE_PULL_SECRET_NAME"),
	})

	return pod
}

func GetNewPod(server *networkv1alpha1.Server, namespace string) *corev1.Pod {
	labels := server.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}
	labels["server"] = server.Name
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      server.Name + "-pod",
			Namespace: namespace,
			Labels:    labels,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(server, networkv1alpha1.GroupVersion.WithKind("Server")),
			},
		},
		Spec: *getPodSpec(server),
	}
	return pod
}
