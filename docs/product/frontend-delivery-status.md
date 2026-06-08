# Infinitech 2.0 前端交付状态盘点

日期：2026-05-29
目的：确认当前“用户端效果图 -> 用户端原生微信小程序 -> 用户端 Flutter App -> 商户/骑手/管理端 Flutter -> 管理端 Web/App -> 中间件/后端/数据库联调”的真实进度。

## 1. 本次检查结论

- 仓库不是纯 UI 项目，已经包含 `apps`、`services`、`packages`、`infra` 和 `docs`，可统计源码/配置约 62442 行。
- 当前验证基线通过：`npm run verify` 成功，覆盖架构守卫、packages、Admin Web、BFF、workers 和 Go API 测试。
- 用户端 UI 效果图已经基本完成：`docs/product/user-miniprogram-reference/` 下 `00-welcome-legacy.png`、`00-login-legacy.png`、`00-register-legacy.png`、`00-auth-login-register.png`、`01-home.png` 到 `35-medicine-order-detail.png` 均存在。
- 用户端原生微信小程序已经补齐欢迎/登录/注册和 01-35 业务参考图的页面路由：`apps/user-wechat-miniprogram/app.json` 当前注册 39 个页面，每个页面都有 `index.json`、`index.wxml`、`index.ts`。
- 已有页面已优先接当前 BFF/API；本轮已把 `01` 到 `35` 按参考图完成首轮精修，并把手机号验证码生产短信 provider 首版、真实微信 `wx.login` 页面流程、评价、钱包账单/总览/提现申请、优惠券、会员积分/签到、邀请好友、搜索聚合、买药首页、跑腿下单/详情、反馈、圈子动态、找饭搭资料人工审核/候选/举报拉黑/举报处置/同校同楼隐私/设备风控、消息会话/聊天/已读回执/离线补偿/WebSocket 首版实时投递、消息 PostgreSQL 规范化首版、realtime gateway Redis adapter 多副本 fanout 首版、realtime gateway WebSocket 签名鉴权首版、realtime gateway 会话成员权限校验首版、会话免打扰偏好首版、群资料/成员预览首版、消息敏感信息风控、红包发送/详情/领取/退回/过期退回/领取风控、客服工单分派/关闭/回访、客服 SLA 状态/超时升级、客服质检/绩效、Admin Web 客服工作台首版、处方影像上传票据/上传回调/对象扫描门禁/确认/OCR 识别/药师复核/留档/审核绑定、药品订单确认/详情和药品库存锁定补成 API-Go + BFF + 小程序调用闭环。IM 实时投递已通过 realtime-gateway + outbox relay + Redis Pub/Sub + 签名 token 鉴权 + 后端会话成员校验 + 会话免打扰偏好 + 群资料/成员预览 + 小程序商户群接入；动态群成员、断线重连/压测与剩余 PostgreSQL 规范化仍是后续生产化边界，详细对接表见 `docs/product/user-miniprogram-backend-integration.md`。
- 商户端、骑手端、移动管理端当前实际目录是 `apps/merchant-uni`、`apps/rider-uni`、`apps/admin-uni`，README 已标注为历史目录，目标应迁移到 Flutter/Dart，但仓库里还没有 `apps/merchant-flutter`、`apps/rider-flutter`、`apps/admin-flutter` 或 `apps/user-flutter`。
- 管理端 Web 已经比端侧更深入，`apps/admin-web` 有运营控制台、审计、通知运营、Outbox、RBAC、商户资质、售后、退款等首版能力。
- 后端、中间件和数据库已经有不少实现，不是完全未开始，但按当前新顺序，后续应在所有前端主流程完成后再集中做真实生产联调、支付、IM/RTC、压测和容灾。

## 2. 用户端效果图完成度

参考图目录：`docs/product/user-miniprogram-reference/`

完成情况：39 张参考图，包含 1.0 复刻欢迎/登录/注册页、合并登录注册探索图和业务页 `01-home.png` 到 `35-medicine-order-detail.png`。

已存在文件：

| 编号 | 页面 | 文件 |
| --- | --- | --- |
| 00A | 1.0 欢迎页 | `00-welcome-legacy.png` |
| 00B | 1.0 登录页 | `00-login-legacy.png` |
| 00C | 1.0 注册页 | `00-register-legacy.png` |
| 00 | 登录注册 | `00-auth-login-register.png` |
| 01 | 首页 | `01-home.png` |
| 02 | 商家列表 | `02-shop-list.png` |
| 03 | 店铺详情 | `03-shop-detail.png` |
| 04 | 购物车 | `04-cart.png` |
| 05 | 确认订单 | `05-order-confirm.png` |
| 06 | 订单列表 | `06-order-list.png` |
| 07 | 订单详情 | `07-order-detail.png` |
| 08 | 通知偏好 | `08-notification-preferences.png` |
| 09 | 支付密码 | `09-payment-password.png` |
| 10 | 地址列表 | `10-address-list.png` |
| 11 | 圈子 | `11-circle.png` |
| 12 | 找饭搭 | `12-meal-match.png` |
| 13 | 申请售后 | `13-after-sales.png` |
| 14 | 评价订单 | `14-order-review.png` |
| 15 | 消息中心 | `15-messages.png` |
| 16 | 我的 | `16-profile.png` |
| 17 | 钱包 | `17-wallet.png` |
| 18 | 红包优惠 | `18-coupons.png` |
| 19 | 会员积分 | `19-member-points.png` |
| 20 | 邀请好友 | `20-invite-friends.png` |
| 21 | 搜索 | `21-search.png` |
| 22 | 买药 | `22-medicine-home.png` |
| 23 | 快递跑腿 | `23-errand-home.png` |
| 24 | 跑腿订单详情 | `24-errand-order-detail.png` |
| 25 | 商户群聊 | `25-merchant-group-chat.png` |
| 26 | 发红包 | `26-red-packet-send.png` |
| 27 | 红包详情 | `27-red-packet-detail.png` |
| 28 | 在线客服 | `28-customer-service-chat.png` |
| 29 | 工单详情 | `29-service-ticket-detail.png` |
| 30 | 投诉建议 | `30-complaint-feedback.png` |
| 31 | 反馈记录 | `31-feedback-records.png` |
| 32 | 上传处方 | `32-prescription-upload.png` |
| 33 | 处方审核结果 | `33-prescription-review-result.png` |
| 34 | 药品订单确认 | `34-medicine-order-confirm.png` |
| 35 | 药品订单详情 | `35-medicine-order-detail.png` |

补充发现：`docs/product/user-miniprogram-page-prompts.md` 已补齐 00A/00B/00C 1.0 复刻欢迎/登录/注册、00 合并登录注册探索图和 35 药品订单详情，避免后续设计留痕断档。

## 3. 用户端原生微信小程序代码完成度

目录：`apps/user-wechat-miniprogram/`

当前 `app.json` 注册页面：39 个；其中 1.0 欢迎页动画、登录、注册和 01-35 业务页均已注册，并新增商户群的群设置页。

| 效果图 | 当前代码状态 | 说明 |
| --- | --- | --- |
| 00A 欢迎页 | 已有 `pages/welcome/welcome/index` | 已按 1.0 移植蓝色渐变启动页、logo 脉冲、`悦享e食` 打字机、完整标题缩放归位、副标题和底部登录/游客卡片渐入。 |
| 00B 登录页 | 已有 `pages/auth/login/index` | 微信登录、手机号验证码登录和手机号密码登录已接 BFF/API-Go；微信登录通过真实 `wx.login` code 换取服务端 token，生产失败不再自动写开发 token；短信 provider 配置入口、验证码隐藏、发送状态和频控首版已接，本地开发验证码仍会回填。 |
| 00C 注册页 | 已有 `pages/auth/register/index` | 已按 1.0 极简注册页生成；微信注册/登录、手机号验证码注册、密码设置和协议确认已接 BFF/API-Go；开发 token 兜底仅限本地 API 或显式开关。 |
| 01 首页 | 已有 `pages/index/index` | 已按参考图补位置/搜索、8 宫格、精选团购、猜你喜欢、饭搭入口和底部 tab；接入首页模块/卡片。 |
| 02 商家列表 | 已有 `pages/shop/list/index` | 已按参考图改成“附近商家”简洁列表，去掉搜索栏和频道 tabs，保留筛选行和 5 张紧凑商家卡。 |
| 03 店铺详情 | 已有 `pages/shop/detail/index` | 已补大图头图、覆盖导航、店铺信息卡、活动标签、外卖/团购/评价/商家 tab、左分类右商品、真实评价卡片、评价图片区三宫格预览/余量提示、菜品评分摘要、配送服务评分、商家资质/联系入口和底部购物车栏；已接 `GET /api/shops/{shopID}/detail`。 |
| 04 购物车 | 已有 `pages/cart/index` | 已按参考图改为店铺详情上的底部抽屉形态，含遮罩、清空/关闭、步进器和去结算栏；已接真实购物车摘要、商品增减和清空。 |
| 05 确认订单 | 已有 `pages/order/confirm/index` | 已补地址、预计送达、商品、优惠/餐具/备注、费用层级和提交后微信/余额支付选择；已接默认地址回填、地址选择和真实购物车费用汇总。 |
| 06 订单列表 | 已有 `pages/order/list/index` | 已补状态 tab、四张订单卡、操作按钮和底部 tabbar；已接订单 `shop_name`、真实时间、`reviewed` 状态和“去评价”快捷入口。 |
| 07 订单详情 | 已有 `pages/order/detail/index` | 已补状态卡、地图式配送卡、时间线、地址、联系商家/骑手和售后底部操作；已接订单 `shop_name`、`address_snapshot`、`reviewed` 和事件时间展示。 |
| 08 通知偏好 | 已有 `pages/notification-preferences/index` | 已按参考图补三类偏好卡、渠道开关、静默时间和保存状态，并接用户通知偏好 API。 |
| 09 支付密码 | 已有 `pages/wallet/payment-password/index` | 已按参考图改为首次输入/再次确认两步流程，补盾牌图标、安全弹窗和支付安全说明；已接支付密码 API。 |
| 10 地址列表 | 已有 `pages/address/list/index` | 已按参考图补三张地址卡、默认标签、编辑/删除/设默认轻按钮和底部新增地址；已接地址列表/新增地址 API，并支持默认地址回写与确认订单页选址返回。 |
| 11 圈子 | 已有 `pages/circle/index` | 已按参考图补推荐/附近/商户群/官方 tabs、发布入口、动态流、悬浮发布和底部 tab；已接圈子动态列表/发布 API。 |
| 12 找饭搭 | 已有 `pages/meal-match/index` | 前置要求状态、资料审核中/未通过/可开启状态、候选匹配、同校/同楼隐私范围、模糊位置、设备风控、举报和拉黑已接 `GET/PUT /api/meal-match/profile`、`GET /api/meal-match/candidates`、`POST /api/meal-match/reports`、`POST /api/meal-match/blocks`；后台审核/举报处置已接 `GET /api/admin/meal-match/moderation`、`POST /api/admin/meal-match/moderation/{recordID}/review`。 |
| 13 申请售后 | 已有 `pages/after-sales/index` | 已按参考图补订单摘要、售后类型、金额、原因、凭证位、处理说明、售后记录和底部操作；已接真实订单摘要、订单级售后记录筛选、进度时间线、凭证预览和补充凭证上传首版，并兼容严格对象存储回调环境的前端确认兜底。 |
| 14 评价订单 | 已有 `pages/order/review/index` | 已按参考图补订单摘要、星级、标签、匿名开关、菜品评价和图片位；已接真实订单摘要、订单级评价查询/更新、匿名回填、逐项菜品评分、配送服务独立评分、评价图片上传票据/确认和星级交互，并兼容严格对象存储回调环境的前端确认兜底。 |
| 15 消息中心 | 已有 `pages/messages/index` | 已按参考图补搜索、分段、通知入口、会话列表、免打扰标签和底部 tab；已接会话列表 API，商户群页已接消息列表/发送、群资料/成员预览、游标同步、已读回执和会话免打扰偏好 API。 |
| 16 我的 | 已有 `pages/profile/index` | 已按参考图补蓝色头图、会员资料卡、资产区、订单快捷、服务宫格、账户安全和底部 tab；已接 `GET /api/user/profile`。 |
| 17 钱包 | 已有 `pages/wallet/index` | 已按参考图补余额蓝卡、充值/提现、资产统计、安全说明、账单筛选和操作列表；已接 `GET /api/wallet/overview`、`POST /api/wallet/withdraw`、充值/支付密码/流水接口。 |
| 18 红包优惠 | 已有 `pages/coupons/index` | 已按参考图补优惠摘要、优惠券/红包/已使用/已过期 tabs、分类筛选、券卡和规则说明，并新增商户群福利卡；已接 `GET /api/user/coupons`、`POST /api/user/coupons/claim`，支持先入商户群再领取群券。 |
| 19 会员积分 | 已有 `pages/member-points/index` | 已按参考图补会员等级、权益、赚积分任务、积分兑换、积分明细和退款提示；已接 `GET /api/user/points`、`POST /api/user/points/check-in`。 |
| 20 邀请好友 | 已有 `pages/invite-friends/index` | 已按参考图补邀请活动卡、邀请码复制、分享、生成分享图入口和邀请记录；已接 `GET /api/user/invite-summary`。 |
| 21 搜索 | 已有 `pages/search/index` | 已按参考图补搜索框、分类 tabs、猜你想搜、排序条和混合结果卡；已接 `GET /api/search?keyword=&category=`。 |
| 22 买药 | 已有 `pages/medicine/home/index` | 已按参考图补校医务室卡、用药提示、搜索、分类、AI 问诊、药品列表、处方上传入口和底部结算栏；已接 `GET /api/medicine/home`，库存会随已锁定订单扣减。 |
| 23 快递跑腿 | 已有 `pages/errand/home/index` | 已按参考图补服务类型、取送信息、物品要求、费用预估、可用优惠和确认下单栏；已接 `POST /api/errand/orders`。 |
| 24 跑腿订单详情 | 已有 `pages/errand/order-detail/index` | 已按参考图补状态卡、地图轨迹式区域、骑手卡、进度时间线、取送信息、费用明细、提示和底部操作；已接 `GET /api/errand/orders/{orderID}`。 |
| 25 商户群聊 | 已有 `pages/messages/merchant-group/index` | 已按参考图补群资料、公告、聊天内券卡、气泡消息、订单卡、红包卡、底部输入和快捷面板；新增 `pages/messages/group-settings/index` 作为群设置子页。已接消息列表/发送、群资料概览、活跃成员预览、商户群自助加入/退出、群券领取资格、离线消息游标同步、打开会话清未读、`message.sent` outbox、WebSocket 实时接收、Redis Pub/Sub 多副本网关 fanout、WebSocket 签名 token 鉴权、会话成员权限校验和会话免打扰偏好首版。 |
| 26 发红包 | 已有 `pages/red-packet/send/index` | 已按参考图补红色红包头图、拼手气/普通切换、金额/个数/祝福语、总额和塞钱按钮；已接红包创建 API，后端会冻结余额并写入钱包流水。 |
| 27 红包详情 | 已有 `pages/red-packet/detail/index` | 已按参考图补红包进度、领取记录、钱包流水和底部操作；已接红包详情、领取、退回、过期批量退回状态和领取风控提示。 |
| 28 在线客服 | 已有 `pages/customer-service/chat/index` | 已按参考图补客服状态、场景 tabs、关联订单、聊天气泡、处理建议、快捷入口和输入栏；已接客服工单列表/创建/事件追加 API，并补发送前敏感信息提示和服务端风控拦截。 |
| 29 工单详情 | 已有 `pages/service-ticket/detail/index` | 已按参考图补工单状态、关联订单、问题描述、处理进度、方案、材料和底部操作；已接工单详情、补充说明、确认关闭、回访评分、SLA 状态和超时升级展示。 |
| 30 投诉建议 | 已有 `pages/feedback/complaint/index` | 已按参考图补反馈类型、关联订单、问题说明、影响程度、联系方式、凭证和底部提交；已同时接反馈与客服工单创建 API。 |
| 31 反馈记录 | 已有 `pages/feedback/records/index` | 已按参考图补 30 天统计、状态 tabs、搜索筛选、记录卡和帮助卡；已聚合反馈列表与客服工单列表 API，并能展示待确认/已关闭状态。 |
| 32 上传处方 | 已有 `pages/prescription/upload/index` | 已按参考图补处方药商品卡、上传处方、用药人信息、校医审核和底部提交；已接处方影像上传票据、上传回调、对象扫描门禁、确认、OCR 识别摘要、处方留档和审核创建 API，并在上传区展示安全扫描状态。 |
| 33 处方审核结果 | 已有 `pages/prescription/review-result/index` | 已按参考图补审核通过状态、药品卡、处方信息、审核节点、配送方式和加入购物车；已接处方审核查询 API，展示 OCR 置信度、识别剂量、留档编号，并保留审核单的处方对象 key/hash。 |
| 34 药品订单确认 | 已有 `pages/medicine/order-confirm/index` | 已按参考图补配送地址、校医务室、药品清单、处方通过提示、费用行、备注和提交订单；已接药品订单创建 API，库存不足会提示用户返回调整药品。 |
| 35 药品订单详情 | 已有 `pages/medicine/order-detail/index` | 已按参考图补骑手取药状态、配送信息、订单进度、药品清单、费用明细和底部客服/确认收药；已接药品订单详情 API，并返回每项药品库存锁定和剩余库存信息。 |

下一步用户端小程序代码建议：

1. 做一轮微信开发者工具真机视觉走查，重点检查 25-35 的底部固定栏、长文本、rpx 间距和低端机滚动性能。
2. 按 `docs/product/user-miniprogram-backend-integration.md` 的剩余缺口继续补剩余 PostgreSQL 规范化、动态群成员、断线重连、顺序保障和压测生产化。

## 4. Flutter App 现状

当前没有真实 Flutter 工程目录和 `pubspec.yaml`：

- 未开始：`apps/user-flutter`
- 未开始：`apps/merchant-flutter`
- 未开始：`apps/rider-flutter`
- 未开始：`apps/admin-flutter`

当前存在的 `apps/merchant-uni`、`apps/rider-uni`、`apps/admin-uni` 是历史 uni-app 目录，README 已写明后续目标目录为 Flutter/Dart。它们可以作为信息架构和接口调用参考，但不能算 Flutter App 已落地。

建议顺序：

1. 用户端原生微信小程序已补齐并首轮精修 35 个业务页面，下一步做真机视觉走查和生产化缺口收口。
2. 新建 `apps/user-flutter`，按同一套信息架构做 Android/iOS 用户端 App。
3. 新建 `apps/merchant-flutter`，迁移 `merchant-uni` 已有经营、订单、商品、团购、资质、通知偏好能力。
4. 新建 `apps/rider-flutter`，迁移 `rider-uni` 抢单大厅和站长工作台，并补任务、钱包、健康证、申诉。
5. 新建 `apps/admin-flutter`，只做移动管理高频处理，不镜像桌面后台全部模块。

## 5. 管理端 Web 现状

目录：`apps/admin-web/`

当前状态：进行中，但已经是可测的运营控制台首版。

已覆盖：

- 管理员登录操作台。
- 运营快照和 P0 视图。
- 订单、售后、商户资质、骑手/站长、骑手绩效、派单、退款策略。
- 审计检索、审计导出、留存报告、归档请求、归档校验历史。
- 通知运营、通知偏好、通知偏好变更申请/审批/灰度应用。
- Outbox 运维、死信、恢复、租约、人工发布/失败。
- RBAC 策略查询、变更申请、审批、应用、回滚。

未完成重点：

- 首页卡片、优惠券、圈子饭搭、评价、客服、RTC、支付中心、钱包财务、风控、数据备份、系统设置等仍有 planned 模块。
- 生产 WORM、真实告警渠道、字段级/租户级 RBAC、KMS/链式签名仍待补。

## 6. 后端、中间件、数据库现状

当前并不是“等前端完成后才开始后端”的状态，仓库已存在大量后端和基础设施骨架：

- Go 核心 API：`services/api-go`
- BFF：`services/bff`
- 实时网关首版：`services/realtime-gateway`
- 调度、支付、通知、集成、结算、对象扫描、对象生命周期、审计归档、outbox relay workers
- PostgreSQL 迁移：`infra/db/migrations`
- Docker/K8s 骨架：`infra/docker`、`infra/k8s`
- 领域包和 contracts：`packages/*`

当前边界：

- 真实生产微信支付、微信原路退款、提现、商户结算和对账未完成。
- 真实 RTC 和地图定位未完成；消息 PostgreSQL 规范化首版、realtime gateway WebSocket 首版投递、Redis adapter 多副本 fanout、WebSocket 签名 token 鉴权、会话成员权限校验、会话免打扰偏好、用户端消息游标离线补偿与已读回执已接，动态群成员、断线重连策略、顺序保障和压测仍待生产化。
- 真实短信/企业微信/微信订阅消息/push 生产账号和模板审批未完成。
- 10 万在线压测和容灾演练未完成。

后续按用户要求，应等各端前端主流程完成后，再做一次完整的端到端联调和生产化闭环。

## 7. 下一步执行顺序

### Phase 1 用户端原生微信小程序

- 精修 01-35 页面到参考图级视觉，包括店铺详情、购物车抽屉、订单列表、钱包、消息等高频页。
- 补后端缺口并替换页面上的“待接后端 / 本地预览”状态。
- 持续更新 `app.json`、README、导航入口和 API 对接文档。
- 用微信开发者工具或可替代截图流程留证。

### Phase 2 用户端 Flutter App

- 新建 `apps/user-flutter`。
- 复用用户端 35 页信息架构。
- Android/iOS 同一套 Flutter UI，先做前端和 mock/API adapter。

### Phase 3 商户端 Flutter

- 新建 `apps/merchant-flutter`。
- 从 `apps/merchant-uni` 迁移经营概况、订单、商品、团购核销、资质、通知偏好。
- 补登录/邀请注册、店铺装修、钱包、消息、资质过期强弹窗。

### Phase 4 骑手端 Flutter

- 新建 `apps/rider-flutter`。
- 从 `apps/rider-uni` 迁移抢单大厅、站长工作台。
- 补任务详情、地图导航、收入钱包、提现、健康证、保险、违规申诉、客服。

### Phase 5 管理端 Web 和移动管理端

- 继续补 Admin Web planned 模块。
- 新建 `apps/admin-flutter`，只覆盖手机紧急运营动作。

### Phase 6 后端、中间件、数据库联调

- 串起所有前端真实 API。
- 补支付、IM/RTC、通知 provider、定位地图、生产数据库、缓存、队列、对象存储。
- 做压测、容灾、验收证据归档。

## 8. 本次验证命令

```bash
npm run verify
```

结果：通过。
