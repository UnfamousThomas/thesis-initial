package handlers

import (
	"bytes"
	"encoding/json"
	"github.com/unfamousthomas/thesis-sidecar/internal/app"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIsShutdownRequested(t *testing.T) {
	a := &app.App{ShutdownRequested: true}
	req := httptest.NewRequest(http.MethodGet, "/shutdown", nil)
	rec := httptest.NewRecorder()

	handler := http.HandlerFunc(IsShutdownRequested(a))
	handler.ServeHTTP(rec, req)

	resp := rec.Result()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status code: %d. Expected 200", resp.StatusCode)
	}

	var response ShutdownRequest
	err := json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Error decoding response: %v", err)
	}

	if response.Shutdown != true {
		t.Fatalf("Delete allowed should be true, got %v", response.Shutdown)
	}
}

func TestSetShutdownRequested(t *testing.T) {
	a := &app.App{ShutdownRequested: false}
	requestBody, err := json.Marshal(ShutdownRequest{Shutdown: true})
	if err != nil {
		t.Fatalf("Error encoding request body: %v", err)
	}
	req := httptest.NewRequest(http.MethodGet, "/allow_delete", bytes.NewReader(requestBody))
	rec := httptest.NewRecorder()

	handler := http.HandlerFunc(SetShutdownRequested(a))
	handler.ServeHTTP(rec, req)

	resp := rec.Result()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status code: %d. Expected 200", resp.StatusCode)
	}

	var response ShutdownRequest
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Error decoding response: %v", err)
	}

	if response.Shutdown != true {
		t.Fatalf("Delete allowed should be true, got %v", response.Shutdown)
	}
}

func TestSetShutdownRequestedInvalid(t *testing.T) {
	a := &app.App{ShutdownRequested: false}

	invalidBody := bytes.NewBufferString("{invalid_json}")

	req := httptest.NewRequest(http.MethodPost, "/set-delete-allowed", invalidBody)
	rec := httptest.NewRecorder()

	handler := http.HandlerFunc(SetShutdownRequested(a))
	handler.ServeHTTP(rec, req)

	resp := rec.Result()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400 Bad Request Error, got %v", resp.StatusCode)
	}

	if a.ShutdownRequested != false {
		t.Errorf("expected ShutdownAllowed=false, got %v", a.ShutdownRequested)
	}
}
