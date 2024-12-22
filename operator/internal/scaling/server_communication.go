package scaling

import networkv1alpha1 "github.com/unfamousthomas/thesis-operator/api/v1alpha1"

type Deletion interface {
	IsDeletionAllowed(*networkv1alpha1.Server) (bool, error)
}

type PlayerCount interface {
	GetPlayerCount(*networkv1alpha1.Server) (int32, error)
}

type ProdDeletionChecker struct{}

func (p ProdDeletionChecker) GetPlayerCount(server *networkv1alpha1.Server) (int32, error) {
	return 0, nil
}

func (p ProdDeletionChecker) IsDeletionAllowed(server *networkv1alpha1.Server) (bool, error) {
	return false, nil
}
