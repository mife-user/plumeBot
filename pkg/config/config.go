// Package config 负责加载和解析 config.yaml 配置文件。
package config

import (
	"github.com/spf13/viper"
)

// Config 是应用程序的根配置结构体。
type Config struct {
	Bot     BotConfig     `mapstructure:"bot"`
	Log     LogConfig     `mapstructure:"log"`
	Control ControlConfig `mapstructure:"control"`
}

// BotConfig 包含 bot 基础信息。
type BotConfig struct {
	Name   string `mapstructure:"name"`
	SelfID string `mapstructure:"self_id"`
}

// LogConfig 包含日志配置。
type LogConfig struct {
	Level string `mapstructure:"level"` // debug, info, warn, error
}

// ControlConfig 包含触发控制配置。
type ControlConfig struct {
	Mode string `mapstructure:"mode"` // mention, auto
}

// Load 从 path 加载 YAML 配置文件，返回解析后的 Config。
// 默认值由 config.yaml 提供；Go 侧仅对缺失的关键字段做兜底。
func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	// 兜底：config.yaml 可能被删字段或留空，保证关键字段有值
	if cfg.Bot.Name == "" {
		cfg.Bot.Name = "PlumeBot"
	}
	if cfg.Log.Level == "" {
		cfg.Log.Level = "info"
	}
	if cfg.Control.Mode == "" {
		cfg.Control.Mode = "mention"
	}

	return &cfg, nil
}
