package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
)

func putFlagRequest(t *testing.T, key string, body string) *httptest.ResponseRecorder {
	t.Helper()
	mux := newMux()
	req := httptest.NewRequest(http.MethodPut, "/flags/"+key, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

func TestPutFlagUpdatesFieldsAndKeepsKey(t *testing.T) {
	store = NewFlagStore()
	_ = store.Create(Flag{Key: "k", Enabled: false, Description: "old", RolloutPercent: 10})

	rec := putFlagRequest(t, "k", `{"enabled":true,"description":"new","rollout_percent":50}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var got Flag
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unexpected decode error: %v", err)
	}
	if got.Key != "k" {
		t.Fatalf("key changed: got %q, want %q", got.Key, "k")
	}
	if !got.Enabled || got.Description != "new" || got.RolloutPercent != 50 {
		t.Fatalf("unexpected flag: %+v", got)
	}
}

func TestPutFlagPartialUpdateLeavesFieldsUnchanged(t *testing.T) {
	store = NewFlagStore()
	_ = store.Create(Flag{Key: "k", Enabled: false, Description: "old", RolloutPercent: 10})

	rec := putFlagRequest(t, "k", `{"enabled":true}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var got Flag
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("unexpected decode error: %v", err)
	}
	if !got.Enabled {
		t.Fatal("expected enabled to be true")
	}
	if got.Description != "old" || got.RolloutPercent != 10 {
		t.Fatalf("unchanged fields modified: %+v", got)
	}
}

func TestPutFlagUnknownKey(t *testing.T) {
	store = NewFlagStore()

	rec := putFlagRequest(t, "missing", `{"enabled":true}`)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
	var body map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body["error"] == "" {
		t.Fatal("expected JSON error object")
	}
}

func TestPutFlagInvalidRolloutPercent(t *testing.T) {
	store = NewFlagStore()
	_ = store.Create(Flag{Key: "k", Enabled: false})

	for _, pct := range []int{-1, 101} {
		body, _ := json.Marshal(map[string]int{"rollout_percent": pct})
		rec := putFlagRequest(t, "k", string(body))
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for rollout_percent %d, got %d", pct, rec.Code)
		}
	}
}

func TestPutFlagInvalidJSON(t *testing.T) {
	store = NewFlagStore()
	_ = store.Create(Flag{Key: "k"})

	rec := putFlagRequest(t, "k", `not json`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestPutFlagConcurrentDeleteNeverReturnsEmptyFlag(t *testing.T) {
	store = NewFlagStore()
	mux := newMux()

	var wg sync.WaitGroup
	var mu sync.Mutex
	var failures []string

	for i := 0; i < 200; i++ {
		wg.Add(2)

		go func() {
			defer wg.Done()
			_ = store.Create(Flag{Key: "k", Enabled: false, Description: "old", RolloutPercent: 10})
			req := httptest.NewRequest(http.MethodPut, "/flags/k", strings.NewReader(`{"enabled":true}`))
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)

			switch rec.Code {
			case http.StatusOK:
				var f Flag
				if err := json.Unmarshal(rec.Body.Bytes(), &f); err != nil || f.Key != "k" {
					mu.Lock()
					failures = append(failures, "200 with empty/invalid flag: "+rec.Body.String())
					mu.Unlock()
				}
			case http.StatusNotFound:
			default:
				mu.Lock()
				failures = append(failures, "unexpected status "+strconv.Itoa(rec.Code))
				mu.Unlock()
			}
		}()

		go func() {
			defer wg.Done()
			store.Delete("k")
		}()
	}

	wg.Wait()

	if len(failures) > 0 {
		t.Fatalf("handler returned unexpected response under concurrent delete: %v", failures)
	}
}
