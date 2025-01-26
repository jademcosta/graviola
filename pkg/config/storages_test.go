package config_test

import (
	"testing"

	"github.com/jademcosta/graviola/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStoragesValidate(t *testing.T) {

	sut := config.StoragesConfig{}
	require.Error(t, sut.IsValid(), "should error when groups <= 0")

	sut.Groups = make([]config.RemoteGroupsConfig, 0)
	sut.Groups = append(sut.Groups, config.RemoteGroupsConfig{Name: "first!!!", OnQueryFailStrategy: "fail_all",
		Servers: []config.RemoteConfig{{Name: "remote 1", Address: "http://non-existent.something"}}})
	require.Error(t, sut.IsValid(), "should error when merge strategy is invalid")

	sut.MergeConf = config.MergeStrategyConfig{}
	require.Error(t, sut.IsValid(), "should error when merge strategy errors")

	sut.MergeConf = config.MergeStrategyConfig{Strategy: "always_merge"}
	require.NoError(t, sut.IsValid(), "should NOT error when everything is valid")

	sut.Groups = append(sut.Groups, config.RemoteGroupsConfig{})
	require.Error(t, sut.IsValid(), "should error when an underlying group returns a validation error")

	sut.Groups = []config.RemoteGroupsConfig{
		{Name: "first", OnQueryFailStrategy: "fail_all"},
		{Name: "second", OnQueryFailStrategy: "fail_all"},
		{Name: "first", OnQueryFailStrategy: "fail_all"},
	}
	require.Error(t, sut.IsValid(), "should error when group names are duplicated")
}

func TestStoragesFillDefaults(t *testing.T) {

	sut := config.StoragesConfig{}

	sut = sut.FillDefaults()
	assert.Equal(t, config.DefaultMergeStrategyType, sut.MergeConf.Strategy, "should have called FillDefaults on children configs")

	sut = config.StoragesConfig{Groups: []config.RemoteGroupsConfig{{}, {}}}
	sut = sut.FillDefaults()

	for _, group := range sut.Groups {
		assert.Equal(t, config.DefaultOnFailStrategy, group.OnQueryFailStrategy, "should have called FillDefaults on all children groups")
	}
}
