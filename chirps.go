package main

import (
	"chirpy/internal/database"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"sort"
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

func getUuidByRequestPath(r *http.Request) (uuid.UUID, error) {
	stringId := r.PathValue("id")
	if stringId == "" {
		return uuid.Nil, errors.New("cannot find id parameter in request path")
	}

	id, err := uuid.Parse(stringId)
	if err != nil {
		return uuid.Nil, errors.New("unable to parse string id as uuid")
	}

	return id, nil
}

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}
	type response struct {
		Chirp
	}
	userId, err := authorizeUser(r.Header, cfg.secretKey)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't decode request parameters", err)
		return
	}

	if len(params.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}
	cleanedWords := cleanProfaneWords(params.Body)

	chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   cleanedWords,
		UserID: userId,
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
	searchParams := r.URL.Query()
	authorQuery := searchParams.Get("author_id")
	sortQuery := searchParams.Get("sort")

	chirps, err := cfg.db.GetChirps(r.Context())
	var response []Chirp
	for _, chirp := range chirps {
		if authorQuery != "" && authorQuery != chirp.UserID.String() {
			continue
		}
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
	if sortQuery == "desc" {
		sort.Slice(response, func(i, j int) bool { return response[i].CreatedAt.After(response[j].CreatedAt) })
	}

	respondWithJSON(w, http.StatusOK, response)
}

func (cfg *apiConfig) handlerGetSingleChirp(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Chirp
	}

	id, err := getUuidByRequestPath(r)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "couldn't get id", err)
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

func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	userId, err := authorizeUser(r.Header, cfg.secretKey)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	chirpId, err := getUuidByRequestPath(r)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "couldn't get id", err)
		return
	}

	chirp, err := cfg.db.GetChirp(r.Context(), chirpId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Couldn't find chirp with id: " + chirpId.String()))
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve chirp", err)
		return
	}

	if chirp.UserID != userId {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if err := cfg.db.DeleteChirp(r.Context(), database.DeleteChirpParams{
		ID:     chirpId,
		UserID: userId,
	}); err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldn't delete chirp", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
