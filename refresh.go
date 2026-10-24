package main

import (
	"chirpy/internal/auth"
	"database/sql"
	"errors"
	"net/http"
	"time"
)

func (cfg *apiConfig) handlerRefreshToken(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Token string `json:"token"`
	}
	bearerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "problem parsing bearer token", err)
		return
	}

	userId, err := cfg.db.GetUserFromRefreshToken(r.Context(), bearerToken)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Couldn't get user from refresh token", err)
	}

	newToken, err := auth.MakeJWT(userId, cfg.secretKey, 1*time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldn't make new JWT", err)
	}

	respondWithJSON(w, http.StatusOK, response{
		Token: newToken,
	})
}

func (cfg *apiConfig) handlerRevokeRefreshToken(w http.ResponseWriter, r *http.Request) {
	bearerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "problem parsing bearer token", err)
		return
	}

	if err := cfg.db.RevokeRefreshToken(r.Context(), bearerToken); err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldn't revoke refresh token", err)
	}

	w.WriteHeader(http.StatusNoContent)
}
