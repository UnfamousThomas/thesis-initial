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
  gameName: gametype-sample #Gametype gametype-sample has to exist
  policy:
    type: webhook #Currently only support Webhook
    webhook: #Webhook specifications, either path and url OR service need to be defined
      path: "/scale"
      url: "http://localhost:2411"
      service:
        name: someService123
        namespace: default
        port: 8080
  sync:
    type: fixedinterval #Currently only supports fixedinterval
    interval: 10m #How often to send the webhook
```

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