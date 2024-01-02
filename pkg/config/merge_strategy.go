package config

import (
	"fmt"
	"slices"
)

const DefaultMergeStrategyType = "keep_biggest"

type MergeStrategyConfig struct {
	Type string `yaml:"type"`
	// Time string `yaml:"time"` //TODO: will be used in the future when dedup by time window is implemented
}

func (mergeStratConf MergeStrategyConfig) FillDefaults() MergeStrategyConfig {
	if mergeStratConf.Type == "" {
		mergeStratConf.Type = DefaultMergeStrategyType
	}

	return mergeStratConf
}

func (mergeStratConf MergeStrategyConfig) IsValid() error {
	if !slices.Contains(listSupportedMergeStrategies(), mergeStratConf.Type) {
		return fmt.Errorf("merge strategy type %s is invalid", mergeStratConf.Type)
	}

	return nil
}

func listSupportedMergeStrategies() []string {
	return []string{"keep_biggest", "always_merge"}
}
