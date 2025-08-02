package config_test

import (
	"testing"

	"github.com/jademcosta/graviola/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestAcceptSpecificMergeStrategiesNames(t *testing.T) {

	testCases := []struct {
		value       string
		shouldError bool
	}{
		{"debug", true},
		{"", true},
		{"something", true},
		{"keepbiggest", true},
		{"alwaysmerge", true},

		{"always_merge", false},
		{"keep_biggest", false},
	}

	for _, tc := range testCases {
		sut := config.MergeStrategyConfig{Strategy: tc.value}
		err := sut.IsValid()

		if tc.shouldError {
			assert.Error(t, err, "type value %s should result in error when calling Validate", tc.value)
		} else {
			assert.NoError(t, err, "type value %s should NOT result in error when calling Validate", tc.value)
		}
	}
}

func TestMergeStrategyDefaultValues(t *testing.T) {
	sut := config.MergeStrategyConfig{}
	newSut := sut.FillDefaults()

	assert.Equal(t, config.DefaultMergeStrategyType, newSut.Strategy,
		"merge strategy type should be set to %s if the provided value is empty", config.DefaultMergeStrategyType)
}
