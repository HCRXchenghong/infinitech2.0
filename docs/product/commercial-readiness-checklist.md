# Infinitech 2.0 商业级验收清单

更新时间：2026-05-22

商业级不是“页面看起来完整”，而是关键业务在真实用户、真实支付、真实履约、真实故障下还能可靠运行。未完成压测和容灾前，只能说“按商业化目标建设中”。

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
- 用户端原生微信小程序。
- 商户端独立 uni-app。
- 骑手端独立 uni-app。
- 管理端桌面 Web 和移动管理 uni-app。
- UI 使用旧版 logo 和 `#009bf5`。
- 核心流程必须有真机/浏览器截图验收。
- 端侧只访问 BFF 或公开 API，不直连内部 Worker 和数据库。
- 当前证据：`apps/admin-web` 已有最小运营控制台首版，接入管理员登录、邀约准入、退款策略、售后列表、对象清理、outbox 运维和订单状态补偿等 BFF/API；已补订单监控、售后审核、商户资质、骑手/站长、骑手绩效、派单审计、退款策略 P0 业务视图首版；`npm run test --workspace @infinitech/admin-web` 和浏览器打开验证已通过。仍需接真实列表 API、完整详情页、细分 RBAC、操作审计、敏感字段脱敏和截图归档。

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
- 已有商户员工健康证和补充资料首批闭环测试：商户可读取/提交员工健康证与门头/后厨等补充资料，保存时校验失效日期和店铺归属，跨商户店铺写入会被隐藏拒绝；BFF 已覆盖 Authorization 转发，商户端 uni-app 已新增资质资料页。
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
- 已有商户端 uni-app 经营概况、订单处理、商品管理、团购核销和资质资料首批页面。
- 已有骑手端 uni-app 抢单大厅和站长工作台首批页面。
