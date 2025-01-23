package config

import (
	"fmt"
	"time"
)

const DefaultQueryMaxSamples = 100000
const DefaultQueryLookbackDelta = "5m"
const DefaultQueryConcurrentQueries = 20
const DefaultTimeout = "1m"

type QueryConfig struct {
	MaxSamples        int    `yaml:"max_samples"`
	LookbackDelta     string `yaml:"lookback_delta"`
	ConcurrentQueries int    `yaml:"max_concurrent_queries"`
	Timeout           string `yaml:"timeout"`
}

func (qc QueryConfig) FillDefaults() QueryConfig {
	if qc.MaxSamples == 0 {
		qc.MaxSamples = DefaultQueryMaxSamples
	}

	if qc.LookbackDelta == "" {
		qc.LookbackDelta = DefaultQueryLookbackDelta
	}

	if qc.ConcurrentQueries == 0 {
		qc.ConcurrentQueries = DefaultQueryConcurrentQueries
	}

	if qc.Timeout == "" {
		qc.Timeout = DefaultTimeout
	}

	return qc
}

func (qc QueryConfig) IsValid() error {
	if qc.MaxSamples <= 0 {
		return fmt.Errorf("max_samples cannot be <= 0")
	}

	if qc.LookbackDelta == "" {
		return fmt.Errorf("lookback_delta cannot empty")
	}

	parsed, err := ParseDuration(qc.LookbackDelta)
	if err != nil {
		return fmt.Errorf("error validating query lookback_delta: %w", err)
	}

	if parsed == 0 {
		return fmt.Errorf("error validating query lookback_delta: it cannot be zero")
	}

	if qc.ConcurrentQueries <= 0 {
		return fmt.Errorf("error validating query concurrent_queries: it cannot be <= 0")
	}

	_, err = ParseDuration(qc.Timeout)
	if err != nil {
		return fmt.Errorf("timeout must be a valid number: %w", err)
	}

	return nil
}

func (qc QueryConfig) LookbackDeltaDuration() time.Duration {
	parsed, err := ParseDuration(qc.LookbackDelta)
	if err != nil {
		panic(err)
	}

	return parsed
}

func (qc QueryConfig) TimeoutDuration() time.Duration {
	parsed, err := ParseDuration(qc.Timeout)
	if err != nil {
		panic(err)
	}

	return parsed
}
