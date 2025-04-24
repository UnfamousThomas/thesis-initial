package utils

type EventReason string

const (
	ReasonServerInitialized        EventReason = "ServerInitialized"
	ReasonServerDeletionAllowed    EventReason = "ServerDeletionAllowed"
	ReasonServerDeletionNotAllowed EventReason = "ServerDeletionNotAllowed"
	ReasonServerPodDeleted         EventReason = "ServerPodDeleted"
	ReasonServerPodCreationFailed  EventReason = "ServerPodCreationFailed"
)
