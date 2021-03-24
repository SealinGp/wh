package config

import "github.com/spf13/viper"

type ConfigOption struct {
	filePath string
	vi       *viper.Viper
	options  *Options
}

type Options struct {
	User string
	Pass string
}

func NewConfigOption(filePath string) *ConfigOption {
	configOption := &ConfigOption{
		filePath: filePath,
		vi:       viper.New(),
		options:  &Options{},
	}

	return configOption
}

func (configOption *ConfigOption) Init() error {
	configOption.vi.SetConfigType("yml")
	configOption.vi.SetConfigFile(configOption.filePath)
	err := configOption.vi.ReadInConfig()
	if err != nil {
		return err
	}

	err = configOption.vi.Unmarshal(configOption.options)
	if err != nil {
		return err
	}

	return nil
}

func (configOption *ConfigOption) GetOptions() *Options {
	return configOption.options
}
