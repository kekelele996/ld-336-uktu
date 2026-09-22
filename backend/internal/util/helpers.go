package util

import (
	"fmt"
	"strconv"
	"time"
)

// Uint64String 将 uint 转为字符串。
func Uint64String(v uint) string {
	return strconv.FormatUint(uint64(v), 10)
}

// HoursDuration 将小时数转为 time.Duration。
func HoursDuration(h int) time.Duration {
	return time.Duration(h) * time.Hour
}

// GenSerial 生成带前缀与时间戳的单号。
func GenSerial(prefix string) string {
	return fmt.Sprintf("%s%s%s", prefix, time.Now().Format("20060102150405"), strconv.FormatInt(time.Now().UnixNano()%100000, 10))
}

// GenBarcode 生成设备分发条码（资产编号前缀 + 随机段）。
func GenBarcode(assetCode string) string {
	return "BAR-" + assetCode + "-" + strconv.FormatInt(time.Now().UnixNano()%100000, 10)
}
