package config

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type GraviolaConfig struct {
	APIConf      APIConfig      `yaml:"api"`
	LogConf      LogConfig      `yaml:"log"`
	StoragesConf StoragesConfig `yaml:"storages"`
	QueryConf    QueryConfig    `yaml:"query"`
}

func Parse(data []byte) (GraviolaConfig, error) {
	parsedConf := &GraviolaConfig{}

	err := yaml.Unmarshal(data, parsedConf)
	if err != nil {
		return GraviolaConfig{}, err
	}

	return *parsedConf, nil
}

func (gravConf GraviolaConfig) FillDefaults() GraviolaConfig {
	gravConf.APIConf = gravConf.APIConf.FillDefaults()
	gravConf.LogConf = gravConf.LogConf.FillDefaults()
	gravConf.StoragesConf = gravConf.StoragesConf.FillDefaults()
	gravConf.QueryConf = gravConf.QueryConf.FillDefaults()

	return gravConf
}

func (gravConf GraviolaConfig) IsValid() error {
	err := gravConf.APIConf.IsValid()
	if err != nil {
		return err
	}

	err = gravConf.LogConf.IsValid()
	if err != nil {
		return err
	}

	err = gravConf.StoragesConf.IsValid()
	if err != nil {
		return err
	}

	err = gravConf.QueryConf.IsValid()
	if err != nil {
		return err
	}

	err = gravConf.checkGroupHasRepeatedNames()
	if err != nil {
		return err
	}

	return nil
}

func (gravConf GraviolaConfig) checkGroupHasRepeatedNames() error {
	names := make(map[string]bool)

	for _, group := range gravConf.StoragesConf.Groups {
		if names[group.Name] {
			return fmt.Errorf("repeated group name: %s", group.Name)
		}
		names[group.Name] = true
	}

	return nil
}
