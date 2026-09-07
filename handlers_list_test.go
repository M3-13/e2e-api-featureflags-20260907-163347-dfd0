package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListFlagsEmpty(t *testing.T) {
	prev := store
	store = NewFlagStore()
	defer func() { store = prev }()

	req := httptest.NewRequest(http.MethodGet, "/flags", nil)
	rec := httptest.NewRecorder()

	listFlags(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var flags []Flag
	if err := json.NewDecoder(rec.Body).Decode(&flags); err != nil {
		t.Fatalf("expected valid JSON list, got error: %v", err)
	}
	if flags == nil {
		t.Fatal("expected non-nil empty JSON list, got nil")
	}
	if len(flags) != 0 {
		t.Fatalf("expected empty list, got %d flags", len(flags))
	}
}

func TestListFlagsReturnsAll(t *testing.T) {
	prev := store
	store = NewFlagStore()
	defer func() { store = prev }()

	_ = store.Create(Flag{Key: "a", Enabled: true, Description: "first", RolloutPercent: 50})
	_ = store.Create(Flag{Key: "b", Enabled: false, Description: "second"})

	req := httptest.NewRequest(http.MethodGet, "/flags", nil)
	rec := httptest.NewRecorder()

	listFlags(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var flags []Flag
	if err := json.NewDecoder(rec.Body).Decode(&flags); err != nil {
		t.Fatalf("expected valid JSON list, got error: %v", err)
	}
	if len(flags) != 2 {
		t.Fatalf("expected 2 flags, got %d", len(flags))
	}

	byKey := map[string]Flag{}
	for _, f := range flags {
		byKey[f.Key] = f
	}
	if got := byKey["a"]; got != (Flag{Key: "a", Enabled: true, Description: "first", RolloutPercent: 50}) {
		t.Fatalf("flag a mismatch: %+v", got)
	}
	if got := byKey["b"]; got != (Flag{Key: "b", Enabled: false, Description: "second"}) {
		t.Fatalf("flag b mismatch: %+v", got)
	}
}
