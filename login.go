package main

import (
	"chirpy/internal/auth"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		ExpiresInSeconds *int   `json:"expires_in_seconds"`
	}
	type response struct {
		Token string `json:"token"`
		User
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
	}

	user, err := cfg.db.FindUserByEmail(r.Context(), params.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Couldn't find chirp with email: " + params.Email))
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve chirp", err)
		return
	}

	match, err := auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed checking password hash", err)
		return
	}
	if !match {
		respondWithError(w, http.StatusUnauthorized, "Wrong password or email", nil)
		return
	}

	var expirationTime time.Duration
	if params.ExpiresInSeconds != nil {
		expirationTime = time.Duration(*params.ExpiresInSeconds) * time.Second
	} else {
		expirationTime = 1 * time.Hour
	}

	token, err := auth.MakeJWT(user.ID, cfg.secretKey, expirationTime)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Problem creating JWT", err)
	}

	respondWithJSON(w, http.StatusOK, response{
		User: User{
			Id:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
		},
		Token: token,
	})
}
