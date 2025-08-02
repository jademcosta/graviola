package config_test

import (
	"testing"

	"github.com/jademcosta/graviola/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueryValidate(t *testing.T) {
	sut := config.QueryConfig{}.FillDefaults()
	err := sut.IsValid()
	require.NoError(t, err, "filled with defaults should be valid")

	sut = config.QueryConfig{MaxSamples: -4}.FillDefaults()
	err = sut.IsValid()
	require.Error(t, err, "should return error when max_samples is < 0")

	sut = config.QueryConfig{ConcurrentQueries: 2, LookbackDelta: "5m", Timeout: "1m", MaxSamples: -1}.FillDefaults()
	err = sut.IsValid()
	require.Error(t, err, "should return error when max_samples is < 0")

	sut = config.QueryConfig{LookbackDelta: "something"}.FillDefaults()
	err = sut.IsValid()
	require.Error(t, err, "should return error when lookback_delta is not a number")

	sut = config.QueryConfig{LookbackDelta: "0"}.FillDefaults()
	err = sut.IsValid()
	require.Error(t, err, "should return error when lookback_delta is 0")

	sut = config.QueryConfig{}.FillDefaults()
	sut.LookbackDelta = ""
	err = sut.IsValid()
	require.Error(t, err, "should return error when lookback_delta is empty")

	sut = config.QueryConfig{LookbackDelta: "-1ms"}.FillDefaults()
	err = sut.IsValid()
	require.Error(t, err, "should return error when lookback_delta is <0")

	sut = config.QueryConfig{LookbackDelta: "0s"}.FillDefaults()
	err = sut.IsValid()
	require.Error(t, err, "should return error when lookback_delta is 0s")

	sut = config.QueryConfig{}.FillDefaults()
	sut.ConcurrentQueries = 0
	err = sut.IsValid()
	require.Error(t, err, "should return error when concurrent_queries is 0")

	sut = config.QueryConfig{ConcurrentQueries: -1}.FillDefaults()
	err = sut.IsValid()
	require.Error(t, err, "should return error when concurrent_queries is < 0")

	sut = config.QueryConfig{Timeout: "111"}.FillDefaults()
	err = sut.IsValid()
	require.Error(t, err, "should return error when timeout has no unit")

	sut = config.QueryConfig{Timeout: "111mo"}.FillDefaults()
	err = sut.IsValid()
	require.Error(t, err, "should return error when timeout has invalid unit")

	sut = config.QueryConfig{Timeout: "111y"}.FillDefaults()
	err = sut.IsValid()
	require.Error(t, err, "should return error when timeout has invalid unit")

	sut = config.QueryConfig{Timeout: "-111s"}.FillDefaults()
	err = sut.IsValid()
	require.Error(t, err, "should return error when timeout is negative")

	sut = config.QueryConfig{Timeout: "-111s"}.FillDefaults()
	err = sut.IsValid()
	require.Error(t, err, "should return error when timeout is negative")

	sut = config.QueryConfig{MaxSamples: 1, LookbackDelta: "1ms", ConcurrentQueries: 1, Timeout: "3m"}
	err = sut.IsValid()
	require.NoError(t, err,
		"should return NO error when max_samples is >= 1, lookback_delta is > 0, concurrent_queries > 0 and timeout > 0")

	sut = config.QueryConfig{MaxSamples: 84782, LookbackDelta: "17m", ConcurrentQueries: 44, Timeout: "7s"}
	err = sut.IsValid()
	require.NoError(t, err,
		"should return NO error when max_samples is >= 1, lookback_delta is > 0, concurrent_queries > 0 and timeout > 0")
}

func TestQueryDefaultValues(t *testing.T) {
	sut := config.QueryConfig{}
	newSut := sut.FillDefaults()

	assert.Equal(t, config.DefaultQueryMaxSamples, newSut.MaxSamples,
		"query max_samples should be set to %d if the provided value is empty", config.DefaultQueryMaxSamples)

	assert.Equal(t, config.DefaultQueryLookbackDelta, newSut.LookbackDelta,
		"query lookback_delta should be set to %s if the provided value is empty", config.DefaultQueryLookbackDelta)

	assert.Equal(t, config.DefaultQueryConcurrentQueries, newSut.ConcurrentQueries,
		"query max_samples should be set to %d if the provided value is empty", config.DefaultQueryConcurrentQueries)

	assert.Equalf(t, config.DefaultTimeout, newSut.Timeout,
		"api port should be set to %s if the provided value is empty", config.DefaultTimeout)
}
