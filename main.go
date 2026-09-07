package main

import (
	"log"
	"net/http"
)

func healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func newMux() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /flags", postFlags)
	mux.HandleFunc("GET /flags", listFlags)
	mux.HandleFunc("GET /flags/{key}", getFlag)
	mux.HandleFunc("PUT /flags/{key}", putFlag)
	mux.HandleFunc("DELETE /flags/{key}", deleteFlag)
	mux.HandleFunc("GET /flags/{key}/evaluate", evaluateFlag)
	mux.HandleFunc("GET /healthz", healthz)

	mux.HandleFunc("/flags", methodNotAllowed)
	mux.HandleFunc("/flags/{key}", methodNotAllowed)
	mux.HandleFunc("/flags/{key}/evaluate", methodNotAllowed)
	mux.HandleFunc("/healthz", methodNotAllowed)

	return mux
}

func main() {
	handler := logRequests(newMux())
	log.Println("listening on :8080")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatal(err)
	}
}
