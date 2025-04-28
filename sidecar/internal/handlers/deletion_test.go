package handlers

import (
	"bytes"
	"encoding/json"
	"github.com/unfamousthomas/thesis-sidecar/internal/app"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIsDeleteAllowed(t *testing.T) {
	a := &app.App{DeleteAllowed: true}
	req := httptest.NewRequest(http.MethodGet, "/allow_delete", nil)
	rec := httptest.NewRecorder()

	handler := http.HandlerFunc(IsDeleteAllowed(a))
	handler.ServeHTTP(rec, req)

	resp := rec.Result()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status code: %d. Expected 200", resp.StatusCode)
	}

	var response DeleteRequest
	err := json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Error decoding response: %v", err)
	}

	if response.Allowed != true {
		t.Fatalf("Delete allowed should be true, got %v", response.Allowed)
	}
}

func TestSetDeleteAllowed(t *testing.T) {
	a := &app.App{DeleteAllowed: false}
	requestBody, err := json.Marshal(DeleteRequest{Allowed: true})
	if err != nil {
		t.Fatalf("Error encoding request body: %v", err)
	}
	req := httptest.NewRequest(http.MethodGet, "/allow_delete", bytes.NewReader(requestBody))
	rec := httptest.NewRecorder()

	handler := http.HandlerFunc(SetDeleteAllowed(a))
	handler.ServeHTTP(rec, req)

	resp := rec.Result()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status code: %d. Expected 200", resp.StatusCode)
	}

	var response DeleteRequest
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Error decoding response: %v", err)
	}

	if response.Allowed != true {
		t.Fatalf("Delete allowed should be true, got %v", response.Allowed)
	}
}

func TestSetDeleteAllowedInvalid(t *testing.T) {
	a := &app.App{DeleteAllowed: false}

	invalidBody := bytes.NewBufferString("{invalid_json}")

	req := httptest.NewRequest(http.MethodPost, "/set-delete-allowed", invalidBody)
	rec := httptest.NewRecorder()

	handler := http.HandlerFunc(SetDeleteAllowed(a))
	handler.ServeHTTP(rec, req)

	resp := rec.Result()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400 Bad Request Error, got %v", resp.StatusCode)
	}

	if a.DeleteAllowed != false {
		t.Errorf("expected DeleteAllowed=false, got %v", a.DeleteAllowed)
	}
}
