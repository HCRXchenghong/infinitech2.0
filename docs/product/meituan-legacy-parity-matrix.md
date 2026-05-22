# 美团与旧版能力对标矩阵

更新日期：2026-05-21

目标：Infinitech 2.0 不能低于旧版 `HCRXchenghong/infinitech` 的闭环能力；凡旧版已出现的用户端、商户端、骑手端、后台、Go API、Socket 网关能力，2.0 必须保留，并用更清晰的领域模型、幂等、审计、风控和测试补强。

参考来源：

- 旧版 `HCRXchenghong/infinitech` 本地只读审计：`/tmp/infinitech-legacy`
- 旧版用户端页面：`user-vue/pages.json`、`app-mobile/pages.json`
- 旧版商户端页面：`merchant-app/pages.json`
- 旧版骑手端页面：`rider-app/pages.json`
- 旧版后台导航：`packages/admin-core/src/navigation-catalog.js`
- 旧版 Go 服务：`backend/go/internal/service`
- 圈子参考：`HCRXchenghong/InfiniLink`
- 美团公开 FAQ：退款到余额、提现、退款申诉、满减不含配送/包装费、超时赔付、异常账户限制、取消订单、商家未接单自动取消、催单等。
- 美团外卖商家版公开介绍：接单、催单、退款、结算对账、门店管理、营业信息、商品库存。
- 美团商户服务规范：评价回复、在线回复及时率、退款/投诉处理、配送准时、送错地址、不接单管控等。

## 1. 用户端必须保留

| 能力 | 旧版证据 | 2.0 要求 | 增强点 |
| --- | --- | --- | --- |
| 登录/注册/微信回调/忘记密码/设置密码 | `pages/auth/*` | 原生微信小程序保留微信登录、手机号、密码找回 | 登录态、OAuth state、短信验证码、设备风控统一 |
| 首页/分类/附近商家/商家详情/点餐 | `pages/index`、`pages/category`、`pages/shop` | 首页模块后台配置，外卖/团购主展示 | 首页卡片、店铺能力、资质状态、营业状态统一由 BFF 聚合 |
| 搜索 | `pages/search/index` | 搜商品、店铺、团购、药品、跑腿服务 | 后续接 OpenSearch，首期用 DB + 缓存 |
| 今日推荐/精选商品 | `pages/product/featured` | 后台可管控首页商品卡片 | 投放时间、排序、灰度、上下架联动 |
| 商品详情/规格/购物车 | `pages/product`、`pages/shop/cart-popup` | 商品图片、描述、配料表、规格、库存、购物车 | 金额整数分、库存锁定、价格快照 |
| 选择位置/收货地址/编辑地址 | `pages/location`、`pages/profile/address-*` | 用户地址簿、默认地址、经纬度、配送范围校验 | 地址不可用时阻止下单并解释原因 |
| 确认订单/备注/餐具数量 | `pages/order/confirm`、`remark`、`tableware` | 下单前支持备注、餐具、无接触配送、发票意向 | 订单选项进入订单快照和商户小票 |
| 支付结果/订单列表/订单详情/追踪 | `pages/pay/success`、`pages/order`、`medicine/tracking` | 全订单状态流转和实时推送 | 实时网关 + 事件日志 + 离线补偿 |
| 评价订单/我的评价 | `pages/order/review`、`profile/my-reviews` | 用户评价商户、商品、骑手 | 审核、隐藏、申诉、评分聚合 |
| 申请售后 | `pages/order/refund`、`AfterSales` 后台 | 退款、部分退款、投诉、食品安全售后 | 商户处理 + 平台仲裁 + 证据附件 + 审计 |
| 消息/聊天/站内语音 | `pages/message`、`pages/rtc/call` | IM、群聊、客服、RTC 语音 | 消息落库、已读、撤回、RTC 审计 |
| 收藏 | `profile/favorites` | 收藏店铺、商品、团购套餐 | 收藏状态进 BFF 聚合 |
| 红包优惠/优惠券 | `profile/coupon-list`、后台 coupon | 用户券包、可用券筛选 | 资金责任区分商户/平台 |
| 钱包/充值/提现/账单 | `profile/wallet/*` | 用户钱包、充值跳转、提现、账单 | 所有余额变更走流水，提现风控 |
| 积分商城/会员中心 | `points-mall`、`vip-center` | 积分、成长值、会员等级、权益 | 退款扣回积分，权益配置版本化 |
| 邀请好友 | `invite-friends` | 邀请用户页 | 邀请关系、奖励、风控审计 |
| 客服/反馈与合作 | `customer-service`、`cooperation` | 平台客服、反馈、合作申请 | 工单化、客服转接、处理 SLA |
| 跑腿四类 | `errand/buy/deliver/pickup/do` | 帮买、帮送、帮取、帮办 | 和骑手履约统一调度 |
| 买药/AI 问诊/极速买药 | `medicine/home/chat/order` | 买药入口、药品/医务室资质、极速买药 | 合规资质、特殊配送规则 |
| 同频饭友 | `dining-buddy/index` | 找饭搭 | 性别、真实性承诺、平台免责承诺、问卷前置 |
| 悦享公益 | `charity/index` | 公益模块后台可开关 | 首期禁用但保留能力入口 |

## 2. 商户端必须保留

| 能力 | 旧版证据 | 2.0 要求 | 增强点 |
| --- | --- | --- | --- |
| 登录/忘记密码/设置密码 | `merchant-app/pages/login` | 邀请注册后登录 | 禁止自助注册，凭管理员链接开通 |
| 经营概况 | `pages/index` | 今日订单、营业额、退款、评分、待处理 | 和后台口径一致 |
| 订单管理/订单详情 | `pages/orders` | 接单、拒单、出餐、退款沟通 | 操作留痕、异常关闭 |
| 在线沟通 | `messages/chat` | 与用户、骑手、客服沟通 | 支持红包、订单卡片、RTC |
| 商品管理/新增/编辑 | `menu/list/add/edit` | 图片、描述、配料表、规格、库存、上下架 | 价格库存快照，资质异常不可售 |
| 店铺设置/基础设置/创建店铺/切换店铺 | `store/*` | 店铺详情页自主管理 | 多店铺、资质有效期、营业状态 |
| 钱包 | `store/wallet` | 商户钱包、保证金、结算、提现 | 商户券成本与平台补贴结算分开 |
| 优惠券参与 | 旧后台 coupon + 本轮需求 | 商户自发券、平台活动弹窗同意 | 商户承担成本必须有确认记录 |

## 3. 骑手端必须保留

| 能力 | 旧版证据 | 2.0 要求 | 增强点 |
| --- | --- | --- | --- |
| 登录/忘记密码/设置密码 | `rider-app/pages/login` | 邀约制注册后登录 | 骑手/站长分权 |
| 抢单大厅/我的任务/任务详情 | `hall`、`tasks` | 10 分钟抢单大厅，超时派单 | 优先级、日固定单量免责不接 |
| 钱包/账单/充值/提现/收入 | `profile/wallet*`、`earnings` | 骑手收入、提现、保证金 | 保证金/微信免押/退押金纠纷延期 |
| 数据统计 | `data-stats` | 接单时间、平均数、完成率、等级 | 团队均值相对评估 |
| 个人信息/头像/手机号/密码 | `personal-info`、`avatar-upload` | 资料维护 | 敏感操作短信验证和审计 |
| 接单设置 | `order-settings` | 在线、接单范围、偏好 | 后台可限制 |
| 健康证 | `health-cert` | 上传、失效日期、审核 | 过期限制接单 |
| 保险保障 | `insurance` | 骑手保障信息 | 事故处理流程、凭证 |
| 违规申诉 | `appeal` | 取消、处罚、投诉可申诉 | 站长/平台审核流 |
| 骑手之家/历史订单/客服 | `rider-home`、`history`、`service` | 保留 | 消息和工单统一 |

## 4. 后台必须保留

| 能力 | 旧版证据 | 2.0 要求 | 增强点 |
| --- | --- | --- | --- |
| 仪表盘/监控 | `Dashboard`、`MonitorChat` | 订单、在线、消息、告警总览 | Prometheus/Grafana 指标接入 |
| 用户/商户/骑手/骑手等级 | `Users`、`Merchants`、`Riders`、`RiderRanks` | 分页、搜索、禁用、等级 | 站点团队均值、派单优先级 |
| 订单/售后 | `Orders`、`AfterSales` | 订单处理、退款、投诉、仲裁 | 资金幂等、证据、审计 |
| 客服/RTC/电话审计 | `SupportChat`、`RTCConsole`、`ContactPhoneAudits` | IM、RTC、电话点击审计 | 风控和隐私留痕 |
| 首页入口/活动/精选商品 | `HomeEntrySettings`、`HomeCampaigns`、`FeaturedProducts` | 首页卡片后台管控 | 商品状态联动、灰度投放 |
| 内容设置/官方通知/推送 | `ContentSettings`、`OfficialNotifications` | 轮播、推送、站内通知 | 推送队列、ACK、失败重试 |
| 优惠券 | `CouponManagement`、`CouponLanding` | 发券、领券、券链接 | 资金责任、商户确认、补贴对账 |
| 财务/支付/交易日志 | `FinanceCenter`、`PaymentCenter`、`TransactionLogs` | 钱包、支付、提现、对账 | 回调重放、提现复核 |
| 数据管理 | `DataManagement` | 导入导出、完整备份包 | 恢复演练和校验报告 |
| API/OAuth/权限/文档 | `ApiManagement`、`ApiPermissions`、`ApiDocumentation` | 第三方接入平台 | scope、签名、限流、审计 |
| 系统日志/设置 | `SystemLogs`、`Settings` | 系统配置与审计 | Secrets 不入库明文 |
| 饭搭治理 | `DiningBuddyGovernance` | 问卷、协议、安全提示、举报 | 风险用户限制 |

## 5. 2.0 新增补强

- 外卖/团购/买药/跑腿统一订单与资金模型，但履约模式分开。
- 商家账号和店铺分离，一个商家可多店铺，一个店铺可开外卖和团购。
- 商户和骑手都邀约制，资质、保证金、微信免押、退押金规则可审计。
- 所有退款默认可退余额，后台可切换原路返回。
- 官方群、商户群、进群领券、余额红包、拼手气红包进入钱包流水。
- 圈子和小微墙后台可控，InfiniLink 只作信息架构参考。
- 10 万在线只作为设计目标，必须通过 10k/30k/60k/100k 压测证明。

## 6. 必须补齐的验收证据

- 旧版页面能力在 2.0 的任务映射表。
- API contract 测试覆盖地址、购物车、下单、支付、退款、售后、评价、收藏、积分、会员、推送、钱包、提现、团购验券。
- 用户端真机截图覆盖首页、商家、购物车、下单、支付、订单、售后、评价、消息、钱包、跑腿、买药、圈子。
- 商户端截图覆盖接单、商品、店铺、资质、券、钱包、消息。
- 骑手端截图覆盖抢单、任务、接单设置、健康证、保险、申诉、钱包。
- 后台截图覆盖售后、券、首页投放、内容推送、数据备份、风控、RTC、电话审计。
