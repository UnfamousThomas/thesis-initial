package utils

import (
	"fmt"
	networkv1alpha1 "github.com/unfamousthomas/thesis-operator/api/v1alpha1"
	"strconv"
)

func CreateServerManifest(name string, namespace string, image string) string {
	manifest := fmt.Sprintf(`
apiVersion: network.unfamousthomas.me/v1alpha1
kind: Server
metadata:
  name: %s
  namespace: %s
spec:
  timeout: 5m
  allowForceDelete: false
  pod:
    containers:
      - name: gameserver
        image: %s
        ports:
          - containerPort: 8081
            protocol: TCP
`, name, namespace, image)

	return manifest
}

func CreateFleetManifest(name string, namespace string, image string, replicas int, prioritize bool, priority networkv1alpha1.Priority) string {
	manifest := fmt.Sprintf(`
apiVersion: network.unfamousthomas.me/v1alpha1
kind: Fleet
metadata:
  name: %s
  namespace: %s
spec:
  scaling:
    replicas: %d
    prioritizeAllowed: %s
    agePriority: %s
  spec:
    timeout: 5m
    allowForceDelete: false
    pod:
      containers:
      - name: gameserver
        image: %s
        ports:
        - containerPort: 8081
          protocol: TCP
`, name, namespace, replicas, strconv.FormatBool(prioritize), priority, image)

	return manifest
}

func CreateGameTypeManifest(name string, namespace string, image string, replicas int, prioritize bool, priority networkv1alpha1.Priority) string {
	manifest := fmt.Sprintf(`
apiVersion: network.unfamousthomas.me/v1alpha1
kind: GameType
metadata:
  name: %s
  namespace: %s
spec:
  fleetSpec:
    scaling:
      replicas: %d
      prioritizeAllowed: %s
      agePriority: %s
    spec:
      timeout: 5m
      allowForceDelete: false
      pod:
        containers:
        - name: gameserver
          image: %s
          ports:
          - containerPort: 8081
            protocol: TCP
`, name, namespace, replicas, strconv.FormatBool(prioritize), priority, image)

	return manifest
}
