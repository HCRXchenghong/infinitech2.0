# Infinitech 2.0

Infinitech 2.0 是面向外卖、团购、买药、快递/跑腿、圈子小微墙的一体化本地生活平台。当前仓库从空目录开始，旧版 `HCRXchenghong/infinitech` 只作为 UI、品牌、业务模块和工程经验的复用来源，不再作为 2.0 主线直接延续；`HCRXchenghong/InfiniLink` 只作为圈子信息流参考，不整仓嵌入。

## 当前交付物

- [平台总计划](./PLATFORM_MASTER_PLAN.md)
- [当前状态总览](./PROJECT_STATUS.md)
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
  - `apps/admin-web/src/adminViews.mjs`
  - `apps/admin-web/src/adminSnapshot.mjs`
  - `apps/admin-web/src/config.mjs`
  - 已接 `/api/admin/operations/snapshot` 运营快照入口
  - 已用运营快照生成 P0 指标和表格首版
  - 已接 `/api/admin/audit-logs` 操作审计入口
- 管理端审计账本：
  - 关键后台写操作记录 actor、action、target、request_id、ip_hash、非敏感 payload 和创建时间
  - 商户/骑手邀约、退款策略、订单退款、状态补偿、售后审核、对象清理和 outbox 运维首批纳入审计
  - PostgreSQL-backed Store 已使用规范化 `audit_logs` 表查询审计，并兼容旧快照审计回填
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

`@infinitech/object-lifecycle-worker` 会通过管理 API 拉取过期未确认或扫描拒绝的对象清理候选，删除对象存储里的文件后回写票据 `deleted` 状态；删除失败会回写 `cleanup_attempts`、`last_cleanup_error` 和 `last_cleanup_failed_at`，后台可通过清理统计查看 pending、failed、deleted 和累计尝试次数，并继续重试和接入告警。生产必须使用最小权限删除凭据，并继续补 STS/Vault 临时凭证、隔离 bucket 策略、告警投递和删除凭据轮换。
