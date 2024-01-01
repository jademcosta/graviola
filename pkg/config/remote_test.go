package config_test

import (
	"testing"

	"github.com/jademcosta/graviola/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestRemoteValidate(t *testing.T) {
	sut := config.RemoteConfig{}
	err := sut.IsValid()
	assert.Error(t, err, "should return error when address is empty")

	sut = config.RemoteConfig{Address: "aaaa"}
	err = sut.IsValid()
	assert.Error(t, err, "should return error when address doesn't start with http:// or https://")

	sut = config.RemoteConfig{Address: "http://something"}
	err = sut.IsValid()
	assert.NoError(t, err, "should NOT return error when address starts with http://")

	sut = config.RemoteConfig{Address: "https://something"}
	err = sut.IsValid()
	assert.NoError(t, err, "should return error when address doesn't start with http:// or https://")
}
