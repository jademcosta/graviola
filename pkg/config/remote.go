package config

import (
	"fmt"
	"regexp"
)

type RemoteConfig struct {
	Name           string           `yaml:"name"`
	Address        string           `yaml:"address"`
	PathPrefix     string           `yaml:"path_prefix"`
	TimeWindowConf TimeWindowConfig `yaml:"time_window"`
}

func (sc RemoteConfig) FillDefaults() RemoteConfig {
	return sc
}

func (sc RemoteConfig) IsValid() error {
	if sc.Address == "" {
		return fmt.Errorf("address of server cannot be nil")
	}

	if sc.Name == "" {
		return fmt.Errorf("name of server cannot be nil")
	}

	rex, _ := regexp.Compile("^https?://.+$")

	if !rex.Match([]byte(sc.Address)) {
		return fmt.Errorf("address should start with http:// or https://")
	}

	return nil
}
