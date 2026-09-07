package main

import (
	"hash/fnv"
	"net/http"
)

func evaluateFlag(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	user := r.URL.Query().Get("user")

	if user == "" || len(user) > 255 {
		writeError(w, http.StatusBadRequest, "user required and must not exceed 255 characters")
		return
	}

	flag, ok := store.Get(key)
	if !ok {
		writeError(w, http.StatusNotFound, "flag not found")
		return
	}

	h := fnv.New32a()
	_, _ = h.Write([]byte(key + ":" + user))
	result := int(h.Sum32()%100) < flag.RolloutPercent

	writeJSON(w, http.StatusOK, map[string]bool{"result": result})
}
