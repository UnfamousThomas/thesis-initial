package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	networkv1alpha1 "github.com/unfamousthomas/thesis-operator/api/v1alpha1"
	"io"
	"net/http"
	"time"
)

type Webhook interface {
	SendScaleWebhookRequest(autoscaler *networkv1alpha1.GameAutoscaler, gametype *networkv1alpha1.GameType) (AutoscaleResponse, error)
}

type ProductionWebhookRequest struct{}

func (w ProductionWebhookRequest) SendScaleWebhookRequest(autoscaler *networkv1alpha1.GameAutoscaler,
	gametype *networkv1alpha1.GameType) (AutoscaleResponse, error) {
	autoscalerSpec := autoscaler.Spec.AutoscalePolicy.WebhookAutoscalerSpec

	var url string
	if autoscalerSpec.Url != nil {
		url = *autoscalerSpec.Url
	} else {
		service := autoscalerSpec.Service
		url = fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", service.Name, service.Namespace, service.Port)
	}
	if autoscalerSpec.Path == nil {
		return AutoscaleResponse{}, errors.New("missing path")
	}
	path := *autoscalerSpec.Path
	url = url + "/" + path

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	request := AutoscaleRequest{
		GameName:        autoscaler.Spec.GameName,
		CurrentReplicas: int(gametype.Spec.FleetSpec.Scaling.Replicas),
	}
	requestBody, err := json.Marshal(request)
	if err != nil {
		return AutoscaleResponse{}, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return AutoscaleResponse{}, err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return AutoscaleResponse{}, err
	}

	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return AutoscaleResponse{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return AutoscaleResponse{}, fmt.Errorf("invalid request response: %d. Raw response: %s", resp.StatusCode, string(bodyBytes))
	}

	var response AutoscaleResponse
	err = json.Unmarshal(bodyBytes, &response)
	if err != nil {
		return AutoscaleResponse{}, fmt.Errorf("failed to decode response: %w\nRaw response: %s\n", err, string(bodyBytes))
	}
	return response, nil
}

type AutoscaleRequest struct {
	GameName        string `json:"game_name"`
	CurrentReplicas int    `json:"current_replicas"`
}

type AutoscaleResponse struct {
	Scale           bool `json:"scale"`
	DesiredReplicas int  `json:"desired_replicas"`
}
