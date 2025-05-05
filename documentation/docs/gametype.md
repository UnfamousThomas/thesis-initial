# GameType

A **GameType** represents a scalable container for one or more fleets, enabling the management and upgrading of server versions. It serves as the central unit for scaling and versioning of game server deployments.

### Manifest

The manifest for a **GameType** object looks like this:

```yaml
apiVersion: network.unfamousthomas.me/v1alpha1
kind: GameType
metadata:
  name: gametype-sample
spec:
  fleetSpec: # (1)! The specification for the fleet associated with this GameType
    scaling:
      replicas: 3 # (2)!
      prioritizeAllowed: true # (3)!
      agePriority: oldest_first # (4)!
    spec: # (5)!
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

1. The specification for the fleet associated with this GameType.
2. The number of replicas (servers) to maintain.
3. Whether to delete allowed servers first when downscaling.
4. Whether to prioritize deleting oldest or newest servers when scaling down.
5. The spec for the servers within this fleet (same as a Server object spec).
### Purpose

The **GameType** object currently acts as a wrapper for 1-2 fleets. While its manifest closely mirrors the **Fleet** object, it provides the additional role of handling multiple fleet versions. This allows for gradual upgrades or changes in server configurations, with the flexibility to roll out new fleet versions in a controlled manner.

### Upgrade Process

When the pod spec (the configuration of the containers) for the servers changes, the GameType initiates the following process:
* **Create New Fleet**: A new fleet is created with the same number of replicas and the updated spec.

* **Gradual Upgrade**:  The old fleet is gradually removed according to the server deletion rules set in the fleet (e.g., prioritizing allowed deletions or age-based deletions).

* **Trigger Old Fleet Deletion**: Once the new fleet is running and stable, the old fleet is triggered for deletion, again adhering to the server's deletion policies.

This ensures a smooth upgrade process, minimizing downtime and adhering to the configured server policies.

### Future Changes
The **GameType** spec may evolve in the future to provide more advanced scaling or versioning capabilities. However, for now, it functions as a simple container for one or two fleets, with the primary goal of supporting controlled upgrades and scaling of game servers.