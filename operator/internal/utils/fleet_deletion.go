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

type FleetDeletionChecker interface {
	isDeleteAllowed(ctx context.Context, server *networkv1alpha1.Server, c *client.Client) (bool, error)
}

// FindDeleteServer is used to find the server that should be deleted.
// It is based on the specs agepriority field.
func FindDeleteServer(ctx context.Context, fleet *networkv1alpha1.Fleet, servers *networkv1alpha1.ServerList, client client.Client, checker FleetDeletionChecker) (*networkv1alpha1.Server, error) {
	strategy := fleet.Spec.Scaling.AgePriority
	deleteFirst := fleet.Spec.Scaling.PrioritizeAllowed

	if strategy == networkv1alpha1.OldestFirst {
		return getOldestServer(ctx, servers, deleteFirst, &client, checker)
	}

	if strategy == networkv1alpha1.NewestFirst {
		return getNewestServer(ctx, servers, deleteFirst, &client, checker)

	}
	return nil, fmt.Errorf("invalid scaling strategy: %s", strategy)
}

// getOldestServer gets the server of the fleet, that is the oldest
// If deleteFirst is enabled, then it tries to get the server that can be deleted, but that is the oldest out of those.
// If it cannot find any where deletion is allowed, it returns the oldest server.
func getOldestServer(ctx context.Context, servers *networkv1alpha1.ServerList, deleteFirst bool, client *client.Client, checker FleetDeletionChecker) (*networkv1alpha1.Server, error) {
	var oldestServer *networkv1alpha1.Server
	var oldestTime *metav1.Time
	var oldestAllowedServer *networkv1alpha1.Server
	var oldestAllowTime *metav1.Time

	//Go over all servers
	for i := range servers.Items {
		server := &servers.Items[i]
		//If time is null, or this server was created before current oldest
		if oldestTime == nil || server.CreationTimestamp.Before(oldestTime) {
			//Set new oldest
			oldestTime = &server.CreationTimestamp
			oldestServer = server
		}
		// Check if we want to prioritize allowed
		if deleteFirst {
			//Check if this is allowed
			allowed, err := checker.isDeleteAllowed(ctx, server, client)
			if err != nil {
				return nil, err
			}
			if allowed {
				//If it is allowed, check current oldest allowed
				if oldestAllowTime == nil || server.CreationTimestamp.Before(oldestAllowTime) {
					//Update to new oldest allowed
					oldestAllowTime = &server.CreationTimestamp
					oldestAllowedServer = server
				}
			}
		}
	}

	if oldestServer == nil {
		return nil, fmt.Errorf("no servers found")
	}

	//If this is not nil, it means we found an oldest server and the boolean was true
	if oldestAllowedServer != nil {
		return oldestAllowedServer, nil
	}

	//If above was nil, return this instead
	return oldestServer, nil
}

// getNewestServer gets the server of the fleet, that is the newest
// If deleteFirst is enabled, then it tries to get the server that can be deleted, but that is the youngest out of those.
// If it cannot find any where deletion is allowed, it returns the youngest server.
func getNewestServer(ctx context.Context, servers *networkv1alpha1.ServerList, deleteFirst bool, client *client.Client, checker FleetDeletionChecker) (*networkv1alpha1.Server, error) {

	var newestServer *networkv1alpha1.Server
	var newestTime *metav1.Time
	var newestAllowedServer *networkv1alpha1.Server
	var newestAllowTime *metav1.Time

	// Go over all of the servers
	for i := range servers.Items {
		server := &servers.Items[i]
		//If time is null, or this server was created after current youngest
		if newestTime == nil || server.CreationTimestamp.After(newestTime.Time) {
			// Update with new youngest
			newestTime = &server.CreationTimestamp
			newestServer = server
		}
		// Check if we want to prioritize allowed
		if deleteFirst {
			//Check if this is allowed
			allowed, err := checker.isDeleteAllowed(ctx, server, client)
			if err != nil {
				return nil, err
			}
			if allowed {
				// If it is, we want to check if there already is a prioritized time, and is it after the current iterations one
				if newestAllowTime == nil || server.CreationTimestamp.After(newestAllowTime.Time) {
					newestAllowTime = &server.CreationTimestamp
					newestAllowedServer = server
				}
			}
		}
	}

	if newestServer == nil {
		return nil, fmt.Errorf("no servers found")
	}

	//If this is not nil, it means we found a youngest server and the boolean was true
	if newestAllowedServer != nil {
		return newestAllowedServer, nil
	}

	//If above was nil, return this instead
	return newestServer, nil
}

// isDeleteAllowed is a utility for a server object, to communicate with the sidecar to see if deletion is allowed
func (ProdDeletionChecker) isDeleteAllowed(ctx context.Context, server *networkv1alpha1.Server, c *client.Client) (bool, error) {
	podName := server.Name + "-pod"
	pod := &v1.Pod{}
	err := (*c).Get(ctx, types.NamespacedName{Namespace: server.Namespace, Name: podName}, pod)
	if err != nil {
		return false, err
	}

	allowed, err := IsDeleteAllowed(pod)
	if err != nil {
		return false, nil
	}

	return allowed, nil
}
