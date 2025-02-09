// SPDX-FileCopyrightText: 2025 The Cipher Host Team <team@cipher.host>
//
// SPDX-License-Identifier: MIT

package cmdkit_test

import (
	"errors"
	"testing"

	"go.cipher.host/cmdkit"
)

var errGeneric cmdkit.Error = "generic error"

func TestError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		give       error
		wantString string
		wantError  error
	}{
		{
			name:       "generic test",
			give:       errGeneric,
			wantString: "generic error",
			wantError:  errGeneric,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.give.Error(); got != tt.wantString {
				t.Errorf("Error(): got %v, want %v", got, tt.wantString)
			}

			if !errors.Is(tt.give, tt.wantError) {
				t.Errorf("Error(): got %v, want %v", tt.give, tt.wantError)
			}
		})
	}
}

func TestExitError_NewExitError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		giveError error
		giveCode  int
		wantErr   string
		wantCode  int
	}{
		{
			name:      "Valid error and code",
			giveError: errGeneric,
			giveCode:  1,
			wantErr:   "generic error",
			wantCode:  1,
		},
		{
			name:      "Nil error and valid code",
			giveError: nil,
			giveCode:  2,
			wantErr:   "undefined error",
			wantCode:  2,
		},
		{
			name:      "Zero values",
			giveError: nil,
			giveCode:  0,
			wantErr:   "undefined error",
			wantCode:  0,
		},
		{
			name:      "Negative code",
			giveError: errGeneric,
			giveCode:  -1,
			wantErr:   "generic error",
			wantCode:  -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := cmdkit.NewExitError(tt.giveError, tt.giveCode)

			if got.Error() != tt.wantErr {
				t.Errorf("New(%v, %d).Error() = %q, want %q", tt.giveError, tt.giveCode, got.Error(), tt.wantErr)
			}

			if got.Code() != tt.wantCode {
				t.Errorf("New(%v, %d).Code() = %d, want %d", tt.giveError, tt.giveCode, got.Code(), tt.wantCode)
			}
		})
	}
}

func TestExitError_Unwrap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		giveError error
		giveCode  int
		wantErr   error
	}{
		{
			name:      "With underlying error",
			giveError: errGeneric,
			giveCode:  1,
			wantErr:   errGeneric,
		},
		{
			name:      "With nil underlying error",
			giveError: nil,
			giveCode:  2,
			wantErr:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var (
				err       = cmdkit.NewExitError(tt.giveError, tt.giveCode)
				unwrapped = err.Unwrap()
			)

			if !errors.Is(unwrapped, tt.wantErr) {
				t.Fatalf("Unwrap() = %v, want %v", unwrapped, tt.wantErr)
			}
		})
	}
}
