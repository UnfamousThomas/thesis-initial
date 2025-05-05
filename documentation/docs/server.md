# Server

A **server** is the core resource in this project, serving as the foundation for managing game server instances in Kubernetes.

It has a **1-to-1 relationship with a pod**, meaning each server corresponds to a single pod. The server acts as a simple wrapper around the pod, with an additional **sidecar** inserted to manage specialized functionality (such as lifecycle control and custom metrics). For more information about the sidecar, please see the [Sidecar](sidecar.md) documentation.

### Manifest

The manifest for a **Server** object looks like this:

```yaml
apiVersion: network.unfamousthomas.me/v1alpha1
kind: Server
metadata:
  labels:
    someLabel: example-label # (1)!
  name: server-sample
spec:
  timeout: 5m # (2)!
  allowForceDelete: false # (3)!
  pod:
    containers:
      - name: example-container
        image: nginx:latest
        ports:
          - containerPort: 80
            protocol: TCP
        resources:
          limits:
            cpu: "500m"
            memory: "256Mi"
          requests:
            cpu: "250m"
            memory: "128Mi"
```

1. Labels are copied down to the pod.
2. Time the controller waits before allowing a forced deletion of the server.
3. Determines if the controller can forcefully delete the server without permission. This would mean that it will not ask the [sidecar](sidecar.md) for permission.

The `spec.pod` sub-object follows the typical Kubernetes pod spec format, allowing you to define multiple containers, volumes, and other pod configurations as necessary.

### Special Environment Variables

For ease of development and to provide useful context inside containers, the following environment variables are automatically introduced to all containers:

- `CONTAINER_IMAGE` - The image of the current container
- `SERVER_NAME` - The name of the server object
- `FLEET_NAME` - The name of the parent fleet object (if applicable)
- `GAME_NAME` - The name of the parent game object (if applicable)
- `POD_IP` - The IP address of the pod
- `NODE_NAME` - The name of the node where the pod is running

### Server Management

- **Duration**: The `duration` field determines how long the controller will wait before permitting a forced deletion of the server. This allows you to implement a delay before the server is deleted to prevent premature shutdowns.

- **AllowForceDelete**: The `allowForceDelete` field, when set to `true`, instructs the controller to delete the server without waiting for user permission. If set to `false`, the server will only be deleted after a manual approval.

By setting these fields, you can fine-tune how the server lifecycle is managed within your Kubernetes environment.

### Tips and Considerations
- **Resource Allocation**: It’s crucial to define the `cpu` and `memory` limits and requests according to the expected usage of the server to avoid performance issues.