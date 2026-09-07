package main

import (
	"errors"
	"net/http"
)

func postFlags(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Key            string `json:"key"`
		Enabled        *bool  `json:"enabled"`
		Description    string `json:"description"`
		RolloutPercent *int   `json:"rollout_percent"`
	}

	if err := decodeJSON(w, r, &input); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			writeError(w, http.StatusRequestEntityTooLarge, "request body too large")
			return
		}
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	if input.Key == "" {
		writeError(w, http.StatusBadRequest, "key is required")
		return
	}
	if input.Enabled == nil {
		writeError(w, http.StatusBadRequest, "enabled is required")
		return
	}

	rollout := 0
	if input.RolloutPercent != nil {
		if *input.RolloutPercent < 0 || *input.RolloutPercent > 100 {
			writeError(w, http.StatusBadRequest, "rollout_percent must be between 0 and 100")
			return
		}
		rollout = *input.RolloutPercent
	}

	flag := Flag{
		Key:            input.Key,
		Enabled:        *input.Enabled,
		Description:    input.Description,
		RolloutPercent: rollout,
	}

	if err := store.Create(flag); err != nil {
		if errors.Is(err, ErrExists) {
			writeError(w, http.StatusConflict, "flag already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to create flag")
		return
	}

	writeJSON(w, http.StatusCreated, flag)
}
