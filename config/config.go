package config

import (
	"fmt"
	"orange-agent/domain"
	"sync"

	"github.com/spf13/viper"
)

var (
	AppConfig *domain.Config
	Once      sync.Once
)

func NewConfig() *domain.Config {
	Once.Do(func() {
		var err error
		AppConfig, err = loadConfig("config.yaml")
		if err != nil {
			panic(err)
		}
	})
	return AppConfig
}
func loadConfig(configPath string) (*domain.Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml") // 或 json, toml, env 等

	// 支持环境变量覆盖
	viper.AutomaticEnv()
	viper.SetEnvPrefix("APP")

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var config domain.Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	return &config, nil
}
