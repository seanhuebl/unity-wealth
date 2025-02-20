package auth

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/seanhuebl/unity-wealth/internal/constants"
	"github.com/seanhuebl/unity-wealth/internal/database"
	authmocks "github.com/seanhuebl/unity-wealth/internal/mocks/auth"
	dbmocks "github.com/seanhuebl/unity-wealth/internal/mocks/database"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestLogin(t *testing.T) {
	validUserID := uuid.New()
	dummyUserRow := database.GetUserByEmailRow{
		ID:             validUserID.String(),
		HashedPassword: "hashedpassword",
	}
	tests := []struct {
		name                   string
		input                  LoginInput
		expectedResponse       LoginResponse
		loginError             error
		expectedErrorSubstring string
	}{
		{
			name: "login successful",
			input: LoginInput{
				Email:    "user@example.com",
				Password: "password123",
			},
			expectedResponse: LoginResponse{
				UserID:       validUserID,
				JWT:          "JWT",
				RefreshToken: "refresh",
			},
			loginError:             nil,
			expectedErrorSubstring: "",
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			req, _ := http.NewRequest("GET", "/", nil)
			req.Header.Set("X-Device-info", "os=Android; os_version=11; device_type=Mobile; browser=Chrome; browser_version=100.0")
			reqWithCtx := req.WithContext(context.WithValue(req.Context(), constants.RequestKey, req))

			ctx.Request = reqWithCtx
			mockSqlTxQ := dbmocks.NewSqlTxQuerier(t)
			mockUserQ := dbmocks.NewUserQuerier(t)
			mockTokenGen := authmocks.NewTokenGenerator(t)
			mockExtractor := authmocks.NewTokenExtractor(t)
			mockHasher := authmocks.NewPasswordHasher(t)

			dummyTx := &sql.Tx{}

			dummyQueries := dbmocks.NewSqlTransactionalQuerier(t)

			mockUserQ.On("GetUserByEmail", ctx.Request.Context(), tc.input.Email).Return(dummyUserRow, nil)

			mockSqlTxQ.On("BeginTx", ctx.Request.Context(), (*sql.TxOptions)(nil)).Return(dummyTx, nil)
			mockSqlTxQ.On("WithTx", dummyTx).Return(dummyQueries)

			mockHasher.On("HashPassword", "refresh").Return("hashedrefresh")
			mockHasher.On("CheckPasswordHash", tc.input.Password, dummyUserRow.HashedPassword).Return(nil)

			mockTokenGen.On("MakeJWT", validUserID, 15*time.Minute).Return("JWT", nil)
			mockTokenGen.On("MakeRefreshToken").Return("refresh", nil)

			dummyQueries.On("GetDeviceInfoByUser", ctx.Request.Context(), mock.Anything).Return("deviceid", nil)
			dummyQueries.On("CreateDeviceInfo", ctx.Request.Context(), mock.Anything).Return("deviceid", nil)
			dummyQueries.On("RevokeToken", ctx.Request.Context(), mock.Anything).Return(nil)
			dummyQueries.On("CreateRefreshToken", ctx.Request.Context(), mock.Anything).Return(nil)

			svc := NewAuthService(mockSqlTxQ, mockUserQ, mockTokenGen, mockExtractor, mockHasher)
			response, err := svc.Login(ctx.Request.Context(), tc.input)
			if tc.expectedErrorSubstring != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedErrorSubstring)
			} else {
				require.NoError(t, err)
				if diff := cmp.Diff(tc.expectedResponse, response); diff != "" {
					t.Errorf("response mismatch (-want +got)\n%s", diff)
				}
			}
			mockSqlTxQ.AssertExpectations(t)
			mockUserQ.AssertExpectations(t)
			mockTokenGen.AssertExpectations(t)
			mockExtractor.AssertExpectations(t)
			mockHasher.AssertExpectations(t)
		})
	}
}
