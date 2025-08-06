package auth

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type RealPasswordHasher struct{}

func NewRealPwdHasher() *RealPasswordHasher {
	return &RealPasswordHasher{}
}

func (rph *RealPasswordHasher) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("error generating password: %w", err)
	}
	return string(hash), nil
}

func (rph *RealPasswordHasher) CheckPasswordHash(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return errors.Join(ErrPwdHashMismatch, err)
	}
	return nil
}
