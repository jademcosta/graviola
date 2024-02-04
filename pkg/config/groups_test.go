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
		sut := config.GroupsConfig{OnQueryFailStrategy: tc.value, Name: "some name",
			Servers: []config.RemoteConfig{{Name: "some name", Address: "http://non-existent.something"}}}
		err := sut.IsValid()

		if tc.shouldError {
			assert.Error(t, err, fmt.Sprintf("value %s should result in error when calling Validate", tc.value))
		} else {
			assert.NoError(t, err, fmt.Sprintf("value %s should NOT result in error when calling Validate", tc.value))
		}
	}
}

func TestGroupsValidate(t *testing.T) {
	sut := config.GroupsConfig{OnQueryFailStrategy: "fail_all"}
	assert.Error(t, sut.IsValid(), "should error if name is empty")

	sut = config.GroupsConfig{Name: "group 1"}
	assert.Error(t, sut.IsValid(), "should error if failure strategy is empty")

	sut = config.GroupsConfig{Name: "group 1", OnQueryFailStrategy: "fail_all"}
	assert.Error(t, sut.IsValid(), "should error remotes is empty")

	sut = config.GroupsConfig{Name: "group 1", OnQueryFailStrategy: "fail_all",
		Servers: []config.RemoteConfig{
			{Name: "some name", Address: "http://non-existent.something"},
			{Name: "some name 2", Address: "http://non-existent.something"},
		}}
	assert.NoError(t, sut.IsValid(), "should NOT error when everything is correct")

	sut = config.GroupsConfig{Name: "group 1", OnQueryFailStrategy: "fail_all",
		Servers: []config.RemoteConfig{
			{Name: "some name", Address: "http://non-existent.something"},
			{Name: "some name2", Address: "non-existent.something"}}}
	assert.Error(t, sut.IsValid(), "should error when underlying remote returns error")

	sut = config.GroupsConfig{Name: "group 1", OnQueryFailStrategy: "fail_all",
		Servers: []config.RemoteConfig{
			{Name: "some name", Address: "http://non-existent.something"},
			{Name: "some name", Address: "http://non-existent.something"}}}
	assert.Error(t, sut.IsValid(), "should error when remotes have the same name")
}

func TestOnQueryFailDefaultValues(t *testing.T) {
	sut := config.GroupsConfig{}
	newSut := sut.FillDefaults()

	assert.Equal(t, config.DefaultOnFailStrategy, newSut.OnQueryFailStrategy,
		fmt.Sprintf("query failure strategy should be set to %s if the provided value is empty", config.DefaultOnFailStrategy))
}
