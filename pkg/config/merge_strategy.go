package config

import (
	"fmt"
	"slices"
)

const (
	MergeStrategyAlwaysMerge = "always_merge"
	MergeStrategyKeepBiggest = "keep_biggest"
)
const DefaultMergeStrategy = MergeStrategyAlwaysMerge

type MergeStrategyConfig struct {
	Strategy string `yaml:"type"`
	// Time string `yaml:"time"` //TODO: will be used in the future when dedup by time window is implemented
}

func (mergeStratConf MergeStrategyConfig) FillDefaults() MergeStrategyConfig {
	if mergeStratConf.Strategy == "" {
		mergeStratConf.Strategy = DefaultMergeStrategy
	}

	return mergeStratConf
}

func (mergeStratConf MergeStrategyConfig) IsValid() error {
	if !slices.Contains(listSupportedMergeStrategies(), mergeStratConf.Strategy) {
		return fmt.Errorf("merge strategy Strategy %s is invalid", mergeStratConf.Strategy)
	}

	return nil
}

func listSupportedMergeStrategies() []string {
	return []string{MergeStrategyAlwaysMerge, MergeStrategyKeepBiggest}
}
