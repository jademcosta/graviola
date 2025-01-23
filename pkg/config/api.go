package config

import (
	"fmt"
)

const DefaultPort = 9197

type APIConfig struct {
	Port int `yaml:"port"`
}

func (apiConf APIConfig) FillDefaults() APIConfig {
	if apiConf.Port == 0 {
		apiConf.Port = DefaultPort
	}

	return apiConf
}

func (apiConf APIConfig) IsValid() error {
	if apiConf.Port == 0 {
		return fmt.Errorf("port cannot be zero")
	}

	return nil
}
