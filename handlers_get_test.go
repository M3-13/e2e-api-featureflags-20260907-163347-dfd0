package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetFlagFound(t *testing.T) {
	store.Create(Flag{Key: "feature-x", Enabled: true, Description: "test", RolloutPercent: 50})

	mux := newMux()
	req := httptest.NewRequest(http.MethodGet, "/flags/feature-x", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var got Flag
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if got.Key != "feature-x" || !got.Enabled || got.Description != "test" || got.RolloutPercent != 50 {
		t.Fatalf("unexpected flag body: %+v", got)
	}
}

func TestGetFlagUnknown(t *testing.T) {
	mux := newMux()
	req := httptest.NewRequest(http.MethodGet, "/flags/does-not-exist", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}

	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body["error"] == "" {
		t.Fatalf("expected error field in JSON body, got %+v", body)
	}
}
