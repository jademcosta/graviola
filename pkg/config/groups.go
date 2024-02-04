package config

import (
	"fmt"
	"slices"
	"strings"
)

const StrategyFailAll = "fail_all"
const StrategyPartialResponse = "partial_response"
const DefaultOnFailStrategy = StrategyFailAll

type GroupsConfig struct {
	Name                string           `yaml:"name"`
	Servers             []RemoteConfig   `yaml:"remotes"`
	TimeWindow          TimeWindowConfig `yaml:"time_window"`
	OnQueryFailStrategy string           `yaml:"on_query_fail"`
}

func (gc GroupsConfig) FillDefaults() GroupsConfig {
	if gc.OnQueryFailStrategy == "" {
		gc.OnQueryFailStrategy = DefaultOnFailStrategy
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

	if len(gc.Servers) == 0 {
		return fmt.Errorf("remotes cannot be empty")
	}

	for _, remote := range gc.Servers {
		err := remote.IsValid()
		if err != nil {
			return err
		}
	}

	return gc.ensureNonDuplicatedRemoteNames()
}

func (gc GroupsConfig) ensureNonDuplicatedRemoteNames() error {
	seen := make(map[string]bool)
	for _, remote := range gc.Servers {
		if seen[remote.Name] {
			return fmt.Errorf("remote name %s is duplicated", remote.Name)
		}
		seen[remote.Name] = true
	}

	return nil
}

func listSupportedFailureStrategies() []string {
	return []string{StrategyFailAll, StrategyPartialResponse}
}
