package utils

import (
	"context"
	"github.com/go-logr/logr"
	networkv1alpha1 "github.com/unfamousthomas/thesis-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetFleetsForType(ctx context.Context, c client.Client, gametype *networkv1alpha1.GameType, logger logr.Logger) (*networkv1alpha1.FleetList, error) {
	fleetList := &networkv1alpha1.FleetList{}

	labelSelector := client.MatchingLabels{
		"type": gametype.Name,
	}

	if err := c.List(ctx, fleetList, labelSelector); err != nil {
		logger.Error(err, "Failed to list Fleets", "GameType", gametype.Name)
		return nil, err
	}

	return fleetList, nil
}

func GetFleetObjectForType(gametype *networkv1alpha1.GameType) *networkv1alpha1.Fleet {
	labels := gametype.Labels
	if labels == nil {
		labels = map[string]string{}
	}
	labels["type"] = gametype.Name

	fleet := &networkv1alpha1.Fleet{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: gametype.Name + "-",
			Namespace:    gametype.Namespace,
			Labels:       labels,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(gametype, networkv1alpha1.GroupVersion.WithKind("GameType")),
			},
		},
		Spec: gametype.Spec.FleetSpec,
	}

	return fleet
}
