package main

import "net/http"

func deleteFlag(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if !store.Delete(key) {
		writeError(w, http.StatusNotFound, "flag not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
