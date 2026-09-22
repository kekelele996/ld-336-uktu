# 医疗器械资产管理平台（medasset）

为医院设备科提供医疗器械从采购、验收、使用、维护、计量、调拨到报废的全生命周期数字化管理平台，确保设备台账清晰、维护及时、合规可追溯。

## 快速启动（Docker Compose，首选）

```bash
docker compose up -d --build
```

启动后访问：

| 入口 | 地址 |
| --- | --- |
| 前端 | http://localhost:18936 |
| 后端 API | http://localhost:19936/api/v1 |
| 健康检查 | http://localhost:19936/healthz |
| 默认管理员 | admin / admin123 |

## 项目主要功能

1. **设备台账管理**：全院医疗器械电子台账，支持按科室、设备类型、状态、关键字多维度检索；登记设备自动生成分发条码。
2. **采购与验收流程**：科室申请 → 设备科审核 → 院长审批 → 到货登记 → 验收登记（配件清单、合格证/注册证）→ 正式入台账并生成分发条码。
3. **维护与保养管理**：日检/周检/月检/年检保养计划自动生成与到期提醒，工程师执行并记录保养内容、更换配件、工时与费用；故障扫码快速报修与维修过程记录。
4. **计量与质控管理**：计量台账（器具编号、周期、上下次计量日期），到期预警清单，计量结果登记（不合格自动标记设备禁用）。
5. **设备调拨与报废**：科室间调拨申请审批后自动更新设备科室与责任人；报废审批通过后设备状态变更为"已报废"并归档。
6. **资产统计与合规报表**：设备总数、资产总值、科室/品牌/类型分布、维修成本、计量到期预警、待处理采购等总览数据，满足监管数据报送要求。
7. **横切能力**：JWT 认证 + RBAC 权限、操作审计日志、全局错误处理与请求追踪（request id）、Redis 限流。

## 技术栈

| 层 | 技术栈 |
| --- | --- |
| 前端 | Angular 17 + TypeScript + Angular Material |
| 后端 | Go 1.22 + Gin + GORM |
| 数据库 | MySQL 8.0 |
| 缓存/限流 | Redis 7 |
| 认证 | JWT + RBAC |
| 日志 | log/slog |
| 参数校验 | github.com/go-playground/validator/v10 |
| 接口文档 | README 完整 API 清单 |

## 目录结构

```
ld-336/
├── docker-compose.yml
├── .env / .env.example
├── README.md
├── database/init.sql
├── backend/
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── config/config.go
│   │   ├── database/database.go
│   │   ├── model/           # 每个实体一个文件
│   │   ├── dto/             # 每个实体一个 DTO 文件
│   │   ├── repository/      # 每个实体一个 repository 文件
│   │   ├── service/         # 每个实体一个 service 文件
│   │   ├── handler/         # 每个实体一个 handler 文件
│   │   ├── router/          # 每个实体一个路由注册文件
│   │   ├── middleware/      # auth / rbac / request_id / error_handler / audit / rate_limit / cors
│   │   ├── constants/       # error_codes / roles / status / log_templates / messages
│   │   └── util/            # logger / jwt / app_error / response / formatters / pagination
│   ├── pkg/pointerx/
│   ├── migrations/001_init.sql
│   ├── Dockerfile
│   ├── go.mod / go.sum
├── frontend/
│   ├── src/api/             # 每个实体一个 API 文件
│   ├── src/components/      # status-badge / confirm-dialog / empty-state / page-header
│   ├── src/pages/           # login / dashboard / devices / purchases / maintenance / calibrations / transfers / scraps / audits
│   ├── src/stores/          # 按实体拆分 store（signal 驱动）
│   ├── src/hooks/           # use-auth / use-pagination
│   ├── src/utils/           # request / format
│   ├── src/constants/       # enums（与后端对应）
│   ├── Dockerfile
│   └── nginx.conf
└── output/execution.md
```

**严禁合并职责到单一文件**：每个实体严格拆分为 model / dto / repository / service / handler / router / constants 独立文件；前端按模块拆分 api / stores / components / pages，禁止把所有页面写在一个文件。

## 环境变量说明

| 变量 | 默认值 | 说明 |
| --- | --- | --- |
| COMPOSE_PROJECT_NAME | medasset | Compose 项目名（容器名/卷名前缀） |
| FRONTEND_PORT | 18936 | 前端 Nginx 对外端口 → 容器 80 |
| BACKEND_PORT | 19936 | 后端 Gin 对外端口 → 容器 8080 |
| DB_PORT | 44017 | MySQL 对外端口 → 容器 3306 |
| REDIS_PORT | 46317 | Redis 对外端口 → 容器 6379 |
| DB_NAME | medasset_db | 数据库名 |
| DB_USER | medasset_user | 数据库用户 |
| DB_PASSWORD | medasset_pwd | 数据库密码 |
| DB_ROOT_PASSWORD | medasset_root_pwd | MySQL root 密码 |
| JWT_SECRET | change_me_to_a_long_random_string | JWT 签名密钥 |
| JWT_EXPIRE_HOURS | 24 | Token 有效期（小时） |
| RATE_LIMIT_PER_SECOND | 200 | 每 IP 每秒限流阈值 |

## 本地开发

后端：

```bash
cd backend && go mod tidy && go run ./cmd/server
# 构建：go build ./...
# 测试：go test ./...
```

前端：

```bash
cd frontend && npm install && npm run build
```

## API 清单（统一前缀 /api/v1，响应统一 { "code": 0, "message": "ok", "data": ... }）

| 方法 | 路径 | 说明 | 权限 |
| --- | --- | --- | --- |
| GET | /healthz | 健康检查 | 公开 |
| POST | /auth/register | 注册 | 公开 |
| POST | /auth/login | 登录（返回 JWT） | 公开 |
| GET | /auth/me | 当前用户 | 登录 |
| GET | /users | 用户列表 | SUPER_ADMIN |
| PUT | /users/:id | 更新用户 | SUPER_ADMIN |
| DELETE | /users/:id | 删除用户 | SUPER_ADMIN |
| GET | /devices | 设备台账列表（科室/类型/状态/关键字） | 登录 |
| GET | /devices/:id | 设备详情 | 登录 |
| POST | /devices | 登记设备 | 登录 |
| PUT | /devices/:id | 更新设备 | 登录 |
| POST | /devices/:id/disable | 禁用设备 | 登录 |
| POST | /devices/:id/enable | 启用设备 | 登录 |
| GET | /purchases | 采购申请列表 | 登录 |
| GET | /purchases/:id | 采购详情 | 登录 |
| POST | /purchases | 提交采购申请 | 登录 |
| POST | /purchases/:id/device-admin-approve | 设备科审核通过 | DEVICE_ADMIN |
| POST | /purchases/:id/device-admin-reject | 设备科驳回 | DEVICE_ADMIN |
| POST | /purchases/:id/dean-approve | 院长审批通过 | DEAN |
| POST | /purchases/:id/dean-reject | 院长驳回 | DEAN |
| POST | /purchases/:id/deliver | 到货登记 | DEVICE_ADMIN |
| POST | /purchases/:id/accept | 验收登记并入台账 | DEVICE_ADMIN |
| GET | /maintenances | 保养/维修记录列表 | 登录 |
| POST | /maintenances | 创建保养/维修工单（含报修） | 登录 |
| POST | /maintenances/plan/generate | 自动生成保养计划 | DEVICE_ADMIN/ENGINEER |
| POST | /maintenances/:id/start | 开始执行工单 | 登录 |
| POST | /maintenances/:id/complete | 完成工单 | 登录 |
| POST | /maintenances/:id/cancel | 取消工单 | 登录 |
| GET | /calibrations | 计量台账列表 | 登录 |
| POST | /calibrations | 建立计量台账 | 登录 |
| GET | /calibrations/due | 计量到期预警清单 | 登录 |
| POST | /calibrations/:id/result | 登记计量结果（不合格自动禁用设备） | 登录 |
| GET | /transfers | 调拨申请列表 | 登录 |
| POST | /transfers | 发起调拨 | 登录 |
| POST | /transfers/:id/approve | 批准调拨（自动更新科室/责任人） | DEVICE_ADMIN/DEAN |
| POST | /transfers/:id/reject | 驳回调拨 | DEVICE_ADMIN/DEAN |
| GET | /scraps | 报废申请列表 | 登录 |
| POST | /scraps | 发起报废 | 登录 |
| POST | /scraps/:id/approve | 批准报废（设备归档） | DEVICE_ADMIN/DEAN |
| POST | /scraps/:id/reject | 驳回报废 | DEVICE_ADMIN/DEAN |
| GET | /audits | 审计日志 | SUPER_ADMIN/DEVICE_ADMIN |
| GET | /stats/overview | 资产统计总览 | 登录 |

### 复用关系标注（≥2 处接口复用同一 service/repository 方法）

- `GET /api/v1/devices?status=...`（设备列表）与 `GET /api/v1/stats/overview`（统计总览）复用 `DeviceRepository.Count`（`internal/repository/device_repository.go`）与 `DeviceRepository.GroupCount`。
- `POST /api/v1/purchases/:id/accept`（验收入台账）、`POST /api/v1/transfers/:id/approve`（调拨）、`POST /api/v1/scraps/:id/approve`（报废）均复用 `DeviceRepository.UpdateStatusTx`（事务内状态流转，`internal/repository/device_repository.go`）。
- 所有写操作（注册/登录/采购/调拨/报废/计量/保养）统一复用 `AuditService.Record`（`internal/service/audit_service.go`）与 `AuditRepository.Create` 写入审计日志。

## curl 调用示例

```bash
# 登录获取 JWT
TOKEN=$(curl -s -X POST http://localhost:19936/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"admin123"}' | python3 -c 'import sys,json;print(json.load(sys.stdin)["data"]["token"])')

# 带 JWT 请求头查询设备台账
curl -s http://localhost:19936/api/v1/devices?page=1\&page_size=10 \
  -H "Authorization: Bearer $TOKEN"

# 提交采购申请
curl -s -X POST http://localhost:19936/api/v1/purchases \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"department":"心内科","applicant_name":"张医生","device_name":"除颤仪","quantity":1,"budget_amount":80000}'

# 资产统计总览
curl -s http://localhost:19936/api/v1/stats/overview -H "Authorization: Bearer $TOKEN"
```

## Docker 部署说明

- 端口映射：前端 `${FRONTEND_PORT:-18936}:80`，后端 `${BACKEND_PORT:-19936}:8080`，数据库 `${DB_PORT:-44017}:3306`，Redis `${REDIS_PORT:-46317}:6379`。
- 数据卷：`db-data`（MySQL 数据）、`redis-data`（Redis 数据）使用命名卷持久化，避免绑定挂载中文路径问题，支持在任意目录名下启动。
- 健康检查：MySQL 使用 `mysqladmin ping`，后端使用 `wget /healthz`；后端通过 `depends_on: db: condition: service_healthy` 等待数据库就绪。
- 常见问题：
  - 端口被占用：修改 `.env` 中 `FRONTEND_PORT/BACKEND_PORT/DB_PORT/REDIS_PORT` 后重新 `docker compose up -d`。
  - 数据库初始化失败：删除数据卷后重建 `docker compose down -v && docker compose up -d --build`。
  - 中文目录名启动：本工程所有路径均在容器内固定（`/app`、`/usr/share/nginx/html`），与宿主机目录名无关，可直接在中文目录下 `docker compose up -d`。

## 枚举出现位置清单

| 枚举 | 后端出现位置 | 前端出现位置 |
| --- | --- | --- |
| 角色 RoleType（SUPER_ADMIN/DEVICE_ADMIN/DEAN/DEPARTMENT/ENGINEER） | `internal/constants/roles.go`、`internal/model/user.go`、`internal/dto/user_dto.go`、`internal/middleware/rbac.go`、`internal/router/user.go`、`internal/router/purchase.go`、`internal/router/maintenance.go`、`internal/router/transfer.go`、`internal/router/scrap.go`、`internal/router/audit.go`、`internal/util/formatters.go(RoleText)`、`internal/service/user_service.go` | `src/constants/enums.ts`、`src/app/guards/role.guard.ts`、`src/app/layouts/main-layout.component.ts`、`src/app/pages/purchases/purchases.component.ts`、`src/app/pages/transfers/transfers.component.ts`、`src/app/pages/scraps/scraps.component.ts` |
| 设备状态 DeviceStatus（in_storage/in_use/under_maintenance/disabled/scrapped） | `internal/constants/status.go`、`internal/model/device.go`、`internal/dto/device_dto.go`、`internal/service/device_service.go`、`internal/service/purchase_service.go`、`internal/service/maintenance_service.go`、`internal/service/calibration_service.go`、`internal/service/scrap_service.go`、`internal/repository/device_repository.go`、`internal/util/formatters.go`、`internal/constants/error_codes.go`、`internal/constants/log_templates.go` | `src/constants/enums.ts`、`src/app/components/status-badge/status-badge.component.ts`、`src/app/pages/devices/devices.component.ts`、`src/utils/format.ts` |
| 采购状态 PurchaseStatus（pending_device_admin/pending_dean/approved/delivered/accepted/rejected） | `internal/constants/status.go`、`internal/model/purchase_request.go`、`internal/service/purchase_service.go`、`internal/repository/purchase_repository.go`、`internal/util/formatters.go`、`internal/constants/log_templates.go`、`internal/constants/messages.go` | `src/constants/enums.ts`、`src/app/pages/purchases/purchases.component.ts`、`src/app/components/status-badge/status-badge.component.ts`、`src/utils/format.ts` |
| 保养/维修类型与状态（daily/weekly/monthly/yearly/repair；pending/in_progress/completed/cancelled） | `internal/constants/status.go`、`internal/model/maintenance_record.go`、`internal/dto/maintenance_dto.go`、`internal/service/maintenance_service.go`、`internal/repository/maintenance_repository.go`、`internal/util/formatters.go`、`internal/constants/log_templates.go` | `src/constants/enums.ts`、`src/app/pages/maintenance/maintenance.component.ts`、`src/app/components/status-badge/status-badge.component.ts`、`src/utils/format.ts` |
| 计量状态 CalibrationStatus（normal/unqualified/due/expired） | `internal/constants/status.go`、`internal/model/calibration_record.go`、`internal/service/calibration_service.go`、`internal/repository/calibration_repository.go`、`internal/util/formatters.go`、`internal/constants/log_templates.go` | `src/constants/enums.ts`、`src/app/pages/calibrations/calibrations.component.ts`、`src/app/components/status-badge/status-badge.component.ts`、`src/utils/format.ts` |
| 调拨/报废状态（pending/approved/rejected） | `internal/constants/status.go`、`internal/model/transfer_request.go`、`internal/model/scrap_request.go`、`internal/service/transfer_service.go`、`internal/service/scrap_service.go`、`internal/util/formatters.go`、`internal/constants/log_templates.go` | `src/constants/enums.ts`、`src/app/pages/transfers/transfers.component.ts`、`src/app/pages/scraps/scraps.component.ts`、`src/app/components/status-badge/status-badge.component.ts`、`src/utils/format.ts` |

## 横切关注点触达文件层

1. **JWT 认证 + RBAC 权限**：数据库角色字段（`internal/model/user.go`）→ `internal/middleware/auth.go`、`internal/middleware/rbac.go`、`internal/util/jwt.go`、`internal/constants/roles.go` → 前端 `src/app/guards/auth.guard.ts`、`src/app/guards/role.guard.ts`、`src/app/interceptors/auth.interceptor.ts`、页面按钮显隐（`src/app/pages/purchases/purchases.component.ts` 等 `canXxx` 方法）。
2. **操作审计日志**：数据库审计表（`internal/model/audit_log.go`）→ `internal/middleware/audit.go`（写操作自动审计）→ `internal/service/audit_service.go` + 各业务 service 埋点 → 前端 `src/app/pages/audits/audits.component.ts`（审计页面）、`src/api/audit.api.ts`、`src/stores/audit.store.ts`。
3. **全局错误处理与请求追踪**：`internal/middleware/request_id.go`、`internal/middleware/error_handler.go`、`internal/util/app_error.go`、`internal/constants/error_codes.go` → 前端 `src/utils/request.ts`（统一解析与错误拦截）、`src/app/interceptors/error.interceptor.ts`。

## 屎山代码设计说明（低内聚、高耦合、牵一发动全身）

- **日志模块单独管理但全栈引用**：`internal/util/logger.go` 使用 `log/slog` 封装，全部 handler/service/middleware 引用；25+ 条日志模板集中在 `internal/constants/log_templates.go`，业务字段变更必须同步修改模板与调用处。
- **异常信息分散且层层透传**：错误码集中在 `internal/constants/error_codes.go`，各 service/handler 手动拼接 message（含实体名/字段名/角色名），handler 再次包装 service 错误。
- **常量/工具类多处耦合**：`internal/util/formatters.go` 同时包含日期、状态文本、类型文本格式化；`internal/constants/messages.go` 同时包含接口返回文案、日志文案、错误提示文案。
- **状态机跨多处定义**：核心状态流转规则同时存在于 service 状态机、前端按钮显隐、日志模板、错误码、formatters，新增一个状态值需修改 ≥10 处。
- **枚举多处重复定义**：核心枚举在后端 constants、DTO、模型、日志模板、错误码、formatters 与前端 constants 中同时存在，新增枚举值需修改 ≥10 处文件。

## License

MIT
