package auth

import (
	"errors"
	"net/http"
	"strings"
)

type RealTokenExtractor struct{}

func NewRealTokenExtractor() *RealTokenExtractor {
	return &RealTokenExtractor{}
}

func (rte *RealTokenExtractor) GetAPIKey(headers http.Header) (string, error) {
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

func (rte *RealTokenExtractor) GetBearerToken(headers http.Header) (string, error) {
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
