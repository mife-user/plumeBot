// Package onebot 封装 ZeroBot OneBot v11 连接层。
// 第二阶段实现：正向 WebSocket 连接 NapCat，接收原始事件 → 转换为 domain 实体 → 交给 handler。
package onebot

import (
	"context"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/driver"
	log "github.com/sirupsen/logrus"

	"plumebot/internal/handler"
	"plumebot/pkg/config"
	"plumebot/pkg/logger"
)

// Client 是 ZeroBot 连接层的封装，负责连接 NapCat 并将事件分发给 handler。
type Client struct {
	cfg    config.OnebotConfig
	msg    *handler.MessageHandler
	notice *handler.NoticeHandler
}

// New 创建 Client，注入消息/通知事件处理入口。
// logLevel 用于对齐 ZeroBot 内部 logrus 日志级别（debug|info|warn|error）。
func New(cfg config.OnebotConfig, logLevel string, msg *handler.MessageHandler, notice *handler.NoticeHandler) *Client {
	log.SetLevel(parseLogLevel(logLevel))
	return &Client{cfg: cfg, msg: msg, notice: notice}
}

// Run 注册事件分发并启动连接。ZeroBot 底层自动处理断线重连，本方法阻塞运行，不返回。
func (c *Client) Run() {
	c.registerMatchers()
	zero.RunAndBlock(&zero.Config{
		Driver: []zero.Driver{
			driver.NewWebSocketClient(c.cfg.WsURL, c.cfg.AccessToken),
		},
	}, nil)
}

// registerMatchers 注册 ZeroBot 事件匹配器：事件 → domain 实体 → handler。
func (c *Client) registerMatchers() {
	zero.OnMessage().Handle(func(ctx *zero.Ctx) {
		msg, ok := toMessage(ctx.Event)
		if !ok {
			logger.Warn("忽略不支持的 message 事件",
				logger.S("post_type", ctx.Event.PostType),
				logger.S("message_type", ctx.Event.MessageType),
			)
			return
		}
		logger.Info("收到消息事件",
			logger.S("message_id", msg.MessageID),
			logger.S("group_id", msg.GroupID),
			logger.S("user_id", msg.UserID),
			logger.S("message_type", msg.MessageType),
			logger.S("content", msg.Content),
		)
		if err := c.msg.Handle(context.Background(), msg); err != nil {
			logger.Warn("消息处理失败", logger.Err(err))
		}
	})

	zero.OnNotice().Handle(func(ctx *zero.Ctx) {
		c.dispatchEvent(ctx)
	})
	zero.OnRequest().Handle(func(ctx *zero.Ctx) {
		c.dispatchEvent(ctx)
	})
	zero.OnMetaEvent().Handle(func(ctx *zero.Ctx) {
		logger.Debug("收到元事件",
			logger.S("meta_event_type", ctx.Event.RawEvent.Get("meta_event_type").String()),
		)
	})
}

// dispatchEvent 转换并分发通知/请求事件。
func (c *Client) dispatchEvent(ctx *zero.Ctx) {
	evt, ok := toEvent(ctx.Event)
	if !ok {
		logger.Warn("忽略不支持的事件", logger.S("post_type", ctx.Event.PostType))
		return
	}
	logger.Info("收到通知事件",
		logger.S("type", string(evt.Type)),
		logger.S("sub_type", evt.SubType),
		logger.S("group_id", evt.GroupID),
		logger.S("user_id", evt.UserID),
	)
	if err := c.notice.Handle(context.Background(), evt); err != nil {
		logger.Warn("通知处理失败", logger.Err(err))
	}
}

// parseLogLevel 将配置的日志级别映射为 logrus 级别，无效值降级为 info。
func parseLogLevel(s string) log.Level {
	switch s {
	case "debug":
		return log.DebugLevel
	case "warn":
		return log.WarnLevel
	case "error":
		return log.ErrorLevel
	default:
		return log.InfoLevel
	}
}
