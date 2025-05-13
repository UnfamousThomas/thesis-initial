package utils

type EventReason string

const (
	ReasonServerInitialized        EventReason = "ServerInitialized"
	ReasonServerDeletionAllowed    EventReason = "ServerDeletionAllowed"
	ReasonServerDeletionNotAllowed EventReason = "ServerDeletionNotAllowed"
	ReasonServerPodDeleted         EventReason = "ServerPodDeleted"
	ReasonServerPodCreationFailed  EventReason = "ServerPodCreationFailed"
	ReasonServerUpdateFAiled       EventReason = "ServerUpdateFailed"

	ReasonFleetInitialized    EventReason = "FleetInitialized"
	ReasonFleetUpdateFailed   EventReason = "FleetUpdateFailed"
	ReasonFleetServersRemoved EventReason = "FleetServersRemoved"
	ReasonFleetScaleServers   EventReason = "FleetScaleServers"

	ReasonGametypeInitialized     EventReason = "GametypeInitialized"
	ReasonGameTypeDeleting        EventReason = "GameTypeDeleting"
	ReasonGametypeServersDeleted  EventReason = "GametypeServersDeleted"
	ReasonGametypeSpecUpdated     EventReason = "GametypeSpecUpdated"
	ReasonGametypeReplicasUpdated EventReason = "GametypeReplicasUpdated"

	ReasonGameAutoscalerInvalidServer          EventReason = "GameAutoscalerInvalidServer"
	ReasonGameAutoscalerInvalidAutoscalePolicy EventReason = "GameautoscalerInvalidAutoscalePolicy"
	ReasonGameAutoscalerInvalidSyncType        EventReason = "GameautoscalerInvalidSyncType"
	ReasonGameautoscalerWebhook                EventReason = "GameautoscalerWebhook"
	ReasonGameautoscalerScale                  EventReason = "GameautoscalerScale"
)
