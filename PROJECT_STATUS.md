# Infinitech 2.0 当前状态总览

更新时间：2026-05-24
目标仓库：`https://github.com/HCRXchenghong/infinitech2.0`  
当前结论：项目已经完成架构基线、monorepo 骨架、首批端侧页面、核心 API 大量业务闭环、BFF 代理、Worker 骨架、PostgreSQL 规范化、outbox/对象存储、管理端审计、服务端 RBAC 策略矩阵和 RBAC 权限治理查询/变更申请审计首版等多条商业化底座链路；但还没有完成真实生产支付、真实 IM/RTC、完整管理端、真实高可用基础设施、10 万在线压测和容灾演练，所以不能宣称已经商业级可上线，只能说正在按商业级标准推进。

最近完成、当前未完成和下一批优先级已汇总到 `docs/product/recent-progress-roadmap.md`。这份文档用于快速查看最近提交后的项目状态、商业级阻塞项和后续推进顺序。

## 0. 最近进展摘要

最近已完成：

- 管理端 P0 业务视图首版。
- 管理端运营快照接口 `/api/admin/operations/snapshot`。
- 管理端快照数据绑定到 KPI、队列和 P0 表格。
- BFF 浏览器 CORS 白名单和 `OPTIONS` 预检。
- 管理端操作审计日志 `/api/admin/audit-logs`。
- 审计日志 PostgreSQL `audit_logs` 规范化表、查询索引、旧快照回填和 `platform_sequences` 行级锁发号。
- 管理端审计中心增强首版，支持 actor/action/target/after/before/limit 筛选、before 游标翻页、保存筛选、详情抽屉、跨模块跳转和脱敏 payload 摘要。
- 管理端审计服务端安全边界首版，新增 `security_auditor` 只读审计角色，并把 audit payload 白名单/敏感字段掩码下沉到 Store 与 PostgreSQL 路径。
- 管理端审计完整性证明首版，审计日志返回 `integrity_algorithm`、`integrity_hash`、`integrity_verified`，本地默认 `sha256:v1`，生产配置 `AUDIT_LOG_SIGNING_SECRET` 后使用 `hmac-sha256:v1` 检测审计字段或白名单 payload 篡改。
- 管理端服务端 RBAC 策略矩阵首版，新增 `super_admin`、`ops_admin`、`finance_admin`、`dispatch_admin`、`support_admin` 等后台角色和服务端 scope，邀约、退款、售后、对象清理、outbox、调度、运营快照和审计入口已按权限边界守护。
- RBAC 权限治理查询与变更申请审计首版，新增 `/api/admin/rbac/policy` 和 `/api/admin/rbac/change-requests`，BFF 与 Admin Web “权限治理”模块已接入；权限变更申请只写入审计并保持 `pending_approval`，当前不会自动修改运行时权限。
- 退款策略配置、管理端订单退款、售后审核、订单状态补偿、对象清理完成/失败、outbox 运维与商户/骑手邀约审计同事务首版，`PUT /api/admin/refund-settings` 通过 `SaveRefundSettingsWithAudit` 同事务更新退款策略并写入审计，`POST /api/orders/{orderID}/refund` 通过 `RefundOrderWithAudit` 同事务写入退款业务账本和 `admin.order.refunded` 审计，`POST /api/after-sales/{requestID}/review` 通过 `ReviewAfterSalesWithAudit` 同事务写入售后审核、必要退款和 `after_sales.reviewed` 审计，`POST /api/admin/orders/{orderID}/state/compensate` 通过 `CompensateOrderStateWithAudit` 同事务写入状态补偿结果和 `admin.order_state.compensated` 审计，`POST /api/admin/object-storage/cleanup-complete` 与 `POST /api/admin/object-storage/cleanup-failed` 分别通过 `CompleteObjectStorageCleanupWithAudit`、`RecordObjectStorageCleanupFailureWithAudit` 同事务写入对象清理结果和审计，outbox claim/lease renew/publish/fail/replay/batch replay 分别通过 `ClaimOutboxEventsWithAudit`、`RenewOutboxEventLeaseWithAudit`、`MarkOutboxEventPublishedWithAudit`、`MarkOutboxEventFailedWithAudit`、`ReplayOutboxEventWithAudit`、`ReplayOutboxEventsWithAudit` 同事务更新 `platform_outbox_events` 和 `audit_logs`；商户/骑手/站长邀约创建通过 `CreateMerchantInviteWithAudit`、`CreateRiderInviteWithAudit` 同事务写入最终邀约和审计。

当前最重要的未完成项：

- 管理端角色/权限审批与应用 UI、字段级/租户级 RBAC、权限变更双人审批、审计导出、留存、异常告警、KMS/链式不可抵赖签名和审计策略治理。
- 剩余关键业务写操作与审计写入同事务强制提交，继续扫描后台配置、运营处置、资金和风控写路径。
- 真实微信支付、微信原路退款、对账、提现、商户结算和骑手收入。
- 真实 IM、客服工作台、RTC 信令与通话审计。
- 生产高可用基础设施和 10 万在线压测/容灾报告。

下一批计划：

- 第一批补后台审计中心、RBAC 治理审批/应用 UI 和字段级/租户级权限。
- 第二批补管理端订单/售后/商户资质/骑手站长详情页。
- 第三批补真实资金链路。
- 第四批补 IM 与 RTC。
- 第五批做 10 万在线容量和容灾验收。

## 1. 架构现状

架构名称：自建/混合云 Kubernetes 上的模块化核心 API + 事件驱动 Worker + 多端 BFF + 实时网关架构。

当前架构原则已经写入 `PLATFORM_MASTER_PLAN.md`：

- `api-go` 是模块化核心 API，订单、支付、钱包、抢单、派单等关键链路优先保持强一致事务边界。
- BFF 面向微信小程序、商户 uni-app、骑手 uni-app、管理端做聚合与端侧差异适配。
- Worker 负责 outbox relay、调度、支付、通知、集成、结算、对象扫描、对象生命周期清理等异步任务。
- `realtime-gateway` 独立承载 WebSocket-only 实时能力，但目前还没有完成完整 IM 落库、离线补偿和 RTC 信令闭环。
- 数据侧以 PostgreSQL 为主库，Redis/Kafka/MinIO/Vault/Prometheus/Grafana/Loki/Tempo/OpenTelemetry 已进入规划和部分部署骨架。
- 10 万在线只是目标架构口径，未完成压测与容灾报告前不得宣称已支撑 10 万同时在线。

## 2. 仓库结构

当前 monorepo 已建立：

- `apps/user-wechat-miniprogram`：用户端原生微信小程序。
- `apps/merchant-uni`：商户端 uni-app。
- `apps/rider-uni`：骑手端 uni-app。
- `apps/admin-web`：桌面管理端 Web，已完成最小运营控制台首版。
- `apps/admin-uni`：移动管理端 uni-app，目前是骨架。
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
- 已持续记录执行台账：`EXECUTION_LEDGER.md`，当前记录到 `DONE-20260524-098`。
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

### 3.4 商户端 uni-app

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

### 3.5 骑手端 uni-app

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
- 管理端已新增审计服务端安全边界首版，`security_auditor` 可只读审计账本但不能执行后台写操作；`auth_sessions` 与身份迁移允许该主体类型；审计 payload 在服务端白名单过滤后才写入或返回，`object_key` 等敏感允许字段会被掩码，password、token、phone、nested/raw_request 等非白名单或敏感字段会被丢弃。
- 管理端已新增审计完整性证明首版，`audit_logs` 表和 API 返回 `integrity_algorithm`、`integrity_hash`、`integrity_verified`；内存 Store 与 PostgreSQL 写入会签封规范化审计字段和服务端白名单 payload，查询时验证是否被篡改；Admin Web 审计中心可展示完整性状态、算法和哈希。
- 管理端已新增退款策略配置、管理端订单退款、售后审核、订单状态补偿、对象清理完成/失败、outbox 运维与商户/骑手邀约审计同事务首版，HTTP 退款策略保存入口改走 `SaveRefundSettingsWithAudit`，管理端订单退款入口改走 `RefundOrderWithAudit`，售后审核入口改走 `ReviewAfterSalesWithAudit`，订单状态补偿入口改走 `CompensateOrderStateWithAudit`，对象清理完成/失败入口改走 `CompleteObjectStorageCleanupWithAudit` 与 `RecordObjectStorageCleanupFailureWithAudit`，outbox 运维入口改走 `ClaimOutboxEventsWithAudit`、`RenewOutboxEventLeaseWithAudit`、`MarkOutboxEventPublishedWithAudit`、`MarkOutboxEventFailedWithAudit`、`ReplayOutboxEventWithAudit` 和 `ReplayOutboxEventsWithAudit`，商户/骑手邀约入口改走 `CreateMerchantInviteWithAudit` 与 `CreateRiderInviteWithAudit`；PostgreSQL-backed Store 分别使用单个数据库事务同时写入业务表、`platform_outbox_events` 或邀约快照与 `audit_logs`，并由 HTTP 防回退测试、Store 原子审计测试和架构守卫固定路径。
- 管理端已新增服务端 RBAC 策略矩阵首版，后台角色包含兼容 `admin`、`super_admin`、`ops_admin`、`finance_admin`、`dispatch_admin`、`support_admin` 和 `security_auditor`；邀约、退款、运营快照、售后、对象清理、outbox、订单状态补偿、派单读写和审计读取已改为服务端 scope 判断，Admin Web RBAC 配置与后端 scope 命名保持一致。
- 管理端已新增 RBAC 权限治理查询与变更申请审计首版：`GET /api/admin/rbac/policy` 返回服务端真实 RBAC 矩阵、策略版本和当前角色能力；`POST /api/admin/rbac/change-requests` 仅允许 `admin`/`super_admin` 提交变更申请，申请写入 `admin.rbac.change_requested` 审计并保持待审批，不自动生效；BFF 与 Admin Web “权限治理”模块已接入。
- BFF 已补浏览器来源 CORS 白名单和 `OPTIONS` 预检处理，默认覆盖本地管理端/uni 调试来源，并可通过 `BFF_ALLOWED_ORIGINS` 配置部署来源。
- `apps/admin-uni` 骨架。
- `packages/admin-core` 已定义关键运营模块。
- 核心后台 API 已覆盖很多运营动作：退款策略、售后、outbox、对象存储清理、派单事件、站长任务、骑手绩效等。

未完成：

- 桌面管理端完整业务页面和详情页。
- 管理端 P0 视图已能读取运营快照生成首批表格/指标，关键写操作已有审计账本、审计检索页、服务端 RBAC 策略矩阵、RBAC 查询/变更申请审计和完整性证明首版；仍需补订单/售后/资质详情抽屉、审核表单、角色/权限审批与应用 UI、字段级/租户级权限、审计导出/留存/告警、KMS/链式不可抵赖签名。
- 移动管理端实际页面。
- 角色/权限审批与应用 UI、字段级/租户级 RBAC、权限变更双人审批、审计导出留存、异常告警、KMS/链式不可抵赖签名和审计策略治理。
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
- 管理端服务端 RBAC 策略矩阵与权限治理查询/申请审计首版：新增后台角色和 scope 常量，核心后台路由已按 `CanManageInvites`、`CanManageRefunds`、`CanReadAdminAfterSales`、`CanManageDispatch`、`CanReadOutbox`、`CanManageOutbox` 等服务端策略守护，`GET /api/admin/rbac/policy` 可读取真实矩阵，`POST /api/admin/rbac/change-requests` 可写入待审批申请审计，`auth_sessions` 和迁移允许新增后台主体类型。
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
- 角色/权限审批与应用 UI、字段级/租户级 RBAC、剩余业务写操作与审计写入同事务强制提交、审计导出/留存/告警、KMS/链式签名和完整审计后台。

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
- `notification-worker`：订单和消息事件通知 payload 骨架，消费幂等。
- `integration-worker`：OAuth/API provider 配置和同步事件骨架，消费幂等。
- `settlement-worker`：完成订单结算计算骨架，金额整数分。
- `object-scan-worker`：对象上传事件、下载、大小限制、ClamAV INSTREAM、回调签名、消费幂等。
- `object-lifecycle-worker`：清理候选读取、对象删除、404 幂等成功、删除完成回写、删除失败回写。

未完成：

- 真实 Kafka/NATS broker 运维接入和生产拓扑。
- 通知通道真实接入：微信订阅消息、短信、站内信、Push。
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
- 后台权限配置 UI、字段级/租户级 RBAC、操作审计导出留存和敏感字段脱敏治理。
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
- 已做审计中心增强首版：actor/action/target/after/before/limit 筛选、before 游标翻页、保存筛选、详情抽屉、跨模块跳转和脱敏 payload 摘要。
- 已做审计服务端安全边界首版：`security_auditor` 只读审计角色、审计 payload 服务端白名单和敏感字段掩码。
- 已做审计完整性证明首版：`sha256:v1`/`hmac-sha256:v1` 签封审计规范化字段和白名单 payload，Admin Web 可展示验证状态。
- 已做退款策略配置、管理端订单退款、售后审核、订单状态补偿、对象清理完成/失败、outbox 运维与商户/骑手邀约审计同事务首版：后台退款策略保存会在仓储级原子路径内同时更新配置和写入审计，管理端订单退款会在仓储级原子路径内同时写入退款业务账本和审计，售后审核会在仓储级原子路径内同时写入审核结果、必要退款和审计，订单状态补偿会在仓储级原子路径内同时写入修复结果和审计，对象清理完成/失败会在仓储级原子路径内同时写入上传票据清理状态和审计，outbox 运维会在仓储级原子路径内同时更新 outbox 事件状态和审计，商户/骑手邀约会在仓储级原子路径内同时生成最终邀约和审计。
- 已做管理端服务端 RBAC 策略矩阵首版：`ops_admin`、`finance_admin`、`dispatch_admin`、`support_admin`、`security_auditor` 等角色已由服务端 scope 守护关键后台路由。
- 已做 RBAC 权限治理查询与变更申请审计首版：后台可读取真实服务端矩阵，`admin`/`super_admin` 可提交待审批权限申请并写入审计，当前不自动生效。
- 下一步：把订单/售后/商户/骑手视图继续拆详情页与审核表单，并补角色/权限审批与应用 UI、字段级/租户级 RBAC、剩余后台配置/运营处置/资金风控写路径审计同事务、审计导出留存、异常告警和 KMS/链式不可抵赖签名。

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
