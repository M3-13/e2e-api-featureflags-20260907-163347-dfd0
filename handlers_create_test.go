package main

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func newCreateMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /flags", postFlags)
	return mux
}

func TestPostFlagsValid(t *testing.T) {
	store = NewFlagStore()
	mux := newCreateMux()

	body := `{"key":"feature-x","enabled":true,"description":"desc","rollout_percent":50}`
	req := httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(body))
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rec.Code)
	}

	got, ok := store.Get("feature-x")
	if !ok {
		t.Fatal("expected flag to be created")
	}
	if !got.Enabled || got.Description != "desc" || got.RolloutPercent != 50 {
		t.Fatalf("unexpected flag: %+v", got)
	}
}

func TestPostFlagsValidDefaults(t *testing.T) {
	store = NewFlagStore()
	mux := newCreateMux()

	body := `{"key":"feature-y","enabled":false}`
	req := httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(body))
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", rec.Code)
	}

	got, ok := store.Get("feature-y")
	if !ok {
		t.Fatal("expected flag to be created")
	}
	if got.Enabled || got.RolloutPercent != 0 {
		t.Fatalf("unexpected flag: %+v", got)
	}
}

func TestPostFlagsDuplicate(t *testing.T) {
	store = NewFlagStore()
	mux := newCreateMux()

	body := `{"key":"dup","enabled":true}`
	req1 := httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(body))
	rec1 := httptest.NewRecorder()
	mux.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusCreated {
		t.Fatalf("expected first create 201, got %d", rec1.Code)
	}

	req2 := httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(body))
	rec2 := httptest.NewRecorder()
	mux.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusConflict {
		t.Fatalf("expected status 409, got %d", rec2.Code)
	}
	if !strings.Contains(rec2.Body.String(), `"error"`) {
		t.Fatalf("expected JSON error object, got %q", rec2.Body.String())
	}
}

func TestPostFlagsEmptyKey(t *testing.T) {
	store = NewFlagStore()
	mux := newCreateMux()

	body := `{"key":"","enabled":true}`
	req := httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(body))
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestPostFlagsRolloutOutOfRange(t *testing.T) {
	store = NewFlagStore()
	mux := newCreateMux()

	for _, rp := range []int{-1, 101} {
		body := fmtBody(rp)
		req := httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(body))
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("rollout_percent %d: expected 400, got %d", rp, rec.Code)
		}
	}
}

func TestPostFlagsMissingEnabled(t *testing.T) {
	store = NewFlagStore()
	mux := newCreateMux()

	body := `{"key":"no-enabled"}`
	req := httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(body))
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestPostFlagsBodyTooLarge(t *testing.T) {
	store = NewFlagStore()
	mux := newCreateMux()

	body := `{"key":"big","enabled":true,"description":"` + strings.Repeat("a", maxBodyBytes) + `"}`
	req := httptest.NewRequest(http.MethodPost, "/flags", strings.NewReader(body))
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge && rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 or 413, got %d", rec.Code)
	}
}

func fmtBody(rollout int) string {
	return `{"key":"r","enabled":true,"rollout_percent":` + strconv.Itoa(rollout) + `}`
}
