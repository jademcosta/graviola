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

// MustParse parses the configuration from the given byte slice and panics if there is an error.
// It also calls FillDefaults on the parsed configuration.
// This is useful for testing or when you want to ensure that the configuration is always valid.
func MustParse(data []byte) GraviolaConfig {
	parsedConf := &GraviolaConfig{}

	err := yaml.Unmarshal(data, parsedConf)
	if err != nil {
		panic(err)
	}

	conf := parsedConf.FillDefaults()
	err = conf.IsValid()
	if err != nil {
		panic(err)
	}

	return conf
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
