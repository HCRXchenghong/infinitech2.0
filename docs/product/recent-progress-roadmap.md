# Infinitech 2.0 最近进展与路线图

更新时间：2026-05-29
目标仓库：`https://github.com/HCRXchenghong/infinitech2.0`  
当前最新提交：以 GitHub `main` 分支最新提交为准

## 当前结论

项目已经完成架构基线、monorepo 工程骨架、首批端侧页面、核心 API 大量业务闭环、BFF 代理、Worker 骨架、PostgreSQL 规范化账本、用户端消息 PostgreSQL 规范化首版、realtime gateway Redis adapter 多副本 fanout 首版、realtime gateway WebSocket 签名鉴权首版、realtime gateway 会话成员权限校验首版、会话免打扰偏好首版、群资料/成员预览首版、outbox relay、对象扫描/清理、审计归档 worker、审计归档完成回写和归档记录查询首版、审计归档对象下载校验/回查首版、审计归档校验历史查询首版、审计归档校验历史可视化面板首版、管理端 P0 运营视图、管理端 P0 业务详情面板首版、管理端高风险操作二次确认与结果追踪首版、管理端失败回放入口首版、管理端 P0 业务筛选分页首版、管理端售后审核表单首版、管理端订单退款表单首版、管理端 Outbox 单事件恢复首版、管理端 Outbox 发布/失败人工处置首版、管理端 Outbox 领取/续租首版、管理端 Outbox 死信分诊/解封首版、管理端 Outbox 单事件事故辅助明细首版、商户资质审核后端与管理端表单首版、商户资质待审列表与明细接口首版、商户资质审核结果可靠通知首版、商户站内通知中心首版、通知运营查询接口首版、通知投递回执台账首版、Admin Web 通知运营页首版、通知失败回执告警首版、通知失败重试编排首版、通知 provider 执行器首版、通知 provider 回调验签入账首版、通知 provider 模板映射与渠道 payload 规范首版、通知偏好与静默窗口首版、通知偏好后端账本与 API 首版、通知 worker 后端偏好读取首版、通知静默 queued 再投递调度首版、通知静默到期自动扫描调度首版、商户端通知偏好设置首版、管理端通知偏好操作入口首版、用户端通知偏好设置首版、用户端真实微信 `wx.login` 页面流程首版、用户端找饭搭资料人工审核与举报处置首版、用户端找饭搭同校隐私与设备风控首版、客服消息敏感信息风控首版、通知 worker 偏好缓存与失败关闭首版、通知偏好变更事件与 worker 主动失效首版、通知偏好批量保存与策略审计首版、通知偏好变更审批与应用首版、管理端审计账本、审计中心增强首版、审计导出首版、审计留存/告警健康报告首版、审计留存告警 outbox 投递首版、审计 WORM/冷归档请求首版、审计服务端安全边界首版、审计完整性证明首版、管理端服务端 RBAC 策略矩阵首版、RBAC 权限治理查询与变更申请审计首版、RBAC 权限申请审批/驳回台账首版、RBAC 权限变更手动应用首版、RBAC 权限变更审计回滚首版，以及退款策略配置、管理端订单退款、售后审核、订单状态补偿、对象清理完成/失败、outbox 运维、商户资质审核和商户/骑手邀约审计同事务首版。

但项目仍未达到商业级可上线状态。真实生产支付、微信原路退款、提现结算、真实 IM/RTC、真实短信/企业微信/端内 push 通知渠道生产账号、模板审批、真实 provider 字段映射联调、渠道模板/供应商配置审批、升级策略、完整后台页面、字段级/租户级权限治理、生产 WORM bucket 策略、归档校验生产演练回查、保留期删除审批、真实高可用基础设施、10 万在线压测和容灾演练还没完成。未完成这些验收前，只能说“按商业级目标推进中”，不能说“已经商业级上线”。

## 最近已完成

### 0. 客服消息敏感信息风控首版

- IM 消息发送、客服工单创建和客服工单补充事件已共用敏感信息风控。
- 消息会返回 `risk_state`、`risk_reason_code`、`risk_reason` 和 `risk_checked_at`；敏感提及允许发送并标记 `flagged`，疑似泄露验证码、支付密码或银行卡号时返回 `RISK_CONTROL_REJECTED`。
- 用户端在线客服页已补发送前提示和拦截卡片，Admin Web 客服工作台已把消息风控纳入防护清单。

### 0.1 用户端找饭搭同校隐私与设备风控首版

- 找饭搭资料新增学校、校区、楼栋、隐私范围、位置精度和设备 ID，候选列表返回同校/同楼、隐私提示和设备风控结果。
- 候选只在同校范围展示；任一方选择同楼保护时必须同楼才展示，页面不再暴露精确米数。
- 缺失设备或共享设备会进入人工复核，已知风险设备会被服务端风控拒绝。

### 0.2 用户端真实微信 wx.login 页面流程首版

- 登录页和注册页共用真实 `wx.login` code 流程，再通过 BFF/API 换取服务端 token。
- 生产接口失败时页面停留并提示，不再自动写入开发 token。
- 本地 API 或显式 `allowPreviewAuth=true` 时仍可进入预览兜底，方便微信开发者工具调试。

### 0.3 用户端消息 PostgreSQL 规范化首版

- `messages` 已补发送人名称、风险状态、风险原因和风控检查时间字段，消息会按会话时间索引恢复。
- `conversation_read_states` 已存储用户在会话内的已读游标，承接离线补偿和打开会话清未读。
- `PostgresStore` 启动会把快照消息同步到规范化表，再从 `messages` 和 `conversation_read_states` 恢复；`message.sent` outbox 仍可继续交给 realtime-gateway 投递。

### 0.4 Realtime Gateway Redis 多副本 fanout 首版

- realtime-gateway 新增 Redis Pub/Sub cluster adapter，内部发布先投递本机连接，再广播到共享 channel。
- Redis envelope 带 `source_id`，其他副本收到后按 `thread_id` 投递本机 WebSocket 连接，发布副本会跳过自己的回环消息。
- Docker Compose 已接本地 Redis URL，K8s base 已接 `realtime-redis-url` secret 和统一 channel。

### 0.5 Realtime Gateway WebSocket 签名鉴权首版

- realtime-gateway 支持 `REALTIME_WS_AUTH_REQUIRED=true` 后强制校验 API-Go 同款 HMAC access token。
- WebSocket upgrade 会校验 `sub`、`role`、`exp`，并拦截 token 用户与 URL `user_id` 不一致的冒充连接。
- 用户端小程序登录后保存 `userId`，商户群 WebSocket 连接使用真实当前用户 ID 并继续传 `Authorization`。

### 0.6 Realtime Gateway 会话成员权限校验首版

- API-Go 新增 `POST /internal/realtime/authorize` 内部授权口，通过 `REALTIME_INTERNAL_TOKEN` 保护，按 `conversation_members` 校验会话主体是否可访问指定 `thread_id`。
- 用户端消息列表、聊天记录、离线同步、已读回执和发送消息会隐藏非成员访问，避免只靠前端传参判断会话权限。
- realtime-gateway 在 WebSocket upgrade 通过签名 token 后会调用成员授权口；非成员返回 403，授权服务不可用返回 503，Docker Compose/K8s 已接 `REALTIME_MEMBERSHIP_AUTH_URL`。

### 0.7 会话免打扰偏好首版

- API-Go 新增 `GET/PUT /api/messages/{threadID}/preference`，用户可按会话读取和切换 `muted`；非成员仍返回隐藏式 404。
- `Store` 与 `PostgresStore` 会同步 `conversation_members.muted`，会话列表、内部实时授权返回和小程序商户群页都使用同一份静默状态。
- 原生微信小程序消息页会展示“免打扰”标签，商户群页右上角可直接切换“已静默/消息提醒”。

### 0.8 群资料与成员预览首版

- API-Go 新增 `GET /api/messages/{threadID}/overview` 和 `GET /api/messages/{threadID}/members`，小程序可读取群摘要、群公告、活跃成员预览和会话当前静默状态。
- `Store` 会基于 `conversation_members` 输出群资料和成员视图，并把商户群参考图里的 `小林`、`阿杰` 等消息参与者补成可读成员数据。
- 原生微信小程序商户群页已改成“群资料 + 群设置”入口，新增 `pages/messages/group-settings/index` 子页承接免打扰开关与成员预览。

### 0.9 商户群自助入退群与群券资格首版

- API-Go 新增 `GET /api/messages/{threadID}/membership`、`POST /api/messages/{threadID}/join`、`POST /api/messages/{threadID}/leave`，商户群成员数会跟随当前成员快照动态变化。
- `Store` / `PostgresStore` 现在把 `conversation_members` 作为会话成员真来源，支持用户自助加入、退出以及按群成员资格领取 `GROUP8` 商户群券。
- 原生微信小程序店铺详情、商户群页、群设置页和红包优惠页已经挂上同一条入群/退群/领券链路，用户可从商家页入群、在群里领券、在群设置退出。

### 0.10 店铺详情评价与商家信息聚合首版

- API-Go 新增 `GET /api/shops/{shopID}/detail`，把头图、公告、活动标签、评分摘要、评价列表和商家资料聚到一个公开接口里。
- `Store` 会把默认评价种子、动态新增评价和商户资质/员工联系电话一起汇总，供店铺详情页直接消费。
- 原生微信小程序店铺详情页已补“评价 / 商家”两栏，支持展示真实评价卡、商家资质、营业时间、联系电话和复制地址操作。

### 0.11 下单链路购物车与地址回填首版

- `CartSummary` 新增 `shop_name`，`Order` 新增 `shop_name` 和 `address_snapshot`，下单后订单详情可以直接拿到商家名和配送地址快照。
- 原生微信小程序购物车页已接真实购物车摘要、商品增减和清空；确认订单页会读取地址列表和购物车费用汇总，支持从地址列表选中配送地址返回。
- 地址列表页支持按 `POST /api/user/addresses` 回写默认地址，订单详情页已补真实进度时间和状态文案，不再依赖静态商家名/地址占位。

### 0.12 订单评价真实回填与更新首版

- `Review` 新增 `anonymous`，`GET /api/reviews` 支持按 `order_id` 过滤，同一用户对同一订单再次提交会原地更新已有评价。
- `Order` 新增 `reviewed`，订单列表和订单详情可直接识别“待评价 / 已评价”状态，评价入口不再停留在静态按钮层。
- 原生微信小程序评价页已接真实订单摘要、已评价内容回填、匿名开关、星级交互和订单级提交；商家评价聚合会对匿名评价做用户名脱敏。

### 0.13 售后页订单上下文与进度预览首版

- `AfterSalesRequest` 新增 `shop_name`、`order_status`、`order_item_summary`、`latest_event_message` 和 `latest_event_at`，`GET /api/after-sales` 支持按 `order_id` 过滤。
- 原生微信小程序售后页已改成真实选单模式，能回填订单摘要、剩余可退金额、订单级售后记录、最近进度和凭证预览，不再错误跳去客服工单详情页。
- 售后页在当前订单已有处理中申请时会优先展示进度时间线，避免用户重复提交后端本就会拦截的重复申请。

### 0.14 售后页补充凭证上传首版

- 原生微信小程序售后页已接 `POST /api/after-sales/{requestID}/evidence/upload-ticket` 与 `POST /api/after-sales/{requestID}/evidence/confirm`，用户可在处理中售后单内直接补传图片凭证。
- 前端已兼容对象存储严格门禁：默认环境可直接确认；若环境要求对象存储上传回调或扫描通过，会先补 `POST /api/object-storage/upload-callback`，再按需要补 `POST /api/object-storage/scan-result` 后重试确认。
- 售后凭证区已从静态占位改成真实凭证卡、上传中状态、补充入口和历史凭证预览，页面细节继续往参考图收口。

### 0.15 评价页图片上传与逐项菜品评分首版

- `Review` 新增 `item_ratings`，原生微信小程序评价页现在会按订单商品回填逐项评分，并把菜品评分随订单评价一起提交。
- 评价页已接 `POST /api/reviews/upload-ticket` 与 `POST /api/reviews/upload-confirm`，用户可在提交评价前先上传图片；默认环境可直接确认，严格环境会补对象存储回调与扫描通过结果后重试确认。
- 页面层把原来的静态图片占位和只读菜品标签改成真实上传卡、上传中状态、图片预览和逐项星级交互，整体更接近参考图。

### 0.16 店铺详情评价图片区与配送服务评分首版

- `Review` 新增 `rider_rating`，评价页的“配送服务”不再和商家总体评分混成一项，后端会独立保存并在订单评价回填时带回。
- 店铺详情聚合评价卡现在会返回 `image_urls`、`item_highlights` 和 `rider_stars_text`，原生微信小程序评价区已经补上图片预览、菜品评分摘要和配送服务标识。
- 这轮把“评价页填写什么”和“店铺详情展示什么”打成了一条线，评价链开始有成品态，而不只是单页表单闭环。

### 0.17 骑手配送评分聚合与站长绩效可视化首版

- `StationRiderPerformance` 现在会聚合订单评价里的 `rider_rating`，返回骑手配送平均分和评价样本数，开始把配送评价真正接进骑手绩效口径。
- 站长工作台和管理端骑手绩效视图已把原来的占位列换成配送评分展示，站长能直接看到“接单耗时 + 完成率 + 配送评分”的组合信息。
- 用户端店铺评价图区也顺手补成了稳定三宫格预览，多图时最后一格会提示余量，页面更接近参考图常见节奏。

### 0.18 骑手配送评分进入派单分与优先级首版

- 骑手绩效分现在会把配送评分样本折算进 `score`，并在站点内刷新 `dispatch_priority`，派单优先级不再只看历史接单效率。
- 自动派单在挑选候选骑手前会先刷新站点骑手优先级，确保新产生的配送评分能在下一次派单里生效。
- 站长工作台和管理端绩效视图也同步把“配送评分 + 样本数 + 派单分”展示出来，方便解释为什么某个骑手更靠前。

### 0.19 骑手绩效拆解与站长可解释展示首版

- `RiderPerformance` 现在会返回 `score_breakdown`，拆出接单加分、单量加分、履约加分、评分加分、评分置信度和站点团队均值。
- 站长工作台新增绩效拆解卡，选中骑手后能直接看到派单分来源，不再只有一个总分数字。
- 管理端骑手绩效表和详情抽屉同步补了“派单分 + 评分拆解”列，后台可以更直观地解释排序原因。

### 0.20 骑手绩效趋势与异常履约钻取首版

- `RiderPerformance` 继续补出 `recent_trend`、`recent_reviews` 和 `exception_summary`，把最近 3 天走势、最新评价摘录、超时/拒单/售后/低分异常一起带回站长端和后台。
- 站长工作台绩效拆解卡新增趋势、评价和异常履约区，站长不只知道“分高不高”，还知道“这几天为什么变高或变低”。
- 管理端骑手绩效详情抽屉同步补了趋势摘要、最近评价和异常履约摘要，方便派单复盘和运营追责。

### 0.21 骑手异常履约明细深链首版

- `RiderPerformance` 继续补出 `exception_details`，把超时、拒单、售后介入、低分评价整理成可排序的异常明细，直接带上 `order_id`、`dispatch_event_id`、`after_sales_request_id` 或 `review_id`。
- 管理端骑手绩效详情抽屉会把异常明细逐条展开，并给出“查看派单事件 / 查看售后 / 查看审计”的预填动作，异常不再只是一行汇总数字。
- 后台售后列表也补了 `request_id / order_id / status` 过滤，异常明细动作终于能真正落到后端查询，而不是停在假入口。

### 0.22 售后时间线与凭证钻取首版

- 管理端控制台继续补出 `after-sales-events` 和 `after-sales-evidence` 两个只读操作，能直接按售后单回看完整时间线和凭证列表。
- 售后审核详情抽屉现在会直接给出“查看售后时间线 / 查看售后凭证 / 查看订单派单事件”的动作，排查路径更短。
- 骑手异常履约明细里的售后类异常也会优先跳到这两个读操作，避免运营只看到筛选列表却还得自己再翻一层。

### 0.23 售后与派单结果可视化首版

- 管理端接口操作台对 `after-sales-events`、`after-sales-evidence` 和 `dispatch-order-events` 新增可视化预览，不再只吐原始 JSON。
- 售后时间线会按“用户提交 / 商户回复 / 客服跟进 / 审核结果”卡片化展示，顺手带上用户可见/内部备注和附件入口。
- 售后凭证和派单事件也会分别渲染成凭证卡片与事件卡片，方便运营快速看图、看状态、看派单原因，再决定要不要继续钻原始返回。

### 0.24 售后聚合详情首版

- 后端新增 `GET /api/admin/after-sales/{requestID}`，把售后工单、时间线、凭证和派单关联一次性聚合返回。
- 管理端售后详情抽屉与骑手异常履约里的售后异常都补出“查看售后详情”动作，不再要求运营自己点四个接口拼上下文。
- 聚合详情也接进了管理端结果预览，会先展示工单概览、时间线概览、凭证概览和派单概览，再保留原始 JSON 兜底排障。

### 0.25 售后资金与客服深链首版

- 售后聚合详情继续补出退款摘要、关联退款记录、关联客服工单和客服工单摘要，后台可以顺着同一个售后单追到资金和客服处理面。
- 管理端客服工单列表新增 `related_order_id` 过滤，售后抽屉和骑手异常履约里的售后异常都能直接按订单打开对应客服工单。
- 管理端结果预览里的售后聚合详情也同步补出“退款概览 / 客服工单”两张卡，减少运营在退款、售后、客服三块之间来回切换。

### 0.26 售后审计与工单详情深链首版

- 售后聚合详情继续补出关联审计摘要和最近审计记录，先把订单审计、售后审核审计和客服工单审计的事实面聚回到同一响应里。
- 管理端售后抽屉和骑手异常履约明细新增“查看订单审计 / 查看退款审计”动作，排查退款或仲裁原因时不需要再手动拼过滤条件。
- 管理端新增 `support-ticket-detail` 操作和结果预览，客服工作台现在可以直接打开单个工单详情，看责任人、SLA、时间线和最近处理进展。

### 0.27 退款流水与管理员工单详情首版

- 后端新增 `GET /api/admin/refunds`，支持按订单、用户、退款去向、状态筛选退款流水，资金侧排查终于不用只靠审计日志倒推。
- 后端新增 `GET /api/admin/service-tickets/{ticketID}` 管理员详情口，客服工作台不再借用户工单查询路径取详情。
- 管理端订单监控、售后审核和骑手异常履约明细新增“查看退款流水”动作；退款流水也接入了结果预览卡片，方便快速核对金额、去向和幂等键。

### 0.28 售后预览卡同屏联动首版

- 管理端 `after-sales-detail` 预览卡现在不再只是摘要展示，已补出“工单列表 / 时间线 / 凭证 / 派单事件 / 退款流水 / 工单详情 / 审计检索”的预填按钮。
- 退款流水预览卡也已补出“退款审计”联动，运营核对金额、幂等键后可以直接顺着同一订单跳到审计检索，不需要再手工抄订单号。
- 这轮没有新增后端接口，重点是把已经接好的售后、客服、退款、审计能力在 Admin Web 操作台里收成同屏可切换的工作面。

### 0.29 订单聚合详情首版

- 后端新增 `GET /api/admin/orders/{orderID}`，把订单主信息、售后记录、退款流水、客服工单、派单事件和关联审计一次性聚合回来。
- 管理端新增 `order-detail` 操作；订单监控详情抽屉、售后详情抽屉和退款/客服预览卡都开始能顺着订单切到这个总览，不再只能从售后单往下钻。
- Admin Web 结果预览也已补出“订单聚合详情”卡组，订单现在可以作为后台排查的总入口，继续往售后、退款、工单和审计四条线分发。

### 0.30 订单总览联动结果区首版

- Admin Web 接口操作台新增“联动结果区”，预览卡上的只读动作会直接拉结果，不再只是预填表单；退款、工单、派单、审计等结果会保留成最近 3 个并排面板。
- 联动结果支持去重、关闭和“设为主结果”，后台排查可以先并排对照，再把其中一条结果提到主结果位继续往下钻。
- 这轮没有新增后端接口，重点是把已经接好的订单/售后/退款/工单/审计能力真正收成一个更像工作面的后台体验。

### 0.31 同单工作区首版

- Admin Web 操作台新增“同单工作区”工具条：当主结果已经带出订单、售后或工单上下文时，会直接给出上下文标签和一组挂板动作。
- 订单上下文支持一键展开退款流水、客服工单、派单事件和订单审计；售后上下文支持一键展开时间线、凭证、售后审计和退款流水。
- 这一段依旧没有新增后端接口，重点是把订单聚合详情和售后聚合详情进一步收成更像后台排查工位的使用体验。

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

### 3.1 管理端 P0 业务详情面板首版

- Admin Web 已新增 `adminDetails` 详情适配层。
- 订单、售后、商户、骑手/站长、骑手绩效、派单、审计、退款策略和权限治理表格行可打开详情面板。
- 详情面板展示当前行字段、模块化核查清单和下一步操作按钮，可跳到订单状态补偿、审计检索、Outbox 事件、对象清理候选、退款策略、RBAC 策略/申请等已有操作入口。
- Admin Web 单测和架构守卫已覆盖详情适配器、详情按钮和下一步动作，避免 P0 视图回退成纯静态表格。

### 3.2 管理端高风险操作二次确认与结果追踪首版

- Admin Web 已新增 `adminOperations` 高风险操作适配层。
- 邀约、退款策略、审计导出/告警/归档、RBAC 变更、Outbox 领取/续租/死信解封/单事件恢复/批量恢复/标记失败/标记已发布和订单状态补偿执行前进入 `pending_confirmation` 确认面板。
- 确认面板展示方法、路径、区域、风险原因和即将提交的参数快照，避免运营人员误触高危写操作。
- 操作执行后会保留最近结果，展示成功/失败、HTTP 状态、请求路径和返回消息，并已由失败回放入口承接首版重试。
- Admin Web 单测和架构守卫已覆盖风险识别、待确认 payload、结果追踪和确认按钮。

### 3.3 管理端失败回放入口首版

- 操作结果追踪中的失败记录会展示“重试”入口。
- 点击重试会恢复原操作和参数快照，减少运营人员手工复制请求参数的误差。
- 高风险动作重试时仍会重新进入 `pending_confirmation` 面板，不绕过二次确认。
- Admin Web 单测、架构守卫和本地浏览器渲染检查已覆盖失败记录重试链路。

### 3.4 管理端 P0 业务筛选分页首版

- Admin Web 已新增 `adminTable` 表格适配层。
- P0 业务视图支持关键字筛选、每页条数和上一页/下一页控制。
- 筛选后仍保留原始行索引，详情面板会打开正确业务行。
- Admin Web 单测、架构守卫和本地浏览器渲染检查已覆盖订单监控筛选、清除和分页计数。

### 3.5 管理端售后审核表单首版

- Admin Web 操作目录新增 `after-sales-review`，连接现有 `POST /api/after-sales/{requestID}/review`。
- 售后审核模块和售后详情抽屉可预填工单 ID、审核结果、审核原因、退款去向和退款幂等键。
- 售后审核已纳入高风险操作，执行前会进入 `pending_confirmation` 二次确认，确认面板展示方法、路径、区域和参数快照。
- Admin Web 单测、架构守卫和本地浏览器渲染检查已覆盖表单预填、请求构造、二次确认和 console error 为 0。

### 3.6 管理端订单退款表单首版

- Admin Web 操作目录新增 `order-refund`，连接现有 `POST /api/orders/{orderID}/refund`。
- 订单监控模块和订单详情抽屉可预填订单 ID、退款原因、退款幂等键、可选退款金额和退款去向。
- 订单退款已纳入高风险操作，执行前会进入 `pending_confirmation` 二次确认，避免资金动作误触。
- Admin Web 单测、架构守卫和本地浏览器渲染检查已覆盖表单预填、请求构造、二次确认和 console error 为 0。

### 3.7 管理端 Outbox 单事件恢复表单首版

- Admin Web 操作目录新增 `outbox-replay-event`，连接现有 `POST /api/admin/outbox/events/{eventID}/replay`。
- 运营首页和 Outbox 队列详情抽屉可预填事件 ID，减少故障处置时手工切换接口和复制参数的错误。
- Outbox 单事件恢复已纳入高风险操作，执行前会进入 `pending_confirmation` 二次确认，确认面板展示方法、路径、区域、风险原因和参数快照。
- Admin Web 单测、架构守卫和本地浏览器渲染检查已覆盖表单预填、请求构造、二次确认和浏览器日志为空。

### 3.8 管理端 Outbox 发布/失败人工处置表单首版

- Admin Web 操作目录新增 `outbox-mark-failed` 和 `outbox-mark-published`，连接现有 `POST /api/admin/outbox/events/{eventID}/failed` 与 `POST /api/admin/outbox/events/{eventID}/published`。
- 运营首页和 Outbox 队列详情抽屉可预填事件 ID、失败原因、重试延迟和最大尝试次数，让人工 ACK/FAIL 处置不再依赖手工拼接口。
- 两个动作均纳入高风险操作，执行前会进入 `pending_confirmation` 二次确认，确认面板展示方法、路径、区域、风险原因和参数快照。
- Admin Web 单测、架构守卫和本地浏览器渲染检查已覆盖表单预填、请求构造、二次确认和浏览器日志为空。

### 3.9 管理端 Outbox 领取/续租表单首版

- Admin Web 操作目录新增 `outbox-claim-events` 和 `outbox-renew-lease`，连接现有 `POST /api/admin/outbox/events/claim` 与 `POST /api/admin/outbox/events/{eventID}/lease/renew`。
- 运营首页和 Outbox 队列详情抽屉可预填 topic、limit、lease owner、lease seconds 和事件 ID，让人工接管 relay 租约处置不再依赖手工拼接口。
- 两个动作均纳入高风险操作，执行前会进入 `pending_confirmation` 二次确认，确认面板展示方法、路径、区域、风险原因和参数快照。
- Admin Web 单测、架构守卫和本地浏览器渲染检查已覆盖表单预填、请求构造、二次确认和浏览器日志为空。

### 3.10 管理端 Outbox 死信分诊/解封表单首版

- Admin Web 操作目录新增 `outbox-dead-letter-triage` 和 `outbox-release-dead-letter`，复用现有 `GET /api/admin/outbox/events?status=dead_letter` 与 `POST /api/admin/outbox/events/{eventID}/replay`。
- 运营首页、今日队列和 Outbox 队列详情抽屉可直接进入死信分诊列表，默认预填 `topic=order.paid`、`status=dead_letter` 和 `limit=20`。
- 死信解封使用专门标题和风险原因，预填 `obe_dead_1`，执行前进入 `pending_confirmation` 二次确认，避免把毒消息误放回可靠投递队列。
- Admin Web 单测、架构守卫和本地浏览器渲染检查已覆盖表单预填、请求构造、二次确认、详情入口和浏览器日志为空。

### 3.11 商户资质审核后端与管理端表单首版

- 商户上传资质从“提交即通过”改为 `pending_review`，只有后台审核通过的营业执照/健康证才会计入商户接单资格。
- 新增 `POST /api/admin/merchant-qualifications/{qualificationID}/review`，由 `merchant:qualification_review` scope 守护，审核结果写入 `admin.merchant_qualification.reviewed` 审计；PostgreSQL-backed Store 在同一事务内更新规范化资质表、审计表和快照。
- BFF 已允许该后台路径，Admin Web 操作目录和商户详情抽屉已新增“审核商户资质”表单，支持 approve/reject、原因和审核时间，并进入高风险二次确认。
- Store、HTTP、BFF、Admin Web 单测与架构守卫已覆盖待审、审批、权限和请求构造。

### 3.12 Outbox 单事件事故辅助明细首版

- 新增 `GET /api/admin/outbox/events/{eventID}`，由 `outbox:read` scope 守护，内存 Store 和 PostgreSQL Store 均返回一致的事件事故明细。
- 返回内容包含事件状态、ready/blocked/lease 信号、retry/lease 剩余秒数、payload 摘要、关联业务目标、最近 outbox 审计、推荐下一步操作和人工处置核查清单。
- BFF 已允许该只读路径，Admin Web 操作目录新增 `outbox-event-detail`，运营首页和 Outbox 队列详情抽屉可预填 `event_id` 查看事故明细，再跳转恢复、续租、领取或审计。

### 3.13 商户资质待审列表与明细接口首版

- 新增 `GET /api/admin/merchant-qualifications`，复用 `merchant:qualification_review` scope，可按状态、商户、资质类型和 limit 查询资质队列；默认返回 `pending_review`。
- 新增 `GET /api/admin/merchant-qualifications/{qualificationID}`，返回资质状态、商户账号、店铺、保证金、缺失资质、接单资格、最近 `admin.merchant_qualification.reviewed` 审计、推荐下一步操作和审核核查清单。
- BFF 已允许列表/明细只读路径，Admin Web 操作目录新增 `merchant-qualifications` 与 `merchant-qualification-detail`，运营首页、商户模块和商户详情抽屉可先查看待审上下文，再进入高风险审核确认。
- Store、HTTP、BFF、Admin Web 单测与架构守卫已覆盖请求构造、权限边界、事故元数据和代理转发。

### 3.14 商户资质审核结果可靠通知首版

- `ReviewMerchantQualificationWithAudit` 审核通过/驳回后会生成 `merchant.qualification_reviewed` outbox 事件，事件 payload 包含商户、资质、审核结果、原因、有效期、目标角色和通知文案。
- HTTP 审核响应返回 `outbox_event`，运营可从 Outbox 队列继续查看或恢复该审核结果事件。
- `notification-worker` 已订阅 `merchant.qualification_reviewed` 并生成面向商户的审核结果通知 payload；outbox relay 默认 topic、Docker Compose 和 K8s 部署骨架已加入该事件。
- 当前边界：这是可靠事件和通知 worker payload 首版，站内信落库和投递回执已由后续通知中心增量承接；仍不是实际短信、企业微信、微信小程序订阅消息或商户端 push。

### 3.15 商户站内通知中心首版

- 新增平台通知账本 `PlatformNotification`、幂等写入请求、商户通知列表请求和已读请求，默认使用 `in_app` 渠道与 `unread/read` 状态。
- 内存 Store 与 PostgreSQL-backed Store 均支持 `CreateNotification`、`Notifications` 和 `MarkNotificationRead`；PostgreSQL 路径会确保 `platform_notifications` 表、`idempotency_key` 唯一约束、目标/状态/时间索引和 source 索引，并从规范表恢复快照。
- 新增 `POST /api/notifications` 后台/worker 写入入口，由 `notification:write` scope 守护；新增 `GET /api/merchant/notifications` 与 `POST /api/merchant/notifications/{notificationID}/read`，商户只能查看和标记自己的通知，管理员可带 `merchant_id` 辅助排查。
- `notification-worker` 已能把可靠事件规范化为站内通知创建请求，并通过 worker token 写入 API；BFF 已允许商户通知列表和已读路径。
- 当前边界：这是站内通知账本和商户读写 API 首版，投递回执、Admin Web 通知运营页和失败回执告警已由后续增量承接；仍不是短信、企业微信、微信小程序订阅消息、商户端 push 或跨端消息中心。

### 3.16 通知运营查询接口首版

- 新增 `notification:read` scope 和 `CanReadNotifications` 权限判断，`ops_admin` 可读写通知账本，`support_admin` 只能只读排查，`security_auditor` 不读取运营通知。
- 新增 `GET /api/admin/notifications`，支持按 `target_role`、`target_id`、`status`、`source_topic`、`source_event_id` 和 `limit` 查询通知账本。
- 内存 Store 与 PostgreSQL 查询路径均支持来源 topic/event 过滤，BFF 已允许该后台只读路径。
- 当前边界：这是后台 API 查询首版，投递回执、Admin Web 通知运营页和失败回执告警已由后续增量承接；外部触达渠道仍待补。

### 3.17 通知投递回执台账首版

- 新增 `PlatformNotificationDelivery`、回执写入请求和回执列表请求，支持 `queued`、`delivered`、`failed` 状态，保留 provider message id、错误码、错误信息、attempted/delivered 时间和幂等键。
- 内存 Store 与 PostgreSQL-backed Store 均支持 `RecordNotificationDelivery` 和 `NotificationDeliveries`；PostgreSQL 路径会确保 `platform_notification_deliveries` 表、幂等唯一约束、通知/目标/渠道状态索引，并从规范表恢复快照。
- 新增 `POST /api/notifications/{notificationID}/deliveries` worker/后台回执写入入口和 `GET /api/admin/notification-deliveries` 后台查询入口，BFF 已允许对应路径。
- `notification-worker` 写入 in-app 通知成功后会继续记录 `delivered` 回执；后台可查询 `failed` 回执，失败回执告警已由后续增量承接。
- 当前边界：这是回执账本和 API 首版，Admin Web 通知运营页、失败回执告警、失败重试编排、provider 执行器骨架、provider 回调验签入账和模板 payload 规范已由后续增量承接；真实短信/企业微信/订阅消息/push 生产账号、模板审批、provider sandbox 字段联调、通知偏好和跨端消息中心仍待补。

### 3.18 Admin Web 通知运营页首版

- Admin Web `notifications` 模块已从 planned 推进到 wired，新增“通知运营”业务视图，展示未读站内信、失败回执、渠道和待接 provider 的运营指标位。
- 操作目录接入 `GET /api/admin/notifications`、`GET /api/admin/notification-deliveries` 和 `POST /api/notifications/{notificationID}/deliveries`，支持按目标商户、通知 ID、状态、来源 topic/event、渠道和 provider 查询或补录回执。
- 详情抽屉可按通知行预填通知台账、投递回执、补录回执、审计和来源 outbox 事件动作；补录回执被标记为高风险操作，执行前进入二次确认。
- Admin Web RBAC 配置同步补 `notification:read` 和 `notification:write`，`ops_admin` 可读写通知回执，`support_admin` 可只读排查通知争议。
- 当前边界：这是运营可视化和人工排查入口首版，失败回执告警、失败重试编排、provider 执行器骨架、provider 回调验签入账和模板 payload 规范已由后续增量承接；真实短信/企业微信/订阅消息/push 生产账号、模板审批、provider sandbox 字段联调、通知偏好和跨端消息中心仍待补。

### 3.19 通知失败回执告警首版

- 新增 `NotificationFailureAlertEmission` 和 `POST /api/admin/notification-deliveries/failure-alerts/emit`，由 `notification:write` scope 守护，运营可按目标、渠道、provider 和 limit 汇总 failed 回执。
- 内存 Store 与 PostgreSQL-backed Store 均会生成 `notification.delivery_failed_alerts` outbox 事件，事件类型为 `notification.delivery_failed_alerts.emitted`，并写入 `admin.notification_delivery_failure_alerts.emitted` 审计；没有失败回执时返回 skipped。
- BFF 已允许该后台路径，Admin Web `notifications` 模块和详情抽屉已新增“投递失败告警”动作，执行前进入高风险二次确认。
- notification-worker 已订阅该 topic 并生成面向安全目标的失败告警通知 payload，outbox relay 默认 topic、Docker Compose 和 K8s 部署骨架已覆盖。
- 当前边界：这是失败回执到可靠 outbox 告警的首版，失败重试编排已由后续增量承接；仍不是真实短信、企业微信、微信订阅消息、商户端 push provider 执行器，也还没有升级策略、静默窗口和通知偏好。

### 3.20 通知失败重试编排首版

- 新增 `NotificationDeliveryRetrySchedule` 和 `POST /api/admin/notification-deliveries/retries/schedule`，由 `notification:write` scope 守护，运营可按目标、渠道、provider、limit 和 `retry_after_seconds` 汇总 failed 回执并安排重试。
- 内存 Store 与 PostgreSQL-backed Store 均会生成 `notification.delivery_retries` outbox 事件，事件类型为 `notification.delivery_retries.scheduled`，并把 `available_at` 对齐到 `retry_at`，让重试计划只在退避窗口后进入 ready 队列；没有失败回执时返回 skipped 并保留审计证据。
- 调度动作写入 `admin.notification_delivery_retries.scheduled` 审计，payload 记录计划数量、渠道、provider、退避秒数、`retry_at`、重试策略和 outbox event id。
- BFF 已允许该后台路径，Admin Web `notifications` 模块和通知详情抽屉已新增“安排投递重试”动作，并纳入高风险二次确认；失败行会预填商户、渠道、provider、limit 和默认退避秒数。
- notification-worker 已订阅 `notification.delivery_retries` 并生成面向安全/通知运营目标的重试计划通知 payload；outbox relay 默认 topic、Docker Compose 和 K8s 部署骨架已覆盖。
- 当前边界：这是可靠重试编排、延迟 outbox 和审计证据首版，provider 执行器、provider 回调验签、模板 payload 规范、端侧通知偏好 UI 和通知偏好审批应用已由后续增量承接；仍缺真实渠道生产账号、模板审批、provider sandbox 字段联调、渠道模板/供应商配置审批、升级策略和跨端消息中心。

### 3.21 通知 provider 执行器首版

- `notification.delivery_retries` outbox payload 已携带原始通知快照，除了失败回执 ID 外还包含原通知标题、正文、目标和类型，避免 provider 重试时只拿到 notification id 无法重发正文。
- `notification-worker` 新增 provider dispatcher，可把初始通知或重试事件转成短信、企业微信、微信订阅消息、push 的 provider dispatch，并按 `NOTIFICATION_PROVIDER_CHANNELS`、`NOTIFICATION_PROVIDER_ENDPOINTS` 或单渠道 endpoint 进行配置。
- provider dispatcher 会调用配置 endpoint/adapter；成功时把 provider message id 写回 `/api/notifications/{notificationID}/deliveries`，失败时写回 `provider_not_configured`、HTTP 错误或 adapter 异常，避免把未触达的外部渠道误标为 delivered。
- 重试事件会对原始通知 ID 写回 provider 回执，站内重试计划通知仍写给安全/通知运营目标，便于运营同时看到“重试计划已触发”和“原通知真实渠道回执”。
- Docker Compose 与 K8s 部署骨架已新增 `notification-worker` provider 环境变量位，后续可挂真实渠道 endpoint、token 和模板配置。
- 当前边界：这是 provider 执行、回执写回和部署配置骨架首版；provider 回调验签、模板 payload 规范、通知偏好 UI 和通知偏好审批应用已由后续增量承接，仍不是生产短信、企业微信、微信订阅消息或 push 账号接入，也还缺渠道模板审批、供应商配置审批、升级策略和跨端消息中心。

### 3.22 通知 provider 回调验签入账首版

- 新增 `POST /api/notifications/provider-callback`，外部通知渠道可回调 delivered/failed/queued 状态、provider message id、错误码、错误信息和 callback 时间。
- API 可通过 `NOTIFICATION_PROVIDER_CALLBACK_SECRET` 启用 HMAC-SHA256 canonical lines 验签；secret 为空时保留本地开发通道，生产配置后缺失或错误签名会返回 `INVALID_NOTIFICATION_PROVIDER_SIGNATURE`。
- 回调入账会归一化 channel/provider/status/timestamps，按 callback idempotency key 或 provider message id 生成幂等键，写入 `platform_notification_deliveries`，重复回调返回同一条回执。
- BFF 已允许该公开回调路径；`notification-worker` 新增 `signProviderCallback` 与 `normalizeProviderCallbackPayload`，可在 provider sandbox/mock 联调时生成与后端一致的签名 payload；Docker Compose 与 K8s 部署骨架已预留 `NOTIFICATION_PROVIDER_CALLBACK_SECRET`/secretKeyRef。
- 当前边界：这是通用 provider 回调安全边界和投递台账入账首版；模板 payload 规范、通知偏好 UI 和通知偏好审批应用已由后续增量承接，仍缺真实短信、企业微信、微信订阅消息、push 生产账号、渠道模板审批、各 provider 字段映射/sandbox 联调、供应商配置审批、升级策略和跨端消息中心。

### 3.23 通知 provider 模板映射与渠道 payload 规范首版

- `notification-worker` 新增 `normalizeProviderTemplates`、`applyProviderTemplate` 和 channel payload 构建逻辑，可解析 `NOTIFICATION_PROVIDER_TEMPLATES`。
- provider dispatch 会按 notification type/template_key、channel 和 provider 查找模板配置，生成 `template_id`、`template_params` 和 `provider_payload`，并保持原 `title/body/target/idempotency_key` 兼容字段。
- 微信订阅消息 payload 覆盖 `touser`、`template_id`、`page`、`data`、`lang`；短信 payload 覆盖 `phone`、`template_code`、`sign_name`、`params`；企业微信 payload 覆盖 `touser`、`agentid`、`msgtype`、`template_id`、`text/params`；push payload 覆盖 `audience`、`title`、`body`、`template_id`、`extras`。
- Docker Compose 与 K8s 部署骨架已预留 `NOTIFICATION_PROVIDER_TEMPLATES`，后续可把审核后的渠道模板映射注入 worker。
- 当前边界：这是 provider 模板配置与渠道 payload 结构首版；偏好过滤、静默窗口、端侧偏好设置 UI 和通知偏好审批应用已由后续增量承接，仍缺真实短信、企业微信、微信订阅消息、push 生产账号、渠道模板审批、各 provider sandbox 字段联调、供应商配置审批、升级策略和跨端消息中心。

### 3.24 通知偏好与静默窗口首版

- `notification-worker` 新增 `normalizeDeliveryPreferences` 和 `notificationDeliveryPreferenceDecision`，可解析 `NOTIFICATION_DELIVERY_PREFERENCES`。
- 偏好规则按 `default`、`target_role`、`target_role:target_id`、`type:{notification_type}` 和 `target_role:target_id:{notification_type}` 合并，支持 `enabled_channels`、`disabled_channels` 和 `quiet_hours`。
- 被偏好禁用或落入静默窗口的外部 provider 投递不会调用 provider endpoint/adapter，会生成 `queued` 回执，分别写入 `notification_preference_disabled` 或 `notification_quiet_window` 原因，保留运营证据且避免误触达。
- Docker Compose 与 K8s 部署骨架已预留 `NOTIFICATION_DELIVERY_PREFERENCES`。
- 当前边界：这是 worker 侧偏好过滤和静默窗口首版；偏好后端账本、worker 后端偏好读取、静默 queued 调度、静默到期自动扫描、端侧偏好设置 UI 和通知偏好审批应用已由后续增量承接，仍缺策略灰度、升级策略、跨端消息中心和真实渠道生产联调。

### 3.25 通知偏好后端账本与 API 首版

- 新增 `NotificationPreference`、`SaveNotificationPreferenceRequest` 和 `NotificationQuietHours` 合约，支持目标/通知类型级 `enabled_channels`、`disabled_channels` 与 `quiet_hours`。
- 内存 Store 与 PostgreSQL-backed Store 均支持 `NotificationPreferences`、`SaveNotificationPreference` 和 `SaveNotificationPreferenceWithAudit`；PostgreSQL 路径新增 `platform_notification_preferences` 规范表、偏好 key 唯一约束和目标/类型索引。
- 新增 `GET/PUT /api/merchant/notification-preferences`，商户只能读写自身通知类型偏好；新增 `GET/PUT /api/admin/notification-preferences`，`support_admin` 可只读排查，`ops_admin` 可写入，运营写入会记录 `admin.notification_preferences.saved` 审计。
- BFF 已允许商户端和运营端偏好路径；Store、HTTP、BFF 测试与架构守卫已覆盖。
- 当前边界：这是偏好持久化和 API 首版；notification-worker 动态读取后端偏好、静默 queued 调度、静默到期自动扫描、商户端偏好 UI、管理端偏好操作入口、用户端偏好 UI、偏好变更主动失效、批量保存和审批应用已由后续增量承接，仍需真实渠道生产联调。

### 3.26 通知 worker 后端偏好读取首版

- 后端通知偏好列表新增 `preference_key` 精确查询，worker 可按决策 key 拉取偏好，不需要扫全量偏好账本。
- `notification-worker` 新增 `deliveryPreferencesFromRecords` 和 `createNotificationPreferenceResolver`，会按 `default`、角色、目标、通知类型和目标+通知类型五类 key 查询 `/api/admin/notification-preferences`。
- provider 投递前会合并静态 `NOTIFICATION_DELIVERY_PREFERENCES` 与后端偏好账本；后端偏好可以禁用短信/企微/订阅消息/push 或设置静默窗口。
- 偏好读取失败时采取失败关闭策略：不调用外部 provider，而是记录 `queued` 回执和 `notification_preference_lookup_failed`，避免绕过商户偏好误触达。
- 当前边界：这是 worker 动态读取偏好首版；静默 queued 调度、静默到期自动扫描、商户端偏好 UI、管理端偏好操作入口、用户端偏好 UI、偏好变更主动失效、批量保存和审批应用已由后续增量承接，仍需真实渠道生产账号和 provider 字段联调。

### 3.27 通知静默 queued 再投递调度首版

- `NotificationDeliveryRetryScheduleRequest` 新增 `status`、`error_code` 和 `retry_at`，默认仍兼容 failed 回执退避重试。
- `GET /api/admin/notification-deliveries` 支持按 `error_code` 查询，运营可筛出 `status=queued` 且 `error_code=notification_quiet_window` 的静默窗口回执。
- 内存 Store 与 PostgreSQL-backed Store 可把 queued quiet-window 回执调度成 `notification.delivery_retries` outbox 事件，事件 `available_at` 对齐指定 `retry_at`，payload 继续携带原通知快照，worker 到点按原通知重发 provider。
- Admin Web 通知回执查询新增错误码筛选，重试表单新增 queued/failed 状态、错误码和指定重试时间，保留高风险二次确认。
- 当前边界：这是静默回执到延迟重投的可运营调度首版；自动扫描、商户端偏好 UI、管理端偏好操作入口、用户端偏好 UI、偏好变更主动失效、批量保存和审批应用已由后续增量承接，仍需真实渠道生产账号和 provider sandbox 字段联调。

### 3.28 通知静默到期自动扫描调度首版

- 静默窗口 queued provider 回执现在会记录 `retry_at`，用于表达偏好静默结束后的最早可重投时间。
- `GET /api/admin/notification-deliveries` 支持 `retry_at_before` 查询；后台可通过 `POST /api/admin/notification-deliveries/quiet-window-retries/schedule` 扫描 `status=queued`、`error_code=notification_quiet_window` 且 `retry_at` 已到期的回执，并生成 `notification.delivery_retries` 延迟 outbox。
- `notification-worker` 新增静默重试自动调度循环，可通过 `NOTIFICATION_QUIET_RETRY_AUTO_SCHEDULE`、`NOTIFICATION_QUIET_RETRY_INTERVAL_MS` 和 `NOTIFICATION_QUIET_RETRY_LIMIT` 在部署侧开启；Docker Compose、K8s、BFF、Admin Web 操作目录、详情抽屉、测试和架构守卫已接入。
- 当前边界：这是静默到期后自动扫描并调度可靠重投的首版；商户端偏好 UI、管理端偏好操作入口、用户端偏好 UI、偏好变更主动失效、批量保存和审批应用已由后续增量承接，仍需真实短信/企业微信/订阅消息/push 生产账号、模板审批和 provider 字段联调。

### 3.29 商户端通知偏好设置首版

- `apps/merchant-flutter` 新增 `lib/features/notifications/merchant_notification_preferences_page.dart`，首页新增“通知”入口。
- 商户可按 `order.status_changed`、`merchant.qualification_reviewed` 配置微信订阅、短信、企业微信、端内 Push 的开关，并可开启静默时段、设置开始/结束时间、时区和静默渠道。
- 商户端 API client 新增 `getMerchantNotificationPreferences` 与 `saveMerchantNotificationPreference`，直接调用 `GET/PUT /api/merchant/notification-preferences`，保存后的偏好进入后端账本，供 `notification-worker` 投递前读取。
- 架构守卫已检查页面、路由、API client、渠道字段、通知类型和 `quiet_hours` payload，避免后续把商户自助偏好入口删回去。
- 当前边界：这是商户端偏好设置 UI 首版；管理端偏好操作入口、用户端偏好 UI、偏好变更主动失效、批量保存和审批应用已由后续增量承接，仍需真实渠道账号/模板审批和 provider sandbox 字段联调。

### 3.30 管理端通知偏好操作入口首版

- Admin Web 操作目录新增 `notification-preferences` 和 `notification-preference-save`，可调用运营侧 `GET/PUT /api/admin/notification-preferences`。
- 通知运营页新增查询/保存通知偏好的快捷动作，详情抽屉会按通知目标、来源 topic 和失败渠道预填目标角色、目标 ID、通知类型、禁用渠道和 `quiet_hours` JSON。
- 保存通知偏好被纳入高风险二次确认，原因明确标注会改变外部触达渠道和静默窗口；执行后继续进入操作结果追踪与失败回放。
- Admin Web 请求构造已支持 `json` 字段类型，`quiet_hours` 可从表单 JSON 解析成对象，非法 JSON 会阻断请求并提示。
- 架构守卫和 Admin Web 单测已覆盖操作目录、通知视图动作、详情抽屉预填、高风险确认、GET/PUT 请求构造、CSV 渠道数组和 `quiet_hours` JSON。
- 当前边界：这是运营侧偏好操作入口首版，仍不是完整通知策略中心；用户端偏好 UI、批量保存和审批应用已由后续增量承接，后续还需要真实渠道账号/模板审批和 provider sandbox/生产字段联调。

### 3.31 用户端通知偏好设置首版

- 后端新增 `GET/PUT /api/user/notification-preferences`，用户角色只能读取和保存自身 `target_role=user` 偏好；保存时服务端会覆盖请求里的 `target_role/target_id`，防止越权写入商户或其他用户偏好。
- BFF 已放行用户端通知偏好 GET/PUT 路由，回归测试覆盖 Authorization 透传和保存 payload。
- 原生微信小程序新增 `pages/notification-preferences/index` 和首页入口，用户可按 `order.status_changed`、`after_sales.updated`、`coupon.campaign` 配置微信订阅、短信、App Push 开关、静默时段和静默渠道。
- 小程序 API client 新增 `getUserNotificationPreferences` 与 `saveUserNotificationPreference`，页面加载时读取既有偏好，保存时写入 `enabled_channels`、`disabled_channels`、`quiet_hours` 和更新时间。
- 架构守卫已覆盖用户端页面、路由、API client、通知类型、渠道字段和 `quiet_hours` payload。
- 当前边界：这是用户端偏好设置 UI 与 API 首版；偏好缓存与失败关闭、偏好变更主动失效、批量保存和审批应用已由后续增量承接，后续还需要真实渠道账号/模板审批、provider sandbox/生产字段联调、升级策略和跨端消息中心。

### 3.32 通知 worker 偏好缓存与失败关闭首版

- `createNotificationPreferenceResolver` 新增按 preference key 的内存缓存，默认 `NOTIFICATION_PREFERENCE_CACHE_TTL_MS=30000`、`NOTIFICATION_PREFERENCE_CACHE_STALE_MS=300000`、`NOTIFICATION_PREFERENCE_CACHE_MAX_KEYS=500`。
- 新鲜缓存命中时不再重复请求 `/api/admin/notification-preferences`，TTL 过期后再刷新，降低高频通知对核心 API 和偏好账本的读放大。
- 刷新失败且仍在 stale 窗口内时继续使用旧偏好；没有可用缓存时保持原来的失败关闭策略，外部 provider 投递转为 `queued` 回执并记录 `notification_preference_lookup_failed`。
- Docker Compose 与 K8s notification-worker 已预留缓存参数位，架构守卫覆盖代码、测试和部署配置，worker 单测覆盖 TTL 命中、TTL 过期刷新和 stale-if-error。
- 当前边界：这是单 worker 进程内短 TTL 缓存首版；偏好变更事件驱动主动失效、批量策略保存和审批应用已由后续增量承接，后续仍需要策略灰度和真实渠道联调。

### 3.33 通知偏好变更事件与 worker 主动失效首版

- 保存通知偏好现在会生成 `notification.preferences_changed` outbox 事件，payload 携带 `preference_key`、`preference_keys`、目标角色/ID、通知类型、渠道和静默窗口。
- PostgreSQL-backed Store 会在同一事务内 upsert `platform_notification_preferences` 并插入 `platform_outbox_events`，避免偏好已保存但缓存失效事件丢失。
- outbox relay 默认 topic、Docker Compose 和 K8s 已加入 `notification.preferences_changed`。
- `notification-worker` 订阅该 topic，新增 `notificationPreferenceInvalidationKeys` 和 `invalidateNotificationPreferenceCache`；消费偏好变更事件时只删除对应 resolver 缓存 key，不会创建站内通知。
- 当前边界：这是偏好变更可靠事件和 worker 本地缓存主动失效首版；批量策略和审批应用已由后续增量承接，后续仍需要策略灰度、真实渠道账号/模板审批、provider sandbox/生产字段联调、升级策略和跨端消息中心。

### 3.34 通知偏好批量保存与策略审计首版

- 后端新增 `SaveNotificationPreferenceBatchRequest` 和 `NotificationPreferenceBatchSaveResult`，运营可通过 `POST /api/admin/notification-preferences/batch` 一次保存最多 50 条偏好策略，并必须提供变更原因。
- 内存 Store 会校验批量内重复 preference key，逐条保存偏好并为每条偏好生成 `notification.preferences_changed` outbox 事件，批量动作写入 `admin.notification_preferences.batch_saved` 审计。
- PostgreSQL-backed Store 会在同一事务内 upsert 多条 `platform_notification_preferences`、插入对应 outbox 事件并写入批量审计，避免批量策略已落账但缓存失效或审计证据缺失。
- BFF 已允许 `/api/admin/notification-preferences/batch`，Admin Web 通知运营页新增“批量保存通知偏好”操作，`preferences` JSON 和变更原因进入高风险二次确认。
- 架构守卫覆盖合约、仓储、HTTP、BFF、Admin Web 操作目录和测试，防止批量策略入口退化。
- 当前边界：这是批量策略保存与审计首版，不是完整策略治理中心；审批应用已由后续增量承接，后续仍需要策略灰度、真实渠道账号/模板审批、provider sandbox/生产字段联调、升级策略和跨端消息中心。

### 3.35 通知偏好变更审批与应用首版

- 后端新增 `GET/POST /api/admin/notification-preferences/change-requests`，运营可提交待审批通知偏好变更申请，申请只写 `admin.notification_preferences.change_requested` 审计，不会直接改动偏好账本。
- 新增 `POST /api/admin/notification-preferences/change-requests/{changeRequestID}/review`，另一名具备通知写权限的管理员可审批或驳回，服务端禁止申请人自审，审批结果写入 `admin.notification_preferences.change_reviewed`。
- 新增 `POST /api/admin/notification-preferences/change-requests/{changeRequestID}/apply`，只允许 `approved` 状态进入应用阶段，且禁止申请人自己应用；应用复用批量保存原子路径，同事务写入偏好、`notification.preferences_changed` outbox 和 `admin.notification_preferences.change_applied` 审计。
- 申请台账从审计日志重建，支持按 `pending_approval`、`approved`、`rejected`、`applied` 查询，并返回各状态计数、申请人、审批人、应用人、批次和审计 ID。
- BFF 已代理申请/审批/应用路由，Admin Web 通知运营操作目录已接入查询、提交、审批和应用动作，三类写动作均进入高风险二次确认。
- HTTP 回归测试覆盖无权限提交、自审拦截、审批后应用、申请人不能应用、已应用可查询、驳回后不能应用；BFF/Admin Web/架构守卫测试覆盖代理和操作目录。
- 当前边界：这是通知偏好治理的申请、审批、应用首版，不等于完整通知策略中心；灰度应用已由后续增量承接，后续仍需要真实短信/企业微信/订阅消息/push 生产账号、模板审批、provider sandbox/生产字段联调、升级策略和跨端消息中心。

### 3.36 通知偏好变更灰度应用首版

- 通知偏好变更申请新增 `rollout` 策略，支持 `all`、`target_ids` 和 deterministic `percentage` 三种模式，并可通过 `max_targets` 控制首批目标范围。
- 申请阶段会把 rollout 与 `preference_requests`、`preference_keys`、`requested_count` 一起写入 `admin.notification_preferences.change_requested` 审计，审批阶段保留同一策略，避免审批后临时扩大影响面。
- 应用阶段会按 rollout 计算实际落账偏好；被跳过的 preference key 不进入批量保存、不触发 `notification.preferences_changed`，并在 `admin.notification_preferences.change_applied` 审计中记录 `applied_preference_keys`、`skipped_preference_keys`、applied/skipped count 和 rollout 参数。
- 申请台账可从审计日志恢复 rollout、applied/skipped 范围和批量保存结果，Admin Web 提交通知偏好变更时可填写灰度策略 JSON，BFF 与架构守卫已覆盖透传和防退化。
- 当前边界：这是通知偏好变更的首个可审计灰度应用能力；后续仍需策略回滚/升级批次、真实渠道生产联调、跨端消息中心和生产告警联动。

### 4. BFF 浏览器 CORS 白名单

- BFF 已支持本地管理端和 Flutter 调试来源。
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

### 8.1 管理端审计导出首版

- 已新增 `GET /api/admin/audit-logs/export`，复用 actor/action/target/after/before/limit 审计筛选条件导出 CSV。
- 导出返回 `filename`、`row_count`、`generated_at`、`content_type` 和 CSV 内容，CSV 包含审计主体、动作、目标、请求、完整性状态和服务端白名单 payload。
- 导出动作本身写入 `admin.audit_logs.exported`，payload 记录导出格式、行数、筛选条件和生成时间，避免审计导出无人留痕。
- BFF、Admin Web 操作目录、审计适配层、HTTP 回归测试和架构守卫已覆盖导出入口。

### 8.2 管理端审计留存/告警健康报告首版

- 已新增 `GET /api/admin/audit-logs/retention-report`，默认按 2555 天留存、180 天热存和 500 条完整性抽样生成审计健康报告。
- 报告返回总日志数、最早/最新时间、过期日志、冷归档候选、完整性失败、导出事件、关键动作覆盖、缺失关键动作和 `ok`/`warning`/`critical` 状态。
- 内存 Store 可验证完整性篡改、留存过期、冷归档候选和关键动作缺口；PostgreSQL-backed Store 使用规范化 `audit_logs` 聚合查询总量、留存、热存、导出和动作覆盖。
- BFF、Admin Web 操作目录、审计适配层、HTTP 回归测试和架构守卫已覆盖报告入口。
- 这只是留存/告警的可见性首版，后续仍需生产 WORM bucket 策略、归档校验生产演练回查、真实告警渠道投递、导出审批/水印和 KMS/链式不可抵赖签名。

### 8.3 管理端审计留存告警 outbox 投递首版

- 已新增 `POST /api/admin/audit-logs/retention-alerts/emit`，复用审计留存报告口径，把 critical/warning 告警投递为 `audit.retention_alerts` outbox 事件。
- 投递动作会写入 `admin.audit_retention_alerts.emitted` 审计，payload 只保留告警数量、critical/warning 数量、留存窗口、完整性失败、冷归档候选、outbox event id 和幂等键。
- 新增 `audit:write` scope，`security_auditor` 仍只能读取审计报告和日志，不能触发告警投递；Admin Web 与 BFF 已接入投递入口。
- `notification-worker` 已订阅 `audit.retention_alerts`，后续可把该 topic 接到短信、企业微信、电话值班、工单或事故流程。
- 这仍不是完整真实渠道告警，后续还要补告警路由策略、值班升级、静默/抑制窗口和投递回执。

### 8.4 管理端审计 WORM/冷归档请求首版

- 已新增 `POST /api/admin/audit-logs/archive/request`，按热存天数和 limit 筛选冷归档候选审计日志。
- 请求会生成 `sha256:v1` manifest hash、归档路径、manifest entries 和 `audit.archive_requested` outbox 事件，归档事件 payload 包含归档 ID、存储前缀、归档 key、日志数量、完整性失败数和 manifest 摘要。
- 归档请求会写入 `admin.audit_archive.requested` 审计；`security_auditor` 仍保持只读，只有具备 `audit:write` 的管理员角色可触发请求。
- BFF、Admin Web 操作目录、审计适配层、HTTP 回归测试、Store 测试和架构守卫已接入。
- 这仍是归档请求/manifest 首版，不是归档上传执行器、生产 bucket 保留策略、冷归档删除审批或法律级不可抵赖归档。

### 8.5 管理端审计归档 worker 首版

- 已新增 `services/audit-archive-worker`，直接领取 `audit.archive_requested` outbox 事件，避免通用 relay 在归档文件上传前把事件标记完成。
- worker 会按后端 manifest hash 合约校验归档清单，校验失败则标记 outbox failed 并进入 retry/dead-letter 策略。
- worker 会上传 `application/x-ndjson` 归档文件，文件包含 manifest header 和逐条审计 manifest entry，并写入对象锁请求头、保留期和审计 hash 元数据。
- Docker Compose workers profile 和 Kubernetes 2 副本 Deployment 已补 `audit-archive-worker`；通用 `outbox-relay-worker` 只继续 relay `audit.retention_alerts` 等普通事件。
- 这仍不是最终生产 WORM 治理：还要补 bucket object-lock 强制策略、归档校验生产演练回查、保留期删除审批、KMS/链式不可抵赖签名和真实存储演练报告。

### 8.6 审计归档完成回写和记录查询首版

- 已新增 `POST /api/admin/audit-logs/archive/complete`，由归档 worker 在对象上传成功后回写归档完成证据，写入 `admin.audit_archive.completed` 审计。
- 完成证据包含归档 ID、storage key、manifest hash、content hash、字节数、对象锁模式、保留期、上传时间和 outbox event id。
- 已新增 `GET /api/admin/audit-logs/archive/records`，从审计账本重建归档完成记录，支持按归档 ID、时间范围和 limit 查询。
- BFF 与 Admin Web 操作目录已接入归档记录查询，worker 已在成功上传后先回写完成证据，再把 outbox 标记 published。
- 后续 `8.7` 已补归档对象下载校验/回查首版。

### 8.7 审计归档对象下载校验/回查首版

- 已新增 `POST /api/admin/audit-logs/archive/verify`，根据已完成归档记录从配置的 `AUDIT_ARCHIVE_DOWNLOAD_BASE_URL` 读取 JSONL 归档对象。
- 校验会对比归档文件 content hash、manifest header 中的 archive id / manifest hash、完成记录 bytes 和 manifest entry 数，返回 `verified` 或 `failed` 状态。
- 校验动作写入 `admin.audit_archive.verified` 审计，payload 记录 expected/actual content hash、expected/actual bytes、manifest/hash/log count 匹配结果和错误码。
- `security_auditor` 可触发该只读回查，BFF 与 Admin Web 操作目录已接入，Store/HTTP/Admin/BFF/架构守卫测试已覆盖。
- 后续 `8.8` 已补校验历史查询首版，`8.9` 已补审计中心可视化面板首版；这仍不是最终法律级不可抵赖归档，后续还要补生产 WORM bucket 强制对象锁、保留期删除审批、KMS/链式签名和真实存储演练报告。

### 8.8 审计归档校验历史查询首版

- 已新增 `GET /api/admin/audit-logs/archive/verifications`，从 `admin.audit_archive.verified` 审计账本重建归档校验历史。
- 查询支持按 `archive_id`、`status`、`after`、`before` 和 `limit` 过滤，可回看每次校验的 expected/actual content hash、bytes、manifest header 匹配结果、条目数和错误码。
- 内存 Store 与 PostgreSQL-backed Store 共用同一套审计 payload 反序列化逻辑，避免只在本地可查、生产 SQL 不可查。
- BFF 已放行该路由，Admin Web 操作目录新增“归档校验历史”，HTTP/Store/Admin/BFF/架构守卫测试已覆盖。
- 后续 `8.9` 已补审计中心可视化面板首版；这仍不是生产演练报告和法律级不可抵赖归档，后续还要补生产 WORM bucket 强制对象锁、保留期删除审批、KMS/链式签名和真实存储演练报告。

### 8.9 审计归档校验历史可视化面板首版

- Admin Web 审计检索页已内嵌“归档校验历史”面板，管理员可在同一页面按归档 ID、状态和条数查询 `/api/admin/audit-logs/archive/verifications`。
- 面板展示归档 ID、storage key、状态、校验时间、manifest hash、expected/actual content hash、expected/actual bytes、manifest/log count、匹配摘要和异常码。
- 每条校验历史支持展开原始详情，便于把 API 回查结果、审计账本和运营排查页面串起来。
- 筛选条件会写入 Admin Web 本地状态，刷新后仍保留常用查询口径；Admin Web 单测、架构守卫和本地浏览器渲染检查已覆盖。
- 这仍不是生产 WORM bucket 策略、保留期删除审批、KMS/链式不可抵赖签名或真实存储演练报告。

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
- 这仍不是最终法律级不可抵赖方案，后续还要补 KMS/Vault 密钥轮换、链式账本、生产 WORM bucket 策略、归档校验生产演练回查和真实告警渠道投递。

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
- 当前只是后台关键写路径与审计同事务首版，后续仍需继续扫描其他后台写路径、生产 WORM bucket 策略、归档校验生产演练回查、真实告警渠道投递和不可抵赖归档。

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
- 字段级/租户级 RBAC、权限变更产品化审批页面、生产 WORM bucket 策略、归档校验生产演练回查、保留期删除审批、真实告警渠道投递、KMS/链式不可抵赖签名和审计策略治理。
- 剩余关键业务写操作与审计写入同事务强制提交，继续扫描后台配置、运营处置、资金和风控写路径。
- 真实 IM 动态成员、token 撤销联动、断线重连、顺序保障、客服质检审计同事务和离线补偿生产化。
- RTC 语音通话信令、通话审计和订单/会话关联。
- PostgreSQL HA、Redis Cluster、Kafka 多 Broker、MinIO 生产权限、Vault/KMS 密钥治理。
- 10 万在线连接压测、订单写入压测、抢单冲突压测、故障注入和容灾报告。

### P1 产品闭环缺口

- 用户端邀请页、邀请奖励、钱包账单、售后入口、评价、收藏、积分会员、搜索、定位、消息页。
- 商户端完整邀请注册 UI、店铺装修、店铺展示页管理、商户钱包、结算、消息页、资质过期强弹窗。
- 骑手端完整任务详情、地图导航、轨迹、收入、钱包、账单、提现、违规申诉。
- 优惠券资金责任闭环：商户自发券、平台补贴券、商户确认活动券、结算审计。
- 群聊、红包、拼手气红包、官方群默认静音、商户群领券限制的 API 和页面闭环。
- 圈子/小微墙、饭搭问卷、真实性承诺、免责协议、资料审核、举报拉黑和举报处置已有首版；同校/同楼隐私策略、设备风控和 PostgreSQL 规范化仍待补。
- 买药完整药房/医务室资质、商品、监管审计和履约。
- 快递/跑腿完整下单、计价、异常处理和骑手履约。

### P2 工程与运营缺口

- 真实 Kafka/NATS broker 运维和 relay 积压恢复演练。
- 通知通道真实接入：微信订阅消息、短信、企业微信、Push 的生产账号、模板审批、provider sandbox 字段联调、真实渠道联调和跨端消息中心；站内通知账本、投递回执账本、Admin Web 通知运营页、失败回执可靠告警、失败重试编排、provider 执行器骨架、provider 回调验签入账、模板 payload 规范、偏好过滤、静默窗口、静默 queued 再投递调度、静默到期自动扫描调度、商户端通知偏好设置、管理端通知偏好操作入口、用户端通知偏好设置、worker 偏好缓存失败关闭、偏好变更主动失效、通知偏好批量保存和通知偏好变更审批/应用已有首版。
- 对象存储接真实 MinIO SDK、STS/Vault 临时凭证、隔离 bucket、扫描/删除失败告警。
- 更多管理端后端业务明细接口、首页卡片、精选商品、首页活动、圈子饭搭、红包、开放平台、系统日志等完整后台面板。
- 自动化发布、镜像构建、灰度、回滚、release/tag 策略和生产 runbook。

## 下一批优先推进顺序

### 第一批：后台审计中心补全

- 继续把剩余关键业务写操作和 `audit_logs` 写入推进到同一业务事务内强制提交，退款策略配置、管理端订单退款、售后审核、订单状态补偿、对象清理完成/失败、outbox 运维、商户资质审核和商户/骑手邀约已完成首版，下一步继续扫后台配置、运营处置、资金和风控写路径。
- 在服务端 RBAC 策略矩阵与变更申请/审批/应用/回滚审计首版基础上继续补字段级权限、站点/商户数据域、产品化审批队列和菜单隐藏策略。
- 补生产 WORM bucket 策略、归档校验生产演练回查、保留期删除审批、真实告警渠道投递、KMS/链式不可抵赖签名和冷热归档策略治理。

### 第二批：管理端明细接口与审核闭环

- 订单、售后、商户资质、骑手/站长后端明细接口。
- Outbox 事故处置辅助信息和更完整的审核辅助信息；售后审核、订单退款、商户资质审核、商户资质审核结果可靠通知、商户站内通知中心、通知运营查询、通知投递回执、Admin Web 通知运营页、通知失败回执告警、通知失败重试编排、通知 provider 执行器、通知 provider 回调验签入账、通知 provider 模板映射与渠道 payload 规范、通知偏好与静默窗口、通知静默 queued 再投递调度、通知静默到期自动扫描调度、商户端通知偏好设置、管理端通知偏好操作入口、用户端通知偏好设置、通知 worker 偏好缓存失败关闭、通知偏好变更主动失效、通知偏好批量保存与审批应用、Outbox 单事件恢复、Outbox 发布/失败人工处置、Outbox 领取/续租和 Outbox 死信分诊/解封表单首版已接入，后续继续补真实通知渠道、明细接口和审核/运维辅助信息。

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
- `git diff --check`

### 0.32 同单工作区同步刷新

- Admin Web 同单工作区在既有“联动结果区 + 打开订单/售后工作区”基础上，新增共享上下文同步首版：联动卡片会标记“已对齐 / 待同步”，并显示同步后的目标订单/售后/工单/用户 token。
- 联动结果区新增按分组筛选和上下文定位后的同步动作，支持一键“同步当前筛选 / 同步全部”，把已经挂出来的退款、客服工单、派单事件、审计卡片直接刷新到当前主结果上下文，不必手动逐张重挂。
- 相关实现集中在 `apps/admin-web/src/adminLinkedWorkspace.mjs` 与 `apps/admin-web/src/main.js`，并补了 `adminLinkedWorkspace.test.mjs` 覆盖共享参数重绑、待同步识别和 focus/filter 匹配逻辑。

### 0.33 同单工作区上下文切换

- Admin Web 同单工作区新增“切到此上下文”和上下文候选区，不必再先把联动卡提成主结果，运营就能直接围绕某张退款/客服/派单/审计卡继续排查。
- 工作区切换后会保留主结果，并补出“跟随主结果”回退入口；同时会把焦点定位带到新的订单/售后/工单/用户 token，让“筛选 + 同步 + 带参刷新”围绕同一上下文继续展开。
- 相关实现继续收口在 `apps/admin-web/src/adminLinkedWorkspace.mjs`、`apps/admin-web/src/main.js` 和 `apps/admin-web/src/styles.css`，并补充 `adminLinkedWorkspace.test.mjs` 覆盖上下文归一化、候选聚合和主焦点推导。

### 0.34 同单工作区分组同步与表单预填

- Admin Web 同单工作区继续补了一层操作台联动：除了“同步当前筛选 / 同步全部”，现在还会按退款、客服、派单、审计分组给出同步按钮，让运营能只刷新当前关注的一条线。
- 操作台当前表单会跟随工作区上下文自动预填 `order_id / related_order_id / request_id / ticket_id / user_id / target_type / target_id` 等关键字段，切换上下文或切换操作时都不必再手工重抄参数。
- 这轮也顺手修掉了联动结果卡分组 badge 一直落成“其他”的显示问题；实现仍集中在 `apps/admin-web/src/adminLinkedWorkspace.mjs`、`apps/admin-web/src/main.js`、`apps/admin-web/src/styles.css`，并补了 `adminLinkedWorkspace.test.mjs` 覆盖分组同步与表单预填断言。

### 0.35 同单工作区上下文自动刷新

- Admin Web 同单工作区继续把“切完上下文还得自己点同步”的那一步往下压：现在切到联动卡上下文、从候选上下文切换、回到“跟随主结果”，以及把某张联动卡设为主结果后，当前筛选下待同步的联动卡会自动刷新。
- 这轮还修正了焦点匹配口径：当工作区焦点已经切到新订单/售后/工单/用户时，联动卡同步判断会同时识别卡片的当前 token 和同步后的目标 token，避免“刚切完上下文最该刷新的卡反而被焦点漏掉”。
- 相关实现仍集中在 `apps/admin-web/src/adminLinkedWorkspace.mjs` 与 `apps/admin-web/src/main.js`，并扩充 `adminLinkedWorkspace.test.mjs` 覆盖焦点态下的同步动作与分组行为。

### 0.36 同单工作区同步反馈

- Admin Web 同单工作区继续往“真工作台”收口：现在会直接显示同步中/同步反馈文案，并在工具条里把退款、客服、派单、审计四组状态汇总出来。
- 当某次同步正在进行时，当前筛选、全部同步和分组同步按钮会进入对应的“刷新中”态；完成后会立刻回写“已刷新当前筛选/全部/某分组 N 张卡”，让运营知道刚刚到底发生了什么。
- 这轮实现继续收口在 `apps/admin-web/src/adminLinkedWorkspace.mjs`、`apps/admin-web/src/main.js` 和 `apps/admin-web/src/styles.css`，并扩充 `adminLinkedWorkspace.test.mjs` 覆盖分组汇总口径。

### 0.37 同单工作区同步时间与分组完成感

- Admin Web 同单工作区继续把同步反馈补到“可确认”层级：同步反馈现在会附带最近一次完成时间和涉及分组摘要，比如“14:32:08 · 退款 2 / 客服 1”，不用再靠记忆判断刚刚刷了什么。
- 分组状态摘要也会在同步完成后短暂切到“刚刷新 N”，让运营能一眼看到这次同步影响了哪些线，而不是只剩一片“已对齐”。
- 这轮实现继续收口在 `apps/admin-web/src/adminLinkedWorkspace.mjs`、`apps/admin-web/src/main.js` 和 `apps/admin-web/src/styles.css`，并扩充 `adminLinkedWorkspace.test.mjs` 覆盖同步动作分组统计。

### 0.38 同单工作区局部失败反馈

- Admin Web 同单工作区继续把同步反馈补到“能处理异常”的层级：如果一次同步里只有部分卡刷新失败，反馈文案会直接带上失败张数，而不是笼统地只说“已刷新”。
- 退款、客服、派单、审计四组摘要也会同步带上“失败 N”红色状态，这样运营不用逐卡排查，先看工作区就知道是哪条线掉队了。
- 这轮实现继续收口在 `apps/admin-web/src/main.js`，复用 `adminLinkedWorkspace.mjs` 里的动作分组统计能力，把成功/失败按组回写到工作区状态里。

### 0.39 同单工作区失败项重试

- Admin Web 同单工作区继续往“能补救”推进：工作区现在会在失败后给出“重试失败项”入口，并支持按退款/客服等失败分组就地重试。
- 重试链路会沿用当前工作区上下文、当前筛选和当前焦点，不需要运营先切操作台、再抄参数、再重新挂卡。
- 这轮实现继续收口在 `apps/admin-web/src/adminLinkedWorkspace.mjs`、`apps/admin-web/src/main.js` 和 `apps/admin-web/src/styles.css`，并补充 `adminLinkedWorkspace.test.mjs` 覆盖失败项作用域与分组统计。

### 0.40 同单工作区失败定位与单卡重试

- Admin Web 同单工作区继续往“能直接排错”推进：联动结果区现在新增“失败”筛选，失败卡会直接展示失败原因，不用再先展开原始 JSON 才知道挂在哪。
- 失败卡片本身也补了“重试此卡”入口；如果运营只想补一张掉队的退款卡或客服卡，不必整组重跑。
- 工作区工具条同步补了“仅看失败项”，配合失败筛选可以先把红卡集中起来，再决定逐卡补救还是整组重试。

### 0.41 同单工作区失败卡请求事实

- Admin Web 红卡继续往“可解释”推进：失败卡现在会直接展示 HTTP 状态、最后一次请求路径和原参数摘要，不用再先展开原始返回才能判断是接口挂了还是参数跑偏了。
- 同时会把“按当前工作区上下文重试后将改写成什么参数”直接显示出来，运营能先确认这次补跑会不会把 `order_id / request_id / user_id` 切到新的主上下文。
- 这轮实现继续收口在 `apps/admin-web/src/adminLinkedWorkspace.mjs`、`apps/admin-web/src/main.js` 和 `apps/admin-web/src/styles.css`，并扩充 `adminLinkedWorkspace.test.mjs` 覆盖失败事实摘要与重试参数提示。

### 0.42 同单工作区失败卡回填与对照痕迹

- Admin Web 红卡继续往“可补救”推进：失败卡现在可以一键把当前重试参数回填到操作台，不必再手动抄 `order_id / request_id / ticket_id`。
- 联动结果 entry 也开始保留最近一次失败和最近一次成功的对照痕迹，红卡上能直接看出它是一直失败，还是刚从成功态掉下来。
- 这轮实现继续收口在 `apps/admin-web/src/adminLinkedWorkspace.mjs`、`apps/admin-web/src/main.js` 和 `apps/admin-web/src/adminLinkedWorkspace.test.mjs`，把痕迹保存、卡片渲染和回填动作连成一条线。

### 0.43 同单工作区回填确认与局部历史

- Admin Web 失败卡继续往“拿来就能处理”推进：点了“回填到操作台”后，操作台会直接出现一块回填确认面板，给出来源、当前参数、原失败请求事实，以及“执行当前查询 / 收起提示”两个下一步。
- 红卡里的最近失败/最近成功也不再是一行散文案，而是带时间戳、状态 badge 和参数摘要的局部历史列表，排查时能更快看出这张卡是不是刚刚恢复、又或者刚刚掉线。
- 这轮实现收口在 `apps/admin-web/src/adminLinkedWorkspace.mjs`、`apps/admin-web/src/main.js`、`apps/admin-web/src/styles.css` 和 `apps/admin-web/src/adminLinkedWorkspace.test.mjs`，把回填动作、确认态和局部历史渲染连成一条线。

## 当前风险提醒

- 当前代码里很多能力已经有首版和测试，但仍有大量链路是“骨架/首版/模拟接入”状态。
- 钱包、支付、退款、结算、审计、消息、RTC、容灾压测是商业化风险最高区域。
- 未完成真实支付、真实高可用、真实压测和生产演练前，不应对外承诺“已支撑 10 万在线”或“可商业上线”。
