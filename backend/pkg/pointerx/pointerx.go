// Package pointerx 提供无业务依赖的指针构造工具。
package pointerx

import "time"

// TimePtr 返回 time.Time 的指针。
func TimePtr(t time.Time) *time.Time { return &t }

// BoolPtr 返回 bool 的指针。
func BoolPtr(b bool) *bool { return &b }
