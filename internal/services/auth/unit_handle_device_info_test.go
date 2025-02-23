package auth

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/seanhuebl/unity-wealth/internal/database"
	dbmocks "github.com/seanhuebl/unity-wealth/internal/mocks/database"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestHandleDeviceInfo(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	validDeviceIDStr := uuid.NewString()
	newDeviceIDStr := uuid.NewString()

	inputDeviceInfo := DeviceInfo{
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
					UserID:         userID.String(),
					DeviceType:     inputDeviceInfo.DeviceType,
					Browser:        inputDeviceInfo.Browser,
					BrowserVersion: inputDeviceInfo.BrowserVersion,
					Os:             inputDeviceInfo.Os,
					OsVersion:      inputDeviceInfo.OsVersion,
				}).Return(validDeviceIDStr, nil)

				tokenQ.On("RevokeToken", ctx, mock.MatchedBy(func(params database.RevokeTokenParams) bool {
					return params.UserID == userID.String() && params.DeviceInfoID == validDeviceIDStr
				})).Return(nil)
			},
			expectedErrorSubstring: "",
			expectedDeviceID:       uuid.MustParse(validDeviceIDStr),
		},
		{
			name: "device not found and new device created",
			setupMocks: func(deviceQ *dbmocks.DeviceQuerier, tokenQ *dbmocks.TokenQuerier) {
				deviceQ.On("GetDeviceInfoByUser", ctx, database.GetDeviceInfoByUserParams{
					UserID:         userID.String(),
					DeviceType:     inputDeviceInfo.DeviceType,
					Browser:        inputDeviceInfo.Browser,
					BrowserVersion: inputDeviceInfo.BrowserVersion,
					Os:             inputDeviceInfo.Os,
					OsVersion:      inputDeviceInfo.OsVersion,
				}).Return("", sql.ErrNoRows)

				deviceQ.On("CreateDeviceInfo", ctx, mock.MatchedBy(func(params database.CreateDeviceInfoParams) bool {
					return params.UserID == userID.String() &&
						params.DeviceType == inputDeviceInfo.DeviceType &&
						params.Browser == inputDeviceInfo.Browser
				})).Return(newDeviceIDStr, nil)
			},
			expectedErrorSubstring: "",
			expectedDeviceID:       uuid.MustParse(newDeviceIDStr),
		},
		{
			name: "device not found and creation fails",
			setupMocks: func(deviceQ *dbmocks.DeviceQuerier, tokenQ *dbmocks.TokenQuerier) {
				deviceQ.On("GetDeviceInfoByUser", ctx, database.GetDeviceInfoByUserParams{
					UserID:         userID.String(),
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
					UserID:         userID.String(),
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
					UserID:         userID.String(),
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
					UserID:         userID.String(),
					DeviceType:     inputDeviceInfo.DeviceType,
					Browser:        inputDeviceInfo.Browser,
					BrowserVersion: inputDeviceInfo.BrowserVersion,
					Os:             inputDeviceInfo.Os,
					OsVersion:      inputDeviceInfo.OsVersion,
				}).Return(validDeviceIDStr, nil)

				tokenQ.On("RevokeToken", ctx, mock.MatchedBy(func(params database.RevokeTokenParams) bool {
					return params.DeviceInfoID == validDeviceIDStr
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

			if tc.setupMocks != nil {
				tc.setupMocks(mockDeviceQ, mockTokenQ)
			}

			deviceID, err := NewAuthService(nil, nil, nil, nil, nil).handleDeviceInfo(ctx, mockDeviceQ, mockTokenQ, userID, inputDeviceInfo)

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
