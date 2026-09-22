package service

import (
	"net/http"
	"testing"

	"github.com/medasset/medasset/internal/constants"
	"github.com/medasset/medasset/internal/dto"
	"github.com/medasset/medasset/internal/repository"
	"github.com/medasset/medasset/internal/util"
)

func TestDeviceServiceCreateDisableEnable(t *testing.T) {
	env := newTestServiceEnv(t)
	svc := NewDeviceService(repository.NewDeviceRepository(env.db), env.audit, env.logger)

	dev, err := svc.Create(&dto.CreateDeviceReq{AssetCode: "MA-100", Name: "呼吸机", Category: "生命支持", Department: "ICU"}, "admin")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if dev.Barcode == "" {
		t.Error("expected barcode generated")
	}
	if dev.Status != constants.DeviceStatusInStorage {
		t.Errorf("status = %q", dev.Status)
	}

	// 重复资产编号。
	if _, err := svc.Create(&dto.CreateDeviceReq{AssetCode: "MA-100", Name: "呼吸机2"}, "admin"); err == nil {
		t.Error("expected duplicate asset code error")
	}

	// 禁用/启用。
	disabled, err := svc.Disable(dev.ID, "admin")
	if err != nil {
		t.Fatalf("Disable failed: %v", err)
	}
	if disabled.Status != constants.DeviceStatusDisabled {
		t.Errorf("status = %q", disabled.Status)
	}
	enabled, err := svc.Enable(dev.ID, "admin")
	if err != nil {
		t.Fatalf("Enable failed: %v", err)
	}
	if enabled.Status != constants.DeviceStatusInStorage {
		t.Errorf("status = %q", enabled.Status)
	}

	// 列表。
	result, err := svc.List(1, 10, "", "", "", "")
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("total = %d, want 1", result.Total)
	}

	// 不存在的设备。
	if _, err := svc.Get(999); err == nil {
		t.Error("expected not found error")
	} else if appErr, ok := err.(*util.AppError); ok && appErr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", appErr.Code)
	}
}
