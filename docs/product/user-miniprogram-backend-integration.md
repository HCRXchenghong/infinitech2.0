# 用户端微信小程序 UI 与后端对接留痕

日期：2026-05-29

## 2026-06-07 消息体系会话模型修正留痕

- 修正消息详情页理解：不能把所有聊天都做成一个“平台 logo 头像”的通用群聊页；头像必须按真实身份区分，平台 logo 只用于官方客服或悦享e食总群标识，不作为普通商家、骑手、用户头像。
- 用户端消息体系按真实关系拆分：`悦享e食总群` 为注册账号后默认加入的平台总群；`商家群聊` 是某个商家的群，群内包含商家和用户们；`骑手私聊` 是用户和骑手；`饭搭/好友私聊` 是用户和用户；`官方客服` 走客服会话。
- `骑手-商家` 私聊属于骑手端和商户端消息能力，用户端不应越权展示该会话内容；后续做骑手端/商户端 Flutter App 时复用消息线程能力单独落页面。
- 当前后端已有 `GET /api/messages/threads`、线程概览、成员、同步、已读、偏好和实时 WebSocket 闭环；现阶段需要把前端消息列表与详情页按线程 `type / members` 准确渲染，避免继续用标题关键词粗暴归并。
- 已重新生成消息体系预览图：`docs/product/generated-user-assets/message-architecture-preview-2026-06-07.png`。该图只用于确认结构和视觉方向，暂未替换小程序代码。

## 2026-06-07 消息页会话图标生图替换留痕

- 按用户要求替换消息页四类会话入口图标：官方客服使用项目既有 `assets/brand/logo.jpg` 生成小尺寸运行图，商户、骑手、用户/群聊图标使用生图能力重新生成。
- 新增运行资产 `assets/generated/messages/message-merchant.png`、`message-rider.png`、`message-user.png`、`message-service-logo.jpg`，均压缩为 160x160 小图，避免继续扩大真机调试包体。
- `pages/messages/index` 从文字圆角标改为图片图标渲染，继续保留商家消息、骑手消息、群聊、官方客服四类聚合入口和后端 `GET /api/messages/threads` 归并能力。

## 2026-06-07 消息页极简聚合改版留痕

- 按用户确认的极简预览稿落地 `pages/messages/index`：去掉原消息页 Tabs、通知卡、复杂会话列表，只保留标题、搜索框和一个白色会话列表卡。
- 消息页同屏聚合四类入口：商家消息、骑手消息、群聊、官方客服；每行展示图标、标题、最近一条摘要、时间和未读点/未读数。
- 数据层保留 `GET /api/messages/threads` 同步能力，接口返回后按商家、骑手、群聊、客服归并到四行展示；接口不可用时使用本地四类兜底文案，不再把页面堆复杂。

## 2026-06-07 顶部标题统一留痕

- 按用户确认要求统一用户端小程序自定义顶部导航：标题统一为 `34rpx / 900`，使用全局绝对居中规则，不再被左侧返回键、右侧状态文案或微信胶囊避让区带偏。
- 全局隐藏顶部右侧 `nav-action / mp-action / nav-spacer` 视觉入口，搜索页不再在标题栏右侧显示“接口异常”，订单、搜索、消息、我的等页面标题规格保持一致。
- 保留必要返回按钮，但返回按钮只承担页面返回，不参与标题居中计算；后续若需要“设置、清空、地图、发布”等操作入口，应迁移到页面内容区按钮承接，避免顶部导航再次变乱。

## 2026-06-07 我的订单页 UI 改版留痕

- 按用户确认方向直接落代码，不再生成二次预览：`pages/order/list/index` 改为贴近首页的浅蓝顶区、白卡列表、原生胶囊避让和蓝色选中态风格。
- 删除预览稿里的“本月订单 / 进行中 / 待评价 / 售后”统计概览行；页面从标题、搜索框、状态 Tab 后直接进入订单卡列表。
- 状态 Tab 调整为“全部、待支付、进行中、待评价、退款/售后”，默认选中“全部”；新增订单关键词搜索，支持按商家、商品、状态、时间过滤。
- 订单卡改为一行一个大卡：商家图、商家名、时间、状态胶囊、商品摘要、金额、配送进度条和操作按钮；进行中订单支持“联系骑手 / 查看详情”，待评价订单支持“再来一单 / 去评价”。
- 该页仍保留 `GET /api/orders` 后端同步逻辑，后端无数据或失败时使用页面兜底数据预览；底部导航继续使用微信原生 `tabBar`，不再绘制自定义底栏。

## 2026-06-06 原生 Tab、真实定位与代码治理留痕

- 问题确认：底部导航此前是各页面自绘 `view`，点击使用 `navigateTo/redirectTo`，所以用户体感像普通页面跳转；改为微信原生 `tabBar` 后，四个主入口统一走小程序原生 Tab 行为。
- `app.json` 已新增原生 `tabBar`：首页、订单、消息、我的使用本轮单独生成并压缩后的图片图标，保留欢迎页、登录页、注册页作为入口页，不删除原有动效与认证流程。
- 新增 `utils/navigation.ts`，统一处理 `openRoute/switchTabRoute/backOrSwitchTab`；后续进入原生 Tab 页统一用 `wx.switchTab`，避免 Tab 页被当普通页跳转。
- 已清理首页、订单列表、消息中心、我的页和圈子页残留的自绘底部导航节点、底部固定样式与冗余跳转函数；圈子页不是原生 Tab 页，不再伪装成“消息”Tab 选中态。
- 登录、注册、欢迎页进入首页已改为 `switchTabRoute("/pages/index/index")`；全项目已扫描并修正直接 `redirectTo` 到首页/订单/消息/我的的返回兜底。
- 首页左上角位置从硬编码“望京SOHO”改为真实定位能力：进入首页会调用 `wx.getLocation({ type: "gcj02" })` 获取真实坐标，点击位置进入独立选择位置页，并通过该页缓存到 `homeLocation`；`app.json` 已声明 `scope.userLocation`、`getLocation`、`chooseLocation`。
- 新增 `utils/location.ts` 统一读取首页定位缓存；首页展示可使用该缓存，确认订单、地址列表、买药订单、跑腿下单和订单详情均改为后端地址/真实地图点或“待同步/待补充”文案，不再把缓存定位伪造成收货地址。
- 2026-06-06 按已确认预览稿新增“选择位置”页：`pages/location/select/index`。首页位置点击改为进入独立页面，页面提供搜索入口、重新定位、微信地图手动选点、后端已保存地址和确认当前位置；确认后统一写入 `homeLocation` 并回到首页刷新展示。
- 2026-06-06 地址真实化修正：移除选择位置页的预览地点、地址列表的 `cached_home_location` 兜底、外卖/买药确认单的默认联系人、跑腿下单/详情的演示取送地址和备注。正式收货地址只读取 `GET /api/user/addresses`，新增地址必须通过 `wx.chooseLocation` 获取真实坐标并用登录态联系人提交 `POST /api/user/addresses`；无后端地址时展示空状态，不再伪造地址。
- 2026-06-06 跑腿地址真实化：跑腿首页取件/送达两行改为微信地图真实选点入口，提交 `POST /api/errand/orders` 前校验真实取件地址、真实送达地址、登录联系人和物品描述；不再提交默认手机号或演示备注。
- 2026-06-06 买药配送地址真实化：买药确认订单页读取后端默认/已选地址，点击地址卡进入地址选择；无真实后端地址时阻止提交 `POST /api/medicine/orders`，失败时不再跳转本地预览订单。
- 订单列表完成第一批页面精修：状态 Tab 改为真实筛选、增加订单概览区、空状态、状态胶囊和更贴近首页浅蓝体系的卡片样式；同时保留后端 `GET /api/orders` 同步逻辑和无后端预览兜底。
- 代码乱点审计：当前仍存在较多页面级兜底假数据、局部页面样式重复和部分页面视觉未统一；本轮先治理会影响稳定性的导航/定位基础层，后续逐页精修时继续把页面兜底数据迁到 BFF/CDN 配置与统一组件/样式层。

## 2026-06-06 底部导航图标单独生图接入

- 按用户确认方向单独生成 8 个底部导航图标：`首页 / 订单 / 消息 / 我的` 各自包含选中态与未选中态；不是从预览图裁切。
- 源图留存在 `docs/product/generated-user-assets/tabbar-icons-source/`，最终小程序运行资产为 `apps/user-wechat-miniprogram/assets/generated/tabbar/tabbar-*.png`。
- 已做绿底抠透明、裁切留白和 96x96 PNG 压缩，最终总览图为 `docs/product/generated-user-assets/tabbar-icons-final-contact-sheet-2026-06-06.png`。
- `utils/media.ts` 新增 `tabbarIconImage(key, active)`，统一返回底栏图标路径，后续可替换为后台/CDN 配置。
- 首页、订单列表、消息中心、圈子页和我的页的自绘底部导航已切换为图片图标 + 文本标签；选中态使用蓝色图标，未选中态使用加深灰蓝图标，避免未选中态过淡。
- 真机调试包体修复：微信报 `source size 3341KB exceed max limit 2MB` 后，压缩小程序运行包图片资产。`assets/generated` 从约 2.9MB 降到约 672KB，`apps/user-wechat-miniprogram` 目录从约 3.9MB 降到约 1.68MB；高清源图继续留在 `docs/product/generated-user-assets/`，运行包只保留小图兜底。

## 2026-06-05 顶部导航统一与首页确认稿

- 小程序全局 `window.navigationStyle` 保持 `default` 兜底；已注册页面通过各自 `index.json` 明确导航模式，避免系统导航和页面自定义导航叠在一起。
- `app.wxss` 已统一 `.plain-nav`、`.blue-nav`、`.nav`、`.profile-nav`、`.mp-nav` 的三列布局、安全区、标题居中、返回按钮和右侧操作区样式。
- 购物车、通知偏好、地址、圈子、找饭搭、支付密码、消息、订单、我的等业务页已补齐或对齐顶部导航结构；欢迎页、登录页、注册页保持沉浸式入口，不做替换。
- 已生成首页新版效果图确认稿，文件见 `docs/product/generated-user-assets/homepage-redesign-preview-single-carousel-2026-06-05.png`；已按该稿替换小程序首页代码，只保留上方主轮播，分类下方改为“附近优选 / 今日团购 / 后台推荐”的静态卡片。
- 已修复微信系统胶囊遮挡问题：业务页 `index.json` 显式配置 `navigationStyle: custom`；首页移除自绘微信胶囊，店铺详情移除自绘 `•••`；全局导航统一增加顶部状态栏高度和右侧胶囊避让区，右侧操作按钮不会再压到系统胶囊。
- 首页参照美团顶部的原生胶囊骨架调整：`pages/index/index.json` 使用 `navigationStyle: custom`，页面顶部只保留微信系统状态栏/胶囊，不再自绘左上角房子或假胶囊；第一行是定位信息 + 轻量操作胶囊 + 系统胶囊避让位，第二行是搜索框。
- 首页定位与搜索区保持 `home-sticky` 吸顶层，向下滚动时固定在顶部；背景仍使用本项目浅蓝系，不沿用美团黄色。
- 首页顶部“消息”胶囊已删除，定位左侧从字符符号改为 CSS 绘制的蓝色定位针；8 个首页分类图标已替换为统一重制版，运行资产覆盖 `apps/user-wechat-miniprogram/assets/generated/category-*.png`，拼图留痕见 `docs/product/generated-user-assets/category-icons-v2-contact-sheet.png`。
- 2026-06-05 按用户确认稿落代码：首页主视觉改为 3 张 `swiper` 自动轮播；八宫格下方先放横滑小卡「今日团购」，首屏约 2.5 张并追加“查看更多”入口；「外卖优选」下移为一行一个的大卡列表；首页「后台推荐」展示区已删除，后续轮播/团购/外卖可继续迁移为后台 CDN 链接配置。
- 2026-06-05 首页外卖优选卡片微调：右下角从“去看看”按钮改为价格展示，外卖兜底数据补 `price` 字段，后续接后台时可由商品/店铺推荐接口返回。

## 2026-06-04 生图资产与 CDN 化接入

- 本轮不再从参考图裁素材；首页、商家列表和店铺详情首批业务图片已改为内置生图能力生成，高清留痕见 `docs/product/generated-user-assets/`。
- 小程序运行包只保留压缩 JPG 兜底图，位置为 `apps/user-wechat-miniprogram/assets/generated/`，当前目录约 1MB，避免主包被高清 PNG 撑大。
- 页面统一消费链接字段：`imageUrl`、`image_url`、`cover_url`、`logo_url`。后端/BFF/API 返回值优先，本地生图只做无后台配置时的预览兜底。
- 首页已接 `heroImageUrl`、`featuredDeal.imageUrl`、`recommendCards[].imageUrl`、`mealMatchCard.imageUrl` 和 `homeCards[].imageUrl`。
- 首页八个分类入口已从表情图标改为生图分类图标，页面优先读取 `HomeModule.icon_url`，无后台配置时按模块 key 回退到本地 `category-*.png`。
- 商家列表已接 `shops[].imageUrl`，会优先读取后端 `cover_url`、`logo_url`、`image_url`。
- 店铺详情已接 `shop.coverUrl` 和 `products[].imageUrl`，会优先读取 `GET /api/shops/{shopID}/detail` 的 `cover_url/logo_url` 与 `GET /api/shops/{shopID}/products` 的 `image_url`。
- 购物车、确认订单、订单列表和订单详情已接商品图片字段：后端 `CartItem.image_url`、`OrderItem.image_url` 会从 `merchant_products.image_url` 带出；小程序缺省时回退到生图兜底资产。
- 买药首页、买药确认订单和买药订单详情已接医疗封面/药品图片字段：`MedicineClinic.cover_url`、`MedicineProduct.image_url`、`MedicineOrderItem.image_url` 优先，缺省时回退本轮生成的校医务室、退热贴、处方胶囊和急救用品图。
- 跑腿首页和跑腿订单详情已接跑腿场景/包裹图片字段：下单请求和详情返回都支持 `image_url`，缺省时回退本轮生成的跑腿封面和包裹缩略图。
- 搜索聚合页已接 `SearchResult.image_url`，商家、菜品、团购、买药、跑腿结果均优先显示后端/CDN 链接，缺省时按类型回退本地生图资产。
- BFF `GET /api/home/cards` 和 API-Go 默认 `HomeCard.ImageURL` 已带默认图片链接；API-Go 店铺 `Shop.CoverURL`、商品 `MerchantProduct.ImageURL` 已从旧 `/assets/mock` 迁到本轮生成图路径。
- 后续上线 CDN 时，把后台配置或对象存储公共 URL 改成 `https://cdn...` 即可；前端页面不需要再从本地路径改结构。

## 2026-06-05 首页分类图标生图替换

- 生成 8 个首页分类图标：外卖、团购、买药、快递跑腿、圈子、找饭搭、红包优惠、会员积分；高清 PNG 留痕在 `docs/product/generated-user-assets/`，小程序运行图在 `apps/user-wechat-miniprogram/assets/generated/category-*.png`。
- `pages/index/index`：`modules[].iconImageUrl` 优先渲染图片，`iconGlyph` 只作为兜底，不再默认使用表情图标。
- `utils/media.ts`：新增 `categoryIconImage(key)`，统一按模块 key 获取本地兜底图标。
- API-Go：`HomeModule` 新增 `icon_url`；`DefaultHomeModules()` 默认启用 8 个分类，并返回对应图标路径。
- BFF：`defaultHomeModules()` 同步返回 8 个分类和 `icon_url`，本地预览与 API-Go 保持一致。

## 2026-06-04 买药与跑腿图片化精修

- `pages/medicine/home/index`：校医务室卡片新增 `clinic.coverUrl`，药品列表新增商品图；远程 `cover_url/image_url` 优先，本地生图兜底。
- `pages/medicine/order-confirm/index`：药品清单从“药”字块改为药品图片，提交 `POST /api/medicine/orders` 时同步传 `items[].image_url`。
- `pages/medicine/order-detail/index`：订单详情药品清单读取后端 `items[].image_url`，旧订单或无图订单按药品 ID 回退。
- API-Go：`MedicineClinic` 新增 `cover_url`；默认 `MedicineProduct` 和 `MedicineOrderItem` 写入本轮生成图路径，药品下单与详情复用同一字段。
- `pages/errand/home/index`：跑腿首页封面改为生图资产，物品与要求区域新增包裹缩略图，下单传 `image_url`。
- `pages/errand/order-detail/index`：跑腿订单详情读取 `image_url`，无后台配置时显示本地包裹兜底图。
- API-Go：`ErrandOrderRequest`、`ErrandOrderDetail` 新增 `image_url`，默认值为 `/assets/generated/errand-parcel.jpg`。
- `pages/search/index`：搜索结果卡片从文字缩略改为图片缩略；`GET /api/search` 的 `image_url` 为空时只显示类型占位，不再用前端本地生图冒充接口图片。
- API-Go：搜索结果里的商家补 `Shop.CoverURL/LogoURL`，买药补默认药品结果，跑腿补包裹图，供聚合搜索直接渲染。

## 2026-06-04 交易主链路图片化精修

- `pages/cart/index`：购物车抽屉商品行已从文字缩略改为商品图片，远程 `image_url` 优先，本地生图兜底。
- `pages/order/confirm/index`：确认订单商品明细已补商品图、数量和金额三列结构，减少纯文本清单感。
- `pages/order/list/index`：订单卡片已补首个商品/店铺图片，后端订单项带图时直接显示 CDN 链接。
- `pages/order/detail/index`：订单商品明细已补商品图，和列表、确认页保持统一。
- API-Go：`CartItem`、`OrderItem` 新增 `image_url`；内存 Store 与 PostgreSQL 查询都从 `merchant_products.image_url` 回填，不新增数据库字段。

## 本次结论

- 用户端原生微信小程序已注册 39 个页面：`00A` 欢迎页、`00B` 登录页、`00C` 注册页和业务参考图 `01` 到 `35`，并新增商户群群设置页。
- 已补齐每个页面的 `index.json`、`index.wxml`、`index.ts`，可在微信开发者工具中逐页打开。
- 高频交易页 `01` 到 `07` 已按参考图完成首轮精修：首页、附近商家、店铺详情、购物车抽屉、确认订单、订单列表和订单详情；其中店铺详情的评价/商家两栏已接 `GET /api/shops/{shopID}/detail` 聚合接口。
- 2026-06-05：按用户确认的两张生成图重做首页八宫格分类图标；外卖使用单独的 Infinitech S logo 送餐箱图，其余七类从确认的 4x2 contact sheet 拆分，已替换小程序 `apps/user-wechat-miniprogram/assets/generated/category-*.png`，源图与应用预览留存在 `docs/product/generated-user-assets/`。
- 2026-06-07：我的订单页按二次反馈完成顶部细节调整；标题改为整屏居中，右侧“客服”入口移除，订单卡片内配送进度条移除，保留搜索、状态筛选和订单操作能力。
- 2026-06-07：搜索页去掉前端本地 mock 初始化、离线缓存兜底和开发预览 token 依赖；小程序空关键词只展示后端 `suggestions` 生成的“猜你喜欢”，输入关键词后才展示 `GET /api/search` 返回的真实结果，接口失败或无数据时显示空态。API-Go `/api/search` 同步改为匿名可读公共目录搜索，`suggestions` 由真实目录结果生成；BFF 和 API-Go 均已补搜索闭环回归测试。
- 2026-06-07：我的订单页同步去掉三条前端预览订单和自动注入 `user_1` 开发 token；订单搜索/状态筛选只针对 `GET /api/orders` 的真实返回数据生效，未登录或接口异常时显示空态。
- 设置与消息页 `08` 到 `15` 已按参考图完成首轮精修：通知偏好、支付密码、地址、圈子、找饭搭、售后、评价、消息中心。
- 资产与服务页 `16` 到 `24` 已按参考图完成首轮精修：我的、钱包、红包优惠、会员积分、邀请好友、搜索、买药、快递跑腿、跑腿订单详情。
- 已扩展 `apps/user-wechat-miniprogram/utils/api.ts`，优先接入当前 BFF/API 已存在的用户端接口。
- 手机号验证码登录/注册、真实微信 `wx.login` 页面流程、生产短信 provider 配置入口、评价、钱包账单、钱包总览、提现申请、优惠券、会员积分/签到、邀请好友、搜索聚合、买药首页、跑腿下单/详情、反馈、圈子动态、找饭搭资料人工审核/候选/举报拉黑/举报处置/同校同楼隐私/设备风控、消息会话/聊天/已读回执/离线补偿/WebSocket 首版投递、消息 PostgreSQL 规范化表/已读状态表首版、realtime gateway Redis adapter 多副本 fanout 首版、realtime gateway WebSocket 签名鉴权首版、realtime gateway 会话成员权限校验首版、会话免打扰偏好首版、群资料/成员预览首版、商户群自助加入/退出、商户群券资格校验、商户群聊、消息敏感信息风控、红包发送/详情/领取/退回/过期退回/领取风控、客服工单分派/关闭/回访、客服 SLA 状态/超时升级、客服质检/绩效、Admin Web 客服工作台可视化首版、处方影像上传票据/上传回调/确认/对象扫描门禁/OCR 识别/药师复核/留档/审核绑定、药品订单确认/详情和药品库存锁定已补首批 API-Go/BFF/小程序调用闭环；短信模板审批、断线重连/压测、真实打款和剩余 PostgreSQL 规范化专项仍是后续生产化边界。

## 已接入或部分接入后端

| 页面/能力 | 当前接入 |
| --- | --- |
| 首页 | `GET /api/home/modules`；轮播、今日团购、外卖优选当前使用前端兜底图片/数据，后续迁移为后台 CDN 链接配置 |
| 微信登录 | `POST /api/auth/wechat-mini/login`；页面通过 `wx.login` 获取真实 code 后提交 BFF/API，生产失败会停留并提示；仅本地 API 或显式 `allowPreviewAuth=true` 时允许开发预览 token 兜底 |
| 手机号验证码登录/注册 | `POST /api/auth/phone/code`、`POST /api/auth/phone/login`、`POST /api/auth/phone/register`；当前支持开发验证码回传、生产短信 provider endpoint 配置、验证码生产模式隐藏、发送状态、重发冷却、手机号小时/日频控、手机号验证码登录、密码登录、注册协议确认和注册后签发用户 token |
| 附近商家/搜索/买药药房 | `GET /api/shops` |
| 店铺详情 | `GET /api/shops/{shopID}/detail`；当前聚合头图、公告、活动标签、评分摘要、评价列表、营业时间、电话、地址、服务承诺和商家资质，评价卡会额外返回 `image_urls`、`item_highlights` 和 `rider_stars_text`，供店铺详情页的“评价 / 商家”两栏直接渲染 |
| 店铺商品 | `GET /api/shops/{shopID}/products` |
| 店铺团购 | `GET /api/shops/{shopID}/groupbuy-deals`、`POST /api/groupbuy/orders` |
| 地址列表/新增地址 | `GET /api/user/addresses`、`POST /api/user/addresses`；新增地址通过微信地图真实选点获得 `latitude/longitude`，并要求登录态联系人和可识别城市；当前同一路径也用于按 `id` 回写默认地址，确认订单页/买药确认页可从地址列表选择配送地址 |
| 我的 | `GET /api/user/profile`；返回会员、资产、订单计数、支付密码状态 |
| 购物车 | `GET /api/cart`、`POST /api/cart/items`；当前 `GET /api/cart` 会返回 `shop_name`、商品清单和费用汇总，购物车页支持真实增减与清空 |
| 外卖确认订单 | `POST /api/orders/checkout`；确认订单页会先读取 `GET /api/cart` 与 `GET /api/user/addresses`，再按选中地址提交下单 |
| 药品订单 | `POST /api/medicine/orders`、`GET /api/medicine/orders/{orderID}`；买药确认页提交前必须选中后端真实配送地址，当前含校医务室、处方审核状态、药品清单、库存锁定/剩余库存、费用明细和履约时间线；库存不足返回 `INSUFFICIENT_STOCK` |
| 跑腿下单/详情 | `POST /api/errand/orders`、`GET /api/errand/orders/{orderID}`；跑腿下单页取件/送达均走微信地图真实选点，提交前校验联系人、地址和物品描述，当前为内存 Store 首版专属地址、备注、费用、骑手和进度 |
| 订单列表/详情 | `GET /api/orders`、`GET /api/orders/{orderID}`；当前订单会带 `shop_name`、`address_snapshot` 和 `reviewed`，供用户端直接渲染商家名、配送地址和待评价状态 |
| 售后申请/记录 | `GET /api/after-sales?order_id=`、`POST /api/after-sales`；当前售后记录支持按订单筛选，并返回 `shop_name`、`order_status`、`order_item_summary` 和最近用户可见进度 |
| 售后工单事件/证据 | `GET /api/after-sales/{requestID}/events`、`POST /api/after-sales/{requestID}/events`、`GET /api/after-sales/{requestID}/evidence`、`POST /api/after-sales/{requestID}/evidence/upload-ticket`、`POST /api/after-sales/{requestID}/evidence/confirm`、`POST /api/object-storage/upload-callback`、`POST /api/object-storage/scan-result`；当前用户端售后页已接补充凭证上传首版，默认环境可直确认，严格环境会先补对象存储上传回调，再按需要补扫描通过结果后重试确认 |
| 评价订单 | `GET /api/reviews?order_id=`、`POST /api/reviews`、`POST /api/reviews/upload-ticket`、`POST /api/reviews/upload-confirm`；当前支持订单级评价查询、同订单更新覆盖、匿名评价、文字、评分、标签、`item_ratings` 逐项菜品评分、`rider_rating` 配送服务独立评分，以及评价图片上传票据/确认；默认环境可直确认，严格环境会先补对象存储上传回调，再按需要补扫描通过结果后重试确认 |
| 通知偏好 | `GET /api/user/notification-preferences`、`PUT /api/user/notification-preferences` |
| 钱包总览/充值/账单/提现/支付密码/余额支付工具 | `GET /api/wallet/overview`、`POST /api/wallet/credit`、`GET /api/wallet/transactions`、`POST /api/wallet/withdraw`、`POST /api/wallet/payment-password`、`POST /api/wallet/pay` |
| 微信预支付工具 | `POST /api/payments/wechat/prepay` |
| 团购券列表 | `GET /api/groupbuy/vouchers` |
| 红包优惠/普通优惠券 | `GET /api/user/coupons`、`POST /api/user/coupons/claim`；当前含平台券、商户群券、团购券和买药券首版，商户群券需先满足群成员资格 |
| 会员积分/签到 | `GET /api/user/points`、`POST /api/user/points/check-in`；当前含任务、权益、兑换项和流水首版 |
| 邀请好友 | `GET /api/user/invite-summary`；当前含邀请码、分享路径、邀请记录和风控提示首版 |
| 混合搜索 | `GET /api/search?keyword=&category=`；当前聚合商家、菜品、团购、买药和跑腿结果 |
| 买药首页 | `GET /api/medicine/home`；当前含校医务室、分类、药品列表和购物车摘要首版 |
| 找饭搭 | `GET /api/meal-match/profile`、`PUT /api/meal-match/profile`、`GET /api/meal-match/candidates`、`POST /api/meal-match/reports`、`POST /api/meal-match/blocks`、`GET /api/admin/meal-match/moderation`、`POST /api/admin/meal-match/moderation/{recordID}/review`；当前含前置资料、资料人工审核状态、候选匹配分、同校/同楼隐私范围、模糊位置、设备风控、隐私提示、举报待审/处置和不感兴趣拉黑 |
| 反馈投诉 | `GET /api/feedback`、`POST /api/feedback` |
| 客服工单 | `GET /api/service-tickets`、`POST /api/service-tickets`、`GET /api/service-tickets/{ticketID}`、`POST /api/service-tickets/{ticketID}/events`、`POST /api/service-tickets/{ticketID}/close`、`POST /api/service-tickets/{ticketID}/follow-up`、`GET /api/admin/service-tickets?sla_status=`、`POST /api/admin/service-tickets/{ticketID}/assign`、`POST /api/admin/service-tickets/{ticketID}/escalate`、`POST /api/admin/service-tickets/{ticketID}/resolve`、`POST /api/admin/service-tickets/{ticketID}/quality-review`、`GET /api/admin/service-ticket-quality-reviews`、`GET /api/admin/service-ticket-performance`；当前含工单列表、详情、SLA 状态、超时升级、客服分派、处理方案、用户确认关闭、回访评分、质检抽检、客服绩效和敏感信息风控 |
| 圈子小微墙 | `GET /api/circle/posts`、`POST /api/circle/posts`；当前为内存 Store 首版发布/列表 |
| 消息中心/群聊 | `GET /api/messages/threads`、`GET /api/messages/{threadID}/overview`、`GET /api/messages/{threadID}/members`、`GET /api/messages/{threadID}/membership`、`POST /api/messages/{threadID}/join`、`POST /api/messages/{threadID}/leave`、`GET /api/messages/{threadID}`、`GET /api/messages/{threadID}/sync?since_id=&mark_read=`、`GET /api/messages/{threadID}/preference`、`PUT /api/messages/{threadID}/preference`、`POST /api/messages/{threadID}/read`、`POST /api/messages/{threadID}`、`ws://.../ws?thread_id=`、`POST /internal/realtime/publish`、`POST /internal/realtime/authorize`；当前支持按游标补离线消息、打开会话自动清未读、手动已读回执、群资料概览、活跃成员预览、商户群自助加入/退出、会话免打扰切换、敏感信息风控、`message.sent` outbox 事件、PostgreSQL `messages` / `conversation_read_states` 规范化首版、`conversation_members` 成员快照落库、realtime gateway Redis Pub/Sub 多副本 fanout 首版、WebSocket 签名 token 鉴权首版和握手期会话成员权限校验，outbox relay 会把 `message.sent` 路由到任意 realtime-gateway 副本，小程序商户群页可实时接收同会话消息 |
| 红包发送/详情/领取/退回 | `POST /api/red-packets`、`GET /api/red-packets/{packetID}`、`POST /api/red-packets/{packetID}/claim`、`POST /api/red-packets/{packetID}/refund`、`POST /api/admin/red-packets/expire`；当前支持余额冻结、领取入账、发包人退回、过期批量退回、领取频次/金额风控和钱包流水 |
| 处方审核 | `POST /api/prescriptions/upload-ticket`、`POST /api/object-storage/upload-callback`、`POST /api/object-storage/scan-result`、`POST /api/prescriptions/upload-confirm`、`POST /api/prescriptions`、`GET /api/prescriptions/{reviewID}`、`GET /api/admin/prescriptions`、`POST /api/admin/prescriptions/{reviewID}/review`；当前支持处方影像对象存储上传票据、上传回调、扫描通过后确认、OCR 识别摘要、处方留档、药师复核、审核单绑定对象 key/hash/公开 URL，并返回审核节点、校医和药品信息 |

## 后端仍待补齐

| 页面/能力 | 缺口 |
| --- | --- |
| 手机号短信生产化 | provider endpoint、验证码隐藏、发送状态、重发冷却和手机号小时/日频控首版已接；真实短信模板审批、图形/设备风控、黑名单、验证码审计归档和 provider 回执重试待补 |
| 消息/群聊/私聊 | 会话列表、聊天记录、发送、已读回执、离线游标补偿、`message.sent` outbox、realtime gateway WebSocket 首版投递、Redis adapter 多副本 fanout、WebSocket 签名 token 鉴权、握手期会话成员权限校验、会话免打扰偏好、群资料概览、活跃成员预览、商户群自助加入/退出、PostgreSQL `messages` 规范化消息表和 `conversation_read_states` 已读状态表首版已接；动态群管理、token 撤销/会话状态联动、断线重连策略、消息顺序保障、Redis failover 演练和压测待补 |
| 红包 | 余额冻结、领取入账、发包人退回、过期批量退回和领取风控首版已接；真实 24 小时调度器、支付密码校验、群成员资格、红包 PostgreSQL 规范化表和财务对账待补 |
| 钱包提现 | 首版申请已接；真实微信零钱/银行卡打款、回调、失败重试、风控复核和 PostgreSQL 规范化状态机待补 |
| 优惠券 | 首版列表/兑换已接；商户群券已支持按群成员资格领取，使用锁定、核销、补贴结算和反作弊待补 |
| 会员积分/邀请 | 首版积分/签到/邀请摘要已接；真实奖励结算、邀请归因、积分商城库存和反作弊待补 |
| 找饭搭生产化 | 资料人工审核、候选、同校/同楼隐私范围、模糊位置、设备风控、举报待审/处置和拉黑已接首版；真实设备指纹 provider、IP/地理异常、举报分级策略、精准地理隐私 provider 和 PostgreSQL 规范化待补 |
| 反馈投诉 | 客服工单分派、SLA 状态、超时升级、处理方案、用户关闭、回访评分、后台工单列表、质检抽检、客服绩效、客服消息敏感信息风控和 Admin Web 客服工作台首版已接；客服质检审计同事务和 PostgreSQL 规范化表待补 |
| 处方/药品订单 | 处方影像上传票据、上传回调、对象扫描门禁、上传确认、OCR 识别摘要、药师复核工作台、处方留档、审核单绑定、状态查询、药品订单确认/详情和药品库存锁定已接；真实 OCR provider、药品批次效期和 PostgreSQL 规范化表待补 |
| 跑腿详情 | 首版跑腿下单/详情已接；微信地图真实取送地址已接，距离计费、骑手轨迹、补差价和取消费规则待补 |

## 验证命令

```bash
node - <<'NODE'
const fs = require('fs');
const app = JSON.parse(fs.readFileSync('apps/user-wechat-miniprogram/app.json','utf8'));
const missing = [];
for (const page of app.pages) {
  for (const ext of ['json','wxml','ts']) {
    const file = `apps/user-wechat-miniprogram/${page}.${ext}`;
    if (!fs.existsSync(file)) missing.push(file);
  }
}
console.log(`pages=${app.pages.length}`);
if (missing.length) throw new Error(missing.join('\n'));
console.log('all page json/wxml/ts files exist');
NODE
git diff --check
npm run verify
```
