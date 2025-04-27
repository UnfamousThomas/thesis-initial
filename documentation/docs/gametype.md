# GameType

GameType acts as the thing to scale, as well as a container for different versions of servers.

### Manifest

The general manifest looks something like this:
```yaml
apiVersion: network.unfamousthomas.me/v1alpha1
kind: GameType
metadata:
  name: gametype-sample
spec:
  fleetSpec: #The fleet spec
    scaling:
      replicas: 3
      prioritizeAllowed: true
      agePriority: oldest_first
    spec: #The server spec
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

