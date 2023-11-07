package config

import (
	"fmt"
	"slices"
	"strings"
)

const StrategyFailAll = "fail_all"
const StrategyPartialResponse = "partial_response"
const DefaultStrategy = StrategyFailAll

type GroupsConfig struct {
	Name                string           `yaml:"name"`
	Servers             []RemoteConfig   `yaml:"remotes"`
	TimeWindow          TimeWindowConfig `yaml:"time_window"`
	OnQueryFailStrategy string           `yaml:"on_query_fail"`
}

func (gc GroupsConfig) FillDefaults() GroupsConfig {
	if gc.OnQueryFailStrategy == "" {
		gc.OnQueryFailStrategy = DefaultStrategy
	}
	return gc
}

func (gc GroupsConfig) IsValid() error {
	if gc.Name == "" {
		return fmt.Errorf("group name cannot be empty")
	}

	if !slices.Contains(listSupportedFailureStrategies(), strings.ToLower(gc.OnQueryFailStrategy)) {
		return fmt.Errorf("on_query_fail should be one of %v", listSupportedFailureStrategies())
	}
	return nil
}

func listSupportedFailureStrategies() []string {
	return []string{StrategyFailAll, StrategyPartialResponse}
}
