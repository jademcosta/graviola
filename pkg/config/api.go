package config

import (
	"fmt"
	"time"
)

// const StrategyFailAll = "fail_all"
// const StrategyPartialResponse = "partial_response"
const DefaultPort = 9197
const DefaultTimeout = "1m"

type ApiConfig struct {
	Port    int    `yaml:"port"`
	Timeout string `yaml:"timeout"`
}

func (apiConf ApiConfig) FillDefaults() ApiConfig {
	if apiConf.Port == 0 {
		apiConf.Port = DefaultPort
	}

	if apiConf.Timeout == "" {
		apiConf.Timeout = DefaultTimeout
	}

	return apiConf
}

func (apiConf ApiConfig) IsValid() error {
	if apiConf.Port == 0 {
		return fmt.Errorf("port cannot be zero")
	}

	_, err := ParseDuration(apiConf.Timeout)
	if err != nil {
		return fmt.Errorf("error validating api config: %w", err)
	}

	return nil
}

func (apiConf ApiConfig) TimeoutDuration() time.Duration {
	parsed, err := ParseDuration(apiConf.Timeout)
	if err != nil {
		panic(err)
	}

	return parsed
}
