package validation

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestValidateStringLength(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		minLength int
		maxLength int
		wantErr   bool
	}{
		{
			name:      "Valid Length",
			value:     "test",
			minLength: 2,
			maxLength: 6,
			wantErr:   false,
		},
		{
			name:      "Too Short",
			value:     "a",
			minLength: 2,
			maxLength: 6,
			wantErr:   true,
		},
		{
			name:      "Too Long",
			value:     "testing",
			minLength: 2,
			maxLength: 6,
			wantErr:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateStringLength(tc.value, tc.minLength, tc.maxLength)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:    "Valid Email",
			value:   "test@example.com",
			wantErr: false,
		},
		{
			name:    "Invalid Email 1",
			value:   "testexample.com",
			wantErr: true,
		},
		{
			name:    "Invalid Email 2",
			value:   "test@examplecom",
			wantErr: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateEmail(tc.value)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
