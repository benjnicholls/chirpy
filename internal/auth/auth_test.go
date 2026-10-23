package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCheckPasswordHash(t *testing.T) {
	// First, we need to create some hashed passwords for testing
	password1 := "correctPassword123!"
	password2 := "anotherPassword456!"
	hash1, _ := HashPassword(password1)
	hash2, _ := HashPassword(password2)

	tests := []struct {
		name          string
		password      string
		hash          string
		wantErr       bool
		matchPassword bool
	}{
		{
			name:          "Correct password",
			password:      password1,
			hash:          hash1,
			wantErr:       false,
			matchPassword: true,
		},
		{
			name:          "Incorrect password",
			password:      "wrongPassword",
			hash:          hash1,
			wantErr:       false,
			matchPassword: false,
		},
		{
			name:          "Password doesn't match different hash",
			password:      password1,
			hash:          hash2,
			wantErr:       false,
			matchPassword: false,
		},
		{
			name:          "Empty password",
			password:      "",
			hash:          hash1,
			wantErr:       false,
			matchPassword: false,
		},
		{
			name:          "Invalid hash",
			password:      password1,
			hash:          "invalidhash",
			wantErr:       true,
			matchPassword: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, err := CheckPasswordHash(tt.password, tt.hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckPasswordHash() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && match != tt.matchPassword {
				t.Errorf("CheckPasswordHash() expects %v, got %v", tt.matchPassword, match)
			}
		})
	}
}

func TestValidateJWT(t *testing.T) {
	// First, we need to create some hashed passwords for testing
	realSecretKey := "thisisareallysecretkey"
	fakeSecretKey := "thisisareallyfakesecretkey"
	userId, _ := uuid.NewUUID()
	jwt1, _ := MakeJWT(userId, realSecretKey, 5*time.Hour)
	jwt2, _ := MakeJWT(userId, realSecretKey, 0)
	jwt3, _ := MakeJWT(userId, realSecretKey, 1*time.Hour)

	tests := []struct {
		name      string
		id        uuid.UUID
		jwt       string
		secretKey string
		wantErr   bool
	}{
		{
			name:      "Valid token",
			id:        userId,
			jwt:       jwt1,
			secretKey: realSecretKey,
			wantErr:   false,
		},
		{
			name:      "Expired JWT",
			id:        userId,
			jwt:       jwt2,
			secretKey: realSecretKey,
			wantErr:   true,
		},
		{
			name:      "Wrong secret",
			id:        userId,
			jwt:       jwt3,
			secretKey: fakeSecretKey,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := ValidateJWT(tt.jwt, tt.secretKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateJWT() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && id != tt.id {
				t.Errorf("CheckPasswordHash() expects %v, got %v", tt.id, id)
			}
		})
	}
}
