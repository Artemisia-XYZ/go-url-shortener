package helpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStrRandom(t *testing.T) {
	tests := []struct {
		name   string
		length uint
	}{
		{
			name:   "random 64 characters",
			length: 64,
		},
		{
			name:   "random 0 characters",
			length: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, len(StrRandom(tt.length)), int(tt.length))
		})
	}
}
