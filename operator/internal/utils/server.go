package utils

import (
	"github.com/unfamousthomas/thesis-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateServerForFleet(fleet v1alpha1.Fleet, namespace string) *v1alpha1.Server {
	labels := fleet.Labels
	if labels == nil {
		labels = make(map[string]string)
	}
	labels["fleet"] = fleet.Name
	server := v1alpha1.Server{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fleet.Name + "-",
			Namespace:    namespace,
			Labels:       labels,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(&fleet, v1alpha1.GroupVersion.WithKind("Fleet")),
			},
		},
		Spec: fleet.Spec.ServerSpec,
	}

	return &server
}
