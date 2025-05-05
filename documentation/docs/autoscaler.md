# GameAutoscaler

The **GameAutoscaler** is a resource that periodically triggers a webhook to adjust the replica count of a **GameType**. It sends requests to the defined webhook endpoint, and based on the response, it adjusts the number of replicas.

### Manifest

The manifest for a **GameAutoscaler** object looks like this:
```yaml
apiVersion: network.unfamousthomas.me/v1alpha1
kind: GameAutoscaler
metadata:
  name: gameautoscaler-sample
spec:
  gameName: gametype-sample # (1)!
  policy:
    type: webhook # (2)! 
    webhook: # (3)!
      url: "http://localhost:2411" # (4)!
      path: "/scale" # (5)!
      service: # (6)!
        name: someService123 
        namespace: default
        port: 8080
  sync:
    type: fixedinterval # (7)!
    interval: 10m # (8)!
```

1. Gametype gametype-sample has to exist.
2. Currently only support Webhook.
3. Webhook specifications, either path and url OR service need to be defined.
4. The url to send the request to. Combined with path if provided.
5. Path to send the request to. Combined with url.
6. Service to send the request to.
7. Currently only supports fixedinterval.
8. How often to send the webhook.
### Basic Concept
The GameAutoscaler object simplifies the autoscaling of GameType resources by using a webhook. The key fields to configure are:

* `gameName`: The name of the (already existing) **GameType** that needs to be scaled
* `policy.webhook`: The configuration of the webhook
* `sync`: Defines how often and how the webhook should be triggered. 

### Request JSONs
When the **GameAutoscaler** triggers the webhook, it sends a request in the following format:
```json
{
  "game_name": "gametype-sample",
  "current_replicas": 5
}
```

### Response JSONs
The webhook should respond with a JSON object indicating whether scaling is required and the desired replica count:

```json
{
  "scale": true,
  "desired_replicas": 10,
}
```
If the webhook returns the above, the **GameAutoscaler** will update the **GameType** to have 10 replicas.

#### Go Structs
For those interested in implementing the webhook in Go, here are the Go structs representing the request and response formats:

```go
type AutoscaleRequest struct {
	GameName        string `json:"game_name"`
	CurrentReplicas int    `json:"current_replicas"`
}

type AutoscaleResponse struct {
	Scale           bool `json:"scale"`
	DesiredReplicas int  `json:"desired_replicas"`
}
```