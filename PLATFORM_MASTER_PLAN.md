# Infinitech 2.0 平台总计划

更新时间：2026-05-24
当前状态：规划基线和首批工程骨架已建立，正在推进商业级数据、安全、多主体认证和核心交易事务化。

最近完成、当前未完成和下一批推进顺序单独维护在 `docs/product/recent-progress-roadmap.md`。

## 1. 项目目标

Infinitech 2.0 要做成一个对标美团外卖的本地生活平台，第一阶段覆盖：

- 外卖：用户下单、商户接单、骑手配送、售后、评价。
- 团购：团购商品、券码、核销、退款、营销活动。
- 买药：药品/健康商品浏览、用药咨询入口、极速买药订单、配送追踪。
- 快递/跑腿：帮买、帮送、帮取、帮办，后续可扩展校园/同城快递。
- 钱包：余额支付、账单、退款、提现、网页充值跳转。
- 微信支付：用户支付、退款、回调、对账。
- 实时沟通：用户与骑手、用户与商家、商家与骑手、官方客服。
- 网络语音通话：基于 WebRTC 的站内语音，先做信令和点对点通话，后续扩 SFU/录音/质检。
- 邀请用户页：用户拉新、商户入驻、骑手入职邀请都纳入统一邀请体系。
- 第三方 OAuth/API：微信、后续支付宝/高德/短信/地图/其他平台数据接入。
- 容灾和容量：以 10 万同时在线为正式容量目标，必须有压测报告和容灾演练证据。
- 商家与店铺：商家账号负责主体、资质、结算和权限；店铺负责展示页、商品、团购套餐、服务能力和履约方式。

## 1.1 总架构

架构名称：自建/混合云 Kubernetes 上的模块化核心 API + 事件驱动 Worker + 多端 BFF + 实时网关架构。

执行原则：

- `api-go` 是模块化核心 API，订单、支付、钱包、抢单、派单保持强一致事务边界。
- BFF 负责多端聚合和差异适配，所有端只访问 BFF 或公开 API。
- Worker 通过 Kafka 消费事件，调度、支付、通知、集成、结算任务必须幂等。
- `realtime-gateway` 独立部署，使用 WebSocket-only 模式承载 IM、订单通知、骑手位置、抢单/派单事件和 RTC 信令。
- 基础设施按自建/混合云 Kubernetes 规划，包含 PostgreSQL HA + PgBouncer、Redis Cluster、Kafka、MinIO、Vault、Prometheus/Grafana/Loki/Tempo/OpenTelemetry。
- 最近推进路线维护在 `docs/product/recent-progress-roadmap.md`。
- 商业级验收口径维护在 `docs/product/commercial-readiness-checklist.md`，未满足对应验收证据前不得宣称生产可商用。

## 2. 硬性产品与技术约束

- UI 主题色、logo、整体视觉语言沿用旧版 `HCRXchenghong/infinitech`。
- 品牌主色固定为 `#009bf5`，辅色优先使用 `#0081cc`、`#0284c7`。
- 用户端第一阶段必须是原生微信小程序，不再沿用旧版用户端 uni-app 作为主线代码。
- 商户端和骑手端第一阶段分别使用独立 `uni-app` 工程。
- 管理端第一阶段为桌面 Web，加移动管理端 `uni-app`。
- 后续用户端、骑手端、商户端、管理端移动形态都预留迁移原生 App 的接口和模块边界。
- 骑手接单采用抢单加派单组合。
- 骑手在线状态下收到派单，每日允许一次免责取消；超过后按后台配置规则处理。
- 骑手订单计价规则必须后台可配置。
- 骑手端账号分为站长账号和骑手账号；站长可查看站点全部数据、配置每日任务时长并手动派单。
- 骑手和站长都采用邀约制注册，不开放公开自助注册。
- 骑手开始接单前必须缴纳 50 元保证金，或申请微信免押并审核通过。
- 商户开始接单前必须缴纳 50 元保证金，商户不支持免押。
- 支付必须支持微信支付和余额支付。
- 钱包充值入口只做跳转到外部浏览器充值网站，充值网站本身不在本阶段范围；后续通过 API/OAuth/二维码扫码登录打通。

## 3. 旧版复用结论

旧仓库不是直接续写对象，但可以复用以下资产和经验：

- 品牌资产：已复制到 `assets/brand/`。
- UI 色板：`#009bf5`、`#0081cc`、`#0284c7`、`#f3f4f6`、`#ffffff`。
- 用户端视觉：旧版 `user-vue` / `app-mobile` 的首页、分类、订单、消息、钱包、邀请页、买药、跑腿页面可作为原生小程序视觉参考。
- 商户端：旧版 `merchant-app` 的经营、订单、商品、店铺、钱包页面可作为 uni-app 起点。
- 骑手端：旧版 `rider-app` 已有抢单大厅、任务、钱包、历史订单、接单设置等页面，可作为 uni-app 起点。
- 管理端：旧版 `admin-vue` 已有订单、商户、骑手、支付中心、客服工作台、RTC 管理台、跑腿配置、邀请落地页等页面，可复用 UI 和信息架构。
- 后端：旧版 Go API、BFF、Socket.IO 网关已有订单、钱包、支付、团购、买药、消息、RTC、邀请、设置等模块雏形，可作为 2.0 领域模型参考。
- 容量验收：旧版已有 `Realtime 100k Acceptance` 文档，应保留“不能只口头宣称 10 万在线，必须压测留证”的原则。

不建议直接照搬：

- 旧版根目录交付形态过多，2.0 应重新定义清晰 monorepo 边界。
- 用户端旧版是 uni-app，小程序主线需重写为原生微信小程序。
- 支付命名里出现旧的 `ifpay` 习惯，2.0 需要统一为 `wechat_pay`、`balance`、`wallet_recharge` 等明确枚举。
- 支付宝和银行 sidecar 暂不纳入首发，避免支付链路过早分散。
- Tauri 桌面壳暂不作为主线，管理端先稳定 Web。

## 4. 目标工程结构

```text
.
├── apps
│   ├── user-wechat-miniprogram      # 原生微信小程序，用户端首发主线
│   ├── merchant-uni                 # 商户端 uni-app
│   ├── rider-uni                    # 骑手端 uni-app
│   ├── admin-web                    # 桌面管理端 Web
│   └── admin-uni                    # 移动管理端 uni-app
├── services
│   ├── api-go                       # 核心业务 API
│   ├── bff                          # Web/小程序聚合层
│   ├── realtime-gateway             # IM、通知、RTC 信令
│   ├── dispatch-worker              # 派单、抢单、骑手状态、计价
│   ├── payment-worker               # 微信支付、余额支付、退款、对账
│   └── integration-worker           # OAuth、外部 API、地图、短信、开放平台
├── packages
│   ├── contracts                    # API schema、枚举、响应规范
│   ├── design-tokens                # 颜色、间距、logo、图标规范
│   ├── domain-core                  # 领域规则和纯函数
│   ├── client-sdk                   # 请求、鉴权、实时连接 SDK
│   └── admin-core                   # 管理端菜单、权限、资源模型
├── infra
│   ├── docker                       # 本地开发依赖
│   ├── k8s                          # 生产部署模板
│   ├── observability                # Prometheus/Grafana/Loki/Tempo
│   └── loadtest                     # 10k/30k/60k/100k 压测脚本
├── docs
├── assets
└── scripts
```

## 5. 技术架构

### 5.1 前端端划分

用户端原生微信小程序：

- 使用微信小程序原生能力：`WXML/WXSS/TypeScript`。
- 视觉复刻旧版用户端，组件重新实现。
- 首发页面：登录、首页、圈子/小微墙、找饭搭、定位、分类、商家列表、商家详情、购物车、确认订单、支付、订单列表/详情、消息、客服、钱包、邀请、买药、跑腿。
- 旧版已有页面必须进入 2.0 路线图：搜索、今日推荐、商品详情、购物车、地址、备注、餐具数量、评价、售后、收藏、红包优惠、钱包充值/提现/账单、设置、修改手机号、积分商城、会员中心、反馈合作、公益。
- 支持微信登录、微信支付 JSAPI、订阅消息、地图定位、扫码。

商户端 uni-app：

- 独立工程，不和骑手端混用。
- 首发页面：登录、经营概况、订单管理、商品管理、资质资料、店铺设置、钱包、客服消息。
- 复用旧版 merchant-app 视觉和页面结构。

骑手端 uni-app：

- 独立工程。
- 首发页面：登录、抢单大厅、派单弹窗、我的任务、任务详情、钱包、历史订单、接单设置、客服。
- 旧版已有页面必须保留：账单明细、钱包充值/提现、收入明细、数据统计、头像、个人信息、修改手机号、修改密码、健康证、保险保障、违规申诉、骑手之家。
- 复用旧版 rider-app 视觉和页面结构。
- 支持在线/离线、位置上报、抢单、派单确认、每日一次免责取消。

管理端 Web：

- 桌面浏览器使用 Vue 3 + Element Plus 或同等级企业后台栈。
- 首发模块：仪表盘、订单、售后、用户、商户、骑手、骑手等级、商品/类目、首页卡片、精选商品、首页活动、优惠券、圈子/饭搭、团购、买药、跑腿配置、派单规则、骑手计价、支付中心、钱包/财务、积分会员、内容推送、客服工作台、RTC 审计、电话联系审计、数据备份恢复、风控、API/OAuth 管理、系统设置。

移动管理端 uni-app：

- 面向手机临时处理：订单监控、骑手调度、客服、支付/退款审核、系统告警。
- 不做完整桌面后台的镜像，优先高频运营动作。

### 5.2 后端服务

核心业务 API：

- 使用 Go，按领域模块组织，先做模块化单体，后续按容量拆服务。
- 领域模块：认证、用户、地址、收藏、评价、积分、会员、商户、骑手、店铺、商品、首页配置、优惠券、圈子/小微墙、饭搭、购物车、订单、售后、配送承诺、团购、买药、跑腿、钱包、提现、支付、消息、推送、客服、RTC、邀请、风控、数据管理、配置、开放平台。
- 当前首批已落地接口：微信小程序登录、商户邀约注册、商户主体登录、管理员 bootstrap 登录、骑手/站长邀约注册、骑手/站长主体登录、商户资料、商户资质上传、商户员工健康证、商户补充资料、商户订单列表、商户接单、商户出餐、商户商品列表、商户商品创建/编辑、商户商品状态调整、商户保证金查询/缴纳、店铺列表、店铺商品、店铺团购套餐、团购下单、团购券列表、商户扫码核销团购券、用户地址、购物车、购物车结算订单、订单列表、订单详情、钱包充值入账模拟、余额支付密码、余额支付、订单退款到平台余额、部分退款累计防超退、退款策略配置、售后申请、商户/管理员售后审核、售后部分退款、微信支付预下单、微信支付回调验签、骑手保证金查询/缴纳/微信免押/退押申请、骑手上线/离线、10 分钟后自动派单、骑手拒绝后顺延派单、派单确认超时自动转派、订单状态机补偿、骑手取货、骑手送达完成、固定单量完成后免责拒派决策、站长站点骑手列表、站长站点调度订单列表、站长手动派单、站长每日任务时长/固定单量配置、站长骑手绩效等级快照、派单事件查询/审计、站点区域匹配首版、骑手抢单、每日一次免责取消、管理端运营快照聚合、管理端操作审计查询；PostgreSQL-backed Store 下管理端审计查询直接读取规范化 `audit_logs` 表，Admin Web 已提供审计中心增强和完整性状态展示首版，后端已提供后台兼容 `admin`、`super_admin`、`ops_admin`、`finance_admin`、`dispatch_admin`、`support_admin`、`security_auditor` 的服务端 RBAC 策略矩阵首版、审计 payload 服务端白名单/敏感字段掩码和审计完整性证明首版。
- 当前首批安全边界：微信登录已拆分真实 `code2session` provider resolver 和平台用户绑定，生产启动必须配置 `WECHAT_MINI_APP_ID`/`WECHAT_MINI_APP_SECRET` 换取真实 `openid`；商户、骑手、站长邀约注册必须设置密码，服务端使用 bcrypt 保存哈希，后续通过 `account_id + password` 登录并签发 session-backed token，PostgreSQL snapshot 会持久化主体 password hash；管理员 bootstrap 密码登录默认关闭，没有内置默认密码，仅同时配置 `ADMIN_BOOTSTRAP_ACCOUNT_ID` 和 `ADMIN_BOOTSTRAP_PASSWORD` 且密码长度 8-72 字节时启用；登录 token 已带服务端 session id，`DATABASE_URL` 存在时会写入 PostgreSQL `auth_sessions` 并校验未撤销状态，`POST /api/auth/logout` 可撤销当前 session；生产启动默认关闭 `Authorization: Bearer role:subject_id` 开发 token，仅 `WECHAT_MINI_LOGIN_MODE=dev` 或 `ALLOW_DEV_BEARER_AUTH=true` 时启用；`security_auditor` 已作为服务端已知角色和会话主体类型，只能读取审计账本，不能执行管理端写操作；`ops_admin`、`finance_admin`、`dispatch_admin`、`support_admin` 等后台角色已通过服务端 scope 守护邀约、退款、售后、对象清理、outbox、派单和运营快照入口；审计 payload 已在 Store 与 PostgreSQL 路径统一白名单过滤和敏感字段掩码；审计日志本地默认 `sha256:v1` 完整性证明，生产配置 `AUDIT_LOG_SIGNING_SECRET` 后使用 `hmac-sha256:v1` 检测规范化审计字段和白名单 payload 篡改。后续继续升级 JWT/OAuth、设备风控、多端 session 管理、异常登录治理、KMS/链式签名、角色/权限配置 UI、字段级/租户级权限和权限变更审计。
- 当前店铺履约门槛：用户加购、购物车结算、商户接单统一校验店铺 `active`、商户资质有效和商户保证金已缴。
- API 响应统一 envelope，错误码统一，接口 schema 放入 `packages/contracts`。

BFF：

- 面向小程序、Web 管理端、uni-app 的聚合层。
- 做接口编排、兼容旧视觉数据结构、灰度开关、运行时配置。
- 避免把端侧直接绑死到内部领域 API。
- 当前 BFF 已代理微信登录、商户邀约注册、商户主体登录、管理员 bootstrap 登录、骑手/站长邀约注册与主体登录、商户资质/员工/补充资料/订单/商品、店铺、商品、团购下单/券/核销、地址、购物车、结算、订单查询、钱包、微信支付预下单、骑手在线、自动派单、拒单顺延、派单确认超时转派、订单状态机补偿、派单事件查询、站长手动派单、站长任务配置、骑手绩效快照、管理端运营快照和管理端操作审计相关公开接口；已补浏览器来源 CORS 白名单与 `OPTIONS` 预检处理，支撑管理端 Web 从本地静态预览页访问 BFF；后续继续扩展字段适配、缓存、限流、观测和端侧灰度。

实时网关：

- 支持 IM、订单通知、骑手位置、抢单/派单事件、RTC 信令。
- 首发可沿用 Socket.IO + Redis adapter 思路；高并发阶段必须 websocket-only 或确认 sticky session。
- 消息持久化不放在网关内，网关只做鉴权、路由、广播、重试和降级。

调度服务：

- 管理抢单池、派单、骑手在线状态、位置、负载、取消、超时。
- 订单可配置抢单优先、派单优先、混合策略。
- 派单权重：距离、骑手等级、接单数、超时率、取消率、商家出餐时间、配送时效。

支付服务：

- 微信支付下单、回调、退款、对账。
- 余额支付使用钱包可用余额，必须有幂等、冻结、扣减、流水、退款反向流水。
- 钱包充值只生成外部充值 URL，端侧点击后打开浏览器。

第三方集成服务：

- OAuth provider registry：微信、后续其他平台。
- API provider registry：地图、短信、天气、药品/商家数据、第三方订单导入。
- 每个 provider 必须有授权、限流、回调验签、错误隔离、审计日志。

## 6. 核心业务设计

### 6.1 商家、店铺与业务能力

2.0 的商家模型按“账号主体”和“店铺展示/经营能力”分开：

- `merchant_account` 是商家主体账号，负责登录、权限、结算主体、合同、资质、钱包。
- 商家账号只能由管理员生成邀请链接注册，不提供公开自助注册入口；邀请必须有管理员创建者、有效期和状态。
- `shop` 是面向用户展示的店铺，一个商家账号可以拥有一个或多个店铺。
- 店铺展示页高度仿照美团外卖的信息架构：头图、logo、评分、月售、配送信息、公告、优惠、商品分类、评价、团购套餐、商家资质、联系商家/客服。
- 外卖和团购不是两套商家账号，而是店铺能力：同一个店铺可以同时开通 `takeout` 和 `groupbuy`。
- 医务室、药房不是普通外卖店铺的随意开关，而是 `account_type/category` 加资质约束；只有 `pharmacy` 或 `clinic` 类账号/店铺才允许开通 `medicine` 能力。
- 快递/跑腿不属于普通商家店铺能力，属于平台服务入口，接入骑手端，履约流程与外卖配送共用骑手状态、派单、位置、收入和客服链路。
- 商户必须上传营业执照和健康证，并填写各自失效日期；资质过期或未审核通过时，店铺状态变为 `qualification_expired`，用户端暂时关闭下单入口，商户端弹窗提示补资料。
- 商户端必须支持维护员工信息，包括姓名、手机号、角色、状态、健康证文件和健康证失效日期。
- 商户端支持补充资料上传，例如门头照、后厨照、许可证、法人/负责人材料等；补充资料可按类型配置是否需要失效日期。

店铺能力：

- `takeout`：外卖商品，骑手配送。
- `groupbuy`：团购套餐，用户到店扫码验券。
- `medicine`：买药/医务室，需药房或医务室资质，骑手配送或按监管要求调整。

履约模式：

- `rider_delivery`：外卖、买药走骑手配送。
- `platform_errand`：快递/跑腿走平台骑手履约，流程与外卖配送一致。
- `in_store_redemption`：团购到店扫码验券，不进入骑手配送。

### 6.2 骑手、站长与保证金

骑手端账号：

- `station_manager`：站长账号，通过管理员或上级站点邀约注册，可查看站点全部订单、骑手、收入、位置、任务时长和异常数据。
- `rider`：骑手账号，通过站长或管理员邀约注册，只能查看自己的订单、收入、保证金、任务时长和客服会话。

站长权限：

- 手动派单。
- 查看站点所有在线骑手和全部订单数据。
- 配置每日任务时长，例如每日目标在线/跑单时长。
- 创建或管理骑手邀约链接。

保证金：

- 骑手保证金固定 `5000` 分，即 50 元；状态为 `paid` 才可接单。
- 骑手可申请微信免押，状态为 `wechat_exempt_approved` 后可接单。
- 商户保证金固定 `5000` 分，即 50 元；商户只支持缴纳保证金，不支持免押。
- 商户未缴纳保证金时不能接单。
- 骑手退保证金从“最后一单完成时间”和“提交离职时间”中较晚者起算，一周后可退；如存在纠纷，则从纠纷截止/关闭时间起顺延一周。
- 后台需要展示骑手接单时间、平均接单耗时、日均接单数、完成率、站点/团队平均水平和等级评估结果。
- 骑手等级按站点团队整体水平相对评估，等级越高派单优先级越高。
- 后台配置每日固定订单数，骑手完成固定订单数后，再收到派单可免责不接。

### 6.3 订单类型

- `takeout`：外卖订单。
- `groupbuy`：团购订单。
- `medicine`：买药订单。
- `courier`：快递/跑腿订单。
- `errand_buy`：帮买。
- `errand_deliver`：帮送。
- `errand_pickup`：帮取。
- `errand_do`：帮办。

订单主状态：

- `created`：已创建。
- `pending_payment`：待支付。
- `paid`：已支付。
- `merchant_pending`：待商家接单。
- `preparing`：商家备货/出餐。
- `dispatching`：待分配骑手。
- `rider_assigned`：已分配骑手。
- `rider_arrived_store`：骑手到店。
- `picked_up`：已取货。
- `delivering`：配送中。
- `completed`：已完成。
- `cancelled`：已取消。
- `refund_pending`：退款中。
- `refunded`：已退款。

团购券状态：

- `paid_unused`、`redeemed`、`expired`、`refund_pending`、`refunded`。
- 团购履约方式固定为到店扫码验券，首期核销方式为 `qr_scan`。

下单附加信息：

- 用户地址必须包含联系人、手机号、城市、详细地址和经纬度；地址不可用时不能下单。
- 外卖/买药/跑腿订单必须保存地址快照，后续改地址不影响历史订单。
- 订单必须支持备注、餐具数量、无接触配送、发票意向。
- 订单金额必须拆分商品金额、配送费、包装费、优惠金额、实付金额。
- 商户、骑手、系统、管理员取消订单都必须写取消原因和操作人。

售后与评价：

- 售后类型：仅退款、部分退款、投诉、食品安全。
- 售后状态：待商户处理、商户同意、商户拒绝、平台审核、已退款、已关闭。
- 售后申请必须关联订单、用户、原因、申请金额、证据附件和处理日志。
- 评价对象：订单、店铺、骑手、商品；评价支持评分、文字、图片、隐藏、删除和申诉。
- 收藏对象：店铺、商品、团购套餐。

积分会员与公益：

- 用户消费可发放积分和成长值，退款订单必须扣回对应积分。
- 会员等级至少支持无、银卡、金卡、黑金，权益和门槛后台配置。
- 积分商城和公益入口延续旧版能力，首期可后台关闭，但不能从路线图删除。

通知、推送与风控：

- 通知渠道：站内、微信订阅消息、短信、App Push。
- 推送必须有队列、送达状态、ACK、失败重试和统计。
- 风控事件至少覆盖异常下单、恶意退款、虚假交易、联系滥用。
- 异常用户可限制下单、提现、发红包、发帖、找饭搭。

### 6.4 抢单加派单

抢单：

- 已支付并可配送订单进入抢单池。
- 外卖、买药、快递/跑腿进入骑手履约；团购不进入抢单池，用户到店扫码验券。
- 下单后 10 分钟内进入抢单大厅。
- 骑手在线、资质有效、未超接单上限才可看到。
- 骑手必须满足保证金规则或微信免押规则才可看到和接受订单。
- 抢单必须使用 Redis 原子锁或数据库唯一约束保证一个订单只被一个骑手拿到。
- 抢单成功后立刻写订单事件流，并通过实时网关通知用户、商户、骑手。

派单：

- 订单下单后超过 10 分钟仍未被抢，自动进入派单。
- 派单候选骑手按距离、负载、等级、历史履约、取消率、配送方向评分。
- 派单给骑手后有确认倒计时，超时自动转派。
- 骑手拒绝派单后，系统立即派给下一位符合条件的在线骑手。
- 站长可以从后台或骑手端站长账号手动派单。
- 在线骑手每日一次免责取消派单，按自然日或运营日由后台配置。
- 免责取消不影响等级和罚金，但必须留审计。
- 超过免责次数后按后台规则：扣分、限制派单、罚款、暂停接单，具体策略可配置。

### 6.5 骑手计价

后台配置项：

- 基础配送费。
- 距离阶梯费。
- 重量/体积加价。
- 夜间/恶劣天气/高峰补贴。
- 跨区补贴。
- 等待时间补贴。
- 团购核销配送规则。
- 买药特殊配送规则。
- 快递/跑腿按里程、重量、时间或任务类型计价。
- 平台抽佣、骑手结算周期、最低保底。

计价要求：

- 计价版本必须随订单保存，避免后台改价影响历史订单。
- 每笔骑手收入都要有可解释明细。
- 骑手端显示预估收入，结算时显示最终收入和差异原因。

### 6.6 支付和钱包

支付方式：

- `wechat_pay`：微信支付。
- `balance`：余额支付。
- `mixed`：余额加微信补差，放在第二阶段。
- 用户余额支付必须先设置余额支付密码；未设置或被锁定时不能余额支付。
- 退款策略后台可配置为默认原路返回或默认退回平台余额；平台默认策略为退回余额。
- 团购商品/套餐不可售、下架或库存为 0 时，未核销券自动退款到平台余额，除非后台把默认策略改为原路返回。

钱包账户：

- 用户钱包、商户钱包、骑手钱包分账本。
- 字段至少包括：可用余额、冻结余额、累计收入、累计支出、风控状态、版本号。
- 所有余额变化必须通过流水驱动，不允许直接改余额。
- 关键操作必须幂等：支付、退款、充值回调、提现、结算。

网页充值：

- 小程序/uni-app 只请求后端拿 `walletRechargeUrl`。
- 端侧点击后打开外部浏览器或 webview，后续由独立网站完成 OAuth/二维码扫码登录和充值。
- 充值完成后由充值网站回调平台 API，平台更新钱包流水。

### 6.7 实时消息和语音

IM：

- 会话类型：用户-商户、用户-骑手、商户-骑手、用户-官方客服、商户-官方客服、骑手-官方客服。
- 消息类型：文本、图片、语音消息、订单卡片、位置、系统通知。
- 必须有消息落库、已读、撤回、敏感词、客服转接、附件上传、离线推送。
- 消息页支持群聊。
- 官方群聊在用户注册后自动加入，默认消息不通知。
- 商户可创建群聊，可配置优惠券必须进群后才能领取和使用。
- 群聊和私聊都支持余额红包；发送方可以设置总金额、数量和红包类型。
- 红包类型支持普通定额红包和拼手气红包，资金从平台余额扣除。

圈子/小微墙：

- 首页增加 `circle` 圈子入口，后台可控制是否展示、排序、灰度和入口文案。
- 参考 `HCRXchenghong/InfiniLink` 的原生小程序圈子、发帖、信息流和消息入口，但不整仓嵌入，不复用其旧支付、会员和 Go 后端。
- 首期做轻量“小微墙”：文本、图片、找饭搭动态、标签、点赞、评论、举报、后台审核。
- 圈子内容必须支持待审核、已发布、隐藏、删除、置顶、敏感词和运营推荐。
- 首页推荐卡片里的商品、店铺、团购券、优惠券、圈子动态都由后台配置，支持开关、排序、投放时间和灰度。

找饭搭：

- 找饭搭入口归属圈子能力，后台可单独开关。
- 用户使用前必须完善个人资料里的性别。
- 用户必须签署身份真实性承诺。
- 用户必须签署“自己的所有行为与平台无关”的免责承诺。
- 用户通过答题生成性格特征和饮食习惯标签，再基于标签做饭搭匹配。
- 匹配结果只作为平台内社交推荐，必须提供举报、拉黑、安全提示和风控审计。

优惠券：

- 商户可自行发券，优惠成本由商户承担。
- 平台可发券，类型一是平台承担成本，平台账户产生补贴流水并在结算时补贴给商户。
- 平台可发券，类型二是商户弹窗点击同意参与活动后，优惠成本由商户承担。
- 优惠券适用范围支持仅限某一个商家，或参与活动的全部商家。
- 优惠券资金责任必须写入券配置、订单优惠明细、商户结算和审计日志。

语音通话：

- 首发：WebRTC 点对点语音，Socket 网关做信令。
- 必须支持呼叫、响铃、接听、拒绝、取消、挂断、超时、异常结束。
- 通话记录写入 RTC 审计，关联订单/会话。
- 后续：接入 SFU、录音、质检、客服坐席、弱网降级。

### 6.8 第三方 OAuth 和 API

统一接入模型：

- `provider`：微信、高德、短信平台、第三方商家平台等。
- `credential`：appid、secret、证书、回调密钥。
- `scope`：授权范围。
- `binding`：平台用户/商户/骑手与外部身份绑定。
- `sync_job`：外部数据拉取/推送任务。
- `audit_log`：所有授权、回调、同步、失败都留审计。

安全要求：

- OAuth state 防 CSRF。
- 回调验签。
- Token 加密存储。
- provider 级限流和熔断。
- 第三方故障不拖垮主链路。

## 7. 10 万同时在线与容灾方案

容量目标不是代码里写个数字，而是压测验收结果。验收前只能说“按 10 万在线目标设计”，不能说“已经支撑 10 万”。

基础拓扑：

- API 多副本，Kubernetes HPA。
- BFF 多副本，独立扩容。
- 实时网关多副本，Redis adapter 或一致性路由。
- Redis Cluster 或 Redis Sentinel，高可用部署。
- PostgreSQL 主从加 PgBouncer，读写分离，慢查询治理。
- 消息队列：Kafka、NATS 或 RabbitMQ，用于订单事件、通知、支付异步、对账。
- 对象存储：图片、聊天附件、证件、日志归档。
- CDN：静态资源、图片缩略图。
- 可观测：Prometheus、Grafana、Loki、Tempo、OpenTelemetry。

容量验收阶段：

- 10k：基础链路和单域稳定性。
- 30k：多节点扩容、广播风暴、重连风暴。
- 60k：数据库连接池、Redis 热点、队列积压、支付回调隔离。
- 100k：混合业务压测和故障演练。

每一档必须验证：

- 登录鉴权成功率。
- IM 发送/接收 p95/p99。
- 订单通知 p95/p99。
- 骑手位置上报吞吐。
- 抢单/派单冲突率。
- RTC 信令延迟。
- 节点故障恢复时间。
- Redis 故障切换。
- 数据库连接池和慢查询。

目标阈值：

- Socket 鉴权成功率 `>= 99.95%`。
- 核心实时链路 `p95 < 150ms`。
- 写接口 `p95 < 250ms`。
- RTC 信令 `p95 < 150ms`。
- RTC 邀请送达 `p95 < 2s`。
- 总错误率 `< 0.1%`。
- 单节点故障恢复 `p95 < 15s`。

容灾目标：

- 同城多可用区部署。
- 关键数据 RPO `<= 5min`，核心订单/支付链路尽量接近 0 数据丢失。
- RTO 第一阶段 `<= 30min`，成熟期 `<= 10min`。
- 支付回调、订单事件、钱包流水必须可重放。
- 每次发布前至少跑一次核心回滚演练。

## 8. 里程碑计划

### M0 规划与治理

状态：已完成。

- 完成旧版审计和品牌资产迁移。
- 建立总计划、执行台账、目录规范。
- 明确 2.0 不直接续写旧版混乱结构。
- 输出接口规范、领域枚举、设计 tokens。

验收：

- 文档和台账存在。
- 品牌资产统一入口存在。
- 各端目标边界明确。

### M1 基础工程与账号体系

状态：进行中。

- 建立 monorepo。
- 建立 Go API、BFF、实时网关基础服务。
- 建立数据库迁移、Redis、本地 Docker。
- 建立统一鉴权：用户、商户、骑手、管理员。
- 微信小程序登录。
- OAuth provider registry 初版。
- 管理端登录和权限。

验收：

- 本地一键启动基础依赖。
- 用户/商户/骑手/管理员可登录。
- API envelope、错误码、鉴权中间件可测。

### M2 用户端原生微信小程序 MVP

状态：进行中。

当前已落地：

- 首页、圈子、小微墙、找饭搭入口。
- 商家列表、店铺详情、购物车、确认订单、订单列表、订单详情、余额支付密码、地址列表预览页。
- 小程序 API 客户端和微信登录、店铺/商品/购物车/结算/订单查询/支付密码/微信支付预下单接入点。
- 小程序默认访问 BFF，本地默认地址 `http://127.0.0.1:25500`。
- 小程序端内复制旧版 logo 到 `apps/user-wechat-miniprogram/assets/brand/`，主题色继续使用 `#009bf5`。
- Go API 已提供店铺、商品、地址、购物车、结算订单接口，并用测试覆盖“选店-加购-地址-结算-余额支付-骑手抢单”闭环。

- 复刻旧版浅蓝 UI。
- 首页、圈子入口、小微墙、找饭搭入口、分类、商家、商品、购物车、确认订单。
- 首页推荐商品/店铺/团购/券/圈子卡片从后台配置读取。
- 搜索、今日推荐、商品详情、购物车弹窗、位置选择、收货地址、订单备注、餐具数量、支付结果。
- 微信支付、余额支付入口。
- 订单列表、订单详情、客服、消息、邀请页。
- 评价订单、申请售后、我的收藏、我的评价、红包优惠、钱包充值/提现/账单。
- 积分商城、会员中心、设置、修改手机号、反馈合作、公益入口。
- 买药入口和极速买药页面。
- 跑腿/快递入口和下单页。
- 商户未缴纳保证金或资质过期时，用户端店铺页展示暂停营业/不可下单状态。
- 用户可设置/修改余额支付密码。
- 消息页支持官方群、商户群、群红包、私聊红包。
- 支持进群后领取/使用指定优惠券。
- 找饭搭使用前要求完善性别、签署身份真实性承诺和平台免责承诺、完成性格与饮食问卷。

验收：

- 真机可跑通登录、浏览、下单、支付模拟、查看订单。
- UI 主题色、logo、布局与旧版风格一致。

### M3 商户端 uni-app MVP

状态：进行中。

当前已落地：

- 管理员创建商户邀请接口。
- 商户通过邀请注册接口，注册时必须设置密码。
- 商户账号密码登录接口，登录后签发 session-backed merchant token。
- 管理员 bootstrap 登录接口，默认关闭，仅通过环境变量启用且无默认密码。
- 商户资料 token 签发。
- 营业执照和健康证上传接口。
- 资质缺失检查，未补齐资质或未缴保证金不可接单。
- 商户资料接口支持种子店铺账号。
- 商户订单列表接口。
- 商户接单接口，支付成功后的外卖/买药店铺订单先进入 `merchant_pending`。
- 商户出餐接口，出餐后订单进入骑手调度 `dispatching`。
- 商户商品列表接口。
- 商户商品创建/编辑接口。
- 商户商品上架、售罄、下架状态接口。
- 商户端经营概况页面首版。
- 商户端订单处理页面首版。
- 商户端商品管理页面首版。
- 团购套餐公开列表接口。
- 团购下单接口。
- 支付成功后自动发放团购券。
- 用户团购券列表接口。
- 商户扫码核销团购券接口。
- 商户端团购核销页面首版。

- 登录。
- 管理员邀请链接注册，禁止自助注册。
- 营业执照、健康证上传和失效日期维护。
- 资质过期/未审核通过时弹窗提示补资料，并显示店铺暂时关闭原因。
- 员工信息维护和补充资料上传。
- 经营概况。
- 订单接单/拒单/出餐。
- 商品管理。
- 优惠券和活动承担方式展示。
- 菜品/商品详情管理：图片、描述、配料表、价格、库存、上下架。
- 店铺设置。
- 店铺详情页自主管理：头图、logo、公告、优惠、评价展示、资质展示、外卖商品和团购套餐。
- 团购商户端提供二维码扫描器，扫描用户团购券二维码后核销。
- 钱包。
- 商户可自行发券，商户券优惠成本由商户承担。
- 商户可在平台活动弹窗中同意参与，参与后活动券优惠成本由商户承担。
- 商户可发红包给用户、骑手或商户群。
- 与用户/骑手/客服沟通。

验收：

- 商户可以完整处理一笔外卖订单。
- 商户端 UI 沿用旧版 merchant-app 风格。
- 商户缴纳 50 元保证金后才可接单。
- 商户可完整维护一个美团外卖式店铺详情页和菜品详情。
- 团购券扫码验券可用。

### M4 骑手端 uni-app MVP

状态：进行中。

当前已落地：

- 骑手端 uni-app 已有抢单大厅首屏，可执行上线/离线、每日一次免责取消和拒绝当前派单。
- 站长工作台首屏已接入站点骑手列表、待调度订单列表、手动派单、每日任务时长、固定单量配置和骑手等级/接单耗时展示。
- Go API 已支持骑手/站长邀约注册、骑手/站长账号密码登录、骑手在线状态、10 分钟后自动派单、拒单顺延、站长站点骑手/订单视图、站长手动派单、站长任务配置和站点骑手绩效等级快照。
- BFF 已代理骑手/站长邀约注册、主体登录以及骑手和站长首批调度接口。

- 站长/骑手邀约注册，禁止公开自助注册。
- 站长账号与骑手账号分权登录。
- 骑手登录、在线/离线。
- 抢单大厅。
- 10 分钟抢单大厅和 10 分钟后自动派单。
- 拒绝派单后自动顺延下一位在线骑手。
- 站长查看全站点数据、配置每日任务时长、手动派单。
- 骑手接单时间、平均接单耗时、日均完成单数、完成率和等级展示。
- 派单候选按后台评估等级优先级、接单速度、距离和负载排序。
- 派单弹窗和确认倒计时。
- 每日一次免责取消。
- 完成后台配置的每日固定订单数后，再收到派单可免责不接。
- 骑手 50 元保证金或微信免押通过后才可接单。
- 骑手离职退押金规则展示和申请。
- 我的任务、任务详情、配送状态流转。
- 钱包、收入、历史订单、接单设置。
- 群聊/私聊红包收发记录。
- 与用户/商户/客服沟通。

验收：

- 骑手能抢单、接派单、取消、取货、送达。
- 免责取消计数和后台审计可查。
- 未缴纳保证金且未通过微信免押的骑手不能接单。
- 站长能手动派单并查看站点全部数据。

### M5 管理端 Web 与移动管理端

状态：进行中。桌面 Web 已完成最小运营控制台首版、运营快照聚合 API 入口和 P0 快照绑定首版，移动管理端仍是骨架。

- 桌面 Web 后台：已具备登录操作台、邀请准入、运营快照、退款策略、售后列表、outbox/对象清理运维、订单状态补偿和审计检索入口；订单、售后、商户、骑手、骑手绩效、派单、审计检索、退款策略 P0 视图已可按运营快照或审计接口生成表格和指标，后续补详情与审核表单。
- 移动管理端 uni-app 高频操作。
- 订单、商户、骑手、用户、商品、团购、买药、跑腿配置。
- 首页卡片管理：控制首页展示商品、店铺、团购、优惠券和圈子动态。
- 精选商品、首页活动、首页入口、公益入口、会员权益和积分商城配置。
- 圈子/小微墙管理：开关、审核、隐藏、删除、置顶、举报和敏感词。
- 找饭搭管理：问卷题库、性格/饮食标签、开关、安全提示和协议版本。
- 优惠券管理：商户自发券、平台补贴券、商户同意参与券、单店/活动商家范围。
- 支付中心、钱包、退款、对账。
- 售后与投诉、评价管理、收藏数据、异常订单风控。
- 骑手计价规则。
- 派单策略。
- OAuth/API 管理。
- 客服工作台、RTC 审计。
- 商户邀请链接创建、撤销、过期管理。
- 商户资质审核、资质过期监控、强制临时关店和恢复。
- 站长和骑手邀请链接创建、撤销、过期管理。
- 站点每日任务时长配置、站长手动派单审计。
- 骑手/商户保证金缴纳、微信免押审核、退押金和纠纷延期管理。
- 骑手接单时间、平均接单耗时、日均接单数、团队均值、等级和派单优先级管理。
- 每日固定订单数配置，完成固定数后的免责不接规则。
- 退款默认策略配置：默认退余额或默认原路返回。
- 群聊、进群领券、红包风控和红包流水管理。
- 平台补贴券需要补贴账户、结算流水和对账报表；商户承担券需要商户确认记录和活动审计。
- 数据导出、完整备份包、恢复演练和系统日志必须保留旧版能力。

验收：

- 运营人员可以配置平台并处理首发业务闭环。

### M6 支付、钱包和结算

状态：进行中。

- 微信支付下单、回调、退款、对账。
- 余额支付、钱包流水、冻结/解冻。已完成购物车结算、余额支付和退款 PostgreSQL 事务化首版。
- 余额支付密码设置、校验、锁定和重置。
- 退款默认策略：退回平台余额或原路返回。已完成后台策略配置、管理员订单退款、余额退款入账、原路返回 pending outbox、PostgreSQL 事务化首版和部分退款累计防超退。
- 售后申请和审核：已完成用户申请、商户/管理员列表、商户/管理员审核、审核通过后触发余额/原路退款；`order_after_sales` PostgreSQL 规范表、恢复首版、审核 SQL 事务化首版、部分退款资金账本、`order_after_sales_events` 仲裁/客服介入处理日志首版、可退金额展示、证据上传票据、证据确认、`order_after_sales_evidence_upload_tickets` 签发票据账本、`order_after_sales_evidence` 附件元数据、对象存储上传签名配置化首版、对象 HEAD 存在性校验开关、上传回调验签、扫描门禁首版、对象扫描 worker 首版、ClamAV 下载扫描适配首版、对象生命周期清理 worker 首版、对象清理失败账本首版和对象清理统计接口首版已完成，确认附件必须匹配已签发未过期票据，生产开启门禁后需上传回调且扫描通过；真实 MinIO SDK/STS/Vault 临时凭证、隔离 bucket、扫描/删除失败告警投递和生产对象存储权限策略待推进。
- 团购不可售、下架或库存为 0 后，未核销券自动退款。
- 群聊/私聊红包扣款、领取、退回和风控流水。
- 用户充值跳转外部浏览器网站。
- 商户结算、骑手收入结算。
- 风控限额和异常审计。

验收：

- 支付、退款、余额扣减、回调重放、对账全部有测试；购物车结算、余额支付扣减和订单退款已由 `PostgresStore` 使用事务、行锁和幂等约束守护。

### M7 实时消息、通知和语音

状态：未开始。

- IM 会话和消息落库。
- 官方群自动加入且默认不通知。
- 商户群、进群领券和群成员关系。
- 群聊红包、私聊红包和拼手气红包事件。
- 实时网关鉴权、房间、事件。
- 订单通知、离线推送。
- WebRTC 语音信令。
- RTC 审计。
- 客服工作台。

验收：

- 用户、商户、骑手、官方客服之间能实时沟通。
- 语音通话能完成呼叫到挂断完整流程。

### M8 容量、容灾和发布体系

状态：未开始。

- 建立压测环境。
- 10k/30k/60k/100k 阶段压测。
- Redis、数据库、网关、API 故障演练。
- 发布、回滚、灰度、监控告警。

验收：

- 100k 报告归档并通过验收门禁。
- 容灾演练有证据。

### M9 原生 App 迁移准备

状态：未开始。

- 用户端原生 App 壳和模块边界。
- 商户端原生 App 计划。
- 骑手端原生 App 计划。
- 管理端移动原生化计划。
- 将业务 SDK、接口、设计 tokens 与平台能力沉淀为可迁移资产。

验收：

- 每端都有明确的原生迁移路线和复用清单。

## 9. 首批待建接口清单

认证：

- `POST /api/auth/wechat-mini/login`
- `POST /api/admin/merchant-invites`
- `GET /api/onboarding/merchant-invites/:token`
- `POST /api/onboarding/merchant-invites/:token/register`
- `POST /api/admin/rider-invites`
- `POST /api/station-manager/rider-invites`
- `POST /api/auth/rider/invite-register`
- `POST /api/auth/merchant/login`
- `POST /api/auth/rider/login`
- `POST /api/auth/admin/login`
- `POST /api/oauth/:provider/callback`

订单：

- `POST /api/orders`
- `GET /api/orders`
- `GET /api/orders/:id`
- `POST /api/orders/:id/pay`
- `POST /api/orders/:id/cancel`
- `POST /api/orders/:id/refund`
- `POST /api/orders/:id/status`
- `GET /api/admin/rider-performance`
- `GET /api/admin/rider-performance/:riderId`
- `PUT /api/admin/rider-level-rules`

调度：

- `GET /api/rider/order-pool`
- `POST /api/rider/online`
- `POST /api/rider/orders/:id/grab`
- `POST /api/dispatch/orders/:id/auto-assign`
- `POST /api/rider/orders/:id/reject-assignment`
- `POST /api/rider/dispatch/:assignmentId/accept`
- `POST /api/rider/dispatch/:assignmentId/cancel`
- `POST /api/rider/dispatch/cancel-free`
- `POST /api/rider/location`
- `POST /api/station-manager/dispatch/:orderId/manual-assign`
- `GET /api/station-manager/orders`
- `GET /api/station-manager/riders`
- `GET /api/station-manager/task-duration`
- `PUT /api/station-manager/task-duration`
- `GET /api/station-manager/rider-performance`
- `GET /api/admin/dispatch/rules`
- `PUT /api/admin/dispatch/rules`

计价：

- `GET /api/admin/rider-pricing-rules`
- `POST /api/admin/rider-pricing-rules`
- `POST /api/pricing/rider/quote`
- `GET /api/admin/daily-order-quota`
- `PUT /api/admin/daily-order-quota`

支付和钱包：

- `POST /api/payments/wechat/prepay`
- `POST /api/payments/wechat/callback`
- `POST /api/payments/refund`
- `POST /api/wallet/pay`
- `POST /api/wallet/payment-password`
- `PUT /api/wallet/payment-password`
- `POST /api/deposits/rider/pay`
- `POST /api/deposits/rider/wechat-exempt/apply`
- `POST /api/deposits/rider/refund-request`
- `POST /api/deposits/merchant/pay`
- `GET /api/admin/deposits/review-queue`
- `POST /api/admin/deposits/:id/review`
- `GET /api/admin/refund-settings`
- `PUT /api/admin/refund-settings`
- `GET /api/wallet`
- `GET /api/wallet/transactions`
- `POST /api/wallet/recharge-url`

消息和 RTC：

- `GET /api/messages/conversations`
- `GET /api/messages/conversations/:id`
- `POST /api/messages`
- `POST /api/messages/:id/read`
- `POST /api/rtc/calls`
- `GET /api/rtc/calls/:id`
- `POST /api/rtc/calls/:id/status`
- `GET /api/groups`
- `POST /api/groups`
- `POST /api/groups/:groupId/join`
- `POST /api/groups/:groupId/red-packets`
- `POST /api/messages/red-packets`

邀请：

- `POST /api/invites`
- `GET /api/invites/:token`
- `POST /api/invites/:token/accept`
- `GET /api/users/me/invite-stats`

开放平台：

- `GET /api/admin/integrations/providers`
- `POST /api/admin/integrations/providers`
- `POST /api/admin/integrations/providers/:id/test`
- `POST /api/integrations/:provider/sync`

商户资料：

- `GET /api/merchant/shops/:shopId/profile`
- `PUT /api/merchant/shops/:shopId/profile`
- `POST /api/merchant/shops/:shopId/qualifications`
- `GET /api/merchant/shops/:shopId/qualifications`
- `POST /api/merchant/shops/:shopId/staff`
- `GET /api/merchant/shops/:shopId/staff`
- `POST /api/merchant/shops/:shopId/supplemental-documents`
- `POST /api/merchant/shops/:shopId/products`
- `PUT /api/merchant/shops/:shopId/products/:productId`
- `POST /api/merchant/groupbuy/vouchers/scan`
- `POST /api/merchant/groups`
- `POST /api/merchant/coupons/:couponId/group-requirement`
- `GET /api/admin/merchant-qualifications/review-queue`
- `POST /api/admin/merchant-qualifications/:id/review`

## 10. 数据模型基线

核心表：

- `users`
- `user_addresses`
- `user_favorites`
- `user_reviews`
- `user_points_accounts`
- `user_points_transactions`
- `membership_tiers`
- `membership_entitlements`
- `merchants`
- `shops`
- `merchant_onboarding_invites`
- `merchant_qualifications`
- `merchant_staff`
- `merchant_supplemental_documents`
- `riders`
- `rider_onboarding_invites`
- `rider_accounts`
- `station_task_configs`
- `deposit_accounts`
- `deposit_transactions`
- `deposit_refund_requests`
- `products`
- `categories`
- `carts`
- `cart_items`
- `orders`
- `order_items`
- `order_options`
- `order_events`
- `order_after_sales`
- `order_reviews`
- `delivery_promises`
- `dispatch_assignments`
- `rider_locations`
- `rider_cancel_quotas`
- `rider_pricing_rules`
- `rider_performance_daily`
- `rider_level_snapshots`
- `daily_order_quota_rules`
- `wallet_accounts`
- `wallet_transactions`
- `wallet_payment_passwords`
- `payment_transactions`
- `refund_transactions`
- `refund_settings`
- `withdraw_requests`
- `withdraw_payouts`
- `home_cards`
- `featured_products`
- `home_campaigns`
- `coupons`
- `coupon_activity_merchants`
- `coupon_subsidy_transactions`
- `circle_feature_configs`
- `circle_posts`
- `circle_post_audits`
- `meal_match_profiles`
- `meal_match_questionnaires`
- `meal_match_agreements`
- `groupbuy_deals`
- `groupbuy_vouchers`
- `group_chats`
- `group_members`
- `group_coupon_requirements`
- `red_packets`
- `red_packet_claims`
- `medicine_items`
- `errand_requests`
- `conversations`
- `messages`
- `notifications`
- `push_deliveries`
- `rtc_calls`
- `invite_tokens`
- `oauth_providers`
- `oauth_bindings`
- `api_provider_configs`
- `risk_events`
- `risk_decisions`
- `data_export_jobs`
- `data_import_jobs`
- `backup_bundles`
- `audit_logs`

关键原则：

- 订单、支付、钱包必须有事件/流水，不只存最终状态。
- 金额使用整数分。
- 所有回调和高风险操作必须幂等。
- 配置规则必须版本化。
- 敏感字段加密或脱敏。

## 11. 验证标准

每个模块至少具备：

- 单元测试。
- API contract 测试。
- 关键链路集成测试。
- 前端真机或浏览器截图验收。
- 失败和重试用例。
- 权限和越权测试。

发布前必须具备：

- `go test ./...`
- 前端类型检查和构建。
- Socket 网关测试。
- 支付回调重放测试。
- 钱包并发扣减测试。
- 抢单并发冲突测试。
- 关键业务 E2E。
- 容灾和压测报告。

## 12. 当前完成/未完成总览

已完成：

- 已确认当前 `infinitech2.0` 目录为空目录起步。
- 已审计旧版 GitHub 仓库的 README、目录、主题色、logo、端划分和部分后端能力。
- 已迁移品牌 logo 到 `assets/brand/`。
- 已建立 2.0 平台总计划。
- 已建立旧版复用审计。
- 已建立执行台账。
- 已初始化 monorepo 工程骨架。
- 已建立 GitHub Actions、PR 模板、Issue 模板、CODEOWNERS 和 Dependabot 协作基线；push/PR 会运行 `npm run verify` 和 uncached Go 测试。
- 已建立原生微信小程序、商户端 uni-app、骑手端 uni-app、管理端 Web、移动管理端 uni-app 的首层工程。
- 已将管理端 Web 从占位页推进到最小运营控制台首版，包含登录操作台、商户/站长/骑手邀约、退款策略、售后列表、对象清理、outbox 运维、订单状态补偿、P0 指标位、今日必盯队列、模块状态和 RBAC 首版矩阵。
- 已补管理端 Web P0 业务视图首版：订单监控、售后审核、商户资质、骑手/站长、骑手绩效、派单审计、退款策略均有独立页面结构、指标、表格、操作入口和安全约束。
- 已新增管理端运营快照首版：`/api/admin/operations/snapshot` 聚合订单、商户资质/保证金、骑手/站长、骑手绩效、售后、派单审计、退款策略、outbox 健康和对象清理统计，BFF 与管理端操作台已接入。
- 已新增管理端运营快照绑定首版：有管理员 token 时可刷新快照，P0 KPI、今日队列、订单、售后、商户、骑手、骑手绩效、派单和退款策略视图可由快照数据生成，后端返回字段进入页面前会 HTML 转义。
- 已新增管理端操作审计日志首版：关键后台写操作记录 actor、action、target、request_id、ip_hash、服务端白名单 payload 和 created_at，管理员可通过 `/api/admin/audit-logs` 查询，BFF 与 Admin Web 操作台已接入；PostgreSQL-backed Store 已把审计推进到规范化 `audit_logs` 表和 actor/action/target 查询索引，旧快照审计会幂等补入表。
- 已新增管理端审计中心增强首版：`/api/admin/audit-logs` 支持 after/before 时间范围查询，Admin Web 支持 actor/action/target/after/before/limit 筛选、before 游标翻页、保存筛选、详情抽屉、按目标筛选、跨模块跳转、payload 白名单摘要和敏感字段脱敏详情；RBAC 首版矩阵已加入安全审计员、运营、财务、调度和客服角色。
- 已新增管理端审计服务端安全边界首版：`security_auditor` 作为服务端真实角色可只读 `/api/admin/audit-logs`，写操作继续要求管理员；`auth_identities`、`auth_sessions` 和运行时建表均允许该主体类型；审计 payload 在 Store、PostgreSQL 写入、读取和镜像恢复路径统一调用白名单/掩码策略，避免敏感字段原样持久化或返回。
- 已新增管理端审计完整性证明首版：`audit_logs` 表和 API 返回 `integrity_algorithm`、`integrity_hash`、`integrity_verified`，内存 Store 与 PostgreSQL 路径对规范化审计字段和服务端白名单 payload 生成稳定证明；本地默认 `sha256:v1`，生产配置 `AUDIT_LOG_SIGNING_SECRET` 后使用 `hmac-sha256:v1`，Admin Web 已展示完整性状态、算法和哈希。
- 已新增退款策略配置、管理端订单退款、售后审核、订单状态补偿、对象清理完成/失败、outbox 运维与商户/骑手邀约审计同事务首版：`PUT /api/admin/refund-settings` 通过仓储级 `SaveRefundSettingsWithAudit` 同时提交退款策略和 `admin.refund_settings.updated` 审计；`POST /api/orders/{orderID}/refund` 通过仓储级 `RefundOrderWithAudit` 同时提交退款业务账本和 `admin.order.refunded` 审计；`POST /api/after-sales/{requestID}/review` 通过仓储级 `ReviewAfterSalesWithAudit` 同时提交售后审核、必要退款和 `after_sales.reviewed` 审计；`POST /api/admin/orders/{orderID}/state/compensate` 通过仓储级 `CompensateOrderStateWithAudit` 同时提交订单状态补偿和 `admin.order_state.compensated` 审计；`POST /api/admin/object-storage/cleanup-complete` 与 `POST /api/admin/object-storage/cleanup-failed` 通过仓储级 `CompleteObjectStorageCleanupWithAudit` 与 `RecordObjectStorageCleanupFailureWithAudit` 同时提交对象清理票据状态和 `admin.object_cleanup.completed`/`admin.object_cleanup.failed` 审计；outbox claim/lease renew/publish/fail/replay/batch replay 通过 `ClaimOutboxEventsWithAudit`、`RenewOutboxEventLeaseWithAudit`、`MarkOutboxEventPublishedWithAudit`、`MarkOutboxEventFailedWithAudit`、`ReplayOutboxEventWithAudit` 和 `ReplayOutboxEventsWithAudit` 同时提交 `platform_outbox_events` 状态变化和对应审计；`POST /api/admin/merchant-invites`、`POST /api/admin/rider-invites` 与 `POST /api/station-manager/rider-invites` 通过 `CreateMerchantInviteWithAudit`、`CreateRiderInviteWithAudit` 同时提交最终邀约和对应审计；PostgreSQL-backed Store 在同一个数据库事务内写入对应业务表、outbox 表或邀约快照与 `audit_logs`，审计 payload 只保留服务端根据最终业务结果规范化后的字段。
- 已新增管理端服务端 RBAC 策略矩阵首版：后台角色包含兼容 `admin`、`super_admin`、`ops_admin`、`finance_admin`、`dispatch_admin`、`support_admin` 和 `security_auditor`；邀约、退款、运营快照、售后读取/审核/客服事件、对象清理、outbox 运维、订单状态补偿、派单读写和审计读取已按服务端 scope 守护；Admin Web 的 RBAC 配置已同步服务端 scope 命名。
- 已新增 BFF 浏览器 CORS 白名单首版：默认允许本地管理端/uni 调试来源，支持 `BFF_ALLOWED_ORIGINS` 配置部署来源，并明确放行 `Authorization`、`Content-Type` 和 `X-Client-Kind` 请求头。
- 已建立 Go 核心 API、BFF、实时网关和 Worker 骨架。
- 已补充领域模型：商家/店铺、资质、保证金、骑手站长、抢单派单、退款、群聊红包、圈子、找饭搭、优惠券、旧版闭环能力。
- 已实现用户端外卖首批闭环接口和测试：店铺、商品、地址、购物车、结算、余额支付、骑手抢单。
- 已实现用户小程序首批预览页：首页、圈子、找饭搭、商家列表、店铺详情、购物车、确认订单、地址。
- 已实现商户端 uni-app 首批经营、订单、商品、团购核销和资质资料页面。
- 已实现骑手端 uni-app 首批抢单大厅和站长工作台页面，骑手可执行取货/送达、保证金缴纳/免押/退押申请，站长台可看骑手等级、接单耗时并配置任务，站长可创建本站骑手邀约，骑手/站长 API 工具已支持接受邀约带密码和账号密码登录。
- 已实现商户订单状态机、订单状态机补偿、平台 outbox 事件首版、outbox relay worker 首版、outbox relay 可运行化与部署骨架、outbox 积压观测首版、outbox 手动恢复/重放首版、outbox 批量恢复/重放首版、outbox 死信隔离首版、outbox relay 租约领取首版、outbox relay 租约续租首版、PostgreSQL outbox 规范化 relay 路径首版、outbox 租约健康观测首版、消费端幂等落库首版、支付/钱包 PostgreSQL 规范化恢复首版、订单创建 PostgreSQL 事务化首版、购物车结算 PostgreSQL 事务化首版、余额支付 PostgreSQL 事务扣减首版、退款策略与余额退款核心闭环首版、payment-worker 原路退款事件规范化首版、退款 PostgreSQL 事务化首版、售后申请与审核核心闭环首版、售后 PostgreSQL 规范化恢复首版、售后审核 PostgreSQL 事务化首版、售后部分退款资金账本首版、售后仲裁与客服介入处理日志首版、售后可退金额与证据上传票据首版、售后上传票据账本与确认防伪首版、售后对象存在性 HEAD 校验开关首版、售后上传回调验签与扫描门禁首版、对象扫描 worker 首版、对象扫描 worker ClamAV 适配与下载首版、对象生命周期清理 worker 首版、对象清理失败账本首版、对象清理统计接口首版、派单审计事件 PostgreSQL 规范化恢复首版、管理端审计日志 PostgreSQL 规范化表首版、管理端审计服务端安全边界首版、管理端审计完整性证明首版、商家订单流转 PostgreSQL 事务化首版、商户商品管理、商户员工健康证与补充资料、商户/骑手保证金、店铺接单门槛、团购发券核销、商户邀约注册带密码、商户主体登录、管理员 bootstrap 登录、骑手/站长邀约注册、骑手/站长主体登录、骑手自动派单、拒单顺延、派单确认超时自动转派、骑手取货/送达完成、固定单量完成后免责拒派决策、站点区域匹配首版、派单事件快照持久化与查询审计、站长手动派单、站长任务配置和骑手绩效快照的接口与测试。

未完成：

- 角色/权限配置 UI、字段级/租户级权限、权限变更审计、设备风控、多端 session 管理和异常登录治理。
- 分支保护、必过检查、release/tag 策略、镜像构建和部署流水线。
- 管理端 P0 视图已完成运营快照绑定、操作审计首版、审计规范化表首版、审计中心增强首版、审计服务端安全边界首版、审计完整性证明首版、服务端 RBAC 策略矩阵首版、退款策略配置、管理端订单退款、售后审核、订单状态补偿、对象清理完成/失败、outbox 运维和商户/骑手邀约同事务审计首版；还需业务分页筛选、详情页、审核表单、角色/权限配置 UI、字段级/租户级 RBAC、剩余后台配置/运营处置/资金风控写路径同事务审计、审计导出/留存/告警、KMS/链式不可抵赖签名和冷热归档，商户端/骑手端剩余业务页面。
- 微信支付生产参数、微信原路退款 API 调用、提现/结算资金链路仍需继续推进；订单创建、购物车结算、支付/钱包、余额支付事务扣减、退款事务化、售后核心闭环、售后 PostgreSQL 恢复、审核事务化、部分退款资金账本、售后仲裁/客服介入处理日志、售后可退金额与上传票据、售后上传票据账本与确认防伪、售后对象存在性 HEAD 校验开关、售后上传回调验签与扫描门禁、对象扫描 worker、对象生命周期清理 worker、对象清理失败账本、对象清理统计接口、原路退款事件规范化、商家订单流转与派单审计事件已完成规范化/事务化或核心首版。
- 真实 Kafka/NATS broker 运维和 relay 积压恢复。
- 微信支付生产参数、余额支付风控。
- 钱包充值跳转。
- 骑手计价后台配置。
- IM、客服、RTC 语音。
- OAuth/API 接入平台。
- 10 万在线压测和容灾演练。
