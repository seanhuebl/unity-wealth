package auth_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/seanhuebl/unity-wealth/internal/database"
	dbmocks "github.com/seanhuebl/unity-wealth/internal/mocks/database"
	"github.com/seanhuebl/unity-wealth/internal/models"
	"github.com/seanhuebl/unity-wealth/internal/services/auth"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestHandleDeviceInfo(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	validDeviceID := uuid.New()
	newDeviceID := uuid.New()

	inputDeviceInfo := models.DeviceInfo{
		DeviceType:     "Mobile",
		Browser:        "Chrome",
		BrowserVersion: "100.0",
		Os:             "Android",
		OsVersion:      "11",
	}

	tests := []struct {
		name                   string
		setupMocks             func(deviceQ *dbmocks.DeviceQuerier, tokenQ *dbmocks.TokenQuerier)
		expectedErrorSubstring string
		expectedDeviceID       uuid.UUID
	}{
		{
			name: "found valid device and token revoked",
			setupMocks: func(deviceQ *dbmocks.DeviceQuerier, tokenQ *dbmocks.TokenQuerier) {
				deviceQ.On("GetDeviceInfoByUser", ctx, database.GetDeviceInfoByUserParams{
					UserID:         userID,
					DeviceType:     inputDeviceInfo.DeviceType,
					Browser:        inputDeviceInfo.Browser,
					BrowserVersion: inputDeviceInfo.BrowserVersion,
					Os:             inputDeviceInfo.Os,
					OsVersion:      inputDeviceInfo.OsVersion,
				}).Return(validDeviceID, nil)

				tokenQ.On("RevokeToken", ctx, mock.MatchedBy(func(params database.RevokeTokenParams) bool {
					return params.UserID == userID && params.DeviceInfoID == validDeviceID
				})).Return(nil)
			},
			expectedErrorSubstring: "",
			expectedDeviceID:       validDeviceID,
		},
		{
			name: "device not found and new device created",
			setupMocks: func(deviceQ *dbmocks.DeviceQuerier, tokenQ *dbmocks.TokenQuerier) {
				deviceQ.On("GetDeviceInfoByUser", ctx, database.GetDeviceInfoByUserParams{
					UserID:         userID,
					DeviceType:     inputDeviceInfo.DeviceType,
					Browser:        inputDeviceInfo.Browser,
					BrowserVersion: inputDeviceInfo.BrowserVersion,
					Os:             inputDeviceInfo.Os,
					OsVersion:      inputDeviceInfo.OsVersion,
				}).Return("", sql.ErrNoRows)

				deviceQ.On("CreateDeviceInfo", ctx, mock.MatchedBy(func(params database.CreateDeviceInfoParams) bool {
					return params.UserID == userID &&
						params.DeviceType == inputDeviceInfo.DeviceType &&
						params.Browser == inputDeviceInfo.Browser
				})).Return(newDeviceID, nil)
			},
			expectedErrorSubstring: "",
			expectedDeviceID:       newDeviceID,
		},
		{
			name: "device not found and creation fails",
			setupMocks: func(deviceQ *dbmocks.DeviceQuerier, tokenQ *dbmocks.TokenQuerier) {
				deviceQ.On("GetDeviceInfoByUser", ctx, database.GetDeviceInfoByUserParams{
					UserID:         userID,
					DeviceType:     inputDeviceInfo.DeviceType,
					Browser:        inputDeviceInfo.Browser,
					BrowserVersion: inputDeviceInfo.BrowserVersion,
					Os:             inputDeviceInfo.Os,
					OsVersion:      inputDeviceInfo.OsVersion,
				}).Return("", sql.ErrNoRows)

				deviceQ.On("CreateDeviceInfo", ctx, mock.Anything).Return("", errors.New("create error"))
			},
			expectedErrorSubstring: "failed to create new device",
			expectedDeviceID:       uuid.Nil,
		},
		{
			name: "unexpected error fetching device info",
			setupMocks: func(deviceQ *dbmocks.DeviceQuerier, tokenQ *dbmocks.TokenQuerier) {
				deviceQ.On("GetDeviceInfoByUser", ctx, database.GetDeviceInfoByUserParams{
					UserID:         userID,
					DeviceType:     inputDeviceInfo.DeviceType,
					Browser:        inputDeviceInfo.Browser,
					BrowserVersion: inputDeviceInfo.BrowserVersion,
					Os:             inputDeviceInfo.Os,
					OsVersion:      inputDeviceInfo.OsVersion,
				}).Return("", errors.New("db error"))
			},
			expectedErrorSubstring: "failed to fetch device info",
			expectedDeviceID:       uuid.Nil,
		},
		{
			name: "found device with invalid UUID",
			setupMocks: func(deviceQ *dbmocks.DeviceQuerier, tokenQ *dbmocks.TokenQuerier) {
				deviceQ.On("GetDeviceInfoByUser", ctx, database.GetDeviceInfoByUserParams{
					UserID:         userID,
					DeviceType:     inputDeviceInfo.DeviceType,
					Browser:        inputDeviceInfo.Browser,
					BrowserVersion: inputDeviceInfo.BrowserVersion,
					Os:             inputDeviceInfo.Os,
					OsVersion:      inputDeviceInfo.OsVersion,
				}).Return("invalid-UUID", nil)
			},
			expectedErrorSubstring: "failed to parse device ID",
			expectedDeviceID:       uuid.Nil,
		},
		{
			name: "failed to revoke token",
			setupMocks: func(deviceQ *dbmocks.DeviceQuerier, tokenQ *dbmocks.TokenQuerier) {
				deviceQ.On("GetDeviceInfoByUser", ctx, database.GetDeviceInfoByUserParams{
					UserID:         userID,
					DeviceType:     inputDeviceInfo.DeviceType,
					Browser:        inputDeviceInfo.Browser,
					BrowserVersion: inputDeviceInfo.BrowserVersion,
					Os:             inputDeviceInfo.Os,
					OsVersion:      inputDeviceInfo.OsVersion,
				}).Return(validDeviceID, nil)

				tokenQ.On("RevokeToken", ctx, mock.MatchedBy(func(params database.RevokeTokenParams) bool {
					return params.DeviceInfoID == validDeviceID
				})).Return(errors.New("revoke error"))
			},
			expectedErrorSubstring: "failed to revoke token",
			expectedDeviceID:       uuid.Nil,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			mockDeviceQ := dbmocks.NewDeviceQuerier(t)
			mockTokenQ := dbmocks.NewTokenQuerier(t)
			nopLogger := zap.NewNop()
			if tc.setupMocks != nil {
				tc.setupMocks(mockDeviceQ, mockTokenQ)
			}

			deviceID, err := auth.NewAuthService(nil, nil, nil, nil, nil, nopLogger).HandleDeviceInfo(ctx, mockDeviceQ, mockTokenQ, userID, inputDeviceInfo)

			if tc.expectedErrorSubstring != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectedErrorSubstring)
				require.Equal(t, uuid.Nil, deviceID)
			} else {
				require.NoError(t, err)
				if diff := cmp.Diff(tc.expectedDeviceID, deviceID); diff != "" {
					t.Errorf("handleDeviceInfo() mismatch (-want +got):\n%s", diff)
				}
			}
			mockDeviceQ.AssertExpectations(t)
			mockTokenQ.AssertExpectations(t)
		})
	}
}
