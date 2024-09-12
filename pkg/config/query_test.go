package config_test

import (
	"testing"

	"github.com/jademcosta/graviola/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueryValidate(t *testing.T) {
	sut := config.QueryConfig{}
	err := sut.IsValid()
	require.Error(t, err, "should return error when max_samples is 0")

	sut = config.QueryConfig{MaxSamples: -1}
	err = sut.IsValid()
	require.Error(t, err, "should return error when max_samples is < 0")

	sut = config.QueryConfig{MaxSamples: 1, LookbackDelta: "something"}
	err = sut.IsValid()
	require.Error(t, err, "should return error when lookback_delta is not a number")

	sut = config.QueryConfig{MaxSamples: 1, LookbackDelta: "0"}
	err = sut.IsValid()
	require.Error(t, err, "should return error when lookback_delta is 0")

	sut = config.QueryConfig{MaxSamples: 1, LookbackDelta: ""}
	err = sut.IsValid()
	require.Error(t, err, "should return error when lookback_delta is empty")

	sut = config.QueryConfig{MaxSamples: 1, LookbackDelta: "-1ms"}
	err = sut.IsValid()
	require.Error(t, err, "should return error when lookback_delta is <0")

	sut = config.QueryConfig{MaxSamples: 1, LookbackDelta: "0s"}
	err = sut.IsValid()
	require.Error(t, err, "should return error when lookback_delta is 0s")

	sut = config.QueryConfig{MaxSamples: 1, LookbackDelta: "1s"}
	err = sut.IsValid()
	require.Error(t, err, "should return error when concurrent_queries is 0")

	sut = config.QueryConfig{MaxSamples: 1, LookbackDelta: "1s", ConcurrentQueries: -1}
	err = sut.IsValid()
	require.Error(t, err, "should return error when concurrent_queries is < 0")

	sut = config.QueryConfig{MaxSamples: 1, LookbackDelta: "1ms", ConcurrentQueries: 1}
	err = sut.IsValid()
	require.NoError(t, err, "should return NO error when max_samples is > 1, lookback_delta is > 0 and concurrent_queries > 0")

	sut = config.QueryConfig{MaxSamples: 84782, LookbackDelta: "17m", ConcurrentQueries: 44}
	err = sut.IsValid()
	require.NoError(t, err, "should return NO error when max_samples is > 1, lookback_delta is > 0 and concurrent_queries > 0")
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
}
