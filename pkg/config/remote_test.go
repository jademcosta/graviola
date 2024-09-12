package config_test

import (
	"testing"

	"github.com/jademcosta/graviola/pkg/config"
	"github.com/stretchr/testify/require"
)

func TestRemoteValidate(t *testing.T) {
	sut := config.RemoteConfig{Address: "http://something.com"}
	err := sut.IsValid()
	require.Error(t, err, "should return error when name is empty")

	sut = config.RemoteConfig{Name: "a name"}
	err = sut.IsValid()
	require.Error(t, err, "should return error when address is empty")

	sut = config.RemoteConfig{Name: "a name", Address: "aaaa"}
	err = sut.IsValid()
	require.Error(t, err, "should return error when address doesn't start with http:// or https://")

	sut = config.RemoteConfig{Name: "a name", Address: "http://something"}
	err = sut.IsValid()
	require.NoError(t, err, "should NOT return error when address starts with http://")

	sut = config.RemoteConfig{Name: "a name", Address: "https://something"}
	err = sut.IsValid()
	require.NoError(t, err, "should NOT return error when address starts with http:// or https://")
}
