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
	require.NoError(t, err, "should return NO error when every option is correct")

	sut = config.APIConfig{}.FillDefaults()
	err = sut.IsValid()
	require.NoError(t, err, "filled with defaults should be valid")
}

func TestApiDefaultValues(t *testing.T) {
	sut := config.APIConfig{}
	newSut := sut.FillDefaults()

	assert.Equalf(t, config.DefaultPort, newSut.Port,
		"api port should be set to %d if the provided value is empty", config.DefaultPort)
}
