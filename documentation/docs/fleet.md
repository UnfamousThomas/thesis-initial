# Fleet

A **Fleet** represents a group of servers managed together. Fleets simplify the process of scaling the number of servers up or down, based on the needs of the game or application. They act as a container for managing the lifecycle and scaling rules of multiple servers.

### Manifest

The manifest for a **Fleet** object looks like this:

```yaml
apiVersion: network.unfamousthomas.me/v1alpha1
kind: Fleet
metadata:
  name: fleet-sample
  labels:
    some-label: value1 # (5)!
spec:
  scaling:
    replicas: 3 # (1)!
    prioritizeAllowed: true # (2)!
    agePriority: oldest_first # (3)!
  spec: # (4)!
    timeout: 5m
    allowForceDelete: false
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

1. The number of servers that should exist in this fleet.
2. Whether to delete allowed servers first during downscaling.
3. Whether to delete oldest servers first or newest (`newest_first`).
4. The spec for the servers in this fleet (same as a Server object spec). **Note**: When this is changed it applies to new servers created 
but not to old ones.
5. Labels are copied down to the [Server](server.md) resource.

### Scaling Behaviour
* **Replicas**: The `replicas` field defines how many server instances should exist in the fleet at any time. This field ensures that the fleet always maintains the desired number of servers.
* **PrioritizeAllowed**: The `prioritizeAllowed` field determines whether servers that have deletion allowed should be deleted first during a downscale event. If `true`, servers that are marked as allowed for deletion will be removed first to maintain the specified number of replicas.
* **AgePriority**: The `agePriority` fields determines the order in which servers are deleted when downscaling. By default, `oldest_first` will remove the oldest servers first. If set to newest_first, the most recently created servers will be deleted first. You can add additional priorities here as needed.  


Simply put, this just acts as a simple container for multiple servers.

### Default Behaviour
* **Replicas**: If not specified, the controller will default to creating 1 replica of the server
* **PrioritizeAllowed**: If not specified, it defaults to `true`, meaning the controller will prioritize deleting allowed servers during downscaling.
* **AgePriority**: If not specified, the controller will default to `oldest_first`, ensuring that the oldest servers are removed first during downscaling.

### Spec Inheritance
The `spec.spec` field inside the Fleet manifest is identical to the spec used in the Server object. This means that each server created by a fleet will inherit its configuration from this spec, including container settings, resource requests, and limits.

For more information on the Server spec, see the [Server](server.md) documentation.