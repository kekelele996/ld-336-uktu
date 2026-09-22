package router

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/medasset/medasset/internal/config"
	"github.com/medasset/medasset/internal/handler"
	"github.com/medasset/medasset/internal/middleware"
	"github.com/medasset/medasset/internal/repository"
	"github.com/medasset/medasset/internal/service"
	"github.com/medasset/medasset/internal/util"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Deps 路由装配依赖。
type Deps struct {
	DB   *gorm.DB
	Cfg  *config.Config
	Log  *slog.Logger
	RDB  *redis.Client
}

// New 构建 Gin 引擎并注册全部路由。
func New(deps Deps) *gin.Engine {
	if deps.Cfg.RunMode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(middleware.RequestID())
	r.Use(middleware.CORS())
	r.Use(middleware.ErrorHandler(deps.Log))
	r.Use(middleware.RateLimit(deps.RDB, deps.Cfg.RateLimit))

	// 健康检查。
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, util.Resp{Code: 0, Message: "ok", Data: gin.H{"status": "healthy"}})
	})

	// 仓储层。
	userRepo := repository.NewUserRepository(deps.DB)
	deviceRepo := repository.NewDeviceRepository(deps.DB)
	purchaseRepo := repository.NewPurchaseRepository(deps.DB)
	maintenanceRepo := repository.NewMaintenanceRepository(deps.DB)
	calibrationRepo := repository.NewCalibrationRepository(deps.DB)
	transferRepo := repository.NewTransferRepository(deps.DB)
	scrapRepo := repository.NewScrapRepository(deps.DB)
	auditRepo := repository.NewAuditRepository(deps.DB)

	// 服务层。
	auditSvc := service.NewAuditService(auditRepo, deps.Log)
	userSvc := service.NewUserService(userRepo, auditSvc, deps.Cfg.JWTSecret, 24, deps.Log)
	deviceSvc := service.NewDeviceService(deviceRepo, auditSvc, deps.Log)
	purchaseSvc := service.NewPurchaseService(purchaseRepo, deviceRepo, auditSvc, deps.Log)
	maintenanceSvc := service.NewMaintenanceService(maintenanceRepo, deviceRepo, auditSvc, deps.Log)
	calibrationSvc := service.NewCalibrationService(calibrationRepo, deviceRepo, auditSvc, deps.Log)
	transferSvc := service.NewTransferService(transferRepo, deviceRepo, auditSvc, deps.Log)
	scrapSvc := service.NewScrapService(scrapRepo, deviceRepo, auditSvc, deps.Log)
	statsSvc := service.NewStatsService(deviceRepo, maintenanceRepo, calibrationRepo, purchaseRepo, auditSvc, deps.Log)

	// 处理器层。
	authHandler := handler.NewAuthHandler(userSvc)
	userHandler := handler.NewUserHandler(userSvc)
	deviceHandler := handler.NewDeviceHandler(deviceSvc)
	purchaseHandler := handler.NewPurchaseHandler(purchaseSvc)
	maintenanceHandler := handler.NewMaintenanceHandler(maintenanceSvc)
	calibrationHandler := handler.NewCalibrationHandler(calibrationSvc)
	transferHandler := handler.NewTransferHandler(transferSvc)
	scrapHandler := handler.NewScrapHandler(scrapSvc)
	auditHandler := handler.NewAuditHandler(auditSvc)
	statsHandler := handler.NewStatsHandler(statsSvc)

	auth := middleware.Auth(deps.Cfg)
	api := r.Group("/api/v1")
	{
		registerPublicAuthRoutes(api, authHandler)
		api.Use(auth)
		registerProtectedAuthRoutes(api, authHandler)
		registerUserRoutes(api, userHandler)
		registerDeviceRoutes(api, deviceHandler)
		registerPurchaseRoutes(api, purchaseHandler)
		registerMaintenanceRoutes(api, maintenanceHandler)
		registerCalibrationRoutes(api, calibrationHandler)
		registerTransferRoutes(api, transferHandler)
		registerScrapRoutes(api, scrapHandler)
		registerAuditRoutes(api, auditHandler)
		registerStatsRoutes(api, statsHandler)
	}
	return r
}
