package user

import (
	"context"

	"github.com/seanhuebl/unity-wealth/internal/services/user"
)

type UserService interface {
	SignUp(ctx context.Context, input user.SignUpInput) error
}
