package main

import "net/http"

func listFlags(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, store.List())
}
