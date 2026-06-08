# Infinitech 2.0

Infinitech 2.0 是面向外卖、团购、买药、快递/跑腿、圈子小微墙的一体化本地生活平台。当前仓库从空目录开始，旧版 `HCRXchenghong/infinitech` 只作为 UI、品牌、业务模块和工程经验的复用来源，不再作为 2.0 主线直接延续；`HCRXchenghong/InfiniLink` 只作为圈子信息流参考，不整仓嵌入。

## 当前交付物

- [平台总计划](./PLATFORM_MASTER_PLAN.md)
- [当前状态总览](./PROJECT_STATUS.md)
- [最近进展与路线图](./docs/product/recent-progress-roadmap.md)
- [旧版复用审计](./docs/LEGACY_REUSE_AUDIT.md)
- [美团与旧版能力对标矩阵](./docs/product/meituan-legacy-parity-matrix.md)
- [执行台账](./EXECUTION_LEDGER.md)
- 用户端原生微信小程序店铺详情页已接 `GET /api/shops/{shopID}/detail`，评价与商家信息不再是占位文案。
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
  - `apps/admin-web/src/adminDetails.mjs`
  - `apps/admin-web/src/adminOperations.mjs`
  - `apps/admin-web/src/adminTable.mjs`
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
  - 已接 `/api/admin/audit-logs/archive/records`、`/api/admin/audit-logs/archive/verify` 和 `/api/admin/audit-logs/archive/verifications` 归档记录、归档对象校验与校验历史入口
	  - 已接 `/api/admin/rbac/policy`、`/api/admin/rbac/change-requests`、`/api/admin/rbac/change-requests/{id}/review`、`/api/admin/rbac/change-requests/{id}/apply` 和 `/api/admin/rbac/change-requests/{id}/rollback` 权限治理入口
	  - 已接 `/api/admin/merchant-qualifications` 和 `/api/admin/merchant-qualifications/{qualificationID}` 商户资质待审列表/明细入口，可查看商户、店铺、保证金、资质风险、最近审核审计、推荐动作和核查清单
	  - 已接 `/api/admin/merchant-qualifications/{qualificationID}/review` 商户资质审核入口，商户资质上传后进入待审，商户模块和详情抽屉可预填审核表单并进入高风险二次确认
	  - 商户资质审核结果已投递 `merchant.qualification_reviewed` 可靠事件，`notification-worker` 可生成商户通知 payload，outbox relay 默认 topic 与 Docker/K8s 部署骨架已覆盖
	  - 已新增商户站内通知中心首版，`notification-worker` 可把可靠通知写入 `/api/notifications`，商户可通过 `/api/merchant/notifications` 查看未读消息并通过 `/api/merchant/notifications/{notificationID}/read` 标记已读
		  - 已新增 `/api/admin/notifications` 通知运营查询入口，可按目标商户、状态、来源 topic/event 查询通知台账，支持客服只读排查和运营回溯
		  - 已新增通知投递回执台账首版，worker 可通过 `/api/notifications/{notificationID}/deliveries` 记录投递成功/失败，后台可通过 `/api/admin/notification-deliveries` 查询失败原因和回执状态
		  - 已新增 Admin Web 通知运营页首版，通知模块可打开通知台账、通知回执和补录回执表单，运营可从详情抽屉追踪来源事件、失败原因和 provider 错误码，补录回执进入高风险二次确认
			  - 已新增通知失败回执告警首版，运营可通过 `/api/admin/notification-deliveries/failure-alerts/emit` 把 failed 回执汇总投递到 `notification.delivery_failed_alerts` outbox topic，并写入 `admin.notification_delivery_failure_alerts.emitted` 审计
			  - 已新增通知失败重试编排首版，运营可通过 `/api/admin/notification-deliveries/retries/schedule` 按渠道/provider 退避时间安排 failed 回执重试，生成 `notification.delivery_retries` outbox 事件并写入 `admin.notification_delivery_retries.scheduled` 审计
			  - 已新增通知 provider 执行器骨架，`notification-worker` 可按 `NOTIFICATION_PROVIDER_CHANNELS` 或重试事件生成短信/企微/订阅消息/push provider dispatch，调用配置的 provider endpoint/adapter，并把 delivered/failed 回执写回 `/api/notifications/{notificationID}/deliveries`
			  - 已新增通知 provider 回调验签入账首版，外部渠道可回调 `/api/notifications/provider-callback`，生产配置 `NOTIFICATION_PROVIDER_CALLBACK_SECRET` 后按 HMAC-SHA256 canonical lines 验签并幂等写入投递回执账本
			  - 已新增通知 provider 模板映射与渠道 payload 规范首版，`notification-worker` 可通过 `NOTIFICATION_PROVIDER_TEMPLATES` 把通知 type/template_key 映射成短信、企微、订阅消息、push 的 `template_id`、模板参数和渠道 payload
			  - 已新增通知偏好与静默窗口首版，`notification-worker` 可通过 `NOTIFICATION_DELIVERY_PREFERENCES` 按目标/类型禁用外部渠道或在静默时段把外部投递转为 queued 回执，避免误触达又保留运营证据
			  - 已新增通知偏好后端账本与 API 首版，新增 `platform_notification_preferences`、`GET/PUT /api/admin/notification-preferences` 和 `GET/PUT /api/merchant/notification-preferences`，商户可维护自身通知类型偏好，运营写入偏好会记录 `admin.notification_preferences.saved` 审计
			  - 已新增通知 worker 后端偏好读取首版，`notification-worker` 会按 `preference_key` 精确读取后端偏好账本并在 provider 投递前执行；偏好读取失败时外部渠道转为 queued 回执并记录 `notification_preference_lookup_failed`
			  - 已新增通知静默 queued 再投递调度首版，后台可按 `status=queued`、`error_code=notification_quiet_window` 和 `retry_at` 把静默窗口回执调度成 `notification.delivery_retries` 延迟 outbox，Admin Web 可筛错误码并指定重试时间
			  - 已新增通知静默到期自动扫描调度首版，quiet-window queued 回执会记录静默结束 `retry_at`，后台可通过 `/api/admin/notification-deliveries/quiet-window-retries/schedule` 扫描到期回执并生成延迟重试 outbox，worker 可通过环境变量开启自动调度循环
			  - 已新增商户端通知偏好设置首版，`apps/merchant-flutter` 可打开“通知偏好”页，按通知类型配置短信、微信订阅、企业微信、push 开关和静默时间，并调用 `GET/PUT /api/merchant/notification-preferences`
			  - 已新增管理端通知偏好操作入口首版，Admin Web 通知运营页和详情抽屉可查询 `/api/admin/notification-preferences`，也可预填目标商户、通知类型、禁用渠道和 `quiet_hours` JSON 后保存偏好，保存动作进入高风险二次确认
			  - 已新增用户端通知偏好设置首版，原生微信小程序新增 `pages/notification-preferences/index` 和首页入口，用户可按订单、售后和优惠活动配置微信订阅、短信、App Push 开关与静默时间，并调用 `GET/PUT /api/user/notification-preferences`
			  - 已新增通知 worker 偏好缓存与失败关闭首版，按 preference key 缓存后端偏好，短时抖动时使用 stale 偏好，无法读取且无缓存时仍把外部投递转为 queued，避免绕过用户/商户选择
			  - 已新增通知偏好变更事件与 worker 主动失效首版，保存偏好会写入 `notification.preferences_changed` outbox，relay 默认发布，worker 消费后只失效对应偏好缓存 key，不会误创建站内通知
			  - 已新增通知偏好批量保存与策略审计首版，运营可通过 `/api/admin/notification-preferences/batch` 一次保存多条偏好，PostgreSQL-backed Store 在同一事务内写入偏好、变更 outbox 和 `admin.notification_preferences.batch_saved` 审计，Admin Web 已接高风险二次确认
			  - 已新增通知偏好变更审批与灰度应用首版，运营可通过 `/api/admin/notification-preferences/change-requests` 提交申请并固化 `all`、`target_ids` 或 `percentage` 灰度策略，另一名管理员审批后再通过 `/apply` 按范围进入批量保存、偏好变更 outbox 和 `admin.notification_preferences.change_applied` 审计路径
			  - 已接 `/api/after-sales/{requestID}/review` 售后审核入口，售后模块和详情抽屉可预填审核表单并进入高风险二次确认
			  - 已新增 Admin Web 客服工作台首版，支持 `GET /api/admin/service-tickets?sla_status=` 查询工单队列，并通过 `POST /api/admin/service-tickets/{ticketID}/assign` 分派客服、`POST /api/admin/service-tickets/{ticketID}/escalate` 升级 SLA 超时工单、`POST /api/admin/service-tickets/{ticketID}/resolve` 提交处理方案；客服质检/绩效已接 `POST /api/admin/service-tickets/{ticketID}/quality-review`、`GET /api/admin/service-ticket-quality-reviews`、`GET /api/admin/service-ticket-performance`，客服消息敏感信息风控已覆盖 IM 发送、工单创建和工单事件追加，写动作进入高风险二次确认
  - 已接 `/api/orders/{orderID}/refund` 订单退款入口，订单模块和详情抽屉可预填退款表单并进入高风险二次确认
  - 已接 `/api/admin/outbox/events/{eventID}/replay` Outbox 单事件恢复入口，运营首页和 Outbox 详情抽屉可预填事件 ID 并进入高风险二次确认
  - 已接 `/api/admin/outbox/events/{eventID}/failed` 和 `/api/admin/outbox/events/{eventID}/published` Outbox 人工失败/发布处置入口，运营首页和 Outbox 详情抽屉可预填处置表单并进入高风险二次确认
  - 已接 `/api/admin/outbox/events/claim` 和 `/api/admin/outbox/events/{eventID}/lease/renew` Outbox 租约领取/续租入口，运营首页和 Outbox 详情抽屉可预填租约表单并进入高风险二次确认
  - 已接 `GET /api/admin/outbox/events?status=dead_letter` Outbox 死信分诊预设和 `/api/admin/outbox/events/{eventID}/replay` 死信解封入口，运营首页和 Outbox 详情抽屉可直接进入死信分诊与高风险解封确认
  - 已接 `GET /api/admin/outbox/events/{eventID}` Outbox 单事件事故辅助明细入口，可查看事件状态、payload 摘要、关联目标、最近处置审计、推荐下一步动作和人工处置核查清单
  - 已新增审计中心增强首版，支持 actor/action/target/after/before/limit 筛选、保存筛选、详情抽屉、跨模块跳转、脱敏 payload 摘要和归档校验历史可视化面板
	  - 已新增 P0/P1 业务详情面板首版，订单、售后、商户、骑手、绩效、派单、退款策略、客服工作台、通知运营和权限治理表格行可打开详情，并直接跳到补偿、售后审核、客服分派、SLA 升级、工单方案、客服质检、客服绩效、审计、outbox、对象清理、通知回执、通知失败告警和 RBAC 等下一步操作
  - 已新增高风险操作二次确认和结果追踪首版，邀约、售后审核、客服工单分派、客服 SLA 升级、客服处理方案、订单退款、退款策略、审计导出/告警/归档、RBAC 变更、Outbox 领取/续租/死信解封/单事件恢复/批量恢复/标记失败/标记已发布和订单状态补偿执行前会进入确认面板，执行后保留最近操作结果
  - 已新增失败回放入口首版，失败记录可一键恢复原操作和参数，高风险动作重试时仍需再次二次确认
  - 已新增 P0 业务筛选分页首版，业务视图支持关键字筛选、每页条数和上一页/下一页控制
  - 已新增商户资质审核后端与表单首版，支持 `approve`/`reject` 决策、审核原因和审核时间，审核通过后才计入商户接单资格
  - 已新增售后审核表单首版，支持 `approve`、`reject`、`escalate` 决策、审核原因、退款去向和退款幂等键
  - 已新增订单退款表单首版，支持退款原因、幂等键、可选退款金额和余额/原路退款去向
  - 已新增 Outbox 单事件恢复表单首版，支持 `event_id` 预填和可选 `now` 参数，执行前必须二次确认
  - 已新增 Outbox 发布/失败人工处置表单首版，支持失败原因、重试延迟、最大尝试次数和发布时间，执行前必须二次确认
  - 已新增 Outbox 领取/续租表单首版，支持 topic、limit、lease owner、lease seconds 和事件 ID，执行前必须二次确认
  - 已新增 Outbox 死信分诊/解封表单首版，支持 `dead_letter` 查询预设、事件 ID 预填和高风险解封二次确认
- 管理端审计账本：
  - 关键后台写操作记录 actor、action、target、request_id、ip_hash、服务端白名单 payload 和创建时间
  - 已落地 `security_auditor` 只读审计角色，可读 `/api/admin/audit-logs`，不能执行邀请、补偿等后台写操作
  - 已落地审计导出首版，`GET /api/admin/audit-logs/export` 可按同一套筛选条件导出 CSV，导出行为写入 `admin.audit_logs.exported` 审计
  - 已落地审计留存/告警健康报告首版，`GET /api/admin/audit-logs/retention-report` 可返回留存窗口、冷归档候选、完整性失败、导出事件和关键动作覆盖告警
  - 已落地审计留存告警 outbox 投递首版，`POST /api/admin/audit-logs/retention-alerts/emit` 可把报告告警投递到 `audit.retention_alerts` 并写入 `admin.audit_retention_alerts.emitted`
  - 已落地审计 WORM/冷归档请求首版，`POST /api/admin/audit-logs/archive/request` 可按热存窗口生成归档 manifest、投递 `audit.archive_requested` outbox 事件并写入 `admin.audit_archive.requested`
  - 已落地审计归档 worker 首版，`services/audit-archive-worker` 可领取 `audit.archive_requested`、校验 manifest hash、上传 JSONL 归档文件、写入对象锁请求头，回写 `/api/admin/audit-logs/archive/complete` 完成证据，并把 outbox 标记为 published/failed
  - 已落地归档对象下载校验/回查首版，`POST /api/admin/audit-logs/archive/verify` 可读取配置的 `AUDIT_ARCHIVE_DOWNLOAD_BASE_URL` 归档对象，校验 content hash、manifest header、bytes 和条目数，并写入 `admin.audit_archive.verified` 审计
  - 已落地归档校验历史查询和可视化面板首版，`GET /api/admin/audit-logs/archive/verifications` 可按归档 ID、状态、时间范围和 limit 回看 `admin.audit_archive.verified` 校验台账，Admin Web 审计中心可展示状态、hash、bytes/logs、匹配摘要和原始详情
  - 已落地服务端 RBAC 策略矩阵首版，后台角色包含兼容 `admin`、`super_admin`、`ops_admin`、`finance_admin`、`dispatch_admin`、`support_admin` 和 `security_auditor`，邀约、退款、售后、对象清理、outbox、调度、运营快照和审计读取已按 scope 守护
  - 已落地 RBAC 权限治理查询、变更申请、审批/驳回、手动应用和审计回滚首版，分权后台角色可读服务端真实矩阵，`admin`/`super_admin` 可提交待审批申请，另一名管理员可审批或驳回；已审批申请可通过单独应用动作进入运行时权限矩阵，当前已应用申请可按应用前 scopes 回滚，申请、审批、应用和回滚分别写入 `admin.rbac.change_requested`、`admin.rbac.change_reviewed`、`admin.rbac.change_applied` 和 `admin.rbac.change_rolled_back` 审计
  - 商户/骑手邀约、退款策略、订单退款、状态补偿、售后审核、对象清理和 outbox 运维首批纳入审计
  - PostgreSQL-backed Store 已使用规范化 `audit_logs` 表查询审计，并兼容旧快照审计回填
  - 审计 payload 已在内存 Store、PostgreSQL 写入、SQL 读取和镜像恢复路径统一白名单过滤，对 token、phone、object key、签名等敏感字段做服务端掩码
  - 审计日志已新增 `integrity_algorithm`、`integrity_hash`、`integrity_verified` 完整性证明字段；本地默认 `sha256:v1`，生产配置 `AUDIT_LOG_SIGNING_SECRET` 后使用 `hmac-sha256:v1`，查询时可发现审计字段或白名单 payload 被篡改
  - 退款策略配置、管理端订单退款、售后审核、订单状态补偿、对象清理完成/失败、outbox 运维和商户/骑手邀约已分别通过仓储级 `SaveRefundSettingsWithAudit`、`RefundOrderWithAudit`、`ReviewAfterSalesWithAudit`、`CompensateOrderStateWithAudit`、`CompleteObjectStorageCleanupWithAudit`、`RecordObjectStorageCleanupFailureWithAudit`、outbox `WithAudit` 方法以及 `CreateMerchantInviteWithAudit`、`CreateRiderInviteWithAudit` 完成业务写入和审计写入同事务首版；PostgreSQL-backed Store 在同一数据库事务内写入业务表、`platform_outbox_events` 或邀约快照与 `audit_logs`
  - Admin Web 已对审计 payload 做白名单摘要和敏感字段脱敏展示，并可按时间范围回溯、保存常用筛选、查看完整性状态和跳到相关运营模块
- BFF 浏览器接入：
  - 已支持本地管理端/Flutter 调试来源的 CORS 白名单与 `OPTIONS` 预检
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
- 用户微信小程序继续使用微信原生语言栈 `WXML/WXSS/TypeScript`，不改成 Flutter 或跨端编译。
- 用户端已新增通知偏好入口，订单、售后和优惠活动触达可由用户自助配置。
- 用户端微信登录页已接真实 `wx.login` code 流程，生产接口失败不会自动写开发 token；本地 API 或显式 `allowPreviewAuth=true` 时才允许预览兜底。
- 用户端登录/注册手机号验证码已补生产短信 provider 配置入口、验证码隐藏、重发冷却和手机号频控；开发模式仍可回填验证码便于本地预览。
- 用户端消息/商户群聊已补游标离线补偿、打开会话清未读、手动已读回执、`message.sent` outbox 出站事件、PostgreSQL 规范化消息表和已读状态表首版、realtime-gateway WebSocket 首版投递、Redis adapter 多副本 fanout、WebSocket 签名 token 鉴权、会话成员权限校验、会话免打扰偏好，以及群资料/成员预览、自助入群/退群和商户群券资格首版，小程序商户群已接实时连接、群设置、进群领券和免打扰切换。
- 用户端红包已补领取风控首版，支持同群短时频次、24 小时累计金额校验、重复领取幂等返回和小程序风险提示展示。
- 用户端找饭搭已补资料人工审核、审核状态展示、候选放行、同校/同楼隐私范围、模糊位置、设备风控、举报待审和举报成立后暂停展示首版。
- 用户端买药处方链路已补处方影像上传票据、上传回调、对象扫描门禁、上传确认、OCR 识别摘要、药师复核、处方留档、审核单对象元数据绑定和药品库存锁定首版。
- 首页新增后台可控圈子入口和推荐卡片，找饭搭要求性别、学校/楼栋、同校或同楼隐私范围、设备安全校验、真实性承诺、平台免责承诺和问卷。
- 商户端、骑手端先分别做独立 `Flutter/Dart` 客户端。
- 管理端桌面优先 Web，移动管理端用 `Flutter/Dart`，后续各端再按需要补原生能力。
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

`npm run dev:api-go` 会使用 `WECHAT_MINI_LOGIN_MODE=dev` 的本地微信登录解析器。生产启动必须配置 `WECHAT_MINI_APP_ID` 和 `WECHAT_MINI_APP_SECRET`，用于调用微信小程序 `code2session` 换取真实 `openid`。用户端页面只有在本地 API 或显式 `allowPreviewAuth=true` 时才会启用开发 token 兜底。

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
