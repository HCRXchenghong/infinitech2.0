# 旧版仓库复用审计

审计时间：2026-05-21  
来源仓库：[HCRXchenghong/infinitech](https://github.com/HCRXchenghong/infinitech)  
补充参考仓库：[HCRXchenghong/InfiniLink](https://github.com/HCRXchenghong/InfiniLink)  
审计方式：查看 GitHub README，并浅克隆到 `/tmp/infinitech-legacy`、`/tmp/InfiniLink-reference` 做只读分析。

## 1. 旧版概况

旧版 README 显示它是一个本地生活 monorepo，包含：

- `admin-vue`：管理端 Web、官网、邀请页、下载页、运行时设置、客服工作台。
- `user-vue`：用户端 跨端/小程序风格工程。
- `app-mobile`：用户端 App 形态。
- `merchant-app`：旧版商户端跨端页面，可作为 Flutter/Dart 重建参考。
- `rider-app`：旧版骑手端跨端页面，可作为 Flutter/Dart 重建参考。
- `backend/go`：核心业务 API。
- `backend/bff`：聚合层。
- `socket-server`：Socket.IO 实时网关、聊天、通知、RTC 信令。
- `packages`：contracts、client-sdk、domain-core、mobile-core、admin-core。

这说明旧版已经具备不少业务雏形，但结构过重、端形态混杂，不适合作为 2.0 直接续写基础。

## 2. 品牌资产

已复制到 2.0：

- `assets/brand/logo.svg` 来源：`app-mobile/static/images/logo.svg`
- `assets/brand/logo.jpg` 来源：`app-mobile/static/images/logo.png`
- `assets/brand/admin-logo.png` 来源：`admin-vue/public/logo.png`

旧版 SVG logo 关键色：

- `#009bf5`
- `#0081cc`
- 透明背景

## 3. UI 主题

旧版高频颜色：

- 主色：`#009bf5`
- 深一点的品牌蓝：`#0081cc`
- 渐变/强调蓝：`#0284c7`
- 页面背景：`#f3f4f6`
- 轻背景：`#f5f7fb`、`#f8fbff`
- 文字主色：`#1f2937`
- 文字次色：`#6b7280`
- 边框：`#e5e7eb`
- 警示/补贴：`#ea580c`
- 错误：`#ef4444`

2.0 设计 tokens 建议：

```css
:root {
  --it-color-primary: #009bf5;
  --it-color-primary-dark: #0081cc;
  --it-color-primary-strong: #0284c7;
  --it-color-bg: #f3f4f6;
  --it-color-surface: #ffffff;
  --it-color-text: #1f2937;
  --it-color-text-muted: #6b7280;
  --it-color-border: #e5e7eb;
  --it-color-warning: #ea580c;
  --it-color-danger: #ef4444;
}
```

## 4. 可复用模块

| 旧版位置 | 复用价值 | 2.0 处理 |
| --- | --- | --- |
| `app-mobile/static/images/logo.svg` | 品牌 logo | 已复制，作为统一品牌源 |
| `user-vue` / `app-mobile` | 用户端视觉、页面信息架构 | 原生微信小程序重写，视觉参考 |
| `merchant-app` | 商户端页面和交互 | 可迁移到 2.0 `merchant-flutter` |
| `rider-app` | 抢单大厅、任务、钱包、接单设置 | 可迁移到 2.0 `rider-flutter` |
| `admin-vue` | 后台模块、邀请页、客服、RTC 管理 | 可复用 UI 和业务信息架构 |
| `backend/go` | 订单、钱包、支付、团购、买药、消息、RTC、邀请 | 作为领域模型参考，代码择优迁移 |
| `backend/bff` | 前端聚合层思路 | 可保留 BFF 边界 |
| `socket-server` | IM、通知、RTC 信令、Redis adapter | 可作为实时网关起点 |
| `packages/contracts` | 响应规范、上传、身份模型 | 可升级为 2.0 contracts |
| `packages/mobile-core` | 移动端业务纯函数和测试 | 提取无框架逻辑，避免端强绑定 |
| `docs/operations/realtime-100k-acceptance.md` | 10 万在线验收原则 | 直接继承验收思想 |

## 5. 不建议复用的部分

- 旧版用户端跨端工程代码不能作为用户端主线，因为 2.0 要先做原生微信小程序。
- `admin-win`、`admin-mac` 暂不纳入首发，避免桌面壳增加维护面。
- 支付宝 sidecar、银行 payout sidecar 暂不首发，先聚焦微信支付和余额。
- 旧版根目录过多端并列，2.0 需要先定义清楚 `apps/services/packages/infra`。
- 旧版部分支付字段使用 `ifpay` 这类历史命名，2.0 应统一业务枚举。

## 6. 已发现旧版能力

用户端页面能力：

- 首页、分类、美食/甜品/商超/水果/买药、商家详情、购物车、订单、消息、RTC、钱包、邀请好友、客服、跑腿、买药。
- 继续细分：搜索、今日推荐、位置选择、收货地址、订单备注、餐具数量、支付结果、评价订单、申请售后、通知详情、我的收藏、我的评价、红包优惠、充值、提现、账单、设置、修改手机号、积分商城、会员中心、反馈合作、同频饭友、悦享公益。

商户端页面能力：

- 登录、经营概况、订单管理、在线沟通、商品管理、店铺设置、钱包。

骑手端页面能力：

- 登录、抢单大厅、任务、钱包、收入、历史订单、接单设置、健康证、保险、申诉、客服。
- 继续细分：忘记密码、设置密码、账单明细、钱包充值、钱包提现、数据统计、头像上传、个人信息、修改手机号、修改密码、骑手之家、开发者选项。

管理端页面能力：

- 仪表盘、订单、用户、商户、骑手、支付中心、财务、客服工作台、RTC 管理台、RTC 审计、跑腿配置、内容设置、邀请落地页、API 管理。
- 继续细分：售后、电话联系审计、优惠券领取落地页、优惠券管理、数据管理、饭搭治理、精选商品、首页活动、首页入口设置、通知编辑/预览/官方通知、官网内容、交易日志、骑手等级设置、店铺编辑、商品菜单管理、系统日志、开放平台权限与文档。

后端能力：

- Go API 已有订单、团购、买药、消息、钱包、支付、微信登录、RTC 审计、邀请、用户地址、通知、商户、骑手等模块。
- Socket 网关已有 support namespace、rider namespace、notify namespace、RTC namespace。
- 旧版已有 Redis adapter、Socket 鉴权、运行时配置、实时容量验收文档。

## 7. 2.0 迁移策略

第一步：

- 保留旧版品牌和视觉。
- 用文档冻结 2.0 的端划分和服务边界。
- 不把旧版整体复制进新仓库。

第二步：

- 从旧版提取设计 tokens、图标、页面结构。
- 原生小程序重写用户端。
- 商户端和骑手端用 Flutter/Dart 重建，旧版跨端工程只作为页面结构和交互参考。
- 管理端从旧版 `admin-vue` 迁移核心模块。

第三步：

- 后端按 2.0 数据模型和接口规范重建。
- 从旧版迁移可测试、可解释的纯业务逻辑。
- 支付、钱包、订单、调度必须补并发和幂等测试。

## 8. InfiniLink 圈子参考边界

InfiniLink 是原生微信小程序 + Go 后端的社区型项目，包含 content feed、circles、messaging、profile、search、payment entry flows 和 admin console。

2.0 处理方式：

- 只参考圈子列表、发帖、信息流、消息入口和用户资料入口的信息架构。
- 不整仓嵌入 InfiniLink。
- 不复用 InfiniLink 的旧支付、会员、成长体系和后端部署结构。
- 圈子能力在 2.0 中作为首页可配置模块，由 BFF 和后台配置控制开关、排序、审核和灰度。
- 找饭搭是 Infinitech 2.0 新业务，需要身份真实性承诺、平台免责承诺、问卷和风控审计，不直接等同于 InfiniLink 社交圈。

第四步：

- 建立压测环境。
- 完成 10k、30k、60k、100k 四档容量验收。
- 没有压测证据前不宣称已支撑 10 万在线。
