package config

import "github.com/spf13/viper"

type ConfigOption struct {
	filePath string
	vi       *viper.Viper
	*Options
}

type Options struct {
	User            string
	Pass            string
	HttpProxyAddrs  []string
	SocksProxyAddrs []string
	LogPath         string
}

func NewConfigOption(filePath string) *ConfigOption {
	configOption := &ConfigOption{
		filePath: filePath,
		vi:       viper.New(),
		Options:  &Options{},
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

	err = configOption.vi.Unmarshal(configOption.Options)
	if err != nil {
		return err
	}

	return nil
}
