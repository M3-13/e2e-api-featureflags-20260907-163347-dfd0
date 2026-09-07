package main

import "net/http"

func putFlag(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	var body struct {
		Enabled        *bool   `json:"enabled"`
		Description    *string `json:"description"`
		RolloutPercent *int    `json:"rollout_percent"`
	}
	if err := decodeJSON(w, r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	flag, ok := store.Get(key)
	if !ok {
		writeError(w, http.StatusNotFound, "flag not found")
		return
	}

	if body.Enabled != nil {
		flag.Enabled = *body.Enabled
	}
	if body.Description != nil {
		flag.Description = *body.Description
	}
	if body.RolloutPercent != nil {
		if *body.RolloutPercent < 0 || *body.RolloutPercent > 100 {
			writeError(w, http.StatusBadRequest, "rollout_percent must be between 0 and 100")
			return
		}
		flag.RolloutPercent = *body.RolloutPercent
	}

	updated, ok := store.Update(key, flag)
	if !ok {
		writeError(w, http.StatusNotFound, "flag not found")
		return
	}
	writeJSON(w, http.StatusOK, updated)
}
