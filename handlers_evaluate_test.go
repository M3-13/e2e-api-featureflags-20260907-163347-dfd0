package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func resetStore() {
	store = NewFlagStore()
}

func evalResult(t *testing.T, mux *http.ServeMux, key, user string) (int, bool, bool) {
	t.Helper()
	target := "/flags/" + key + "/evaluate"
	if user != "" {
		target += "?user=" + user
	}
	req := httptest.NewRequest(http.MethodGet, target, nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		return rec.Code, false, false
	}
	var body struct {
		Result bool `json:"result"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unexpected response body %q: %v", rec.Body.String(), err)
	}
	return rec.Code, body.Result, true
}

func TestEvaluateDeterministic(t *testing.T) {
	resetStore()
	if err := store.Create(Flag{Key: "feature", Enabled: true, RolloutPercent: 50}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mux := newMux()

	_, first, _ := evalResult(t, mux, "feature", "user-a")
	for i := 0; i < 10; i++ {
		_, got, _ := evalResult(t, mux, "feature", "user-a")
		if got != first {
			t.Fatalf("expected deterministic result %v, got %v", first, got)
		}
	}
}

func TestEvaluateRolloutZeroAlwaysFalse(t *testing.T) {
	resetStore()
	if err := store.Create(Flag{Key: "off", Enabled: true, RolloutPercent: 0}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mux := newMux()

	for _, user := range []string{"a", "b", "c", "z", "user-with-odd-chars"} {
		code, got, _ := evalResult(t, mux, "off", user)
		if code != http.StatusOK || got {
			t.Fatalf("rollout 0: expected false for user %q, got status %d result %v", user, code, got)
		}
	}
}

func TestEvaluateRolloutHundredAlwaysTrue(t *testing.T) {
	resetStore()
	if err := store.Create(Flag{Key: "on", Enabled: true, RolloutPercent: 100}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mux := newMux()

	for _, user := range []string{"a", "b", "c", "z", "user-with-odd-chars"} {
		code, got, _ := evalResult(t, mux, "on", user)
		if code != http.StatusOK || !got {
			t.Fatalf("rollout 100: expected true for user %q, got status %d result %v", user, code, got)
		}
	}
}

func TestEvaluateMissingOrEmptyUser(t *testing.T) {
	resetStore()
	if err := store.Create(Flag{Key: "feature", Enabled: true, RolloutPercent: 50}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mux := newMux()

	req := httptest.NewRequest(http.MethodGet, "/flags/feature/evaluate", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing user, got %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/flags/feature/evaluate?user=", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for empty user, got %d", rec.Code)
	}
}

func TestEvaluateUserTooLong(t *testing.T) {
	resetStore()
	if err := store.Create(Flag{Key: "feature", Enabled: true, RolloutPercent: 50}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mux := newMux()

	longUser := strings.Repeat("x", 256)
	req := httptest.NewRequest(http.MethodGet, "/flags/feature/evaluate?user="+longUser, nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for overlong user, got %d", rec.Code)
	}
}

func TestEvaluateUnknownKey(t *testing.T) {
	resetStore()
	mux := newMux()

	req := httptest.NewRequest(http.MethodGet, "/flags/nonexistent/evaluate?user=a", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown key, got %d", rec.Code)
	}
}
