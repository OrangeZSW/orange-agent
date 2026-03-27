package config

import (
	"fmt"
	"sync"

	"github.com/spf13/viper"
)

var (
	AppConfig *Config
	Once      sync.Once
)

type Config struct {
	Telegram Telegram       `mapstructure:"telegram"`
	Database DatabaseConfig `mapstructure:"database"`
	Logger   Logger         `mapstructure:"logger"`
}

type Telegram struct {
	BotToken string `mapstructure:"bot_token"`
	Proxy    string `mapstructure:"proxy"`
}

type DatabaseConfig struct {
	Driver   string `mapstructure:"driver"`
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
}

type Logger struct {
	Level      string `mapstructure:"level"`
	Output     string `mapstructure:"output"`
	FilePath   string `mapstructure:"file_path"`
	FileName   string `mapstructure:"file_name"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxAge     int    `mapstructure:"max_age"`
	MaxBackups int    `mapstructure:"max_backups"`
	Compress   bool   `mapstructure:"compress"`
	ShowCaller bool   `mapstructure:"show_caller"`
	Module     string `mapstructure:"module"`
}

func NewConfig() *Config {
	Once.Do(func() {
		var err error
		AppConfig, err = loadConfig("config.yaml")
		if err != nil {
			panic(err)
		}
	})
	return AppConfig
}
func loadConfig(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml") // 或 json, toml, env 等

	// 支持环境变量覆盖
	viper.AutomaticEnv()
	viper.SetEnvPrefix("APP")

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	return &config, nil
}
