package auth

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type PasswordHasher interface {
	HashPassword(password string) (string, error)
	CheckPasswordHash(password, hash string) error
}

type TokenGenerator interface {
	MakeJWT(userID uuid.UUID, expiresIn time.Duration) (string, error)
	ValidateJWT(tokenString string) (*jwt.RegisteredClaims, error)
	MakeRefreshToken() (string, error)
}

type TokenExtractor interface {
	GetAPIKey(headers http.Header) (string, error)
	GetBearerToken(headers http.Header) (string, error)
}
