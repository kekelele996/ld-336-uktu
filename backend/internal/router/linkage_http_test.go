package router

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/medasset/medasset/internal/config"
	"github.com/medasset/medasset/internal/model"
	"github.com/medasset/medasset/internal/repository"
	"github.com/medasset/medasset/internal/service"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func newHTTPEnv(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	if err := db.AutoMigrate(&model.User{}, &model.Device{}, &model.PurchaseRequest{},
		&model.MaintenanceRecord{}, &model.CalibrationRecord{}, &model.TransferRequest{},
		&model.ScrapRequest{}, &model.AuditLog{}); err != nil {
		t.Fatal(err)
	}
	auditSvc := service.NewAuditService(repository.NewAuditRepository(db), slog.Default())
	if err := service.NewUserService(repository.NewUserRepository(db), auditSvc, "test-secret", 24, slog.Default()).SeedAdmin(); err != nil {
		t.Fatal(err)
	}
	gin.SetMode(gin.TestMode)
	engine := New(Deps{DB: db, Cfg: &config.Config{ServerPort: "8080", RunMode: "test", JWTSecret: "test-secret", RateLimit: 10000}, Log: slog.Default(), RDB: nil})
	return engine, db
}

func loginToken(t *testing.T, engine *gin.Engine) string {
	t.Helper()
	body := `{"username":"admin","password":"admin123"}`
	w := doReq(engine, http.MethodPost, "/api/v1/auth/login", "", body)
	if w.Code != http.StatusOK {
		t.Fatalf("login status=%d body=%s", w.Code, w.Body.String())
	}
	var resp struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Data.Token == "" {
		t.Fatal("empty token")
	}
	return resp.Data.Token
}

func doReq(engine *gin.Engine, method, path, token, body string) *httptest.ResponseRecorder {
	var buf *bytes.Buffer
	if body != "" {
		buf = bytes.NewBufferString(body)
	} else {
		buf = bytes.NewBuffer(nil)
	}
	req := httptest.NewRequest(method, path, buf)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	return w
}

func mustCreateDeviceHTTP(t *testing.T, engine *gin.Engine, token, assetCode string) uint {
	t.Helper()
	body := `{"asset_code":"` + assetCode + `","name":"测试设备","category":"生命支持","department":"ICU"}`
	w := doReq(engine, http.MethodPost, "/api/v1/devices", token, body)
	if w.Code != http.StatusOK {
		t.Fatalf("create device status=%d body=%s", w.Code, w.Body.String())
	}
	var resp struct {
		Data struct {
			ID uint `json:"id"`
		} `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Data.ID == 0 {
		t.Fatalf("device id=0 body=%s", w.Body.String())
	}
	return resp.Data.ID
}

func createRepairHTTP(engine *gin.Engine, token string, deviceID uint) *httptest.ResponseRecorder {
	body, _ := json.Marshal(map[string]any{"device_id": deviceID, "type": "repair"})
	return doReq(engine, http.MethodPost, "/api/v1/maintenances", token, string(body))
}

func TestHTTPLinkageFlow(t *testing.T) {
	engine, _ := newHTTPEnv(t)
	token := loginToken(t, engine)
	devID := mustCreateDeviceHTTP(t, engine, token, "E2E-001")

	// 1. 创建第一张维修工单成功。
	w := createRepairHTTP(engine, token, devID)
	if w.Code != http.StatusOK {
		t.Fatalf("first repair: status=%d body=%s", w.Code, w.Body.String())
	}
	var m1 struct {
		Data struct {
			ID       uint   `json:"id"`
			RecordNo string `json:"record_no"`
			Status   string `json:"status"`
		} `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &m1)

	// 2. 同设备再报修 -> 409，message 含阻塞原因与原工单号。
	w = createRepairHTTP(engine, token, devID)
	if w.Code != http.StatusConflict {
		t.Fatalf("second repair status=%d body=%s", w.Code, w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte("维修工单")) || !bytes.Contains(w.Body.Bytes(), []byte(m1.Data.RecordNo)) {
		t.Fatalf("block reason not shown: %s", w.Body.String())
	}

	// 3. 开始维修成功，设备 -> under_maintenance。
	w = doReq(engine, http.MethodPost, "/api/v1/maintenances/"+itoa(m1.Data.ID)+"/start", token, `{"engineer":"张工"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("start status=%d body=%s", w.Code, w.Body.String())
	}

	// 4. 重复开始 -> 409。
	w = doReq(engine, http.MethodPost, "/api/v1/maintenances/"+itoa(m1.Data.ID)+"/start", token, `{"engineer":"张工"}`)
	if w.Code != http.StatusConflict {
		t.Fatalf("repeat start status=%d body=%s", w.Code, w.Body.String())
	}

	// 5. 维修期间发起调拨 -> 409，body 含"维修中"。
	tb, _ := json.Marshal(map[string]any{"device_id": devID, "to_department": "心内科"})
	w = doReq(engine, http.MethodPost, "/api/v1/transfers", token, string(tb))
	if w.Code != http.StatusConflict || !bytes.Contains(w.Body.Bytes(), []byte("维修中")) {
		t.Fatalf("transfer during repair status=%d body=%s", w.Code, w.Body.String())
	}

	// 6. 维修期间发起报废 -> 409。
	sb, _ := json.Marshal(map[string]any{"device_id": devID, "reason": "损坏无法修复"})
	w = doReq(engine, http.MethodPost, "/api/v1/scraps", token, string(sb))
	if w.Code != http.StatusConflict || !bytes.Contains(w.Body.Bytes(), []byte("维修中")) {
		t.Fatalf("scrap during repair status=%d body=%s", w.Code, w.Body.String())
	}

	// 7. 完成维修（无计量）-> 200，device_restored=true, device_status=in_use。
	w = doReq(engine, http.MethodPost, "/api/v1/maintenances/"+itoa(m1.Data.ID)+"/complete", token, `{"content":"更换主板","cost":500}`)
	if w.Code != http.StatusOK {
		t.Fatalf("complete status=%d body=%s", w.Code, w.Body.String())
	}
	var cr struct {
		Data struct {
			DeviceStatus   string `json:"device_status"`
			DeviceRestored bool   `json:"device_restored"`
			Notice         string `json:"notice"`
		} `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &cr)
	if !cr.Data.DeviceRestored || cr.Data.DeviceStatus != "in_use" || cr.Data.Notice != "" {
		t.Fatalf("complete response unexpected: %s", w.Body.String())
	}

	// 8. 刷新设备详情 -> in_use（刷新后一致）。
	w = doReq(engine, http.MethodGet, "/api/v1/devices/"+itoa(devID), token, "")
	if w.Code != http.StatusOK || !bytes.Contains(w.Body.Bytes(), []byte(`"status":"in_use"`)) {
		t.Fatalf("device after complete: status=%d body=%s", w.Code, w.Body.String())
	}

	// 9. 维修结束后调拨可正常发起。
	tb2, _ := json.Marshal(map[string]any{"device_id": devID, "to_department": "心内科"})
	w = doReq(engine, http.MethodPost, "/api/v1/transfers", token, string(tb2))
	if w.Code != http.StatusOK {
		t.Fatalf("transfer after repair status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestHTTPUnqualifiedCalibrationHoldsDisabled(t *testing.T) {
	engine, _ := newHTTPEnv(t)
	token := loginToken(t, engine)
	devID := mustCreateDeviceHTTP(t, engine, token, "E2E-002")

	// 建立计量台账并登记不合格。
	cb := `{"instrument_no":"JL-E2E-002","device_id":` + itoa(devID) + `,"calibration_cycle_months":12}`
	w := doReq(engine, http.MethodPost, "/api/v1/calibrations", token, cb)
	if w.Code != http.StatusOK {
		t.Fatalf("create calibration: %s", w.Body.String())
	}
	var cid struct {
		Data struct {
			ID uint `json:"id"`
		} `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &cid)
	w = doReq(engine, http.MethodPost, "/api/v1/calibrations/"+itoa(cid.Data.ID)+"/result", token, `{"result":"unqualified"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("record unqualified: %s", w.Body.String())
	}
	// 不合格 -> 设备禁用。
	w = doReq(engine, http.MethodGet, "/api/v1/devices/"+itoa(devID), token, "")
	if !bytes.Contains(w.Body.Bytes(), []byte(`"status":"disabled"`)) {
		t.Fatalf("device should be disabled: %s", w.Body.String())
	}

	// 报修 -> 开始 -> 完成。
	w = createRepairHTTP(engine, token, devID)
	var m struct {
		Data struct{ ID uint } `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &m)
	if w := doReq(engine, http.MethodPost, "/api/v1/maintenances/"+itoa(m.Data.ID)+"/start", token, `{"engineer":"张工"}`); w.Code != http.StatusOK {
		t.Fatalf("start: %s", w.Body.String())
	}
	w = doReq(engine, http.MethodPost, "/api/v1/maintenances/"+itoa(m.Data.ID)+"/complete", token, `{"content":"修复"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("complete: %s", w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte(`"device_restored":false`)) ||
		!bytes.Contains(w.Body.Bytes(), []byte(`"device_status":"disabled"`)) ||
		!bytes.Contains(w.Body.Bytes(), []byte("计量")) {
		t.Fatalf("unqualified complete response missing hold+notice: %s", w.Body.String())
	}
	// 刷新设备仍为禁用。
	w = doReq(engine, http.MethodGet, "/api/v1/devices/"+itoa(devID), token, "")
	if !bytes.Contains(w.Body.Bytes(), []byte(`"status":"disabled"`)) {
		t.Fatalf("device should stay disabled after repair: %s", w.Body.String())
	}
}

func itoa(u uint) string {
	if u == 0 {
		return "0"
	}
	b := make([]byte, 0, 8)
	for u > 0 {
		b = append([]byte{byte('0' + u%10)}, b...)
		u /= 10
	}
	return string(b)
}
