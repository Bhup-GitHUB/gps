package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	res := httptest.NewRecorder()

	HealthHandler(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", res.Code)
	}

	var body map[string]string
	err := json.Unmarshal(res.Body.Bytes(), &body)
	if err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}

	if body["status"] != "ok" {
		t.Fatalf("expected ok got %q", body["status"])
	}
}
