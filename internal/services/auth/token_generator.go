package auth

import (
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type RealTokenGenerator struct {
	tokenSecret     string
	tokenTypeAccess TokenType
}

func NewRealTokenGenerator(tokenSecret string, tokenTypeAccess TokenType) *RealTokenGenerator {
	return &RealTokenGenerator{
		tokenSecret:     tokenSecret,
		tokenTypeAccess: tokenTypeAccess,
	}
}
func (rtg *RealTokenGenerator) MakeJWT(userID uuid.UUID, expiresIn time.Duration) (string, error) {
	if rtg.tokenSecret == "" {
		return "", errors.New("tokenSecret must not be empty")
	}
	if expiresIn <= 0 {
		return "", errors.New("expiresIn must be positive")
	}
	signingKey := []byte(rtg.tokenSecret)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    string(rtg.tokenTypeAccess),
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
		Subject:   userID.String(),
	})
	return token.SignedString(signingKey)
}

func (rtg *RealTokenGenerator) ValidateJWT(tokenString string) (*jwt.RegisteredClaims, error) {
	// Create an instance of RegisteredClaims to hold the parsed token claims.
	var claims jwt.RegisteredClaims

	// Parse the token using the claims instance.
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(rtg.tokenSecret), nil
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

	if claims.Issuer != string(rtg.tokenTypeAccess) {
		return nil, errors.New("invalid issuer")
	}

	if _, err := uuid.Parse(claims.Subject); err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	return &claims, nil
}

func (rtg *RealTokenGenerator) MakeRefreshToken() (string, error) {
	token := make([]byte, 32)
	_, err := RandReader(token)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(token), nil
}
