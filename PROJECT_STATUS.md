# Infinitech 2.0 当前状态总览

更新时间：2026-05-25
目标仓库：`https://github.com/HCRXchenghong/infinitech2.0`  
当前结论：项目已经完成架构基线、monorepo 骨架、首批端侧页面、核心 API 大量业务闭环、BFF 代理、Worker 骨架、PostgreSQL 规范化、outbox/对象存储、管理端审计、审计导出首版、审计留存/告警健康报告首版、审计留存告警 outbox 投递首版、审计 WORM/冷归档请求首版、审计归档 worker 首版、审计归档完成回写和归档记录查询首版、审计归档对象下载校验/回查首版、审计归档校验历史查询首版、审计归档校验历史可视化面板首版、管理端 P0 业务详情面板首版、管理端高风险操作二次确认与结果追踪首版、管理端失败回放入口首版、管理端 P0 业务筛选分页首版、管理端售后审核表单首版、管理端订单退款表单首版、管理端 Outbox 单事件恢复表单首版、管理端 Outbox 发布/失败人工处置表单首版、管理端 Outbox 领取/续租表单首版、管理端 Outbox 死信分诊/解封表单首版、管理端 Outbox 单事件事故辅助明细首版、商户资质审核后端与管理端表单首版、商户资质待审列表与明细接口首版、商户资质审核结果可靠通知首版、商户站内通知中心首版、通知运营查询接口首版、通知投递回执台账首版、Admin Web 通知运营页首版、通知失败回执告警首版、通知失败重试编排首版、通知 provider 执行器首版、通知 provider 回调验签入账首版、通知 provider 模板映射与渠道 payload 规范首版、通知偏好与静默窗口首版、通知偏好后端账本与 API 首版、通知 worker 后端偏好读取首版、通知静默 queued 再投递调度首版、通知静默到期自动扫描调度首版、商户端通知偏好设置首版、管理端通知偏好操作入口首版、用户端通知偏好设置首版、通知 worker 偏好缓存与失败关闭首版、通知偏好变更事件与 worker 主动失效首版、通知偏好批量保存与策略审计首版、通知偏好变更审批与应用首版、服务端 RBAC 策略矩阵、RBAC 权限治理查询/变更申请审计、权限申请审批/驳回台账、权限变更手动应用和权限变更审计回滚首版等多条商业化底座链路；但还没有完成真实生产支付、真实 IM/RTC、完整管理端、真实高可用基础设施、10 万在线压测和容灾演练，所以不能宣称已经商业级可上线，只能说正在按商业级标准推进。

最近完成、当前未完成和下一批优先级已汇总到 `docs/product/recent-progress-roadmap.md`。这份文档用于快速查看最近提交后的项目状态、商业级阻塞项和后续推进顺序。

## 0. 最近进展摘要

最近已完成：

- 管理端 P0 业务视图首版。
- 管理端运营快照接口 `/api/admin/operations/snapshot`。
- 管理端快照数据绑定到 KPI、队列和 P0 表格。
- 管理端 P0 业务详情面板首版，订单、售后、商户、骑手、绩效、派单、退款策略和权限治理表格行可打开详情并跳到审计/补偿/outbox/对象清理/RBAC 等下一步操作。
- 管理端高风险操作二次确认与结果追踪首版，邀约、商户资质审核、退款策略、审计导出/告警/归档、RBAC 变更、Outbox 领取/续租/死信解封/单事件恢复/批量恢复/标记失败/标记已发布和订单状态补偿执行前进入确认面板，执行后保留最近结果。
- 管理端失败回放入口首版，失败记录可恢复原操作和参数，高风险动作重试仍需再次确认。
- 管理端 P0 业务筛选分页首版，业务视图可按关键字筛选、调整每页条数并分页查看。
- 管理端售后审核表单首版，售后模块和详情抽屉可预填工单号、审核结果、原因、退款去向和退款幂等键，提交前进入高风险二次确认，确认后调用现有 `POST /api/after-sales/{requestID}/review` 原子审计路径。
- 管理端订单退款表单首版，订单模块和详情抽屉可预填订单号、退款原因、退款幂等键、退款金额和退款去向，提交前进入高风险二次确认，确认后调用现有 `POST /api/orders/{orderID}/refund` 原子审计路径。
- 管理端 Outbox 单事件恢复表单首版，运营首页和 Outbox 队列详情可预填事件 ID，提交前进入高风险二次确认，确认后调用现有 `POST /api/admin/outbox/events/{eventID}/replay` 原子审计路径。
- 管理端 Outbox 发布/失败人工处置表单首版，运营首页和 Outbox 队列详情可预填失败/发布参数，提交前进入高风险二次确认，确认后调用现有 `POST /api/admin/outbox/events/{eventID}/failed` 或 `POST /api/admin/outbox/events/{eventID}/published` 原子审计路径。
- 管理端 Outbox 领取/续租表单首版，运营首页和 Outbox 队列详情可预填 topic、limit、lease owner、lease seconds 和事件 ID，提交前进入高风险二次确认，确认后调用现有 `POST /api/admin/outbox/events/claim` 或 `POST /api/admin/outbox/events/{eventID}/lease/renew` 原子审计路径。
- 管理端 Outbox 死信分诊/解封表单首版，运营首页和 Outbox 队列详情可预填 `dead_letter` 查询、死信事件 ID 和解封时间，解封前进入高风险二次确认，确认后调用现有 `POST /api/admin/outbox/events/{eventID}/replay` 原子审计路径。
- 管理端 Outbox 单事件事故辅助明细首版，新增 `GET /api/admin/outbox/events/{eventID}`，返回事件状态、ready/blocked/lease 信号、payload 摘要、关联业务目标、最近 outbox 审计、推荐下一步操作和人工处置核查清单；BFF 和 Admin Web 操作目录/详情抽屉已接入。
- 商户资质审核后端与管理端表单首版，商户上传营业执照/健康证后进入 `pending_review`，只有后台 `POST /api/admin/merchant-qualifications/{qualificationID}/review` 审核通过后才计入接单资格；审核写入 `admin.merchant_qualification.reviewed` 审计，BFF 与 Admin Web 商户详情抽屉已接入并进入高风险二次确认。
- 商户资质待审列表与明细接口首版，新增 `GET /api/admin/merchant-qualifications` 和 `GET /api/admin/merchant-qualifications/{qualificationID}`，可按状态/商户/类型查看待审资质、商户/店铺/保证金上下文、接单门槛、最近审核审计、推荐下一步动作和核查清单；BFF 与 Admin Web 商户模块/详情抽屉已接入。
- 商户资质审核结果可靠通知首版，审核原子路径会写入 `merchant.qualification_reviewed` outbox 事件，HTTP 响应返回 `outbox_event`，notification-worker 可生成商户通知 payload，outbox relay 默认 topic 与 Docker/K8s 部署骨架已覆盖。
- 商户站内通知中心首版，新增 `platform_notifications` 账本、`POST /api/notifications` 写入入口、`GET /api/merchant/notifications` 商户列表和 `POST /api/merchant/notifications/{notificationID}/read` 已读入口；notification-worker 可用 worker token 把商户资质审核结果可靠事件幂等写入站内信。
- 通知运营查询接口首版，新增 `GET /api/admin/notifications` 和 `notification:read` scope，客服可只读按商户、状态、来源 topic/event 查询通知账本，运营可追溯商户是否已收到/已读站内通知。
- 通知投递回执台账首版，新增 `platform_notification_deliveries` 账本、`POST /api/notifications/{notificationID}/deliveries` 写入入口和 `GET /api/admin/notification-deliveries` 查询入口，支持记录 delivered/failed、provider message id、错误码和错误信息。
- Admin Web 通知运营页首版，通知模块已从 planned 推进到 wired，可打开通知台账、通知回执和补录回执表单；详情抽屉可按通知、商户、来源 topic 和失败状态预填查询/补录动作，补录回执进入高风险二次确认。
- 通知失败回执告警首版，新增 `POST /api/admin/notification-deliveries/failure-alerts/emit`，运营可按商户、渠道和 provider 汇总 failed 回执，投递 `notification.delivery_failed_alerts` outbox 事件并写入 `admin.notification_delivery_failure_alerts.emitted` 审计；Admin Web 已接入高风险二次确认。
- 通知失败重试编排首版，新增 `POST /api/admin/notification-deliveries/retries/schedule`，运营可按目标、渠道、provider 和退避秒数安排 failed 回执重试，重试事件在 `retry_at` 后才出现在 ready outbox 队列，并写入 `admin.notification_delivery_retries.scheduled` 审计。
- 通知 provider 执行器首版，`notification-worker` 可按 `NOTIFICATION_PROVIDER_CHANNELS` 或 `notification.delivery_retries` 事件生成短信/企微/订阅消息/push provider dispatch，调用配置 endpoint/adapter，并把 delivered/failed 回执写回通知投递台账；重试事件会携带原始通知快照，避免只拿到 notification id 无法重发正文。
- 通知 provider 回调验签入账首版，新增 `POST /api/notifications/provider-callback`，生产配置 `NOTIFICATION_PROVIDER_CALLBACK_SECRET` 后按 HMAC-SHA256 canonical lines 验签，支持 delivered/failed/queued 异步回执幂等入账。
- 通知 provider 模板映射与渠道 payload 规范首版，`notification-worker` 可通过 `NOTIFICATION_PROVIDER_TEMPLATES` 把 notification type/template_key 映射成各渠道 `template_id`、变量和 provider payload，短信、微信订阅消息、企业微信、push 均有规范化 payload。
- 通知偏好与静默窗口首版，`notification-worker` 可通过 `NOTIFICATION_DELIVERY_PREFERENCES` 按默认/目标角色/目标 ID/通知类型禁用外部渠道，或在静默窗口把 provider 投递转为 `queued` 回执并记录原因。
- 通知偏好后端账本与 API 首版，新增 `platform_notification_preferences`、商户 `GET/PUT /api/merchant/notification-preferences` 和运营 `GET/PUT /api/admin/notification-preferences`；商户可维护自身通知类型渠道偏好，运营写入会记录 `admin.notification_preferences.saved` 审计。
- 通知 worker 后端偏好读取首版，`notification-worker` 会在外部 provider 投递前按 `preference_key` 从 `/api/admin/notification-preferences` 精确读取后端偏好；读取失败时不调用 provider，而是记录 `queued` 回执和 `notification_preference_lookup_failed`。
- 通知静默 queued 再投递调度首版，`/api/admin/notification-deliveries/retries/schedule` 现可按 `status=queued`、`error_code=notification_quiet_window` 和 `retry_at` 调度静默窗口回执，生成 `notification.delivery_retries` 延迟 outbox，Admin Web 支持错误码筛选和指定重试时间。
- 通知静默到期自动扫描调度首版，quiet-window queued 回执会记录静默结束 `retry_at`；`/api/admin/notification-deliveries/quiet-window-retries/schedule` 可扫描到期回执并生成可靠重投 outbox，`notification-worker` 可按环境变量开启自动调度循环。
- 商户端通知偏好设置首版，`apps/merchant-flutter` 新增“通知偏好”页和首页入口，商户可按通知类型配置短信、微信订阅、企业微信、push 开关与静默时间，并调用商户偏好 API 保存。
- 管理端通知偏好操作入口首版，Admin Web 通知运营页可查询和保存 `/api/admin/notification-preferences`，详情抽屉可按失败回执预填目标商户、通知类型、禁用渠道和 `quiet_hours` JSON，保存前进入高风险二次确认。
- 用户端通知偏好设置首版，原生微信小程序新增“通知偏好”页和首页入口，用户可按订单状态、售后进度和优惠活动配置微信订阅、短信、App Push 开关与静默时间，并调用用户偏好 API 保存。
- 通知 worker 偏好缓存与失败关闭首版，偏好 resolver 按 key 缓存后端偏好，TTL 过期后刷新，后端短时不可读时使用 stale 偏好；无缓存时仍把外部 provider 投递转为 queued。
- 通知偏好变更事件与 worker 主动失效首版，保存偏好会生成 `notification.preferences_changed` outbox，relay 默认发布，worker 消费后只失效对应偏好缓存 key。
- 通知偏好批量保存与策略审计首版，运营可一次保存多条偏好策略，后端同事务写入偏好、变更 outbox 和批量审计，Admin Web 已接入高风险二次确认。
- 通知偏好变更审批与灰度应用首版，运营可提交偏好变更申请并固化 `all`、`target_ids` 或 `percentage` rollout，另一名管理员审批或驳回，批准后再按灰度范围手动应用到批量保存路径；应用同事务写入偏好、缓存失效 outbox 和 `admin.notification_preferences.change_applied` 审计，审计记录 applied/skipped 范围，驳回申请不能应用。
- BFF 浏览器 CORS 白名单和 `OPTIONS` 预检。
- 管理端操作审计日志 `/api/admin/audit-logs`。
- 审计日志 PostgreSQL `audit_logs` 规范化表、查询索引、旧快照回填和 `platform_sequences` 行级锁发号。
- 管理端审计中心增强首版，支持 actor/action/target/after/before/limit 筛选、before 游标翻页、保存筛选、详情抽屉、跨模块跳转和脱敏 payload 摘要。
- 管理端审计导出首版，新增 `/api/admin/audit-logs/export`，支持按审计筛选条件导出 CSV，导出行为本身写入 `admin.audit_logs.exported` 审计，BFF 与 Admin Web 已接入。
- 管理端审计留存/告警健康报告首版，新增 `/api/admin/audit-logs/retention-report`，按 7 年留存、180 天热存、完整性抽样、导出事件、关键动作覆盖生成 `ok`/`warning`/`critical` 状态和告警列表，BFF 与 Admin Web 已接入。
- 管理端审计留存告警 outbox 投递首版，新增 `/api/admin/audit-logs/retention-alerts/emit`，按审计健康报告生成 `audit.retention_alerts` 可靠事件并写入 `admin.audit_retention_alerts.emitted` 审计，notification-worker 已订阅。
- 管理端审计 WORM/冷归档请求首版，新增 `/api/admin/audit-logs/archive/request`，按热存窗口筛选冷归档候选，生成可验证 manifest hash 和归档路径，投递 `audit.archive_requested` outbox 事件并写入 `admin.audit_archive.requested` 审计，BFF 与 Admin Web 已接入。
- 管理端审计归档 worker 首版，新增 `services/audit-archive-worker`，可直接领取 `audit.archive_requested` outbox 事件，校验后端 manifest hash，上传 JSONL 归档文件，附带对象锁请求头和审计元数据，回写 `/api/admin/audit-logs/archive/complete` 完成证据，并按成功/失败回写 outbox。
- 管理端审计归档记录查询首版，新增 `/api/admin/audit-logs/archive/records`，可按归档 ID 查询已完成归档的 storage key、manifest hash、content hash、字节数、对象锁模式、保留期、上传时间和 outbox event id。
- 管理端审计归档对象下载校验/回查首版，新增 `POST /api/admin/audit-logs/archive/verify`，可根据已完成归档记录从配置的归档下载地址读取 JSONL，对比 content hash、manifest header、字节数和 manifest entry 数，并写入 `admin.audit_archive.verified` 审计；`security_auditor` 可触发只读回查。
- 管理端审计归档校验历史查询首版，新增 `GET /api/admin/audit-logs/archive/verifications`，可按归档 ID、状态、时间范围和 limit 回看 `admin.audit_archive.verified` 校验台账；BFF、Admin Web 操作目录、Store/PostgreSQL 仓储和 HTTP 回归测试已覆盖。
- 管理端审计归档校验历史可视化面板首版，审计中心已内嵌“归档校验历史”查询区，可按归档 ID、状态和条数查询并展示状态、校验时间、manifest/content hash、字节数、日志条数、匹配摘要和原始详情。
- 管理端审计服务端安全边界首版，新增 `security_auditor` 只读审计角色，并把 audit payload 白名单/敏感字段掩码下沉到 Store 与 PostgreSQL 路径。
- 管理端审计完整性证明首版，审计日志返回 `integrity_algorithm`、`integrity_hash`、`integrity_verified`，本地默认 `sha256:v1`，生产配置 `AUDIT_LOG_SIGNING_SECRET` 后使用 `hmac-sha256:v1` 检测审计字段或白名单 payload 篡改。
- 管理端服务端 RBAC 策略矩阵首版，新增 `super_admin`、`ops_admin`、`finance_admin`、`dispatch_admin`、`support_admin` 等后台角色和服务端 scope，邀约、退款、售后、对象清理、outbox、调度、运营快照和审计入口已按权限边界守护。
- RBAC 权限治理查询与变更申请审计首版，新增 `/api/admin/rbac/policy` 和 `/api/admin/rbac/change-requests`，BFF 与 Admin Web “权限治理”模块已接入；权限变更申请只写入审计并保持 `pending_approval`，当前不会自动修改运行时权限。
- RBAC 权限申请审批/驳回台账首版，新增权限申请列表与 `/api/admin/rbac/change-requests/{id}/review` 审批入口，审批人不能审批自己提交的申请，审批/驳回继续写入审计，运行时生效必须走单独应用动作。
- RBAC 权限变更手动应用首版，新增 `/api/admin/rbac/change-requests/{id}/apply`，只允许已审批申请进入运行时权限矩阵，应用动作写入 `admin.rbac.change_applied` 审计，API 启动时会从应用审计日志重放已应用策略。
- RBAC 权限变更审计回滚首版，新增 `/api/admin/rbac/change-requests/{id}/rollback`，只允许目标角色最新且当前仍已应用的申请按应用前 scopes 回滚，回滚动作写入 `admin.rbac.change_rolled_back` 审计，API 启动时会按应用/回滚审计时间顺序重放运行时策略。
- 退款策略配置、管理端订单退款、售后审核、订单状态补偿、对象清理完成/失败、outbox 运维与商户/骑手邀约审计同事务首版，`PUT /api/admin/refund-settings` 通过 `SaveRefundSettingsWithAudit` 同事务更新退款策略并写入审计，`POST /api/orders/{orderID}/refund` 通过 `RefundOrderWithAudit` 同事务写入退款业务账本和 `admin.order.refunded` 审计，`POST /api/after-sales/{requestID}/review` 通过 `ReviewAfterSalesWithAudit` 同事务写入售后审核、必要退款和 `after_sales.reviewed` 审计，`POST /api/admin/orders/{orderID}/state/compensate` 通过 `CompensateOrderStateWithAudit` 同事务写入状态补偿结果和 `admin.order_state.compensated` 审计，`POST /api/admin/object-storage/cleanup-complete` 与 `POST /api/admin/object-storage/cleanup-failed` 分别通过 `CompleteObjectStorageCleanupWithAudit`、`RecordObjectStorageCleanupFailureWithAudit` 同事务写入对象清理结果和审计，outbox claim/lease renew/publish/fail/replay/batch replay 分别通过 `ClaimOutboxEventsWithAudit`、`RenewOutboxEventLeaseWithAudit`、`MarkOutboxEventPublishedWithAudit`、`MarkOutboxEventFailedWithAudit`、`ReplayOutboxEventWithAudit`、`ReplayOutboxEventsWithAudit` 同事务更新 `platform_outbox_events` 和 `audit_logs`；商户/骑手/站长邀约创建通过 `CreateMerchantInviteWithAudit`、`CreateRiderInviteWithAudit` 同事务写入最终邀约和审计。

当前最重要的未完成项：

- 字段级/租户级 RBAC、权限变更产品化审批队列、生产 WORM bucket 对象锁策略、归档校验生产演练回查、保留期删除审批、真实告警渠道投递、KMS/链式不可抵赖签名和审计策略治理。
- 商户资质审核通知仍未接真实短信、企业微信、微信小程序/商户端 push 生产账号、模板审批、真实 provider 字段映射联调、真实渠道联调和跨端消息中心；站内通知账本、投递回执账本、Admin Web 通知运营页、失败回执可靠告警、失败重试编排、provider 执行器骨架、provider 回调验签入账、模板 payload 规范、偏好过滤、静默窗口、静默 queued 再投递调度、静默到期自动扫描调度、商户端通知偏好设置、管理端通知偏好操作入口、用户端通知偏好设置、worker 偏好缓存失败关闭、偏好变更主动失效、通知偏好批量保存和通知偏好变更审批/灰度应用已有首版。
- 剩余关键业务写操作与审计写入同事务强制提交，继续扫描后台配置、运营处置、资金和风控写路径。
- 真实微信支付、微信原路退款、对账、提现、商户结算和骑手收入。
- 真实 IM、客服工作台、RTC 信令与通话审计。
- 生产高可用基础设施和 10 万在线压测/容灾报告。

下一批计划：

- 第一批补后台审计中心生产 WORM bucket 保留策略、归档校验生产演练回查、保留期删除审批、真实告警渠道投递、RBAC 产品化审批页、字段级/租户级权限和菜单按权限隐藏。
- 第二批继续补管理端后端明细接口和更完整的审核辅助信息；P0 行详情面板、高风险操作二次确认与结果追踪、失败回放入口、P0 业务筛选分页、售后审核表单、订单退款表单、Outbox 单事件恢复表单、Outbox 发布/失败人工处置表单、Outbox 领取/续租表单、Outbox 死信分诊/解封表单、Outbox 单事件事故辅助明细、商户资质审核、通知运营页和通知失败回执告警首版已完成。
- 第三批补真实资金链路。
- 第四批补 IM 与 RTC。
- 第五批做 10 万在线容量和容灾验收。

## 1. 架构现状

架构名称：自建/混合云 Kubernetes 上的模块化核心 API + 事件驱动 Worker + 多端 BFF + 实时网关架构。

当前架构原则已经写入 `PLATFORM_MASTER_PLAN.md`：

- `api-go` 是模块化核心 API，订单、支付、钱包、抢单、派单等关键链路优先保持强一致事务边界。
- BFF 面向原生微信小程序、商户 Flutter、骑手 Flutter、管理端 Web/Flutter 做聚合与端侧差异适配。
- Worker 负责 outbox relay、调度、支付、通知、集成、结算、对象扫描、对象生命周期清理、审计归档等异步任务。
- `realtime-gateway` 独立承载 WebSocket-only 实时能力，但目前还没有完成完整 IM 落库、离线补偿和 RTC 信令闭环。
- 数据侧以 PostgreSQL 为主库，Redis/Kafka/MinIO/Vault/Prometheus/Grafana/Loki/Tempo/OpenTelemetry 已进入规划和部分部署骨架。
- 10 万在线只是目标架构口径，未完成压测与容灾报告前不得宣称已支撑 10 万同时在线。

## 2. 仓库结构

当前 monorepo 已建立：

- `apps/user-wechat-miniprogram`：用户端原生微信小程序。
- `apps/user-wechat-miniprogram` 继续使用微信原生语言栈 `WXML/WXSS/TypeScript`。
- `apps/merchant-flutter`：商户端 Flutter/Dart。
- `apps/rider-flutter`：骑手端 Flutter/Dart。
- `apps/admin-web`：桌面管理端 Web，已完成最小运营控制台首版。
- `apps/admin-flutter`：移动管理端 Flutter/Dart，目前是骨架。
- `services/api-go`：核心业务 API。
- `services/bff`：多端 BFF。
- `services/realtime-gateway`：实时网关骨架。
- `services/dispatch-worker`：调度 worker 规则骨架。
- `services/payment-worker`：支付 worker 事件规范化骨架。
- `services/notification-worker`：通知 worker 骨架。
- `services/integration-worker`：第三方 OAuth/API worker 骨架。
- `services/settlement-worker`：结算 worker 骨架。
- `services/outbox-relay-worker`：outbox relay worker，可运行化首版。
- `services/object-scan-worker`：对象扫描 worker，ClamAV 下载扫描首版。
- `services/object-lifecycle-worker`：对象生命周期清理 worker。
- `services/audit-archive-worker`：审计归档 worker，WORM/冷归档上传首版。
- `packages/contracts`：枚举、接口响应规范。
- `packages/domain-core`：核心领域纯规则。
- `packages/client-sdk`：请求与鉴权工具。
- `packages/design-tokens`：品牌颜色、logo、tokens。
- `packages/admin-core`：管理端资源与操作模块定义。
- `infra/db/migrations`：PostgreSQL 迁移。
- `infra/docker`：本地开发依赖 compose。
- `infra/k8s/base`：Kubernetes 部署骨架。
- `infra/loadtest`：10 万在线压测计划骨架。
- `infra/observability`：观测说明。
- `docs`：旧版复用、架构、容量容灾、商业化验收和产品对标文档。

## 3. 已完成的产品与工程能力

### 3.1 文档和治理

- 已完成平台总计划：`PLATFORM_MASTER_PLAN.md`。
- 已完成旧版复用审计：`docs/LEGACY_REUSE_AUDIT.md`。
- 已完成美团/旧版能力对标矩阵：`docs/product/meituan-legacy-parity-matrix.md`。
- 已完成商业级验收清单：`docs/product/commercial-readiness-checklist.md`。
- 已完成容量与容灾规划：`docs/operations/capacity-and-dr.md`。
- 已持续记录执行台账：`EXECUTION_LEDGER.md`，当前记录到 `DONE-20260525-137`。
- 已完成 GitHub 协作与质量门禁首版：`verify.yml` 会在 `push`/`pull_request` 跑 `npm run verify` 和 uncached Go 测试；PR 模板要求商业影响、验证和回滚说明；Issue 模板区分 bug、feature、commercial readiness gap；CODEOWNERS 和 Dependabot 已建立。

### 3.2 品牌和 UI 基线

- 已复制旧版 logo 到 `assets/brand/` 和用户小程序品牌资产目录。
- 已固定主色 `#009bf5`，并通过 `packages/design-tokens` 输出 tokens。
- 用户端页面视觉按旧版浅蓝风格推进。
- 已生成用户小程序预览图：`docs/product/user-miniprogram-preview.png`。

### 3.3 用户端原生微信小程序

已完成首批页面与 API 接入点：

- 首页。
- 圈子/小微墙入口。
- 找饭搭页面和前置信息要求。
- 商家列表。
- 店铺详情。
- 团购套餐购买入口。
- 购物车。
- 确认订单。
- 订单列表。
- 订单详情。
- 地址预览页。
- 余额支付密码页面。
- 小程序 API 客户端。
- 微信登录、店铺/商品/团购/购物车/结算/订单查询/支付密码/微信支付预下单接入点。
- 默认访问 BFF，不直接访问内部 Worker 或数据库。

未完成：

- 真机完整首单验收。
- 真实 `wx.login` 页面流程和用户授权体验。
- 定位、搜索、评价、售后入口、邀请页、钱包账单、红包、官方群/商户群、客服消息、买药、跑腿完整页面。

### 3.4 商户端 Flutter

已完成首批页面和 API 工具：

- 经营概况页面。
- 订单处理页面。
- 商品管理页面。
- 团购核销页面。
- 资质资料页面。
- 商户 API 客户端。
- 商户邀请注册设置密码。
- 商户账号密码登录。
- 资质上传。
- 员工健康证。
- 补充资料。
- 商户保证金查询和缴纳。
- 商户订单列表。
- 接单、出餐。
- 商品列表、创建、编辑、上下架、售罄。
- 团购券扫码核销入口。

未完成：

- 完整登录/邀请注册 UI。
- 店铺装修和店铺展示页管理。
- 商户钱包、结算、消息页。
- 资质过期强弹窗。
- 商户自发券、平台活动券确认参与、商户承担活动结算审计。

### 3.5 骑手端 Flutter

已完成首批页面和 API 工具：

- 抢单大厅首屏。
- 骑手上线/离线。
- 保证金缴纳。
- 微信免押申请入口。
- 退押申请入口。
- 取货。
- 送达完成。
- 每日一次免责取消。
- 拒绝当前派单。
- 站长工作台首屏。
- 站点骑手列表。
- 骑手等级、接单耗时。
- 待调度订单列表。
- 站长手动派单。
- 每日任务时长和固定单量配置。
- 站长创建骑手邀约。
- 骑手接受邀约带密码。
- 骑手/站长账号密码登录 API 调用入口。

未完成：

- 完整任务详情、地图导航、轨迹、收入、钱包、账单、提现。
- 健康证、保险、违规申诉、骑手之家。
- 真机定位和后台持续定位验证。
- 与实时网关的派单弹窗、位置和消息联动。

### 3.6 管理端

已完成：

- `apps/admin-web` 最小运营控制台首版：静态入口、运营导航、P0 指标位、今日必盯队列、模块状态、RBAC 首版矩阵和接口操作台。
- `apps/admin-web` 接口操作台已接入管理员登录、商户/站长/骑手邀约、退款策略、售后列表、对象清理、outbox 运维和订单状态补偿等现有 BFF/API。
- `apps/admin-web` 已补 P0 业务视图首版：订单监控、售后审核、商户资质、骑手/站长、骑手绩效、派单审计、退款策略均有独立页面结构、指标、表格、操作入口和安全约束。
- 管理端已新增 `/api/admin/operations/snapshot` 运营快照首版，按后台视角聚合订单、商户资质/保证金、骑手/站长、骑手绩效、售后、派单审计、退款策略、outbox 健康和对象清理统计；BFF 与管理端操作台已接入该入口。
- 管理端 P0 视图已开始绑定运营快照：有管理员 token 时可刷新快照，订单、售后、商户、骑手、骑手绩效、派单、退款策略和顶部 KPI 会按快照数据生成；快照字段渲染已做 HTML 转义。
- 管理端已新增 `/api/admin/audit-logs` 操作审计首版，商户/骑手邀约、退款策略、订单退款、状态补偿、售后审核、对象清理和 outbox 运维等关键写操作会写入可查询审计账本；PostgreSQL-backed Store 下查询直接走规范化 `audit_logs` 表；BFF 与 Admin Web 操作台已接入。
- 管理端已新增审计中心增强首版，支持按操作者、动作、目标、after/before 时间范围和条数查询审计记录；页面以白名单摘要展示 payload，并对 password、secret、token、authorization、openid、phone、object key、签名等敏感字段做脱敏；常用筛选可保存，审计详情可在抽屉查看，并可按目标继续筛选或跳到相关运营模块。
- 管理端已新增审计导出首版：`GET /api/admin/audit-logs/export` 使用同一套 actor/action/target/after/before/limit 筛选条件导出 CSV，返回 filename、row_count、generated_at 和 CSV 内容；导出动作写入 `admin.audit_logs.exported` 审计，记录导出筛选、行数和格式，BFF 与 Admin Web 操作目录已接入。
- 管理端已新增审计留存/告警健康报告首版：`GET /api/admin/audit-logs/retention-report` 使用默认 2555 天留存、180 天热存和 500 条完整性抽样，统计总日志数、最早/最新时间、过期日志、冷归档候选、完整性失败、导出事件和关键审计动作覆盖，返回 `ok`、`warning` 或 `critical` 状态与告警列表；PostgreSQL-backed Store 使用规范化 `audit_logs` 聚合查询，BFF 与 Admin Web 操作目录已接入。
- 管理端已新增审计留存告警 outbox 投递首版：`POST /api/admin/audit-logs/retention-alerts/emit` 复用留存报告口径，把 critical/warning 告警投递为 `audit.retention_alerts` outbox 事件，并写入 `admin.audit_retention_alerts.emitted` 审计；新增 `audit:write` scope，`security_auditor` 仍保持只读，notification-worker 已订阅该 topic。
- 管理端已新增审计 WORM/冷归档请求首版：`POST /api/admin/audit-logs/archive/request` 使用热存窗口和 limit 筛选冷归档候选，生成 `sha256:v1` manifest hash、归档路径和 manifest entries，投递 `audit.archive_requested` outbox 事件并写入 `admin.audit_archive.requested` 审计；新增请求只允许 `audit:write`，`security_auditor` 保持只读。
- 管理端已新增审计归档 worker 首版：`services/audit-archive-worker` 直接领取 `audit.archive_requested` outbox 事件，按后端 hash 合约校验 manifest，上传 `application/x-ndjson` 归档文件，附带 `x-amz-object-lock-mode`、保留期和审计 hash 元数据，成功调用 `/api/admin/audit-logs/archive/complete` 记录完成证据，再标记 published；失败标记 failed 并进入 retry/dead-letter 策略。通用 `outbox-relay-worker` 不再 relay `audit.archive_requested`，避免归档文件落盘前被错误标记完成。
- 管理端已新增审计归档完成记录查询首版：`GET /api/admin/audit-logs/archive/records` 从 `admin.audit_archive.completed` 审计账本重建归档完成记录，Admin Web 操作目录和 BFF 均已接入，可查 storage key、manifest hash、content hash、bytes、object lock、retain until、uploaded at 和 outbox event id。
- 管理端已新增审计归档对象下载校验/回查首版：`POST /api/admin/audit-logs/archive/verify` 根据归档完成记录读取配置的归档下载地址，校验 JSONL 文件 content hash、manifest header、字节数和 manifest entry 数，校验结果写入 `admin.audit_archive.verified` 审计，`security_auditor` 可触发该只读回查。
- 管理端已新增审计归档校验历史查询首版：`GET /api/admin/audit-logs/archive/verifications` 从 `admin.audit_archive.verified` 审计账本重建校验历史，支持按归档 ID、状态、after/before 和 limit 查询，BFF 与 Admin Web 操作目录已接入。
- 管理端已新增审计归档校验历史可视化面板首版：审计检索页内可直接查询和展示归档校验历史，包含状态、校验时间、manifest/content hash、bytes/logs、匹配摘要和原始详情。
- 管理端已新增审计服务端安全边界首版，`security_auditor` 可只读审计账本但不能执行后台写操作；`auth_sessions` 与身份迁移允许该主体类型；审计 payload 在服务端白名单过滤后才写入或返回，`object_key` 等敏感允许字段会被掩码，password、token、phone、nested/raw_request 等非白名单或敏感字段会被丢弃。
- 管理端已新增审计完整性证明首版，`audit_logs` 表和 API 返回 `integrity_algorithm`、`integrity_hash`、`integrity_verified`；内存 Store 与 PostgreSQL 写入会签封规范化审计字段和服务端白名单 payload，查询时验证是否被篡改；Admin Web 审计中心可展示完整性状态、算法和哈希。
- 管理端已新增退款策略配置、管理端订单退款、售后审核、订单状态补偿、对象清理完成/失败、outbox 运维与商户/骑手邀约审计同事务首版，HTTP 退款策略保存入口改走 `SaveRefundSettingsWithAudit`，管理端订单退款入口改走 `RefundOrderWithAudit`，售后审核入口改走 `ReviewAfterSalesWithAudit`，订单状态补偿入口改走 `CompensateOrderStateWithAudit`，对象清理完成/失败入口改走 `CompleteObjectStorageCleanupWithAudit` 与 `RecordObjectStorageCleanupFailureWithAudit`，outbox 运维入口改走 `ClaimOutboxEventsWithAudit`、`RenewOutboxEventLeaseWithAudit`、`MarkOutboxEventPublishedWithAudit`、`MarkOutboxEventFailedWithAudit`、`ReplayOutboxEventWithAudit` 和 `ReplayOutboxEventsWithAudit`，商户/骑手邀约入口改走 `CreateMerchantInviteWithAudit` 与 `CreateRiderInviteWithAudit`；PostgreSQL-backed Store 分别使用单个数据库事务同时写入业务表、`platform_outbox_events` 或邀约快照与 `audit_logs`，并由 HTTP 防回退测试、Store 原子审计测试和架构守卫固定路径。
- 管理端已新增 Outbox 单事件恢复表单首版：运营首页和 Outbox 队列详情抽屉可预填事件 ID，进入 `POST /api/admin/outbox/events/{eventID}/replay` 前必须二次确认，继续复用后端 `ReplayOutboxEventWithAudit` 原子审计路径。
- 管理端已新增 Outbox 发布/失败人工处置表单首版：运营首页和 Outbox 队列详情抽屉可预填失败原因、重试延迟、最大尝试次数和事件 ID，进入 `POST /api/admin/outbox/events/{eventID}/failed` 或 `POST /api/admin/outbox/events/{eventID}/published` 前必须二次确认，继续复用后端 `MarkOutboxEventFailedWithAudit` 与 `MarkOutboxEventPublishedWithAudit` 原子审计路径。
- 管理端已新增 Outbox 领取/续租表单首版：运营首页和 Outbox 队列详情抽屉可预填 topic、limit、lease owner、lease seconds 和事件 ID，进入 `POST /api/admin/outbox/events/claim` 或 `POST /api/admin/outbox/events/{eventID}/lease/renew` 前必须二次确认，继续复用后端 `ClaimOutboxEventsWithAudit` 与 `RenewOutboxEventLeaseWithAudit` 原子审计路径。
- 管理端已新增 Outbox 死信分诊/解封表单首版：运营首页和 Outbox 队列详情抽屉可预填 `status=dead_letter` 查询和死信事件 ID，进入 `POST /api/admin/outbox/events/{eventID}/replay` 解封前必须二次确认，继续复用后端 `ReplayOutboxEventWithAudit` 原子审计路径。
- 管理端已新增 Outbox 单事件事故辅助明细首版：`GET /api/admin/outbox/events/{eventID}` 由 `outbox:read` 守护，返回事件 ready/blocked/lease 状态、retry/lease 剩余秒数、payload 摘要、关联业务目标、最近 outbox 审计、推荐操作和人工核查清单；BFF 与 Admin Web `outbox-event-detail` 已接入。
- 管理端已新增商户资质审核后端与表单首版：商户上传资质默认进入 `pending_review`，后台 `ops_admin` 通过 `POST /api/admin/merchant-qualifications/{qualificationID}/review` 审核后才会把资质计入接单资格，审核结果写入 `admin.merchant_qualification.reviewed` 审计；BFF 与 Admin Web 商户模块/详情抽屉已接入并进入高风险二次确认。
- 管理端已新增商户资质待审列表与明细接口首版：`GET /api/admin/merchant-qualifications` 和 `GET /api/admin/merchant-qualifications/{qualificationID}` 复用 `merchant:qualification_review` 权限，可返回资质状态、商户/店铺/保证金上下文、接单资格、最近审核审计、推荐动作和审核核查清单；BFF 与 Admin Web `merchant-qualifications` / `merchant-qualification-detail` 已接入。
- 管理端已新增商户资质审核结果可靠通知首版：`ReviewMerchantQualificationWithAudit` 会在审核、审计同一事务链路中生成 `merchant.qualification_reviewed` outbox 事件，notification-worker 订阅后可生成面向商户的审核结果通知 payload，outbox relay 默认 topic 与 Compose/K8s 配置已覆盖。
- 商户端已新增站内通知中心首版：平台通知账本支持幂等写入、按商户/状态查询和已读标记；PostgreSQL-backed Store 会确保 `platform_notifications` 表、幂等键唯一约束和目标/状态/时间索引；BFF 已代理商户通知列表和已读路径，notification-worker 可把可靠通知写入 `/api/notifications`。
- 管理端已新增通知运营查询接口首版：`GET /api/admin/notifications` 支持 target/status/source 过滤，`ops_admin` 可读写通知账本，`support_admin` 只读排查，`security_auditor` 不可读运营通知。
- 管理端已新增通知投递回执台账首版：`POST /api/notifications/{notificationID}/deliveries` 可记录站内信或外部渠道的 delivered/failed 回执，`GET /api/admin/notification-deliveries` 可按商户、通知、渠道、provider 和状态查询失败原因。
- Admin Web 已新增通知运营页首版：`notifications` 模块状态为 wired，页面动作接入通知台账、通知回执和补录通知回执；运营可在详情抽屉里按通知 ID、商户目标、来源 topic 和失败状态预填查询，补录回执纳入高风险二次确认。
- 管理端已新增通知失败回执告警首版：`POST /api/admin/notification-deliveries/failure-alerts/emit` 由 `notification:write` 守护，按目标、渠道、provider 和 limit 汇总 failed 回执，生成 `notification.delivery_failed_alerts` outbox 事件并写入 `admin.notification_delivery_failure_alerts.emitted` 审计；BFF、Admin Web、notification-worker、outbox relay 和部署骨架已接入。
- 管理端已新增通知失败重试编排首版：`POST /api/admin/notification-deliveries/retries/schedule` 由 `notification:write` 守护，按目标、渠道、provider、limit 和 `retry_after_seconds` 汇总 failed 回执并生成 `notification.delivery_retries` outbox 事件；事件 `available_at` 对齐 `retry_at`，在 provider 退避窗口后进入 ready 队列，操作写入 `admin.notification_delivery_retries.scheduled` 审计；BFF、Admin Web、notification-worker、outbox relay 和部署骨架已接入。
- 通知 worker 已新增 provider 执行器首版：`notification.delivery_retries` outbox payload 会携带原始通知快照，worker 可按渠道生成 provider dispatch、调用配置 endpoint/adapter，并把真实 delivered/failed 结果写回 `/api/notifications/{notificationID}/deliveries`；未配置 provider 时会记录 `provider_not_configured` 失败回执，避免外部渠道状态被误标为成功。
- 通知 provider 回调验签入账首版：外部渠道可回调 `POST /api/notifications/provider-callback`；生产配置 `NOTIFICATION_PROVIDER_CALLBACK_SECRET` 后，API 会按 notification/channel/provider/status/provider_message_id/error/idempotency/timestamp canonical lines 做 HMAC-SHA256 验签，再把 delivered/failed/queued 回执幂等写入 `platform_notification_deliveries`；BFF、notification-worker 签名工具、Docker Compose 和 K8s secret 位已覆盖。
- 通知 provider 模板映射与渠道 payload 规范首版：`notification-worker` 可解析 `NOTIFICATION_PROVIDER_TEMPLATES`，按 notification type/template_key 和 channel 选择模板，生成 `template_id`、`template_params` 和面向短信、微信订阅消息、企业微信、push 的 provider payload；Docker Compose 和 K8s 已预留模板配置位。
- 通知偏好与静默窗口首版：`notification-worker` 可解析 `NOTIFICATION_DELIVERY_PREFERENCES`，按 default、target_role、target_role:target_id、type 和 target_role:target_id:type 合并偏好规则；禁用渠道或静默窗口内的外部 provider 投递不会调用 provider endpoint，而是记录 `queued` 回执和 `notification_preference_disabled`/`notification_quiet_window` 原因，避免误触达又保留运营证据。
- 通知偏好后端账本与 API 首版：`platform_notification_preferences` 持久化目标/类型级偏好，支持 `enabled_channels`、`disabled_channels` 和 `quiet_hours`；商户可通过 `/api/merchant/notification-preferences` 维护自身通知类型偏好，运营可通过 `/api/admin/notification-preferences` 查询/写入并生成 `admin.notification_preferences.saved` 审计；BFF 与 HTTP/Store 测试已覆盖。
- 通知 worker 后端偏好读取首版：`notification-worker` 新增 `createNotificationPreferenceResolver` 和 `deliveryPreferencesFromRecords`，按 `default`、角色、目标、类型和目标+类型 `preference_key` 精确读取 `/api/admin/notification-preferences`；后端偏好与静态环境偏好合并后进入 provider 投递决策，读取失败时外部投递转 `queued` 并写入 `notification_preference_lookup_failed`。
- 通知静默 queued 再投递调度首版：通知重试调度请求新增 `status`、`error_code` 和 `retry_at`，内存 Store 与 PostgreSQL-backed Store 可筛 `queued` + `notification_quiet_window` 回执并生成 `notification.delivery_retries` 延迟 outbox；Admin Web 通知回执筛选支持错误码，重试表单支持 queued 与指定重试时间。
- 通知静默到期自动扫描调度首版：静默窗口 queued 回执记录 `retry_at`，后台可按 `retry_at_before` 查询到期回执，`/api/admin/notification-deliveries/quiet-window-retries/schedule` 可批量调度到期 quiet-window 回执，worker 可通过 `NOTIFICATION_QUIET_RETRY_AUTO_SCHEDULE` 开启周期扫描。
- 商户端通知偏好设置首版：商户端 Flutter 新增 `lib/features/notifications/merchant_notification_preferences_page.dart`，可读取/保存 `order.status_changed` 和 `merchant.qualification_reviewed` 的外部渠道开关、静默时间、静默渠道；首页新增通知入口，API client 接入 `GET/PUT /api/merchant/notification-preferences`。
- 管理端通知偏好操作入口首版：Admin Web 通知运营模块新增 `notification-preferences` 和 `notification-preference-save` 操作，支持读取/保存目标级通知偏好；保存表单解析 `enabled_channels`、`disabled_channels` 和 `quiet_hours` JSON，并纳入高风险二次确认、结果追踪、详情抽屉预填和架构守卫。
- 用户端通知偏好设置首版：原生微信小程序新增 `pages/notification-preferences/index`，可读取/保存 `order.status_changed`、`after_sales.updated` 和 `coupon.campaign` 的外部渠道开关、静默时间、静默渠道；首页新增通知偏好入口，API client 接入 `GET/PUT /api/user/notification-preferences`，服务端强制按当前用户写入 `target_role=user`。
- 通知 worker 偏好缓存与失败关闭首版：`notification-worker` 可按 preference key 短 TTL 缓存后端偏好，TTL 过期刷新；刷新失败且仍在 stale 窗口内继续使用旧偏好，没有缓存时保持 `notification_preference_lookup_failed` queued 失败关闭。
- 通知偏好变更审批与灰度应用首版：`POST /api/admin/notification-preferences/change-requests` 可提交待审批偏好变更申请并固化 `all`、`target_ids` 或 `percentage` rollout，`GET /api/admin/notification-preferences/change-requests` 可按状态读取台账，`POST /api/admin/notification-preferences/change-requests/{id}/review` 支持另一名管理员审批/驳回且禁止自审，`POST /api/admin/notification-preferences/change-requests/{id}/apply` 只允许已审批申请按灰度范围进入批量保存路径并写入 `admin.notification_preferences.change_applied` 审计；BFF、Admin Web 操作目录和 HTTP/BFF/Admin Web/架构守卫测试已覆盖。
- 管理端已新增服务端 RBAC 策略矩阵首版，后台角色包含兼容 `admin`、`super_admin`、`ops_admin`、`finance_admin`、`dispatch_admin`、`support_admin` 和 `security_auditor`；邀约、退款、运营快照、售后、对象清理、outbox、订单状态补偿、派单读写和审计读取已改为服务端 scope 判断，Admin Web RBAC 配置与后端 scope 命名保持一致。
- 管理端已新增 RBAC 权限治理查询与变更申请审计首版：`GET /api/admin/rbac/policy` 返回服务端真实 RBAC 矩阵、策略版本和当前角色能力；`POST /api/admin/rbac/change-requests` 仅允许 `admin`/`super_admin` 提交变更申请，申请写入 `admin.rbac.change_requested` 审计并保持待审批；BFF 与 Admin Web “权限治理”模块已接入。
- 管理端已新增 RBAC 权限申请审批/驳回台账首版：`GET /api/admin/rbac/change-requests` 可从审计账本重建申请状态，`POST /api/admin/rbac/change-requests/{id}/review` 可审批或驳回，禁止同一管理员自审，审批结果写入 `admin.rbac.change_reviewed` 审计。
- 管理端已新增 RBAC 权限变更手动应用首版：`POST /api/admin/rbac/change-requests/{id}/apply` 只允许已审批申请应用到运行时权限矩阵，禁止提交人直接应用自己的申请，应用动作写入 `admin.rbac.change_applied` 审计；`api-go` 启动时会从 `admin.rbac.change_applied` 审计日志恢复已应用策略。
- 管理端已新增 RBAC 权限变更审计回滚首版：`POST /api/admin/rbac/change-requests/{id}/rollback` 只允许目标角色最新且当前仍处于已应用状态的申请回滚到应用前 scopes，禁止申请人直接回滚自己的申请，回滚动作写入 `admin.rbac.change_rolled_back` 审计；`api-go` 启动时会按应用/回滚审计时间顺序恢复运行时策略。
- BFF 已补浏览器来源 CORS 白名单和 `OPTIONS` 预检处理，默认覆盖本地管理端/Flutter 调试来源，并可通过 `BFF_ALLOWED_ORIGINS` 配置部署来源。
- `apps/admin-flutter` 骨架。
- `packages/admin-core` 已定义关键运营模块。
- 核心后台 API 已覆盖很多运营动作：退款策略、售后、outbox、对象存储清理、派单事件、站长任务、骑手绩效等。

未完成：

- 桌面管理端完整业务页面和详情页。
- 管理端 P0 视图已能读取运营快照生成首批表格/指标，关键写操作已有审计账本、审计检索页、审计 CSV 导出、审计留存/告警健康报告、审计留存告警 outbox 投递、审计 WORM/冷归档请求、审计归档 worker、归档完成记录查询、归档对象下载校验/回查、归档校验历史查询和可视化面板、服务端 RBAC 策略矩阵、RBAC 查询/变更申请审计、审批/驳回台账、手动应用、审计回滚、完整性证明、P0 行详情面板、P0 业务筛选分页、售后审核表单、订单退款表单、Outbox 单事件恢复表单、Outbox 发布/失败人工处置表单、Outbox 领取/续租表单、Outbox 死信分诊/解封表单、Outbox 单事件事故辅助明细和商户资质审核首版；仍需补更多后端业务明细接口、字段级/租户级权限、生产 WORM bucket 保留策略、归档校验生产演练回查、保留期删除审批、真实告警渠道投递、KMS/链式不可抵赖签名。
- 移动管理端实际页面。
- 字段级/租户级 RBAC、权限变更产品化审批页面、生产 WORM bucket 保留策略、归档校验生产演练回查、保留期删除审批、真实告警渠道投递、KMS/链式不可抵赖签名和审计策略治理。
- 订单、售后、用户、商户、骑手、首页卡片、优惠券、圈子/饭搭、团购、买药、跑腿、客服、RTC、OAuth/API、对象存储告警等后台面板。

### 3.7 核心 API 和数据链路

`services/api-go` 已完成大量核心闭环：

- 用户微信登录签名 token 骨架。
- 真实微信 `code2session` provider resolver。
- auth session 持久化和 logout 撤销。
- 生产默认关闭开发 token。
- 商户邀约注册、资质、员工健康证、补充资料。
- 商户主体登录。
- 管理员 bootstrap 登录。
- 骑手/站长邀约注册。
- 骑手/站长主体登录。
- 商户资料、订单列表、接单、出餐。
- 商户商品管理。
- 商户/骑手保证金。
- 店铺接单门槛。
- 团购下单发券和扫码核销。
- 骑手在线状态。
- 10 分钟后自动派单。
- 拒单顺延派单。
- 派单确认超时自动转派。
- 订单状态机补偿首版。
- 平台 outbox 事件首版。
- outbox 手动恢复、批量恢复、死信隔离、租约领取、租约续租、租约健康观测。
- 消费端幂等落库首版。
- 支付/钱包 PostgreSQL 规范化恢复首版。
- 订单创建 PostgreSQL 事务化首版。
- 购物车结算 PostgreSQL 事务化首版。
- 余额支付 PostgreSQL 事务扣减首版。
- 退款策略与余额退款核心闭环首版。
- payment-worker 原路退款事件规范化首版。
- 退款 PostgreSQL 事务化首版。
- 售后申请与审核核心闭环首版。
- 售后 PostgreSQL 规范化恢复首版。
- 售后审核 PostgreSQL 事务化首版。
- 售后部分退款资金账本首版。
- 售后仲裁与客服介入处理日志首版。
- 售后可退金额与证据上传票据首版。
- 售后证据确认与附件元数据首版。
- 对象存储上传签名配置化首版。
- 售后上传票据账本与确认防伪首版。
- 售后对象存在性 HEAD 校验开关首版。
- 售后上传回调验签与扫描门禁首版。
- 对象扫描 worker 首版。
- 对象扫描 worker ClamAV 适配与下载首版。
- 管理端运营快照 API 首版：订单、商户、骑手、售后、调度、退款策略、outbox 和对象清理统一聚合。
- 管理端运营快照绑定首版：P0 表格、KPI 和队列可从 `/api/admin/operations/snapshot` 响应生成，并对展示字段做 HTML 转义。
- 管理端操作审计日志首版：关键后台写操作记录 actor、action、target、request_id、ip_hash、服务端白名单 payload 和创建时间，管理员可按条件查询。
- 审计日志 PostgreSQL 规范化表首版：`PostgresStore` 确保 `audit_logs` 表/索引存在，把旧快照审计幂等补入表，通过 `platform_sequences` 行级锁生成审计 `aud_N`，PostgreSQL 查询路径直接读取规范化表。
- 管理端审计完整性证明首版：审计日志签封规范化字段和白名单 payload，返回 `integrity_algorithm`、`integrity_hash`、`integrity_verified`，本地默认 SHA256，生产可用 `AUDIT_LOG_SIGNING_SECRET` 启用 HMAC。
- 退款策略配置、管理端订单退款、售后审核、订单状态补偿、对象清理完成/失败、outbox 运维与商户/骑手邀约审计同事务首版：`SaveRefundSettingsWithAudit` 将退款策略配置写入和 `admin.refund_settings.updated` 审计写入收敛到仓储级原子路径；`RefundOrderWithAudit` 将管理端订单退款业务账本和 `admin.order.refunded` 审计写入收敛到仓储级原子路径；`ReviewAfterSalesWithAudit` 将售后审核、必要退款和 `after_sales.reviewed` 审计写入收敛到仓储级原子路径；`CompensateOrderStateWithAudit` 将订单状态补偿和 `admin.order_state.compensated` 审计写入收敛到仓储级原子路径；`CompleteObjectStorageCleanupWithAudit` 与 `RecordObjectStorageCleanupFailureWithAudit` 将对象清理完成/失败票据状态和 `admin.object_cleanup.completed`/`admin.object_cleanup.failed` 审计写入收敛到仓储级原子路径；outbox claim/lease renew/publish/fail/replay/batch replay 通过对应 `WithAudit` 仓储方法在同一事务内更新 `platform_outbox_events` 与 `audit_logs`；商户/骑手/站长邀约通过 `CreateMerchantInviteWithAudit` 与 `CreateRiderInviteWithAudit` 把最终生成 token 的邀约和审计写入收敛到仓储级原子路径。
- 管理端服务端 RBAC 策略矩阵与权限治理查询/申请/审批/应用/回滚审计首版：新增后台角色和 scope 常量，核心后台路由已按 `CanManageInvites`、`CanManageRefunds`、`CanReadAdminAfterSales`、`CanManageDispatch`、`CanReadOutbox`、`CanManageOutbox` 等服务端策略守护，`GET /api/admin/rbac/policy` 可读取真实矩阵，`GET /api/admin/rbac/change-requests` 可读取申请台账，`POST /api/admin/rbac/change-requests` 可写入待审批申请审计，`POST /api/admin/rbac/change-requests/{id}/review` 可写入审批/驳回审计，`POST /api/admin/rbac/change-requests/{id}/apply` 可把已审批申请应用到运行时权限矩阵并写入应用审计，`POST /api/admin/rbac/change-requests/{id}/rollback` 可把当前已应用申请回滚到应用前 scopes 并写入回滚审计，`auth_sessions` 和迁移允许新增后台主体类型。
- 对象生命周期清理 worker 首版。
- 对象清理失败账本首版。
- 对象清理统计接口首版。
- 派单审计事件 PostgreSQL 规范化恢复首版。
- 商家订单流转 PostgreSQL 事务化首版。
- 骑手取货/送达完成。
- 固定单量完成后免责拒派决策。
- 站长站点骑手/订单视图。
- 站长手动派单。
- 站长任务时长/固定单量配置。
- 站点骑手绩效等级快照。
- 派单事件持久化与查询审计。
- 站点区域匹配首版。
- 店铺、商品、地址、购物车、结算订单、订单列表、订单详情。
- 余额支付、余额支付密码。
- 微信支付预下单/回调验签骨架。
- 骑手抢单。
- 每日一次免责取消。

当前仍未完成：

- 微信支付生产参数完整接入。
- 微信原路退款 API 真调用。
- 微信支付对账和差错处理。
- 钱包提现。
- 商户结算和骑手收入结算真实资金链路。
- 充值外部网页跳转/OAuth/二维码登录闭环。
- 买药完整订单与监管审计。
- 跑腿/快递完整下单、计价、异常处理。
- 用户邀请页和邀请奖励闭环。
- 优惠券、红包、群聊资金闭环的 API 实装。
- 评价、收藏、积分会员、推送、风控完整闭环。
- 字段级/租户级 RBAC、剩余业务写操作与审计写入同事务强制提交、生产 WORM bucket 保留策略、归档校验生产演练回查、保留期删除审批、真实告警渠道投递、KMS/链式签名和完整审计后台。

### 3.8 BFF

已完成代理：

- 运行时配置。
- 首页模块和首页卡片。
- 用户端微信登录。
- 商户邀约和主体登录。
- 管理员 bootstrap 登录。
- 骑手/站长邀约和主体登录。
- 店铺、商品、地址、购物车、结算、订单、钱包、微信支付预下单。
- 退款策略和订单退款。
- 商户订单、商品、保证金。
- 骑手调度、保证金。
- 派单事件查询。
- 派单确认超时转派。
- 订单状态机补偿。
- 管理端操作审计查询。
- 对象存储清理候选、统计、完成、失败代理。
- 站长任务和绩效核心接口代理。
- 浏览器来源 CORS 白名单、`OPTIONS` 预检、`Authorization`/`Content-Type`/`X-Client-Kind` 请求头放行。

未完成：

- BFF 缓存策略。
- 限流策略。
- 多端错误文案适配。
- 灰度发布配置。
- BFF 观测指标。

### 3.9 Worker

已完成：

- `outbox-relay-worker`：可轮询、claim、发布、续租、失败回写、Kafka REST publisher 可选。
- `dispatch-worker`：调度规则测试覆盖最近骑手、10 分钟抢单大厅、拒单跳过、等级优先、固定单量免责。
- `payment-worker`：支付回调和原路退款事件规范化，消费幂等。
- `notification-worker`：订单、消息、审计留存告警、商户资质审核结果、通知失败回执告警和通知失败重试计划 payload 骨架，已可把面向商户或安全目标的站内通知写入 API，记录 in-app delivered 回执，并通过 provider 执行器首版生成短信/企微/订阅消息/push dispatch、按 `NOTIFICATION_PROVIDER_TEMPLATES` 套用 `template_id`/变量/渠道 payload、按 `NOTIFICATION_DELIVERY_PREFERENCES` 执行渠道偏好和静默窗口、调用配置 endpoint/adapter、写回 provider delivered/failed/queued 回执，提供 provider 回调 HMAC 签名 payload 工具，消费幂等；后台可查询通知和回执账本。
- `integration-worker`：OAuth/API provider 配置和同步事件骨架，消费幂等。
- `settlement-worker`：完成订单结算计算骨架，金额整数分。
- `object-scan-worker`：对象上传事件、下载、大小限制、ClamAV INSTREAM、回调签名、消费幂等。
- `object-lifecycle-worker`：清理候选读取、对象删除、404 幂等成功、删除完成回写、删除失败回写。
- `audit-archive-worker`：领取审计归档 outbox、校验 manifest hash、上传 JSONL 归档、附带对象锁请求头、成功/失败回写 outbox，消费幂等。

未完成：

- 真实 Kafka/NATS broker 运维接入和生产拓扑。
- 通知通道真实接入：微信订阅消息、短信、企业微信、Push 的生产账号、模板审批、真实 provider 字段映射联调、真实渠道联调和跨端消息中心；站内通知账本、投递回执账本、Admin Web 通知运营页、失败回执可靠告警、失败重试编排、provider 执行器骨架、provider 回调验签入账、模板 payload 规范、偏好过滤、静默窗口、静默 queued 再投递调度、静默到期自动扫描调度、商户端通知偏好设置、管理端通知偏好操作入口、用户端通知偏好设置、worker 偏好缓存失败关闭、偏好变更主动失效、通知偏好批量保存和通知偏好变更审批/灰度应用已有首版。
- payment-worker 真正调用微信退款/对账 API。
- settlement-worker 真正落库商户结算、骑手收入、平台抽佣。
- integration-worker 真正接入微信以外 provider、地图、短信、第三方 API。
- object-scan/object-lifecycle 的真实 MinIO SDK、STS/Vault 临时凭证、隔离 bucket、告警投递和最小权限 IAM。

### 3.10 实时网关

已完成：

- `realtime-gateway` WebSocket-only 容量守卫和基础健康接口测试。
- 架构计划已定义 IM、订单通知、骑手位置、抢单/派单事件和 RTC 信令由该网关承载。

未完成：

- 用户/商户/骑手/客服 IM 落库。
- 已读、撤回、离线补偿。
- 官方群自动入群和默认免打扰。
- 商户群、券进群门槛。
- 红包消息和余额资金链路。
- 骑手位置上报、轨迹抽样。
- RTC 呼叫、接听、拒绝、取消、挂断、超时、审计。
- 10 万连接压测。

### 3.11 基础设施

已完成：

- PostgreSQL 核心迁移：`0001_core.sql`、`0002_auth_payment.sql`、`0003_platform_store_snapshots.sql`、`0004_platform_outbox.sql`。
- Docker Compose 开发依赖骨架。
- Kubernetes base namespace 和 app stack 骨架。
- outbox relay、object scan、object lifecycle worker 部署骨架。
- 观测说明文档。
- 10 万在线压测计划 JSON 骨架。

未完成：

- PostgreSQL HA + PgBouncer。
- Redis Cluster/Sentinel。
- Kafka 多 Broker。
- MinIO 生产 bucket、隔离策略和权限。
- Vault/KMS 真接入。
- Ingress、CDN、WAF、API Gateway、灰度发布。
- Prometheus/Grafana/Loki/Tempo/OpenTelemetry 实际部署和告警规则。
- CI/CD、镜像构建、版本发布、回滚演练。
- 数据备份恢复和灾备演练报告。

## 4. 当前验证情况

最近一次验证通过：

```bash
npm run verify
cd services/api-go && go test -count=1 ./...
npm run verify:architecture
```

`npm run verify` 覆盖：

- 架构守卫。
- contracts。
- domain-core。
- client-sdk。
- admin-core。
- BFF。
- realtime-gateway。
- dispatch-worker。
- payment-worker。
- notification-worker。
- integration-worker。
- object-scan-worker。
- object-lifecycle-worker。
- settlement-worker。
- outbox-relay-worker。
- api-go 全包测试。

## 5. 明确未完成清单

### 5.1 商业级上线必须补齐

- 真实微信支付生产参数、证书、验签、退款、对账。
- 钱包提现、商户结算、骑手结算。
- 管理端 Web 完整运营后台。
- 移动管理端高频运营处理。
- 用户端真机首单闭环。
- 商户端和骑手端 HBuilderX 真机闭环。
- IM 消息落库、已读、离线补偿。
- RTC 语音通话信令和审计。
- 骑手实时位置、轨迹和地图。
- 买药完整资质和监管审计。
- 快递/跑腿完整任务模型和计价。
- 邀请用户页和奖励闭环。
- 优惠券、群聊、红包、平台补贴、商户承担活动结算。
- 评价、内容审核、举报、风控、隐私合规。
- 后台权限配置 UI、字段级/租户级 RBAC、操作审计归档/真实告警和敏感字段脱敏治理。
- 真实 Kafka/Redis/PostgreSQL HA/MinIO/Vault 生产部署。
- Prometheus/Grafana/Loki/Tempo 告警和仪表盘。
- 10k/30k/60k/100k 在线压测报告。
- 容灾演练报告。

### 5.2 不能提前宣称完成的事项

- 不能说已经支撑 10 万在线。
- 不能说已经商业级上线。
- 不能说微信支付生产闭环完成。
- 不能说实时 IM/RTC 完成。
- 不能说资金结算闭环完成。
- 不能说对象存储生产权限和 Vault 完成。

## 6. 接下来建议推进顺序

### 第 1 优先级：GitHub 和协作基线

- 已完成初始化 Git 仓库并推送到 `HCRXchenghong/infinitech2.0`。
- 已完成 GitHub Actions：`npm run verify`、`cd services/api-go && go test -count=1 ./...`。
- 已完成 PR 模板、Issue 模板、CODEOWNERS、Dependabot。
- 已确认 `.gitignore` 不上传本地二进制、日志、env 和临时文件。
- 下一步：观察第一次 GitHub Actions 运行结果，若 GitHub 托管环境缺少 Go 1.25 工具链则把 workflow 调整为平台可用版本并同步 `go.mod` 策略。

### 第 2 优先级：管理端 Web 最小可运营

- 已建立 admin-web 技术栈和 workspace 测试。
- 已做登录操作台。
- 已做运营首页首版。
- 已接入商户/站长/骑手邀约、退款策略、售后、对象清理、outbox 和订单状态补偿操作。
- 已做订单监控 P0 业务视图首版。
- 已做售后审核 P0 业务视图首版。
- 已做商户资质 P0 业务视图首版。
- 已做骑手/站长管理 P0 业务视图首版。
- 已做派单审计和对象存储清理入口。
- 已做退款策略和对象清理统计展示入口。
- 已做 P0 业务详情面板首版：订单、售后、商户、骑手/站长、骑手绩效、派单、审计、退款策略和权限治理表格行可打开详情并跳到下一步操作。
- 已做高风险操作二次确认与结果追踪首版：邀约、退款策略、审计导出/告警/归档、RBAC 变更、Outbox 领取/续租/死信解封/单事件恢复/批量恢复/标记失败/标记已发布和订单状态补偿执行前进入确认面板，执行后保留最近结果。
- 已做失败回放入口首版：失败记录可恢复原操作和参数，高风险动作重试仍需再次确认。
- 已做 P0 业务筛选分页首版：业务视图可按关键字筛选、调整每页条数并分页查看。
- 已做售后审核表单首版：售后模块和详情抽屉可预填 `request_id`、`decision`、`reason`、`refund_destination` 和 `refund_idempotency_key`，进入 `POST /api/after-sales/{requestID}/review` 前必须二次确认。
- 已做订单退款表单首版：订单模块和详情抽屉可预填 `order_id`、`reason`、`idempotency_key`、`amount_fen` 和 `destination`，进入 `POST /api/orders/{orderID}/refund` 前必须二次确认。
- 已做 Outbox 单事件恢复表单首版：运营首页和 Outbox 队列详情抽屉可预填 `event_id`，进入 `POST /api/admin/outbox/events/{eventID}/replay` 前必须二次确认。
- 已做 Outbox 发布/失败人工处置表单首版：运营首页和 Outbox 队列详情抽屉可预填 `event_id`、失败原因、重试延迟和最大尝试次数，进入 `POST /api/admin/outbox/events/{eventID}/failed` 或 `POST /api/admin/outbox/events/{eventID}/published` 前必须二次确认。
- 已做 Outbox 领取/续租表单首版：运营首页和 Outbox 队列详情抽屉可预填 topic、limit、lease owner、lease seconds 和 `event_id`，进入 `POST /api/admin/outbox/events/claim` 或 `POST /api/admin/outbox/events/{eventID}/lease/renew` 前必须二次确认。
- 已做 Outbox 死信分诊/解封表单首版：运营首页和 Outbox 队列详情抽屉可预填 `status=dead_letter` 查询和死信事件 ID，进入 `POST /api/admin/outbox/events/{eventID}/replay` 解封前必须二次确认。
- 已做 Outbox 单事件事故辅助明细首版：运营首页和 Outbox 队列详情抽屉可预填 `event_id`，进入 `GET /api/admin/outbox/events/{eventID}` 查看事故状态、payload 摘要、关联目标、最近审计、推荐操作和处置核查清单。
- 已做商户资质审核后端与表单首版：商户资质上传进入 `pending_review`，后台审核通过后才计入接单资格，审核动作写入审计并在 Admin Web 商户详情抽屉进入二次确认。
- 已做商户资质审核结果可靠通知首版：审核通过/驳回会生成 `merchant.qualification_reviewed` outbox 事件，notification-worker 可生成商户通知 payload，relay 默认投递 topic 与部署骨架已覆盖。
- 已做商户站内通知中心首版：通知 worker 可把审核结果写入平台通知账本，商户可查询未读通知并标记已读。
- 已做通知运营查询接口首版：后台可按商户、状态和来源事件查询通知账本，支持客服排查通知争议。
- 已做通知投递回执台账首版：worker 写入站内通知后会记录 delivered 回执，后台可查询 failed 回执和错误信息。
- 已做通知失败回执告警首版：后台可把 failed 回执按商户、渠道和 provider 汇总成 `notification.delivery_failed_alerts` outbox 事件，写入审计，并由 notification-worker 生成安全告警通知 payload。
- 已做通知失败重试编排首版：后台可把 failed 回执按商户、渠道和 provider 汇总成 `notification.delivery_retries` outbox 事件，按 provider 退避窗口延迟进入 ready 队列，写入审计，并由 notification-worker 生成重试计划通知 payload。
- 已做通知 provider 执行器首版：`notification-worker` 可把初始通知或重试任务转成 provider dispatch，调用配置 endpoint/adapter，并把 provider message id、失败码和失败原因写回投递回执；重试 outbox payload 已携带原始通知快照，便于真实渠道重发正文。
- 已做通知 provider 回调验签入账首版：`POST /api/notifications/provider-callback` 可验签并幂等记录外部 provider 异步 delivered/failed/queued 回执，BFF、worker 签名工具和部署 secret 位已接入。
- 已做通知 provider 模板映射与渠道 payload 规范首版：`notification-worker` 可把通知 type/template_key 映射成真实渠道模板和参数，provider endpoint/adapter 收到的 dispatch 已包含 `template_id`、`template_params` 和 `provider_payload`。
- 已做通知偏好与静默窗口首版：`notification-worker` 可按目标/类型偏好禁用外部渠道，或在静默窗口把 provider 投递转为 queued 回执并写入原因，避免上线后误触达。
- 已做通知偏好后端账本与 API 首版：商户和运营都可通过 API 读写目标/类型级通知偏好，运营写入带审计，PostgreSQL-backed Store 有规范表和索引。
- 已做通知 worker 后端偏好读取首版：worker 投递 provider 前会从后端偏好账本读取精确规则，偏好不可读时失败关闭为 queued 回执，避免绕过商户偏好误触达。
- 已做通知静默 queued 再投递调度首版：后台可按 queued 状态、`notification_quiet_window` 错误码和 `retry_at` 把静默窗口回执调度为延迟 `notification.delivery_retries` outbox，worker 到点可按原通知快照重发 provider。
- 已做通知静默到期自动扫描调度首版：静默窗口 queued 回执会记录静默结束 `retry_at`，后台可扫描到期回执生成 `notification.delivery_retries`，worker 可通过环境变量开启周期扫描。
- 已做商户端通知偏好设置首版：商户端可配置订单状态、资质审核通知的外部触达渠道和静默时间，直接写入后端偏好账本供 worker 投递前读取。
- 已做用户端通知偏好设置首版：用户小程序可配置订单状态、售后进度和优惠活动通知的外部触达渠道与静默时间，后端会把偏好强制限定在当前用户自身。
- 已做通知 worker 偏好缓存与失败关闭首版：worker 按 key 缓存后端偏好并支持 stale-if-error，偏好服务不可读且无缓存时仍 queued 外部投递，避免误触达。
- 已做通知偏好变更事件与 worker 主动失效首版：保存偏好会写入 `notification.preferences_changed` outbox，relay 默认发布，worker 消费后只失效对应偏好缓存 key。
- 已做通知偏好批量保存与策略审计首版：`POST /api/admin/notification-preferences/batch` 支持一次保存多条偏好，PostgreSQL-backed Store 在同一事务内 upsert 偏好、插入偏好变更 outbox 并写入 `admin.notification_preferences.batch_saved` 审计，Admin Web 操作入口已纳入高风险二次确认。
- 已做通知偏好变更审批与灰度应用首版：后台可提交通知偏好变更申请并固化 rollout，由另一名管理员审批或驳回，已审批申请再按灰度范围手动应用到批量保存路径；应用动作写入 `admin.notification_preferences.change_applied` 并记录 applied/skipped 范围，驳回申请不能应用。
- 已做审计中心增强首版：actor/action/target/after/before/limit 筛选、before 游标翻页、保存筛选、详情抽屉、跨模块跳转、脱敏 payload 摘要、CSV 导出和留存/告警健康报告。
- 已做审计导出首版：审计 CSV 导出复用检索筛选条件，导出动作写入 `admin.audit_logs.exported`，导出 payload 只保留格式、行数、筛选条件和生成时间。
- 已做审计留存/告警健康报告首版：审计报告按留存窗口、热存窗口、完整性抽样、导出事件和关键动作覆盖生成状态与告警；这只是报告能力，不是 WORM 归档或真实渠道告警。
- 已做审计留存告警 outbox 投递首版：留存报告中的告警可通过 `POST /api/admin/audit-logs/retention-alerts/emit` 进入 `audit.retention_alerts` 可靠事件队列，并写入 `admin.audit_retention_alerts.emitted` 审计；这仍不是短信、企业微信、电话等真实渠道投递。
- 已做审计 WORM/冷归档请求首版：冷归档请求可通过 `POST /api/admin/audit-logs/archive/request` 生成归档 manifest、manifest hash、归档路径和 `audit.archive_requested` outbox 事件，并写入 `admin.audit_archive.requested` 审计。
- 已做审计归档 worker、归档回查和校验历史查询首版：`services/audit-archive-worker` 可领取归档 outbox、校验 manifest hash、上传 JSONL 归档、附带对象锁请求头、回写归档完成证据，并按结果回写 outbox；`GET /api/admin/audit-logs/archive/records` 可查询归档完成记录，`POST /api/admin/audit-logs/archive/verify` 可下载归档对象并校验 content hash、manifest header、字节数和条目数，`GET /api/admin/audit-logs/archive/verifications` 可按归档 ID/状态/时间范围回看校验台账。这仍不是生产 bucket 策略、保留期删除审批或法律级不可抵赖链。
- 已做审计服务端安全边界首版：`security_auditor` 只读审计角色、审计 payload 服务端白名单和敏感字段掩码。
- 已做审计完整性证明首版：`sha256:v1`/`hmac-sha256:v1` 签封审计规范化字段和白名单 payload，Admin Web 可展示验证状态。
- 已做退款策略配置、管理端订单退款、售后审核、订单状态补偿、对象清理完成/失败、outbox 运维与商户/骑手邀约审计同事务首版：后台退款策略保存会在仓储级原子路径内同时更新配置和写入审计，管理端订单退款会在仓储级原子路径内同时写入退款业务账本和审计，售后审核会在仓储级原子路径内同时写入审核结果、必要退款和审计，订单状态补偿会在仓储级原子路径内同时写入修复结果和审计，对象清理完成/失败会在仓储级原子路径内同时写入上传票据清理状态和审计，outbox 运维会在仓储级原子路径内同时更新 outbox 事件状态和审计，商户/骑手邀约会在仓储级原子路径内同时生成最终邀约和审计。
- 已做管理端服务端 RBAC 策略矩阵首版：`ops_admin`、`finance_admin`、`dispatch_admin`、`support_admin`、`security_auditor` 等角色已由服务端 scope 守护关键后台路由。
- 已做 RBAC 权限治理查询、变更申请、审批/驳回台账、手动应用与审计回滚首版：后台可读取真实服务端矩阵，`admin`/`super_admin` 可提交待审批权限申请并写入审计，另一名管理员可审批或驳回并写入审计；已审批申请可手动应用到运行时权限矩阵并写入 `admin.rbac.change_applied`，当前已应用申请可按应用前 scopes 回滚并写入 `admin.rbac.change_rolled_back`，服务启动会按应用/回滚审计重放恢复策略。
- 下一步：把 P0 行详情面板继续接更多后端明细接口和更完整的审核辅助信息，并补字段级/租户级 RBAC、剩余后台配置/运营处置/资金风控写路径审计同事务、生产 WORM bucket 保留策略、归档校验生产演练回查、保留期删除审批、真实通知/告警渠道生产账号、模板审批、provider sandbox 联调、通知策略升级与回滚、跨端消息中心和 KMS/链式不可抵赖签名。

### 第 3 优先级：微信支付生产链路

- 接入真实 JSAPI 下单参数。
- 接入微信支付平台证书/公钥。
- 接入真实回调验签。
- 接入微信退款 API。
- 接入退款回调和对账。
- 给支付、退款、钱包补并发和重放压测。

### 第 4 优先级：实时 IM 和骑手位置

- 设计消息表、会话表、群成员表、已读表。
- 用户/骑手/商户/客服会话落库。
- 官方群自动入群默认免打扰。
- 商户群、进群领券门槛。
- 骑手位置上报与订单订阅。
- 离线补偿和重连恢复。

### 第 5 优先级：骑手计价、结算和提现

- 后台配置骑手计价规则。
- 订单保存计价版本和明细。
- 骑手端展示收入明细。
- settlement-worker 落库商户结算、骑手收入、平台抽佣。
- 提现申请、审核、打款状态、失败重试。

### 第 6 优先级：营销和社交闭环

- 用户邀请页。
- 优惠券资金责任：商户承担、平台承担、商户同意活动承担。
- 群聊红包、拼手气红包、余额资金冻结/领取/退回。
- 圈子内容审核、举报、风控。
- 找饭搭问卷、真实性承诺、免责协议、匹配规则。

### 第 7 优先级：生产基础设施和 10 万在线验证

- Docker 镜像构建。
- K8s 环境拆分：dev/staging/prod。
- PostgreSQL HA、PgBouncer。
- Redis Cluster/Sentinel。
- Kafka 多 Broker。
- MinIO + Vault/STS。
- Prometheus/Grafana/Loki/Tempo/OpenTelemetry。
- 10k、30k、60k、100k 压测。
- API/BFF/Realtime/Redis/PostgreSQL/Kafka 故障演练。

## 7. 当前关键文档索引

- 总计划：`PLATFORM_MASTER_PLAN.md`
- 执行台账：`EXECUTION_LEDGER.md`
- 商业级验收清单：`docs/product/commercial-readiness-checklist.md`
- 旧版复用审计：`docs/LEGACY_REUSE_AUDIT.md`
- 美团和旧版对标矩阵：`docs/product/meituan-legacy-parity-matrix.md`
- 容量与容灾：`docs/operations/capacity-and-dr.md`
- 系统架构图文档：`docs/architecture/system-architecture.md`
- 用户小程序预览图：`docs/product/user-miniprogram-preview.png`

## 8. 当前仓库上传注意事项

- 不上传 `.env`、`.env.*`、日志、`node_modules`、临时目录。
- 不上传 Go 测试二进制，例如 `services/api-go/platform.test`。
- 远程仓库当前只有一个短 README，本地 README 会成为主 README。
- 如果远程拒绝推送，优先检查 GitHub 凭据和分支保护，不建议强推。
