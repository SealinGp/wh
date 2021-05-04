package self_report

import (
	"github.com/spf13/viper"
)

func ConfigInit(configPath string) (*ConfigOpt, error) {
	cfg := &ConfigOpt{}
	viper.SetConfigType("yml")
	viper.SetConfigFile(configPath)
	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	err = viper.Unmarshal(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
