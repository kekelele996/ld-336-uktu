package repository

import "errors"

// 仓储层哨兵错误。
var (
	ErrNotFound = errors.New("record not found")
	ErrConflict = errors.New("record conflict")
)
