package config

import (
	"fmt"
	"slices"
	"strings"
)

const DefaultLogLevel = "info"

type LogConfig struct {
	Level string `yaml:"level"`
}

func (lc LogConfig) FillDefaults() LogConfig {
	if lc.Level == "" {
		lc.Level = DefaultLogLevel
	}

	return lc
}

func (lc LogConfig) IsValid() error {
	if !slices.Contains(listSupportedLogLevels(), strings.ToLower(lc.Level)) {
		return fmt.Errorf("unsupported log level %s", lc.Level)
	}
	return nil
}

func listSupportedLogLevels() []string {
	return []string{"debug", "info", "warn", "error"}
}
