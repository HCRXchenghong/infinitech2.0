# Infinitech 2.0 商业级验收清单

更新时间：2026-05-25

商业级不是“页面看起来完整”，而是关键业务在真实用户、真实支付、真实履约、真实故障下还能可靠运行。未完成压测和容灾前，只能说“按商业化目标建设中”。

最近完成、当前缺口和后续推进顺序见 `docs/product/recent-progress-roadmap.md`。

## 1. 产品闭环

- 状态：进行中
- 外卖：浏览、加购、地址、结算、微信支付、余额支付、商户接单、骑手抢/派单、配送追踪、完成、评价、售后。
- 团购：购买、券码、商户扫码核销、过期/售罄退款、评价、结算。
- 买药：药房/医务室资质、药品商品、用药咨询入口、配送履约、监管审计。
- 快递/跑腿：任务下单、价格预估、骑手履约、异常处理。
- 圈子/饭搭：发帖、审核、举报、饭搭资料前置、风控免责。

## 2. 资金安全

- 状态：进行中
- 金额全用整数分。
- 钱包余额只能由流水驱动。
- 支付、退款、充值回调、提现、结算必须幂等。
- 用户余额支付必须有支付密码和错误锁定。
- 微信支付回调必须验签、重放保护和对账。
- 退款默认策略可配置为余额或原路返回，必须写审计。

## 3. 账号和权限

- 状态：进行中
- 用户、商户、骑手、站长、管理员分角色鉴权。
- 商户和骑手/站长为邀约制注册。
- 商户资质、员工资料、保证金控制接单能力。
- 骑手保证金或微信免押控制接单能力。
- 管理端 RBAC、操作日志、敏感字段脱敏。

## 4. 数据和审计

- 状态：进行中
- PostgreSQL 迁移版本化。
- 订单、订单项、订单事件、钱包账户、钱包流水、派单任务、消息、审计日志持久化。
- 地址、订单、计价、商品名必须保存快照。
- 关键配置版本化，历史订单不被后台改价影响。
- 备份、恢复、数据导出、数据删除合规流程。

## 5. 实时和履约

- 状态：未完成
- IM 消息落库、已读、撤回、离线补偿。
- 骑手位置上报和轨迹抽样。
- 抢单 1 万并发只允许一个成功。
- 10 分钟抢单大厅后进入自动派单。
- 拒绝后派给下一位在线骑手。
- 每日一次免责取消和固定订单数后免责不接可配置可审计。
- RTC 语音信令、呼叫状态和通话审计。

## 6. 端侧质量

- 状态：进行中
- 用户端原生微信小程序，继续使用 `WXML/WXSS/TypeScript`。
- 商户端独立 Flutter/Dart。
- 骑手端独立 Flutter/Dart。
- 管理端桌面 Web 和移动管理 Flutter/Dart。
- UI 使用旧版 logo 和 `#009bf5`。
- 核心流程必须有真机/浏览器截图验收。
- 端侧只访问 BFF 或公开 API，不直连内部 Worker 和数据库。
- 当前证据：`apps/admin-web` 已有最小运营控制台首版，接入管理员登录、邀约准入、运营快照、操作审计、退款策略、订单退款表单、售后列表、售后审核表单、商户资质审核表单、对象清理、outbox 运维、订单状态补偿和 RBAC 权限治理等 BFF/API；已补订单监控、售后审核、商户资质、骑手/站长、骑手绩效、派单审计、退款策略、权限治理 P0 业务视图首版；`/api/admin/operations/snapshot` 已能聚合订单、商户、骑手、售后、派单、退款策略、outbox 和对象清理状态；`/api/admin/audit-logs` 已能查询商户/骑手邀约、退款策略、订单退款、状态补偿、售后审核、商户资质审核、对象清理、outbox 运维和 RBAC 变更申请/审批/应用/回滚等关键写操作审计，PostgreSQL-backed Store 下直接读取规范化 `audit_logs` 表，并支持 after/before 时间范围；`/api/admin/audit-logs/export` 已能按同一套筛选导出 CSV，并写入 `admin.audit_logs.exported` 审计；`/api/admin/audit-logs/retention-report` 已能生成审计留存/告警健康报告，覆盖留存窗口、冷归档候选、完整性失败、导出事件和关键动作覆盖；`/api/admin/audit-logs/archive/request` 已能生成冷归档 manifest、`sha256:v1` manifest hash、归档路径和 `audit.archive_requested` outbox 事件；`services/audit-archive-worker` 已能领取该归档事件、校验 manifest hash、上传 JSONL 归档文件，回写 `/api/admin/audit-logs/archive/complete` 完成证据，再回写 outbox；`/api/admin/audit-logs/archive/records` 已能查询归档完成记录；`/api/admin/audit-logs/archive/verify` 已能从归档下载地址读取 JSONL 并校验 content hash、manifest header、字节数和条目数；`/api/admin/audit-logs/archive/verifications` 已能按归档 ID、状态、时间范围和 limit 回看归档校验台账；后端已新增 `security_auditor` 只读审计角色和 `ops_admin`、`finance_admin`、`dispatch_admin`、`support_admin` 等服务端 RBAC scope 边界，`GET /api/admin/rbac/policy` 可返回真实服务端矩阵，`GET /api/admin/rbac/change-requests` 可读取申请台账，`POST /api/admin/rbac/change-requests` 可提交待审批申请，`POST /api/admin/rbac/change-requests/{id}/review` 可审批/驳回且禁止自审，`POST /api/admin/rbac/change-requests/{id}/apply` 可把已审批申请手动应用到运行时权限矩阵并写入 `admin.rbac.change_applied`，`POST /api/admin/rbac/change-requests/{id}/rollback` 可把当前已应用申请按应用前 scopes 回滚并写入 `admin.rbac.change_rolled_back`，服务启动会按应用/回滚审计恢复策略；审计 payload 已在 Store 与 PostgreSQL 路径统一白名单过滤和敏感字段掩码；审计完整性证明首版已返回 `integrity_algorithm`、`integrity_hash`、`integrity_verified`，生产可用 `AUDIT_LOG_SIGNING_SECRET` 启用 HMAC 篡改检测；退款策略配置、管理端订单退款、售后审核、商户资质审核、订单状态补偿、对象清理完成/失败、outbox 运维和商户/骑手邀约已完成业务写入与审计写入同事务首版，PostgreSQL-backed Store 分别在同一事务内写入 `refund_settings`/退款业务账本/售后审核账本/资质审核结果/订单状态补偿/对象清理票据状态/`platform_outbox_events`/邀约快照与 `audit_logs`；Admin Web 已新增审计中心增强首版，支持 actor/action/target/after/before/limit 筛选、before 游标翻页、保存筛选、详情抽屉、按目标筛选、跨模块跳转、白名单摘要、敏感字段脱敏详情、完整性状态展示、CSV 导出、留存告警报告、归档请求、归档记录、归档对象校验、校验历史查询操作和归档校验历史可视化面板；管理端已能用运营快照生成顶部 KPI、今日队列和 P0 表格，已补 P0 行详情面板首版、高风险操作二次确认与结果追踪首版、失败回放入口首版、P0 业务筛选分页首版、售后审核表单首版、商户资质审核表单首版、订单退款表单首版、Outbox 单事件恢复表单首版、Outbox 发布/失败人工处置表单首版、Outbox 领取/续租表单首版、Outbox 死信分诊/解封表单首版和 Outbox 单事件事故辅助明细首版，并对快照展示字段做 HTML 转义；BFF 已支持浏览器来源 CORS 白名单和 `OPTIONS` 预检，允许管理端 Web 携带管理员 token 访问快照、审计、资质审核和 RBAC 接口；本轮 `npm run verify`、`git diff --check`、`cd services/api-go && go test -count=1 ./...` 和本地浏览器渲染检查已通过。仍需补更多后端业务明细接口、字段级/租户级 RBAC、剩余后台配置/运营处置/资金风控写路径同事务审计、生产 WORM bucket 策略、归档校验生产演练回查、保留期删除审批、真实告警投递、KMS/链式不可抵赖签名和截图归档。

- 本轮补充证据：已新增 `POST /api/admin/audit-logs/retention-alerts/emit`，可把审计留存报告中的 critical/warning 告警投递为 `audit.retention_alerts` outbox 事件，并写入 `admin.audit_retention_alerts.emitted` 审计；BFF、Admin Web 操作目录、notification-worker、HTTP 回归测试和架构守卫已覆盖。该能力是可靠事件投递首版，真实短信、企业微信、电话值班和工单渠道仍待补。

- 本轮补充证据：已新增 `POST /api/admin/audit-logs/archive/request`，可按热存窗口生成审计冷归档 manifest、`sha256:v1` manifest hash、归档路径和 `audit.archive_requested` outbox 事件，并写入 `admin.audit_archive.requested` 审计；BFF、Admin Web、Store/HTTP 测试和架构守卫已覆盖。该能力是 WORM/冷归档请求首版，归档执行已由后续 `audit-archive-worker` 承接。

- 本轮补充证据：已新增 `services/audit-archive-worker`，可领取 `audit.archive_requested` outbox、校验 manifest hash、上传 `application/x-ndjson` 归档、附带对象锁请求头和审计 hash 元数据，回写 `admin.audit_archive.completed` 完成证据，并按结果回写 outbox；Docker Compose 与 K8s 部署骨架已覆盖。仍需生产 bucket 强制对象锁、归档校验生产演练回查、保留期删除审批、KMS/链式不可抵赖签名和真实演练报告。

- 本轮补充证据：已新增 `GET /api/admin/audit-logs/archive/records` 和 `POST /api/admin/audit-logs/archive/complete`，归档 worker 上传成功后先记录完成证据，后台可查询 storage key、manifest/content hash、bytes、对象锁模式、保留期、上传时间和 outbox event id；BFF、Admin Web 操作目录、HTTP/Store/worker 测试和架构守卫已覆盖。

- 本轮补充证据：已新增 `POST /api/admin/audit-logs/archive/verify`，可根据归档完成记录下载 JSONL 归档对象，校验 content hash、manifest header、bytes 和 manifest entry 数，并写入 `admin.audit_archive.verified` 审计；BFF、Admin Web 操作目录、HTTP/Store 测试和架构守卫已覆盖。该能力是下载校验首版，校验历史查询已由后续 `/archive/verifications` 承接，生产 WORM bucket 策略、保留期删除审批和 KMS/链式不可抵赖签名仍待补。

- 本轮补充证据：已新增 `GET /api/admin/audit-logs/archive/verifications`，可从 `admin.audit_archive.verified` 审计账本重建归档校验历史，支持 archive id、状态、时间范围和 limit 查询；BFF、Admin Web 操作目录、HTTP/Store/Admin/BFF 测试和架构守卫已覆盖。该能力已由审计中心可视化面板承接展示，生产演练报告和法律级不可抵赖归档仍待补。

- 本轮补充证据：Admin Web 审计检索页已新增“归档校验历史”可视化面板，可按归档 ID、状态和条数查询并展示状态、校验时间、manifest/content hash、bytes/logs、匹配摘要和原始详情；Admin Web 单测、架构守卫和本地浏览器渲染检查已覆盖。

- 本轮补充证据：Admin Web 已新增 P0 业务详情面板首版，订单、售后、商户、骑手/站长、骑手绩效、派单、审计、退款策略和权限治理表格行可打开详情，展示字段、核查清单和下一步操作入口；`npm run test --workspace @infinitech/admin-web` 与 `npm run verify:architecture` 已覆盖。

- 本轮补充证据：Admin Web 已新增高风险操作二次确认与结果追踪首版，邀约、退款策略、审计导出/告警/归档、RBAC 变更、Outbox 领取/续租/死信解封/单事件恢复/批量恢复/标记失败/标记已发布和订单状态补偿执行前会进入 `pending_confirmation` 确认面板，确认执行后会保留最近操作结果；Admin Web 单测、架构守卫和本地浏览器渲染检查已覆盖。

- 本轮补充证据：Admin Web 已新增失败回放入口首版，失败记录可一键恢复原操作和参数，高风险动作重试时仍重新进入 `pending_confirmation` 二次确认面板；Admin Web 单测、架构守卫和本地浏览器渲染检查已覆盖。

- 本轮补充证据：Admin Web 已新增 P0 业务筛选分页首版，业务视图可按关键字筛选、选择每页 4/10/20 条并上一页/下一页翻看；筛选后详情按钮仍对应原始业务行；Admin Web 单测、架构守卫和本地浏览器渲染检查已覆盖。

- 本轮补充证据：Admin Web 已新增售后审核表单首版，售后模块和详情抽屉可预填工单 ID、审核结果、审核原因、退款去向和退款幂等键，进入现有 `POST /api/after-sales/{requestID}/review` 前必须二次确认；Admin Web 单测、架构守卫和本地浏览器渲染检查已覆盖。

- 本轮补充证据：Admin Web 已新增订单退款表单首版，订单模块和详情抽屉可预填订单 ID、退款原因、退款幂等键、可选退款金额和退款去向，进入现有 `POST /api/orders/{orderID}/refund` 前必须二次确认；Admin Web 单测、架构守卫和本地浏览器渲染检查已覆盖。

- 本轮补充证据：Admin Web 已新增 Outbox 单事件恢复表单首版，运营首页和 Outbox 队列详情抽屉可预填事件 ID，进入现有 `POST /api/admin/outbox/events/{eventID}/replay` 前必须二次确认；Admin Web 单测、架构守卫和本地浏览器渲染检查已覆盖。

- 本轮补充证据：Admin Web 已新增 Outbox 发布/失败人工处置表单首版，运营首页和 Outbox 队列详情抽屉可预填事件 ID、失败原因、重试延迟和最大尝试次数，进入现有 `POST /api/admin/outbox/events/{eventID}/failed` 或 `POST /api/admin/outbox/events/{eventID}/published` 前必须二次确认；Admin Web 单测、架构守卫和本地浏览器渲染检查已覆盖。

- 本轮补充证据：Admin Web 已新增 Outbox 领取/续租表单首版，运营首页和 Outbox 队列详情抽屉可预填 topic、limit、lease owner、lease seconds 和事件 ID，进入现有 `POST /api/admin/outbox/events/claim` 或 `POST /api/admin/outbox/events/{eventID}/lease/renew` 前必须二次确认；Admin Web 单测、架构守卫和本地浏览器渲染检查已覆盖。

- 本轮补充证据：Admin Web 已新增 Outbox 死信分诊/解封表单首版，运营首页、今日队列和 Outbox 队列详情抽屉可预填 `status=dead_letter` 查询和死信事件 ID，进入现有 `POST /api/admin/outbox/events/{eventID}/replay` 解封前必须二次确认；Admin Web 单测、架构守卫和本地浏览器渲染检查已覆盖。

- 本轮补充证据：已新增 `GET /api/admin/outbox/events/{eventID}` Outbox 单事件事故辅助明细，返回 ready/blocked/lease 信号、payload 摘要、关联目标、最近审计、推荐操作和人工核查清单；BFF、Admin Web 操作目录/详情抽屉、Store/HTTP/BFF/Admin Web 测试和架构守卫已覆盖。

- 本轮补充证据：商户资质审核已从“上传即通过”改为待审闭环，`POST /api/merchant/qualifications` 只写入 `pending_review`，`POST /api/admin/merchant-qualifications/{qualificationID}/review` 由 `merchant:qualification_review` scope 守护并写入 `admin.merchant_qualification.reviewed` 审计，审核通过后才计入接单资格；BFF、Admin Web 商户模块/详情抽屉、Store/HTTP/Admin/BFF 测试和架构守卫已覆盖。

- 本轮补充证据：已新增 `GET /api/admin/merchant-qualifications` 和 `GET /api/admin/merchant-qualifications/{qualificationID}` 商户资质待审列表/明细，返回资质状态、商户/店铺/保证金上下文、缺失资质、接单资格、最近审核审计、推荐操作和核查清单；BFF、Admin Web 操作目录/详情抽屉、Store/HTTP/BFF/Admin Web 测试和架构守卫已覆盖。

- 本轮补充证据：商户资质审核结果已进入 `merchant.qualification_reviewed` outbox topic，审核响应返回 `outbox_event`，`notification-worker` 可生成面向商户的审核结果通知 payload，outbox relay 默认 topic、Docker Compose 和 K8s 配置已覆盖；站内通知账本、商户已读/未读 API、后台查询、投递回执、Admin Web 通知运营页、失败回执告警、失败重试编排、provider 执行器骨架、通知偏好 UI/API 和通知偏好审批应用已由后续通知中心增量承接，真实短信、企业微信、微信小程序订阅消息、商户端 push 生产账号、真实渠道联调和跨端消息中心仍待补。

- 本轮补充证据：已新增商户站内通知中心首版，`POST /api/notifications` 由 `notification:write` scope 守护并支持 worker 幂等写入，`GET /api/merchant/notifications` 可按商户和状态查询，`POST /api/merchant/notifications/{notificationID}/read` 可标记已读；PostgreSQL-backed Store 已新增 `platform_notifications` 规范表、幂等唯一约束和查询索引，BFF、notification-worker、HTTP/Store 测试和架构守卫已覆盖。

- 本轮补充证据：已新增通知运营查询接口首版，`GET /api/admin/notifications` 由 `notification:read` scope 守护，支持按目标商户、状态、source topic/event 和 limit 查询通知账本；`support_admin` 可只读排查通知争议，`ops_admin` 可读写通知账本，`security_auditor` 不读取运营通知。BFF、HTTP/RBAC 测试和架构守卫已覆盖。

- 本轮补充证据：已新增通知投递回执台账首版，`POST /api/notifications/{notificationID}/deliveries` 可记录 delivered/failed 回执、provider message id、错误码和错误信息，`GET /api/admin/notification-deliveries` 可按商户、通知、渠道、provider 和状态查询；PostgreSQL-backed Store 已新增 `platform_notification_deliveries` 规范表、幂等唯一约束和查询索引，notification-worker 写入站内通知成功后会记录 in-app delivered 回执，BFF、HTTP/Store/worker 测试和架构守卫已覆盖。

- 本轮补充证据：已新增 Admin Web 通知运营页首版，`notifications` 模块可打开通知台账、通知回执和补录回执表单；详情抽屉可按通知 ID、商户、来源 topic 和失败状态预填查询/补录动作，补录回执纳入高风险二次确认；Admin Web RBAC 已同步 `notification:read`/`notification:write`，Admin Web 单测和架构守卫已覆盖。

- 本轮补充证据：已新增通知失败回执告警首版，`POST /api/admin/notification-deliveries/failure-alerts/emit` 由 `notification:write` scope 守护，可把 failed 回执按商户、渠道和 provider 汇总成 `notification.delivery_failed_alerts` outbox 事件并写入 `admin.notification_delivery_failure_alerts.emitted` 审计；BFF、Admin Web 高风险二次确认、notification-worker、outbox relay 默认 topic 和部署骨架已覆盖。

- 本轮补充证据：已新增通知失败重试编排首版，`POST /api/admin/notification-deliveries/retries/schedule` 由 `notification:write` scope 守护，可把 failed 回执按商户、渠道和 provider 汇总成 `notification.delivery_retries` outbox 事件，事件 `available_at` 对齐 `retry_at`，只在 provider 退避窗口后进入 ready 队列，并写入 `admin.notification_delivery_retries.scheduled` 审计；BFF、Admin Web 高风险二次确认、notification-worker、outbox relay 默认 topic 和部署骨架已覆盖。

- 本轮补充证据：已新增通知 provider 执行器首版，`notification.delivery_retries` outbox payload 会携带原始通知快照，`notification-worker` 可按 `NOTIFICATION_PROVIDER_CHANNELS` 或重试事件生成短信/企业微信/微信订阅消息/push provider dispatch，调用配置 endpoint/adapter，并把 delivered/failed、provider message id、错误码和错误信息写回 `/api/notifications/{notificationID}/deliveries`；未配置 provider 时记录 `provider_not_configured` 失败回执，Docker Compose 与 K8s 部署骨架已预留 provider endpoint/token 环境变量。

- 本轮补充证据：已新增通知 provider 回调验签入账首版，`POST /api/notifications/provider-callback` 可接收外部渠道 delivered/failed/queued 异步回执，生产配置 `NOTIFICATION_PROVIDER_CALLBACK_SECRET` 后按 HMAC-SHA256 canonical lines 验签，通过 callback idempotency/provider message id 幂等写入 `platform_notification_deliveries`；BFF 代理、notification-worker 签名 payload 工具、Docker Compose 和 K8s secret 位、HTTP/worker/BFF 测试与架构守卫已覆盖。

- 本轮补充证据：已新增通知 provider 模板映射与渠道 payload 规范首版，`notification-worker` 可解析 `NOTIFICATION_PROVIDER_TEMPLATES`，按 notification type/template_key 与 channel 匹配模板，给 provider dispatch 增加 `template_id`、`template_params` 和面向微信订阅消息、短信、企业微信、push 的 `provider_payload`；worker 测试、架构守卫、Docker Compose 与 K8s 配置位已覆盖。

- 本轮补充证据：已新增通知偏好与静默窗口首版，`notification-worker` 可解析 `NOTIFICATION_DELIVERY_PREFERENCES`，按 default/target_role/target_id/type 合并规则，禁用渠道或静默窗口内的外部 provider 投递不会调用 provider endpoint，而是写入 `queued` 回执并记录 `notification_preference_disabled` 或 `notification_quiet_window`，避免误触达又保留运营证据；worker 测试、架构守卫、Docker Compose 与 K8s 配置位已覆盖。

- 本轮补充证据：已新增通知偏好后端账本与 API 首版，`platform_notification_preferences` 可持久化目标/通知类型级 `enabled_channels`、`disabled_channels` 和 `quiet_hours`；商户 `GET/PUT /api/merchant/notification-preferences`、运营 `GET/PUT /api/admin/notification-preferences`、运营写入审计 `admin.notification_preferences.saved`、BFF 代理、HTTP/Store/BFF 测试和架构守卫已覆盖。worker 动态读取、商户端 UI、管理端操作入口、用户端 UI、批量保存和审批应用已由后续增量承接，真实渠道联调仍待补。

- 本轮补充证据：已新增通知 worker 后端偏好读取首版，后端偏好列表支持 `preference_key` 精确查询，`notification-worker` 会在 provider 投递前按决策 key 读取 `/api/admin/notification-preferences` 并合并静态偏好；读取失败时外部渠道转为 `queued` 回执并记录 `notification_preference_lookup_failed`。静默到期自动扫描、商户端偏好 UI、管理端操作入口、用户端偏好 UI、批量保存和审批应用已由后续增量承接，该能力仍需真实渠道生产联调承接。

- 本轮补充证据：已新增通知静默 queued 再投递调度首版，`POST /api/admin/notification-deliveries/retries/schedule` 可按 `status=queued`、`error_code=notification_quiet_window` 和 `retry_at` 汇总静默窗口回执，生成 `notification.delivery_retries` 延迟 outbox，并继续携带原通知快照供 worker 到点重发；`GET /api/admin/notification-deliveries` 支持 `error_code` 筛选，Admin Web 重试表单支持 queued 状态和指定重试时间。自动扫描、商户端偏好 UI、管理端操作入口、用户端偏好 UI、批量策略保存和审批应用已由后续增量承接，真实渠道联调仍待补。

- 本轮补充证据：已新增通知静默到期自动扫描调度首版，静默窗口 queued provider 回执会记录静默结束 `retry_at`；`GET /api/admin/notification-deliveries` 支持 `retry_at_before`，`POST /api/admin/notification-deliveries/quiet-window-retries/schedule` 可扫描 `queued + notification_quiet_window + retry_at<=now` 回执并调度 `notification.delivery_retries` 延迟 outbox；`notification-worker` 可通过 `NOTIFICATION_QUIET_RETRY_AUTO_SCHEDULE` 开启周期扫描，BFF、Admin Web、Docker Compose、K8s、测试和架构守卫已覆盖。商户端偏好 UI、管理端操作入口、用户端偏好 UI、批量策略保存和审批应用已由后续增量承接，真实渠道生产联调仍待补。

- 本轮补充证据：已新增商户端通知偏好设置首版，`apps/merchant-flutter` 新增“通知偏好”页面和首页入口，商户可按订单状态、资质审核通知配置微信订阅、短信、企业微信、push 开关、静默时间和静默渠道；API client 已接 `GET/PUT /api/merchant/notification-preferences`，架构守卫覆盖页面、路由、payload 字段和通知类型。管理端偏好操作入口、用户端偏好 UI、批量策略保存和审批应用已由后续增量承接，真实渠道生产账号、模板审批和 provider 字段联调仍待补。

- 本轮补充证据：已新增管理端通知偏好操作入口首版，Admin Web 通知运营页可查询 `GET /api/admin/notification-preferences` 并保存 `PUT /api/admin/notification-preferences`；通知详情抽屉会按目标商户、来源 topic 和失败渠道预填 `target_role`、`target_id`、`notification_type`、`disabled_channels` 和 `quiet_hours` JSON。保存动作已纳入高风险二次确认，Admin Web 单测和架构守卫覆盖操作目录、请求构造、CSV 渠道数组、JSON 解析和非法 JSON 阻断。用户端偏好 UI、批量策略保存和审批应用已由后续增量承接，仍需真实渠道账号、模板审批和 provider 字段联调。

- 本轮补充证据：已新增用户端通知偏好设置首版，原生微信小程序新增“通知偏好”页面和首页入口，用户可按订单状态、售后进度和优惠活动配置微信订阅、短信、App Push 开关、静默时间和静默渠道；API client 已接 `GET/PUT /api/user/notification-preferences`，后端强制把保存请求限定到当前 `target_role=user` 和当前用户 ID，BFF、HTTP 测试和架构守卫覆盖用户端读写路径。批量策略保存和审批应用已由后续增量承接，仍需真实渠道账号、模板审批和 provider 字段联调。

- 本轮补充证据：已新增通知 worker 偏好缓存与失败关闭首版，`notification-worker` 可按 preference key 短 TTL 缓存后端偏好，默认 `NOTIFICATION_PREFERENCE_CACHE_TTL_MS=30000`、`NOTIFICATION_PREFERENCE_CACHE_STALE_MS=300000`、`NOTIFICATION_PREFERENCE_CACHE_MAX_KEYS=500`；TTL 命中时不重复请求偏好 API，刷新失败且仍在 stale 窗口内继续用旧偏好，没有缓存时保持 `notification_preference_lookup_failed` queued 失败关闭。worker 单测、Docker Compose、K8s 和架构守卫已覆盖。

- 本轮补充证据：已新增通知偏好变更事件与 worker 主动失效首版，保存偏好会生成 `notification.preferences_changed` outbox 事件，PostgreSQL-backed Store 在同一事务内 upsert 偏好并插入 outbox；outbox relay 默认 topic、Docker Compose 和 K8s 已加入该事件，`notification-worker` 消费后只失效对应 resolver 缓存 key，不会误创建站内通知。批量策略保存和审批应用已由后续增量承接，仍需真实渠道账号、模板审批和 provider 字段联调。

- 本轮补充证据：已新增通知偏好批量保存与策略审计首版，`POST /api/admin/notification-preferences/batch` 可一次保存最多 50 条偏好策略并要求变更原因；后端拒绝同批重复 preference key，PostgreSQL-backed Store 在同一事务内 upsert 偏好、插入每条 `notification.preferences_changed` outbox 并写入 `admin.notification_preferences.batch_saved` 审计；BFF、Admin Web 高风险二次确认、HTTP/Store/BFF/Admin Web 测试和架构守卫已覆盖。审批应用已由后续增量承接，仍需真实渠道账号、模板审批和 provider 字段联调。

- 本轮补充证据：已新增通知偏好变更审批与灰度应用首版，`POST /api/admin/notification-preferences/change-requests` 只提交申请并写入 `admin.notification_preferences.change_requested`，同时固化 `all`、`target_ids` 或 deterministic `percentage` rollout；`POST /api/admin/notification-preferences/change-requests/{id}/review` 支持另一名管理员审批或驳回且禁止自审，`POST /api/admin/notification-preferences/change-requests/{id}/apply` 只允许已审批申请按 rollout 进入批量保存原子路径。应用同事务写入偏好、`notification.preferences_changed` outbox 和 `admin.notification_preferences.change_applied` 审计，并记录 applied/skipped preference keys。BFF、Admin Web 高风险二次确认、HTTP/BFF/Admin Web 测试和架构守卫已覆盖。仍需真实渠道账号、模板审批、provider 字段联调、策略升级/回滚和跨端消息中心。

## 7. 容量和容灾

- 状态：未完成
- Kubernetes 多副本部署。
- PostgreSQL HA + PgBouncer。
- Redis Cluster/Sentinel。
- Kafka 多 Broker。
- MinIO 对象存储。
- Prometheus/Grafana/Loki/Tempo/OpenTelemetry。
- 10k/30k/60k/100k 在线连接压测。
- API 节点、Realtime 节点、Redis、数据库、Kafka 故障演练。

## 8. 安全和合规

- 状态：进行中
- HTTPS、WAF、限流、重放保护。
- OAuth state 防 CSRF。
- 回调验签。
- Token 和证书加密存储。
- 敏感信息脱敏和最小权限。
- 内容审核、饭搭免责协议、未成年人和高风险行为风控策略。

## 9. 当前已具备的证据

- `npm run verify` 通过。
- `cd services/api-go && go test ./...` 通过。
- GitHub Actions 质量门禁首版已建立：push 和 pull request 会运行 `npm run verify` 与 `cd services/api-go && go test -count=1 ./...`。
- GitHub PR 模板、Issue 模板、CODEOWNERS 和 Dependabot 已建立，用于把商业影响、验证证据、回滚说明和商业化缺口纳入协作流程。
- 已有核心数据库迁移：`infra/db/migrations/0001_core.sql`。
- 已有认证和支付补充迁移：`infra/db/migrations/0002_auth_payment.sql`。
- 已有 API 鉴权骨架：`services/api-go/internal/httpapi/auth.go`。
- 已有微信登录签名 token、微信支付预下单和回调验签骨架。
- 已有商户邀约注册、营业执照/健康证资质上传和资质缺失检查；商户邀约注册已要求设置密码，商户可通过 `account_id + password` 登录并签发 session-backed token。
- 已有商户员工健康证和补充资料首批闭环测试：商户可读取/提交员工健康证与门头/后厨等补充资料，保存时校验失效日期和店铺归属，跨商户店铺写入会被隐藏拒绝；BFF 已覆盖 Authorization 转发，商户端 Flutter 已新增资质资料页。
- 已有管理员 bootstrap 密码登录闭环：默认关闭，无内置默认密码，仅配置 `ADMIN_BOOTSTRAP_ACCOUNT_ID` 和 `ADMIN_BOOTSTRAP_PASSWORD` 后启用，错误密码返回 401。
- 已有骑手/站长邀约注册首批闭环测试：站长只能邀请本站骑手，管理员可邀请骑手或站长，接受邀约后签发对应骑手端 token，已使用邀约不可复用。
- 已有店铺接单门槛统一判断：用户加购、购物车结算、商户接单都校验店铺状态、商户资质和商户保证金。
- 已有保证金首批闭环测试：骑手保证金查询/缴纳、微信免押、退押申请、商户保证金缴纳，并同步骑手/商户接单门槛状态。
- 已有商户外卖处理首批闭环测试：商户订单列表、接单、备餐、出餐后进入骑手调度、出餐前禁止骑手抢单。
- 已有商户商品管理首批闭环测试：商户自有店铺商品列表、创建/编辑、上架/售罄/下架，售罄/下架商品对用户端隐藏。
- 已有团购首批闭环测试：团购套餐列表、团购下单、余额支付后发券、用户券列表、商户本店扫码核销、重复核销拒绝。
- 已有骑手调度首批闭环测试：骑手上线、10 分钟后自动派单、10 分钟内自动派单拒绝、骑手拒绝后顺延下一位在线骑手。
- 已有派单审计事件首版测试：自动派单、拒单顺延、站长手动派单和骑手抢单会写入 `DispatchEvent`；管理员可查询派单事件，站长只能查询本站订单派单事件；BFF 已代理查询并保留 Authorization 转发。
- 已有派单确认超时自动转派首版测试：未到确认窗口拒绝转派，到点后写入 `dispatch.timeout` 审计事件，跳过超时骑手并顺延同站点下一位在线可接单骑手；HTTP/BFF 均覆盖 Authorization 转发。
- 已有订单状态机补偿首版测试：管理员可触发 `POST /api/admin/orders/{orderID}/state/compensate`，基于支付交易、钱包流水、订单事件和派单审计事件修复订单状态/骑手漂移，写入 `order.state.compensated` 审计事件；重复补偿幂等，已完成订单不会被派单事件倒退；BFF 已覆盖 Authorization 转发。
- 已有平台 outbox 事件首版测试：订单支付、商户履约、骑手履约、团购核销、状态补偿和派单审计事件会生成可恢复 outbox；管理员可查询 pending、标记 published、标记 failed 并按 backoff 恢复；BFF 已覆盖 Authorization 转发，PostgreSQL 快照和 `platform_outbox_events` 迁移已纳入架构检查。
- 已有 outbox relay worker 首版测试：`@infinitech/outbox-relay-worker` 可轮询 pending outbox、发布到可插拔 publisher、成功标记 published、失败标记 failed 并写入 retry backoff；API client 会转发 `OUTBOX_RELAY_TOKEN` Authorization，已纳入 workspace 和架构检查。
- 已有 outbox relay 可运行化证据：worker 支持长轮询、环境变量配置批量/间隔/backoff、可选 Kafka REST publisher；Docker Compose workers profile 和 K8s 2 副本 Deployment 已落地，并由架构检查校验。
- 已有 outbox 积压观测首版测试：管理员可通过 `GET /api/admin/outbox/stats` 按 topic 查看 total/pending/failed/published、ready/blocked、oldest ready lag 和 per-topic 聚合；失败事件在 retry backoff 期间计为 blocked，到期后计为 ready，BFF 已覆盖 Authorization 与查询参数透传。
- 已有 outbox 手动恢复/重放首版测试：管理员可通过 `POST /api/admin/outbox/events/{eventID}/replay` 将 failed/backoff 事件立即拉回 pending-ready 队列，保留 attempts 审计并拒绝重放已 published 事件；BFF 和 outbox relay worker API client 均覆盖透传。
- 已有 outbox 批量恢复/重放首版测试：管理员可通过 `POST /api/admin/outbox/events/replay` 按 topic/limit 批量恢复仍处于 backoff 阻塞期的 failed/pending 事件，保留 attempts 审计、跳过已经 ready 和 published 的事件；BFF 和 outbox relay worker API client 均覆盖透传。
- 已有 outbox 死信隔离首版测试：relay 标记失败时可传 `max_attempts`，达到上限后事件进入 `dead_letter` 并退出 pending 查询，stats 单独统计 dead_letter，管理员可显式查询死信并通过单事件 replay 人工解封；批量 replay 跳过 dead_letter，outbox relay worker 部署默认 `OUTBOX_RELAY_MAX_ATTEMPTS=10`。
- 已有 outbox relay 租约领取首版测试：管理员可通过 `POST /api/admin/outbox/events/claim` 按 topic/limit 领取 ready 事件并写入 `lease_owner`/`lease_expires_at`，活动租约会从 pending 查询和 ready backlog 中隐藏并在 stats 中计为 `leased`；租约过期后可被其他 relay 重新领取，published/failed/replay 会清理租约；outbox relay worker 优先 claim 再发布，Compose/K8s 默认配置 `OUTBOX_RELAY_WORKER_ID` 和 `OUTBOX_RELAY_LEASE_SECONDS=60`。
- 已有 outbox relay 租约续租首版测试：管理员可通过 `POST /api/admin/outbox/events/{eventID}/lease/renew` 续租当前 owner 的活动租约，错误 owner、过期租约和非 pending/failed 状态会冲突；outbox relay worker 在慢发布期间按 `OUTBOX_RELAY_LEASE_RENEW_INTERVAL_MS=30000` 心跳续租，完成 publish 后清理 timer，BFF 和架构检查均覆盖该链路。
- 已有 PostgreSQL outbox 规范化 relay 路径首版：`PostgresStore` 启动时确保 `platform_outbox_events` 表/索引存在，把 snapshot outbox 幂等补入规范化表；pending/stats/claim/renew/ack/fail/replay 走规范化表，claim 和批量 replay 使用 `FOR UPDATE SKIP LOCKED`，并由架构检查防回退。
- 已有 outbox 租约健康观测首版：`GET /api/admin/outbox/stats` 支持 `lease_expiring_within_seconds`，返回全局、per-topic 和 per-owner 的 `lease_expiring_soon`、`next_lease_expires_at`、`next_lease_expires_in_seconds`，可用于发现 relay 心跳异常、worker 倾斜和租约即将过期的重复投递风险；内存 Store、PostgreSQL outbox 路径、HTTP、BFF 和架构检查均已覆盖。
- 已有管理端运营快照首版：`GET /api/admin/operations/snapshot` 按统一后台口径聚合订单状态/异常、商户资质与保证金、骑手在线与保证金、站长数量、骑手绩效等级、售后队列、派单审计、退款策略、outbox 健康和对象清理统计；HTTP、BFF、Admin Web 操作台和测试已覆盖。管理端已新增快照适配层，可将快照生成 KPI、队列和 P0 表格，并对后端展示字段做 HTML 转义。BFF 已补浏览器来源 CORS 白名单与预检测试，确保本地或配置域名下的管理端 Web 能携带管理员 token 拉取快照。
- 已有管理端操作审计日志首版：`GET /api/admin/audit-logs` 可按 actor、action、target、limit、after 和 before 查询审计账本，`GET /api/admin/audit-logs/export` 可按同一套筛选导出 CSV 并写入 `admin.audit_logs.exported` 审计，`GET /api/admin/audit-logs/retention-report` 可生成留存窗口、冷归档候选、完整性失败、导出事件和关键动作覆盖报告，`POST /api/admin/audit-logs/archive/request` 可生成冷归档 manifest、`audit.archive_requested` outbox 事件和 `admin.audit_archive.requested` 审计；`audit-archive-worker` 已能把归档请求上传为 JSONL 归档，回写 `admin.audit_archive.completed` 完成证据，再回写 outbox；`GET /api/admin/audit-logs/archive/records` 可查询归档完成记录；`POST /api/admin/audit-logs/archive/verify` 可读取归档对象并校验 content hash、manifest header、字节数和条目数；`GET /api/admin/audit-logs/archive/verifications` 可查询归档校验历史。管理员关键写操作已开始记录 actor、action、target、request_id、ip_hash、服务端白名单 payload 和 created_at，覆盖商户邀约、骑手/站长邀约、退款策略保存、订单退款、订单状态补偿、售后审核、对象清理完成/失败、outbox 领取/续租/发布/失败/重放/批量重放和 RBAC 变更申请/审批/应用/回滚；HTTP、BFF、Admin Web 操作台和测试已覆盖。`PostgresStore` 已确保 `audit_logs` 规范化表与 actor/action/target 索引存在，把旧快照审计幂等补入表，通过 `platform_sequences` 行级锁生成审计 `aud_N`，并从表内最大编号恢复序列；PostgreSQL-backed Store 查询直接读取规范化表，并按 `created_at >= after` 与 `created_at < before` 过滤时间窗口。服务端已新增 `security_auditor` 只读审计角色，并新增后台兼容 `admin`、`super_admin`、`ops_admin`、`finance_admin`、`dispatch_admin`、`support_admin` 的 RBAC 策略矩阵首版，关键后台路由已按服务端 scope 判断；`GET /api/admin/rbac/policy` 可查询真实 RBAC 矩阵，`GET /api/admin/rbac/change-requests` 可查询申请台账，`POST /api/admin/rbac/change-requests` 可提交待审批申请并写入 `admin.rbac.change_requested`，`POST /api/admin/rbac/change-requests/{id}/review` 可审批/驳回并写入 `admin.rbac.change_reviewed`，`POST /api/admin/rbac/change-requests/{id}/apply` 可把已审批申请手动应用到运行时权限矩阵并写入 `admin.rbac.change_applied`，`POST /api/admin/rbac/change-requests/{id}/rollback` 可把当前已应用申请按应用前 scopes 回滚并写入 `admin.rbac.change_rolled_back`，API 启动会按应用/回滚审计日志恢复策略。审计 payload 在 Store、SQL 写入、SQL 读取和镜像恢复路径统一白名单过滤，敏感允许字段如 `object_key` 会掩码，非白名单敏感字段不会持久化或原样返回。审计完整性证明首版已覆盖规范化审计字段和白名单 payload，本地默认 `sha256:v1`，生产可配置 `AUDIT_LOG_SIGNING_SECRET` 启用 `hmac-sha256:v1`，查询时返回 `integrity_verified` 以暴露篡改风险。退款策略配置已通过 `SaveRefundSettingsWithAudit` 完成业务写入与审计写入同事务首版，管理端订单退款已通过 `RefundOrderWithAudit` 完成退款业务账本与审计写入同事务首版，售后审核已通过 `ReviewAfterSalesWithAudit` 完成审核结果、必要退款与审计写入同事务首版，订单状态补偿已通过 `CompensateOrderStateWithAudit` 完成状态修复结果与审计写入同事务首版，对象清理完成/失败已通过 `CompleteObjectStorageCleanupWithAudit` 与 `RecordObjectStorageCleanupFailureWithAudit` 完成上传票据状态与审计写入同事务首版，outbox 领取/续租/发布/失败/重放/批量重放已通过对应 `WithAudit` 仓储方法完成 outbox 状态与审计写入同事务首版，商户/骑手/站长邀约已通过 `CreateMerchantInviteWithAudit` 与 `CreateRiderInviteWithAudit` 完成最终邀约与审计写入同事务首版，PostgreSQL-backed Store 均在同一数据库事务内写入业务表、`platform_outbox_events` 或邀约快照与 `audit_logs`。Admin Web 审计检索页已支持 actor/action/target/after/before/limit 筛选、before 游标翻页、保存筛选、详情抽屉、按目标筛选、跨模块跳转、白名单摘要、敏感字段脱敏详情、完整性状态展示、CSV 导出、留存告警报告、归档请求、归档记录、归档对象校验、校验历史查询操作和归档校验历史可视化面板。当前仍待补字段级/租户级 RBAC、剩余后台配置/运营处置/资金风控写路径同事务审计、生产 WORM bucket 策略、归档校验生产演练回查、保留期删除审批、真实告警投递和 KMS/链式不可抵赖签名。
- 本轮补充证据：审计留存告警已能进入 `audit.retention_alerts` outbox topic，`notification-worker` 已订阅并生成安全告警通知 payload；该能力只完成可靠事件和通知 worker 接入，真实告警渠道、值班升级、静默窗口和回执仍待补。
- 本轮补充证据：商户资质审核结果已能进入 `merchant.qualification_reviewed` outbox topic，`notification-worker` 已订阅并生成商户通知 payload，outbox relay 默认发布 topic 已包含该事件；商户站内通知中心首版已支持通知落库、未读列表和已读标记，通知运营查询接口已支持后台按商户/状态/source 回溯，通知投递回执台账已支持 delivered/failed/queued 记录和失败原因查询，Admin Web 通知运营页已支持运营/客服查账、高风险补录回执和通知偏好操作入口，通知失败回执可汇总投递 `notification.delivery_failed_alerts` 并写审计，失败回执也可编排为 `notification.delivery_retries` 延迟重试 outbox 并写审计，provider 执行器骨架已能生成真实渠道 dispatch 并写回 provider 回执，provider 回调验签入账已能接收外部渠道异步回执，provider 模板映射与渠道 payload 规范已为 sandbox 联调准备好 `template_id`/变量结构，偏好过滤和静默窗口已避免外部渠道误触达，静默到期自动扫描调度已能把到期 queued 回执转成可靠重投，商户端、管理端和用户端均已有通知偏好设置首版，worker 已有偏好缓存、偏好变更主动失效、通知偏好批量保存和通知偏好审批应用首版。真实商户短信/企业微信/订阅消息/push 生产账号、provider 凭证、模板审批、真实 provider 字段映射联调和跨端消息中心仍待补。
- 已有消费端幂等落库首版：`@infinitech/domain-core` 提供 consumed-event ledger 和 `createIdempotentConsumer`，dispatch/payment/notification/integration/settlement 五个 worker 已覆盖重复 outbox 投递只执行一次；PostgreSQL 迁移和启动建表已新增 `platform_consumed_events`，为真实 Kafka/NATS 至少一次投递的消费端落库防重做准备。
- 已有支付/钱包 PostgreSQL 规范化恢复首版：`PostgresStore` 启动和写入后会同步/恢复 `orders`、`order_items`、`order_events`、`wallet_accounts`、`wallet_transactions`、`wallet_payment_passwords`、`payment_transactions`，并重建钱包幂等索引、支付密码和微信 `out_trade_no`/`transaction_id` 索引；测试覆盖余额支付重复请求、支付密码恢复、微信回调重复投递和订单事件恢复。
- 已有订单创建 PostgreSQL 事务化首版：`PostgresStore.CreateOrder` 使用数据库事务写入 `orders` 和初始 `order_events`，通过 `platform_sequences` 的 `orders` 序列行级锁生成订单号，并在序列缺失或灾备恢复后根据现有订单最大编号追平；提交后刷新规范化表镜像和 snapshot，架构检查防止回退到内存先写。
- 已有购物车结算 PostgreSQL 事务化首版：`PostgresStore.CheckoutCart` 使用数据库事务、用户/店铺 advisory lock、店铺/商户/地址/购物车/商品行级锁完成结算，并校验商户资质仍有效，原子写入订单、订单项和订单事件并删除对应购物车行；运行时同步 `merchant_qualifications`、`merchant_products` 和 `cart_items` 规范化表，提交后刷新订单镜像、清空内存购物车和 snapshot，架构检查防止回退到内存先写。
- 已有余额支付 PostgreSQL 事务扣减首版：`PostgresStore.PayOrderWithBalance` 使用数据库事务、`pg_advisory_xact_lock` 幂等键锁、订单/钱包/钱包流水 `FOR UPDATE`、余额检查和 `wallet_transactions.idempotency_key` 唯一约束守护扣减；测试覆盖 SQL 事务提交后的团购发券、`order.paid` outbox 和跨订单/跨用户幂等键重放拒绝。
- 已有退款策略与余额退款核心闭环首版：管理员可读取/保存退款默认策略，订单退款默认退回平台余额并生成幂等退款交易、正向钱包退款流水、订单 `refunded` 状态和 `order.refunded` outbox；原路返回策略首版进入 `refund_pending` 并生成 `payment.refund.requested` outbox，等待后续微信退款 worker 接入；HTTP、BFF、领域测试和架构检查均覆盖。
- 已有 payment-worker 原路退款事件规范化首版：`payment.refund.requested` outbox 会被规范化为 `refund_requested` 任务载荷，包含订单、用户、金额、目的地、pending 状态和稳定幂等键；重复投递继续由 consumed-event ledger 防重。
- 已有退款 PostgreSQL 事务化首版：`refund_settings` 和 `refund_transactions` 已进入迁移与运行时规范表；`PostgresStore.RefundOrder` 通过事务、advisory lock、行级锁和幂等唯一约束守护余额退款/原路退款申请，余额退款会原子写入退款交易、钱包退款流水、钱包余额版本、订单状态和订单事件，提交后补齐 outbox 和 snapshot。
- 已有部分退款资金账本首版：退款累计金额按订单汇总 `success` 与 `pending_original_route` 交易，防止余额退款、原路退款、售后退款跨路径超退；部分退款保留订单原业务状态，累计退满后才进入 `refunded` 或 `refund_pending`；退款 outbox 区分本次退款金额和订单原金额，避免原路退款 worker 误退整单。
- 已有售后申请与审核核心闭环首版：用户可提交关联订单、原因、申请金额和证据附件的售后申请；用户、商户、管理员可按权限查看售后列表；商户仅能审核自有店铺订单，管理员可审核全部；审核支持通过、驳回和转平台审核，通过后复用退款幂等链路退回平台余额，并写入售后 outbox。
- 已有售后 PostgreSQL 规范化恢复首版：`order_after_sales` 已进入迁移和运行时建表，支持 `refund_only`、`partial_refund`、`food_safety` 类型，`PostgresStore` 会把售后申请同步到规范表并在启动恢复时重建售后列表和 `nextAfterSalesID`。
- 已有售后审核 PostgreSQL 事务化首版：`PostgresStore.ReviewAfterSales` 会在同一事务内锁定售后申请、订单、退款幂等键、退款配置和钱包账户，写入售后审核结果、退款交易、钱包退款流水、订单退款状态和订单事件；提交后补齐退款与售后 outbox，架构检查防止回退到内存 Store 审核。
- 已有售后仲裁与客服介入处理日志首版：`order_after_sales_events` 已进入迁移和运行时规范表；售后创建、用户补充证据、商户回复、客服介入、平台仲裁、内部备注、审核通过/驳回/升级都会记录可恢复时间线，用户和商户仅可见公开记录，管理员可见完整审计。
- 已有售后可退金额、证据上传票据、附件确认和对象存储签名配置化首版：售后申请返回订单原金额、已退金额和剩余可退金额；`POST /api/after-sales/{requestID}/evidence/upload-ticket` 会按售后权限发放有效上传票据、限制 10MB、仅允许图片/PDF 类型，并通过 `ObjectStorageConfig` 支持 MinIO provider、bucket、上传端点、CDN 端点、HMAC 签名密钥、TTL、最大大小、HEAD 校验端点、HEAD 超时、上传回调密钥、上传回调门禁和扫描门禁配置；上传票据已进入 `order_after_sales_evidence_upload_tickets` 签发账本，上传回调和扫描结果分别通过 `POST /api/object-storage/upload-callback` 与 `POST /api/object-storage/scan-result` 验签入账；`@infinitech/object-scan-worker` 已能消费 `object.uploaded` 事件、幂等回调上传状态、按签名 URL 或 `OBJECT_STORAGE_DOWNLOAD_BASE_URL` 下载对象、限制扫描大小、通过 ClamAV INSTREAM 协议映射 `passed/rejected` 结果并回调扫描状态；`@infinitech/object-lifecycle-worker` 已能读取过期未确认或扫描拒绝的清理候选，按 `OBJECT_STORAGE_DELETE_BASE_URL` 删除对象，并回写票据 `deleted` 状态、原因和删除时间；删除失败会通过 `POST /api/admin/object-storage/cleanup-failed` 回写 `cleanup_attempts`、`last_cleanup_error` 和 `last_cleanup_failed_at`，候选对象仍保留可重试；`GET /api/admin/object-storage/cleanup-stats` 可汇总 pending、expired、scan rejected、failed、deleted、累计尝试次数和最近失败/删除时间，供管理端和告警接入；`POST /api/after-sales/{requestID}/evidence/confirm` 必须匹配同售后申请、同用户/角色、同对象 key、同类型/大小且未过期的已签发票据，生产开启 `OBJECT_STORAGE_REQUIRE_HEAD_VERIFICATION=true`、`OBJECT_STORAGE_REQUIRE_UPLOAD_CALLBACK=true`、`OBJECT_STORAGE_REQUIRE_SCAN_APPROVAL=true` 后还会校验对象存在、大小、类型、上传回调和扫描通过，确认后才会把对象绑定为售后附件元数据；`GET /api/after-sales/{requestID}/evidence` 可按权限查看附件清单，`order_after_sales_evidence` 支持 PostgreSQL 恢复。当前仍待接真实 MinIO SDK、STS/Vault 临时凭证、隔离 bucket、扫描/删除失败告警投递和生产对象存储权限策略。
- 已有派单审计事件 PostgreSQL 规范化恢复首版：`PostgresStore` 启动和写入后会同步/恢复 `dispatch_events`，并重建站长派单审计、拒单/超时骑手跳过索引和 `nextDispatchEventID`；测试覆盖灾备恢复后站长查询、订单状态补偿修复骑手漂移、拒单索引恢复和后续派单事件 ID 不碰撞。
- 已有商家订单流转 PostgreSQL 事务化首版：`PostgresStore.MerchantAcceptOrder` 和 `MerchantMarkOrderReady` 使用数据库事务、`orders` 行级锁、商户归属/状态/门店启用/保证金校验，并原子写入 `order_events`；提交后刷新规范化表镜像、补齐 `order.status_changed` outbox 和 snapshot，架构检查防止回退到内存先写。
- 已有站点区域匹配首版测试：店铺归属站点后，站长订单视图、自动派单候选、站长手动派单和派单事件查询都按站点过滤，跨站点订单对非本站站长隐藏。
- 已有骑手履约首批闭环测试：骑手抢单后可确认取货进入 `picked_up`，送达后订单进入 `completed`，重复送达会被拒绝。
- 已有固定单量后免责拒派测试：骑手当日完成站点固定单量后拒绝派单会返回免责标记、当日完成数、固定单量和顺延骑手。
- 已有站长调度首批闭环测试：站长查看站点骑手、查看待调度订单、手动派单给同站点在线骑手、跨站点骑手不可见不可派。
- 已有站长任务配置首批闭环测试：读取每日任务时长/固定单量、保存配置、拒绝超过 24 小时的异常任务时长。
- 已有站长骑手绩效首批闭环测试：按站点返回骑手平均接单耗时、日均单量、完成率、等级、分数和派单优先级。
- 已有用户外卖首批闭环测试：店铺、商品、地址、购物车、结算、余额支付密码、余额支付、微信支付、订单查询、商户出餐、骑手抢单。
- 已有用户小程序首批页面和预览图。
- 已有商户端 Flutter 经营概况、订单处理、商品管理、团购核销和资质资料首批页面。
- 已有骑手端 Flutter 抢单大厅和站长工作台首批页面。
