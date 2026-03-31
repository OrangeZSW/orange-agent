package domain

type Config struct {
	Telegram Telegram       `mapstructure:"telegram"`
	Database DatabaseConfig `mapstructure:"database"`
	Logger   Logger         `mapstructure:"logger"`
}

type Telegram struct {
	BotToken string `mapstructure:"bot_token"`
	Proxy    string `mapstructure:"proxy"`
	Promete  string `mapstructure:"promete"`
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
