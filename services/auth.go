package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	tokenTypeAccess TokenType
	tokenSecret string
}

func NewAuthService(tokenType, tokenSecret string) *AuthService {
	return &AuthService{
		tokenTypeAccess: TokenType(tokenType),
		tokenSecret: tokenSecret,
	}
}

type TokenType string

var ErrNoAuthHeaderIncluded = errors.New("no authorization header included")
var RandReader = rand.Read

func (a *AuthService) GetAPIKey(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", ErrNoAuthHeaderIncluded
	}
	splitAuth := strings.Split(authHeader, " ")
	if len(splitAuth) < 2 || splitAuth[0] != "ApiKey" {
		return "", errors.New("malformed authorization header")
	}
	if splitAuth[1] == "" {
		return "", errors.New("malformed authorization header")
	}
	return splitAuth[1], nil
}

func (a *AuthService) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (a *AuthService) CheckPasswordHash(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func (a *AuthService) MakeJWT(userID uuid.UUID, expiresIn time.Duration) (string, error) {
	if a.tokenSecret == "" {
		return "", errors.New("tokenSecret must not be empty")
	}
	if expiresIn <= 0 {
		return "", errors.New("expiresIn must be positive")
	}
	signingKey := []byte(a.tokenSecret)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    string(a.tokenTypeAccess),
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
		Subject:   userID.String(),
	})
	return token.SignedString(signingKey)
}

func (a *AuthService) ValidateJWT(tokenString string) (*jwt.RegisteredClaims, error) {
	// Create an instance of RegisteredClaims to hold the parsed token claims.
	var claims jwt.RegisteredClaims

	// Parse the token using the claims instance.
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(a.tokenSecret), nil
	})
	if err != nil {
		return nil, err
	}

	// Ensure the token is valid.
	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	now := time.Now().Unix()
	if claims.ExpiresAt != nil && claims.ExpiresAt.Unix() < now {
		return nil, fmt.Errorf("token expired")
	}
	if claims.NotBefore != nil && claims.NotBefore.Unix() > now {
		return nil, fmt.Errorf("token not valid yet")
	}

	// Check the issuer (this example assumes you want the issuer to equal TokenTypeAccess).
	if claims.Issuer != string(a.tokenTypeAccess) {
		return nil, errors.New("invalid issuer")
	}

	// Optionally, if you expect the Subject (user ID) to be a valid UUID, you can verify that.
	if _, err := uuid.Parse(claims.Subject); err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// Return the complete claims struct.
	return &claims, nil
}

func (a *AuthService) GetBearerToken(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", ErrNoAuthHeaderIncluded
	}
	splitAuth := strings.Split(authHeader, " ")
	if len(splitAuth) < 2 || splitAuth[0] != "Bearer" {
		return "", errors.New("malformed authorization header")
	}
	if splitAuth[1] == "" {
		return "", errors.New("malformed authorization header")
	}
	return splitAuth[1], nil
}

func (a *AuthService) MakeRefreshToken() (string, error) {
	token := make([]byte, 32)
	_, err := RandReader(token)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(token), nil
}

func (a *AuthService) ValidatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}
	if matched, _ := regexp.MatchString(`[A-Z]`, password); !matched {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}
	if matched, _ := regexp.MatchString(`[a-z]`, password); !matched {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}
	if matched, _ := regexp.MatchString(`\d`, password); !matched {
		return fmt.Errorf("password must contain at least one digit")
	}
	if matched, _ := regexp.MatchString(`[!@#\$%\^&\*\(\)_\+\-=\[\]\{\};':"\\|,.<>\/?]`, password); !matched {
		return fmt.Errorf("password must contain at least one special character")
	}
	return nil
}
