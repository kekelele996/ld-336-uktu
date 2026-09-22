package util

import "strconv"

// ParsePage 解析分页参数，默认第 1 页。
func ParsePage(raw string) int {
	p, err := strconv.Atoi(raw)
	if err != nil || p < 1 {
		return 1
	}
	return p
}

// ParsePageSize 解析每页条数，默认 10，最大 200。
func ParsePageSize(raw string) int {
	s, err := strconv.Atoi(raw)
	if err != nil || s < 1 {
		return 10
	}
	if s > 200 {
		return 200
	}
	return s
}

// Offset 计算分页偏移量。
func Offset(page, pageSize int) int {
	return (page - 1) * pageSize
}
