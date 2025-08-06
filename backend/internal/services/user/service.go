package user

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/seanhuebl/unity-wealth/internal/constants"
	"github.com/seanhuebl/unity-wealth/internal/database"
	"github.com/seanhuebl/unity-wealth/internal/models"
	"github.com/seanhuebl/unity-wealth/internal/sentinels"
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
	start := time.Now()
	reqID, _ := ctx.Value(constants.RequestIDKey).(string)
	preSuccess := u.logger.With(
		zap.String("request_id", reqID),
		zap.String("email", input.Email),
	)
	preSuccess.Info(evtSignUpAttempt)

	if !models.IsValidEmail(input.Email) {
		preSuccess.Warn(evtSignUpInvalidEmail)
		return sentinels.ErrInvalidEmail
	}

	if err := models.ValidatePassword(input.Password); err != nil {
		wrapped := fmt.Errorf("signup: validate pwd: %w: %v", sentinels.ErrInvalidPassword, err)
		preSuccess.Warn(evtSignUpInvalidPassword,
			zap.Error(wrapped),
		)
		return wrapped
	}

	preSuccess.Info(evtSignUpValidInput)

	hashedPW, err := u.pwdHasher.HashPassword(input.Password)
	if err != nil {
		wrapped := fmt.Errorf("signup: hash pwd: %w", err)
		preSuccess.Error(evtSignUpPwdHashFailed,
			zap.Error(wrapped),
		)
		return wrapped
	}

	newID := uuid.New()

	logger := u.logger.With(
		zap.String("request_id", reqID),
		zap.String("user_id", newID.String()),
	)

	dbStart := time.Now()
	if err = u.userQueries.CreateUser(ctx, database.CreateUserParams{
		ID:             newID,
		Email:          input.Email,
		HashedPassword: hashedPW,
	}); err != nil {
		wrapped := fmt.Errorf("signup: %w: %v", sentinels.ErrDBExecFailed, err)
		logger.Error(evtSignUpUserInsertDBFailed,
			zap.Error(wrapped),
		)
		return wrapped
	}
	dbDuration := time.Since(dbStart)
	totalDuration := time.Since(start)

	logger.Info(evtSignUpSuccess,
		zap.Duration("db_duration_ms", dbDuration),
		zap.Duration("total_duration_ms", totalDuration),
	)
	return nil
}
