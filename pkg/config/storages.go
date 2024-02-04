package config

import (
	"fmt"
)

type StoragesConfig struct {
	MergeConf MergeStrategyConfig `yaml:"merge_strategy"`
	Groups    []GroupsConfig      `yaml:"groups"`
}

func (storagesConf StoragesConfig) FillDefaults() StoragesConfig {
	mergeConf := storagesConf.MergeConf.FillDefaults()
	storagesConf.MergeConf = mergeConf

	for i := 0; i < len(storagesConf.Groups); i++ {
		groupConf := storagesConf.Groups[i].FillDefaults()
		storagesConf.Groups[i] = groupConf
	}

	return storagesConf
}

func (storagesConf StoragesConfig) IsValid() error {
	if len(storagesConf.Groups) == 0 {
		return fmt.Errorf("cannot have empty groups list")
	}

	err := storagesConf.MergeConf.IsValid()
	if err != nil {
		return err
	}

	for _, group := range storagesConf.Groups {
		err = group.IsValid()
		if err != nil {
			return err
		}
	}

	return storagesConf.ensureNonDuplicatedGroupNames()
}

func (storagesConf StoragesConfig) ensureNonDuplicatedGroupNames() error {
	seen := make(map[string]bool)
	for _, group := range storagesConf.Groups {
		if seen[group.Name] {
			return fmt.Errorf("group name %s is duplicated", group.Name)
		}
		seen[group.Name] = true
	}

	return nil
}
