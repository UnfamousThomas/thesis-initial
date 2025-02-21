package kube

const (
	crdGroup           = "network.unfamousthomas.me"
	crdVersion         = "v1alpha1"
	scalerResourceName = "gameautoscalers"
	gameResourceName   = "gametypes"
	fleetResourceName  = "fleets"
	serverResourceName = "servers"
)

type Metadata struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Labels    map[string]string `json:"labels"`
}
