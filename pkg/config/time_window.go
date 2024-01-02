package config

type TimeWindowConfig struct {
	Start string `yaml:"start"`
	End   string `yaml:"end"`
}

//FIXME: implement

func (tWindowConf TimeWindowConfig) IsValid() error {

	return nil
}
