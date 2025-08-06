package auth_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/seanhuebl/unity-wealth/internal/models"
)

func TestValidatePassword(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		password   string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:       "Valid password",
			password:   "StrongPass1!",
			wantErr:    false,
			wantErrMsg: "",
		},
		{
			name:       "Too short",
			password:   "Short1!",
			wantErr:    true,
			wantErrMsg: "password must be at least 8 characters long",
		},
		{
			name:       "No uppercase",
			password:   "weakpass1!",
			wantErr:    true,
			wantErrMsg: "password must contain at least one uppercase letter",
		},
		{
			name:       "No lowercase",
			password:   "WEAKPASS1!",
			wantErr:    true,
			wantErrMsg: "password must contain at least one lowercase letter",
		},
		{
			name:       "No digit",
			password:   "WeakPass!",
			wantErr:    true,
			wantErrMsg: "password must contain at least one digit",
		},
		{
			name:       "No special character",
			password:   "WeakPass1",
			wantErr:    true,
			wantErrMsg: "password must contain at least one special character",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := models.ValidatePassword(tc.password)

			if (err != nil) != tc.wantErr {
				t.Errorf("ValidatePassword() error = %v, wantErr %v", err, tc.wantErr)
			}

			if err != nil {
				if diff := cmp.Diff(tc.wantErrMsg, err.Error()); diff != "" {
					t.Errorf("Error message mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}
