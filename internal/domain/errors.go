// Package domain 定义领域层公共错误哨兵，供全部 infra 和 service 层引用。
package domain

import "errors"

var (
	// ErrNotFound 表示请求的资源不存在（记录未找到、文件缺失等）。
	ErrNotFound = errors.New("not found")

	// ErrConflict 表示唯一约束或状态冲突（重复插入、乐观锁冲突等）。
	ErrConflict = errors.New("conflict")

	// ErrClosed 表示资源已关闭、连接已断开。
	ErrClosed = errors.New("closed")

	// ErrRateLimited 表示消息在限流中间件等待令牌超时，已被丢弃。
	// 由连接层识别后回复固定文案，其余层无需处理。
	ErrRateLimited = errors.New("rate limited")

	// ErrSensitiveWord 表示消息命中敏感词，已被拦截丢弃。
	// 由连接层识别后回复固定文案；命中词见 SensitiveWordError.Word。
	ErrSensitiveWord = errors.New("sensitive word")
)

// SensitiveWordError 是携带命中词的敏感词拦截错误。
// 通过 errors.Is(err, ErrSensitiveWord) 判断类型，errors.As 取命中词。
type SensitiveWordError struct {
	Word string // 命中的敏感词（保留配置中的原始大小写，便于日志）
}

func (e *SensitiveWordError) Error() string {
	return "sensitive word: " + e.Word
}

// Unwrap 使 errors.Is 能匹配到 ErrSensitiveWord 哨兵。
func (e *SensitiveWordError) Unwrap() error {
	return ErrSensitiveWord
}
