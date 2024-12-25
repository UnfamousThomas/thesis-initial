package utils

import (
	networkv1alpha1 "github.com/unfamousthomas/thesis-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"time"
)

type Deletion interface {
	IsDeletionAllowed(*networkv1alpha1.Server, *corev1.Pod) (bool, error)
}

type PlayerCount interface {
	GetPlayerCount(*networkv1alpha1.Server) (int32, error)
}

type ProdDeletionChecker struct{}

func (p ProdDeletionChecker) GetPlayerCount(server *networkv1alpha1.Server) (int32, error) {
	return 0, nil
}

func (p ProdDeletionChecker) IsDeletionAllowed(server *networkv1alpha1.Server, pod *corev1.Pod) (bool, error) {
	if server.Spec.AllowForceDelete {
		return true, nil
	}

	if server.Spec.TimeOut != nil {
		timeWhenAllowDelete := server.GetDeletionTimestamp().Time.Add(server.Spec.TimeOut.Duration)
		if timeWhenAllowDelete.Before(time.Now()) {
			return true, nil
		}
	}
	err := RequestShutdown(pod)
	if err != nil {
		return false, err
	}
	allowed, err := IsDeleteAllowed(pod)
	return allowed, err
}
