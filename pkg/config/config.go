// Package config 负责加载和解析 config.yaml 配置文件。
package config

import (
	"github.com/spf13/viper"
)

// Config 是应用程序的根配置结构体。
type Config struct {
	Bot        BotConfig        `mapstructure:"bot"`
	Onebot     OnebotConfig     `mapstructure:"onebot"`
	Log        LogConfig        `mapstructure:"log"`
	Control    ControlConfig    `mapstructure:"control"`
	Middleware MiddlewareConfig `mapstructure:"middleware"`
}

// BotConfig 包含 bot 基础信息。
type BotConfig struct {
	Name   string `mapstructure:"name"`
	SelfID string `mapstructure:"self_id"`
}

// OnebotConfig 包含 OneBot 连接配置（正向 WebSocket）。
type OnebotConfig struct {
	WsURL       string `mapstructure:"ws_url"`       // NapCat WebSocket 服务器地址
	AccessToken string `mapstructure:"access_token"` // 访问令牌，NapCat 未设置时为空
}

// LogConfig 包含日志配置。
type LogConfig struct {
	Level string `mapstructure:"level"` // debug, info, warn, error
}

// ControlConfig 包含触发控制配置。
type ControlConfig struct {
	Mode string `mapstructure:"mode"` // mention, auto
}

// MiddlewareConfig 包含消息管线中间件配置。
type MiddlewareConfig struct {
	RateLimit RateLimitConfig `mapstructure:"rate_limit"`
}

// RateLimitConfig 包含消息限流（令牌桶）配置。
type RateLimitConfig struct {
	Rate           float64 `mapstructure:"rate"`               // 令牌补充速率（个/秒）
	Burst          int     `mapstructure:"burst"`              // 桶容量，允许的突发消息数
	MaxWaitSeconds int     `mapstructure:"max_wait_seconds"`   // 等待令牌的上限秒数，超时后回复固定消息并丢弃
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
	if cfg.Onebot.WsURL == "" {
		cfg.Onebot.WsURL = "ws://127.0.0.1:3001"
	}
	if cfg.Log.Level == "" {
		cfg.Log.Level = "info"
	}
	if cfg.Control.Mode == "" {
		cfg.Control.Mode = "mention"
	}
	if cfg.Middleware.RateLimit.Rate <= 0 {
		cfg.Middleware.RateLimit.Rate = 2 // 每群每秒补充 2 个令牌
	}
	if cfg.Middleware.RateLimit.Burst <= 0 {
		cfg.Middleware.RateLimit.Burst = 20 // 桶容量 20，吸收打字突发
	}
	if cfg.Middleware.RateLimit.MaxWaitSeconds <= 0 {
		cfg.Middleware.RateLimit.MaxWaitSeconds = 10
	}

	return &cfg, nil
}
