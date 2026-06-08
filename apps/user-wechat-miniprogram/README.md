# 用户端原生微信小程序

首期定位：外卖和团购主展示，首页模块化扩展买药、快递/跑腿、后续交友。

当前用户端页面：

- `pages/welcome/welcome`：按 1.0 欢迎页移植的启动动画，包含 logo 脉冲、标题打字机、缩放归位和底部登录/游客卡片渐入。
- `pages/auth/login/index`、`pages/auth/register/index`：按 1.0 极简白底登录/注册页移植，微信登录、手机号验证码、手机号密码登录和注册协议确认已接 BFF；微信登录通过真实 `wx.login` code 换取服务端 token，开发 token 兜底仅限本地 API 或显式开关；验证码会展示发送状态/重发冷却，生产短信模式不回填验证码。
- `pages/index/index`：首页模块、推荐卡片、圈子入口。
- `pages/search/index`：搜索页，当前基于后端店铺列表做本地筛选。
- `pages/shop/list/index`：商家列表，外卖/团购/买药/跑腿入口共用。
- `pages/shop/detail/index`：美团式店铺详情页，包含头图、外卖商品、团购券、真实评价和商家信息；已接 `GET /api/shops/{shopID}/detail`、团购套餐列表和下单入口，评价区会展示真实评价图片流、三宫格预览余量提示、菜品评分摘要和配送服务评分。
- `pages/cart/index`：购物车和金额拆分；已接真实购物车摘要、商品增减、清空购物车和店铺名回填。
- `pages/order/confirm/index`：确认订单、地址、备注、餐具和支付入口；已接真实地址选择、默认地址回填、购物车费用汇总和提交下单。
- `pages/order/list/index`：订单列表；已接真实 `shop_name`、订单时间、待评价状态和订单卡快捷入口。
- `pages/order/detail/index`：订单详情和订单事件流；已接订单 `shop_name`、`address_snapshot`、`reviewed` 和真实进度时间线。
- `pages/order/review/index`、`pages/after-sales/index`：评价订单与售后申请；评价页已接真实订单摘要、已评回填、匿名评价、评价图片上传、逐项菜品评分、配送服务独立评分和订单级更新；售后页已接真实订单摘要、订单级售后记录筛选、事件时间线、凭证预览和补充凭证上传首版，并兼容严格对象存储回调环境的前端确认兜底。
- `pages/wallet/payment-password/index`：余额支付密码设置。
- `pages/wallet/index`：用户钱包，余额总览、充值、账单和提现申请已接后端。
- `pages/address/list/index`：收货地址列表；支持默认地址回写和从确认订单页选择配送地址。
- `pages/coupons/index`、`pages/member-points/index`、`pages/invite-friends/index`：优惠券、会员积分、邀请好友。
- `pages/messages/index`、`pages/messages/merchant-group/index`、`pages/messages/group-settings/index`、`pages/customer-service/chat/index`、`pages/service-ticket/detail/index`：消息、商户群、群设置、客服和售后工单；消息支持会话列表、聊天发送、WebSocket 实时接收、realtime gateway Redis adapter 多副本 fanout、WebSocket 签名 token 鉴权、会话成员权限校验、会话免打扰偏好、群资料/成员预览、商户群自助加入/退出、群券资格校验、游标离线补偿、打开会话清未读、已读回执、PostgreSQL 规范化消息表/已读状态表首版和敏感信息风控，工单支持客服分派、SLA 状态、超时升级、处理方案、用户确认关闭、回访评分和客服消息敏感信息拦截。
- `pages/notification-preferences/index`：通知偏好设置，支持订单状态、售后进度和优惠活动的微信订阅、短信、App Push 开关与静默时间。
- `pages/circle/index`、`pages/meal-match/index`：圈子和找饭搭入口；找饭搭资料人工审核、候选匹配、同校/同楼隐私、模糊位置、设备风控、举报待审/处置和不感兴趣拉黑已接后端首版。
- `pages/medicine/home/index`、`pages/prescription/upload/index`、`pages/prescription/review-result/index`、`pages/medicine/order-confirm/index`、`pages/medicine/order-detail/index`：买药和处方链路；上传处方页已接处方影像上传票据、上传回调、对象扫描门禁、确认、OCR 识别摘要、药师复核、处方留档和审核单绑定首版，并展示安全扫描状态，药品订单会锁定库存并处理库存不足提示。
- `pages/errand/home/index`、`pages/errand/order-detail/index`：快递跑腿下单和履约详情。
- `pages/red-packet/send/index`、`pages/red-packet/detail/index`：余额红包发送、详情、领取、退回和过期退回已接后端首版，服务端会冻结余额并写钱包流水；红包详情页会展示领取风控校验提示。
- `pages/feedback/complaint/index`、`pages/feedback/records/index`：投诉建议、反馈记录和客服工单已接后端首版，能展示处理中、待确认、已关闭状态。
- `utils/api.ts`：小程序 API 客户端，支持微信登录、手机号登录/注册、短信验证码发送状态、店铺详情/商品/团购、购物车、结算、订单、订单级评价查询/提交、评价图片上传票据/确认、订单级售后查询、售后凭证上传票据/确认/对象存储回调补传、支付密码、用户通知偏好、消息同步/已读/发送风控、会话免打扰偏好、群资料/成员预览、自助入群/退群、商户群券领取、实时 Socket URL、钱包、红包、客服工单分派/关闭/回访、找饭搭同校隐私与设备风控、处方影像上传/扫描/确认、处方审核/OCR/留档、药品订单/库存锁定、跑腿和微信支付预下单接口接入；后端已补消息 PostgreSQL 规范化、已读状态表、会话成员权限校验、会话免打扰、群详情、商户群资格、订单 `reviewed` 状态、评价 `item_ratings`、`rider_rating` 与评价图片上传票据、售后订单上下文字段、售后证据票据确认链路，以及店铺详情评价图片区/菜品摘要聚合首版。
- `utils/media.ts`：用户端图片链接解析与兜底资产清单。页面优先消费后端/CDN 的 `image_url`、`cover_url`、`logo_url`，无后台配置时才使用 `assets/generated/` 中的生图压缩兜底。
- 品牌资源已复制到 `assets/brand/`，供小程序本地引用。
- 生图资产高清留痕见 `docs/product/generated-user-assets/`；小程序内压缩兜底见 `assets/generated/`，后续生产环境应由后台配置 CDN 链接覆盖。
- 预览图：`docs/product/user-miniprogram-preview.png`。

约束：

- 原生微信小程序，不使用旧版用户端跨端工程作为主线。
- 技术栈继续使用微信原生 `WXML/WXSS/TypeScript`，不改成 Flutter。
- 只访问 BFF 或公开 API。
- 主色固定 `#009bf5`，logo 来源 `assets/brand/logo.svg` 和欢迎页本地 `assets/brand/logo.jpg`。
- 商家店铺页统一展示外卖商品、团购套餐、评价、公告、商家资质和联系入口。
- 团购订单到店扫码验券；快递/跑腿进入骑手履约流程。
- 首页增加圈子入口，圈子首期是后台可控的小微墙，不整仓嵌入 InfiniLink。
- 找饭搭归属圈子能力，使用前必须完善性别、学校/楼栋、同校或同楼隐私范围、设备安全校验，签署身份真实性承诺和平台免责承诺，完成性格与饮食习惯问卷，并通过平台人工审核。
- 首页推荐商品、店铺、团购、优惠券和圈子卡片都从后台配置读取。
- 旧版已有能力必须保留：搜索、今日推荐、商品详情、购物车、地址、备注、餐具数量、支付结果、订单追踪、评价、售后、收藏、我的评价、红包优惠、钱包充值/提现/账单、积分商城、会员中心、反馈合作、公益。
- 余额支付前必须设置余额支付密码，密码未设置或锁定时不允许余额支付。
- 退款默认退回平台余额，后续以后台退款策略为准。
- 用户可按订单、售后和优惠活动维护外部通知渠道和静默时间，服务端按当前登录用户保存偏好。
- 消息页包含私聊、官方群、商户群；用户注册后自动加入官方群，默认不通知；打开会话会建立实时连接、同步离线消息并回写已读状态，服务端会在消息 API 和 WebSocket 握手期校验会话成员资格；商户群页、群设置页、店铺详情和红包优惠页会读取群资料/成员状态，并支持自助入群、退群、切换免打扰和领取商户群券。
- 用户可在群聊或私聊中用余额发送普通红包或拼手气红包，未领完金额按服务端过期退回能力回到发包人余额。
- 指定优惠券可配置为必须加入商户群后才能领取或使用。

当前后端对接明细见：`docs/product/user-miniprogram-backend-integration.md`。
