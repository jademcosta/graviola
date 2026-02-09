package config

import (
	"fmt"
	"slices"
	"strings"
)

// FIXME: write a test to match the expected values of these consts with the ones in the factories,
// to avoid typos
// TODO append a prefix on these consts
const (
	FailStrategyFailAll         = "fail_all"
	FailStrategyPartialResponse = "partial_response"
	DefaultOnFailStrategy       = FailStrategyFailAll
)

type RemoteGroupsConfig struct {
	Name                string              `yaml:"name"`
	Remotes             []RemoteConfig      `yaml:"remotes"`
	TimeWindow          TimeWindowConfig    `yaml:"time_window"`
	OnQueryFailStrategy string              `yaml:"on_query_fail"`
	MergeStrategy       MergeStrategyConfig `yaml:"merge_strategy"`
}

func (rgc RemoteGroupsConfig) FillDefaults() RemoteGroupsConfig {
	if rgc.OnQueryFailStrategy == "" {
		rgc.OnQueryFailStrategy = DefaultOnFailStrategy
	}

	if rgc.MergeStrategy.Strategy == "" {
		rgc.MergeStrategy.Strategy = DefaultMergeStrategy
	}

	for i, remote := range rgc.Remotes {
		rgc.Remotes[i] = remote.FillDefaults()
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

	if len(rgc.Remotes) == 0 {
		return fmt.Errorf("remotes cannot be empty")
	}

	for _, remote := range rgc.Remotes {
		err := remote.IsValid()
		if err != nil {
			return err
		}
	}

	if err := rgc.MergeStrategy.IsValid(); err != nil {
		return fmt.Errorf("invalid merge strategy config on group %s: %w", rgc.Name, err)
	}

	return rgc.ensureNonDuplicatedRemoteNames()
}

func (rgc RemoteGroupsConfig) ensureNonDuplicatedRemoteNames() error {
	seen := make(map[string]bool)
	for _, remote := range rgc.Remotes {
		if seen[remote.Name] {
			return fmt.Errorf("remote name %s is duplicated", remote.Name)
		}
		seen[remote.Name] = true
	}

	return nil
}

func listSupportedFailureStrategies() []string {
	return []string{FailStrategyFailAll, FailStrategyPartialResponse}
}
