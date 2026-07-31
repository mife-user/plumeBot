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
)
