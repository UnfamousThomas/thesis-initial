package kube

const (
	crdGroup           = "network.unfamousthomas.me"
	crdVersion         = "v1alpha1"
	scalerResourceName = "GameAutoscaler"
	gameResourceName   = "GameTypes"
	fleetResourceName  = "Fleets"
	serverResourceName = "Servers"
)

type Metadata struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Labels    map[string]string `json:"labels"`
}
