package config_test

import (
	"fmt"
	"testing"

	"github.com/jademcosta/graviola/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestQueryValidate(t *testing.T) {
	sut := config.QueryConfig{}
	err := sut.IsValid()
	assert.Error(t, err, "should return error when max_samples is 0")

	sut = config.QueryConfig{MaxSamples: -1}
	err = sut.IsValid()
	assert.Error(t, err, "should return error when max_samples is < 0")

	sut = config.QueryConfig{MaxSamples: 1, LookbackDelta: "something"}
	err = sut.IsValid()
	assert.Error(t, err, "should return error when lookback_delta is not a number")

	sut = config.QueryConfig{MaxSamples: 1, LookbackDelta: "0"}
	err = sut.IsValid()
	assert.Error(t, err, "should return error when lookback_delta is 0")

	sut = config.QueryConfig{MaxSamples: 1, LookbackDelta: ""}
	err = sut.IsValid()
	assert.Error(t, err, "should return error when lookback_delta is empty")

	sut = config.QueryConfig{MaxSamples: 1, LookbackDelta: "-1ms"}
	err = sut.IsValid()
	assert.Error(t, err, "should return error when lookback_delta is <0")

	sut = config.QueryConfig{MaxSamples: 1, LookbackDelta: "0s"}
	err = sut.IsValid()
	assert.Error(t, err, "should return error when lookback_delta is 0s")

	sut = config.QueryConfig{MaxSamples: 1, LookbackDelta: "1ms"}
	err = sut.IsValid()
	assert.NoError(t, err, "should return NO error when max_samples is > 1 and lookback_delta is > 0")

	sut = config.QueryConfig{MaxSamples: 84782, LookbackDelta: "17m"}
	err = sut.IsValid()
	assert.NoError(t, err, "should return NO error when max_samples is > 1 and lookback_delta is > 0")
}

func TestQueryDefaultValues(t *testing.T) {
	sut := config.QueryConfig{}
	newSut := sut.FillDefaults()

	assert.Equal(t, config.DefaultQueryMaxSamples, newSut.MaxSamples,
		fmt.Sprintf("query max_samples should be set to %d if the provided value is empty", config.DefaultQueryMaxSamples))
}
