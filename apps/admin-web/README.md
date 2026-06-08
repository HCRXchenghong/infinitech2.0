# 管理端 Web

桌面后台首期覆盖订单、商户、骑手、商品、团购、买药、跑腿、支付、钱包、调度、计价、客服、RTC 和集成配置。当前已从占位页推进到可打开的最小运营控制台，入口是 `index.html`，默认连接本地 BFF `http://localhost:25500`。

当前首版能力：

- 登录操作台：调用 `/api/auth/admin/login`，返回 token 后可填入当前后台会话。
- 商户/站长/骑手邀约：调用管理员邀请接口，支撑邀约制准入。
- 运营快照：调用 `/api/admin/operations/snapshot` 聚合订单、商户、骑手、骑手配送评分样本、售后、派单、退款策略、outbox 和对象清理状态。
- 操作审计：调用 `/api/admin/audit-logs` 查询关键后台写操作审计账本。
- 审计检索：按 actor、action、target、after/before 时间范围和 limit 查询审计记录，支持保存筛选、详情抽屉、按目标筛选和跨模块跳转，并以白名单摘要、脱敏详情和完整性状态展示 payload。
- 商户资质工作台：调用 `GET /api/admin/merchant-qualifications` 和 `GET /api/admin/merchant-qualifications/{qualificationID}` 查看待审资质、商户/店铺/保证金上下文、最近审核审计、推荐下一步动作和核查清单，再进入资质审核二次确认。
- 通知运营：调用 `GET /api/admin/notifications`、`GET /api/admin/notification-deliveries`、`GET/PUT /api/admin/notification-preferences`、`POST /api/admin/notification-preferences/batch`、`GET/POST /api/admin/notification-preferences/change-requests`、`POST /api/admin/notification-preferences/change-requests/{id}/review`、`POST /api/admin/notification-preferences/change-requests/{id}/apply`、`POST /api/notifications/{notificationID}/deliveries`、`POST /api/admin/notification-deliveries/failure-alerts/emit`、`POST /api/admin/notification-deliveries/retries/schedule` 和 `POST /api/admin/notification-deliveries/quiet-window-retries/schedule`，按商户、状态、来源事件、渠道和 provider 查询通知/回执/偏好，在需要时补录高风险投递回执、保存单条外部触达偏好、批量保存多条偏好策略，或把关键偏好变更走申请、审批、手动应用闭环，并把 failed 或到期 quiet-window queued 回执汇总投递到可靠告警/重试 outbox；notification-worker provider 执行器首版会把配置渠道的 delivered/failed/queued 回执写回同一台账。
- 退款策略和订单退款：读取和保存默认退余额/默认原路返回策略；订单模块可通过 `POST /api/orders/{orderID}/refund` 发起退款，后端保存和退款路径已把业务写入与审计写入收敛到同一仓储事务。
- 售后审核：读取后台售后列表，并可通过 `POST /api/after-sales/{requestID}/review` 提交审核结果。
- 客服工作台：调用 `GET /api/admin/service-tickets` 查看客服工单列表，支持按用户、状态、SLA 状态、客服筛选；可通过 `POST /api/admin/service-tickets/{ticketID}/assign` 分派客服，通过 `POST /api/admin/service-tickets/{ticketID}/escalate` 升级超时工单，通过 `POST /api/admin/service-tickets/{ticketID}/resolve` 提交处理方案；客服质检/绩效已接 `POST /api/admin/service-tickets/{ticketID}/quality-review`、`GET /api/admin/service-ticket-quality-reviews`、`GET /api/admin/service-ticket-performance`，分派、升级、方案提交和质检均进入高风险二次确认。
- Outbox 运维：查看事件、单事件事故辅助明细、健康统计、死信分诊、领取租约、续租、死信解封、单事件恢复、标记失败、标记已发布和批量恢复入口。
- 对象存储清理：查看清理统计和候选对象。
- 订单状态补偿：按订单号触发后台补偿接口；后端补偿路径已把状态修复结果与审计写入收敛到同一仓储事务。
- 运营首页：订单、售后、骑手、资质、outbox、对象清理的 P0 指标位。
- P0/P1 业务视图：订单监控、售后审核、商户资质、骑手/站长、骑手绩效、派单审计、审计检索、退款策略、客服工作台、通知运营；有管理员 token 时会用运营快照生成 P0 指标和表格，表格行可筛选分页、打开详情面板并跳到补偿、审计、outbox、对象清理、通知回执、通知失败告警、通知失败重试、客服分派、SLA 升级、工单方案、客服质检、客服绩效和 RBAC 等下一步操作。
- 骑手绩效详情：运营快照里的 `rider_performance` 已支持 `score_breakdown`、`recent_trend`、`recent_reviews`、`exception_summary` 和 `exception_details`，详情抽屉会展示最近 3 天趋势、最新评价摘录、异常履约摘要，并给出查看派单事件、售后记录和订单审计的预填动作，方便解释派单优先级变化。
- 售后详情动作：售后审核详情抽屉和骑手异常履约明细都已补出 `after-sales-events`、`after-sales-evidence` 和 `dispatch-order-events` 的预填动作，运营可以直接回看售后时间线、凭证和关联派单事件。
- 售后聚合详情：后台现已补出 `GET /api/admin/after-sales/{requestID}`，把工单上下文、时间线、凭证和派单关联一次性拉回；售后抽屉和骑手异常履约里的售后异常都可以直接跳这个聚合详情。
- 售后资金/客服深链：售后聚合详情现已继续带回退款摘要、关联退款记录、关联客服工单和客服工单摘要；客服工单列表也支持按 `related_order_id` 过滤，售后抽屉和异常履约可以顺着订单直接追到客服处置。
- 售后审计/工单详情深链：售后聚合详情现已继续带回关联审计摘要和最近审计记录；售后抽屉、异常履约和客服工作台补出订单审计、退款审计、客服工单详情入口，方便把售后、资金、工单三条线接成一条排查面。
- 退款流水/客服详情深链：后台现已补出 `GET /api/admin/refunds` 和 `GET /api/admin/service-tickets/{ticketID}`；订单监控、售后审核和异常履约可以直接顺着订单打开退款流水，客服工作台也能通过管理员接口直接看工单详情，不再借用户查询口。
- 预览卡同屏联动：售后聚合详情和退款流水预览卡现在已经能直接把操作台预填到“时间线 / 凭证 / 派单事件 / 退款流水 / 工单详情 / 审计检索”，运营看完摘要就能在同屏继续钻取，不用手动重填参数。
- 订单聚合详情：后台现已补出 `GET /api/admin/orders/{orderID}`，把订单主信息、售后、退款、客服工单、派单事件和关联审计一次性拉回；订单监控、售后详情抽屉和结果预览开始都可以把订单当成总入口来追问题。
- 联动结果区：预览卡上的只读动作现在会直接拉结果并保留到操作台联动面板，退款、工单、审计、派单这几条线可以并排对照；需要时还能把任一联动结果提到主结果位继续深钻。
- 同单工作区：当主结果里已经带出订单、售后或工单上下文时，操作台会直接给出“打开订单工作区 / 打开售后工作区”和一组常用挂板按钮，能一键把退款、工单、派单、审计、凭证这些排查面板挂起来。
- 同单工作区同步：联动面板现在会标出“已对齐 / 待同步”状态，支持按分组筛选、按订单/售后/工单/用户定位，并可一键“同步当前筛选 / 同步全部”，让已经挂出来的退款、工单、派单、审计卡片跟随当前主结果上下文一起刷新。
- 同单工作区切换：工作区现在还能直接切到某张联动卡的上下文，不必先把那张卡设为主结果。切换后会保留主结果、补出“跟随主结果”回退入口，并把定位焦点一起带到新的订单/售后/工单/用户上下文。
- 同单工作区预填：工作区工具条现在会按退款、客服、派单、审计分组给出同步按钮，操作台当前表单也会随工作区上下文自动预填 `order_id / request_id / ticket_id / user_id / target_type` 等参数，减少反复抄写和漏填。
- 同单工作区自动刷新：切到联动卡上下文、从候选上下文切换、回到“跟随主结果”以及把某张联动卡设为主结果后，当前筛选下待同步的联动卡会自动刷新；焦点匹配也会同时识别“当前卡上下文”和“同步后目标上下文”，避免刚切完上下文时该刷新的卡被漏掉。
- 同单工作区同步反馈：工作区现在会直接显示“同步中 / 同步反馈”提示，并按退款、客服、派单、审计给出分组状态；能一眼看到哪组正在刷新、哪组仍待同步、哪组已经对齐，不用逐卡判断。
- 同单工作区同步时间：同步反馈现在还会带上最近一次完成时间和涉及分组摘要，例如“14:32:08 · 退款 2 / 客服 1”；分组状态也会在同步完成后短暂显示“刚刷新 N”，方便运营确认刚刚到底动了哪几条线。
- 同单工作区局部失败：如果一轮同步里只有部分卡刷新失败，工作区会直接把失败张数写进反馈文案，并在对应分组上标成“失败 N”，不用再点进每张卡才知道是哪一条线没跟上。
- 同单工作区失败重试：工作区现在会给出“重试失败项”和“重试退款失败 / 重试客服失败”这类入口；如果某一轮同步只挂了一部分，运营可以直接就地补一轮失败项，不必手动筛卡再重跑。
- 操作结果预览：接口操作台对 `after-sales-events`、`after-sales-evidence`、`dispatch-order-events` 和 `refund-transactions` 已补出可视化结果面板，会把时间线、凭证卡片、派单事件卡片和退款流水卡片直接渲染出来，同时保留原始 JSON 便于排障。
- 高风险操作确认：邀约、售后审核、客服工单分派、客服 SLA 升级、客服处理方案、订单退款、退款策略、审计导出/告警/归档、通知回执补录、通知失败告警、通知失败重试、通知静默到期扫描、通知偏好保存、通知偏好批量保存、通知偏好变更申请/审批/灰度应用、RBAC 变更、Outbox 领取/续租/死信解封/单事件恢复/批量恢复/标记失败/标记已发布和订单状态补偿执行前进入二次确认，确认后在操作台保留最近执行结果；失败记录可一键恢复原操作和参数，高风险动作重试仍需再次确认。
- 模块导航：订单、售后、商户、骑手、调度、支付、钱包、优惠券、圈子、群聊红包、客服、RTC、OAuth/API 等模块的状态位。
- RBAC 首版矩阵：后台兼容管理员、超级管理员、运营、财务、调度、客服、安全审计员角色边界已与服务端 scope 对齐。
- BFF 浏览器接入：本地预览页可通过 BFF CORS 白名单和 `OPTIONS` 预检访问后台 API；部署域名通过 `BFF_ALLOWED_ORIGINS` 配置。

验证：

```bash
npm run test --workspace @infinitech/admin-web
npm run test --workspace @infinitech/bff
```

重点运营能力：

- 创建、撤销和审计商户/站长/骑手邀请链接。
- 审核商户营业执照、健康证、员工健康证和补充资料，资质过期时自动临时关店。
- 查看骑手接单时间、平均接单耗时、日均完成单数、配送评分、评价样本数、派单分、接单/单量/履约/评分拆解、团队均值、等级和派单优先级。
- 通过运营快照把 P0 视图从静态说明推进到统一数据口径，顶部 KPI、今日队列、订单、售后、商户、骑手、骑手绩效、派单和退款策略视图已可由该接口生成；订单、售后、商户、骑手、绩效、派单、退款策略、客服工作台、通知运营和权限治理已补业务详情面板、筛选分页、高风险二次确认、最近结果追踪和失败回放入口首版，客服工作台可打开 `GET /api/admin/service-tickets?sla_status=`、`GET /api/admin/service-ticket-quality-reviews`、`GET /api/admin/service-ticket-performance`，并从详情抽屉预填 `POST /api/admin/service-tickets/{ticketID}/assign`、`POST /api/admin/service-tickets/{ticketID}/escalate`、`POST /api/admin/service-tickets/{ticketID}/resolve` 与 `POST /api/admin/service-tickets/{ticketID}/quality-review`；通知运营页可打开 `GET /api/admin/notifications`、`GET /api/admin/notification-deliveries`、`GET/PUT /api/admin/notification-preferences`、`POST /api/admin/notification-preferences/batch`、`GET/POST /api/admin/notification-preferences/change-requests`、`POST /api/admin/notification-preferences/change-requests/{id}/review`、`POST /api/admin/notification-preferences/change-requests/{id}/apply`、`POST /api/notifications/{notificationID}/deliveries`、`POST /api/admin/notification-deliveries/failure-alerts/emit`、`POST /api/admin/notification-deliveries/retries/schedule` 和 `POST /api/admin/notification-deliveries/quiet-window-retries/schedule`，详情抽屉可按通知 ID、商户、来源 topic、回执状态、错误码、渠道和 provider 预填查询、补录、失败告警、重试计划、静默到期扫描、通知偏好保存和通知偏好变更审批/灰度应用动作，操作页也可批量保存多条通知偏好策略并在变更申请中填写 rollout JSON；商户模块和详情抽屉已可打开 `GET /api/admin/merchant-qualifications` 待审列表与 `GET /api/admin/merchant-qualifications/{qualificationID}` 明细，再预填资质审核表单进入 `POST /api/admin/merchant-qualifications/{qualificationID}/review` 二次确认，售后模块和详情抽屉已可预填审核表单并进入 `POST /api/after-sales/{requestID}/review` 二次确认，订单模块和详情抽屉已可预填退款表单并进入 `POST /api/orders/{orderID}/refund` 二次确认，运营首页和 Outbox 队列详情已可预填事件明细、死信分诊、领取租约、续租、死信解封、单事件恢复、标记失败和标记已发布表单，并分别进入 `GET /api/admin/outbox/events/{eventID}`、`GET /api/admin/outbox/events?status=dead_letter`、`POST /api/admin/outbox/events/claim`、`POST /api/admin/outbox/events/{eventID}/lease/renew`、`POST /api/admin/outbox/events/{eventID}/replay`、`POST /api/admin/outbox/events/{eventID}/failed`、`POST /api/admin/outbox/events/{eventID}/published`，高风险写动作进入二次确认，后续继续接更完整的审核辅助信息和后端业务明细接口。
- 通过审计检索页追踪商户/骑手邀约、退款策略、订单退款、状态补偿、售后审核、对象清理、outbox 运维和 RBAC 申请/审批/应用/回滚；后端已在 PostgreSQL-backed Store 下读取规范化 `audit_logs` 表，已支持 `security_auditor` 只读审计角色，并已新增 `ops_admin`、`finance_admin`、`dispatch_admin`、`support_admin` 等服务端 RBAC scope 边界；`GET /api/admin/audit-logs/export` 可按同一套审计筛选导出 CSV，导出行为写入 `admin.audit_logs.exported`；`GET /api/admin/audit-logs/retention-report` 可返回留存窗口、冷归档候选、完整性失败、导出事件和关键动作覆盖告警；`POST /api/admin/audit-logs/archive/request` 可生成审计冷归档 manifest、`sha256:v1` manifest hash 和 `audit.archive_requested` outbox 事件，并写入 `admin.audit_archive.requested` 审计；`GET /api/admin/rbac/policy` 可读取服务端真实矩阵，`GET /api/admin/rbac/change-requests` 可读取权限申请台账，`POST /api/admin/rbac/change-requests` 可提交待审批申请，`POST /api/admin/rbac/change-requests/{id}/review` 可审批或驳回并禁止同一管理员自审，`POST /api/admin/rbac/change-requests/{id}/apply` 可把已审批申请手动应用到运行时权限矩阵并写入 `admin.rbac.change_applied`，`POST /api/admin/rbac/change-requests/{id}/rollback` 可把当前已应用申请按应用前 scopes 回滚并写入 `admin.rbac.change_rolled_back`，服务启动会按应用/回滚审计日志恢复策略；服务端写入/读取审计 payload 前会统一白名单过滤和敏感字段掩码；审计记录已返回 `integrity_algorithm`、`integrity_hash`、`integrity_verified`，生产配置 `AUDIT_LOG_SIGNING_SECRET` 后可用 HMAC 检测篡改；退款策略、订单退款、售后审核、订单状态补偿、对象清理完成/失败、outbox 领取/续租/发布/失败/重放/批量重放和商户/骑手邀约均已通过仓储级 `WithAudit` 原子路径在同一业务事务内写入最终业务结果和审计，相关 payload 均由服务端根据最终业务结果规范化生成；页面支持 after/before 时间范围、before 游标翻页、保存筛选、详情抽屉、目标筛选、跨模块跳转、白名单摘要、敏感字段脱敏详情、完整性状态、CSV 导出、留存告警报告和归档请求操作，后续继续补字段级/租户级权限、剩余后台配置/运营处置/资金风控写路径同事务审计、真实 WORM 对象存储写入/冷热归档 worker、真实告警投递和 KMS/链式不可抵赖签名。
- 审计留存告警已新增 `POST /api/admin/audit-logs/retention-alerts/emit` 操作，可把健康报告中的告警投递到 `audit.retention_alerts` outbox topic，并由 notification-worker 进入后续通知链路。
- 审计 WORM/冷归档请求已新增 `POST /api/admin/audit-logs/archive/request` 操作，可把冷归档候选 manifest 投递到 `audit.archive_requested` outbox topic，供后续真实 WORM 归档 worker 执行。
- 配置站点每日任务时长、每日固定订单数和完成固定数后的免责不接规则。
- 配置退款默认策略：默认退平台余额或默认原路返回，并由后端同事务写入审计账本。
- 配置首页推荐卡片，控制商品、店铺、团购、优惠券和圈子动态在首页展示。
- 管理圈子/小微墙、找饭搭开关、问卷题库、协议版本、审核、举报和敏感词。
- 管理优惠券资金责任：商户自担、平台补贴、商户弹窗同意参与后自担。
- 保留旧版后台闭环：售后、评价、精选商品、首页活动、内容设置、通知推送、积分会员、数据备份恢复、交易日志、电话联系审计、系统日志、开放平台权限和文档。
- 管理官方群、商户群、进群领券规则、红包风控和红包流水。
- 同单工作区失败定位继续补强：联动结果区现在支持“失败”筛选，失败卡会直接显示失败原因，并提供“重试此卡”入口；工作区工具条也补了“仅看失败项”，方便运营先把红卡收拢再处理。
- 失败卡继续往“可解释”推进：红卡现在会直接带出 HTTP 状态、最后一次请求路径、原参数摘要，以及按当前工作区上下文重试后会改写成什么参数，不用再自己对着 JSON 猜。
- 同单工作区失败卡还补了两块收口：一是可以直接把当前重试参数回填到操作台，二是会保留最近一次失败和最近一次成功的对照痕迹，方便判断这张卡到底是一直挂着，还是刚从好转坏。
- 同单工作区失败卡继续往“可执行”推进：点了“回填到操作台”后，操作台会显示一块专门的回填确认面板，直接给出来源、当前参数、最近失败/最近成功痕迹和“执行当前查询”入口；红卡里的痕迹也从一行文案收成了更清楚的局部历史列表。
