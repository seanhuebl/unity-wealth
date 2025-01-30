package handlers_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/seanhuebl/unity-wealth/handlers"
	"github.com/seanhuebl/unity-wealth/internal/database"
	"github.com/seanhuebl/unity-wealth/mocks"
	"github.com/stretchr/testify/mock"
)

func TestHandleDeviceInfo(t *testing.T) {
	userID := uuid.New().String()
	// The device already exists in the DB
	existingDeviceID := uuid.New().String()
	// Define several test scenarios using table-driven tests
	newlyCreatedID := uuid.New().String()
	tests := []struct {
		name             string
		userID           string
		info             handlers.DeviceInfo
		mockSetup        func(*mocks.Quierier) // how we configure the mock DB
		expectedDeviceID uuid.UUID
		expectedError    error
	}{
		{
			name:   "Device found successfully",
			userID: userID,
			info: handlers.DeviceInfo{
				DeviceType:     "mobile",
				Browser:        "Chrome",
				BrowserVersion: "99.0",
				Os:             "Android",
				OsVersion:      "12",
			},
			mockSetup: func(mockQueries *mocks.Quierier) {

				// Mock: GetDeviceInfoByUser returns existingDeviceID (no error)
				mockQueries.On("GetDeviceInfoByUser", context.Background(), database.GetDeviceInfoByUserParams{
					UserID:         userID,
					DeviceType:     "mobile",
					Browser:        "Chrome",
					BrowserVersion: "99.0",
					Os:             "Android",
					OsVersion:      "12",
				}).Return(existingDeviceID, nil)

				// 2) RevokeToken is called for the found device
				mockQueries.On("RevokeToken", mock.Anything, database.RevokeTokenParams{
					UserID:       userID,
					DeviceInfoID: existingDeviceID,
				}).Return(nil) // success
			},
			expectedDeviceID: uuid.MustParse(existingDeviceID), // This won't match the actual call unless you unify them. See notes below.
			expectedError:    nil,
		},
		{
			name:   "No device found -> create device successfully",
			userID: userID,
			info: handlers.DeviceInfo{
				DeviceType:     "desktop",
				Browser:        "Firefox",
				BrowserVersion: "88.0",
				Os:             "Windows",
				OsVersion:      "10",
			},
			mockSetup: func(mockQueries *mocks.Quierier) {
				// 1) Simulate no existing device
				mockQueries.On("GetDeviceInfoByUser", context.Background(), database.GetDeviceInfoByUserParams{
					UserID:         userID,
					DeviceType:     "desktop",
					Browser:        "Firefox",
					BrowserVersion: "88.0",
					Os:             "Windows",
					OsVersion:      "10",
				}).Return("", sql.ErrNoRows)

				// 2) Then simulate creating a new device
				mockQueries.On("CreateDeviceInfo", mock.Anything, mock.MatchedBy(func(params database.CreateDeviceInfoParams) bool {
					return params.UserID == userID && params.DeviceType == "desktop" &&
						params.Browser == "Firefox" && params.BrowserVersion == "88.0" &&
						params.Os == "Windows" && params.OsVersion == "10"
				})).Return(newlyCreatedID, nil)

			},
			expectedDeviceID: uuid.MustParse(newlyCreatedID),
			expectedError:    nil,
		},
		{
			name:   "Error during device lookup",
			userID: userID,
			info: handlers.DeviceInfo{
				DeviceType:     "tablet",
				Browser:        "Safari",
				BrowserVersion: "14.0",
				Os:             "iOS",
				OsVersion:      "14",
			},
			mockSetup: func(mockQueries *mocks.Quierier) {
				mockQueries.On("GetDeviceInfoByUser", context.Background(), database.GetDeviceInfoByUserParams{
					UserID:         userID,
					DeviceType:     "tablet",
					Browser:        "Safari",
					BrowserVersion: "14.0",
					Os:             "iOS",
					OsVersion:      "14",
				}).Return(uuid.Nil.String(), fmt.Errorf("failed to fetch device info"))

			},
			expectedDeviceID: uuid.Nil,
			expectedError:    fmt.Errorf("failed to fetch device info"),
		},
		{
			name:   "No device found -> error creating new device",
			userID: userID,
			info: handlers.DeviceInfo{
				DeviceType:     "tv",
				Browser:        "TizenBrowser",
				BrowserVersion: "3.0",
				Os:             "Tizen",
				OsVersion:      "5.5",
			},
			mockSetup: func(mockQueries *mocks.Quierier) {
				mockQueries.On("GetDeviceInfoByUser", context.Background(), database.GetDeviceInfoByUserParams{
					UserID:         userID,
					DeviceType:     "tv",
					Browser:        "TizenBrowser",
					BrowserVersion: "3.0",
					Os:             "Tizen",
					OsVersion:      "5.5",
				}).Return("", sql.ErrNoRows)

				mockQueries.On("CreateDeviceInfo", mock.Anything, mock.MatchedBy(func(params database.CreateDeviceInfoParams) bool {
					return params.UserID == userID && params.DeviceType == "tv" &&
						params.Browser == "TizenBrowser" && params.BrowserVersion == "3.0" &&
						params.Os == "Tizen" && params.OsVersion == "5.5"
				})).Return(uuid.Nil.String(), errors.New("insert failed"))
			},
			expectedDeviceID: uuid.Nil,
			expectedError:    fmt.Errorf("failed to create new device: insert failed"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create your mocks
			mockQueries := mocks.NewQuierier(t)

			// Set up the mock DB behavior
			tc.mockSetup(mockQueries)

			// Call the function under test
			gotDeviceID, gotErr := handlers.HandleDeviceInfo(
				context.Background(),
				mockQueries, // Our Quierier mock
				uuid.MustParse(tc.userID),
				tc.info,
			)
			// Compare device IDs
			if diff := cmp.Diff(tc.expectedDeviceID, gotDeviceID); diff != "" {
				t.Errorf("DeviceID mismatch (-want +got):\n%s", diff)
			}

			// Compare errors
			if diff := cmp.Diff(
				tc.expectedError,
				gotErr,
				cmp.Comparer(func(x, y error) bool {
					if x == nil || y == nil {
						return x == y
					}
					return x.Error() == y.Error()
				}),
			); diff != "" {
				t.Errorf("Error mismatch (-want +got):\n%s", diff)
			}

			mockQueries.AssertExpectations(t)
		})
	}
}
