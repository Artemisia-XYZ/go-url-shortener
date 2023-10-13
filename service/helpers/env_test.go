package helpers

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetenv(t *testing.T) {
	tests := []struct {
		name     string
		setup    func()
		expected string
	}{
		{
			name: "defined",
			setup: func() {
				os.Setenv("TEST_ENV", "foo")
			},
			expected: "foo",
		},
		{
			name:     "undefined",
			expected: "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
				defer os.Clearenv()
			}

			assert.Equal(t, tt.expected, Getenv("TEST_ENV", "default"))
		})
	}
}
