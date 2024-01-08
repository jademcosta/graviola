package config

import (
	"fmt"
)

const DefaultQueryMaxSamples = 10000

type QueryConfig struct {
	MaxSamples int `yaml:"max_samples"`
}

func (qc QueryConfig) FillDefaults() QueryConfig {
	if qc.MaxSamples == 0 {
		qc.MaxSamples = DefaultQueryMaxSamples
	}
	return qc
}

func (qc QueryConfig) IsValid() error {
	if qc.MaxSamples <= 0 {
		return fmt.Errorf("max samples cannot be <= 0")
	}

	return nil
}
