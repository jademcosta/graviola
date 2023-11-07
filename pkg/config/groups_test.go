package config_test

import (
	"fmt"
	"testing"

	"github.com/jademcosta/graviola/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestOnQueryFailAcceptSpecificValues(t *testing.T) {

	testCases := []struct {
		value       string
		shouldError bool
	}{
		{"fail_all", false},
		{"partial_response", false},
		{"Partial_Response", false},
		{"FAIL_ALL", false},

		{"FAILALL", true},
		{"partialresponse", true},
		{"", true},
		{"anything", true},
	}

	for _, tc := range testCases {
		sut := config.GroupsConfig{OnQueryFailStrategy: tc.value, Name: "some name"}
		err := sut.IsValid()

		if tc.shouldError {
			assert.Error(t, err, fmt.Sprintf("value %s should result in error when calling Validate", tc.value))
		} else {
			assert.NoError(t, err, fmt.Sprintf("value %s should NOT result in error when calling Validate", tc.value))
		}
	}
}

func TestOnQueryFailDefaultValues(t *testing.T) {
	sut := config.GroupsConfig{}
	newSut := sut.FillDefaults()

	assert.Equal(t, config.DefaultStrategy, newSut.OnQueryFailStrategy,
		fmt.Sprintf("query failure strategy should be set to %s if the provided value is empty", config.DefaultStrategy))
}
