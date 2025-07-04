package routes

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type readyResponse struct {
	Ready bool `json:"ready"`
}

func TestHealthHandler_GetRequest_OkResponse(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health/readiness", nil)
	rec := httptest.NewRecorder()

	handler := handleReady(context.Background(), nil)
	handler.ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", res.StatusCode)
	}

	var decoded readyResponse
	err := json.NewDecoder(res.Body).Decode(&decoded)
	if err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}

	if !decoded.Ready {
		t.Errorf("expected ready=true, got ready=%v", decoded.Ready)
	}
}

func TestHealthHandler_PostRequest_InvalidHttpMethod(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/health/readiness", nil)
	rec := httptest.NewRecorder()

	handler := handleReady(context.Background(), nil)
	handler.ServeHTTP(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf(
			"expected status 405, got %d",
			res.StatusCode,
		)
	}
}
