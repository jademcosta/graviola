package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimeParsing(t *testing.T) {

	testCases := []struct {
		input       string
		expected    time.Duration
		shouldError bool
	}{
		{"1Ms", 0, true},
		{"1A", 0, true},
		{"1MO", 0, true},
		{"1mo", 0, true},
		{"1M", 0, true},
		{"1MS", 0, true},
		{"1S", 0, true},

		{"1s", 1 * time.Second, false},
		{"1ms", 1 * time.Millisecond, false},
		{"1m", 1 * time.Minute, false},

		{"84m", 84 * time.Minute, false},
		{"8424ms", 8424 * time.Millisecond, false},
	}

	for _, tc := range testCases {
		result, err := parseDuration(tc.input)

		assert.Equal(t, tc.expected, result, "should be equal")

		if tc.shouldError {
			assert.Errorf(t, err, "input %s should return error", tc.input)
			continue
		}
		assert.NoErrorf(t, err, "input %s should NOT return error", tc.input)
	}
}
