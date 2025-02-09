package interfaces

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type AuthInterface interface {
	GetAPIKey(headers http.Header) (string, error)
	HashPassword(password string) (string, error)
	CheckPasswordHash(password, hash string) error
	MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error)
	ValidateJWT(tokenString, tokenSecret string) (*jwt.RegisteredClaims, error)
	GetBearerToken(headers http.Header) (string, error)
	MakeRefreshToken() (string, error)
	ValidatePassword(password string) error
}
