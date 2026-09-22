package util

import (
	"testing"
	"time"
)

func TestFormatDate(t *testing.T) {
	tm := time.Date(2025, 6, 1, 10, 0, 0, 0, time.Local)
	if got := FormatDate(&tm); got != "2025-06-01" {
		t.Errorf("FormatDate = %q, want 2025-06-01", got)
	}
	if got := FormatDate(nil); got != "-" {
		t.Errorf("FormatDate(nil) = %q, want -", got)
	}
}

func TestStatusTexts(t *testing.T) {
	cases := []struct {
		got, want string
	}{
		{DeviceStatusText("in_storage"), "在库"},
		{DeviceStatusText("scrapped"), "已报废"},
		{PurchaseStatusText("pending_dean"), "待院长审批"},
		{MaintenanceTypeText("monthly"), "月检"},
		{MaintenanceStatusText("completed"), "已完成"},
		{CalibrationStatusText("unqualified"), "不合格"},
		{TransferStatusText("approved"), "已批准"},
		{ScrapStatusText("rejected"), "已驳回"},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("got %q, want %q", c.got, c.want)
		}
	}
}
