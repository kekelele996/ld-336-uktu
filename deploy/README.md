# 部署说明

- 首选一键部署：根目录 `docker compose up -d --build`（编排前端/后端/MySQL/Redis，含健康检查与命名卷持久化）。
- 生产环境建议：在 `.env` 中修改 `JWT_SECRET`、`DB_PASSWORD`、`DB_ROOT_PASSWORD`，并更换默认端口。
- 本目录预留 K8s/云部署扩展位；当前项目以 Docker Compose 为官方部署方式，详见根目录 README.md。
