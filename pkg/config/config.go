package config

import "gopkg.in/yaml.v3"

type GraviolaConfig struct {
	ApiConf      ApiConfig      `yaml:"api"`
	LogConf      LogConfig      `yaml:"log"`
	StoragesConf StoragesConfig `yaml:"storages"`
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
	gravConf.ApiConf = gravConf.ApiConf.FillDefaults()
	gravConf.LogConf = gravConf.LogConf.FillDefaults()
	gravConf.StoragesConf = gravConf.StoragesConf.FillDefaults()

	return gravConf
}
