package user

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/seanhuebl/unity-wealth/internal/database"
	"github.com/seanhuebl/unity-wealth/internal/services/auth"
)

type UserService struct {
	userQueries database.UserQuerier
	pwdHasher   auth.PasswordHasher
}

func NewUserService(userQueries database.UserQuerier, pwdHasher auth.PasswordHasher) *UserService {
	return &UserService{
		userQueries: userQueries,
		pwdHasher:   pwdHasher,
	}
}

func (u *UserService) SignUp(ctx context.Context, input SignUpInput) error {
	if !auth.IsValidEmail(input.Email) {
		return auth.ErrInvalidEmail
	}
	if err := auth.ValidatePassword(input.Password); err != nil {
		return fmt.Errorf("%w, %v", auth.ErrInvalidPassword, err)
	}
	hashedPW, err := u.pwdHasher.HashPassword(input.Password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	newID := uuid.New()

	if err = u.userQueries.CreateUser(ctx, database.CreateUserParams{
		ID:             newID.String(),
		Email:          input.Email,
		HashedPassword: hashedPW,
	}); err != nil {
		return fmt.Errorf("unable to create user: %w", err)
	}
	return nil
}