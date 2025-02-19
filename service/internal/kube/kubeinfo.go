package kube

const (
	crdGroup           = "network.unfamousthomas.me"
	crdVersion         = "v1alpha1"
	scalerResourceName = "GameAutoscaler"
	gameResourceName   = "GameType"
	fleetResourceName  = "Fleet"
	serverResourceName = "Server"
)

type Metadata struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Labels    map[string]string `json:"labels"`
}
