package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"net/http"
	"time"
)

type deleteRequest struct {
	Allowed bool `json:"allowed"`
}

type shutdownRequest struct {
	Shutdown bool `json:"shutdown"`
}

// IsDeleteAllowed sents a request to API/allow_delete to ask the server if it can be shutdown and deleted
func IsDeleteAllowed(pod *v1.Pod) (bool, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(buildPodBaseAddress(pod) + "allow_delete")
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, errors.New("GET request returned: " + resp.Status)
	}

	var request deleteRequest
	err = json.NewDecoder(resp.Body).Decode(&request)
	if err != nil {
		return false, err
	}
	return request.Allowed, nil
}

// RequestShutdown sends a request to API/shutdown to tell the server that operator has requested its shutdown
func RequestShutdown(pod *v1.Pod) error {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	request := shutdownRequest{
		Shutdown: true,
	}
	requestBody, err := json.Marshal(request)
	if err != nil {
		return err
	}

	resp, err := client.Post(buildPodBaseAddress(pod)+"shutdown", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("POST request returned: " + resp.Status)
	}

	return nil
}

func buildPodBaseAddress(pod *v1.Pod) string {
	return fmt.Sprintf("http://%s:8080/", pod.Status.PodIP)
}
