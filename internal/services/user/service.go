package user

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/seanhuebl/unity-wealth/internal/database"
	"github.com/seanhuebl/unity-wealth/internal/models"
	"github.com/seanhuebl/unity-wealth/internal/services/auth"
	"go.uber.org/zap"
)

type UserService struct {
	userQueries database.UserQuerier
	pwdHasher   auth.PasswordHasher
	logger      *zap.Logger
}

func NewUserService(userQueries database.UserQuerier, pwdHasher auth.PasswordHasher, logger *zap.Logger) *UserService {
	return &UserService{
		userQueries: userQueries,
		pwdHasher:   pwdHasher,
		logger:      logger,
	}
}

func (u *UserService) SignUp(ctx context.Context, input SignUpInput) error {
	if !models.IsValidEmail(input.Email) {
		return auth.ErrInvalidEmail
	}
	if err := models.ValidatePassword(input.Password); err != nil {
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
