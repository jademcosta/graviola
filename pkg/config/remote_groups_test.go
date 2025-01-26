package config_test

import (
	"testing"

	"github.com/jademcosta/graviola/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		sut := config.RemoteGroupsConfig{OnQueryFailStrategy: tc.value, Name: "some name",
			Servers: []config.RemoteConfig{{Name: "some name", Address: "http://non-existent.something"}}}
		err := sut.IsValid()

		if tc.shouldError {
			assert.Error(t, err, "value %s should result in error when calling Validate", tc.value)
		} else {
			assert.NoError(t, err, "value %s should NOT result in error when calling Validate", tc.value)
		}
	}
}

func TestGroupsValidate(t *testing.T) {
	sut := config.RemoteGroupsConfig{OnQueryFailStrategy: "fail_all"}
	require.Error(t, sut.IsValid(), "should error if name is empty")

	sut = config.RemoteGroupsConfig{Name: "group 1"}
	require.Error(t, sut.IsValid(), "should error if failure strategy is empty")

	sut = config.RemoteGroupsConfig{Name: "group 1", OnQueryFailStrategy: "fail_all"}
	require.Error(t, sut.IsValid(), "should error remotes is empty")

	sut = config.RemoteGroupsConfig{Name: "group 1", OnQueryFailStrategy: "fail_all",
		Servers: []config.RemoteConfig{
			{Name: "some name", Address: "http://non-existent.something"},
			{Name: "some name 2", Address: "http://non-existent.something"},
		}}
	require.NoError(t, sut.IsValid(), "should NOT error when everything is correct")

	sut = config.RemoteGroupsConfig{Name: "group 1", OnQueryFailStrategy: "fail_all",
		Servers: []config.RemoteConfig{
			{Name: "some name", Address: "http://non-existent.something"},
			{Name: "some name2", Address: "non-existent.something"}}}
	require.Error(t, sut.IsValid(), "should error when underlying remote returns error")

	sut = config.RemoteGroupsConfig{Name: "group 1", OnQueryFailStrategy: "fail_all",
		Servers: []config.RemoteConfig{
			{Name: "some name", Address: "http://non-existent.something"},
			{Name: "some name", Address: "http://non-existent.something"}}}
	require.Error(t, sut.IsValid(), "should error when remotes have the same name")
}

func TestOnQueryFailDefaultValues(t *testing.T) {
	sut := config.RemoteGroupsConfig{}
	newSut := sut.FillDefaults()

	assert.Equalf(t, config.StrategyFailAll, newSut.OnQueryFailStrategy,
		"query failure strategy should be set to %s if the provided value is empty",
		config.StrategyFailAll,
	)
}
