package config_test

import (
	"testing"

	"github.com/jademcosta/graviola/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestStoragesValidate(t *testing.T) {

	sut := config.StoragesConfig{}
	assert.Error(t, sut.IsValid(), "should error when groups <= 0")

	sut.Groups = make([]config.GroupsConfig, 0)
	sut.Groups = append(sut.Groups, config.GroupsConfig{Name: "first!!!", OnQueryFailStrategy: "fail_all",
		Servers: []config.RemoteConfig{{Name: "remote 1", Address: "http://non-existent.something"}}})
	assert.Error(t, sut.IsValid(), "should error when merge strategy is invalid")

	sut.MergeConf = config.MergeStrategyConfig{}
	assert.Error(t, sut.IsValid(), "should error when merge strategy errors")

	sut.MergeConf = config.MergeStrategyConfig{Strategy: "always_merge"}
	assert.NoError(t, sut.IsValid(), "should NOT error when everything is valid")

	sut.Groups = append(sut.Groups, config.GroupsConfig{})
	assert.Error(t, sut.IsValid(), "should error when an underlying group returns a validation error")

	sut.Groups = []config.GroupsConfig{
		{Name: "first", OnQueryFailStrategy: "fail_all"},
		{Name: "second", OnQueryFailStrategy: "fail_all"},
		{Name: "first", OnQueryFailStrategy: "fail_all"},
	}
	assert.Error(t, sut.IsValid(), "should error when group names are duplicated")
}

func TestStoragesFillDefaults(t *testing.T) {

	sut := config.StoragesConfig{}

	sut = sut.FillDefaults()
	assert.Equal(t, config.DefaultMergeStrategyType, sut.MergeConf.Strategy, "should have called FillDefaults on children configs")

	sut = config.StoragesConfig{Groups: []config.GroupsConfig{{}, {}}}
	sut = sut.FillDefaults()

	for _, group := range sut.Groups {
		assert.Equal(t, config.DefaultOnFailStrategy, group.OnQueryFailStrategy, "should have called FillDefaults on all children groups")
	}
}
