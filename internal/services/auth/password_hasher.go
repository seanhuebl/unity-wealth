package auth

import "golang.org/x/crypto/bcrypt"

type RealPasswordHasher struct{}

func NewRealPwdHasher() *RealPasswordHasher {
	return &RealPasswordHasher{}
}

func (rph *RealPasswordHasher) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (rph *RealPasswordHasher) CheckPasswordHash(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
