package user

import (
	"context"
	"fmt"
	"regexp"

	"github.com/google/uuid"
	"github.com/seanhuebl/unity-wealth/internal/database"
	"github.com/seanhuebl/unity-wealth/internal/interfaces"
	"github.com/seanhuebl/unity-wealth/internal/services/auth"
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

var (
	uppercaseRegex = regexp.MustCompile(`[A-Z]`)
	lowercaseRegex = regexp.MustCompile(`[a-z]`)
	digitRegex     = regexp.MustCompile(`\d`)
	specialRegex   = regexp.MustCompile(`[!@#\$%\^&\*\(\)_\+\-=\[\]\{\};':"\\|,.<>\/?]`)
)

func (u *UserService) SignUp(ctx context.Context, input SignUpInput) error {
	if !auth.IsValidEmail(input.Email) {
		return fmt.Errorf("invalid email")
	}
	if err := validatePassword(input.Password); err != nil {
		return fmt.Errorf("invalid password: %w", err)
	}
	hashedPW, err := u.authService.HashPassword(input.Password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	newID := uuid.New()

	if err = u.queries.CreateUser(ctx, database.CreateUserParams{
		ID:             newID.String(),
		Email:          input.Email,
		HashedPassword: hashedPW,
	}); err != nil {
		return fmt.Errorf("unable to create user: %w", err)
	}
	return nil
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}
	if !uppercaseRegex.MatchString(password) {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}
	if !lowercaseRegex.MatchString(password) {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}
	if !digitRegex.MatchString(password) {
		return fmt.Errorf("password must contain at least one digit")
	}
	if !specialRegex.MatchString(password) {
		return fmt.Errorf("password must contain at least one special character")
	}
	return nil
}
