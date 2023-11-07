package config_test

import (
	"fmt"
	"testing"

	"github.com/jademcosta/graviola/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestAcceptSpecificValues(t *testing.T) {

	testCases := []struct {
		value       string
		shouldError bool
	}{
		{"debug", false},
		{"info", false},
		{"warn", false},
		{"error", false},
		{"Debug", false},
		{"INFO", false},
		{"WaRn", false},

		{"panic", true},
		{"warning", true},
		{"", true},
		{"something", true},
		{"infodebug", true},
		{"infoa", true},
	}

	for _, tc := range testCases {
		sut := config.LogConfig{Level: tc.value}
		err := sut.IsValid()

		if tc.shouldError {
			assert.Error(t, err, fmt.Sprintf("value %s should result in error when calling Validate", tc.value))
		} else {
			assert.NoError(t, err, fmt.Sprintf("value %s should NOT result in error when calling Validate", tc.value))
		}
	}
}

func TestDefaultValues(t *testing.T) {
	sut := config.LogConfig{}
	newSut := sut.FillDefaults()

	assert.Equal(t, config.DefaultLogLevel, newSut.Level,
		fmt.Sprintf("log level should be set to %s if the provided value is empty", config.DefaultLogLevel))
}
