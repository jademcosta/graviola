package config_test

import (
	"fmt"
	"testing"

	"github.com/jademcosta/graviola/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestApiValidate(t *testing.T) {
	sut := config.ApiConfig{Port: 0}
	err := sut.IsValid()
	assert.Error(t, err, "should return error when port is zero")

	sut = config.ApiConfig{Port: 100}
	err = sut.IsValid()
	assert.Error(t, err, "should return error when timeout is empty")

	sut = config.ApiConfig{Port: 100, Timeout: "111"}
	err = sut.IsValid()
	assert.Error(t, err, "should return error when timeout has no unit")

	sut = config.ApiConfig{Port: 100, Timeout: "111mo"}
	err = sut.IsValid()
	assert.Error(t, err, "should return error when timeout has invalid unit")

	sut = config.ApiConfig{Port: 100, Timeout: "111y"}
	err = sut.IsValid()
	assert.Error(t, err, "should return error when timeout has invalid unit")

	sut = config.ApiConfig{Port: 100, Timeout: "-111s"}
	err = sut.IsValid()
	assert.Error(t, err, "should return error when timeout has invalid unit")

	sut = config.ApiConfig{Port: 100, Timeout: "111m"}
	err = sut.IsValid()
	assert.NoError(t, err, "should return NO error when every option is correct")
}

func TestApiDefaultValues(t *testing.T) {
	sut := config.ApiConfig{}
	newSut := sut.FillDefaults()

	assert.Equal(t, config.DefaultPort, newSut.Port,
		fmt.Sprintf("api port should be set to %d if the provided value is empty", config.DefaultPort))

	assert.Equal(t, config.DefaultTimeout, newSut.Timeout,
		fmt.Sprintf("api port should be set to %s if the provided value is empty", config.DefaultTimeout))
}
