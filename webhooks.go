package main

import (
	"chirpy/internal/auth"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerPolkaWebhooks(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Event string `json:"event"`
		Data  struct {
			UserID uuid.UUID `json:"user_id"`
		}
	}
	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if apiKey != cfg.polkaKey {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusBadRequest, "couldn't decode request body", err)
		return
	}

	if params.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if err := cfg.db.UpgradeUserToChirpyRed(r.Context(), params.Data.UserID); err != nil {
		respondWithError(w, http.StatusNotFound, "User not found: "+params.Data.UserID.String(), err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
