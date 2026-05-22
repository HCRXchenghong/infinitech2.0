# 容量与容灾验收

目标：

- 10 万认证在线连接。
- 50% 活跃连接。
- 混合流量覆盖首页卡片、圈子小微墙、饭搭匹配、消息、群聊红包、订单、退款、骑手位置、抢单/派单、RTC 信令。

阶段：

- `10k`
- `30k`
- `60k`
- `100k`

每阶段必须归档：

- 拓扑图。
- 压测配置。
- 指标面板截图。
- 场景日志。
- 节点故障演练记录。
- Redis 故障切换记录。
- Kafka 堆积与恢复记录。
- outbox 死信隔离记录，包括 `dead_letter` 查询、人工 replay 解封和 `OUTBOX_RELAY_MAX_ATTEMPTS` 配置。
- outbox relay 租约领取记录，包括 `POST /api/admin/outbox/events/claim`、`leased` backlog 统计、租约过期接管和 `OUTBOX_RELAY_WORKER_ID`/`OUTBOX_RELAY_LEASE_SECONDS` 部署配置。
- outbox relay 租约续租记录，包括 `POST /api/admin/outbox/events/{eventID}/lease/renew`、错误 owner/过期租约冲突、慢发布期间心跳续租和 `OUTBOX_RELAY_LEASE_RENEW_INTERVAL_MS` 部署配置。
- PostgreSQL outbox 规范化 relay 记录，包括 `platform_outbox_events` 幂等补表、`FOR UPDATE SKIP LOCKED` 多副本领取、ack/fail/replay 写回规范化表和 snapshot 镜像同步。
- outbox 租约健康观测记录，包括 `GET /api/admin/outbox/stats?lease_expiring_within_seconds=...`、全局/per-topic/per-owner `lease_expiring_soon`、`next_lease_expires_at` 和 `next_lease_expires_in_seconds`，用于确认 relay 心跳、worker 倾斜和租约临期告警。
- 消费端幂等落库记录，包括 `platform_consumed_events`、`consumer_event_key`、`UNIQUE (consumer_name, idempotency_key)` 和 dispatch/payment/notification/integration/settlement worker 重复 outbox 投递只执行一次的测试证据。
- 支付/钱包 PostgreSQL 规范化恢复记录，包括 `orders`、`order_events`、`wallet_accounts`、`wallet_transactions`、`wallet_payment_passwords`、`payment_transactions` 的同步/恢复，以及余额支付幂等索引、支付密码和微信回调交易索引在重启后仍可用的测试证据。
- 订单创建 PostgreSQL 事务化记录，包括 `platform_sequences` 的 `orders` 序列行级锁、序列缺失时按现有 `orders` 最大编号追平、`orders` 与初始 `order_events` 原子写入，以及恢复后订单详情、用户订单索引、后续订单号和支付链路仍可继续的测试证据。
- 余额支付 PostgreSQL 事务扣减记录，包括 `pg_advisory_xact_lock` 幂等键锁、`orders`/`wallet_accounts`/`wallet_transactions` 行级锁、余额检查、钱包流水唯一幂等写入，以及 SQL 事务提交后团购发券和 `order.paid` outbox 仍可恢复的测试证据。
- 派单审计事件 PostgreSQL 规范化恢复记录，包括 `dispatch_events` 同步/恢复、站长审计查询、拒单/超时骑手跳过索引、订单状态补偿依赖派单事件修复骑手漂移，以及后续派单事件 ID 不碰撞的测试证据。
- 商家订单流转 PostgreSQL 事务化记录，包括 `orders` 行级锁、商户归属/状态/门店启用/保证金校验、`order_events` 原子写入、`order.status_changed` outbox 回灌和恢复后骑手可继续抢单的测试证据。
- 资金类幂等重放记录，包括支付回调、退款、余额红包和钱包流水。

验收阈值：

- Socket 鉴权成功率 `>= 99.95%`。
- 核心实时链路 `p95 < 150ms`。
- 写接口 `p95 < 250ms`。
- RTC 邀请送达 `p95 < 2s`。
- 资金写入幂等成功率 `100%`，不得重复扣款、重复退款或重复入账。
- 总错误率 `< 0.1%`。
- 节点恢复 `p95 < 15s`。
