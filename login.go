package main

import (
	"chirpy/internal/auth"
	"chirpy/internal/database"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	type response struct {
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
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

	token, err := auth.MakeJWT(user.ID, cfg.secretKey, 1*time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Problem creating JWT", err)
	}

	refreshToken, err := cfg.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:  auth.MakeRefreshToken(),
		UserID: user.ID,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Problem creating refresh token", err)
	}

	respondWithJSON(w, http.StatusOK, response{
		User: User{
			Id:          user.ID,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
			Email:       user.Email,
			IsChirpyRed: user.IsChirpyRed,
		},
		Token:        token,
		RefreshToken: refreshToken.Token,
	})
}
