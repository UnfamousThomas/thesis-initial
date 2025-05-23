package utils

import (
	"fmt"
)

func CreateServiceDeployManifest(image string) string {

	manifest := fmt.Sprintf(`
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: http-controller-service-clusterrole
rules:
  # Allow all operations on pods in the core API group
  - apiGroups: [""]
    resources: ["pods"]
    verbs: ["*"]
  # Allow all operations on the custom resources under network.unfamousthomas.me/v1alpha1
  - apiGroups: ["network.unfamousthomas.me"]
    resources:
      - gameautoscalers
      - gametypes
      - fleets
      - servers
    verbs: ["*"]
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: http-controller-service-sa
  namespace: default
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: http-controller-service-clusterrolebinding
subjects:
  - kind: ServiceAccount
    name: http-controller-service-sa
    namespace: default
roleRef:
  kind: ClusterRole
  name: http-controller-service-clusterrole
  apiGroup: rbac.authorization.k8s.io

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: http-controller-service
  labels:
    app: http-controller-service
spec:
  replicas: 1
  selector:
    matchLabels:
      app: http-controller-service
  template:
    metadata:
      labels:
        app: http-controller-service
    spec:
      serviceAccountName: http-controller-service-sa
      containers:
        - name: kube-http-service
          image: %s
          imagePullPolicy: Never
          ports:
            - containerPort: 8080


`, image)
	return manifest
}
