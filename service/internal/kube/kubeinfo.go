package kube

const (
	crdGroup           = "network.unfamousthomas.me"
	crdVersion         = "v1alpha1"
	scalerResourceName = "gameautoscaler"
	gameResourceName   = "gametype"
	fleetResourceName  = "fleet"
	serverResourceName = "server"
)

type Metadata struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Labels    map[string]string `json:"labels"`
}
