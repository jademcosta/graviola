package config

import (
	"fmt"
	"slices"
	"strings"
)

// TODO append a prefix on these consts
const StrategyFailAll = "fail_all"
const StrategyPartialResponse = "partial_response"
const DefaultOnFailStrategy = StrategyFailAll

type RemoteGroupsConfig struct {
	Name                string           `yaml:"name"`
	Servers             []RemoteConfig   `yaml:"remotes"`
	TimeWindow          TimeWindowConfig `yaml:"time_window"`
	OnQueryFailStrategy string           `yaml:"on_query_fail"`
}

func (rgc RemoteGroupsConfig) FillDefaults() RemoteGroupsConfig {
	if rgc.OnQueryFailStrategy == "" {
		rgc.OnQueryFailStrategy = DefaultOnFailStrategy
	}
	return rgc
}

func (rgc RemoteGroupsConfig) IsValid() error {
	if rgc.Name == "" {
		return fmt.Errorf("group name cannot be empty")
	}

	if !slices.Contains(listSupportedFailureStrategies(), strings.ToLower(rgc.OnQueryFailStrategy)) {
		return fmt.Errorf("on_query_fail should be one of %v", listSupportedFailureStrategies())
	}

	if len(rgc.Servers) == 0 {
		return fmt.Errorf("remotes cannot be empty")
	}

	for _, remote := range rgc.Servers {
		err := remote.IsValid()
		if err != nil {
			return err
		}
	}

	return rgc.ensureNonDuplicatedRemoteNames()
}

func (rgc RemoteGroupsConfig) ensureNonDuplicatedRemoteNames() error {
	seen := make(map[string]bool)
	for _, remote := range rgc.Servers {
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
