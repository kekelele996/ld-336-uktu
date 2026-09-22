package service

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/medasset/medasset/internal/constants"
	"github.com/medasset/medasset/internal/dto"
	"github.com/medasset/medasset/internal/model"
	"github.com/medasset/medasset/internal/repository"
	"github.com/medasset/medasset/internal/util"
)

func newMaintenanceLinkSvc(t *testing.T, env *testEnv) *MaintenanceService {
	t.Helper()
	return NewMaintenanceService(
		repository.NewMaintenanceRepository(env.db),
		repository.NewDeviceRepository(env.db),
		repository.NewCalibrationRepository(env.db),
		env.audit,
		env.logger,
	)
}

func newRepairDevice(t *testing.T, env *testEnv) *model.Device {
	t.Helper()
	ds := NewDeviceService(repository.NewDeviceRepository(env.db), env.audit, env.logger)
	d, err := ds.Create(&dto.CreateDeviceReq{AssetCode: uniqueAssetCode(), Name: "除颤仪", Category: "生命支持", Department: "急诊科"}, "admin")
	if err != nil {
		t.Fatalf("create device: %v", err)
	}
	if _, err := ds.Enable(d.ID, "admin"); err != nil {
		t.Fatalf("enable device: %v", err)
	}
	if err := env.db.Model(&model.Device{}).Where("id = ?", d.ID).
		Update("status", constants.DeviceStatusInUse).Error; err != nil {
		t.Fatalf("set in_use: %v", err)
	}
	d2, err := repository.NewDeviceRepository(env.db).FindByID(d.ID)
	if err != nil {
		t.Fatalf("reload device: %v", err)
	}
	return d2
}

var assetCodeSeq uint64

func uniqueAssetCode() string {
	assetCodeSeq++
	return fmt.Sprintf("MA-LINK-%d-%d", time.Now().UnixNano(), assetCodeSeq)
}

func createRepairOrder(t *testing.T, svc *MaintenanceService, deviceID uint) *model.MaintenanceRecord {
	t.Helper()
	m, err := svc.Create(&dto.CreateMaintenanceReq{
		DeviceID:         deviceID,
		Type:             constants.MaintenanceTypeRepair,
		FaultDescription: "无法开机",
	}, "admin")
	if err != nil {
		t.Fatalf("create repair: %v", err)
	}
	return m
}

func startRepair(t *testing.T, svc *MaintenanceService, id uint) *dto.MaintenanceActionResp {
	t.Helper()
	resp, err := svc.Start(id, &dto.StartMaintenanceReq{Engineer: "王工"}, "admin")
	if err != nil {
		t.Fatalf("start repair: %v", err)
	}
	return resp
}

func completeRepair(t *testing.T, svc *MaintenanceService, id uint) *dto.MaintenanceActionResp {
	t.Helper()
	resp, err := svc.Complete(id, &dto.CompleteMaintenanceReq{Content: "更换电源板"}, "admin")
	if err != nil {
		t.Fatalf("complete repair: %v", err)
	}
	return resp
}

func deviceStatus(t *testing.T, env *testEnv, id uint) string {
	t.Helper()
	d, err := repository.NewDeviceRepository(env.db).FindByID(id)
	if err != nil {
		t.Fatalf("find device: %v", err)
	}
	return d.Status
}

func orderStatus(t *testing.T, env *testEnv, id uint) string {
	t.Helper()
	m, err := repository.NewMaintenanceRepository(env.db).FindByID(id)
	if err != nil {
		t.Fatalf("find maintenance: %v", err)
	}
	return m.Status
}

// 开始维修前，同一设备存在其他待处理维修工单时整次拒绝；取消多余工单后原工单方可开始。
func TestRepairStartBlockedByAnotherPendingRepair(t *testing.T) {
	env := newTestServiceEnv(t)
	svc := newMaintenanceLinkSvc(t, env)
	d := newRepairDevice(t, env)

	first := createRepairOrder(t, svc, d.ID)
	second := createRepairOrder(t, svc, d.ID)

	// 两张待处理维修工单并存时，任一工单的开始都整次拒绝。
	if _, err := svc.Start(first.ID, &dto.StartMaintenanceReq{Engineer: "王工"}, "admin"); err == nil {
		t.Fatal("expected first start blocked by the other pending repair")
	} else if appErr, ok := err.(*util.AppError); !ok || appErr.Code != http.StatusConflict {
		t.Fatalf("expected 409 conflict, got %v", err)
	}
	if _, err := svc.Start(second.ID, &dto.StartMaintenanceReq{Engineer: "李工"}, "admin"); err == nil {
		t.Fatal("expected second start blocked by the other pending repair")
	} else if appErr, ok := err.(*util.AppError); !ok || appErr.Code != http.StatusConflict {
		t.Fatalf("expected 409 conflict, got %v", err)
	}
	// 两张工单与设备保持原样。
	if got := deviceStatus(t, env, d.ID); got != constants.DeviceStatusInUse {
		t.Errorf("device status changed to %q, want in_use", got)
	}
	if got := orderStatus(t, env, first.ID); got != constants.MaintenanceStatusPending {
		t.Errorf("first order status = %q, want pending (unchanged)", got)
	}
	if got := orderStatus(t, env, second.ID); got != constants.MaintenanceStatusPending {
		t.Errorf("second order status = %q, want pending (unchanged)", got)
	}

	// 取消重复报修工单后，原工单开始成功，设备进入维修中。
	if _, err := svc.Cancel(second.ID, &dto.CancelMaintenanceReq{Reason: "重复报修"}, "admin"); err != nil {
		t.Fatalf("cancel duplicate: %v", err)
	}
	resp := startRepair(t, svc, first.ID)
	if resp.DeviceStatus != constants.DeviceStatusUnderMaintenance {
		t.Fatalf("device status = %q, want under_maintenance", resp.DeviceStatus)
	}

	// 维修进行中再创建待处理维修工单，其开始同样被整次拒绝。
	third := createRepairOrder(t, svc, d.ID)
	if _, err := svc.Start(third.ID, &dto.StartMaintenanceReq{Engineer: "李工"}, "admin"); err == nil {
		t.Fatal("expected third start blocked by in-progress repair")
	}
	if got := orderStatus(t, env, third.ID); got != constants.MaintenanceStatusPending {
		t.Errorf("third order status = %q, want pending (unchanged)", got)
	}
	if got := deviceStatus(t, env, d.ID); got != constants.DeviceStatusUnderMaintenance {
		t.Errorf("device status = %q, want under_maintenance", got)
	}
}

// 已处理中的工单重复开始只能得到阻塞错误，不产生第二次状态流转。
func TestRepairStartIdempotent(t *testing.T) {
	env := newTestServiceEnv(t)
	svc := newMaintenanceLinkSvc(t, env)
	d := newRepairDevice(t, env)
	m := createRepairOrder(t, svc, d.ID)

	startRepair(t, svc, m.ID)

	_, err := svc.Start(m.ID, &dto.StartMaintenanceReq{Engineer: "王工"}, "admin")
	if err == nil {
		t.Fatal("expected duplicate start to be rejected")
	}
	appErr, ok := err.(*util.AppError)
	if !ok || appErr.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %v", err)
	}
	if got := orderStatus(t, env, m.ID); got != constants.MaintenanceStatusInProgress {
		t.Errorf("order status = %q, want in_progress", got)
	}
	if got := deviceStatus(t, env, d.ID); got != constants.DeviceStatusUnderMaintenance {
		t.Errorf("device status = %q", got)
	}
}

// 待处理工单重复提交开始：自身仍 pending 时不存在其他阻塞工单，第一次生效后第二次必须失败。
func TestConcurrentStartOnlyOneSucceeds(t *testing.T) {
	env := newTestServiceEnv(t)
	svc := newMaintenanceLinkSvc(t, env)
	d := newRepairDevice(t, env)
	m := createRepairOrder(t, svc, d.ID)

	// 串行模拟两次并发提交（SQLite 测试库连接池固定 1 连接，两次调用被串行化；
	// MySQL 下由设备行锁保证同样语义）：只有一次开始生效，另一次收到冲突。
	res1, err1 := svc.Start(m.ID, &dto.StartMaintenanceReq{Engineer: "王工"}, "admin")
	_, err2 := svc.Start(m.ID, &dto.StartMaintenanceReq{Engineer: "王工"}, "admin")

	if res1 == nil || err1 != nil {
		t.Fatalf("first start should succeed, got resp=%v err=%v", res1, err1)
	}
	if err2 == nil {
		t.Fatal("duplicate concurrent start should be rejected")
	}
	if appErr, ok := err2.(*util.AppError); !ok || appErr.Code != http.StatusConflict {
		t.Fatalf("expected 409 conflict on duplicate start, got %v", err2)
	}
	if got := orderStatus(t, env, m.ID); got != constants.MaintenanceStatusInProgress {
		t.Errorf("order status = %q, want in_progress", got)
	}
	if got := deviceStatus(t, env, d.ID); got != constants.DeviceStatusUnderMaintenance {
		t.Errorf("device status = %q, want under_maintenance", got)
	}
}

// 维修期间，调拨和报废申请一并拒绝，原工单、设备与申请保持原样。
func TestTransferScrapRejectedUnderMaintenance(t *testing.T) {
	env := newTestServiceEnv(t)
	mainSvc := newMaintenanceLinkSvc(t, env)
	transferSvc := NewTransferService(repository.NewTransferRepository(env.db), repository.NewDeviceRepository(env.db), env.audit, env.logger)
	scrapSvc := NewScrapService(repository.NewScrapRepository(env.db), repository.NewDeviceRepository(env.db), env.audit, env.logger)
	d := newRepairDevice(t, env)
	m := createRepairOrder(t, mainSvc, d.ID)
	startRepair(t, mainSvc, m.ID)

	if _, err := transferSvc.Create(&dto.CreateTransferReq{DeviceID: d.ID, ToDepartment: "ICU"}, "zhang"); err == nil {
		t.Error("expected transfer create blocked during maintenance")
	} else if appErr, ok := err.(*util.AppError); !ok || appErr.Code != http.StatusConflict {
		t.Errorf("expected 409, got %v", err)
	}
	if _, err := scrapSvc.Create(&dto.CreateScrapReq{DeviceID: d.ID, Reason: "损坏"}, "zhang"); err == nil {
		t.Error("expected scrap create blocked during maintenance")
	} else if appErr, ok := err.(*util.AppError); !ok || appErr.Code != http.StatusConflict {
		t.Errorf("expected 409, got %v", err)
	}

	if got := deviceStatus(t, env, d.ID); got != constants.DeviceStatusUnderMaintenance {
		t.Errorf("device status = %q, want under_maintenance", got)
	}
	if n := countTransfers(env, d.ID); n != 0 {
		t.Errorf("transfers created = %d, want 0 (request must not be persisted)", n)
	}
	if n := countScraps(env, d.ID); n != 0 {
		t.Errorf("scraps created = %d, want 0 (request must not be persisted)", n)
	}
	if got := orderStatus(t, env, m.ID); got != constants.MaintenanceStatusInProgress {
		t.Errorf("maintenance order status = %q, want in_progress (unchanged)", got)
	}
}

// 维修开始前已存在的待审批调拨/报废申请，在维修期间审批也必须被拒绝，申请与设备保持原样。
func TestTransferScrapApproveRejectedUnderMaintenance(t *testing.T) {
	env := newTestServiceEnv(t)
	mainSvc := newMaintenanceLinkSvc(t, env)
	transferSvc := NewTransferService(repository.NewTransferRepository(env.db), repository.NewDeviceRepository(env.db), env.audit, env.logger)
	scrapSvc := NewScrapService(repository.NewScrapRepository(env.db), repository.NewDeviceRepository(env.db), env.audit, env.logger)
	d := newRepairDevice(t, env)

	tr, err := transferSvc.Create(&dto.CreateTransferReq{DeviceID: d.ID, ToDepartment: "ICU"}, "zhang")
	if err != nil {
		t.Fatalf("create transfer: %v", err)
	}
	sr, err := scrapSvc.Create(&dto.CreateScrapReq{DeviceID: d.ID, Reason: "老化"}, "zhang")
	if err != nil {
		t.Fatalf("create scrap: %v", err)
	}

	m := createRepairOrder(t, mainSvc, d.ID)
	startRepair(t, mainSvc, m.ID)

	if _, err := transferSvc.Approve(tr.ID, &dto.TransferApproveReq{}, "admin"); err == nil {
		t.Error("expected transfer approve blocked")
	}
	if _, err := scrapSvc.Approve(sr.ID, &dto.ScrapApproveReq{}, "admin"); err == nil {
		t.Error("expected scrap approve blocked")
	}
	// 申请仍为待审批，设备仍维修中，科室未被调拨改动。
	if got := requestTransferStatus(env, tr.ID); got != constants.TransferStatusPending {
		t.Errorf("transfer status = %q, want pending", got)
	}
	if got := requestScrapStatus(env, sr.ID); got != constants.ScrapStatusPending {
		t.Errorf("scrap status = %q, want pending", got)
	}
	if got := deviceStatus(t, env, d.ID); got != constants.DeviceStatusUnderMaintenance {
		t.Errorf("device status = %q", got)
	}
	dd, _ := repository.NewDeviceRepository(env.db).FindByID(d.ID)
	if dd.Department != "急诊科" {
		t.Errorf("department changed to %q during blocked transfer approve", dd.Department)
	}
}

// 完成维修时最新计量结果为不合格：设备保持禁用并提示未通过计量。
func TestCompleteRepairUnqualifiedCalibrationKeepsDisabled(t *testing.T) {
	env := newTestServiceEnv(t)
	svc := newMaintenanceLinkSvc(t, env)
	calSvc := NewCalibrationService(repository.NewCalibrationRepository(env.db), repository.NewDeviceRepository(env.db), env.audit, env.logger)
	d := newRepairDevice(t, env)

	// 纳入计量并登记一次不合格结果（设备被自动禁用）。
	c, err := calSvc.Create(&dto.CreateCalibrationReq{
		InstrumentNo: "JL-001", DeviceID: d.ID, CalibrationCycleMonths: 12,
	}, "admin")
	if err != nil {
		t.Fatalf("create calibration: %v", err)
	}
	if _, err := calSvc.RecordResult(c.ID, &dto.CalibrationResultReq{Result: constants.CalibrationResultUnqualified}, "admin"); err != nil {
		t.Fatalf("record result: %v", err)
	}

	m := createRepairOrder(t, svc, d.ID)
	startRepair(t, svc, m.ID)
	resp, err := svc.Complete(m.ID, &dto.CompleteMaintenanceReq{Content: "维修完成"}, "admin")
	if err != nil {
		t.Fatalf("complete: %v", err)
	}
	if resp.DeviceStatus != constants.DeviceStatusDisabled {
		t.Errorf("device status = %q, want disabled", resp.DeviceStatus)
	}
	if resp.Notice != constants.MsgRepairCalibrationUnqualified {
		t.Errorf("notice = %q, want unqualified calibration notice", resp.Notice)
	}
	if got := deviceStatus(t, env, d.ID); got != constants.DeviceStatusDisabled {
		t.Errorf("persisted device status = %q, want disabled", got)
	}
}

// 计量合格后完成维修：设备恢复使用中。
func TestCompleteRepairQualifiedCalibrationRestoresInUse(t *testing.T) {
	env := newTestServiceEnv(t)
	svc := newMaintenanceLinkSvc(t, env)
	calSvc := NewCalibrationService(repository.NewCalibrationRepository(env.db), repository.NewDeviceRepository(env.db), env.audit, env.logger)
	d := newRepairDevice(t, env)

	c, err := calSvc.Create(&dto.CreateCalibrationReq{
		InstrumentNo: "JL-002", DeviceID: d.ID, CalibrationCycleMonths: 12,
	}, "admin")
	if err != nil {
		t.Fatalf("create calibration: %v", err)
	}
	if _, err := calSvc.RecordResult(c.ID, &dto.CalibrationResultReq{Result: constants.CalibrationResultQualified}, "admin"); err != nil {
		t.Fatalf("record result: %v", err)
	}

	m := createRepairOrder(t, svc, d.ID)
	startRepair(t, svc, m.ID)
	resp := completeRepair(t, svc, m.ID)
	if resp.DeviceStatus != constants.DeviceStatusInUse {
		t.Errorf("device status = %q, want in_use", resp.DeviceStatus)
	}
}

// 从未纳入计量的设备完成维修：直接恢复使用中。
func TestCompleteRepairWithoutCalibrationRestoresInUse(t *testing.T) {
	env := newTestServiceEnv(t)
	svc := newMaintenanceLinkSvc(t, env)
	d := newRepairDevice(t, env)
	m := createRepairOrder(t, svc, d.ID)
	startRepair(t, svc, m.ID)
	resp := completeRepair(t, svc, m.ID)
	if resp.DeviceStatus != constants.DeviceStatusInUse {
		t.Errorf("device status = %q, want in_use", resp.DeviceStatus)
	}
	if got := deviceStatus(t, env, d.ID); got != constants.DeviceStatusInUse {
		t.Errorf("persisted status = %q", got)
	}
}

// 最新一次计量为合格（此前曾不合格）时，以最新结果为准恢复使用中。
func TestCompleteRepairUsesLatestCalibrationResult(t *testing.T) {
	env := newTestServiceEnv(t)
	svc := newMaintenanceLinkSvc(t, env)
	calSvc := NewCalibrationService(repository.NewCalibrationRepository(env.db), repository.NewDeviceRepository(env.db), env.audit, env.logger)
	d := newRepairDevice(t, env)

	c, err := calSvc.Create(&dto.CreateCalibrationReq{
		InstrumentNo: "JL-003", DeviceID: d.ID, CalibrationCycleMonths: 12,
	}, "admin")
	if err != nil {
		t.Fatalf("create calibration: %v", err)
	}
	if _, err := calSvc.RecordResult(c.ID, &dto.CalibrationResultReq{Result: constants.CalibrationResultUnqualified}, "admin"); err != nil {
		t.Fatalf("record unqualified: %v", err)
	}
	if _, err := calSvc.RecordResult(c.ID, &dto.CalibrationResultReq{Result: constants.CalibrationResultQualified}, "admin"); err != nil {
		t.Fatalf("record qualified: %v", err)
	}

	m := createRepairOrder(t, svc, d.ID)
	startRepair(t, svc, m.ID)
	resp := completeRepair(t, svc, m.ID)
	if resp.DeviceStatus != constants.DeviceStatusInUse {
		t.Errorf("device status = %q, want in_use based on latest qualified result", resp.DeviceStatus)
	}
}

// 保养类（非维修）工单不与设备可用性联动。
func TestMaintenancePlanOrdersDoNotBlockOrChangeDevice(t *testing.T) {
	env := newTestServiceEnv(t)
	svc := newMaintenanceLinkSvc(t, env)
	d := newRepairDevice(t, env)

	m, err := svc.Create(&dto.CreateMaintenanceReq{DeviceID: d.ID, Type: constants.MaintenanceTypeDaily}, "admin")
	if err != nil {
		t.Fatalf("create daily plan: %v", err)
	}
	resp, err := svc.Start(m.ID, &dto.StartMaintenanceReq{Engineer: "王工"}, "admin")
	if err != nil {
		t.Fatalf("start daily: %v", err)
	}
	if resp.DeviceStatus != constants.DeviceStatusInUse {
		t.Errorf("device status = %q, daily plan must not change device status", resp.DeviceStatus)
	}
	// 维修期间可创建第二个同类保养工单并不受维修互斥影响：先开一个维修。
	r := createRepairOrder(t, svc, d.ID)
	startRepair(t, svc, r.ID)

	m2, err := svc.Create(&dto.CreateMaintenanceReq{DeviceID: d.ID, Type: constants.MaintenanceTypeWeekly}, "admin")
	if err != nil {
		t.Fatalf("create weekly plan during repair: %v", err)
	}
	if _, err := svc.Start(m2.ID, &dto.StartMaintenanceReq{Engineer: "王工"}, "admin"); err != nil {
		t.Errorf("maintenance plan start should not be blocked by active repair: %v", err)
	}
}

func countTransfers(env *testEnv, deviceID uint) int64 {
	var n int64
	env.db.Model(&model.TransferRequest{}).Where("device_id = ?", deviceID).Count(&n)
	return n
}

func countScraps(env *testEnv, deviceID uint) int64 {
	var n int64
	env.db.Model(&model.ScrapRequest{}).Where("device_id = ?", deviceID).Count(&n)
	return n
}

func requestTransferStatus(env *testEnv, id uint) string {
	var tr model.TransferRequest
	env.db.First(&tr, id)
	return tr.Status
}

func requestScrapStatus(env *testEnv, id uint) string {
	var sr model.ScrapRequest
	env.db.First(&sr, id)
	return sr.Status
}
