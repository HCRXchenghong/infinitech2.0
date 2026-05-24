# Infinitech 2.0 最近进展与路线图

更新时间：2026-05-24
目标仓库：`https://github.com/HCRXchenghong/infinitech2.0`  
当前最新提交：以 GitHub `main` 分支最新提交为准

## 当前结论

项目已经完成架构基线、monorepo 工程骨架、首批端侧页面、核心 API 大量业务闭环、BFF 代理、Worker 骨架、PostgreSQL 规范化账本、outbox relay、对象扫描/清理、管理端 P0 运营视图、管理端审计账本、审计中心增强首版、审计服务端安全边界首版、审计完整性证明首版、管理端服务端 RBAC 策略矩阵首版、RBAC 权限治理查询与变更申请审计首版、RBAC 权限申请审批/驳回台账首版、RBAC 权限变更手动应用首版、RBAC 权限变更审计回滚首版，以及退款策略配置、管理端订单退款、售后审核、订单状态补偿、对象清理完成/失败、outbox 运维和商户/骑手邀约审计同事务首版。

但项目仍未达到商业级可上线状态。真实生产支付、微信原路退款、提现结算、真实 IM/RTC、完整后台页面、字段级/租户级权限治理、真实高可用基础设施、10 万在线压测和容灾演练还没完成。未完成这些验收前，只能说“按商业级目标推进中”，不能说“已经商业级上线”。

## 最近已完成

### 1. 管理端 P0 业务视图首版

- 已新增订单监控、售后审核、商户资质、骑手/站长、骑手绩效、派单审计、退款策略等 P0 页面结构。
- 每个视图已有关键指标、表格列、操作入口和安全约束说明。
- 管理端 Web 已从占位页推进到可操作的运营控制台基础形态。

### 2. 管理端运营快照接口

- 已新增 `GET /api/admin/operations/snapshot`。
- 后台可统一聚合订单、商户资质、保证金、骑手在线、站长、骑手绩效、售后、派单审计、退款策略、outbox 健康和对象清理统计。
- BFF 已代理该接口。

### 3. 管理端快照绑定

- Admin Web 已新增快照适配层。
- 顶部 KPI、今日队列、订单、售后、商户、骑手、骑手绩效、派单和退款策略视图可由快照响应生成。
- 后端展示字段进入页面前做 HTML 转义，降低后台页面注入风险。

### 4. BFF 浏览器 CORS 白名单

- BFF 已支持本地管理端和 uni-app 调试来源。
- 已支持 `BFF_ALLOWED_ORIGINS` 配置部署来源。
- 已明确允许 `Authorization`、`Content-Type`、`X-Client-Kind` 请求头和 `OPTIONS` 预检。

### 5. 管理端操作审计日志首版

- 已新增 `GET /api/admin/audit-logs`。
- API、BFF、Admin Web 操作台已接入。
- 已覆盖商户/骑手邀约、退款策略保存、订单退款、订单状态补偿、售后审核、对象清理完成/失败、outbox 领取/续租/发布/失败/重放/批量重放等关键后台写操作。
- 审计记录包含 actor、action、target、request_id、ip_hash、服务端白名单 payload 和 created_at。

### 6. 审计日志 PostgreSQL 规范化表

- `PostgresStore` 启动时会确保 `audit_logs` 表和 actor/action/target 查询索引存在。
- 旧快照审计会幂等补入规范化表。
- 新审计 ID 通过 `platform_sequences` 的 `audit_logs` 序列行级锁生成 `aud_N`，避免多副本 API 下 ID 碰撞。
- PostgreSQL-backed Store 查询审计时直接读取规范化 `audit_logs` 表。

### 7. 管理端审计检索页首版

- Admin Web 已新增“审计检索”P0 模块。
- 支持按 actor type、actor id、action、target type、target id、before 时间游标和 limit 查询 `/api/admin/audit-logs`。
- 支持 before 游标翻页，便于按时间倒序继续追查历史操作。
- 审计 payload 在表格中只显示白名单摘要，并对 password、secret、token、authorization、openid、phone、object key、签名等敏感字段做脱敏详情展示。
- RBAC 首版矩阵已从安全审计员 `security_auditor` 扩展到运营、财务、调度和客服等后台分权角色，并与服务端 scope 对齐。

### 8. 管理端审计中心增强首版

- `/api/admin/audit-logs` 已新增 `after` 时间下界，内存 Store 和 PostgreSQL 查询均按 `created_at >= after` 过滤，`before` 继续作为严格小于的时间游标。
- Admin Web 审计检索已支持 after/before 时间范围、保存常用筛选、审计详情抽屉、按当前目标继续筛选和 before 游标翻页。
- 审计目标可根据 target type/action 跳到订单、售后、商户、骑手、outbox、对象清理和退款策略等相关运营模块。
- Admin Web 单测、Go Store/HTTP/PostgreSQL 查询构造测试和架构守卫已覆盖 after、预设、详情和跳转能力。

### 9. 管理端审计服务端安全边界首版

- `api-go` 新增 `security_auditor` 服务端角色，`Principal.CanReadAuditLogs()` 只允许管理员和安全审计员读取 `/api/admin/audit-logs`。
- 安全审计员可以只读审计账本，但不能执行商户邀请、订单状态补偿等后台写操作；HTTP 回归测试已覆盖允许读和拒绝写。
- `auth_identities`、`auth_sessions` 迁移和运行时建表已允许 `security_auditor` 主体类型，避免生产 session 落库失败。
- 审计 payload 白名单和敏感字段掩码已下沉到后端 Store/PostgreSQL 路径；写入、SQL marshal、SQL 读取和镜像恢复都会调用统一 `sanitizeAuditPayload`。
- 白名单字段如 `default_refund_strategy`、`amount_fen` 会保留；`object_key` 等敏感允许字段会掩码；`password`、`token`、`phone`、`nested`、`raw_request` 等非白名单或敏感字段不会持久化或原样返回。

### 10. 管理端审计完整性证明首版

- `audit_logs` 迁移和运行时建表已新增 `integrity_algorithm` 与 `integrity_hash`，API 合同返回 `integrity_verified`。
- 内存 Store 与 PostgreSQL 写入会对审计 ID、actor、action、target、request_id、ip_hash、服务端白名单 payload 和 `created_at` 生成稳定完整性证明；查询时重新验证并暴露验证结果。
- 本地和无密钥环境默认使用 `sha256:v1`；生产配置 `AUDIT_LOG_SIGNING_SECRET` 后使用 `hmac-sha256:v1`，可发现数据库中审计字段或白名单 payload 被篡改。
- 旧快照或旧 SQL 行缺少完整性字段时，可在业务字段完全一致的前提下回填证明；Admin Web 审计中心已展示完整性状态、算法和哈希。
- 这仍不是最终法律级不可抵赖方案，后续还要补 KMS/Vault 密钥轮换、链式账本、WORM/冷归档、导出留存和告警。

### 11. 后台关键写路径审计同事务首版

- `PUT /api/admin/refund-settings` 已从 HTTP 层“先写业务、再补审计”迁移到仓储级 `SaveRefundSettingsWithAudit` 原子路径。
- `POST /api/orders/{orderID}/refund` 已从 HTTP 层“先退款、再补审计”迁移到仓储级 `RefundOrderWithAudit` 原子路径。
- `POST /api/after-sales/{requestID}/review` 已从 HTTP 层“先审核售后、再补审计”迁移到仓储级 `ReviewAfterSalesWithAudit` 原子路径。
- `POST /api/admin/orders/{orderID}/state/compensate` 已从 HTTP 层“先补偿状态、再补审计”迁移到仓储级 `CompensateOrderStateWithAudit` 原子路径。
- `POST /api/admin/object-storage/cleanup-complete` 与 `POST /api/admin/object-storage/cleanup-failed` 已从 HTTP 层“先回写清理结果、再补审计”迁移到仓储级 `CompleteObjectStorageCleanupWithAudit` 与 `RecordObjectStorageCleanupFailureWithAudit` 原子路径。
- `POST /api/admin/outbox/events/claim`、`/lease/renew`、`/published`、`/failed`、`/{eventID}/replay` 和 `/api/admin/outbox/events/replay` 已从 HTTP 层“先改 outbox、再补审计”迁移到仓储级 `ClaimOutboxEventsWithAudit`、`RenewOutboxEventLeaseWithAudit`、`MarkOutboxEventPublishedWithAudit`、`MarkOutboxEventFailedWithAudit`、`ReplayOutboxEventWithAudit` 和 `ReplayOutboxEventsWithAudit` 原子路径。
- `POST /api/admin/merchant-invites`、`POST /api/admin/rider-invites` 与 `POST /api/station-manager/rider-invites` 已从 HTTP 层“先创建邀约、再补审计”迁移到仓储级 `CreateMerchantInviteWithAudit` 与 `CreateRiderInviteWithAudit` 原子路径。
- 内存 Store 在同一把业务锁内完成退款策略保存、订单退款、售后审核、订单状态补偿、对象清理票据状态更新、outbox 运维状态变更或商户/骑手邀约创建并写入对应审计；PostgreSQL-backed Store 在同一个数据库事务内写入业务表、`platform_outbox_events` 或包含邀约状态的 `platform_store_snapshots` 与 `audit_logs`，审计 ID 继续走 `platform_sequences` 行级锁。
- 退款策略审计 payload 以服务端规范化后的 `default_refund_strategy` 为准；订单退款审计 payload 以最终退款交易为准，只保留 `refund_id`、`destination`、`status`、`amount_fen` 和 `idempotency_key`；售后审核审计 payload 以最终审核结果和退款交易为准，只保留 `decision`、`status`、`refund_id`、`amount_fen`、`destination` 和 `idempotency_key`；订单状态补偿审计 payload 以最终补偿结果为准，只保留 `changed`、`previous_status`、`expected_status`、`compensation_type`、`evidence_count` 和必要 rider 字段；对象清理审计 payload 以最终票据状态为准，只保留脱敏 `object_key`、`reason`、`status` 和失败时的 `cleanup_attempts`。
- outbox 运维审计 payload 以最终 outbox 事件或批量结果为准，只保留 `topic`、`status`、`attempts`、`lease_owner`、`lease_seconds`、`retry_after_seconds`、`claimed`、`replayed` 和 `limit` 等白名单字段。
- 商户/骑手邀约审计 payload 以最终生成的邀约为准，只保留 `type`、`expires_at`，骑手邀约额外保留 `station_id`；调用方伪造 token、station 或过期时间不会混入审计。
- HTTP 防回退测试、Store 原子审计测试和架构守卫已固定这些路径，避免未来退回到业务写成功后再单独调用 `RecordAuditLog`。
- 当前只是后台关键写路径与审计同事务首版，后续仍需继续扫描其他后台写路径、导出留存、异常告警和不可抵赖归档。

### 12. 管理端服务端 RBAC 策略矩阵首版

- `api-go` 已把后台角色从单一 `admin`/`security_auditor` 扩展为 `admin`、`super_admin`、`ops_admin`、`finance_admin`、`dispatch_admin`、`support_admin`、`security_auditor`。
- 已新增服务端 scope 矩阵：邀约、退款、运营快照、售后读取/审核/事件、对象清理、outbox 运维、订单状态补偿、调度读写、钱包/结算只读、审计只读等权限不再只靠页面草案约束。
- 关键后台路由已从硬编码 `IsAdmin()` 推进到 `CanManageInvites()`、`CanManageRefunds()`、`CanReadAdminAfterSales()`、`CanManageDispatch()`、`CanReadOutbox()`、`CanManageOutbox()` 等服务端权限判断。
- `security_auditor` 继续只能读审计账本；`finance_admin` 可读写退款策略和退款动作但不能创建邀约；`ops_admin` 可做运营/邀约/售后审核/对象清理/outbox 运维但不能读审计账本；`dispatch_admin` 可看和管理派单但不能操作 outbox；`support_admin` 可读售后并追加客服事件但不能审批退款。
- `auth_sessions` 运行时表和认证迁移已允许新增后台主体类型；签名 token 测试覆盖新增后台角色。
- Admin Web 的 RBAC 配置已同步服务端 scope 命名，架构守卫和 HTTP 回归测试固定角色边界。
- 当前仍是服务端 RBAC 首版，后续还要补角色/权限审批与应用页面、字段级权限、租户/站点数据域、审批流和更完整的后台菜单隐藏策略。

### 13. RBAC 权限治理查询与变更申请审计首版

- 已新增 `GET /api/admin/rbac/policy`，后台分权角色可读取服务端真实 RBAC 矩阵、scope 风险等级、当前角色权限、策略版本和变更模型。
- 已新增 `POST /api/admin/rbac/change-requests`，仅 `admin`/`super_admin` 可提交权限变更申请；服务端校验目标角色、scope 白名单、申请原因和 `*` 权限范围。
- 权限变更申请会写入 `admin.rbac.change_requested` 审计日志，payload 只保留变更申请 ID、目标角色、当前 scopes、申请 scopes、策略版本、原因和 `pending_approval` 状态。
- BFF 已代理权限策略查询和变更申请接口。
- Admin Web 已新增“权限治理”模块，操作台支持读取策略和提交 CSV scopes 申请；审计中心可把 `admin_rbac_role` 目标跳转到权限治理视图。
- 申请阶段不会直接修改运行时权限，必须经过审批和单独应用动作，避免误改生产权限。

### 14. RBAC 权限申请审批/驳回台账首版

- 已新增 `GET /api/admin/rbac/change-requests`，可从规范化审计账本重建 RBAC 申请列表，支持按 `pending_approval`、`approved`、`rejected` 状态筛选。
- 已新增 `POST /api/admin/rbac/change-requests/{id}/review`，`admin`/`super_admin` 可审批或驳回权限申请。
- 服务端禁止提交人审批自己的申请，避免单人自提自批高危权限。
- 审批和驳回会写入 `admin.rbac.change_reviewed` 审计，审计目标为 `admin_rbac_change_request`，记录 decision、status、role、requested scopes、policy version 和原因。
- BFF 已代理申请列表和审批接口；Admin Web 操作台新增权限申请列表与审批操作，审计中心可跳转到权限治理模块。
- 审批结果不会自动改运行时权限，必须由具备权限的管理员走单独应用动作。

### 15. RBAC 权限变更手动应用首版

- 已新增 `POST /api/admin/rbac/change-requests/{id}/apply`，只允许状态为 `approved` 的申请进入应用阶段。
- 服务端禁止申请人直接应用自己的申请，继续保持至少两人参与高危权限变更。
- 应用动作会写入 `admin.rbac.change_applied` 审计，记录 change request、目标角色、应用 scopes、应用前 scopes、策略版本和原因。
- `api-go` 启动时会从 `admin.rbac.change_applied` 审计日志重放已应用策略，运行时 `HasAdminScope` 和 `AdminScopesForRole` 会读取应用后的权限矩阵。
- BFF 和 Admin Web 操作目录已接入应用入口，架构守卫和 HTTP 回归测试固定应用后权限即时生效和重启恢复。

### 16. RBAC 权限变更审计回滚首版

- 已新增 `POST /api/admin/rbac/change-requests/{id}/rollback`，只允许当前仍处于 `applied` 状态的申请回滚。
- 服务端禁止申请人直接回滚自己的申请，继续保持高危权限变更至少两人参与。
- 回滚前会校验该申请仍是目标角色最新一次生效策略事件，且当前运行时 scopes 仍等于该申请应用后的 scopes，避免覆盖后续策略变更。
- 回滚动作会写入 `admin.rbac.change_rolled_back` 审计，记录 change request、目标角色、回滚前 scopes、回滚到 scopes、策略版本和原因。
- `api-go` 启动时会按 `admin.rbac.change_applied` 与 `admin.rbac.change_rolled_back` 审计时间顺序重放运行时策略，避免重启后丢失回滚结果。
- BFF 和 Admin Web 操作目录已接入回滚入口，架构守卫和 HTTP 回归测试固定回滚后权限恢复、重启恢复和状态筛选。

## 当前未完成

### P0 商业级阻塞项

- 真实微信支付生产参数、证书、回调、验签、异常重放和对账闭环。
- 微信原路退款 API 真调用、退款回调、对账和差错处理。
- 钱包提现、商户结算、骑手收入、平台抽佣和财务报表。
- 字段级/租户级 RBAC、权限变更产品化审批页面、审计导出/留存/告警、KMS/链式不可抵赖签名、冷热归档和审计策略治理。
- 剩余关键业务写操作与审计写入同事务强制提交，继续扫描后台配置、运营处置、资金和风控写路径。
- 真实 IM 消息落库、离线补偿、客服工作台和消息风控。
- RTC 语音通话信令、通话审计和订单/会话关联。
- PostgreSQL HA、Redis Cluster、Kafka 多 Broker、MinIO 生产权限、Vault/KMS 密钥治理。
- 10 万在线连接压测、订单写入压测、抢单冲突压测、故障注入和容灾报告。

### P1 产品闭环缺口

- 用户端邀请页、邀请奖励、钱包账单、售后入口、评价、收藏、积分会员、搜索、定位、消息页。
- 商户端完整邀请注册 UI、店铺装修、店铺展示页管理、商户钱包、结算、消息页、资质过期强弹窗。
- 骑手端完整任务详情、地图导航、轨迹、收入、钱包、账单、提现、违规申诉。
- 优惠券资金责任闭环：商户自发券、平台补贴券、商户确认活动券、结算审计。
- 群聊、红包、拼手气红包、官方群默认静音、商户群领券限制的 API 和页面闭环。
- 圈子/小微墙、饭搭问卷、真实性承诺、免责协议、举报拉黑和风控审核。
- 买药完整药房/医务室资质、商品、监管审计和履约。
- 快递/跑腿完整下单、计价、异常处理和骑手履约。

### P2 工程与运营缺口

- 真实 Kafka/NATS broker 运维和 relay 积压恢复演练。
- 通知通道真实接入：微信订阅消息、短信、站内信、Push。
- 对象存储接真实 MinIO SDK、STS/Vault 临时凭证、隔离 bucket、扫描/删除失败告警。
- 管理端首页卡片、精选商品、首页活动、圈子饭搭、红包、开放平台、系统日志等完整后台面板。
- 自动化发布、镜像构建、灰度、回滚、release/tag 策略和生产 runbook。

## 下一批优先推进顺序

### 第一批：后台审计中心补全

- 继续把剩余关键业务写操作和 `audit_logs` 写入推进到同一业务事务内强制提交，退款策略配置、管理端订单退款、售后审核、订单状态补偿、对象清理完成/失败、outbox 运维和商户/骑手邀约已完成首版，下一步继续扫后台配置、运营处置、资金和风控写路径。
- 在服务端 RBAC 策略矩阵与变更申请/审批/应用/回滚审计首版基础上继续补字段级权限、站点/商户数据域、产品化审批队列和菜单隐藏策略。
- 补导出、留存策略、异常告警、KMS/链式不可抵赖签名和冷热归档设计。

### 第二批：管理端详情页与审核闭环

- 订单详情抽屉。
- 售后审核详情。
- 商户资质审核详情。
- 骑手/站长管理详情。
- 退款策略和 outbox 运维的二次确认与结果追踪。

### 第三批：真实资金链路

- 接微信支付生产参数和证书配置。
- 接微信原路退款 API 和退款回调。
- 做支付/退款对账 worker。
- 做商户结算、骑手收入、平台抽佣和财务报表。

### 第四批：实时消息与 RTC

- IM 消息持久化、会话、已读、离线补偿。
- 用户与骑手、用户与商家、用户与客服、商家与骑手消息链路。
- RTC 信令、呼叫状态、通话审计、通话风控。

### 第五批：10 万在线验收

- 明确压测环境和数据规模。
- 压测 realtime-gateway 在线连接、消息投递、骑手位置、RTC 信令。
- 压测订单创建、余额支付、抢单冲突、派单转派、outbox 积压恢复。
- 做 API 节点宕机、Realtime 节点宕机、Redis failover、数据库只读、Kafka 堆积恢复演练。
- 输出容量与容灾报告。

## 最近验证证据

- `cd services/api-go && go test -count=1 ./...`
- `npm run verify`
- `npm run verify:architecture`

## 当前风险提醒

- 当前代码里很多能力已经有首版和测试，但仍有大量链路是“骨架/首版/模拟接入”状态。
- 钱包、支付、退款、结算、审计、消息、RTC、容灾压测是商业化风险最高区域。
- 未完成真实支付、真实高可用、真实压测和生产演练前，不应对外承诺“已支撑 10 万在线”或“可商业上线”。
