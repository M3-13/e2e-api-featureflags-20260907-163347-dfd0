package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDeleteFlagRemovesFlag(t *testing.T) {
	key := "delete-test-flag"
	if err := store.Create(Flag{Key: key, Enabled: true}); err != nil {
		t.Fatalf("unexpected create error: %v", err)
	}

	mux := newMux()
	req := httptest.NewRequest(http.MethodDelete, "/flags/"+key, nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rec.Code)
	}
	if rec.Body.Len() != 0 {
		t.Fatalf("expected empty body, got %q", rec.Body.String())
	}
	if _, ok := store.Get(key); ok {
		t.Fatal("expected flag to be removed from the store")
	}
}

func TestDeleteFlagUnknownKey(t *testing.T) {
	mux := newMux()
	req := httptest.NewRequest(http.MethodDelete, "/flags/does-not-exist", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", ct)
	}
}
