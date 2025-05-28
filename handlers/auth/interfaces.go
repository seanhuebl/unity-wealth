package auth

import (
	"context"
	"github.com/seanhuebl/unity-wealth/internal/models"
)

type AuthService interface {
	Login(ctx context.Context, input models.LoginInput) (models.LoginResponse, error)
}
