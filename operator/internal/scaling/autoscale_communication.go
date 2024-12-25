package scaling

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	networkv1alpha1 "github.com/unfamousthomas/thesis-operator/api/v1alpha1"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

type AutoscaleRequest struct {
	GameName        string `json:"game_name"`
	CurrentReplicas int    `json:"current_replicas"`
	MinReplicas     int    `json:"min_replicas"`
	MaxReplicas     int    `json:"max_replicas"`
}

type AutoscaleResponse struct {
	Scale           bool `json:"scale"`
	DesiredReplicas int  `json:"desired_replicas"`
}

func SendScaleWebhookRequest(context context.Context, autoscaler *networkv1alpha1.GameAutoscaler, gametype *networkv1alpha1.GameType, c client.Client) (AutoscaleResponse, error) {

	autoscalerSpec := autoscaler.Spec.AutoscalePolicy.WebhookAutoscalerSpec

	var url string
	if autoscalerSpec.Url != nil {
		url = *autoscalerSpec.Url
	} else {
		service := autoscalerSpec.Service
		url = fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", service.Name, service.Namespace, service.Port)
	}
	url = url + "/" + autoscalerSpec.Path

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	request := AutoscaleRequest{
		GameName:        autoscaler.Spec.GameName,
		CurrentReplicas: gametype.Status.CurrentReplicas,
		MinReplicas:     gametype.Spec.Scaling.MinReplicas,
		MaxReplicas:     gametype.Spec.Scaling.MaxReplicas,
	}
	requestBody, err := json.Marshal(request)
	if err != nil {
		return AutoscaleResponse{}, err
	}

	req, err := http.NewRequest("GET", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return AutoscaleResponse{}, err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return AutoscaleResponse{}, err
	}

	defer resp.Body.Close()

	var response AutoscaleResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return AutoscaleResponse{}, err
	}
	return response, nil
}
