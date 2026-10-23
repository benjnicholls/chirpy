package main

import (
	"chirpy/internal/auth"
	"chirpy/internal/database"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Chirp struct {
	Id        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserId    uuid.UUID `json:"user_id"`
}

func cleanProfaneWords(s string) string {
	profaneWords := []string{"kerfuffle", "sharbert", "fornax"}
	words := strings.Split(s, " ")
	cleanedWords := make([]string, 0, len(words))

	for _, word := range words {
		cleanedWord := word
		for _, profaneWord := range profaneWords {
			if profaneWord == strings.ToLower(word) {
				cleanedWord = "****"
			}
		}
		cleanedWords = append(cleanedWords, cleanedWord)
	}

	return strings.Join(cleanedWords, " ")
}

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body   string    `json:"body"`
		UserId uuid.UUID `json:"user_id"`
	}
	type response struct {
		Chirp
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't decode request parameters", err)
		return
	}
	bearerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "problem parsing bearer token", err)
	}
	subjectId, err := auth.ValidateJWT(bearerToken, cfg.secretKey)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "problem validating token", err)
	}

	if subjectId != params.UserId {
		respondWithError(w, http.StatusUnauthorized, "invalid token", nil)
	}

	if len(params.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}
	cleanedWords := cleanProfaneWords(params.Body)

	chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   cleanedWords,
		UserID: params.UserId,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create chirp", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, response{
		Chirp: Chirp{
			Id:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserId:    chirp.UserID,
		},
	})
}

func (cfg *apiConfig) handlerGetAllChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.db.GetChirps(r.Context())
	var response []Chirp
	for _, chirp := range chirps {
		response = append(response, Chirp{
			Id:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserId:    chirp.UserID,
		})
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve chirps", err)
		return
	}

	respondWithJSON(w, http.StatusOK, response)
}

func (cfg *apiConfig) handlerGetSingleChirp(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Chirp
	}
	chirpId := r.PathValue("id")
	if chirpId == "" {
		respondWithError(w, http.StatusBadRequest, "cannot get chirp with no id", nil)
		return
	}

	id, err := uuid.Parse(chirpId)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't parse string as uuid", err)
		return
	}

	chirp, err := cfg.db.GetChirp(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Couldn't find chirp with id: " + id.String()))
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve chirp", err)
		return
	}

	respondWithJSON(w, http.StatusOK, response{
		Chirp: Chirp{
			Id:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserId:    chirp.UserID,
		},
	})
}
