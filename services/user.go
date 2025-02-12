package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/seanhuebl/unity-wealth/internal/database"
	"github.com/seanhuebl/unity-wealth/internal/interfaces"
	"github.com/seanhuebl/unity-wealth/models"
)

type UserService struct {
	queries     interfaces.Querier
	authService interfaces.AuthInterface
}

func NewUserService(queries interfaces.Querier, authSvc interfaces.AuthInterface) *UserService {
	return &UserService{
		queries:     queries,
		authService: authSvc,
	}
}

func (u *UserService) SignUp(ctx context.Context, input models.SignUpInput) error {
	if !models.IsValidEmail(input.Email) {
		return fmt.Errorf("invalid email")
	}
	if err := u.authService.ValidatePassword(input.Password); err != nil {
		return fmt.Errorf("invalid password: %w", err)
	}
	hashedPW, err := u.authService.HashPassword(input.Password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	newID := uuid.New()

	if err = u.queries.CreateUser(ctx, database.CreateUserParams{
		ID: newID.String(),
		HashedPassword: hashedPW,
	}); err != nil {
		return fmt.Errorf("unable to create user: %w", err)
	}
	return nil
}
