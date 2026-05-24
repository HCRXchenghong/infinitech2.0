# Infinitech 2.0

Infinitech 2.0 是面向外卖、团购、买药、快递/跑腿、圈子小微墙的一体化本地生活平台。当前仓库从空目录开始，旧版 `HCRXchenghong/infinitech` 只作为 UI、品牌、业务模块和工程经验的复用来源，不再作为 2.0 主线直接延续；`HCRXchenghong/InfiniLink` 只作为圈子信息流参考，不整仓嵌入。

## 当前交付物

- [平台总计划](./PLATFORM_MASTER_PLAN.md)
- [当前状态总览](./PROJECT_STATUS.md)
- [最近进展与路线图](./docs/product/recent-progress-roadmap.md)
- [旧版复用审计](./docs/LEGACY_REUSE_AUDIT.md)
- [美团与旧版能力对标矩阵](./docs/product/meituan-legacy-parity-matrix.md)
- [执行台账](./EXECUTION_LEDGER.md)
- GitHub 协作门禁：
  - `.github/workflows/verify.yml`
  - `.github/pull_request_template.md`
  - `.github/ISSUE_TEMPLATE/`
  - `.github/CODEOWNERS`
  - `.github/dependabot.yml`
- 管理端 Web 最小运营控制台：
  - `apps/admin-web/index.html`
  - `apps/admin-web/src/main.js`
  - `apps/admin-web/src/adminApi.mjs`
  - `apps/admin-web/src/adminAudit.mjs`
  - `apps/admin-web/src/adminViews.mjs`
  - `apps/admin-web/src/adminSnapshot.mjs`
  - `apps/admin-web/src/config.mjs`
  - 已接 `/api/admin/operations/snapshot` 运营快照入口
  - 已用运营快照生成 P0 指标和表格首版
  - 已接 `/api/admin/audit-logs` 操作审计入口
  - 已接 `/api/admin/audit-logs/export` 审计 CSV 导出入口
  - 已接 `/api/admin/audit-logs/retention-report` 审计留存/告警健康报告入口
  - 已接 `/api/admin/audit-logs/retention-alerts/emit` 审计留存告警 outbox 投递入口
  - 已接 `/api/admin/audit-logs/archive/request` 审计 WORM/冷归档请求入口
  - 已接 `/api/admin/rbac/policy`、`/api/admin/rbac/change-requests`、`/api/admin/rbac/change-requests/{id}/review`、`/api/admin/rbac/change-requests/{id}/apply` 和 `/api/admin/rbac/change-requests/{id}/rollback` 权限治理入口
  - 已新增审计中心增强首版，支持 actor/action/target/after/before/limit 筛选、保存筛选、详情抽屉、跨模块跳转和脱敏 payload 摘要
- 管理端审计账本：
  - 关键后台写操作记录 actor、action、target、request_id、ip_hash、服务端白名单 payload 和创建时间
  - 已落地 `security_auditor` 只读审计角色，可读 `/api/admin/audit-logs`，不能执行邀请、补偿等后台写操作
  - 已落地审计导出首版，`GET /api/admin/audit-logs/export` 可按同一套筛选条件导出 CSV，导出行为写入 `admin.audit_logs.exported` 审计
  - 已落地审计留存/告警健康报告首版，`GET /api/admin/audit-logs/retention-report` 可返回留存窗口、冷归档候选、完整性失败、导出事件和关键动作覆盖告警
  - 已落地审计留存告警 outbox 投递首版，`POST /api/admin/audit-logs/retention-alerts/emit` 可把报告告警投递到 `audit.retention_alerts` 并写入 `admin.audit_retention_alerts.emitted`
  - 已落地审计 WORM/冷归档请求首版，`POST /api/admin/audit-logs/archive/request` 可按热存窗口生成归档 manifest、投递 `audit.archive_requested` outbox 事件并写入 `admin.audit_archive.requested`
  - 已落地审计归档 worker 首版，`services/audit-archive-worker` 可领取 `audit.archive_requested`、校验 manifest hash、上传 JSONL 归档文件、写入对象锁请求头，回写 `/api/admin/audit-logs/archive/complete` 完成证据，并把 outbox 标记为 published/failed
  - 已落地服务端 RBAC 策略矩阵首版，后台角色包含兼容 `admin`、`super_admin`、`ops_admin`、`finance_admin`、`dispatch_admin`、`support_admin` 和 `security_auditor`，邀约、退款、售后、对象清理、outbox、调度、运营快照和审计读取已按 scope 守护
  - 已落地 RBAC 权限治理查询、变更申请、审批/驳回、手动应用和审计回滚首版，分权后台角色可读服务端真实矩阵，`admin`/`super_admin` 可提交待审批申请，另一名管理员可审批或驳回；已审批申请可通过单独应用动作进入运行时权限矩阵，当前已应用申请可按应用前 scopes 回滚，申请、审批、应用和回滚分别写入 `admin.rbac.change_requested`、`admin.rbac.change_reviewed`、`admin.rbac.change_applied` 和 `admin.rbac.change_rolled_back` 审计
  - 商户/骑手邀约、退款策略、订单退款、状态补偿、售后审核、对象清理和 outbox 运维首批纳入审计
  - PostgreSQL-backed Store 已使用规范化 `audit_logs` 表查询审计，并兼容旧快照审计回填
  - 审计 payload 已在内存 Store、PostgreSQL 写入、SQL 读取和镜像恢复路径统一白名单过滤，对 token、phone、object key、签名等敏感字段做服务端掩码
  - 审计日志已新增 `integrity_algorithm`、`integrity_hash`、`integrity_verified` 完整性证明字段；本地默认 `sha256:v1`，生产配置 `AUDIT_LOG_SIGNING_SECRET` 后使用 `hmac-sha256:v1`，查询时可发现审计字段或白名单 payload 被篡改
  - 退款策略配置、管理端订单退款、售后审核、订单状态补偿、对象清理完成/失败、outbox 运维和商户/骑手邀约已分别通过仓储级 `SaveRefundSettingsWithAudit`、`RefundOrderWithAudit`、`ReviewAfterSalesWithAudit`、`CompensateOrderStateWithAudit`、`CompleteObjectStorageCleanupWithAudit`、`RecordObjectStorageCleanupFailureWithAudit`、outbox `WithAudit` 方法以及 `CreateMerchantInviteWithAudit`、`CreateRiderInviteWithAudit` 完成业务写入和审计写入同事务首版；PostgreSQL-backed Store 在同一数据库事务内写入业务表、`platform_outbox_events` 或邀约快照与 `audit_logs`
  - Admin Web 已对审计 payload 做白名单摘要和敏感字段脱敏展示，并可按时间范围回溯、保存常用筛选、查看完整性状态和跳到相关运营模块
- BFF 浏览器接入：
  - 已支持本地管理端/uni 调试来源的 CORS 白名单与 `OPTIONS` 预检
  - 可通过 `BFF_ALLOWED_ORIGINS` 配置部署来源
  - 明确允许 `Authorization`、`Content-Type`、`X-Client-Kind` 请求头
- 统一品牌资产：
  - `assets/brand/logo.svg`
  - `assets/brand/logo.jpg`
  - `assets/brand/admin-logo.png`

## 当前结论

- 架构采用：自建/混合云 Kubernetes 上的模块化核心 API + 事件驱动 Worker + 多端 BFF + 实时网关架构。
- 主色沿用旧版浅蓝：`#009bf5`。
- 用户端优先做原生微信小程序，视觉沿用旧版用户端。
- 首页新增后台可控圈子入口和推荐卡片，找饭搭要求性别、真实性承诺、平台免责承诺和问卷。
- 商户端、骑手端先分别做独立 `uni-app`。
- 管理端桌面优先 Web，移动管理端用 `uni-app`，后续各端再逐步迁移原生 App。
- 10 万同时在线必须通过分阶段压测和容灾演练验证后才算达标。

## 验证

```bash
npm run verify
npm run test --workspace @infinitech/admin-web
npm run test --workspace @infinitech/bff
npm run test --workspace @infinitech/audit-archive-worker
cd services/api-go && go test -count=1 ./...
```

单独启动开发服务：

```bash
npm run dev:api-go
npm run dev:bff
npm run dev:realtime
python3 -m http.server 4173 --bind 127.0.0.1 --directory .
```

管理端 Web 本地预览地址：`http://127.0.0.1:4173/apps/admin-web/index.html`。

BFF 默认允许 `http://127.0.0.1:4173`、`http://localhost:4173`、`http://127.0.0.1:5173`、`http://localhost:5173`、`http://127.0.0.1:8080`、`http://localhost:8080` 作为浏览器来源。部署到真实域名时用逗号分隔配置 `BFF_ALLOWED_ORIGINS`，例如 `BFF_ALLOWED_ORIGINS=https://admin.example.com,https://m-admin.example.com`；后台接口需要管理员 token，不接受 `*` 通配来源。

`npm run dev:api-go` 会使用 `WECHAT_MINI_LOGIN_MODE=dev` 的本地微信登录解析器。生产启动必须配置 `WECHAT_MINI_APP_ID` 和 `WECHAT_MINI_APP_SECRET`，用于调用微信小程序 `code2session` 换取真实 `openid`。

管理员 bootstrap 密码登录默认关闭，没有内置默认密码。需要启用时同时配置 `ADMIN_BOOTSTRAP_ACCOUNT_ID` 和 `ADMIN_BOOTSTRAP_PASSWORD`；密码长度必须为 8-72 字节，否则 API 会拒绝启动。商户、骑手和站长主体登录都使用邀约注册时设置的账号密码，并由服务端 bcrypt 哈希保存。

售后证据上传票据已支持对象存储配置。生产环境需要配置：

- `OBJECT_STORAGE_PROVIDER=minio`
- `OBJECT_STORAGE_BUCKET`
- `OBJECT_STORAGE_UPLOAD_BASE_URL`
- `OBJECT_STORAGE_PUBLIC_BASE_URL`
- `OBJECT_STORAGE_HEAD_BASE_URL`
- `OBJECT_STORAGE_SIGNING_SECRET`
- `OBJECT_STORAGE_CALLBACK_SIGNING_SECRET`
- `OBJECT_STORAGE_TICKET_TTL_SECONDS`
- `OBJECT_STORAGE_MAX_UPLOAD_BYTES`
- `OBJECT_STORAGE_REQUIRE_HEAD_VERIFICATION`
- `OBJECT_STORAGE_REQUIRE_UPLOAD_CALLBACK`
- `OBJECT_STORAGE_REQUIRE_SCAN_APPROVAL`
- `OBJECT_STORAGE_HEAD_TIMEOUT_SECONDS`

本地未配置 `OBJECT_STORAGE_SIGNING_SECRET` 时仍可跑通开发测试，但不能用于生产。上传票据会返回 `ticket_id`，确认售后附件时必须携带或匹配同一张未过期票据，平台会校验对象 key、用户/角色、类型和大小后才绑定附件。生产建议开启 `OBJECT_STORAGE_REQUIRE_HEAD_VERIFICATION=true`、`OBJECT_STORAGE_REQUIRE_UPLOAD_CALLBACK=true`、`OBJECT_STORAGE_REQUIRE_SCAN_APPROVAL=true`，确认前通过对象存储/CDN HEAD 校验对象存在、大小和类型，并要求上传服务回调和内容扫描通过。

对象扫描 worker 生产环境还需要配置：

- `OBJECT_SCAN_WORKER_TOKEN`
- `OBJECT_STORAGE_DOWNLOAD_BASE_URL`
- `OBJECT_SCAN_SCANNER=clamav`
- `OBJECT_SCAN_MAX_BYTES`
- `OBJECT_SCAN_DOWNLOAD_TIMEOUT_MS`
- `OBJECT_SCAN_CLAMAV_TIMEOUT_MS`
- `OBJECT_SCAN_CLAMAV_CHUNK_BYTES`
- `CLAMAV_HOST`
- `CLAMAV_PORT`

`@infinitech/object-scan-worker` 会优先使用事件中的签名下载 URL，否则按 `OBJECT_STORAGE_DOWNLOAD_BASE_URL/bucket/object_key` 下载对象，再通过 ClamAV INSTREAM 协议扫描并回调扫描结果。生产必须让 worker 只拿到最小权限的临时下载凭据，后续还要补隔离 bucket、扫描失败告警和告警投递。

对象生命周期 worker 生产环境还需要配置：

- `OBJECT_LIFECYCLE_WORKER_TOKEN`
- `OBJECT_STORAGE_DELETE_BASE_URL`
- `OBJECT_STORAGE_DELETE_TOKEN`
- `OBJECT_LIFECYCLE_INTERVAL_MS`
- `OBJECT_LIFECYCLE_BATCH_LIMIT`
- `OBJECT_LIFECYCLE_GRACE_SECONDS`

`@infinitech/object-lifecycle-worker` 会通过管理 API 拉取过期未确认或扫描拒绝的对象清理候选，删除对象存储里的文件后回写票据 `deleted` 状态；删除失败会回写 `cleanup_attempts`、`last_cleanup_error` 和 `last_cleanup_failed_at`，后台可通过清理统计查看 pending、failed、deleted 和累计尝试次数，并继续重试和接入告警。对象清理完成/失败管理入口已通过仓储级原子审计路径在同一数据库事务内写入票据状态与 `audit_logs`。生产必须使用最小权限删除凭据，并继续补 STS/Vault 临时凭证、隔离 bucket 策略、告警投递和删除凭据轮换。
