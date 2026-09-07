package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestLogRequestsDoesNotLogUser(t *testing.T) {
	dummy := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})

	old := os.Stdout
	rPipe, wPipe, _ := os.Pipe()
	os.Stdout = wPipe

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/flags/x/evaluate?user=geheim", nil)

	logRequests(dummy).ServeHTTP(rec, req)

	_ = wPipe.Close()
	os.Stdout = old

	out, _ := io.ReadAll(rPipe)
	logLine := string(out)

	if !strings.Contains(logLine, http.MethodGet) {
		t.Errorf("log %q does not contain method", logLine)
	}
	if !strings.Contains(logLine, "/flags/x/evaluate") {
		t.Errorf("log %q does not contain path without query", logLine)
	}
	if strings.Contains(logLine, "?") {
		t.Errorf("log %q contains query string", logLine)
	}
	if !strings.Contains(logLine, "418") {
		t.Errorf("log %q does not contain status code", logLine)
	}
	if strings.Contains(logLine, "geheim") {
		t.Errorf("log %q contains the user parameter", logLine)
	}
}
