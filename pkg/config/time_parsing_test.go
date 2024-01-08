package config_test

import (
	"testing"
	"time"

	"github.com/jademcosta/graviola/pkg/config"
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
		{"1D", 0, true},
		{"0", 0, true},

		{"1s", 1 * time.Second, false},
		{"1ms", 1 * time.Millisecond, false},
		{"1m", 1 * time.Minute, false},
		{"1h", 1 * time.Hour, false},
		{"1d", 1 * time.Hour * 24, false},

		{"84m", 84 * time.Minute, false},
		{"8424ms", 8424 * time.Millisecond, false},
		{"2d", 24 * 2 * time.Hour, false},

		{"365d", 24 * 365 * time.Hour, false},
	}

	for _, tc := range testCases {
		result, err := config.ParseDuration(tc.input)

		assert.Equal(t, tc.expected, result, "should be equal")

		if tc.shouldError {
			assert.Errorf(t, err, "input %s should return error", tc.input)
			continue
		}
		assert.NoErrorf(t, err, "input %s should NOT return error", tc.input)
	}
}

func TestDateParsing(t *testing.T) {

	now := time.Now()

	testCases := []struct {
		input       string
		expected    time.Time
		shouldError bool
	}{
		{"-1", time.Unix(0, 0), true},
		{"abc", time.Unix(0, 0), true},
		{"Monday", time.Unix(0, 0), true},
		{"Friday, September 1st", time.Unix(0, 0), true},
		{"now-", time.Unix(0, 0), true},
		{"-now", time.Unix(0, 0), true},
		{"now+6s", time.Unix(0, 0), true},
		{"now-6s-6d", time.Unix(0, 0), true},
		{"2022-04-12T23:20:50.52", time.Unix(0, 0), true},
		{"2022-04-12T23:20:11", time.Unix(0, 0), true},
		{"2022-04-12", time.Unix(0, 0), true},
		{"23:20:50.52Z", time.Unix(0, 0), true},
		{"2022-04-12 23:20:50.52Z", time.Unix(0, 0), true},
		{"2022-04-12-23:20:50.52", time.Unix(0, 0), true},

		{"1234", time.Unix(1234, 0), false},
		{"0", time.Unix(0, 0), false},
		{"1704157075", time.Unix(1704157075, 0), false},

		{"now-6s", now.Add(-(6 * time.Second)), false},
		{"now-1m", now.Add(-(1 * time.Minute)), false},
		{"now-16h", now.Add(-(16 * time.Hour)), false},
		{"now", now, false},

		{"1985-04-12T23:20:50.52Z", parseOrPanic("1985-04-12T23:20:50.52Z"), false},
		{"1996-12-19T16:39:57-08:00", parseOrPanic("1996-12-19T16:39:57-08:00"), false},
		{"2022-12-19T16:39:57+03:00", parseOrPanic("2022-12-19T16:39:57+03:00"), false},
		{" 2022-12-19T16:39:57+03:00", parseOrPanic("2022-12-19T16:39:57+03:00"), false},
		{"2048-04-12T23:20:50.52Z", parseOrPanic("2048-04-12T23:20:50.52Z"), false},
	}

	for _, tc := range testCases {
		result, err := config.ParseDate(tc.input, now)

		if tc.shouldError {
			assert.Errorf(t, err, "input %s should return error", tc.input)
			continue
		}

		assert.Equal(t, tc.expected, result, "should be equal")
		assert.NoErrorf(t, err, "input %s should NOT return error", tc.input)
	}
}

func parseOrPanic(date string) time.Time {
	t, err := time.Parse(time.RFC3339, date)
	if err != nil {
		panic(err)
	}
	return t
}
