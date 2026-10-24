package main

import (
	"chirpy/internal/auth"
	"net/http"

	"github.com/google/uuid"
)

func authorizeUser(headers http.Header, secretKey string) (uuid.UUID, error) {
	token, err := auth.GetBearerToken(headers)
	if err != nil {
		return uuid.Nil, err
	}

	id, err := auth.ValidateJWT(token, secretKey)
	if err != nil {
		return uuid.Nil, err
	}

	return id, nil
}
