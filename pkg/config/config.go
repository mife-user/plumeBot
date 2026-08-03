// Package config 负责加载和解析 config.yaml 配置文件。
package config

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// defaultConfigYAML 是嵌入的默认配置模板，配置文件缺失时写入磁盘。
// 注意：模板与仓库根目录 config.yaml 各自独立，新增配置字段时需同步两处。
//
//go:embed config.default.yaml
var defaultConfigYAML []byte

// 默认值常量：供各消费包在字段为空时兜底，避免默认值字符串在多处漂移。
const (
	DefaultBotName = "PlumeBot"
	DefaultWsURL   = "ws://127.0.0.1:3001"
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
	Rate           float64 `mapstructure:"rate"`             // 令牌补充速率（个/秒）
	Burst          int     `mapstructure:"burst"`            // 桶容量，允许的突发消息数
	MaxWaitSeconds int     `mapstructure:"max_wait_seconds"` // 等待令牌的上限秒数，超时后回复固定消息并丢弃
}

// Load 从 path 加载 YAML 配置文件。
// 配置文件不存在时，将嵌入的默认配置（config.default.yaml）写入 path 后再加载。
// 字段空值/缺省不做 Go 侧兜底，由各消费包自行处理默认值。
func Load(path string) (*Config, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := writeDefault(path); err != nil {
			return nil, err
		}
	}

	v := viper.New()
	v.SetConfigFile(path)
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// writeDefault 将嵌入的默认配置写入 path（自动创建父目录）。
func writeDefault(path string) error {
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("创建配置目录失败: %w", err)
		}
	}
	if err := os.WriteFile(path, defaultConfigYAML, 0o644); err != nil {
		return fmt.Errorf("写入默认配置文件 %s 失败: %w", path, err)
	}
	return nil
}
