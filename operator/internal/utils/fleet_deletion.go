package utils

import (
	"context"
	_ "errors"
	"fmt"
	networkv1alpha1 "github.com/unfamousthomas/thesis-operator/api/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func FindDeleteServer(ctx context.Context, fleet *networkv1alpha1.Fleet, servers *networkv1alpha1.ServerList, client client.Client) (*networkv1alpha1.Server, error) {
	strategy := fleet.Spec.Scaling.AgePriority
	deleteFirst := fleet.Spec.Scaling.PrioritizeAllowed

	if strategy == networkv1alpha1.OldestFirst {
		return getOldestServer(ctx, servers, deleteFirst, client)
	}

	if strategy == networkv1alpha1.NewestFirst {
		return getNewestServer(ctx, servers, deleteFirst, client)

	}
	return nil, fmt.Errorf("invalid scaling strategy: %s", strategy)
}

func getOldestServer(ctx context.Context, servers *networkv1alpha1.ServerList, deleteFirst bool, client client.Client) (*networkv1alpha1.Server, error) {
	var oldestServer *networkv1alpha1.Server
	var oldestTime *metav1.Time
	var oldestAllowedServer *networkv1alpha1.Server
	var oldestAllowTime *metav1.Time

	for _, server := range servers.Items {
		if oldestTime == nil || server.CreationTimestamp.Before(oldestTime) {
			oldestTime = &server.CreationTimestamp
			oldestServer = &server
			if deleteFirst {
				allowed, err := isDeleteAllowed(ctx, &server, client)
				if err != nil {
					return nil, err
				}
				if allowed {
					if oldestAllowTime == nil || server.CreationTimestamp.Before(oldestAllowTime) {
						oldestAllowTime = &server.CreationTimestamp
						oldestAllowedServer = &server
					}
				}
			}
		}
	}

	if oldestServer == nil {
		return nil, fmt.Errorf("no servers found")
	}

	if oldestAllowedServer != nil {
		return oldestAllowedServer, nil
	}

	return oldestServer, nil
}

func getNewestServer(ctx context.Context, servers *networkv1alpha1.ServerList, deleteFirst bool, client client.Client) (*networkv1alpha1.Server, error) {

	var newestServer *networkv1alpha1.Server
	var newestTime *metav1.Time
	var newestAllowedServer *networkv1alpha1.Server
	var newestAllowTime *metav1.Time

	for _, server := range servers.Items {
		if newestTime == nil || server.CreationTimestamp.After(newestTime.Time) {
			newestTime = &server.CreationTimestamp
			newestServer = &server
			if deleteFirst {
				allowed, err := isDeleteAllowed(ctx, &server, client)
				if err != nil {
					return nil, err
				}
				if allowed {
					if newestAllowTime == nil || server.CreationTimestamp.After(newestAllowTime.Time) {
						newestAllowTime = &server.CreationTimestamp
						newestAllowedServer = &server
					}
				}
			}
		}
	}

	if newestServer == nil {
		return nil, fmt.Errorf("no servers found")
	}

	if newestAllowedServer != nil {
		return newestAllowedServer, nil
	}

	return newestServer, nil
}

func isDeleteAllowed(ctx context.Context, server *networkv1alpha1.Server, c client.Client) (bool, error) {
	podName := server.Name + "-pod"
	pod := &v1.Pod{}
	err := c.Get(ctx, types.NamespacedName{Namespace: server.Namespace, Name: podName}, pod)
	if err != nil {
		return false, err
	}

	allowed, err := IsDeleteAllowed(pod)
	if err != nil {
		return false, nil
	}

	return allowed, nil
}
