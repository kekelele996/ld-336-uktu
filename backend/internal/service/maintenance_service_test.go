package service

import (
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/medasset/medasset/internal/constants"
	"github.com/medasset/medasset/internal/dto"
	"github.com/medasset/medasset/internal/model"
	"github.com/medasset/medasset/internal/repository"
	"github.com/medasset/medasset/internal/util"
)

// TestMaintenanceRepairMutex 同一设备只能有一个待处理/处理中的维修工单进入处理中，
// 重复开始 / 另一张待处理工单开始都被整次拒绝，原工单与设备保持原样。
func TestMaintenanceRepairMutex(t *testing.T) {
	env := newTestServiceEnv(t)
	deviceRepo := repository.NewDeviceRepository(env.db)
	maintRepo := repository.NewMaintenanceRepository(env.db)
	calibRepo := repository.NewCalibrationRepository(env.db)
	svc := NewMaintenanceService(maintRepo, deviceRepo, calibRepo, env.audit, env.logger)

	dev, err := NewDeviceService(deviceRepo, env.audit, env.logger).
		Create(&dto.CreateDeviceReq{AssetCode: "MA-200", Name: "除颤仪", Category: "生命支持", Department: "急诊科"}, "admin")
	if err != nil {
		t.Fatalf("create device: %v", err)
	}

	// 第一张待处理维修工单创建成功。
	m1 := mustCreateRepair(t, svc, dev.ID)

	// 同一设备再次报修：创建即被整次拒绝（每台设备至多一张未闭环维修单）。
	if _, err := svc.Create(&dto.CreateMaintenanceReq{DeviceID: dev.ID, Type: constants.MaintenanceTypeRepair}, "reporter2"); !isConflict(err) {
		t.Fatalf("second repair create expected conflict, got %v", err)
	}
	if appErr, ok := err.(*util.AppError); ok && !contains(appErr.Message, m1.RecordNo) {
		t.Fatalf("create block reason should name blocker %s, got %q", m1.RecordNo, appErr.Message)
	}

	// m1 开始成功，设备进入维修中。
	started, err := svc.Start(m1.ID, &dto.StartMaintenanceReq{Engineer: "张工"}, "admin")
	if err != nil {
		t.Fatalf("start m1: %v", err)
	}
	if started.Status != constants.MaintenanceStatusInProgress {
		t.Fatalf("m1 status = %s", started.Status)
	}
	if got := mustDeviceStatus(t, deviceRepo, dev.ID); got != constants.DeviceStatusUnderMaintenance {
		t.Fatalf("device status = %s, want under_maintenance", got)
	}

	// 重复开始 m1：拒绝，状态不变（重复提交只生效一次）。
	if _, err := svc.Start(m1.ID, &dto.StartMaintenanceReq{Engineer: "张工"}, "admin"); !isConflict(err) {
		t.Fatalf("repeat start expected conflict, got %v", err)
	}
	if got := mustMaintenanceStatus(t, maintRepo, m1.ID); got != constants.MaintenanceStatusInProgress {
		t.Fatalf("m1 changed after repeat start: %s", got)
	}

	// 防御性场景：绕过创建校验直接落库另一张待处理维修单，开始仍被拒绝（阻塞原因含工单号）。
	m2 := &model.MaintenanceRecord{
		RecordNo: "MT-BYPASS", DeviceID: dev.ID, DeviceName: dev.Name,
		Type: constants.MaintenanceTypeRepair, Status: constants.MaintenanceStatusPending,
	}
	if err := env.db.Create(m2).Error; err != nil {
		t.Fatalf("bypass create: %v", err)
	}
	if _, err := svc.Start(m2.ID, &dto.StartMaintenanceReq{Engineer: "李工"}, "admin"); !isConflict(err) {
		t.Fatalf("start m2 expected conflict, got %v", err)
	}
	if appErr, ok := err.(*util.AppError); ok {
		if got := appErr.Message; !contains(got, m1.RecordNo) || !contains(got, constants.MsgRepairBlocked) {
			t.Fatalf("block reason should name blocker order, got %q", got)
		}
	}
	// 原工单、设备保持原样。
	if got := mustMaintenanceStatus(t, maintRepo, m2.ID); got != constants.MaintenanceStatusPending {
		t.Fatalf("m2 status = %s, want pending (unchanged)", got)
	}
	if got := mustDeviceStatus(t, deviceRepo, dev.ID); got != constants.DeviceStatusUnderMaintenance {
		t.Fatalf("device status = %s, want under_maintenance (unchanged)", got)
	}
}

// TestMaintenanceCompleteRestoresWithoutCalibration 从未纳入计量的设备维修完成后恢复使用中。
func TestMaintenanceCompleteRestoresWithoutCalibration(t *testing.T) {
	env := newTestServiceEnv(t)
	deviceRepo := repository.NewDeviceRepository(env.db)
	maintRepo := repository.NewMaintenanceRepository(env.db)
	calibRepo := repository.NewCalibrationRepository(env.db)
	svc := NewMaintenanceService(maintRepo, deviceRepo, calibRepo, env.audit, env.logger)

	dev, _ := NewDeviceService(deviceRepo, env.audit, env.logger).
		Create(&dto.CreateDeviceReq{AssetCode: "MA-201", Name: "输液泵", Category: "生命支持", Department: "ICU"}, "admin")
	m := mustCreateRepair(t, svc, dev.ID)
	if _, err := svc.Start(m.ID, &dto.StartMaintenanceReq{Engineer: "张工"}, "admin"); err != nil {
		t.Fatalf("start: %v", err)
	}

	resp, err := svc.Complete(m.ID, &dto.CompleteMaintenanceReq{Content: "更换电池", Cost: 100}, "admin")
	if err != nil {
		t.Fatalf("complete: %v", err)
	}
	if !resp.DeviceRestored || resp.DeviceStatus != constants.DeviceStatusInUse {
		t.Fatalf("device should restore to in_use, got restored=%v status=%s", resp.DeviceRestored, resp.DeviceStatus)
	}
	if resp.Notice != "" {
		t.Fatalf("unexpected notice: %s", resp.Notice)
	}
	if got := mustDeviceStatus(t, deviceRepo, dev.ID); got != constants.DeviceStatusInUse {
		t.Fatalf("device status = %s", got)
	}

	// 重复完成被拒绝。
	if _, err := svc.Complete(m.ID, &dto.CompleteMaintenanceReq{Content: "x"}, "admin"); !isConflict(err) {
		t.Fatalf("repeat complete expected conflict, got %v", err)
	}
}

// TestMaintenanceCompleteQualifiedCalibration 最新计量合格 -> 恢复使用中。
func TestMaintenanceCompleteQualifiedCalibration(t *testing.T) {
	env := newTestServiceEnv(t)
	deviceRepo := repository.NewDeviceRepository(env.db)
	maintRepo := repository.NewMaintenanceRepository(env.db)
	calibRepo := repository.NewCalibrationRepository(env.db)
	svc := NewMaintenanceService(maintRepo, deviceRepo, calibRepo, env.audit, env.logger)
	calibSvc := NewCalibrationService(calibRepo, deviceRepo, env.audit, env.logger)

	dev, _ := NewDeviceService(deviceRepo, env.audit, env.logger).
		Create(&dto.CreateDeviceReq{AssetCode: "MA-202", Name: "监护仪", Category: "生命支持", Department: "ICU"}, "admin")
	c, err := calibSvc.Create(&dto.CreateCalibrationReq{InstrumentNo: "JL-202", DeviceID: dev.ID, CalibrationCycleMonths: 12}, "admin")
	if err != nil {
		t.Fatalf("create calibration: %v", err)
	}
	if _, err := calibSvc.RecordResult(c.ID, &dto.CalibrationResultReq{Result: constants.CalibrationResultQualified}, "admin"); err != nil {
		t.Fatalf("record qualified: %v", err)
	}

	m := mustCreateRepair(t, svc, dev.ID)
	if _, err := svc.Start(m.ID, &dto.StartMaintenanceReq{Engineer: "张工"}, "admin"); err != nil {
		t.Fatalf("start: %v", err)
	}
	resp, err := svc.Complete(m.ID, &dto.CompleteMaintenanceReq{Content: "校准完成"}, "admin")
	if err != nil {
		t.Fatalf("complete: %v", err)
	}
	if !resp.DeviceRestored {
		t.Fatalf("qualified calibration should restore device, notice=%s", resp.Notice)
	}
}

// TestMaintenanceCompleteUnqualifiedHoldsDisabled 最新计量不合格 -> 设备保持禁用并提示未通过计量。
func TestMaintenanceCompleteUnqualifiedHoldsDisabled(t *testing.T) {
	env := newTestServiceEnv(t)
	deviceRepo := repository.NewDeviceRepository(env.db)
	maintRepo := repository.NewMaintenanceRepository(env.db)
	calibRepo := repository.NewCalibrationRepository(env.db)
	svc := NewMaintenanceService(maintRepo, deviceRepo, calibRepo, env.audit, env.logger)
	calibSvc := NewCalibrationService(calibRepo, deviceRepo, env.audit, env.logger)

	dev, _ := NewDeviceService(deviceRepo, env.audit, env.logger).
		Create(&dto.CreateDeviceReq{AssetCode: "MA-203", Name: "分析仪", Category: "检验设备", Department: "检验科"}, "admin")
	c, err := calibSvc.Create(&dto.CreateCalibrationReq{InstrumentNo: "JL-203", DeviceID: dev.ID, CalibrationCycleMonths: 12}, "admin")
	if err != nil {
		t.Fatalf("create calibration: %v", err)
	}
	// 计量不合格 -> 设备禁用。
	if _, err := calibSvc.RecordResult(c.ID, &dto.CalibrationResultReq{Result: constants.CalibrationResultUnqualified}, "admin"); err != nil {
		t.Fatalf("record unqualified: %v", err)
	}

	m := mustCreateRepair(t, svc, dev.ID)
	if _, err := svc.Start(m.ID, &dto.StartMaintenanceReq{Engineer: "张工"}, "admin"); err != nil {
		t.Fatalf("start: %v", err)
	}
	if got := mustDeviceStatus(t, deviceRepo, dev.ID); got != constants.DeviceStatusUnderMaintenance {
		t.Fatalf("device status = %s, want under_maintenance", got)
	}

	resp, err := svc.Complete(m.ID, &dto.CompleteMaintenanceReq{Content: "维修完成"}, "admin")
	if err != nil {
		t.Fatalf("complete: %v", err)
	}
	if resp.DeviceRestored {
		t.Fatal("unqualified calibration must NOT restore device")
	}
	if resp.DeviceStatus != constants.DeviceStatusDisabled {
		t.Fatalf("device status = %s, want disabled", resp.DeviceStatus)
	}
	if !contains(resp.Notice, "计量") {
		t.Fatalf("notice should mention calibration, got %q", resp.Notice)
	}
	if got := mustDeviceStatus(t, deviceRepo, dev.ID); got != constants.DeviceStatusDisabled {
		t.Fatalf("device status = %s, want disabled after complete", got)
	}
	if got := mustMaintenanceStatus(t, maintRepo, m.ID); got != constants.MaintenanceStatusCompleted {
		t.Fatalf("maintenance status = %s, want completed", got)
	}
}

// TestTransferScrapBlockedDuringMaintenance 维修期间调拨/报废申请一并拒绝，原工单、设备与申请保持原样。
func TestTransferScrapBlockedDuringMaintenance(t *testing.T) {
	env := newTestServiceEnv(t)
	deviceRepo := repository.NewDeviceRepository(env.db)
	maintRepo := repository.NewMaintenanceRepository(env.db)
	calibRepo := repository.NewCalibrationRepository(env.db)
	maintSvc := NewMaintenanceService(maintRepo, deviceRepo, calibRepo, env.audit, env.logger)
	transferSvc := NewTransferService(repository.NewTransferRepository(env.db), deviceRepo, env.audit, env.logger)
	scrapSvc := NewScrapService(repository.NewScrapRepository(env.db), deviceRepo, env.audit, env.logger)

	dev, _ := NewDeviceService(deviceRepo, env.audit, env.logger).
		Create(&dto.CreateDeviceReq{AssetCode: "MA-204", Name: "CT机", Category: "影像设备", Department: "放射科"}, "admin")

	// 维修前可正常发起调拨/报废。
	if _, err := transferSvc.Create(&dto.CreateTransferReq{DeviceID: dev.ID, ToDepartment: "ICU", ToPerson: "王五"}, "u1"); err != nil {
		t.Fatalf("transfer before maintenance: %v", err)
	}
	if _, err := scrapSvc.Create(&dto.CreateScrapReq{DeviceID: dev.ID, Reason: "老化"}, "u1"); err != nil {
		t.Fatalf("scrap before maintenance: %v", err)
	}

	m := mustCreateRepair(t, maintSvc, dev.ID)
	if _, err := maintSvc.Start(m.ID, &dto.StartMaintenanceReq{Engineer: "张工"}, "admin"); err != nil {
		t.Fatalf("start: %v", err)
	}

	// 维修期间新发起调拨/报废被拒绝。
	if _, err := transferSvc.Create(&dto.CreateTransferReq{DeviceID: dev.ID, ToDepartment: "心内科"}, "u1"); !isConflict(err) {
		t.Fatalf("transfer create during maintenance expected conflict, got %v", err)
	}
	if _, err := scrapSvc.Create(&dto.CreateScrapReq{DeviceID: dev.ID, Reason: "损坏"}, "u1"); err == nil {
		t.Fatal("scrap create during maintenance expected conflict, got nil")
	} else if appErr, ok := err.(*util.AppError); !ok || !contains(appErr.Message, constants.MsgDeviceUnderMaintenance) {
		t.Fatalf("block reason missing, got %v", err)
	}

	// 既有待审批调拨/报废在维修期间审批也被拒绝，申请保持 pending、设备保持维修中。
	if err := env.db.Where("device_id = ? AND status = ?", dev.ID, constants.TransferStatusPending).
		First(&model.TransferRequest{}).Error; err != nil {
		t.Fatalf("pending transfer missing: %v", err)
	}
	var pendingTransfer model.TransferRequest
	if err := env.db.Where("device_id = ? AND status = ?", dev.ID, constants.TransferStatusPending).First(&pendingTransfer).Error; err != nil {
		t.Fatalf("load pending transfer: %v", err)
	}
	if _, err := transferSvc.Approve(pendingTransfer.ID, &dto.TransferApproveReq{Comment: "ok"}, "dean"); !isConflict(err) {
		t.Fatalf("transfer approve during maintenance expected conflict, got %v", err)
	}
	var pendingScrap model.ScrapRequest
	if err := env.db.Where("device_id = ? AND status = ?", dev.ID, constants.ScrapStatusPending).First(&pendingScrap).Error; err != nil {
		t.Fatalf("load pending scrap: %v", err)
	}
	if _, err := scrapSvc.Approve(pendingScrap.ID, &dto.ScrapApproveReq{Comment: "ok"}, "dean"); !isConflict(err) {
		t.Fatalf("scrap approve during maintenance expected conflict, got %v", err)
	}

	// 原工单、设备保持原样。
	if got := mustMaintenanceStatus(t, maintRepo, m.ID); got != constants.MaintenanceStatusInProgress {
		t.Fatalf("repair status = %s, want in_progress", got)
	}
	if got := mustDeviceStatus(t, deviceRepo, dev.ID); got != constants.DeviceStatusUnderMaintenance {
		t.Fatalf("device status = %s, want under_maintenance", got)
	}
	var nTransfer int64
	if err := env.db.Model(&model.TransferRequest{}).
		Where("device_id = ? AND status = ?", dev.ID, constants.TransferStatusPending).
		Count(&nTransfer).Error; err != nil {
		t.Fatalf("count transfers: %v", err)
	}
	if nTransfer != 1 {
		t.Fatalf("pending transfers = %d, want 1 (unchanged)", nTransfer)
	}
	var nScrap int64
	if err := env.db.Model(&model.ScrapRequest{}).
		Where("device_id = ? AND status = ?", dev.ID, constants.ScrapStatusPending).
		Count(&nScrap).Error; err != nil {
		t.Fatalf("count scraps: %v", err)
	}
	if nScrap != 1 {
		t.Fatalf("pending scraps = %d, want 1 (unchanged)", nScrap)
	}

	// 维修完成（无计量）恢复使用后，审批可正常通过。
	if _, err := maintSvc.Complete(m.ID, &dto.CompleteMaintenanceReq{Content: "修复"}, "admin"); err != nil {
		t.Fatalf("complete: %v", err)
	}
	if _, err := transferSvc.Approve(pendingTransfer.ID, &dto.TransferApproveReq{Comment: "ok"}, "dean"); err != nil {
		t.Fatalf("transfer approve after maintenance: %v", err)
	}
}

// TestConcurrentStartOnlyOneSucceeds 并发开始同一张维修工单只能生效一次。
func TestConcurrentStartOnlyOneSucceeds(t *testing.T) {
	env := newTestServiceEnv(t)
	deviceRepo := repository.NewDeviceRepository(env.db)
	maintRepo := repository.NewMaintenanceRepository(env.db)
	calibRepo := repository.NewCalibrationRepository(env.db)
	svc := NewMaintenanceService(maintRepo, deviceRepo, calibRepo, env.audit, env.logger)

	dev, _ := NewDeviceService(deviceRepo, env.audit, env.logger).
		Create(&dto.CreateDeviceReq{AssetCode: "MA-300", Name: "呼吸机", Category: "生命支持", Department: "ICU"}, "admin")
	m := mustCreateRepair(t, svc, dev.ID)

	const n = 20
	var success int64
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			if _, err := svc.Start(m.ID, &dto.StartMaintenanceReq{Engineer: "并发工程师"}, "admin"); err == nil {
				atomic.AddInt64(&success, 1)
			}
		}()
	}
	close(start)
	wg.Wait()

	if success != 1 {
		t.Fatalf("concurrent start success = %d, want exactly 1", success)
	}
	if got := mustMaintenanceStatus(t, maintRepo, m.ID); got != constants.MaintenanceStatusInProgress {
		t.Fatalf("status = %s, want in_progress", got)
	}
	if got := mustDeviceStatus(t, deviceRepo, dev.ID); got != constants.DeviceStatusUnderMaintenance {
		t.Fatalf("device = %s, want under_maintenance", got)
	}
}

// TestConcurrentCompleteOnlyOneSucceeds 并发完成同一张维修工单只能生效一次。
func TestConcurrentCompleteOnlyOneSucceeds(t *testing.T) {
	env := newTestServiceEnv(t)
	deviceRepo := repository.NewDeviceRepository(env.db)
	maintRepo := repository.NewMaintenanceRepository(env.db)
	calibRepo := repository.NewCalibrationRepository(env.db)
	svc := NewMaintenanceService(maintRepo, deviceRepo, calibRepo, env.audit, env.logger)

	dev, _ := NewDeviceService(deviceRepo, env.audit, env.logger).
		Create(&dto.CreateDeviceReq{AssetCode: "MA-301", Name: "注射泵", Category: "生命支持", Department: "ICU"}, "admin")
	m := mustCreateRepair(t, svc, dev.ID)
	if _, err := svc.Start(m.ID, &dto.StartMaintenanceReq{Engineer: "张工"}, "admin"); err != nil {
		t.Fatalf("start: %v", err)
	}

	const n = 20
	var success int64
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			if _, err := svc.Complete(m.ID, &dto.CompleteMaintenanceReq{Content: "并发完成"}, "admin"); err == nil {
				atomic.AddInt64(&success, 1)
			}
		}()
	}
	close(start)
	wg.Wait()

	if success != 1 {
		t.Fatalf("concurrent complete success = %d, want exactly 1", success)
	}
	if got := mustMaintenanceStatus(t, maintRepo, m.ID); got != constants.MaintenanceStatusCompleted {
		t.Fatalf("status = %s, want completed", got)
	}
	if got := mustDeviceStatus(t, deviceRepo, dev.ID); got != constants.DeviceStatusInUse {
		t.Fatalf("device = %s, want in_use", got)
	}
}

// TestConcurrentTransferScrapBlockedDuringMaintenance 维修期间并发发起调拨/报废全部被拒绝，不产生任何申请。
func TestConcurrentTransferScrapBlockedDuringMaintenance(t *testing.T) {
	env := newTestServiceEnv(t)
	deviceRepo := repository.NewDeviceRepository(env.db)
	maintRepo := repository.NewMaintenanceRepository(env.db)
	calibRepo := repository.NewCalibrationRepository(env.db)
	maintSvc := NewMaintenanceService(maintRepo, deviceRepo, calibRepo, env.audit, env.logger)
	transferSvc := NewTransferService(repository.NewTransferRepository(env.db), deviceRepo, env.audit, env.logger)
	scrapSvc := NewScrapService(repository.NewScrapRepository(env.db), deviceRepo, env.audit, env.logger)

	dev, _ := NewDeviceService(deviceRepo, env.audit, env.logger).
		Create(&dto.CreateDeviceReq{AssetCode: "MA-302", Name: "彩超", Category: "影像设备", Department: "超声科"}, "admin")
	m := mustCreateRepair(t, maintSvc, dev.ID)
	if _, err := maintSvc.Start(m.ID, &dto.StartMaintenanceReq{Engineer: "张工"}, "admin"); err != nil {
		t.Fatalf("start: %v", err)
	}

	const n = 15
	var wg sync.WaitGroup
	var transferOK, scrapOK int64
	start := make(chan struct{})
	for i := 0; i < n; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			<-start
			if _, err := transferSvc.Create(&dto.CreateTransferReq{DeviceID: dev.ID, ToDepartment: "心内科"}, "u"); err == nil {
				atomic.AddInt64(&transferOK, 1)
			}
		}()
		go func() {
			defer wg.Done()
			<-start
			if _, err := scrapSvc.Create(&dto.CreateScrapReq{DeviceID: dev.ID, Reason: "故障"}, "u"); err == nil {
				atomic.AddInt64(&scrapOK, 1)
			}
		}()
	}
	close(start)
	wg.Wait()

	if transferOK != 0 || scrapOK != 0 {
		t.Fatalf("during maintenance transferOK=%d scrapOK=%d, want 0/0", transferOK, scrapOK)
	}
	var nTransfer, nScrap int64
	if err := env.db.Model(&model.TransferRequest{}).Where("device_id = ?", dev.ID).Count(&nTransfer).Error; err != nil {
		t.Fatal(err)
	}
	if err := env.db.Model(&model.ScrapRequest{}).Where("device_id = ?", dev.ID).Count(&nScrap).Error; err != nil {
		t.Fatal(err)
	}
	if nTransfer != 0 || nScrap != 0 {
		t.Fatalf("created during maintenance: transfers=%d scraps=%d, want 0/0", nTransfer, nScrap)
	}
	if got := mustDeviceStatus(t, deviceRepo, dev.ID); got != constants.DeviceStatusUnderMaintenance {
		t.Fatalf("device = %s, want under_maintenance", got)
	}
	if got := mustMaintenanceStatus(t, maintRepo, m.ID); got != constants.MaintenanceStatusInProgress {
		t.Fatalf("maintenance = %s, want in_progress", got)
	}
}

func mustCreateRepair(t *testing.T, svc *MaintenanceService, deviceID uint) *model.MaintenanceRecord {
	t.Helper()
	m, err := svc.Create(&dto.CreateMaintenanceReq{
		DeviceID: deviceID,
		Type:     constants.MaintenanceTypeRepair,
	}, "reporter")
	if err != nil {
		t.Fatalf("create repair: %v", err)
	}
	return m
}

func mustDeviceStatus(t *testing.T, repo *repository.DeviceRepository, id uint) string {
	t.Helper()
	d, err := repo.FindByID(id)
	if err != nil {
		t.Fatalf("find device: %v", err)
	}
	return d.Status
}

func mustMaintenanceStatus(t *testing.T, repo *repository.MaintenanceRepository, id uint) string {
	t.Helper()
	m, err := repo.FindByID(id)
	if err != nil {
		t.Fatalf("find maintenance: %v", err)
	}
	return m.Status
}

func isConflict(err error) bool {
	appErr, ok := err.(*util.AppError)
	return ok && appErr.Code == http.StatusConflict
}

func contains(s, sub string) bool {
	return strings.Contains(s, sub)
}
