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

func getPodSpec(spec *corev1.PodSpec) *corev1.PodSpec {
	pod := addContainer(spec, corev1.Container{
		Name:  "loputoo-sidecar",
		Image: "ghcr.io/unfamousthomas/sidecar:latest",
		Ports: []corev1.ContainerPort{
			{
				Name:          "http",
				ContainerPort: 8080,
			},
		},
	})

	pod.ImagePullSecrets = append(pod.ImagePullSecrets, corev1.LocalObjectReference{
		Name: os.Getenv("IMAGE_PULL_SECRET_NAME"),
	})

	return pod
}

func GetNewPod(server *networkv1alpha1.Server, namespace string) *corev1.Pod {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      server.Name + "-pod",
			Namespace: namespace,
			Labels: map[string]string{
				"server": server.Name,
			},
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(server, networkv1alpha1.GroupVersion.WithKind("Server")),
			},
		},
		Spec: *getPodSpec(&server.Spec.Pod),
	}
	return pod
}
