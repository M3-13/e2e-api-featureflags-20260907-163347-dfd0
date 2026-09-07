package main

import "net/http"

func getFlag(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	flag, ok := store.Get(key)
	if !ok {
		writeError(w, http.StatusNotFound, "flag not found")
		return
	}
	writeJSON(w, http.StatusOK, flag)
}
