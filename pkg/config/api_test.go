package config_test

import (
	"testing"

	"github.com/jademcosta/graviola/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApiValidate(t *testing.T) {
	sut := config.APIConfig{Port: 0}
	err := sut.IsValid()
	require.Error(t, err, "should return error when port is zero")

	sut = config.APIConfig{Port: 100}
	err = sut.IsValid()
	require.Error(t, err, "should return error when timeout is empty")

	sut = config.APIConfig{Port: 100, Timeout: "111"}
	err = sut.IsValid()
	require.Error(t, err, "should return error when timeout has no unit")

	sut = config.APIConfig{Port: 100, Timeout: "111mo"}
	err = sut.IsValid()
	require.Error(t, err, "should return error when timeout has invalid unit")

	sut = config.APIConfig{Port: 100, Timeout: "111y"}
	err = sut.IsValid()
	require.Error(t, err, "should return error when timeout has invalid unit")

	sut = config.APIConfig{Port: 100, Timeout: "-111s"}
	err = sut.IsValid()
	require.Error(t, err, "should return error when timeout has invalid unit")

	sut = config.APIConfig{Port: 100, Timeout: "111m"}
	err = sut.IsValid()
	assert.NoError(t, err, "should return NO error when every option is correct")
}

func TestApiDefaultValues(t *testing.T) {
	sut := config.APIConfig{}
	newSut := sut.FillDefaults()

	assert.Equalf(t, config.DefaultPort, newSut.Port,
		"api port should be set to %d if the provided value is empty", config.DefaultPort)

	assert.Equalf(t, config.DefaultTimeout, newSut.Timeout,
		"api port should be set to %s if the provided value is empty", config.DefaultTimeout)
}
