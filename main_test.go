package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthz(t *testing.T) {
	mux := newMux()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if h := rec.Header().Get("Access-Control-Allow-Origin"); h != "" {
		t.Fatalf("expected no Access-Control-Allow-Origin header, got %q", h)
	}
}
