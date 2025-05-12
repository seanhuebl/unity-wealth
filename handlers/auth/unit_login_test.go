package auth_test

import (
	"bytes"
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	authhttp "github.com/seanhuebl/unity-wealth/handlers/auth"
	"github.com/seanhuebl/unity-wealth/internal/constants"
	"github.com/seanhuebl/unity-wealth/internal/database"
	authmocks "github.com/seanhuebl/unity-wealth/internal/mocks/auth"
	dbmocks "github.com/seanhuebl/unity-wealth/internal/mocks/database"
	"github.com/seanhuebl/unity-wealth/internal/models"
	"github.com/seanhuebl/unity-wealth/internal/services/auth"
	"github.com/seanhuebl/unity-wealth/internal/testhelpers"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestLogin(t *testing.T) {
	validUserID := uuid.New()
	validDeviceID := uuid.New()
	dummyUserRow := database.GetUserByEmailRow{
		ID:             validUserID.String(),
		HashedPassword: "hashedpassword",
	}
	tests := []struct {
		name               string
		deviceFound        bool
		reqBody            string
		expErrSubstr       string
		expectedStatusCode int
		expectedResponse   map[string]interface{}
	}{
		{
			name:               "invalid request body",
			reqBody:            `{"email": "user@example.com", "password": "Validpass1!"`,
			expErrSubstr:       "invalid request body",
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse: map[string]interface{}{
				"data": map[string]interface{}{
					"error": "invalid request body",
				},
			},
		},
		{
			name:               "service error: invalid email",
			reqBody:            `{"email": "user", "password": "Validpass1!"}`,
			expErrSubstr:       "login failed",
			expectedStatusCode: http.StatusUnauthorized,
			expectedResponse: map[string]interface{}{
				"data": map[string]interface{}{
					"error": "login failed",
				},
			},
		},
		{
			name:               "successful login",
			reqBody:            `{"email": "user@example.com", "password": "Validpass1!"}`,
			expErrSubstr:       "",
			expectedStatusCode: http.StatusOK,
			expectedResponse: map[string]interface{}{
				"data": map[string]interface{}{
					"message": "login successful",
				},
			},
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			req, _ := http.NewRequest("POST", "/login", bytes.NewBufferString(tc.reqBody))
			req.Header.Set("X-Device-Info", "os=Android; os_version=11; device_type=Mobile; browser=Chrome; browser_version=100.0")
			req.Header.Set("Content-Type", "application/json")
			reqWithCtx := req.WithContext(context.WithValue(req.Context(), constants.RequestKey, req))

			ctx.Request = reqWithCtx
			mockSqlTxQ := dbmocks.NewSqlTxQuerier(t)
			mockUserQ := dbmocks.NewUserQuerier(t)
			mockTokenGen := authmocks.NewTokenGenerator(t)
			mockExtractor := authmocks.NewTokenExtractor(t)
			mockHasher := authmocks.NewPasswordHasher(t)

			db, sqlMock, err := sqlmock.New()
			require.NoError(t, err)

			sqlMock.ExpectBegin()
			sqlMock.ExpectCommit()

			dummyTx, err := db.Begin()
			require.NoError(t, err)

			dummyQueries := dbmocks.NewSqlTransactionalQuerier(t)

			expectedInput := models.LoginInput{
				Email:    "user@example.com",
				Password: "Validpass1!",
			}
			
			if strings.Contains(tc.name, "invalid") {
				return
			} else {
				mockUserQ.On("GetUserByEmail", ctx.Request.Context(), expectedInput.Email).Return(dummyUserRow, nil)

				mockSqlTxQ.On("BeginTx", ctx.Request.Context(), (*sql.TxOptions)(nil)).Return(dummyTx, nil).Once()
				mockSqlTxQ.On("WithTx", dummyTx).Return(dummyQueries)

				mockHasher.On("HashPassword", "refresh").Return("hashedrefresh", nil)
				mockHasher.On("CheckPasswordHash", expectedInput.Password, dummyUserRow.HashedPassword).Return(nil)

				mockTokenGen.On("MakeJWT", validUserID, 15*time.Minute).Return("JWT", nil)
				mockTokenGen.On("MakeRefreshToken").Return("refresh", nil)

				if tc.deviceFound {
					dummyQueries.On("GetDeviceInfoByUser", ctx.Request.Context(), mock.Anything).Return(validDeviceID.String(), nil)
					dummyQueries.On("RevokeToken", ctx.Request.Context(), mock.Anything).Return(nil)
				} else {
					dummyQueries.On("CreateDeviceInfo", ctx.Request.Context(), mock.Anything).Return(validDeviceID.String(), nil)
				}

				dummyQueries.On("CreateRefreshToken", ctx.Request.Context(), mock.AnythingOfType("database.CreateRefreshTokenParams")).Return(nil)
			}

			svc := auth.NewAuthService(mockSqlTxQ, mockUserQ, mockTokenGen, mockExtractor, mockHasher)
			h := authhttp.NewHandler(svc)
			router := gin.New()
			router.POST("/login", h.Login)
			router.ServeHTTP(w, req)

			testhelpers.CheckHTTPResponse(t, w, tc.expErrSubstr, tc.expectedStatusCode, tc.expectedResponse, testhelpers.ProcessResponse(w, t))

			mockSqlTxQ.AssertExpectations(t)
			mockUserQ.AssertExpectations(t)
			mockTokenGen.AssertExpectations(t)
			mockExtractor.AssertExpectations(t)
			mockHasher.AssertExpectations(t)
		})
	}
}
