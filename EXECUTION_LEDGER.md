# Infinitech 2.0 执行台账

规则：

- 所有新增工作必须进入本台账。
- 已完成工作保留在“已完成”区，方便回溯。
- 未完成工作保留在“进行中/未开始”区，直到验收完成。
- 涉及 10 万在线、支付、钱包、订单、派单、RTC 的任务必须写明验收证据。

## 已完成

### DONE-20260521-001 本地工程状态确认

- 日期：2026-05-21
- 结果：`/Users/seron-cheng/Desktop/infinitech2.0` 是空目录，不是 git 仓库。
- 证据：本地 `ls -la` 和 `git status` 检查。

### DONE-20260521-002 旧版仓库审计

- 日期：2026-05-21
- 来源：[HCRXchenghong/infinitech](https://github.com/HCRXchenghong/infinitech)
- 结果：确认旧版包含用户端、商户端、骑手端、管理端、Go API、BFF、Socket 网关、共享 packages。
- 证据：`docs/LEGACY_REUSE_AUDIT.md`

### DONE-20260521-003 品牌资产迁移

- 日期：2026-05-21
- 结果：已迁移 logo 到 `assets/brand/`。
- 文件：
  - `assets/brand/logo.svg`
  - `assets/brand/logo.jpg`
  - `assets/brand/admin-logo.png`

### DONE-20260521-004 总计划建立

- 日期：2026-05-21
- 结果：已建立平台总计划。
- 文件：`PLATFORM_MASTER_PLAN.md`

### DONE-20260521-005 架构骨架落地

- 日期：2026-05-21
- 结果：已按“模块化核心 API + 事件驱动 Worker + 多端 BFF + 实时网关”落地第一层工程骨架。
- 文件：
  - `package.json`
  - `packages/contracts`
  - `packages/design-tokens`
  - `packages/domain-core`
  - `packages/client-sdk`
  - `packages/admin-core`
  - `services/api-go`
  - `services/bff`
  - `services/realtime-gateway`
  - `services/*-worker`
  - `infra/docker/compose.yml`
  - `infra/k8s/base`
  - `docs/architecture/system-architecture.md`
  - `docs/operations/capacity-and-dr.md`

### DONE-20260521-006 商家店铺与履约模型落地

- 日期：2026-05-21
- 结果：已明确商家账号、店铺、店铺能力、医务室/药房属性、快递/跑腿骑手履约、团购扫码验券模型。
- 文件：
  - `packages/contracts/src/index.mjs`
  - `packages/domain-core/src/index.mjs`
  - `services/api-go/internal/platform/contracts.go`
  - `PLATFORM_MASTER_PLAN.md`
  - `docs/architecture/system-architecture.md`

### DONE-20260521-007 商户邀请注册与资质规则落地

- 日期：2026-05-21
- 结果：已明确商家端只能通过管理员邀请链接注册，营业执照/健康证必须上传文件和失效日期，过期或未审核通过时店铺暂时关闭并要求商户端弹窗补资料，同时支持员工信息和补充资料模型。
- 文件：
  - `packages/contracts/src/index.mjs`
  - `packages/domain-core/src/index.mjs`
  - `services/api-go/internal/platform/contracts.go`
  - `PLATFORM_MASTER_PLAN.md`
  - `docs/architecture/system-architecture.md`
  - `apps/merchant-flutter/README.md`

### DONE-20260521-008 骑手站长、派单窗口与保证金规则落地

- 日期：2026-05-21
- 结果：已明确骑手端站长/骑手账号、邀约制注册、站长手动派单和每日任务时长配置、10 分钟抢单大厅后自动派单、拒绝后顺延下一位在线骑手、骑手/商户 50 元保证金、骑手微信免押和退押金延期规则。
- 文件：
  - `packages/contracts/src/index.mjs`
  - `packages/domain-core/src/index.mjs`
  - `services/dispatch-worker/src/index.mjs`
  - `services/api-go/internal/platform/contracts.go`
  - `PLATFORM_MASTER_PLAN.md`
  - `docs/architecture/system-architecture.md`
  - `apps/rider-flutter/README.md`

### DONE-20260521-009 骑手绩效、商户商品、退款和群聊红包模型落地

- 日期：2026-05-21
- 结果：已补充骑手接单耗时、团队均值、等级、派单优先级、每日固定订单数后免责不接；商户美团式店铺详情页、菜品图片/描述/配料表、团购二维码验券；退款默认退余额/原路返回策略、余额支付密码；官方群自动加入、商户群、进群领券、余额红包和拼手气红包。
- 文件：
  - `packages/contracts/src/index.mjs`
  - `packages/domain-core/src/index.mjs`
  - `packages/admin-core/src/index.mjs`
  - `services/dispatch-worker/src/index.mjs`
  - `services/api-go/internal/platform/contracts.go`
  - `PLATFORM_MASTER_PLAN.md`
  - `docs/architecture/system-architecture.md`
  - `apps/user-wechat-miniprogram/README.md`
  - `apps/merchant-flutter/README.md`
  - `apps/rider-flutter/README.md`
  - `apps/admin-web/README.md`

### DONE-20260521-010 圈子、饭搭、优惠券和首页卡片模型落地

- 日期：2026-05-21
- 结果：已参考 InfiniLink 的圈子/信息流结构，明确只借鉴不整仓嵌入；已补充首页圈子入口、小微墙、找饭搭性别/真实性/免责/问卷前置条件、商户自担券、平台补贴券、商户确认参与券、单店/参与商家券范围，以及后台可控首页卡片。
- 文件：
  - `packages/contracts/src/index.mjs`
  - `packages/domain-core/src/index.mjs`
  - `packages/admin-core/src/index.mjs`
  - `services/bff/src/runtime.mjs`
  - `services/bff/src/server.mjs`
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/httpapi/router.go`
  - `apps/user-wechat-miniprogram/pages/index/index.wxml`
  - `apps/user-wechat-miniprogram/pages/circle/index.wxml`
  - `apps/user-wechat-miniprogram/pages/meal-match/index.wxml`
  - `PLATFORM_MASTER_PLAN.md`
  - `docs/architecture/system-architecture.md`
  - `docs/LEGACY_REUSE_AUDIT.md`

### DONE-20260521-011 美团与旧版闭环能力对标补强

- 日期：2026-05-21
- 结果：已重新审计旧版用户端、商户端、骑手端、后台和 Go 服务，补齐旧版必须保留能力矩阵；已把地址、订单附加信息、售后、评价、收藏、积分会员、配送承诺、通知推送、风控、数据备份恢复等闭环基础能力纳入 contracts/domain/admin/API 计划。
- 文件：
  - `docs/product/meituan-legacy-parity-matrix.md`
  - `docs/LEGACY_REUSE_AUDIT.md`
  - `PLATFORM_MASTER_PLAN.md`
  - `packages/contracts/src/index.mjs`
  - `packages/domain-core/src/index.mjs`
  - `packages/admin-core/src/index.mjs`
  - `services/api-go/internal/platform/contracts.go`
  - `apps/user-wechat-miniprogram/README.md`
  - `apps/rider-flutter/README.md`
  - `apps/admin-web/README.md`

### DONE-20260521-012 用户外卖下单闭环首批实现

- 日期：2026-05-21
- 结果：已把第一条用户端外卖闭环从文档推进到可测代码：店铺列表、店铺商品、地址保存、购物车、购物车结算订单、余额支付、骑手抢单，并补齐小程序端首批预览页面。
- 验收证据：
  - `cd services/api-go && go test ./...`
  - `npm run verify`
- 文件：
  - `infra/db/migrations/0002_auth_payment.sql`
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/runtime.mjs`
  - `apps/user-wechat-miniprogram/app.json`
  - `apps/user-wechat-miniprogram/assets/brand/logo.jpg`
  - `apps/user-wechat-miniprogram/assets/brand/logo.svg`
  - `apps/user-wechat-miniprogram/pages/shop/list/index.wxml`
  - `apps/user-wechat-miniprogram/pages/shop/detail/index.wxml`
  - `apps/user-wechat-miniprogram/pages/cart/index.wxml`
  - `apps/user-wechat-miniprogram/pages/order/confirm/index.wxml`
  - `apps/user-wechat-miniprogram/pages/address/list/index.wxml`
  - `apps/user-wechat-miniprogram/README.md`
  - `docs/product/user-miniprogram-preview.svg`
  - `docs/product/user-miniprogram-preview.png`
  - `PLATFORM_MASTER_PLAN.md`

### DONE-20260521-013 商业级基础安全和数据底座推进

- 日期：2026-05-21
- 结果：已补商业级验收清单、PostgreSQL 核心迁移、API 角色鉴权骨架和小程序 API 客户端接入点；用户/骑手敏感操作不再只信任请求体里的裸 ID。
- 验收证据：
  - `cd services/api-go && go test ./...`
- 文件：
  - `docs/product/commercial-readiness-checklist.md`
  - `infra/db/migrations/0001_core.sql`
  - `scripts/check-architecture.mjs`
  - `services/api-go/internal/httpapi/auth.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `apps/user-wechat-miniprogram/utils/api.ts`
  - `apps/user-wechat-miniprogram/pages/shop/list/index.ts`
  - `apps/user-wechat-miniprogram/pages/shop/detail/index.ts`
  - `apps/user-wechat-miniprogram/pages/cart/index.ts`
  - `apps/user-wechat-miniprogram/pages/order/confirm/index.ts`

### DONE-20260521-014 用户端 BFF 代理层推进

- 日期：2026-05-21
- 结果：已把用户端核心接口接入 BFF 代理层，小程序默认访问 BFF，由 BFF 转发到核心 API 并保留 Authorization，符合端侧不直连内部 API 的架构约束。
- 验收证据：
  - `npm run test --workspace @infinitech/bff`
- 文件：
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `apps/user-wechat-miniprogram/utils/api.ts`

### DONE-20260521-015 用户订单查询链路推进

- 日期：2026-05-21
- 结果：已补用户订单列表和订单详情能力，订单详情包含事件流，为后续售后、评价、客服、配送追踪接入做基础。
- 验收证据：
  - `cd services/api-go && go test ./...`
  - `npm run test --workspace @infinitech/bff`
  - `npm run verify`
  - `npm run verify`
  - `npm run verify`
- 文件：
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `apps/user-wechat-miniprogram/app.json`
  - `apps/user-wechat-miniprogram/utils/api.ts`
  - `apps/user-wechat-miniprogram/pages/order/list/index.wxml`
  - `apps/user-wechat-miniprogram/pages/order/detail/index.wxml`

### DONE-20260521-016 余额支付密码链路推进

- 日期：2026-05-21
- 结果：已补余额支付密码设置与校验，余额支付请求必须携带正确 6 位数字密码；小程序增加余额支付密码设置页。
- 验收证据：
  - `cd services/api-go && go test ./...`
- 文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/server.mjs`
  - `apps/user-wechat-miniprogram/utils/api.ts`
  - `apps/user-wechat-miniprogram/pages/wallet/payment-password/index.wxml`

### DONE-20260521-017 微信登录、微信支付与仓储边界推进

- 日期：2026-05-21
- 结果：已补微信小程序登录接口、签名会话 token、微信支付预下单、微信支付回调验签/幂等确认、支付交易记录模型，以及核心 Repository 边界；内存 Store 仍用于测试，后续 PostgreSQL Store 按同一接口替换。
- 验收证据：
  - `cd services/api-go && go test ./...`
  - `npm run test --workspace @infinitech/bff`
  - `npm run verify:architecture`
- 文件：
  - `services/api-go/internal/httpapi/auth.go`
  - `services/api-go/internal/httpapi/wechat_pay.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `apps/user-wechat-miniprogram/utils/api.ts`
  - `infra/db/migrations/0002_auth_payment.sql`
  - `scripts/check-architecture.mjs`

### DONE-20260521-018 商户邀约注册与资质门槛推进

- 日期：2026-05-21
- 结果：已补管理员创建商户邀请、商户通过邀请注册、商户资料 token、营业执照/健康证上传和资质缺失检查；商户未补齐资质或未缴纳保证金时仍不可接单。
- 验收证据：
  - `cd services/api-go && go test ./...`
  - `npm run test --workspace @infinitech/bff`
  - `npm run verify`
- 文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/httpapi/auth.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/server.mjs`

### DONE-20260521-019 商户接单/出餐状态机推进

- 日期：2026-05-21
- 结果：已把外卖/买药类店铺订单从“支付成功直接进入骑手调度”调整为“商户待接单 -> 备餐中 -> 出餐后进入骑手调度”；骑手在商户出餐前不能抢单；商户订单列表、接单、出餐接口已接入 BFF；商户端 Flutter 已补经营概况和订单处理首版页面。
- 验收证据：
  - `cd services/api-go && go test ./...`
  - `npm run test --workspace @infinitech/bff`
  - `npm run verify`
- 文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `apps/merchant-flutter/lib/router.dart`
  - `apps/merchant-flutter/lib/api/merchant_api.dart`
  - `apps/merchant-flutter/lib/features/home/merchant_home_page.dart`
  - `apps/merchant-flutter/lib/features/orders/merchant_orders_page.dart`

### DONE-20260521-020 商户商品管理闭环推进

- 日期：2026-05-21
- 结果：已补商户商品列表、创建/编辑、上架/售罄/下架接口；商户只能管理自己店铺商品；用户端公开商品列表只展示上架商品，售罄/下架会从公开列表隐藏；BFF 已代理商户商品接口；商户端 Flutter 已新增商品管理页。
- 验收证据：
  - `cd services/api-go && go test ./...`
  - `npm run test --workspace @infinitech/bff`
  - `npm run verify`
- 文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `apps/merchant-flutter/lib/router.dart`
  - `apps/merchant-flutter/lib/api/merchant_api.dart`
  - `apps/merchant-flutter/lib/features/home/merchant_home_page.dart`
  - `apps/merchant-flutter/lib/features/products/merchant_products_page.dart`

### DONE-20260521-021 店铺接单门槛统一判断

- 日期：2026-05-21
- 结果：已给店铺模型补状态字段，并把“资质有效 + 保证金已缴 + 店铺 active”合并为统一接单门槛；用户加购、购物车结算、商户接单都会校验该门槛，避免资质过期或未缴保证金的店铺继续产生新履约订单。
- 验收证据：
  - `cd services/api-go && go test ./...`
  - `npm run verify`
- 文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router_test.go`

### DONE-20260521-022 团购下单发券与扫码核销闭环推进

- 日期：2026-05-21
- 结果：已补店铺团购套餐公开列表、团购下单、支付成功后自动发券、用户团购券列表、商户扫码核销团购券；核销必须使用 `qr_scan`，必须由本店商户执行，重复核销会被拒绝；BFF 已代理团购相关接口；用户小程序店铺页接入团购套餐和购买入口；商户端 Flutter 已新增团购核销页。
- 验收证据：
  - `cd services/api-go && go test ./...`
  - `npm run test --workspace @infinitech/bff`
  - `npm run verify`
- 文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/platform/contracts_test.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `apps/user-wechat-miniprogram/utils/api.ts`
  - `apps/user-wechat-miniprogram/pages/shop/detail/index.ts`
  - `apps/user-wechat-miniprogram/pages/shop/detail/index.wxml`
  - `apps/merchant-flutter/lib/router.dart`
  - `apps/merchant-flutter/lib/api/merchant_api.dart`
  - `apps/merchant-flutter/lib/features/home/merchant_home_page.dart`
  - `apps/merchant-flutter/lib/features/groupbuy/groupbuy_redeem_page.dart`

### DONE-20260521-023 骑手在线自动派单与拒单顺延推进

- 日期：2026-05-21
- 结果：已补骑手上线/离线接口、10 分钟后自动派单接口、骑手拒绝派单后顺延下一位在线骑手接口；自动派单会按在线、容量、保证金/免押、派单优先级、平均接单耗时和距离筛选候选骑手；10 分钟内触发自动派单会被拒绝。
- 验收证据：
  - `cd services/api-go && go test ./...`
  - `npm run test --workspace @infinitech/bff`
  - `npm run verify`
- 文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`

### DONE-20260521-024 站长手动派单与骑手端首屏推进

- 日期：2026-05-21
- 结果：已补站长查看站点骑手、站长查看待调度订单和站长手动派单接口；手动派单要求同站点在线骑手且保证金/免押状态可接单，跨站点骑手对站长不可见不可派；BFF 已代理站长调度接口；骑手端 Flutter 已新增抢单大厅首屏和站长工作台首屏。
- 验收证据：
  - `cd services/api-go && go test ./...`
  - `npm run test --workspace @infinitech/bff`
  - `npm run verify`
- 文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `apps/rider-flutter/lib/router.dart`
  - `apps/rider-flutter/lib/api/rider_api.dart`
  - `apps/rider-flutter/lib/features/home/rider_home_page.dart`
  - `apps/rider-flutter/lib/features/station/station_workspace_page.dart`

### DONE-20260521-025 站长任务时长和固定单量配置推进

- 日期：2026-05-21
- 结果：已补站长读取/保存每日任务时长和每日固定单量接口；配置按站点归属绑定站长，拒绝超过 24 小时的异常任务时长和异常固定单量；BFF 已代理任务配置接口；骑手端站长工作台可直接编辑并保存任务配置。
- 验收证据：
  - `cd services/api-go && go test ./...`
  - `npm run test --workspace @infinitech/bff`
  - `npm run verify`
- 文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `apps/rider-flutter/lib/api/rider_api.dart`
  - `apps/rider-flutter/lib/features/station/station_workspace_page.dart`

### DONE-20260521-026 站点骑手绩效等级快照推进

- 日期：2026-05-21
- 结果：已补站点骑手绩效等级快照接口，按站点返回骑手平均接单耗时、日均单量、完成率、分数、等级和派单优先级；评分以团队均值为参照并同步派单优先级；BFF 已代理该接口；骑手端站长工作台显示骑手等级和接单耗时。
- 验收证据：
  - `cd services/api-go && go test ./...`
  - `npm run test --workspace @infinitech/bff`
  - `npm run verify`
- 文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `apps/rider-flutter/lib/api/rider_api.dart`
  - `apps/rider-flutter/lib/features/station/station_workspace_page.dart`

### DONE-20260521-027 固定单量后免责拒派执行推进

- 日期：2026-05-21
- 结果：已把站长配置的每日固定单量接入骑手拒派决策；骑手当日完成单量达到站点固定单量后，拒绝派单会返回 `can_decline_without_penalty`、当日完成数、固定单量和 `after_fixed_order_count` 原因，同时继续顺延下一位在线骑手；骑手端拒派提示会显示本次免责。
- 验收证据：
  - `cd services/api-go && go test ./...`
  - `npm run test --workspace @infinitech/bff`
  - `npm run verify`
- 文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/bff/src/runtime.test.mjs`
  - `apps/rider-flutter/lib/features/home/rider_home_page.dart`

### DONE-20260521-028 骑手取货送达履约流转推进

- 日期：2026-05-21
- 结果：已补骑手取货和送达完成接口；订单由 `rider_assigned` 进入 `picked_up`，送达后进入 `completed`，重复送达会被拒绝；BFF 已代理骑手抢单/取货/送达路径；骑手端抢单大厅已增加确认取货和送达完成按钮。
- 验收证据：
  - `cd services/api-go && go test ./...`
  - `npm run test --workspace @infinitech/bff`
  - `npm run verify`
- 文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `apps/rider-flutter/lib/api/rider_api.dart`
  - `apps/rider-flutter/lib/features/home/rider_home_page.dart`

### DONE-20260521-029 商户与骑手保证金首批闭环推进

- 日期：2026-05-21
- 结果：已补保证金账户内存实现；骑手可查询/缴纳 50 元保证金、申请微信免押、提交退押申请，退押申请会记录最后完成订单时间和离职提交时间；商户可查询/缴纳 50 元保证金；保证金状态会同步到骑手/商户主体并继续参与接单门槛；BFF 已代理保证金接口；骑手端抢单大厅已增加保证金操作入口。
- 验收证据：
  - `cd services/api-go && go test ./...`
  - `npm run test --workspace @infinitech/bff`
  - `npm run verify`
- 文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `apps/rider-flutter/lib/api/rider_api.dart`
  - `apps/rider-flutter/lib/features/home/rider_home_page.dart`

### DONE-20260521-030 骑手与站长邀约注册闭环推进

- 日期：2026-05-21
- 结果：已补骑手/站长邀约注册内存实现和 HTTP 接口；站长只能创建本站骑手邀约，不能创建站长邀约；管理员可创建骑手或站长邀约；接受邀约后签发对应骑手端登录 token，已使用邀约不可复用；BFF 已代理管理员邀约、站长邀约和邀约注册接口；骑手端工具函数已预留接受邀约和站长创建骑手邀约调用。
- 验收证据：
  - `cd services/api-go && go test ./...`
  - `npm run test --workspace @infinitech/bff`
  - `npm run verify`
- 文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `apps/rider-flutter/lib/api/rider_api.dart`
  - `apps/rider-flutter/README.md`

### DONE-20260521-031 PostgreSQL 持久化第一阶段

- 日期：2026-05-21
- 结果：已补 PostgreSQL-backed Store 启动路径；`api-go` 设置 `DATABASE_URL` 时会使用 PostgreSQL 快照持久化当前模块化 Store，未设置时继续使用内存 Store；新增快照迁移和快照往返测试，确保支付索引、团购券索引和核心状态可恢复。该阶段用于服务重启不丢状态，后续仍需把订单、钱包、支付和派单拆为逐表强一致实现。
- 验收证据：
  - `cd services/api-go && go test ./...`
  - `npm run verify`
- 文件：
  - `services/api-go/cmd/api/main.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/platform/postgres_store_test.go`
  - `infra/db/migrations/0003_platform_store_snapshots.sql`
  - `services/api-go/go.mod`
  - `services/api-go/go.sum`

### DONE-20260521-032 真实微信 code2session 登录入口

- 日期：2026-05-21
- 结果：已把微信小程序登录拆成 provider resolver 和平台用户绑定两层；生产启动必须配置 `WECHAT_MINI_APP_ID`/`WECHAT_MINI_APP_SECRET` 调用微信 `code2session` 换取真实 `openid`，本地开发通过 `WECHAT_MINI_LOGIN_MODE=dev` 显式启用确定性开发 resolver；用户绑定优先使用 provider 返回的 `openid`，不再把前端 `code` 当作长期身份。`npm run dev:api-go` 已改为本地开发模式，README 和平台计划同步记录生产配置边界。
- 验收证据：
  - `cd services/api-go && go test ./...`
- 文件：
  - `services/api-go/cmd/api/main.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/api-go/internal/httpapi/wechat_login.go`
  - `services/api-go/internal/httpapi/wechat_login_test.go`
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `package.json`
  - `README.md`
  - `PLATFORM_MASTER_PLAN.md`

### DONE-20260521-033 auth session 持久化与 token 撤销

- 日期：2026-05-21
- 结果：已让登录/邀约注册签发的 HMAC token 携带服务端 session id，并在会话存储中记录 token hash、主体、设备、IP、User-Agent 和过期时间；`DATABASE_URL` 存在时 API 会使用 PostgreSQL `auth_sessions` 表校验未撤销 session，本地/测试默认使用内存 session；新增 `POST /api/auth/logout` 撤销当前 session，撤销后原 token 立即 401；生产启动默认关闭 `Authorization: Bearer role:subject_id` 开发 token，仅 `WECHAT_MINI_LOGIN_MODE=dev` 或 `ALLOW_DEV_BEARER_AUTH=true` 显式启用；BFF 已代理 logout。
- 验收证据：
  - `cd services/api-go && go test ./...`
  - `npm run test --workspace @infinitech/bff`
  - `npm run verify`
- 文件：
  - `services/api-go/cmd/api/main.go`
  - `services/api-go/internal/httpapi/auth.go`
  - `services/api-go/internal/httpapi/auth_session.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `PLATFORM_MASTER_PLAN.md`

### DONE-20260521-034 骑手/站长主体登录闭环

- 日期：2026-05-21
- 结果：已把骑手/站长邀约注册升级为设置账号密码，密码使用 bcrypt 哈希保存；新增 `POST /api/auth/rider/login`，骑手和站长可用 `account_id + password` 登录并签发 session-backed token；错误密码返回 `INVALID_CREDENTIALS` 401；PostgreSQL snapshot 已持久化 rider password hash，重启后主体登录不丢；BFF 已代理骑手登录；骑手端 Flutter API 工具已支持接受邀约带密码、账号密码登录，并按账号类型保存 rider/station manager token。
- 验收证据：
  - `cd services/api-go && go test ./...`
  - `npm run test --workspace @infinitech/bff`
- 文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `apps/rider-flutter/lib/api/rider_api.dart`
  - `services/api-go/go.mod`
  - `services/api-go/go.sum`

### DONE-20260522-035 商户/管理员主体登录闭环

- 日期：2026-05-22
- 结果：已把商户邀约注册升级为必须设置账号密码，密码使用 bcrypt 哈希保存；新增 `POST /api/auth/merchant/login`，商户可用 `account_id + password` 登录并签发 session-backed merchant token，错误密码返回 `INVALID_CREDENTIALS` 401；PostgreSQL snapshot 已持久化 merchant password hash，重启后商户主体登录不丢；新增 `POST /api/auth/admin/login`，管理员 bootstrap 密码登录默认关闭、无内置默认密码，仅同时配置 `ADMIN_BOOTSTRAP_ACCOUNT_ID` 和 `ADMIN_BOOTSTRAP_PASSWORD` 且密码长度 8-72 字节时启用；BFF 已代理商户登录和管理员登录；商户端 Flutter API 工具已支持邀请注册带密码、账号密码登录，并在匿名注册/登录时不携带开发 token。
- 验收证据：
  - `cd services/api-go && go test ./...`
  - `npm run test --workspace @infinitech/bff`
  - `npm run verify`
- 文件：
  - `services/api-go/cmd/api/main.go`
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/platform/postgres_store_test.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `apps/merchant-flutter/lib/api/merchant_api.dart`
  - `apps/merchant-flutter/README.md`
  - `README.md`
  - `PLATFORM_MASTER_PLAN.md`
  - `docs/product/commercial-readiness-checklist.md`

### DONE-20260522-036 派单事件持久化与站点区域匹配首版

- 日期：2026-05-22
- 结果：已补 `StationServiceArea`、店铺站点归属和派单决策站点字段；自动派单、拒单顺延、站长手动派单、骑手抢单都会写入 `DispatchEvent` 审计事件，并进入 PostgreSQL-backed Store 快照；新增 `GET /api/dispatch/orders/{orderID}/events`，管理员可查询全部订单派单审计，站长只能查询本站订单；站点订单视图、派单候选骑手、手动派单和事件查询均按站点过滤；拒单事件时间不会早于上一派单事件，审计流保持业务顺序；BFF 已代理派单事件查询并保留 Authorization 转发。
- 验收证据：
  - `cd services/api-go && go test ./...`
  - `npm run test --workspace @infinitech/bff`
  - `npm run verify`
- 文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/platform/postgres_store_test.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `PLATFORM_MASTER_PLAN.md`
  - `docs/product/commercial-readiness-checklist.md`

### DONE-20260522-037 派单确认超时自动转派首版

- 日期：2026-05-22
- 结果：已新增 `POST /api/dispatch/orders/{orderID}/timeout-reassign`，管理员和本站站长可触发派单确认超时转派；默认确认窗口为 60 秒，也支持请求体传入 `timeout_seconds`；未到确认窗口会拒绝转派；到点后系统写入 `dispatch.timeout` 审计事件，把超时骑手加入本单跳过列表，并按同站点在线可接单候选顺延下一位骑手；BFF 已代理该接口并保留 Authorization 转发；PostgreSQL-backed Store 会在转派后保存快照。
- 验收证据：
  - `cd services/api-go && go test ./...`
  - `npm run test --workspace @infinitech/bff`
  - `npm run verify`
- 文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `PLATFORM_MASTER_PLAN.md`
  - `docs/product/commercial-readiness-checklist.md`

### DONE-20260522-038 订单状态机补偿首版

- 日期：2026-05-22
- 结果：已新增订单状态机补偿领域能力和管理员接口 `POST /api/admin/orders/{orderID}/state/compensate`；补偿会基于支付交易、钱包流水、团购券、订单事件和派单审计事件推导订单应有状态，修复支付成功但订单仍卡在待支付、派单事件已落但订单状态/骑手丢失等漂移；修复时写入 `order.state.compensated` 审计事件，重复补偿保持幂等，已完成/退款/取消等终态不会被派单事件倒退；BFF 已代理该接口并保留 Authorization 转发；PostgreSQL-backed Store 会在补偿后保存快照。
- 验收证据：
  - `cd services/api-go && go test ./internal/platform ./internal/httpapi`
  - `npm run test --workspace @infinitech/bff`
  - `npm run verify`
- 文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `PLATFORM_MASTER_PLAN.md`
  - `docs/product/commercial-readiness-checklist.md`

### DONE-20260522-039 平台 outbox 事件首版

- 日期：2026-05-22
- 结果：已新增平台 outbox 事件首版：订单支付成功、商户接单/出餐、骑手取货/送达、团购核销、订单状态补偿和派单审计事件会同步写入可恢复 outbox；支持 pending 查询、发布确认、失败重试 backoff，失败事件到期后会重新出现在 pending relay 查询中；管理员接口 `GET /api/admin/outbox/events`、`POST /api/admin/outbox/events/{eventID}/published`、`POST /api/admin/outbox/events/{eventID}/failed` 已落地，BFF 已代理并保留 Authorization；PostgreSQL-backed Store 快照会保存 outbox 事件和幂等索引；新增 `platform_outbox_events` 迁移作为后续 Kafka/NATS relay 的归一化目标表。
- 验收证据：
  - `cd services/api-go && go test ./...`
  - `npm run test --workspace @infinitech/bff`
  - `npm run verify:architecture`
- 文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/platform/postgres_store_test.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `infra/db/migrations/0004_platform_outbox.sql`
  - `scripts/check-architecture.mjs`
  - `PLATFORM_MASTER_PLAN.md`
  - `docs/product/commercial-readiness-checklist.md`

### DONE-20260522-040 outbox relay worker 首版

- 日期：2026-05-22
- 结果：已新增 `@infinitech/outbox-relay-worker` 工作包，支持从管理员 outbox API 轮询 pending 事件、发布到可插拔 publisher、成功后标记 published、失败后标记 failed 并写入 retry backoff；默认覆盖 `order.paid`、`order.status_changed`、`order.completed`、`dispatch.assigned`、`dispatch.timeout`、`dispatch.status_changed` 六类关键主题；API client 会转发 `OUTBOX_RELAY_TOKEN` Authorization；系统架构图已补充 `outbox-relay-worker -> Kafka`，workspace 和架构检查已纳入该 worker。
- 验收证据：
  - `npm run test --workspace @infinitech/outbox-relay-worker`
  - `node --check services/outbox-relay-worker/src/index.mjs`
  - `node --check services/outbox-relay-worker/src/index.test.mjs`
- 文件：
  - `package.json`
  - `services/outbox-relay-worker/package.json`
  - `services/outbox-relay-worker/src/index.mjs`
  - `services/outbox-relay-worker/src/index.test.mjs`
  - `scripts/check-architecture.mjs`
  - `docs/architecture/system-architecture.md`
  - `PLATFORM_MASTER_PLAN.md`
  - `docs/product/commercial-readiness-checklist.md`

### DONE-20260522-041 outbox relay 可运行化与部署骨架

- 日期：2026-05-22
- 结果：已把 outbox relay worker 从一次性函数推进为可运行组件：新增长轮询 `runRelayLoop`、环境变量控制的主题/批量/间隔/backoff、可选 Kafka REST publisher；本地 Docker Compose 已加入 `outbox-relay-worker` workers profile，K8s base 已加入 2 副本 Deployment 并通过 Secret 引用 `outbox-relay-token`；架构检查会校验 Compose/K8s 中的 outbox relay 部署骨架。
- 验收证据：
  - `node --check services/outbox-relay-worker/src/index.mjs`
  - `node --check services/outbox-relay-worker/src/index.test.mjs`
  - `npm run test --workspace @infinitech/outbox-relay-worker`
  - `npm run verify:architecture`
- 文件：
  - `services/outbox-relay-worker/src/index.mjs`
  - `services/outbox-relay-worker/src/index.test.mjs`
  - `infra/docker/compose.yml`
  - `infra/k8s/base/app-stack.yaml`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `EXECUTION_LEDGER.md`
  - `PLATFORM_MASTER_PLAN.md`
  - `docs/product/commercial-readiness-checklist.md`

### DONE-20260522-042 outbox 积压观测首版

- 日期：2026-05-22
- 结果：已新增 outbox 积压统计能力：管理员可通过 `GET /api/admin/outbox/stats` 按 topic 查看 total/pending/failed/published、ready/blocked、oldest ready lag 和 per-topic 聚合；失败事件在 retry backoff 期间会作为 blocked 计入，backoff 到期后会作为 ready backlog 计入，发布确认后退出 ready backlog；BFF 已代理该管理员接口并保留 Authorization 与查询参数透传。
- 验收证据：
  - `node --check services/bff/src/server.mjs`
  - `node --check services/bff/src/runtime.test.mjs`
  - `cd services/api-go && go test ./...`
  - `npm run test --workspace @infinitech/bff`
  - `npm run verify`
- 文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `EXECUTION_LEDGER.md`
  - `PLATFORM_MASTER_PLAN.md`
  - `docs/product/commercial-readiness-checklist.md`

### DONE-20260522-043 outbox 手动恢复/重放首版

- 日期：2026-05-22
- 结果：已新增 outbox 单事件手动恢复/重放能力：管理员可通过 `POST /api/admin/outbox/events/{eventID}/replay` 将 failed 或等待 backoff 的 pending 事件立即拉回 pending-ready 队列；该动作保留 attempts 审计、清理 last_error、更新 available_at，已 published 事件会被拒绝重放，避免人为制造重复投递；BFF 已代理该接口，outbox relay worker API client 已提供 `replayEvent` 运维辅助方法。
- 验收证据：
  - `node --check services/bff/src/server.mjs`
  - `node --check services/bff/src/runtime.test.mjs`
  - `node --check services/outbox-relay-worker/src/index.mjs`
  - `node --check services/outbox-relay-worker/src/index.test.mjs`
  - `cd services/api-go && go test ./...`
  - `npm run test --workspace @infinitech/bff`
  - `npm run test --workspace @infinitech/outbox-relay-worker`
  - `npm run verify`
- 文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `services/outbox-relay-worker/src/index.mjs`
  - `services/outbox-relay-worker/src/index.test.mjs`
  - `EXECUTION_LEDGER.md`
  - `PLATFORM_MASTER_PLAN.md`
  - `docs/product/commercial-readiness-checklist.md`

### DONE-20260522-044 outbox 批量恢复/重放首版

- 日期：2026-05-22
- 结果：已新增 outbox 批量恢复/重放能力：管理员可通过 `POST /api/admin/outbox/events/replay` 按 topic 和 limit 将 failed/pending 且仍处于 backoff 阻塞期的事件批量拉回 pending-ready 队列；该动作保留 attempts 审计、清理 last_error、更新 available_at，跳过已经 ready 的事件和 published 事件，避免运维批处理造成重复投递；BFF 已代理该接口，outbox relay worker API client 已提供 `replayEvents` 运维辅助方法。
- 验收证据：
  - `node --check services/bff/src/server.mjs`
  - `node --check services/bff/src/runtime.test.mjs`
  - `node --check services/outbox-relay-worker/src/index.mjs`
  - `node --check services/outbox-relay-worker/src/index.test.mjs`
  - `cd services/api-go && go test ./...`
  - `npm run test --workspace @infinitech/bff`
  - `npm run test --workspace @infinitech/outbox-relay-worker`
  - `npm run verify`
- 文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `services/outbox-relay-worker/src/index.mjs`
  - `services/outbox-relay-worker/src/index.test.mjs`
  - `EXECUTION_LEDGER.md`
  - `PLATFORM_MASTER_PLAN.md`
  - `docs/product/commercial-readiness-checklist.md`

### DONE-20260522-045 outbox 死信隔离首版

- 日期：2026-05-22
- 结果：已新增 outbox 死信隔离能力：事件失败时可携带 `max_attempts`，达到上限后进入 `dead_letter` 状态并退出 pending relay 查询，避免毒消息无限回流拖垮队列；管理员可通过 `GET /api/admin/outbox/events?status=dead_letter` 查询死信，通过 `GET /api/admin/outbox/stats` 查看 dead_letter 计数，通过单事件 `POST /api/admin/outbox/events/{eventID}/replay` 人工解封并保留 attempts 审计；批量 replay 继续跳过 dead_letter，避免一键恢复误放毒消息；outbox relay worker 默认透传 `OUTBOX_RELAY_MAX_ATTEMPTS=10`，Compose/K8s 部署骨架和架构检查已纳入该护栏。
- 验收证据：
  - `node --check services/bff/src/runtime.test.mjs`
  - `node --check services/outbox-relay-worker/src/index.mjs`
  - `node --check services/outbox-relay-worker/src/index.test.mjs`
  - `node --check scripts/check-architecture.mjs`
  - `cd services/api-go && go test ./...`
  - `npm run test --workspace @infinitech/bff`
  - `npm run test --workspace @infinitech/outbox-relay-worker`
  - `npm run verify`
- 文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/runtime.test.mjs`
  - `services/outbox-relay-worker/src/index.mjs`
  - `services/outbox-relay-worker/src/index.test.mjs`
  - `infra/db/migrations/0004_platform_outbox.sql`
  - `infra/docker/compose.yml`
  - `infra/k8s/base/app-stack.yaml`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `EXECUTION_LEDGER.md`
  - `PLATFORM_MASTER_PLAN.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `docs/operations/capacity-and-dr.md`

### DONE-20260522-046 outbox relay 租约领取首版

- 日期：2026-05-22
- 结果：已新增 outbox relay claim/lease 防重复发布能力：管理员接口 `POST /api/admin/outbox/events/claim` 可按 topic/limit 原子领取 ready 的 pending/failed 事件并写入 `lease_owner`、`lease_expires_at`；活动租约会从 pending relay 查询和 ready backlog 中隐藏并在 stats 中单独计为 `leased`，租约过期后可由其他 relay 重新领取；published/failed/replay 会清理租约。outbox relay worker 已优先走 claim，再发布并 ack/fail，保留旧 pending 查询 fallback；Compose/K8s 已配置 `OUTBOX_RELAY_WORKER_ID` 和 `OUTBOX_RELAY_LEASE_SECONDS`，架构检查锁定 outbox lease 字段、索引和部署变量。
- 验收证据：
  - `node --check services/bff/src/runtime.test.mjs`
  - `node --check services/outbox-relay-worker/src/index.mjs`
  - `node --check services/outbox-relay-worker/src/index.test.mjs`
  - `node --check scripts/check-architecture.mjs`
  - `cd services/api-go && go test ./...`
  - `npm run test --workspace @infinitech/bff`
  - `npm run test --workspace @infinitech/outbox-relay-worker`
  - `npm run verify:architecture`
  - `npm run verify`
- 文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `services/outbox-relay-worker/src/index.mjs`
  - `services/outbox-relay-worker/src/index.test.mjs`
  - `infra/db/migrations/0004_platform_outbox.sql`
  - `infra/docker/compose.yml`
  - `infra/k8s/base/app-stack.yaml`
  - `scripts/check-architecture.mjs`
  - `EXECUTION_LEDGER.md`
  - `PLATFORM_MASTER_PLAN.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `docs/operations/capacity-and-dr.md`

### DONE-20260522-047 outbox relay 租约续租首版

- 日期：2026-05-22
- 结果：已新增 outbox relay lease heartbeat 能力：管理员接口 `POST /api/admin/outbox/events/{eventID}/lease/renew` 只允许当前活动 `lease_owner` 续租 pending/failed 事件，错误 owner、过期租约、已 published/dead-letter 事件都会冲突，避免慢 Kafka 发布或网络抖动期间租约过期被其他 relay 副本重复领取。BFF 已代理续租接口；outbox relay worker 在发布期间按 `OUTBOX_RELAY_LEASE_RENEW_INTERVAL_MS` 周期续租，发布结束后清理 timer，续租失败会被记录但不打断已在进行的 publish；Compose/K8s 默认配置 `OUTBOX_RELAY_LEASE_RENEW_INTERVAL_MS=30000`，架构检查纳入部署变量。
- 验收证据：
  - `node --check services/bff/src/server.mjs`
  - `node --check services/bff/src/runtime.test.mjs`
  - `node --check services/outbox-relay-worker/src/index.mjs`
  - `node --check services/outbox-relay-worker/src/index.test.mjs`
  - `node --check scripts/check-architecture.mjs`
  - `cd services/api-go && go test ./...`
  - `npm run test --workspace @infinitech/bff`
  - `npm run test --workspace @infinitech/outbox-relay-worker`
  - `npm run verify:architecture`
- 文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `services/outbox-relay-worker/src/index.mjs`
  - `services/outbox-relay-worker/src/index.test.mjs`
  - `infra/docker/compose.yml`
  - `infra/k8s/base/app-stack.yaml`
  - `scripts/check-architecture.mjs`
  - `EXECUTION_LEDGER.md`
  - `PLATFORM_MASTER_PLAN.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `docs/operations/capacity-and-dr.md`

### DONE-20260522-048 PostgreSQL outbox 规范化 relay 路径首版

- 日期：2026-05-22
- 结果：已把 `PostgresStore` 的 outbox relay 关键路径从单纯 JSON snapshot 兼容推进到 `platform_outbox_events` 规范化表：启动时确保 outbox 表/索引存在，并把 snapshot 里的 outbox 事件用 `ON CONFLICT (idempotency_key) DO NOTHING` 幂等补入规范化表；管理员 pending 查询、stats、claim、renew、published、failed、单事件 replay、批量 replay 均优先读写规范化表，并把结果回写内存镜像与 snapshot。claim 和批量 replay 使用 `FOR UPDATE SKIP LOCKED`，为多副本 relay 并发领取打基础；renew 明确要求当前 owner 且 `lease_expires_at > now`，继续保持错误 owner/过期租约冲突语义。
- 验收证据：
  - `node --check scripts/check-architecture.mjs`
  - `npm run verify:architecture`
  - `cd services/api-go && go test ./...`
- 文件：
  - `services/api-go/internal/platform/postgres_store.go`
  - `scripts/check-architecture.mjs`
  - `EXECUTION_LEDGER.md`
  - `PLATFORM_MASTER_PLAN.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `docs/operations/capacity-and-dr.md`

### DONE-20260522-049 outbox 租约健康观测首版

- 日期：2026-05-22
- 结果：已增强 `GET /api/admin/outbox/stats` 的 relay 租约健康观测：支持 `lease_expiring_within_seconds` 查询窗口，返回全局 `lease_expiring_soon`、`next_lease_expires_at`、`next_lease_expires_in_seconds`，并按 topic 与 `lease_owner` 聚合活动租约数量、临期数量和最早到期时间。内存 Store 与 PostgreSQL 规范化 outbox 路径共用同一套统计函数，BFF 透传查询参数和响应字段，架构检查锁定字段/路由/统计实现，便于后续 10 万并发压测和 DR 演练定位 relay 心跳中断、租约即将过期和 worker 倾斜。
- 验收证据：
  - `node --check scripts/check-architecture.mjs`
  - `node --check services/bff/src/runtime.test.mjs`
  - `cd services/api-go && go test ./...`
  - `npm run test --workspace @infinitech/bff`
  - `npm run verify:architecture`
- 文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/runtime.test.mjs`
  - `scripts/check-architecture.mjs`
  - `EXECUTION_LEDGER.md`
  - `PLATFORM_MASTER_PLAN.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `docs/operations/capacity-and-dr.md`

### DONE-20260522-050 消费端幂等落库首版

- 日期：2026-05-22
- 结果：已新增下游 worker 消费端幂等首版：`@infinitech/domain-core` 提供 `normalizeOutboxConsumerEvent`、`createConsumedEventLedger` 和 `createIdempotentConsumer`，按 `consumer_name + idempotency_key` 生成 `consumer_event_key` 并在重复 outbox 投递时返回 `duplicate`，避免 relay 重放、租约接管或 Kafka/NATS 至少一次投递导致业务处理重复执行。dispatch/payment/notification/integration/settlement 五个 worker 均暴露对应 `create...Consumer` 入口并覆盖重复投递只执行一次；PostgreSQL 迁移和 `PostgresStore` 启动建表新增 `platform_consumed_events`，后续真实 broker 消费可直接落库防重。
- 验收证据：
  - `node --check scripts/check-architecture.mjs`
  - `node --check packages/domain-core/src/index.mjs`
  - `node --check services/dispatch-worker/src/index.mjs`
  - `node --check services/payment-worker/src/index.mjs`
  - `node --check services/notification-worker/src/index.mjs`
  - `node --check services/integration-worker/src/index.mjs`
  - `node --check services/settlement-worker/src/index.mjs`
  - `cd services/api-go && go test ./...`
  - `npm run verify:architecture`
  - `npm run verify`
- 文件：
  - `packages/domain-core/src/index.mjs`
  - `packages/domain-core/src/index.test.mjs`
  - `services/dispatch-worker/src/index.mjs`
  - `services/dispatch-worker/src/index.test.mjs`
  - `services/payment-worker/src/index.mjs`
  - `services/payment-worker/src/index.test.mjs`
  - `services/notification-worker/src/index.mjs`
  - `services/notification-worker/src/index.test.mjs`
  - `services/integration-worker/src/index.mjs`
  - `services/integration-worker/src/index.test.mjs`
  - `services/settlement-worker/src/index.mjs`
  - `services/settlement-worker/src/index.test.mjs`
  - `infra/db/migrations/0004_platform_outbox.sql`
  - `services/api-go/internal/platform/postgres_store.go`
  - `scripts/check-architecture.mjs`
  - `EXECUTION_LEDGER.md`
  - `PLATFORM_MASTER_PLAN.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `docs/operations/capacity-and-dr.md`

### DONE-20260522-051 支付/钱包 PostgreSQL 规范化恢复首版

- 日期：2026-05-22
- 结果：已把 `PostgresStore` 的支付/钱包关键状态从单纯 JSON snapshot 兼容推进到规范化表同步与恢复首版：启动时确保 `app_users`、`merchant_accounts`、`shops`、`rider_accounts`、`orders`、`order_items`、`order_events`、`wallet_accounts`、`wallet_transactions`、`wallet_payment_passwords`、`payment_transactions` 等关键表存在；每次成功写入后把订单、订单项、订单事件、钱包账户、钱包流水、支付密码和微信支付交易幂等 upsert 到规范化表；启动时再从规范化表恢复订单、钱包、支付密码、钱包幂等索引、微信 `out_trade_no`/`transaction_id` 索引，避免重启后重复扣款、重复回调或支付密码丢失。新增恢复测试覆盖余额支付重复请求、支付密码继续可用、微信回调重复投递和订单事件恢复。
- 验收证据：
  - `cd services/api-go && go test -count=1 ./internal/platform`
  - `npm run verify:architecture`
  - `npm run verify`
- 文件：
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/platform/postgres_store_test.go`
  - `scripts/check-architecture.mjs`
  - `EXECUTION_LEDGER.md`
  - `PLATFORM_MASTER_PLAN.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `docs/operations/capacity-and-dr.md`

### DONE-20260522-052 派单审计事件 PostgreSQL 规范化恢复首版

- 日期：2026-05-22
- 结果：已把派单审计事件从 snapshot-only 恢复推进到 PostgreSQL 规范化表同步与恢复首版：迁移和 `PostgresStore` 启动路径会确保 `dispatch_events` 表、按订单/站点时间索引和 `(idempotency_key, type)` 复合唯一索引存在；每次成功写入后把自动派单、拒单、超时、抢单和手动派单审计事件 upsert 到规范化表；启动时再从表恢复 `dispatchEvents`、拒单/超时骑手跳过索引和 `nextDispatchEventID`，保证站长审计、订单状态补偿和后续派单事件 ID 在灾备恢复后继续可用。新增恢复测试覆盖站长查询派单审计、基于表恢复事件修复订单骑手漂移、拒单骑手索引恢复和新派单事件 ID 不碰撞。
- 验收证据：
  - `cd services/api-go && go test -count=1 ./internal/platform -run TestDispatchEventTableRestoreRebuildsAuditAndCompensationIndexes -v`
  - `cd services/api-go && go test -count=1 ./internal/platform`
  - `npm run verify:architecture`
  - `npm run verify`
- 文件：
  - `infra/db/migrations/0001_core.sql`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/platform/postgres_store_test.go`
  - `scripts/check-architecture.mjs`
  - `EXECUTION_LEDGER.md`
  - `PLATFORM_MASTER_PLAN.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `docs/operations/capacity-and-dr.md`

### DONE-20260522-053 余额支付 PostgreSQL 事务扣减首版

- 日期：2026-05-22
- 结果：已把 `PostgresStore.PayOrderWithBalance` 从“内存扣减后同步规范化表”推进到 PostgreSQL 事务扣减首版：余额支付会在数据库事务中按幂等键获取 `pg_advisory_xact_lock`，再用 `FOR UPDATE` 锁定 `orders`、`wallet_accounts` 和已有 `wallet_transactions`，完成支付密码校验、订单状态校验、余额检查、钱包扣减、订单支付成功事件和钱包流水写入；`wallet_transactions.idempotency_key` 继续作为幂等唯一键，同一幂等键跨订单/跨用户重放会被拒绝。提交后从规范化表刷新内存镜像，并补齐团购发券与 `order.paid` outbox，保证灾备恢复、relay 和端侧返回仍保持一致。
- 验收证据：
  - `cd services/api-go && go test -count=1 ./internal/platform -run 'TestSQLBalancePayment(SideEffectsIssueGroupbuyVoucherAndOutbox|IdempotencyMatcherRejectsCrossOrderReplay)' -v`
  - `cd services/api-go && go test -count=1 ./internal/platform`
  - `npm run verify:architecture`
  - `npm run verify`
- 文件：
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/platform/postgres_store_test.go`
  - `scripts/check-architecture.mjs`
  - `EXECUTION_LEDGER.md`
  - `PLATFORM_MASTER_PLAN.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `docs/operations/capacity-and-dr.md`

### DONE-20260522-054 商家订单流转 PostgreSQL 事务化首版

- 日期：2026-05-22
- 结果：已把 `PostgresStore.MerchantAcceptOrder` 和 `PostgresStore.MerchantMarkOrderReady` 从内存状态变更后再同步推进到 PostgreSQL 事务化首版：商户接单/出餐会在数据库事务中锁定 `orders` 行，校验订单归属商户、当前状态、门店启用状态和商户保证金状态，原子更新订单状态并写入 `order_events`；事务提交后从规范化表刷新内存镜像、补齐 `order.status_changed` outbox 并保存 snapshot，保证商户履约状态、事件审计、relay 和灾备恢复口径一致。架构检查已加入防回退断言，避免商家主链路重新退回 snapshot-first。
- 验收证据：
  - `cd services/api-go && go test -count=1 ./internal/platform -run TestSQLMerchantOrderEventSideEffectsEnqueueStatusOutbox -v`
  - `cd services/api-go && go test ./...`
  - `npm run verify:architecture`
  - `npm run verify`
- 文件：
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/platform/postgres_store_test.go`
  - `scripts/check-architecture.mjs`
  - `EXECUTION_LEDGER.md`
  - `PLATFORM_MASTER_PLAN.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `docs/operations/capacity-and-dr.md`

### DONE-20260522-055 订单创建 PostgreSQL 事务化首版

- 日期：2026-05-22
- 结果：已把 `PostgresStore.CreateOrder` 从内存创建后再同步推进到 PostgreSQL 事务化首版：订单创建会在数据库事务中写入 `orders` 和初始 `order_events`，用 `platform_sequences` 的 `orders` 序列行级锁生成 `ord_N`，并在序列缺失或灾备恢复后根据现有 `orders` 最大编号自动追平，避免恢复后订单号碰撞。提交后从规范化表刷新内存镜像并保存 snapshot，`order.created` 保持本地审计事件、不产生 outbox；架构检查已加入防回退断言，避免订单创建主链路重新退回 snapshot-first。
- 验收证据：
  - `cd services/api-go && go test -count=1 ./internal/platform -run TestSQLCreateOrderSideEffectsRestoreMirror -v`
  - `cd services/api-go && go test ./...`
  - `npm run verify:architecture`
  - `npm run verify`
- 文件：
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/platform/postgres_store_test.go`
  - `infra/db/migrations/0001_core.sql`
  - `scripts/check-architecture.mjs`
  - `EXECUTION_LEDGER.md`
  - `PLATFORM_MASTER_PLAN.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `docs/operations/capacity-and-dr.md`

### DONE-20260522-056 商户资质员工资料端到端首版

- 日期：2026-05-22
- 结果：已把商户员工健康证和补充资料从后端模型推进到商户端首批闭环：Store/HTTP 测试覆盖员工资料读取、保存、健康证失效日期校验、跨商户店铺写入隐藏，以及补充资料读取/提交和跨商户写入隐藏；BFF 已代理 `merchant/me`、资质、员工、补充资料路径并覆盖 Authorization 转发；商户端 Flutter 新增“资质资料”页面，可查看接单门槛、资质缺失、保证金状态、员工健康证和补充资料，并提交营业执照、健康证、员工信息和门头/后厨等资料。
- 验收证据：
  - `cd services/api-go && go test ./internal/platform ./internal/httpapi`
  - `npm run test --workspace @infinitech/bff`
  - `npm run verify:architecture`
- 文件：
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/runtime.test.mjs`
  - `apps/merchant-flutter/lib/api/merchant_api.dart`
  - `apps/merchant-flutter/lib/router.dart`
  - `apps/merchant-flutter/lib/features/home/merchant_home_page.dart`
  - `apps/merchant-flutter/lib/features/compliance/merchant_compliance_page.dart`
  - `apps/merchant-flutter/README.md`
  - `EXECUTION_LEDGER.md`
  - `PLATFORM_MASTER_PLAN.md`
  - `docs/product/commercial-readiness-checklist.md`

### DONE-20260522-057 购物车结算 PostgreSQL 事务化首版

- 日期：2026-05-22
- 结果：已把 `PostgresStore.CheckoutCart` 从 snapshot-first 推进到 PostgreSQL 事务化首版：运行时会确保并同步 `merchant_qualifications`、`merchant_products` 和 `cart_items` 规范化表；结算在数据库事务中按用户/店铺获取 advisory lock，锁定店铺/商户、地址、购物车行和商品行，校验店铺可接单、商户保证金、商户资质有效、地址完整、商品状态和库存，原子写入 `orders`、`order_items`、`order_events` 并删除对应购物车行。事务提交后刷新订单镜像、清空内存购物车并保存 snapshot，架构检查防止主链路退回内存先写。
- 验收证据：
  - `npm run verify`
- 文件：
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/platform/postgres_store_test.go`
  - `scripts/check-architecture.mjs`
  - `EXECUTION_LEDGER.md`
  - `PLATFORM_MASTER_PLAN.md`
  - `docs/product/commercial-readiness-checklist.md`

### DONE-20260522-058 退款策略与余额退款核心闭环首版

- 日期：2026-05-22
- 结果：已新增后台退款策略读取/保存和管理员订单退款接口；平台默认 `balance_first`，可配置 `original_route_first`。余额退款会用幂等键创建退款交易、写入正向钱包退款流水、恢复用户平台余额、更新订单为 `refunded` 并产生 `order.refunded` outbox；原路返回策略首版会把订单置为 `refund_pending`，创建 `pending_original_route` 退款交易并产生 `payment.refund.requested` outbox，等待后续 payment-worker 接入微信退款。重复退款请求按退款幂等键返回同一退款结果，不重复入账。HTTP/BFF 已代理 `GET/PUT /api/admin/refund-settings` 和 `POST /api/orders/{orderID}/refund`，仅管理员可直接退款。
- 验收证据：
  - `cd services/api-go && go test ./internal/platform ./internal/httpapi`
  - `node --test services/bff/src/runtime.test.mjs scripts/check-architecture.mjs`
- 文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `scripts/check-architecture.mjs`
  - `EXECUTION_LEDGER.md`
  - `PLATFORM_MASTER_PLAN.md`
  - `docs/product/commercial-readiness-checklist.md`

### DONE-20260522-059 payment-worker 原路退款事件规范化首版

- 日期：2026-05-22
- 结果：`payment-worker` 默认消费者已按 outbox topic 分发处理；`payment.refund.requested` 不再被当成微信支付回调，而是规范化为 `refund_requested` 任务载荷，包含订单、用户、金额、目的地、pending 状态和稳定幂等键。重复 outbox 投递仍由 consumed-event ledger 防重。
- 验收证据：
  - `npm run test --workspace @infinitech/payment-worker`
- 文件：
  - `services/payment-worker/src/index.mjs`
  - `services/payment-worker/src/index.test.mjs`
  - `EXECUTION_LEDGER.md`
  - `PLATFORM_MASTER_PLAN.md`
  - `docs/product/commercial-readiness-checklist.md`

### DONE-20260522-060 退款 PostgreSQL 事务化首版

- 日期：2026-05-22
- 结果：已新增 `refund_settings` 和 `refund_transactions` 规范化表及迁移；`PostgresStore` 启动和写入会同步/恢复退款策略与退款交易，并重建退款幂等索引。`PostgresStore.RefundOrder` 已从内存先写推进到数据库事务路径：按退款幂等键获取 advisory lock，锁定退款设置、订单、退款交易、钱包账户和钱包流水；余额退款在同一事务内写入退款交易、钱包正向退款流水、钱包余额版本、订单 `refunded` 状态和订单事件；原路返回在同一事务内写入退款交易、订单 `refund_pending` 状态和订单事件。提交后刷新规范化表镜像、补齐 `order.refunded` 或 `payment.refund.requested` outbox 并保存 snapshot，架构检查防止回退到 `Store.RefundOrder`。
- 验收证据：
  - `cd services/api-go && go test ./internal/platform ./internal/httpapi`
  - `node --test scripts/check-architecture.mjs`
- 文件：
  - `infra/db/migrations/0002_auth_payment.sql`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/platform/postgres_store_test.go`
  - `scripts/check-architecture.mjs`
  - `EXECUTION_LEDGER.md`
  - `PLATFORM_MASTER_PLAN.md`
  - `docs/product/commercial-readiness-checklist.md`

### DONE-20260522-061 售后申请与审核核心闭环首版

- 日期：2026-05-22
- 结果：已新增用户售后申请、用户/商户/管理员售后列表和商户/管理员审核接口。售后申请必须关联订单、用户、原因、申请金额和证据附件；首版仅允许整单售后退款，避免部分退款在商品级资金账本未完成前造成账实不一致。审核支持通过、驳回、转平台审核；商户只能处理自己店铺订单，用户不可审核。审核通过后会复用退款幂等链路，默认退回平台余额，写入退款交易、钱包退款流水、订单退款状态和售后 outbox 事件；重复审核已完成售后会被拒绝。
- 验收证据：
  - `cd services/api-go && go test ./internal/platform ./internal/httpapi`
  - `npm run test --workspace @infinitech/bff`
  - `npm run verify:architecture`
- 当前边界：售后请求暂存于 Store snapshot，尚未落 `order_after_sales` PostgreSQL 规范表；部分退款、仲裁处理日志、客服介入和售后证据附件对象存储仍待后续推进。
- 文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `scripts/check-architecture.mjs`
  - `EXECUTION_LEDGER.md`
  - `PLATFORM_MASTER_PLAN.md`
  - `docs/product/commercial-readiness-checklist.md`

### DONE-20260522-062 售后 PostgreSQL 规范化恢复首版

- 日期：2026-05-22
- 结果：已新增 `order_after_sales` 规范表及索引，并把售后申请接入 `PostgresStore` 支付域同步/恢复路径。运行时建表和迁移均包含售后表；snapshot 同步时会清洗证据附件并 upsert 到规范表；启动恢复时会从规范表加载售后申请、重建售后内存索引和 `nextAfterSalesID`，避免售后数据只依赖 snapshot。
- 验收证据：
  - `cd services/api-go && go test ./internal/platform ./internal/httpapi`
  - `npm run verify:architecture`
- 当前边界：售后审核动作本身仍是 Store 级事务后同步到 PostgreSQL，后续还要推进部分退款、仲裁处理日志、客服介入和售后审核 SQL 事务化。
- 文件：
  - `infra/db/migrations/0002_auth_payment.sql`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/platform/postgres_store_test.go`
  - `scripts/check-architecture.mjs`
  - `EXECUTION_LEDGER.md`
  - `PLATFORM_MASTER_PLAN.md`
  - `docs/product/commercial-readiness-checklist.md`

### DONE-20260522-063 售后审核 PostgreSQL 事务化首版

- 日期：2026-05-22
- 结果：`PostgresStore.ReviewAfterSales` 已从 Store 级审核后同步推进到数据库事务路径。审核会在同一事务内锁定 `order_after_sales`、`orders`、退款幂等键、退款配置和钱包账户；商户只能审核自有店铺售后，管理员可审核全部；驳回、转平台审核、审核通过均写入售后表和订单事件。审核通过会在同一事务内写入退款交易、钱包退款流水、钱包余额版本、订单退款状态和售后审核结果，提交后刷新规范表镜像、补齐 `order.refunded`/`payment.refund.requested` 与 `order.after_sales` outbox 并保存 snapshot。架构检查已防止回退到 `Store.ReviewAfterSales`。
- 验收证据：
  - `cd services/api-go && go test ./internal/platform ./internal/httpapi`
  - `npm run verify:architecture`
- 当前边界：首版仍只支持整单售后退款；部分退款、仲裁处理日志、客服介入和证据附件对象存储待后续推进。
- 文件：
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/platform/postgres_store_test.go`
  - `scripts/check-architecture.mjs`
  - `EXECUTION_LEDGER.md`
  - `PLATFORM_MASTER_PLAN.md`
  - `docs/product/commercial-readiness-checklist.md`

### DONE-20260522-064 部分退款资金账本首版

- 日期：2026-05-22
- 结果：退款与售后审核已支持部分退款。核心规则改为按订单累计 `success` 和 `pending_original_route` 退款金额，余额退款和原路退款申请都不能超过订单实付金额；部分退款会保留订单原业务状态，累计退满后才进入 `refunded` 或 `refund_pending`。售后申请可自动识别 `partial_refund`，已完成的部分售后可继续申请剩余金额，重复活跃售后仍会被拦截；退款 outbox 会把本次退款金额放入 `amount_fen`，并把订单原金额放入 `order_amount_fen`，避免原路退款 worker 误按整单退款。
- 验收证据：
  - `cd services/api-go && go test ./internal/platform`
  - `node --test packages/contracts/src/index.test.mjs`
- 当前边界：部分退款已进入内存 Store 和 PostgreSQL 事务路径；下一步继续补仲裁处理日志、客服介入、对象存储证据附件，以及 HTTP/BFF 端对“剩余可退金额”的展示字段。
- 文件：
  - `infra/db/migrations/0002_auth_payment.sql`
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/platform/postgres_store_test.go`
  - `packages/contracts/src/index.test.mjs`
  - `scripts/check-architecture.mjs`
  - `EXECUTION_LEDGER.md`
  - `PLATFORM_MASTER_PLAN.md`
  - `docs/product/commercial-readiness-checklist.md`

### DONE-20260522-065 售后仲裁与客服介入处理日志首版

- 日期：2026-05-22
- 结果：售后已新增独立处理日志账本。用户、商户、管理员可按权限查看售后时间线；用户可补充证据和申请客服介入，客服介入会把售后转入平台审核；商户可公开回复处理情况；管理员可写公开仲裁记录或内部备注。售后创建、审核通过、驳回、升级都会写入时间线；`order_after_sales_events` 已进入迁移和运行时规范表，并接入 PostgreSQL 同步/恢复。
- 验收证据：
  - `cd services/api-go && go test ./internal/platform`
  - `cd services/api-go && go test ./internal/httpapi`
  - `npm run verify:architecture`
  - `node --test packages/contracts/src/index.test.mjs`
- 当前边界：处理日志已具备权限、时间线和 SQL 恢复；下一步继续补对象存储证据附件签名上传、客服工作台视图，以及 HTTP/BFF 端“剩余可退金额/已退金额”展示字段。
- 文件：
  - `infra/db/migrations/0002_auth_payment.sql`
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/platform/postgres_store_test.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `packages/contracts/src/index.mjs`
  - `packages/contracts/src/index.test.mjs`
  - `scripts/check-architecture.mjs`
  - `EXECUTION_LEDGER.md`
  - `PLATFORM_MASTER_PLAN.md`
  - `docs/product/commercial-readiness-checklist.md`

### DONE-20260522-066 售后可退金额与证据上传票据首版

- 日期：2026-05-22
- 结果：售后申请返回值已补齐订单原金额、已退款金额和剩余可退金额，用户端/商户端/管理端不再需要自行推断退款窗口。新增 `POST /api/after-sales/{requestID}/evidence/upload-ticket`，按售后权限发放 15 分钟有效的对象存储上传票据，限制证据附件最大 10MB，并只允许图片和 PDF 类型；返回对象 key、上传 URL、公开 URL、PUT 方法和必需请求头，为后续 MinIO/Vault 签名实现留好接口边界。BFF 已代理该接口。
- 验收证据：
  - `cd services/api-go && go test ./internal/platform`
  - `cd services/api-go && go test ./internal/httpapi`
  - `npm run test --workspace @infinitech/bff`
  - `npm run verify:architecture`
  - `node --test packages/contracts/src/index.test.mjs`
- 当前边界：上传票据当前是平台内签名骨架，尚未接真实 MinIO SDK、STS/Vault 密钥和回调校验；下一步需要把上传票据接入真实对象存储，并在售后事件里绑定已上传对象。
- 文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `scripts/check-architecture.mjs`
  - `EXECUTION_LEDGER.md`
  - `PLATFORM_MASTER_PLAN.md`
  - `docs/product/commercial-readiness-checklist.md`

### DONE-20260522-067 售后证据确认与附件元数据首版

- 日期：2026-05-22
- 结果：售后证据从“拿到上传票据”推进到“上传完成后可确认绑定”。新增 `AfterSalesEvidence` 附件元数据，支持按售后权限确认对象 key、记录文件名/类型/大小/校验摘要/上传人/确认时间，并把公开 URL 回填到售后申请、写入 `evidence_uploaded` 售后时间线和 `order.after_sales.evidence_uploaded` outbox。新增 `GET /api/after-sales/{requestID}/evidence` 与 `POST /api/after-sales/{requestID}/evidence/confirm`，BFF 已代理；PostgreSQL 新增 `order_after_sales_evidence` 规范表，启动恢复和 snapshot 均可恢复附件记录。
- 验收证据：
  - `cd services/api-go && go test ./internal/platform`
  - `cd services/api-go && go test ./internal/httpapi`
  - `npm run test --workspace @infinitech/bff`
  - `npm run test --workspace @infinitech/domain-core`
  - `npm run verify:architecture`
  - `node --test packages/contracts/src/index.test.mjs`
- 当前边界：当前确认接口会校验对象 key 归属、类型和大小，并形成业务附件记录；仍未接真实 MinIO SDK、STS/Vault 临时凭证、对象存在性 HEAD 校验、对象存储回调验签和病毒/涉敏扫描。
- 文件：
  - `infra/db/migrations/0002_auth_payment.sql`
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/platform/postgres_store_test.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `packages/contracts/src/index.mjs`
  - `packages/contracts/src/index.test.mjs`
  - `packages/domain-core/src/index.mjs`
  - `packages/domain-core/src/index.test.mjs`
  - `scripts/check-architecture.mjs`
  - `EXECUTION_LEDGER.md`
  - `PLATFORM_MASTER_PLAN.md`
  - `docs/product/commercial-readiness-checklist.md`

### DONE-20260522-068 对象存储上传签名配置化首版

- 日期：2026-05-22
- 结果：售后证据上传票据从 Store 内硬编码地址推进为可配置对象存储签名边界。新增 `ObjectStorageConfig`、MinIO provider 标识、HMAC-SHA256 上传签名、bucket/CDN/上传端点/TTL/最大上传大小配置，返回票据包含 provider、bucket、对象 key、上传 URL、公开 URL、必需 headers 和签名。API 启动时读取 `OBJECT_STORAGE_PROVIDER`、`OBJECT_STORAGE_BUCKET`、`OBJECT_STORAGE_UPLOAD_BASE_URL`、`OBJECT_STORAGE_PUBLIC_BASE_URL`、`OBJECT_STORAGE_SIGNING_SECRET`、`OBJECT_STORAGE_TICKET_TTL_SECONDS`、`OBJECT_STORAGE_MAX_UPLOAD_BYTES`；本地可空密钥跑通，生产必须配置签名密钥。售后确认附件会使用同一对象存储配置生成公开 URL，避免后续接真实 MinIO/CDN 时改业务逻辑。
- 验收证据：
  - `cd services/api-go && go test ./internal/platform`
  - `cd services/api-go && go test ./internal/httpapi`
  - `npm run verify:architecture`
- 当前边界：当前已具备可配置签名票据和 HMAC 边界，但仍未调用真实 MinIO SDK/STS/Vault 临时凭证，也未做对象存在性 HEAD 校验、上传回调验签、病毒/涉敏扫描和私有 bucket 访问策略自动化。
- 文件：
  - `services/api-go/cmd/api/main.go`
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/object_storage.go`
  - `services/api-go/internal/platform/object_storage_test.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `EXECUTION_LEDGER.md`
  - `PLATFORM_MASTER_PLAN.md`
  - `docs/product/commercial-readiness-checklist.md`

### DONE-20260522-069 售后上传票据账本与确认防伪首版

- 日期：2026-05-22
- 结果：售后证据上传从“返回签名 URL”推进到“签发票据可追溯”。新增 `AfterSalesEvidenceUploadTicket` 账本，记录 ticket、provider、bucket、object key、公开 URL、文件类型/大小、上传人、状态、过期时间和确认时间；`ObjectUploadTicket` 返回 `ticket_id`；确认附件时必须匹配未过期、同售后申请、同用户/角色、同对象 key、同 content-type 和同大小的已签发票据。确认后票据状态变为 `confirmed`，重复确认同对象仍保持幂等。PostgreSQL 新增 `order_after_sales_evidence_upload_tickets` 规范表，snapshot/SQL 同步和恢复均可重建票据账本。
- 验收证据：
  - `cd services/api-go && go test ./internal/platform`
  - `cd services/api-go && go test ./internal/httpapi`
  - `npm run verify:architecture`
- 当前边界：票据账本已防止同 request 前缀伪造对象 key 直接绑定售后附件；下一步仍需接真实 MinIO SDK/STS/Vault、对象 HEAD 校验、上传回调验签、内容扫描和私有 bucket 策略自动化。
- 文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/platform/postgres_store_test.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `infra/db/migrations/0002_auth_payment.sql`
  - `scripts/check-architecture.mjs`
  - `EXECUTION_LEDGER.md`
  - `PLATFORM_MASTER_PLAN.md`
  - `docs/product/commercial-readiness-checklist.md`

### DONE-20260522-070 售后对象存在性 HEAD 校验开关首版

- 日期：2026-05-22
- 结果：售后证据确认链路新增可配置对象存在性校验。`ObjectStorageConfig` 支持 `HeadBaseURL`、`RequireHeadVerification` 和 `HeadTimeout`；API 启动时读取 `OBJECT_STORAGE_HEAD_BASE_URL`、`OBJECT_STORAGE_REQUIRE_HEAD_VERIFICATION`、`OBJECT_STORAGE_HEAD_TIMEOUT_SECONDS`。生产开启后，确认附件会先匹配上传票据，再对对象存储/CDN 发起 HEAD 请求，校验对象存在、Content-Length 与票据大小一致、Content-Type 与票据类型一致，校验通过后才写入售后附件、事件和订单 outbox。确认流程拆成准备校验和提交确认两段，避免网络 I/O 长时间占用 Store 锁。
- 验收证据：
  - `cd services/api-go && go test ./internal/platform`
  - `cd services/api-go && go test ./internal/httpapi`
  - `npm run verify:architecture`
- 当前边界：HEAD 校验已可通过配置打开，适合接 MinIO/CDN 对象存在性门禁；仍未接真实 MinIO SDK/STS/Vault 临时凭证、上传回调验签、病毒/涉敏扫描和私有 bucket 策略自动化。
- 文件：
  - `services/api-go/cmd/api/main.go`
  - `services/api-go/internal/platform/object_storage.go`
  - `services/api-go/internal/platform/object_storage_test.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `EXECUTION_LEDGER.md`
  - `PLATFORM_MASTER_PLAN.md`
  - `docs/product/commercial-readiness-checklist.md`

### DONE-20260522-071 售后上传回调验签与扫描门禁首版

- 日期：2026-05-22
- 结果：售后证据从“可选 HEAD 校验”推进到“上传回调 + 内容扫描门禁”。新增 `ObjectStorageUploadCallbackRequest`、`ObjectStorageScanResultRequest`，以及内部接口 `POST /api/object-storage/upload-callback`、`POST /api/object-storage/scan-result`。`ObjectStorageConfig` 新增 `CallbackSigningSecret`、`RequireUploadCallbackForConfirm`、`RequireScanApprovalForConfirm`，API 启动时读取 `OBJECT_STORAGE_CALLBACK_SIGNING_SECRET`、`OBJECT_STORAGE_REQUIRE_UPLOAD_CALLBACK`、`OBJECT_STORAGE_REQUIRE_SCAN_APPROVAL`。上传回调使用 HMAC 校验 ticket/object/type/size/content_sha/uploaded_at，票据状态从 `issued` 变为 `uploaded`；扫描结果同样验签，记录 `pending/passed/rejected`、扫描结果、扫描器和检查时间；生产开启回调和扫描门禁后，附件确认必须等上传回调和扫描通过后才会写入售后附件、时间线与订单 outbox。
- 验收证据：
  - `cd services/api-go && go test ./internal/platform`
  - `cd services/api-go && go test ./internal/httpapi`
  - `npm run verify:architecture`
- 当前边界：已具备上传回调验签、扫描状态账本、确认前门禁和 PostgreSQL 恢复；真实 MinIO SDK/STS/Vault 临时凭证、扫描 worker/ClamAV 服务、私有 bucket 策略自动化和对象生命周期清理仍待推进。
- 文件：
  - `services/api-go/cmd/api/main.go`
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/object_storage.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/platform/postgres_store_test.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `infra/db/migrations/0002_auth_payment.sql`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `EXECUTION_LEDGER.md`
  - `PLATFORM_MASTER_PLAN.md`
  - `docs/product/commercial-readiness-checklist.md`

### DONE-20260522-072 对象扫描 worker 首版

- 日期：2026-05-22
- 结果：新增独立 `@infinitech/object-scan-worker` workspace，负责消费 `object.uploaded` 和 `after_sales.evidence.object_uploaded` 事件，规范化对象上传信息，生成与 API 一致的上传回调 HMAC 签名，回调 `POST /api/object-storage/upload-callback`，再把扫描结果规范化为 `passed/rejected/pending` 并回调 `POST /api/object-storage/scan-result`。worker 复用 `createIdempotentConsumer`，同一 outbox/对象事件重复投递只扫描和回调一次；已补 Docker Compose 和 Kubernetes 部署骨架，配置 `OBJECT_SCAN_TOPICS`、`OBJECT_STORAGE_CALLBACK_SIGNING_SECRET`、`OBJECT_SCAN_WORKER_TOKEN`、`CLAMAV_HOST`、`CLAMAV_PORT`。
- 验收证据：
  - `npm run test --workspace @infinitech/object-scan-worker`
  - `npm run verify:architecture`
- 当前边界：worker 已具备事件规范化、幂等消费、API 回调和签名生成；真实 ClamAV 流式扫描、对象下载/预签下载、隔离 bucket、扫描失败重试/告警和对象生命周期清理仍待推进。
- 文件：
  - `package.json`
  - `services/object-scan-worker/package.json`
  - `services/object-scan-worker/src/index.mjs`
  - `services/object-scan-worker/src/index.test.mjs`
  - `infra/docker/compose.yml`
  - `infra/k8s/base/app-stack.yaml`
  - `scripts/check-architecture.mjs`
  - `EXECUTION_LEDGER.md`
  - `PLATFORM_MASTER_PLAN.md`
  - `docs/product/commercial-readiness-checklist.md`

### DONE-20260522-073 对象扫描 worker ClamAV 适配与下载首版

- 日期：2026-05-22
- 结果：对象扫描 worker 从“外部注入扫描结果”推进到“可真实下载对象并调用 ClamAV”。上传事件现在可携带 `bucket` 和 `object_url`，scanner 会优先使用签名下载 URL，否则通过 `OBJECT_STORAGE_DOWNLOAD_BASE_URL` 拼接 bucket/object key；下载阶段限制 `OBJECT_SCAN_MAX_BYTES`，并支持下载超时。新增 ClamAV INSTREAM 协议适配，按 `OBJECT_SCAN_CLAMAV_CHUNK_BYTES` 分块写入 socket，读取 `OK/FOUND` 结果后映射为 `passed/rejected` 并继续走原有 API 回调与幂等消费账本。Compose/K8s 已补 `OBJECT_STORAGE_DOWNLOAD_BASE_URL`、扫描大小/超时/分块配置和 ClamAV 服务骨架。
- 验收证据：
  - `npm run test --workspace @infinitech/object-scan-worker`
  - `npm run verify:architecture`
- 当前边界：已具备 ClamAV 下载扫描首版和部署骨架；真实 MinIO SDK/STS/Vault 临时凭证、隔离 bucket、扫描失败重试告警、病毒库更新观测、对象生命周期清理和生产对象存储权限策略仍待推进。
- 文件：
  - `services/object-scan-worker/src/index.mjs`
  - `services/object-scan-worker/src/index.test.mjs`
  - `infra/docker/compose.yml`
  - `infra/k8s/base/app-stack.yaml`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `EXECUTION_LEDGER.md`
  - `PLATFORM_MASTER_PLAN.md`
  - `docs/product/commercial-readiness-checklist.md`

### DONE-20260522-074 对象生命周期清理 worker 首版

- 日期：2026-05-22
- 结果：售后证据对象从“扫描后只做门禁”推进到“过期/拒绝对象可清理且可回写”。新增 `deleted` 上传票据状态、`cleanup_reason` 和 `deleted_at` 持久化字段；核心 API 新增管理员接口 `GET /api/admin/object-storage/cleanup-candidates` 和 `POST /api/admin/object-storage/cleanup-complete`，只返回过期未确认或扫描拒绝且已过保留期的对象，确认过的证据不会进入清理候选。新增独立 `@infinitech/object-lifecycle-worker` workspace，可拉取清理候选、按 `OBJECT_STORAGE_DELETE_BASE_URL` 发起对象删除，404 视为幂等成功，并在删除后回写票据 `deleted` 状态；BFF、Docker Compose、Kubernetes 部署骨架和架构守卫已同步。
- 验收证据：
  - `cd services/api-go && go test ./internal/platform`
  - `cd services/api-go && go test ./internal/httpapi`
  - `npm run test --workspace @infinitech/bff`
  - `npm run test --workspace @infinitech/object-lifecycle-worker`
  - `npm run verify:architecture`
- 当前边界：已具备过期未确认和扫描拒绝对象的清理候选、删除执行和完成回写；真实 MinIO SDK/STS/Vault 临时凭证、隔离 bucket 策略、删除失败告警、删除凭据轮换和生产对象存储 IAM 权限仍待推进。
- 文件：
  - `package.json`
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `services/object-lifecycle-worker/package.json`
  - `services/object-lifecycle-worker/src/index.mjs`
  - `services/object-lifecycle-worker/src/index.test.mjs`
  - `infra/db/migrations/0002_auth_payment.sql`
  - `infra/docker/compose.yml`
  - `infra/k8s/base/app-stack.yaml`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `EXECUTION_LEDGER.md`
  - `PLATFORM_MASTER_PLAN.md`
  - `docs/product/commercial-readiness-checklist.md`

### DONE-20260522-075 对象生命周期清理失败账本首版

- 日期：2026-05-22
- 结果：对象生命周期清理从“删除失败只留在 worker 当轮结果”推进到“失败可入账、可后台查看、可继续重试”。`order_after_sales_evidence_upload_tickets` 新增 `cleanup_attempts`、`last_cleanup_error` 和 `last_cleanup_failed_at` 字段；核心 API 新增 `POST /api/admin/object-storage/cleanup-failed`，管理员或生命周期 worker 可在删除失败时回写失败原因和时间，成功清理后保留尝试次数并清空最后失败信息。清理候选接口会返回失败次数和最后失败信息，BFF 已代理该接口，`@infinitech/object-lifecycle-worker` 删除失败时会自动回写失败账本，便于后续接 Prometheus/Loki 告警和运营后台处理。
- 验收证据：
  - `cd services/api-go && go test ./internal/platform`
  - `cd services/api-go && go test ./internal/httpapi`
  - `npm run test --workspace @infinitech/bff`
  - `npm run test --workspace @infinitech/object-lifecycle-worker`
  - `npm run verify:architecture`
  - `npm run verify`
  - `cd services/api-go && go test -count=1 ./...`
- 当前边界：已具备删除失败入账、候选暴露和 worker 自动上报；真实 MinIO SDK/STS/Vault 临时凭证、隔离 bucket 策略、告警投递、删除凭据轮换和生产对象存储 IAM 权限仍待推进。
- 文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `services/object-lifecycle-worker/src/index.mjs`
  - `services/object-lifecycle-worker/src/index.test.mjs`
  - `infra/db/migrations/0002_auth_payment.sql`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `EXECUTION_LEDGER.md`
  - `PLATFORM_MASTER_PLAN.md`
  - `docs/product/commercial-readiness-checklist.md`

### DONE-20260522-076 对象生命周期清理统计接口首版

- 日期：2026-05-22
- 结果：对象清理从“能查候选和失败明细”推进到“后台可直接读取运营统计”。新增 `ObjectStorageCleanupStats` 合同和 `GET /api/admin/object-storage/cleanup-stats` 管理端接口，按同一套 `now`/`grace_seconds` 口径汇总 pending、expired unconfirmed、scan rejected、failed、deleted、累计清理尝试次数、最近失败时间和最近删除时间；BFF 已代理该统计接口，架构守卫覆盖 API/BFF/测试，便于后续 Admin Web 告警面板、Prometheus exporter 或运营日报直接接入。
- 验收证据：
  - `cd services/api-go && go test ./internal/platform`
  - `cd services/api-go && go test ./internal/httpapi`
  - `npm run test --workspace @infinitech/bff`
  - `npm run verify:architecture`
  - `npm run verify`
  - `cd services/api-go && go test -count=1 ./...`
- 当前边界：已具备对象清理统计 API 与 BFF 代理；仍需把统计接入 Admin Web 页面、Prometheus/Loki 告警规则和真实对象存储 IAM/Vault 策略。
- 文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `EXECUTION_LEDGER.md`
  - `PLATFORM_MASTER_PLAN.md`
  - `docs/product/commercial-readiness-checklist.md`

### DONE-20260522-077 GitHub 协作与质量门禁首版

- 日期：2026-05-22
- 结果：仓库从“已上传代码”推进到“每次 push/PR 有基础质量门禁”。新增 `.github/workflows/verify.yml`，在 `push main` 和 `pull_request` 时设置 Node.js 22 与 Go，安装 workspace 依赖，运行 `npm run verify` 和 `cd services/api-go && go test -count=1 ./...`；新增 PR 模板，要求填写商业影响、验证和回滚说明；新增 Bug、Feature、Commercial readiness gap 三类 Issue 模板；新增 CODEOWNERS 和 Dependabot 配置。架构守卫已覆盖这些协作文件，防止后续误删质量门禁。
- 验收证据：
  - `npm run verify:architecture`
  - `npm run verify`
  - `cd services/api-go && go test -count=1 ./...`
- 当前边界：本地验证已通过，CI 文件已入仓；仍需观察 GitHub 托管环境第一次 Actions 运行结果，后续补分支保护、必过检查、release/tag 策略、镜像构建和部署流水线。
- 文件：
  - `.github/workflows/verify.yml`
  - `.github/pull_request_template.md`
  - `.github/ISSUE_TEMPLATE/bug_report.yml`
  - `.github/ISSUE_TEMPLATE/feature_request.yml`
  - `.github/ISSUE_TEMPLATE/commercial_gap.yml`
  - `.github/CODEOWNERS`
  - `.github/dependabot.yml`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `PROJECT_STATUS.md`
  - `EXECUTION_LEDGER.md`
  - `PLATFORM_MASTER_PLAN.md`
  - `docs/product/commercial-readiness-checklist.md`

### DONE-20260522-078 outbox 租约测试去时间漂移

- 日期：2026-05-22
- 结果：为保证 GitHub Actions 和本地非缓存 Go 测试稳定，修复 outbox 租约相关测试对固定 `2026-05-22T12:00:00Z` 的依赖。`TestClaimOutboxEventsLeasesReadyEvents`、`TestOutboxStatsReportsLeaseHealthByTopicAndOwner`、`TestRenewOutboxEventLeaseRequiresCurrentActiveOwner` 改为从真实 outbox event 的 `CreatedAt` 推导 claim/renew 时间；`TestAdminOutboxHTTPFlow` 改为从 HTTP 返回的 `created_at` 推导 claim/renew/stats 时间和期望租约过期时间，避免当前系统时间超过固定时间后事件被判断为未 ready，导致 CI 随机失败。
- 验收证据：
  - `cd services/api-go && go test -count=1 ./internal/platform ./internal/httpapi`
  - `npm run verify:architecture`
- 当前边界：outbox 租约测试已去固定日期漂移；仍需继续观察 GitHub Actions 首次运行，并后续补分支保护与必过检查。
- 文件：
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `EXECUTION_LEDGER.md`

### DONE-20260522-079 管理端 Web 最小运营控制台首版

- 日期：2026-05-22
- 结果：`apps/admin-web` 从占位页推进到可打开的桌面运营控制台首版。新增静态入口、运营导航、P0 指标位、今日必盯队列、RBAC 草案、模块状态列表和接口操作台；接口操作台已接入现有 BFF/API 的管理员登录、商户邀请、站长/骑手邀请、退款策略读取/保存、售后列表、对象清理统计/候选、outbox 健康/事件/批量恢复、订单状态补偿等操作。根目录 verify 已加入 `verify:apps`，管理端 Web 自带 Node 测试，架构守卫也会检查管理端首版关键文件和操作入口。
- 验收证据：
  - `npm run test --workspace @infinitech/admin-web`
  - `npm run verify:architecture`
  - 浏览器打开 `http://127.0.0.1:4173/apps/admin-web/index.html`，确认页面非白屏，包含“运营首页/接口操作台/今日必须盯住”，点击 Outbox 健康可切换操作表单，浏览器控制台无 error。
- 当前边界：这是最小可运营控制台，不是完整后台。仍需接真实业务列表页、细分 RBAC 服务端策略、操作审计、敏感字段脱敏、商户资质审核实页、骑手/站长管理实页、客服/RTC/优惠券/首页卡片/圈子饭搭/红包等完整面板。
- 文件：
  - `apps/admin-web/index.html`
  - `apps/admin-web/package.json`
  - `apps/admin-web/src/main.js`
  - `apps/admin-web/src/styles.css`
  - `apps/admin-web/src/adminApi.mjs`
  - `apps/admin-web/src/adminApi.test.mjs`
  - `apps/admin-web/src/config.mjs`
  - `apps/admin-web/README.md`
  - `package.json`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `PROJECT_STATUS.md`
  - `PLATFORM_MASTER_PLAN.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260522-080 管理端 P0 业务视图首版

- 日期：2026-05-22
- 结果：管理端 Web 从“全局工作台 + 接口操作台”继续推进到“P0 业务视图首版”。新增 `adminViews` 视图层，订单监控、售后审核、商户资质、骑手/站长、骑手绩效、派单审计、退款策略均有独立页面结构，包含关键指标、操作入口、表格列、首批运营行和安全约束。新增站点骑手、站点订单、骑手绩效、站点任务配置等可执行操作入口，后续可直接接入真实 BFF/API 数据。
- 验收证据：
  - `npm run test --workspace @infinitech/admin-web`
  - `npm run verify:architecture`
  - 浏览器打开 `http://127.0.0.1:4173/apps/admin-web/index.html`，点击订单监控、商户资质、骑手/站长模块，均能展示对应业务视图；浏览器控制台无 error。
- 当前边界：这些是 P0 页面结构和操作入口，仍需把表格行替换为真实 API 数据源，补分页、筛选、详情抽屉、审核表单、操作审计落库和服务端细分 RBAC；审计 payload 服务端白名单/掩码已在后续 `DONE-20260523-089` 落地首版。
- 文件：
  - `apps/admin-web/src/adminViews.mjs`
  - `apps/admin-web/src/main.js`
  - `apps/admin-web/src/styles.css`
  - `apps/admin-web/src/adminApi.mjs`
  - `apps/admin-web/src/adminApi.test.mjs`
  - `apps/admin-web/README.md`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `PROJECT_STATUS.md`
  - `PLATFORM_MASTER_PLAN.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260522-081 管理端运营快照 API 首版

- 日期：2026-05-22
- 结果：新增 `GET /api/admin/operations/snapshot` 管理端运营快照接口，按同一后台口径聚合订单状态、异常订单、商户资质/保证金/店铺资料、骑手/站长/在线/保证金、骑手绩效等级、售后队列、派单审计事件、退款策略、outbox 健康和对象清理统计。BFF 已代理该接口，管理端操作台和 P0 业务视图已加入“运营快照”入口，后续页面可直接绑定该接口或拆分出的明细接口。
- 验收证据：
  - `cd services/api-go && go test -count=1 ./...`
  - `npm run verify`
- 当前边界：这是管理端数据口径聚合首版，不是完整后台实页。仍需把 P0 表格/指标真正绑定接口响应，补分页筛选、详情抽屉、审核表单、操作审计落库、服务端细分 RBAC 和生产级数据库查询优化；审计 payload 服务端白名单/掩码已在后续 `DONE-20260523-089` 落地首版。
- 文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `apps/admin-web/src/adminApi.mjs`
  - `apps/admin-web/src/adminApi.test.mjs`
  - `apps/admin-web/src/adminViews.mjs`
  - `apps/admin-web/README.md`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `PROJECT_STATUS.md`
  - `PLATFORM_MASTER_PLAN.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260522-082 管理端 P0 快照绑定首版

- 日期：2026-05-22
- 结果：管理端 Web 新增 `adminSnapshot` 快照适配层，把 `/api/admin/operations/snapshot` 响应转换为顶部 KPI、今日 P0 队列、订单监控、售后审核、商户资质、骑手/站长、骑手绩效、派单审计和退款策略视图数据。有管理员 token 时可手动刷新运营快照，执行“运营快照”操作成功后也会更新页面数据；登录返回 token 一键填入后会尝试刷新。所有快照字段进入页面前通过 HTML 转义渲染，避免商户名、订单字段、售后字段等后端内容直接注入页面。
- 验收证据：
  - `npm run test --workspace @infinitech/admin-web`
  - `npm run verify:architecture`
- 当前边界：P0 页面已能按运营快照生成首批表格和指标，但仍需分页筛选、详情抽屉、审核表单、操作审计落库和服务端细分 RBAC；审计 payload 服务端白名单/掩码已在后续 `DONE-20260523-089` 落地首版。
- 文件：
  - `apps/admin-web/src/adminSnapshot.mjs`
  - `apps/admin-web/src/main.js`
  - `apps/admin-web/src/styles.css`
  - `apps/admin-web/src/adminApi.test.mjs`
  - `scripts/check-architecture.mjs`
  - `apps/admin-web/README.md`
  - `README.md`
  - `PROJECT_STATUS.md`
  - `PLATFORM_MASTER_PLAN.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260522-083 BFF 浏览器 CORS 白名单首版

- 日期：2026-05-22
- 结果：BFF 新增浏览器来源白名单和 `OPTIONS` 预检处理，默认允许本地管理端/Flutter 调试来源，并支持通过 `BFF_ALLOWED_ORIGINS` 配置生产或测试来源。管理端从静态预览页访问 BFF 时，`Authorization`、`Content-Type` 和 `X-Client-Kind` 头已被明确允许，`/api/admin/operations/snapshot` 代理响应会返回匹配的 `Access-Control-Allow-Origin`。非法来源会得到 `FORBIDDEN_ORIGIN`，不返回跨域放行头；后台带 token 的接口不接受 `*` 通配来源。
- 验收证据：
  - `npm run test --workspace @infinitech/bff`
  - `npm run verify:architecture`
- 当前边界：这是浏览器到 BFF 的基础白名单，不等于完整 API Gateway/WAF 能力；仍需生产域名白名单、限流、灰度、认证失败统一错误码、观测埋点和安全扫描。
- 文件：
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `PROJECT_STATUS.md`
  - `PLATFORM_MASTER_PLAN.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260522-084 管理端操作审计日志首版

- 日期：2026-05-22
- 结果：新增管理端操作审计账本首版，领域层支持 `RecordAuditLog` 与 `AuditLogs`，可按 actor、action、target、limit 和 before 查询。HTTP 新增 `GET /api/admin/audit-logs`，BFF 已代理，Admin Web 操作台已加入“操作审计”入口。管理员关键写操作开始落审计：商户邀约、骑手/站长邀约、退款策略保存、订单退款、订单状态补偿、售后审核、对象清理完成/失败、outbox 领取/续租/发布/失败/重放/批量重放。审计记录包含 actor、action、target、request_id、ip_hash、服务端白名单 payload 和 created_at，并进入 PostgreSQL-backed Store 快照。
- 验收证据：
  - `cd services/api-go && go test -count=1 ./...`
  - `npm run test --workspace @infinitech/bff`
  - `npm run test --workspace @infinitech/admin-web`
  - `npm run verify:architecture`
- 当前边界：这是首版操作审计，不是完整审计后台。当前审计随 Store 快照持久化，后续还要推进到规范化 `audit_logs` 表事务内强制写入、细分 RBAC、审计检索页、导出与留存策略、异常告警和 KMS/链式不可抵赖签名；审计 payload 服务端白名单/掩码已在后续 `DONE-20260523-089` 落地首版，审计完整性证明已在后续 `DONE-20260523-090` 落地首版。
- 文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `apps/admin-web/src/adminApi.mjs`
  - `apps/admin-web/src/adminApi.test.mjs`
  - `apps/admin-web/src/adminViews.mjs`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `PROJECT_STATUS.md`
  - `PLATFORM_MASTER_PLAN.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `apps/admin-web/README.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260522-085 审计日志 PostgreSQL 规范化表首版

- 日期：2026-05-22
- 结果：把管理端操作审计从仅随 Store 快照保存推进到 PostgreSQL `audit_logs` 规范化表。`PostgresStore` 启动时会确保 `audit_logs` 表与 actor/action/target 索引存在，兼容旧 UUID id 表并转为 TEXT；会把旧快照审计幂等补入表，并从表内 `aud_N` 最大编号恢复审计序列。新增审计写入通过 `platform_sequences` 的 `audit_logs` 序列行级锁生成 `aud_N`，再使用 `audit_logs` 表 `ON CONFLICT (id) DO NOTHING` 保持追加式账本语义，查询 `/api/admin/audit-logs` 在 PostgreSQL-backed Store 下直接走规范化表，支持 actor、action、target、before、limit 过滤。
- 验收证据：
  - `cd services/api-go && go test -count=1 ./...`
  - `npm run verify:architecture`
- 当前边界：审计已进入规范化 PostgreSQL 表，但仍不是完整商业审计中心。关键业务写操作与审计写入还不是同一个业务事务内的强制原子提交；仍需补细分 RBAC、审计检索页、留存策略、异常告警、KMS/链式不可抵赖签名和审计归档冷热分层；审计 payload 服务端白名单/掩码已在后续 `DONE-20260523-089` 落地首版，审计完整性证明已在后续 `DONE-20260523-090` 落地首版，审计导出已在后续 `DONE-20260524-102` 落地首版。
- 文件：
  - `infra/db/migrations/0001_core.sql`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/platform/postgres_store_test.go`
  - `scripts/check-architecture.mjs`
  - `EXECUTION_LEDGER.md`

### DONE-20260522-086 最近进展与路线图 Markdown 同步

- 日期：2026-05-22
- 结果：新增 `docs/product/recent-progress-roadmap.md`，集中记录最近已完成、当前未完成、未来将要完成的工作和下一批优先推进顺序；`README.md` 已加入入口，`PROJECT_STATUS.md` 已补最近进展摘要。当前口径明确：项目正在按商业级标准推进，但仍未完成真实生产支付、真实 IM/RTC、完整管理端、高可用基础设施、10 万在线压测和容灾演练，不能宣称已经商业级上线。
- 验收证据：
  - `npm run verify`
- 当前边界：这是项目状态和路线图同步，不是业务能力新增；后续每次完成关键能力都要继续同步本文件、`PROJECT_STATUS.md`、`PLATFORM_MASTER_PLAN.md` 和商业级验收清单。
- 文件：
  - `docs/product/recent-progress-roadmap.md`
  - `README.md`
  - `PROJECT_STATUS.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260523-087 管理端审计检索页首版

- 日期：2026-05-23
- 结果：Admin Web 新增“审计检索”P0 模块，从“接口操作台查 JSON”推进到可扫读的审计中心首版。页面可按 `actor_type`、`actor_id`、`action`、`target_type`、`target_id`、`before` 和 `limit` 查询 `/api/admin/audit-logs`，支持 before 游标翻页；新增 `adminAudit` 适配层，按白名单生成 payload 摘要，并对 password、secret、token、authorization、openid、session、credential、phone、object key、签名等敏感字段做脱敏详情展示。RBAC 草案新增 `security_auditor` 安全审计员角色；架构守卫已锁定审计检索、脱敏适配和筛选参数，防止回退到裸 JSON 操作台。
- 验收证据：
  - `npm run test --workspace @infinitech/admin-web`
  - `npm run verify:architecture`
- 当前边界：这是 Admin Web 审计检索首版，仍不是完整商业审计中心。全域服务端细分 RBAC、关键业务写操作与审计写入同事务强制提交、审计留存、异常告警、KMS/链式不可抵赖签名和冷热归档仍待推进；审计 payload 服务端白名单/掩码已在后续 `DONE-20260523-089` 落地首版，审计完整性证明已在后续 `DONE-20260523-090` 落地首版，审计导出已在后续 `DONE-20260524-102` 落地首版，审计留存/告警健康报告已在后续 `DONE-20260524-103` 落地首版。
- 文件：
  - `apps/admin-web/src/adminAudit.mjs`
  - `apps/admin-web/src/adminApi.mjs`
  - `apps/admin-web/src/adminApi.test.mjs`
  - `apps/admin-web/src/adminViews.mjs`
  - `apps/admin-web/src/config.mjs`
  - `apps/admin-web/src/main.js`
  - `apps/admin-web/src/styles.css`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `PROJECT_STATUS.md`
  - `PLATFORM_MASTER_PLAN.md`
  - `docs/product/recent-progress-roadmap.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `apps/admin-web/README.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260523-088 管理端审计中心增强首版

- 日期：2026-05-23
- 结果：审计中心从“可查首版”继续补到可运营追查首版。`GET /api/admin/audit-logs` 新增 `after` 时间下界，内存 Store 与 PostgreSQL 查询统一按 `created_at >= after` 过滤，`before` 保持严格小于游标；Admin Web 审计检索页支持 actor/action/target/after/before/limit 筛选、保存常用筛选、before 游标翻页、详情抽屉、按当前目标继续筛选和跨模块跳转；目标映射已覆盖订单、售后、商户、骑手、outbox、对象清理和退款策略等已落地运营模块。
- 验收证据：
  - `npm run test --workspace @infinitech/admin-web`
  - `cd services/api-go && go test ./internal/platform ./internal/httpapi`
  - `npm run verify:architecture`
  - `npm run verify`
- 当前边界：审计中心仍未达到完整商业审计闭环。全域服务端细分 RBAC、关键业务写操作与审计写入同事务强制提交、审计留存、异常告警、KMS/链式不可抵赖签名和冷热归档仍待推进；审计 payload 服务端白名单/掩码已在后续 `DONE-20260523-089` 落地首版，审计完整性证明已在后续 `DONE-20260523-090` 落地首版，审计导出已在后续 `DONE-20260524-102` 落地首版，审计留存/告警健康报告已在后续 `DONE-20260524-103` 落地首版。
- 文件：
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/platform/postgres_store_test.go`
  - `apps/admin-web/src/adminAudit.mjs`
  - `apps/admin-web/src/adminApi.mjs`
  - `apps/admin-web/src/adminApi.test.mjs`
  - `apps/admin-web/src/main.js`
  - `apps/admin-web/src/styles.css`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `PROJECT_STATUS.md`
  - `PLATFORM_MASTER_PLAN.md`
  - `docs/product/recent-progress-roadmap.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `apps/admin-web/README.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260523-089 管理端审计服务端安全边界首版

- 日期：2026-05-23
- 结果：把管理端审计从“前端可脱敏展示”推进到服务端可信边界首版。`api-go` 新增 `RoleSecurityAuditor = "security_auditor"` 和 `Principal.CanReadAuditLogs()`，`GET /api/admin/audit-logs` 允许管理员与安全审计员读取，商户邀请、订单状态补偿等后台写操作仍要求管理员；`auth_identities`、`auth_sessions` 迁移和运行时建表已允许 `security_auditor` 主体类型，避免生产会话落库失败。审计 payload 白名单与敏感字段掩码已下沉到 Store/PostgreSQL 路径，`RecordAuditLog`、SQL 写入、SQL 读取和 SQL 镜像恢复都会调用统一 `sanitizeAuditPayload`，保留 `default_refund_strategy`、`amount_fen` 等允许字段，掩码 `object_key` 等敏感允许字段，并丢弃 `password`、`token`、`phone`、`nested`、`raw_request` 等非白名单或敏感字段。
- 验收证据：
  - `cd services/api-go && go test ./internal/platform ./internal/httpapi`
  - `npm run verify:architecture`
- 当前边界：这是审计服务端只读角色和 payload 安全边界首版，不是完整商业审计闭环。关键业务写操作与审计写入仍需推进到同一业务事务强制提交；全域服务端 RBAC 策略矩阵、审计留存、异常告警、KMS/链式不可抵赖签名、冷热归档和策略治理仍待补齐；审计完整性证明已在后续 `DONE-20260523-090` 落地首版，审计导出已在后续 `DONE-20260524-102` 落地首版，审计留存/告警健康报告已在后续 `DONE-20260524-103` 落地首版。
- 文件：
  - `services/api-go/internal/httpapi/auth.go`
  - `services/api-go/internal/httpapi/auth_session.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/platform/postgres_store_test.go`
  - `infra/db/migrations/0001_core.sql`
  - `infra/db/migrations/0002_auth_payment.sql`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `PROJECT_STATUS.md`
  - `PLATFORM_MASTER_PLAN.md`
  - `docs/product/recent-progress-roadmap.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `apps/admin-web/README.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260523-090 管理端审计完整性证明首版

- 日期：2026-05-23
- 结果：把管理端审计账本从“可查、可脱敏、可只读授权”推进到可检测篡改的完整性证明首版。`AuditLog` 新增 `integrity_algorithm`、`integrity_hash`、`integrity_verified`；`audit_logs` 表新增 `integrity_algorithm` 与 `integrity_hash` 并兼容运行时补列。内存 Store 与 PostgreSQL 写入会对审计 ID、actor、action、target、request_id、ip_hash、服务端白名单 payload 和微秒精度 `created_at` 生成稳定证明；payload 内 `time.Time` 会规范为 UTC 字符串，避免 JSONB 读回类型漂移导致误报。无密钥环境默认 `sha256:v1`，生产配置 `AUDIT_LOG_SIGNING_SECRET` 后使用 `hmac-sha256:v1`；查询时重新验证并返回 `integrity_verified`，可发现数据库中审计字段或白名单 payload 被篡改。旧快照或旧 SQL 行缺少完整性字段时，只有业务字段完全一致才会回填证明；Admin Web 审计中心已展示完整性状态、算法和哈希。
- 验收证据：
  - `cd services/api-go && go test ./internal/platform ./internal/httpapi`
  - `npm run test --workspace @infinitech/admin-web`
  - `npm run verify:architecture`
  - `npm run verify`
- 当前边界：这是审计完整性证明首版，不是法律级不可抵赖归档。关键业务写操作与审计写入仍需推进到同一业务事务强制提交；全域服务端 RBAC、留存策略、异常告警、KMS/Vault 密钥轮换、链式账本、WORM/冷归档和策略治理仍待补齐；审计导出已在后续 `DONE-20260524-102` 落地首版。
- 文件：
  - `services/api-go/cmd/api/main.go`
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/platform/postgres_store_test.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `infra/db/migrations/0001_core.sql`
  - `apps/admin-web/src/adminAudit.mjs`
  - `apps/admin-web/src/adminApi.test.mjs`
  - `apps/admin-web/src/adminViews.mjs`
  - `apps/admin-web/src/main.js`
  - `apps/admin-web/src/styles.css`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `PROJECT_STATUS.md`
  - `PLATFORM_MASTER_PLAN.md`
  - `docs/product/recent-progress-roadmap.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `apps/admin-web/README.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260523-091 退款策略配置与审计同事务首版

- 日期：2026-05-23
- 结果：把 `PUT /api/admin/refund-settings` 从 HTTP 层“先保存退款策略、再单独补审计”迁移到仓储级原子路径。`Repository` 新增 `SaveRefundSettingsWithAudit`；内存 Store 在同一把业务锁内更新退款策略并写入 `admin.refund_settings.updated` 审计；PostgreSQL-backed Store 在同一个数据库事务内写入 `refund_settings` 与 `audit_logs`，审计 ID 继续走 `platform_sequences` 行级锁，提交后刷新内存镜像。退款策略审计 payload 由服务端重新生成，只保留规范化后的 `default_refund_strategy`，即便调用方传入其他白名单字段也不会混入该动作的审计。
- 验收证据：
  - `cd services/api-go && go test ./internal/platform ./internal/httpapi`
  - `node --test scripts/check-architecture.mjs`
  - 本轮收尾继续跑 `npm run test --workspace @infinitech/admin-web`、`npm run verify:architecture`、`npm run verify` 和 `git diff --check`
- 当前边界：这是首个关键配置写路径的业务写入与审计写入同事务提交，不代表所有后台写操作都已完成原子审计；后续 `DONE-20260523-092` 已继续迁移管理端订单退款，`DONE-20260523-093` 已继续迁移售后审核，`DONE-20260523-094` 已继续迁移订单状态补偿。对象清理完成/失败、outbox 领取/续租/发布/失败/重放/批量重放、商户/骑手邀约等写路径仍需继续迁移到同一业务事务强制提交；审计留存健康报告已在后续 `DONE-20260524-103` 落地，真实告警投递、KMS/Vault 密钥轮换、链式账本、WORM/冷归档和策略治理仍待补齐；审计导出已在后续 `DONE-20260524-102` 落地首版。
- 文件：
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `PROJECT_STATUS.md`
  - `PLATFORM_MASTER_PLAN.md`
  - `docs/product/recent-progress-roadmap.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `apps/admin-web/README.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260523-092 管理端订单退款与审计同事务首版

- 日期：2026-05-23
- 结果：把 `POST /api/orders/{orderID}/refund` 从 HTTP 层“先执行退款、再单独补审计”迁移到仓储级 `RefundOrderWithAudit` 原子路径。内存 Store 在同一把业务锁内执行退款与写入 `admin.order.refunded` 审计；PostgreSQL-backed Store 在同一个数据库事务内写入退款交易、钱包退款流水、订单状态/事件和 `audit_logs`，审计 ID 继续走 `platform_sequences` 行级锁。退款审计 payload 由服务端根据最终退款交易重新生成，只保留 `refund_id`、`destination`、`status`、`amount_fen`、`idempotency_key`。
- 验收证据：
  - `cd services/api-go && go test ./internal/platform ./internal/httpapi`
  - `node --test scripts/check-architecture.mjs`
  - 本轮收尾继续跑 `npm run test --workspace @infinitech/admin-web`、`npm run verify:architecture`、`npm run verify` 和 `git diff --check`
- 当前边界：这是第二条后台写路径的业务写入与审计写入同事务提交，不代表所有后台写操作都已完成原子审计；后续 `DONE-20260523-093` 已继续迁移售后审核，`DONE-20260523-094` 已继续迁移订单状态补偿。对象清理完成/失败、outbox 领取/续租/发布/失败/重放/批量重放、商户/骑手邀约等仍需继续迁移。
- 文件：
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `PROJECT_STATUS.md`
  - `PLATFORM_MASTER_PLAN.md`
  - `docs/product/recent-progress-roadmap.md`
  - `docs/product/commercial-readiness-checklist.md`
	  - `apps/admin-web/README.md`
	  - `EXECUTION_LEDGER.md`

### DONE-20260523-093 售后审核与审计同事务首版

- 日期：2026-05-23
- 结果：把 `POST /api/after-sales/{requestID}/review` 从 HTTP 层“先审核售后、再单独补审计”迁移到仓储级 `ReviewAfterSalesWithAudit` 原子路径。内存 Store 在同一把业务锁内完成售后审核、必要退款、订单事件和 `after_sales.reviewed` 审计；PostgreSQL-backed Store 在同一个数据库事务内锁定售后申请与订单，写入售后审核结果、售后事件、订单事件、必要退款交易/钱包退款流水/订单退款状态和 `audit_logs`，审计 ID 继续走 `platform_sequences` 行级锁。售后审核审计 payload 由服务端根据最终审核结果和退款交易重新生成，只保留 `decision`、`status`、`refund_id`、`amount_fen`、`destination`、`idempotency_key`。
- 验收证据：
  - `cd services/api-go && go test ./internal/platform ./internal/httpapi`
  - `node --test scripts/check-architecture.mjs`
  - 本轮收尾继续跑 `npm run test --workspace @infinitech/admin-web`、`npm run verify:architecture`、`npm run verify`、`cd services/api-go && go test ./...` 和 `git diff --check`
- 当前边界：这是第三条后台写路径的业务写入与审计写入同事务提交，不代表所有后台写操作都已完成原子审计；后续 `DONE-20260523-094` 已继续迁移订单状态补偿。对象清理完成/失败、outbox 领取/续租/发布/失败/重放/批量重放、商户/骑手邀约等仍需继续迁移。
- 文件：
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `PROJECT_STATUS.md`
  - `PLATFORM_MASTER_PLAN.md`
  - `docs/product/recent-progress-roadmap.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `apps/admin-web/README.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260523-094 订单状态补偿与审计同事务首版

- 日期：2026-05-23
- 结果：把 `POST /api/admin/orders/{orderID}/state/compensate` 从 HTTP 层“先补偿订单状态、再单独补审计”迁移到仓储级 `CompensateOrderStateWithAudit` 原子路径。内存 Store 复用同一把业务锁完成订单状态/骑手/支付方式漂移修复并写入 `admin.order_state.compensated` 审计；PostgreSQL-backed Store 在同一个数据库事务内锁定订单、读取订单事件/钱包流水/支付交易/派单事件，复用状态补偿计划逻辑更新 `orders`、写入 `order_events` 和 `audit_logs`，审计 ID 继续走 `platform_sequences` 行级锁。订单状态补偿审计 payload 由服务端根据最终补偿结果重新生成，只保留 `changed`、`previous_status`、`expected_status`、`compensation_type`、`evidence_count` 和必要 rider 字段，调用方伪造 payload 不会混入审计。
- 验收证据：
  - `cd services/api-go && go test ./internal/platform ./internal/httpapi`
  - `node --test scripts/check-architecture.mjs`
  - 本轮收尾继续跑 `npm run test --workspace @infinitech/admin-web`、`npm run verify:architecture`、`cd services/api-go && go test ./...`、`npm run verify` 和 `git diff --check`
- 当前边界：这是第四条后台写路径的业务写入与审计写入同事务提交，不代表所有后台写操作都已完成原子审计。对象清理完成/失败、outbox 领取/续租/发布/失败/重放/批量重放、商户/骑手邀约等仍需继续迁移。
- 文件：
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `PROJECT_STATUS.md`
  - `PLATFORM_MASTER_PLAN.md`
  - `docs/product/recent-progress-roadmap.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `apps/admin-web/README.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260524-095 outbox 运维与审计同事务首版

- 日期：2026-05-24
- 结果：把管理端 outbox 领取、租约续期、发布确认、失败回写、单条重放和批量重放从 HTTP 层“先更新 outbox、再单独补审计”迁移到仓储级原子路径。新增 `ClaimOutboxEventsWithAudit`、`RenewOutboxEventLeaseWithAudit`、`MarkOutboxEventPublishedWithAudit`、`MarkOutboxEventFailedWithAudit`、`ReplayOutboxEventWithAudit` 和 `ReplayOutboxEventsWithAudit`，内存 Store 在同一把业务锁内更新 outbox 状态并写入审计，PostgreSQL-backed Store 在同一数据库事务内更新 `platform_outbox_events` 并插入 `audit_logs`。outbox 审计 payload 改为服务端根据最终事件/批量结果生成，只保留 `topic`、`status`、`attempts`、`lease_owner`、`lease_seconds`、`retry_after_seconds`、`claimed`、`replayed` 和 `limit` 等白名单字段，调用方伪造 payload 不会进入审计。
- 验收证据：
  - `cd services/api-go && go test -count=1 ./...`
  - `npm run verify`
  - `git diff --check`
- 当前边界：这是 outbox 运维写路径的业务写入与审计写入同事务首版，不代表审计中心已经完整商业闭环。商户/骑手邀约已在后续 `DONE-20260524-096` 迁移，服务端 RBAC 策略矩阵首版已在后续 `DONE-20260524-097` 落地，审计导出已在后续 `DONE-20260524-102` 落地，审计留存/告警健康报告已在后续 `DONE-20260524-103` 落地；剩余后台配置/运营处置/资金风控写路径审计同事务、角色/权限配置 UI、字段级/租户级权限、真实告警投递、KMS/Vault 密钥轮换、链式账本、WORM/冷归档和策略治理仍待补齐。
- 文件：
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `PROJECT_STATUS.md`
  - `PLATFORM_MASTER_PLAN.md`
  - `docs/product/recent-progress-roadmap.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `apps/admin-web/README.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260524-096 商户/骑手邀约与审计同事务首版

- 日期：2026-05-24
- 结果：把 `POST /api/admin/merchant-invites`、`POST /api/admin/rider-invites` 和 `POST /api/station-manager/rider-invites` 从 HTTP 层“先创建邀约、再单独补审计”迁移到仓储级原子路径。新增 `CreateMerchantInviteWithAudit` 与 `CreateRiderInviteWithAudit`，内存 Store 在同一把业务锁内生成最终邀约 token 并写入审计；PostgreSQL-backed Store 在同一个数据库事务内插入 `audit_logs` 并保存包含最终邀约状态的 `platform_store_snapshots`，审计 ID 继续走 `platform_sequences` 行级锁。邀约审计 payload 改为服务端根据最终邀约重新生成，只保留 `type`、`expires_at`，骑手邀约额外保留 `station_id`，调用方伪造 token、station 或过期时间不会进入审计。
- 验收证据：
  - `cd services/api-go && go test -count=1 ./...`
  - `npm run verify`
  - `git diff --check`
- 当前边界：这是商户/骑手/站长邀约写路径的业务写入与审计写入同事务首版，不代表审计中心已经完整商业闭环。服务端 RBAC 策略矩阵首版已在后续 `DONE-20260524-097` 落地，审计导出已在后续 `DONE-20260524-102` 落地，审计留存/告警健康报告已在后续 `DONE-20260524-103` 落地；后续仍需继续扫描后台配置、运营处置、资金和风控写路径，并补角色/权限配置 UI、字段级/租户级权限、真实告警投递、KMS/Vault 密钥轮换、链式账本、WORM/冷归档和策略治理。
- 文件：
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `PROJECT_STATUS.md`
  - `PLATFORM_MASTER_PLAN.md`
  - `docs/product/recent-progress-roadmap.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `apps/admin-web/README.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260524-097 管理端服务端 RBAC 策略矩阵首版

- 日期：2026-05-24
- 结果：把后台权限从单一 `admin`/`security_auditor` 推进到服务端 RBAC 策略矩阵首版。新增后台兼容 `admin`、`super_admin`、`ops_admin`、`finance_admin`、`dispatch_admin`、`support_admin`、`security_auditor` 角色和 scope 常量；邀约、退款、运营快照、售后读取/审核/客服事件、对象清理、outbox 运维、订单状态补偿、派单读写和审计读取已改为服务端权限判断。`security_auditor` 继续只能读审计；`finance_admin` 可读写退款和资金相关入口但不能创建邀约；`ops_admin` 可处理运营、邀约、售后审核、对象清理和 outbox；`dispatch_admin` 可看和管理调度但不能操作 outbox；`support_admin` 可读售后并追加客服事件但不能审批退款。`auth_sessions` 运行时表和认证迁移已允许新增后台主体类型，Admin Web RBAC 配置同步服务端 scope 命名，架构守卫固定该矩阵。
- 验收证据：
  - `cd services/api-go && go test -count=1 ./...`
  - `npm run verify`
  - `git diff --check`
- 当前边界：这是服务端 RBAC 首版，不等于后台权限治理已经商业闭环。后续仍需角色/权限配置 UI、字段级权限、租户/站点数据域、审批流、权限变更审计、菜单隐藏策略、留存策略和异常告警；审计导出已在后续 `DONE-20260524-102` 落地首版。
- 文件：
  - `services/api-go/internal/httpapi/auth.go`
  - `services/api-go/internal/httpapi/auth_test.go`
  - `services/api-go/internal/httpapi/auth_session.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `infra/db/migrations/0001_core.sql`
  - `infra/db/migrations/0002_auth_payment.sql`
  - `apps/admin-web/src/config.mjs`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `PROJECT_STATUS.md`
  - `PLATFORM_MASTER_PLAN.md`
  - `docs/product/recent-progress-roadmap.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `apps/admin-web/README.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260524-098 RBAC 权限治理查询与变更申请审计首版

- 日期：2026-05-24
- 结果：在服务端 RBAC 策略矩阵基础上补齐首个权限治理闭环。新增 `GET /api/admin/rbac/policy` 返回服务端真实角色、scope、当前登录角色权限、策略版本和变更模型；新增 `POST /api/admin/rbac/change-requests` 供 `admin`/`super_admin` 提交权限变更申请，申请会校验目标角色、scope 白名单、原因和 `*` 高危范围，并写入 `admin.rbac.change_requested` 审计日志，当前阶段不自动修改运行时权限。BFF 已代理两个接口；Admin Web 新增“权限治理”模块、操作目录、CSV scope 请求构造、审计跳转和 RBAC 只读展示入口；审计 payload 白名单补充权限申请字段。
- 验收证据：
  - 本轮收尾继续跑 `cd services/api-go && go test -count=1 ./...`
  - 本轮收尾继续跑 `npm run verify`
  - 本轮收尾继续跑 `git diff --check`
- 当时边界：这是权限治理查询与申请审计首版，不代表权限已可在页面动态审批并自动生效。后续审批/驳回台账已在 `DONE-20260524-099` 落地，手动应用已在 `DONE-20260524-100` 落地，审计导出已在 `DONE-20260524-102` 落地；字段级/租户级 RBAC、站点/商户数据域、菜单按权限隐藏、留存策略、异常告警和 KMS/链式不可抵赖签名仍未完成。
- 文件：
  - `services/api-go/internal/httpapi/auth.go`
  - `services/api-go/internal/httpapi/auth_test.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/api-go/internal/platform/store.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `apps/admin-web/src/config.mjs`
  - `apps/admin-web/src/adminApi.mjs`
  - `apps/admin-web/src/adminAudit.mjs`
  - `apps/admin-web/src/adminViews.mjs`
  - `apps/admin-web/src/main.js`
  - `apps/admin-web/src/adminApi.test.mjs`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `PROJECT_STATUS.md`
  - `docs/product/recent-progress-roadmap.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260524-099 RBAC 权限申请审批/驳回台账首版

- 日期：2026-05-24
- 结果：在 RBAC 变更申请基础上补齐审批台账首版。新增 `GET /api/admin/rbac/change-requests`，从规范化审计账本重建权限申请状态，支持按 `pending_approval`、`approved`、`rejected` 筛选；新增 `POST /api/admin/rbac/change-requests/{id}/review`，`admin`/`super_admin` 可审批或驳回权限申请，服务端禁止提交人审批自己的申请。审批/驳回会写入 `admin.rbac.change_reviewed` 审计，审计目标为 `admin_rbac_change_request`，记录 decision、status、role、requested scopes、policy version 和原因；当前仍不自动修改运行时权限。BFF 与 Admin Web 操作目录已接入申请列表和审批操作，审计中心可跳转权限治理。
- 验收证据：
  - `cd services/api-go && go test -count=1 ./...`
  - `npm run test --workspace @infinitech/admin-web`
  - `npm run test --workspace @infinitech/bff`
  - `npm run verify:architecture`
  - 本轮收尾继续跑 `npm run verify` 和 `git diff --check`
- 当时边界：这是审批/驳回台账首版，不是最终权限动态应用系统。后续手动应用已在 `DONE-20260524-100` 落地，审计回滚已在 `DONE-20260524-101` 落地，审计导出已在 `DONE-20260524-102` 落地；字段级/租户级权限、产品化审批队列、留存策略、异常告警和 KMS/链式不可抵赖签名仍未完成。
- 文件：
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `apps/admin-web/src/adminApi.mjs`
  - `apps/admin-web/src/adminApi.test.mjs`
  - `apps/admin-web/src/adminAudit.mjs`
  - `apps/admin-web/src/adminViews.mjs`

### DONE-20260524-100 RBAC 权限变更手动应用首版

- 日期：2026-05-24
- 结果：在 RBAC 申请与审批台账之后补上手动应用链路。新增 `POST /api/admin/rbac/change-requests/{id}/apply`，只允许 `approved` 状态的申请应用到运行时权限矩阵；服务端禁止申请人直接应用自己的申请，应用动作写入 `admin.rbac.change_applied` 审计，payload 记录 change request、目标角色、应用 scopes、应用前 scopes、策略版本和原因。`Principal.HasAdminScope` 与 `AdminScopesForRole` 已读取应用后的运行时策略，`api-go` 启动时会从 `admin.rbac.change_applied` 审计日志重放已应用策略，避免重启后丢失应用结果。BFF 与 Admin Web 操作目录已接入应用入口，审计 payload 白名单补充 `applied_scopes` 和 `previous_scopes`。
- 验收证据：
  - `cd services/api-go && go test -count=1 ./internal/httpapi`
  - `npm run test --workspace @infinitech/admin-web`
  - `npm run test --workspace @infinitech/bff`
  - 本轮收尾继续跑 `go test -count=1 ./...`、`npm run verify` 和 `git diff --check`
- 当时边界：这是运行时权限应用首版，不是完整权限治理后台。后续审计回滚已在 `DONE-20260524-101` 落地，审计导出已在 `DONE-20260524-102` 落地；仍需字段级/租户级 RBAC、站点/商户数据域、产品化审批队列、菜单隐藏、留存策略、异常告警和 KMS/链式不可抵赖签名。
- 文件：
  - `services/api-go/internal/httpapi/auth.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/api-go/internal/platform/store.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `apps/admin-web/src/adminApi.mjs`
  - `apps/admin-web/src/adminApi.test.mjs`
  - `apps/admin-web/src/adminAudit.mjs`
  - `apps/admin-web/src/adminViews.mjs`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `PROJECT_STATUS.md`
  - `docs/product/recent-progress-roadmap.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `apps/admin-web/README.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260524-101 RBAC 权限变更审计回滚首版

- 日期：2026-05-24
- 结果：在 RBAC 权限变更手动应用之后补齐审计回滚链路。新增 `POST /api/admin/rbac/change-requests/{id}/rollback`，只允许当前仍处于 `applied` 状态的申请回滚到应用前 scopes；服务端禁止申请人直接回滚自己的申请，并在回滚前校验该申请仍是目标角色最新一次生效策略事件，且当前运行时 scopes 仍等于该申请应用后的 scopes，避免覆盖后续策略变更。回滚动作写入 `admin.rbac.change_rolled_back` 审计，payload 记录 change request、目标角色、回滚前 scopes、回滚到 scopes、策略版本和原因。RBAC 申请台账新增 `rolled_back` 状态和 `rolled_back_count`，`api-go` 启动时会按 `admin.rbac.change_applied` 与 `admin.rbac.change_rolled_back` 审计时间顺序重放运行时策略，避免重启后丢失回滚结果。BFF 与 Admin Web 操作目录已接入回滚入口，审计 payload 白名单补充 `rollback_from_scopes` 和 `rollback_to_scopes`。
- 验收证据：
  - `cd services/api-go && go test -count=1 ./internal/httpapi`
  - `npm run test --workspace @infinitech/admin-web`
  - `npm run test --workspace @infinitech/bff`
  - `cd services/api-go && go test -count=1 ./...`
  - `npm run verify:architecture`
  - `npm run verify`
  - `git diff --check`
- 当前边界：这是 RBAC 策略应用后的回滚首版，不是完整权限治理后台。后续审计导出已在 `DONE-20260524-102` 落地，审计留存/告警健康报告已在 `DONE-20260524-103` 落地；仍需字段级/租户级 RBAC、站点/商户数据域、产品化审批队列、菜单按权限隐藏、WORM/冷热归档、真实告警投递和 KMS/链式不可抵赖签名。
- 文件：
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/api-go/internal/platform/store.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `apps/admin-web/src/adminApi.mjs`
  - `apps/admin-web/src/adminApi.test.mjs`
  - `apps/admin-web/src/adminAudit.mjs`
  - `apps/admin-web/src/adminViews.mjs`
  - `apps/admin-web/README.md`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `PROJECT_STATUS.md`
  - `docs/product/recent-progress-roadmap.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260524-102 管理端审计导出首版

- 日期：2026-05-24
- 结果：新增 `GET /api/admin/audit-logs/export`，复用审计查询的 actor/action/target/after/before/limit 筛选，返回 CSV 内容、文件名、生成时间、行数和 content type；导出行为本身写入 `admin.audit_logs.exported` 审计，payload 记录筛选条件、导出格式、行数和生成时间。BFF 已允许代理该接口，Admin Web 操作目录与审计页已接入“导出审计 CSV”，审计结果适配器可识别导出响应。
- 验收证据：
  - `cd services/api-go && go test -count=1 ./internal/httpapi`
  - `npm run test --workspace @infinitech/admin-web`
  - `npm run test --workspace @infinitech/bff`
  - `cd services/api-go && go test -count=1 ./...`
  - `npm run verify:architecture`
  - `npm run verify`
  - `git diff --check`
- 当前边界：这是审计 CSV 导出首版，不是完整审计治理。仍需留存策略、WORM/冷热归档、异常告警、导出审批/水印、字段级/租户级权限、KMS/链式不可抵赖签名和审计策略治理。
- 文件：
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/api-go/internal/platform/store.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `apps/admin-web/src/adminApi.mjs`
  - `apps/admin-web/src/adminApi.test.mjs`
  - `apps/admin-web/src/adminAudit.mjs`
  - `apps/admin-web/src/adminViews.mjs`
  - `apps/admin-web/README.md`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `PROJECT_STATUS.md`
  - `docs/product/recent-progress-roadmap.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260524-103 管理端审计留存/告警健康报告首版

- 日期：2026-05-24
- 结果：新增 `GET /api/admin/audit-logs/retention-report`，默认按 2555 天留存、180 天热存和 500 条完整性抽样生成审计健康报告。报告返回总日志数、最早/最新时间、过期日志、冷归档候选、完整性失败、导出事件、关键动作覆盖、缺失关键动作和 `ok`/`warning`/`critical` 状态；内存 Store 可发现留存过期、冷归档候选、完整性篡改和关键动作缺口，PostgreSQL-backed Store 使用规范化 `audit_logs` 做聚合查询。BFF、Admin Web 操作目录、审计适配器、HTTP 回归测试和架构守卫已接入。
- 验收证据：
  - `cd services/api-go && go test -count=1 ./internal/platform ./internal/httpapi`
  - `npm run test --workspace @infinitech/admin-web`
  - `npm run test --workspace @infinitech/bff`
  - `npm run verify:architecture`
  - `cd services/api-go && go test -count=1 ./...`
  - `npm run verify`
  - `git diff --check`
- 当前边界：这是留存/告警的健康报告首版，不是完整审计治理。后续 `DONE-20260524-104` 已补告警 outbox 投递首版，`DONE-20260524-106` 已补审计归档 worker 首版，`DONE-20260524-107` 已补归档完成回写和记录查询，`DONE-20260524-108` 已补归档对象下载校验/回查首版；仍需生产 WORM bucket 策略、真实告警渠道投递、导出审批/水印、字段级/租户级权限、KMS/链式不可抵赖签名和审计策略治理。
- 文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `apps/admin-web/src/adminApi.mjs`
  - `apps/admin-web/src/adminApi.test.mjs`
  - `apps/admin-web/src/adminAudit.mjs`
  - `apps/admin-web/src/adminViews.mjs`
  - `apps/admin-web/README.md`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `PROJECT_STATUS.md`
  - `docs/product/recent-progress-roadmap.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260524-104 管理端审计留存告警 outbox 投递首版

- 日期：2026-05-24
- 结果：新增 `POST /api/admin/audit-logs/retention-alerts/emit`，复用审计留存/告警健康报告口径，把 critical/warning 告警投递为 `audit.retention_alerts` outbox 事件，并写入 `admin.audit_retention_alerts.emitted` 审计。新增 `audit:write` scope，`security_auditor` 保持只读不可投递；BFF 与 Admin Web 操作目录已接入，`notification-worker` 已订阅 `audit.retention_alerts` 并生成安全告警通知 payload。内存 Store 与 PostgreSQL-backed Store 均支持投递，PostgreSQL 路径把 outbox 事件和审计日志放入同一事务。
- 验收证据：
  - `cd services/api-go && go test -count=1 ./internal/platform ./internal/httpapi`
  - `npm run test --workspace @infinitech/admin-web`
  - `npm run test --workspace @infinitech/bff`
  - `npm run test --workspace @infinitech/notification-worker`
  - `npm run verify:architecture`
  - `cd services/api-go && go test -count=1 ./...`
  - `npm run verify`
  - `git diff --check`
- 当前边界：这是告警可靠事件投递首版，不是完整真实渠道告警。短信、企业微信、电话值班、工单、告警抑制/静默窗口、升级策略、投递回执、生产 WORM bucket 策略和 KMS/链式不可抵赖签名仍待补齐；审计归档 worker 首版已在后续 `DONE-20260524-106` 落地，归档完成回写已在后续 `DONE-20260524-107` 落地，归档对象下载校验已在后续 `DONE-20260524-108` 落地。
- 文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/auth.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `services/notification-worker/src/index.mjs`
  - `services/notification-worker/src/index.test.mjs`
  - `apps/admin-web/src/adminApi.mjs`
  - `apps/admin-web/src/adminApi.test.mjs`
  - `apps/admin-web/src/adminAudit.mjs`
  - `apps/admin-web/src/adminViews.mjs`
  - `apps/admin-web/README.md`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `PROJECT_STATUS.md`
  - `docs/product/recent-progress-roadmap.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260524-105 管理端审计 WORM/冷归档请求首版

- 日期：2026-05-24
- 结果：新增 `POST /api/admin/audit-logs/archive/request`，按热存窗口和 limit 生成冷归档候选 manifest、`sha256:v1` manifest hash、归档路径和 `audit.archive_requested` outbox 事件，并写入 `admin.audit_archive.requested` 审计。内存 Store 与 PostgreSQL-backed Store 均支持该路径，PostgreSQL 路径把 outbox 事件和审计日志放入同一事务；`security_auditor` 仍保持只读，只有 `audit:write` 管理员角色可触发请求。BFF、Admin Web 操作目录和审计适配层已接入；后续 `DONE-20260524-106` 已改为由专用 `audit-archive-worker` 处理该 topic。
- 验收证据：
  - `cd services/api-go && go test -count=1 ./internal/platform ./internal/httpapi`
  - `npm run test --workspace @infinitech/admin-web`
  - `npm run test --workspace @infinitech/bff`
  - `npm run test --workspace @infinitech/outbox-relay-worker`
  - `npm run verify:architecture`
  - `cd services/api-go && go test -count=1 ./...`
  - `npm run verify`
  - `git diff --check`
- 当前边界：这是 WORM/冷归档请求和 manifest 首版，不是实际生产 WORM bucket 策略、保留期删除审批、法律级不可抵赖归档或 KMS/链式签名。后续 `DONE-20260524-106` 已补归档 worker 首版，`DONE-20260524-107` 已补归档完成回写，`DONE-20260524-108` 已补归档对象下载校验/回查；对象存储强制保留策略、保留期删除审批、导出审批/水印和真实告警渠道仍待补。
- 文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `services/outbox-relay-worker/src/index.mjs`
  - `services/outbox-relay-worker/src/index.test.mjs`
  - `apps/admin-web/src/adminApi.mjs`
  - `apps/admin-web/src/adminApi.test.mjs`
  - `apps/admin-web/src/adminAudit.mjs`
  - `apps/admin-web/src/adminViews.mjs`
  - `apps/admin-web/README.md`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `PROJECT_STATUS.md`
  - `docs/product/recent-progress-roadmap.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260524-106 管理端审计归档 worker 首版

- 日期：2026-05-24
- 结果：新增 `services/audit-archive-worker`，直接领取 `audit.archive_requested` outbox 事件，校验后端 `sha256:v1` manifest hash，生成 JSONL 归档文件并上传到配置的对象存储地址；上传时附带 `x-amz-object-lock-mode`、保留期、归档 ID、manifest hash 和内容 hash 元数据，成功后标记 outbox published，失败后标记 failed 并进入 retry/dead-letter 策略。通用 `outbox-relay-worker` 不再 relay `audit.archive_requested`，只保留 `audit.retention_alerts` 等普通事件，避免归档文件未落盘就被错误标记完成。后续 `DONE-20260524-107` 已补完成证据回写。
- 验收证据：
  - `npm run test --workspace @infinitech/audit-archive-worker`
  - `npm run test --workspace @infinitech/outbox-relay-worker`
  - `npm run verify:architecture`
  - `npm run verify`
  - `git diff --check`
- 当前边界：这是审计归档执行器首版，不是完整生产 WORM 治理。后续 `DONE-20260524-107` 已补归档完成回写和记录查询，`DONE-20260524-108` 已补归档对象下载校验/回查首版；仍需补 bucket object-lock 强制策略、保留期删除审批、KMS/Vault 密钥轮换、链式不可抵赖签名、真实存储演练报告和告警渠道回执。
- 文件：
  - `package.json`
  - `services/audit-archive-worker/package.json`
  - `services/audit-archive-worker/src/index.mjs`
  - `services/audit-archive-worker/src/index.test.mjs`
  - `services/outbox-relay-worker/src/index.mjs`
  - `services/outbox-relay-worker/src/index.test.mjs`
  - `infra/docker/compose.yml`
  - `infra/k8s/base/app-stack.yaml`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `PROJECT_STATUS.md`
  - `docs/product/recent-progress-roadmap.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260524-107 审计归档完成回写与记录查询首版

- 日期：2026-05-24
- 结果：新增 `POST /api/admin/audit-logs/archive/complete`，归档 worker 在对象上传成功后回写归档完成证据并写入 `admin.audit_archive.completed` 审计；新增 `GET /api/admin/audit-logs/archive/records`，可从审计账本重建已完成归档记录。完成证据包含 archive id、storage key、manifest hash、content hash、bytes、object lock mode、retain until、uploaded at 和 outbox event id。`audit-archive-worker` 已改为先调用完成回写，再把 outbox 标记 published；BFF 和 Admin Web 操作目录已接入归档记录查询。
- 验收证据：
  - `cd services/api-go && go test ./internal/platform ./internal/httpapi`
  - `npm run test --workspace @infinitech/audit-archive-worker`
  - `npm run test --workspace @infinitech/bff`
  - `npm run test --workspace @infinitech/admin-web`
  - `npm run verify:architecture`
- 当前边界：这是归档完成证据账本与查询首版。后续 `DONE-20260524-108` 已补对象下载校验/回查首版，`DONE-20260524-109` 已补校验历史查询首版，`DONE-20260524-110` 已补审计中心可视化面板首版；仍需生产 WORM bucket 策略、保留期删除审批、KMS/链式不可抵赖签名和真实演练报告。
- 文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `services/audit-archive-worker/src/index.mjs`
  - `services/audit-archive-worker/src/index.test.mjs`
  - `apps/admin-web/src/adminApi.mjs`
  - `apps/admin-web/src/adminApi.test.mjs`
  - `apps/admin-web/src/adminAudit.mjs`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `PROJECT_STATUS.md`
  - `docs/product/recent-progress-roadmap.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260524-108 审计归档对象下载校验与回查首版

- 日期：2026-05-24
- 结果：新增 `POST /api/admin/audit-logs/archive/verify`，后台可根据已完成归档记录读取配置的 `AUDIT_ARCHIVE_DOWNLOAD_BASE_URL` 归档对象，校验 JSONL content hash、manifest header 中的 archive id/manifest hash、完成记录 bytes 和 manifest entry 数，并把结果写入 `admin.audit_archive.verified` 审计。`security_auditor` 可触发该只读回查；BFF 与 Admin Web 操作目录已接入；`ObjectStorageConfig` 新增归档下载地址、最大下载字节数和下载超时配置。
- 验收证据：
  - `cd services/api-go && go test ./internal/platform ./internal/httpapi`
  - `npm run test --workspace @infinitech/bff`
  - `npm run test --workspace @infinitech/admin-web`
  - `npm run verify:architecture`
- 当前边界：这是归档对象下载校验/回查首版，不是生产 WORM bucket 强制策略、校验历史详情页、保留期删除审批、KMS/链式不可抵赖签名或真实存储演练报告。
- 文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/object_storage.go`
  - `services/api-go/cmd/api/main.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `apps/admin-web/src/adminApi.mjs`
  - `apps/admin-web/src/adminApi.test.mjs`
  - `apps/admin-web/src/adminAudit.mjs`
  - `apps/admin-web/src/adminViews.mjs`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `PROJECT_STATUS.md`
  - `docs/product/recent-progress-roadmap.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260524-109 审计归档校验历史查询首版

- 日期：2026-05-24
- 结果：新增 `GET /api/admin/audit-logs/archive/verifications`，从 `admin.audit_archive.verified` 审计账本重建归档校验历史，支持按归档 ID、状态、after/before 时间范围和 limit 查询。内存 Store 与 PostgreSQL-backed Store 共用同一套审计 payload 反序列化逻辑；BFF 已放行该路由；Admin Web 操作目录新增“归档校验历史”。
- 验收证据：
  - `cd services/api-go && go test ./internal/platform ./internal/httpapi`
  - `npm run test --workspace @infinitech/bff`
  - `npm run test --workspace @infinitech/admin-web`
  - `npm run verify:architecture`
- 当前边界：这是校验历史查询首版，后续 `DONE-20260524-110` 已补审计中心可视化面板首版；仍不是生产 WORM bucket 强制策略、保留期删除审批、KMS/链式不可抵赖签名或真实存储演练报告。
- 文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `apps/admin-web/src/adminApi.mjs`
  - `apps/admin-web/src/adminApi.test.mjs`
  - `apps/admin-web/src/adminAudit.mjs`
  - `apps/admin-web/src/adminViews.mjs`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `PROJECT_STATUS.md`
  - `docs/product/recent-progress-roadmap.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260524-110 审计归档校验历史可视化面板首版

- 日期：2026-05-24
- 结果：Admin Web 审计检索页新增“归档校验历史”面板，可按归档 ID、状态和条数查询 `/api/admin/audit-logs/archive/verifications`，并展示归档 ID、storage key、状态、校验时间、manifest/content hash、bytes/logs、匹配摘要和原始详情。
- 验收证据：
  - `npm run test --workspace @infinitech/admin-web`
  - `npm run verify:architecture`
  - 本地浏览器打开 `http://127.0.0.1:4174/apps/admin-web/index.html`，进入“审计检索”后可见“归档校验历史”“查询历史”和示例归档 `audit_archive_1`，浏览器 console error 为 0。
- 当前边界：这是管理端可视化回看首版，不是生产 WORM bucket 强制策略、保留期删除审批、KMS/链式不可抵赖签名或真实存储演练报告。
- 文件：
  - `apps/admin-web/src/adminAudit.mjs`
  - `apps/admin-web/src/main.js`
  - `apps/admin-web/src/styles.css`
  - `apps/admin-web/src/adminApi.test.mjs`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `PROJECT_STATUS.md`
  - `docs/product/recent-progress-roadmap.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260524-111 管理端 P0 业务详情面板首版

- 日期：2026-05-24
- 结果：Admin Web 新增 `adminDetails` 业务详情适配层，订单、售后、商户、骑手/站长、骑手绩效、派单、审计、退款策略和权限治理表格行可打开详情面板。详情面板展示当前行字段、模块化核查清单和下一步操作按钮，可把运营人员直接带到订单状态补偿、审计检索、Outbox 事件、对象清理候选、退款策略、RBAC 策略/申请等已有操作入口。
- 验收证据：
  - `npm run test --workspace @infinitech/admin-web`
  - `npm run verify:architecture`
- 当前边界：这是管理端 P0 行详情与下一步动作首版，仍不是完整后端明细接口、业务分页筛选、资质/售后审核表单、二次确认、操作结果追踪或字段级/租户级 RBAC。
- 文件：
  - `apps/admin-web/src/adminDetails.mjs`
  - `apps/admin-web/src/main.js`
  - `apps/admin-web/src/styles.css`
  - `apps/admin-web/src/adminApi.test.mjs`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `PROJECT_STATUS.md`
  - `docs/product/recent-progress-roadmap.md`
  - `apps/admin-web/README.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260524-112 管理端高风险操作二次确认与结果追踪首版

- 日期：2026-05-24
- 结果：Admin Web 新增 `adminOperations` 高风险操作适配层，邀约、退款策略、审计导出/告警/归档、RBAC 变更、Outbox 批量恢复和订单状态补偿执行前会进入 `pending_confirmation` 确认面板。确认面板展示方法、路径、区域、风险原因和提交参数快照；确认执行后会在操作台保留最近结果，展示成功/失败、状态、请求路径和返回消息。
- 验收证据：
  - `npm run test --workspace @infinitech/admin-web`
  - `npm run verify:architecture`
- 当前边界：这是管理端高危操作防误触和最近结果追踪首版，仍不是后端任务级异步进度、失败重试回放中心、审核表单全链路或生产事故工单闭环。
- 文件：
  - `apps/admin-web/src/adminOperations.mjs`
  - `apps/admin-web/src/main.js`
  - `apps/admin-web/src/styles.css`
  - `apps/admin-web/src/adminApi.test.mjs`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `PROJECT_STATUS.md`
  - `docs/product/recent-progress-roadmap.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `apps/admin-web/README.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260524-113 管理端失败回放入口首版

- 日期：2026-05-24
- 结果：Admin Web 操作结果追踪中的失败记录新增“重试”入口。重试会恢复原操作和参数快照；高风险操作不会直接重放，而是重新进入 `pending_confirmation` 二次确认面板，避免绕过人工确认。
- 验收证据：
  - `npm run test --workspace @infinitech/admin-web`
  - `npm run verify:architecture`
  - 本地浏览器打开 `http://127.0.0.1:4174/apps/admin-web/index.html`，确认失败记录出现“重试”，点击后恢复“保存退款策略”并重新进入“需二次确认”面板，参数快照保留，console error 为 0。
- 当前边界：这是最近失败结果的一键回放入口首版，仍不是后端任务级异步进度、批量失败重试队列、事故工单或按审计事件重放的生产恢复中心。
- 文件：
  - `apps/admin-web/src/adminOperations.mjs`
  - `apps/admin-web/src/main.js`
  - `apps/admin-web/src/styles.css`
  - `apps/admin-web/src/adminApi.test.mjs`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `PROJECT_STATUS.md`
  - `docs/product/recent-progress-roadmap.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `apps/admin-web/README.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260524-114 管理端 P0 业务筛选分页首版

- 日期：2026-05-24
- 结果：Admin Web 新增 `adminTable` 表格适配层，P0 业务视图支持关键字筛选、每页 4/10/20 条和上一页/下一页控制。筛选分页会保留原始行索引，详情按钮仍打开正确业务行。
- 验收证据：
  - `npm run test --workspace @infinitech/admin-web`
  - `npm run verify:architecture`
  - 本地浏览器打开 `http://127.0.0.1:4174/apps/admin-web/index.html`，进入“订单监控”后筛选“康宁”只剩 1 条，分页计数显示 `第 1/1 页 · 1 条`，清除后恢复全表，console error 为 0。
- 当前边界：这是前端 P0 业务视图筛选分页首版，仍不是后端分页游标、服务端复杂筛选、导出任务或审核表单全链路。
- 文件：
  - `apps/admin-web/src/adminTable.mjs`
  - `apps/admin-web/src/main.js`
  - `apps/admin-web/src/styles.css`
  - `apps/admin-web/src/adminApi.test.mjs`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `PROJECT_STATUS.md`
  - `docs/product/recent-progress-roadmap.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `apps/admin-web/README.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260525-115 管理端售后审核表单首版

- 日期：2026-05-25
- 结果：Admin Web 操作目录新增 `after-sales-review`，连接现有 `POST /api/after-sales/{requestID}/review` 原子审计路径。售后审核模块和售后详情抽屉可预填工单 ID、审核结果、审核原因、退款去向和退款幂等键；该操作已纳入高风险操作，执行前进入二次确认，确认面板展示方法、路径、区域和参数快照。
- 验收证据：
  - `npm run test --workspace @infinitech/admin-web`
  - `npm run verify:architecture`
  - 本地浏览器打开 `http://127.0.0.1:4174/apps/admin-web/index.html`，进入“售后审核”后点击“审核售后”，确认表单预填 `asr_231`、`approve`、`证据核验通过`、`balance` 和 `after_sales:asr_231`；点击“进入确认”后出现“需二次确认”面板，路径为 `/api/after-sales/:request_id/review`，console error 为 0。打开首行详情后点击详情内“审核售后”同样可预填工单参数。
- 当前边界：这是前端售后审核表单和二次确认首版，复用后端已有原子审计路径；仍不是后端售后明细接口、证据时间线辅助判断、批量审核、资质审核表单、退款/outbox 运维表单或字段级/租户级 RBAC。
- 文件：
  - `apps/admin-web/src/adminApi.mjs`
  - `apps/admin-web/src/adminOperations.mjs`
  - `apps/admin-web/src/adminViews.mjs`
  - `apps/admin-web/src/adminDetails.mjs`
  - `apps/admin-web/src/adminApi.test.mjs`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `PROJECT_STATUS.md`
  - `docs/product/recent-progress-roadmap.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `apps/admin-web/README.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260525-116 管理端订单退款表单首版

- 日期：2026-05-25
- 结果：Admin Web 操作目录新增 `order-refund`，连接现有 `POST /api/orders/{orderID}/refund` 原子审计路径。订单监控模块和订单详情抽屉可预填订单 ID、退款原因、退款幂等键、可选退款金额和退款去向；该操作已纳入高风险操作，执行前进入二次确认，确认面板展示方法、路径、区域和参数快照。
- 验收证据：
  - `npm run test --workspace @infinitech/admin-web`
  - `npm run verify:architecture`
  - 本地浏览器打开 `http://127.0.0.1:4174/apps/admin-web/index.html`，进入“订单监控”后点击“订单退款”，确认表单预填 `ord_10031`、`客服确认退款`、`refund_ord_10031` 和 `balance`；点击“进入确认”后出现“需二次确认”面板，路径为 `/api/orders/:order_id/refund`，console error 为 0。打开首行详情后点击详情内“订单退款”同样可预填订单退款参数。
- 当前边界：这是前端订单退款表单和二次确认首版，复用后端已有原子审计路径；仍不是订单后端明细接口、退款审批流、对账工作台、真实微信原路退款 API 闭环或字段级/租户级 RBAC。
- 文件：
  - `apps/admin-web/src/adminApi.mjs`
  - `apps/admin-web/src/adminOperations.mjs`
  - `apps/admin-web/src/adminViews.mjs`
  - `apps/admin-web/src/adminDetails.mjs`
  - `apps/admin-web/src/adminApi.test.mjs`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `PROJECT_STATUS.md`
  - `docs/product/recent-progress-roadmap.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `apps/admin-web/README.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260525-117 管理端 Outbox 单事件恢复表单首版

- 日期：2026-05-25
- 结果：Admin Web 操作目录新增 `outbox-replay-event`，连接现有 `POST /api/admin/outbox/events/{eventID}/replay` 原子审计路径。运营首页和 Outbox 队列详情抽屉可预填事件 ID；该操作已纳入高风险操作，执行前进入二次确认，确认面板展示方法、路径、区域、风险原因和参数快照，避免运维误触可靠投递重放。
- 验收证据：
  - `npm run test --workspace @infinitech/admin-web`
  - `npm run verify:architecture`
  - 本地浏览器打开 `http://127.0.0.1:4174/apps/admin-web/index.html`，在“运营首页”点击“恢复单个 Outbox”，确认操作台选择“恢复单个 Outbox”、`event_id` 预填 `obe_1`、按钮显示“进入确认”；点击后出现“需二次确认”面板，路径为 `/api/admin/outbox/events/:event_id/replay`，风险原因为“恢复单个 outbox 事件，可能触发一次可靠投递重试”，浏览器日志为空。打开“Outbox”队列详情后点击详情内“恢复单个 Outbox”同样可预填 `obe_1`。
- 当前边界：这是前端 Outbox 单事件恢复表单和二次确认首版，复用后端已有原子审计路径；仍不是完整 outbox 事故中心、broker 运维台、dead-letter 分诊自动化、生产告警联动或多租户/字段级权限闭环。
- 文件：
  - `apps/admin-web/src/adminApi.mjs`
  - `apps/admin-web/src/adminOperations.mjs`
  - `apps/admin-web/src/adminViews.mjs`
  - `apps/admin-web/src/adminDetails.mjs`
  - `apps/admin-web/src/adminApi.test.mjs`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `PROJECT_STATUS.md`
  - `docs/product/recent-progress-roadmap.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `apps/admin-web/README.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260525-118 管理端 Outbox 发布/失败人工处置表单首版

- 日期：2026-05-25
- 结果：Admin Web 操作目录新增 `outbox-mark-failed` 和 `outbox-mark-published`，连接现有 `POST /api/admin/outbox/events/{eventID}/failed` 与 `POST /api/admin/outbox/events/{eventID}/published` 原子审计路径。运营首页和 Outbox 队列详情抽屉可预填事件 ID、失败原因、重试延迟和最大尝试次数；两个动作均纳入高风险操作，执行前进入二次确认，确认面板展示方法、路径、区域、风险原因和参数快照。
- 验收证据：
  - `npm run test --workspace @infinitech/admin-web`
  - `npm run verify:architecture`
  - 本地浏览器打开 `http://127.0.0.1:4174/apps/admin-web/index.html`，在“运营首页”点击“标记 Outbox 失败”，确认表单预填 `obe_1`、`relay down`、`120`、`10`，点击“进入确认”后出现“需二次确认”面板，路径为 `/api/admin/outbox/events/:event_id/failed`。点击“标记 Outbox 已发布”后同样进入二次确认，路径为 `/api/admin/outbox/events/:event_id/published`。打开“Outbox”队列详情后确认详情内存在“标记 Outbox 失败”和“标记 Outbox 已发布”，浏览器日志为空。
- 当前边界：这是前端 Outbox 人工 ACK/FAIL 处置表单和二次确认首版，复用后端已有原子审计路径；仍不是完整 outbox 事故中心、claim/lease 运营表单、broker 运维台、dead-letter 分诊自动化、生产告警联动或多租户/字段级权限闭环。
- 文件：
  - `apps/admin-web/src/adminApi.mjs`
  - `apps/admin-web/src/adminOperations.mjs`
  - `apps/admin-web/src/adminViews.mjs`
  - `apps/admin-web/src/adminDetails.mjs`
  - `apps/admin-web/src/adminApi.test.mjs`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `PROJECT_STATUS.md`
  - `docs/product/recent-progress-roadmap.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `apps/admin-web/README.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260525-119 管理端 Outbox 领取/续租表单首版

- 日期：2026-05-25
- 结果：Admin Web 操作目录新增 `outbox-claim-events` 和 `outbox-renew-lease`，连接现有 `POST /api/admin/outbox/events/claim` 与 `POST /api/admin/outbox/events/{eventID}/lease/renew` 原子审计路径。运营首页和 Outbox 队列详情抽屉可预填 topic、limit、lease owner、lease seconds 和事件 ID；两个动作均纳入高风险操作，执行前进入二次确认，确认面板展示方法、路径、区域、风险原因和参数快照。
- 验收证据：
  - `npm run test --workspace @infinitech/admin-web`
  - `npm run verify:architecture`
  - 本地浏览器打开 `http://127.0.0.1:4174/apps/admin-web/index.html`，在“运营首页”点击“领取 Outbox 租约”，确认表单预填 `order.paid`、`10`、`relay-admin`、`60`，点击“进入确认”后出现“需二次确认”面板，路径为 `/api/admin/outbox/events/claim`。点击“续租 Outbox 租约”后确认表单预填 `obe_1`、`relay-admin`、`60`，二次确认路径为 `/api/admin/outbox/events/:event_id/lease/renew`。打开“Outbox”队列详情后确认详情内存在“领取 Outbox 租约”和“续租 Outbox 租约”，浏览器日志为空。
- 当前边界：这是前端 Outbox 租约领取/续租处置表单和二次确认首版，复用后端已有原子审计路径；仍不是完整 outbox 事故中心、broker 运维台、dead-letter 分诊自动化、生产告警联动或多租户/字段级权限闭环。
- 文件：
  - `apps/admin-web/src/adminApi.mjs`
  - `apps/admin-web/src/adminOperations.mjs`
  - `apps/admin-web/src/adminViews.mjs`
  - `apps/admin-web/src/adminDetails.mjs`
  - `apps/admin-web/src/adminApi.test.mjs`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `PROJECT_STATUS.md`
  - `docs/product/recent-progress-roadmap.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `apps/admin-web/README.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260525-120 管理端 Outbox 死信分诊/解封表单首版

- 日期：2026-05-25
- 结果：Admin Web 操作目录新增 `outbox-dead-letter-triage` 和 `outbox-release-dead-letter`，复用现有 `GET /api/admin/outbox/events?status=dead_letter` 与 `POST /api/admin/outbox/events/{eventID}/replay` 原子审计路径。运营首页、今日队列和 Outbox 队列详情抽屉可预填 `topic=order.paid`、`status=dead_letter`、`limit=20` 和死信事件 ID；死信解封已纳入高风险操作，执行前进入二次确认，确认面板展示方法、路径、区域、风险原因和参数快照。
- 验收证据：
  - `npm run test --workspace @infinitech/admin-web`
  - `npm run verify:architecture`
  - 本地浏览器打开 `http://127.0.0.1:4174/apps/admin-web/index.html`，在“运营首页”点击“Outbox 死信分诊”，确认操作台选择“Outbox 死信分诊”、`topic` 预填 `order.paid`、`status` 预填 `dead_letter`、`limit` 预填 `20`，按钮显示“执行”且不出现二次确认。点击“解封 Outbox 死信”后确认表单预填 `obe_dead_1`，按钮显示“进入确认”；点击后出现“需二次确认”面板，路径为 `/api/admin/outbox/events/:event_id/replay`，风险原因为“解封 outbox dead-letter 事件，可能重新触发可靠投递”。打开“Outbox”队列详情后确认详情内存在“分诊 Outbox 死信”和“解封 Outbox 死信”，浏览器日志为空。
- 当前边界：这是前端 Outbox 死信分诊和解封确认首版，复用后端已有查询/replay 原子审计路径；仍不是完整 outbox 事故中心、broker 运维台、毒消息根因分析、生产告警联动或多租户/字段级权限闭环。
- 文件：
  - `apps/admin-web/src/adminApi.mjs`
  - `apps/admin-web/src/adminOperations.mjs`
  - `apps/admin-web/src/adminViews.mjs`
  - `apps/admin-web/src/adminDetails.mjs`
  - `apps/admin-web/src/config.mjs`
  - `apps/admin-web/src/adminApi.test.mjs`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `PROJECT_STATUS.md`
  - `docs/product/recent-progress-roadmap.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `apps/admin-web/README.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260525-121 商户资质审核后端与管理端表单首版

- 日期：2026-05-25
- 结果：商户资质从“上传即通过”改为待审闭环，`POST /api/merchant/qualifications` 只写入 `pending_review`，营业执照/健康证必须经后台 `POST /api/admin/merchant-qualifications/{qualificationID}/review` 审核通过后才计入商户接单资格。新增 `merchant:qualification_review` scope 和 `CanReviewMerchantQualifications` 权限边界，`ops_admin` 可审核、`support_admin` 不可审核；审核写入 `admin.merchant_qualification.reviewed` 审计，PostgreSQL-backed Store 在同一事务内更新资质、审计和快照。BFF 已允许该后台路径，Admin Web 商户模块和详情抽屉已新增“审核商户资质”表单，并纳入高风险二次确认。
- 验收证据：
  - `go test -count=1 ./internal/platform -run 'TestMerchantInviteRegistrationAndQualificationGate|TestReviewMerchantQualification'`
  - `go test -count=1 ./internal/httpapi -run 'TestMerchantInviteRegisterAndQualificationHTTPFlow|TestAdminMerchantQualificationReviewHTTPUsesAtomicAuditRepositoryPath|TestBackofficeRBAC'`
  - `npm test --workspace @infinitech/bff`
  - `npm run test --workspace @infinitech/admin-web`
  - `npm run verify:architecture`
  - `npm run verify`
  - `git diff --check`
  - `cd services/api-go && go test -count=1 ./...`
  - 本地浏览器打开 `http://127.0.0.1:4174/apps/admin-web/index.html`，进入“商户资质”模块后点击“审核商户资质”，确认表单预填 `qualification_id=mq_merchant_19_health_certificate`、`merchant_id=merchant_19`、`decision=approve`、`reason=资质原件核验通过`，按钮显示“进入确认”；进入确认后展示路径 `/api/admin/merchant-qualifications/:qualification_id/review` 和风险原因“审核商户资质会改变商户接单资格和准入风控状态”。打开 `merchant_19` 行详情后，详情抽屉存在“审核商户资质”动作，点击后可回填同一组审核参数，浏览器日志为空。静态服务器已关闭。
- 当前边界：这是商户主体资质审核和接单门槛的首版，不是完整资质中心；仍需补资质图片真实对象预览、OCR/人工复核队列、药房/医务室专项资质、员工健康证批量审核、过期前通知、字段级/租户级权限和审核 SLA/申诉流程。
- 文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/auth.go`
  - `services/api-go/internal/httpapi/auth_test.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `apps/admin-web/src/adminApi.mjs`
  - `apps/admin-web/src/adminOperations.mjs`
  - `apps/admin-web/src/adminViews.mjs`
  - `apps/admin-web/src/adminDetails.mjs`
  - `apps/admin-web/src/config.mjs`
  - `apps/admin-web/src/adminApi.test.mjs`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `apps/admin-web/README.md`
  - `PROJECT_STATUS.md`
  - `docs/product/recent-progress-roadmap.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260525-122 Outbox 单事件事故辅助明细首版

- 日期：2026-05-25
- 结果：新增 `GET /api/admin/outbox/events/{eventID}` 只读明细接口，由 `outbox:read` 权限守护，内存 Store 与 PostgreSQL-backed Store 均返回一致的事故辅助信息：事件状态、ready/blocked/lease 信号、retry/lease 剩余秒数、payload 摘要、关联业务目标、最近 outbox 审计、推荐下一步操作和人工处置核查清单。BFF 已允许该路径，Admin Web 操作目录新增 `outbox-event-detail`，运营首页和 Outbox 队列详情抽屉可预填 `event_id` 查看明细，再转入恢复、领取、续租或审计。
- 验收证据：
  - `go test -count=1 ./internal/platform -run 'TestOutboxEventDetailBuildsIncidentAssist|TestOutboxFailedBackoffAndPublishedAck'`
  - `go test -count=1 ./internal/httpapi -run 'TestAdminOutboxEventDetailHTTPFlow|TestAdminReplayOutboxEventsHTTPFlow'`
  - `npm test --workspace @infinitech/bff`
  - `npm test --workspace @infinitech/admin-web`
  - `npm run verify:architecture`
- 当前边界：这是 Outbox 单事件事故辅助明细首版，仍不是完整事故中心；后续还需要接真实 broker 指标、告警事件联动、死信根因分类、自动化 runbook、跨租户字段级权限和生产值班截图归档。
- 文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `apps/admin-web/src/adminApi.mjs`
  - `apps/admin-web/src/adminViews.mjs`
  - `apps/admin-web/src/adminDetails.mjs`
  - `apps/admin-web/src/adminApi.test.mjs`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `apps/admin-web/README.md`
  - `PROJECT_STATUS.md`
  - `docs/product/recent-progress-roadmap.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260525-123 商户资质待审列表与明细接口首版

- 日期：2026-05-25
- 结果：新增 `GET /api/admin/merchant-qualifications` 和 `GET /api/admin/merchant-qualifications/{qualificationID}`，复用 `merchant:qualification_review` 权限，运营可按状态/商户/资质类型查询待审资质，并打开单条资质明细查看商户账号、店铺、保证金、缺失资质、接单资格、事故状态、最近 `admin.merchant_qualification.reviewed` 审计、推荐下一步动作和审核核查清单。BFF 已允许列表/明细路径，Admin Web 操作目录新增 `merchant-qualifications` 与 `merchant-qualification-detail`，运营首页、商户模块和商户详情抽屉可先查看待审上下文，再进入资质审核二次确认。
- 验收证据：
  - `go test -count=1 ./internal/platform -run 'TestAdminMerchantQualification|TestMerchantInviteRegistrationAndQualificationGate'`
  - `go test -count=1 ./internal/httpapi -run 'TestAdminMerchantQualification'`
  - `npm test --workspace @infinitech/bff -- --test-name-pattern "proxy"`
  - `npm test --workspace @infinitech/admin-web -- --test-name-pattern "operation catalog|P0 business views|business detail|admin request builder|snapshot adapter"`
  - `git diff --check`
  - `cd services/api-go && go test -count=1 ./...`
  - `npm run verify`
  - 本地浏览器打开 `http://127.0.0.1:4174/apps/admin-web/index.html`，确认运营首页展示“商户资质待审列表”“商户资质明细”操作；切换到 `merchant-qualifications` 后默认 `status=pending_review`、`limit=20`，切换到 `merchant-qualification-detail` 后默认 `qualification_id=mq_merchant_19_health_certificate`、`audit_limit=20`，浏览器 console error 为空。静态服务器已关闭并确认端口不可连接。
- 当前边界：这是资质审核工作台发现/明细首版，仍不是完整招商 CRM；后续还需要补真实文件预览/ OCR、法人主体一致性校验、过期提醒、补件消息触达、字段级/租户级权限和生产归档证据。
- 文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `apps/admin-web/src/adminApi.mjs`
  - `apps/admin-web/src/adminViews.mjs`
  - `apps/admin-web/src/adminDetails.mjs`
  - `apps/admin-web/src/adminSnapshot.mjs`
  - `apps/admin-web/src/adminApi.test.mjs`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `apps/admin-web/README.md`
  - `PROJECT_STATUS.md`
  - `docs/product/recent-progress-roadmap.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260525-146 通知偏好变更灰度应用首版

- 日期：2026-05-25
- 结果：通知偏好治理从“审批后全量应用”继续推进到“审批时固化灰度范围、应用时按范围落账”。后端的通知偏好变更申请新增 `rollout` 策略，支持 `all`、`target_ids` 和 deterministic `percentage` 三种模式，并允许通过 `max_targets` 控制首批目标范围；申请审计写入 rollout、requested count 和 preference keys，审批后不能在应用阶段临时扩大范围。应用接口会按 rollout 计算实际写入的偏好，未命中的 preference key 不进入批量保存、不触发 `notification.preferences_changed`，并在 `admin.notification_preferences.change_applied` 审计中记录 applied/skipped preference keys、applied/skipped count 和 rollout 参数。申请台账可从审计日志恢复 rollout 和 applied/skipped 范围；Admin Web 提交通知偏好变更表单已新增灰度策略 JSON，BFF 和架构守卫覆盖透传与防退化。
- 验收证据：
  - `cd services/api-go && go test -count=1 ./internal/httpapi -run TestMerchantNotificationsHTTPFlow`
  - `npm test --workspace @infinitech/admin-web`
  - `npm test --workspace @infinitech/bff`
  - `npm run verify:architecture`
  - `git diff --check`
  - `npm run verify`
- 当前边界：这是通知偏好变更的首个可审计灰度应用能力，不是完整通知策略中心。策略升级批次、策略回滚、真实渠道账号、模板审批、provider sandbox/生产字段联调、跨端消息中心和生产告警联动仍待补。
- 文件：
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/api-go/internal/platform/store.go`
  - `services/bff/src/runtime.test.mjs`
  - `apps/admin-web/src/adminApi.mjs`
  - `apps/admin-web/src/adminApi.test.mjs`
  - `apps/admin-web/src/adminOperations.mjs`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `apps/admin-web/README.md`
  - `PROJECT_STATUS.md`
  - `docs/product/recent-progress-roadmap.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260525-145 通知偏好变更审批与应用首版

- 日期：2026-05-25
- 结果：通知偏好链路从“可批量保存多条偏好策略”继续推进到“关键偏好变更必须走申请、审批、手动应用闭环”。后端新增 `GET/POST /api/admin/notification-preferences/change-requests`，申请只写入 `admin.notification_preferences.change_requested` 审计，不直接修改偏好账本；新增 `POST /api/admin/notification-preferences/change-requests/{changeRequestID}/review`，支持另一名通知管理员审批或驳回并禁止申请人自审，审批结果写入 `admin.notification_preferences.change_reviewed`；新增 `POST /api/admin/notification-preferences/change-requests/{changeRequestID}/apply`，只允许 `approved` 申请进入批量保存原子路径，应用时同事务写入偏好、`notification.preferences_changed` outbox 和 `admin.notification_preferences.change_applied` 审计，申请人不能自己应用，驳回申请不能应用。申请台账可从审计日志重建并按 `pending_approval`、`approved`、`rejected`、`applied` 查询。BFF 已代理申请/审批/应用路由，Admin Web 通知运营操作目录已接入查询、提交、审批和应用动作，三类写动作均进入高风险二次确认。
- 验收证据：
  - `cd services/api-go && go test -count=1 ./internal/httpapi -run TestMerchantNotificationsHTTPFlow`
  - `npm test --workspace @infinitech/admin-web`
  - `npm test --workspace @infinitech/bff`
  - `npm run verify:architecture`
  - `git diff --check`
  - `npm run verify`
- 当前边界：这是通知偏好治理的审批与应用首版，不是完整通知策略治理中心。灰度应用已由后续 `DONE-20260525-146` 承接；真实渠道账号、模板审批、provider sandbox/生产字段联调、升级策略和跨端消息中心仍待补。
- 文件：
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `apps/admin-web/src/adminApi.mjs`
  - `apps/admin-web/src/adminApi.test.mjs`
  - `apps/admin-web/src/adminOperations.mjs`
  - `apps/admin-web/src/adminViews.mjs`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `apps/admin-web/README.md`
  - `PROJECT_STATUS.md`
  - `docs/product/recent-progress-roadmap.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260525-144 通知偏好批量保存与策略审计首版

- 日期：2026-05-25
- 结果：通知偏好链路从“单条偏好保存与事件失效”继续推进到“运营可批量保存多条偏好策略且证据闭环”。后端新增 `SaveNotificationPreferenceBatchRequest` 与 `NotificationPreferenceBatchSaveResult`，`POST /api/admin/notification-preferences/batch` 支持一次保存最多 50 条偏好策略并要求变更原因；同批重复 preference key 会被拒绝。内存 Store 会逐条保存偏好并生成 `notification.preferences_changed` outbox，批量动作写入 `admin.notification_preferences.batch_saved` 审计。PostgreSQL-backed Store 会在同一事务内 upsert 多条 `platform_notification_preferences`、插入对应 outbox 事件并写入批量审计。BFF 已允许该路由，Admin Web 通知运营页新增“批量保存通知偏好”操作，`preferences` JSON 与变更原因进入高风险二次确认。
- 验收证据：
  - `cd services/api-go && go test -count=1 ./internal/platform -run 'TestMerchantNotificationCenterStoresAndReadsQualificationReview'`
  - `cd services/api-go && go test -count=1 ./internal/httpapi -run 'TestMerchantNotificationsHTTPFlow'`
  - `npm test --workspace @infinitech/admin-web`
  - `npm test --workspace @infinitech/bff`
  - `npm run verify:architecture`
  - `git diff --check`
  - `npm run verify`
- 当前边界：这是批量策略保存与审计首版，不是完整通知策略治理中心。通知偏好审批应用已由后续 `DONE-20260525-145` 承接；策略灰度、真实渠道账号、模板审批、provider sandbox/生产字段联调、升级策略和跨端消息中心仍待补。
- 文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `apps/admin-web/src/adminApi.mjs`
  - `apps/admin-web/src/adminApi.test.mjs`
  - `apps/admin-web/src/adminOperations.mjs`
  - `apps/admin-web/src/adminViews.mjs`
  - `apps/admin-web/README.md`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `PROJECT_STATUS.md`
  - `docs/product/recent-progress-roadmap.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260525-143 通知偏好变更事件与 worker 主动失效首版

- 日期：2026-05-25
- 结果：通知偏好链路从“worker 有短 TTL 缓存”继续推进到“偏好变更可通过可靠事件驱动缓存失效”。保存通知偏好时，内存 Store 与 PostgreSQL-backed Store 都会生成 `notification.preferences_changed` outbox 事件，payload 携带 `preference_key`、`preference_keys`、目标角色/ID、通知类型、渠道和静默窗口；PostgreSQL 保存路径会在同一事务内 upsert 偏好并插入 outbox。outbox relay 默认 topic、Docker Compose 和 K8s 配置已加入该事件。`notification-worker` 订阅 `notification.preferences_changed`，新增 `notificationPreferenceInvalidationKeys` 与 `invalidateNotificationPreferenceCache`，消费该事件时只删除对应 resolver 缓存 key，不会误创建站内通知。
- 验收证据：
  - `npm test --workspace @infinitech/notification-worker`
  - `npm test --workspace @infinitech/outbox-relay-worker`
  - `cd services/api-go && go test -count=1 ./internal/httpapi -run 'TestMerchantNotificationsHTTPFlow'`
  - `cd services/api-go && go test -count=1 ./internal/platform`
  - `npm run verify:architecture`
  - `git diff --check`
  - `npm run verify`
- 当前边界：这是偏好变更可靠事件和 worker 本地缓存主动失效首版，不是完整通知策略治理中心。批量策略已由后续 `DONE-20260525-144` 承接，通知偏好审批应用已由后续 `DONE-20260525-145` 承接；策略灰度、真实渠道账号、模板审批、provider sandbox/生产字段联调、升级策略和跨端消息中心仍待补。
- 文件：
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/notification-worker/src/index.mjs`
  - `services/notification-worker/src/index.test.mjs`
  - `services/outbox-relay-worker/src/index.mjs`
  - `services/outbox-relay-worker/src/index.test.mjs`
  - `infra/docker/compose.yml`
  - `infra/k8s/base/app-stack.yaml`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `PROJECT_STATUS.md`
  - `docs/product/recent-progress-roadmap.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260525-142 通知 worker 偏好缓存与失败关闭首版

- 日期：2026-05-25
- 结果：通知链路从“worker 每次投递前精确读取后端偏好”继续推进到“高频投递可控缓存、偏好服务抖动不绕过用户选择”。`createNotificationPreferenceResolver` 新增按 preference key 的内存缓存，默认 `NOTIFICATION_PREFERENCE_CACHE_TTL_MS=30000`、`NOTIFICATION_PREFERENCE_CACHE_STALE_MS=300000`、`NOTIFICATION_PREFERENCE_CACHE_MAX_KEYS=500`；新鲜缓存命中时不再重复请求 `/api/admin/notification-preferences`，TTL 过期后会刷新，刷新失败且仍在 stale 窗口内会继续使用旧偏好，避免服务短抖动导致外部渠道误触达；没有可用缓存时仍抛出错误，由 dispatcher 把外部 provider 投递转为 `queued + notification_preference_lookup_failed`。Docker Compose、K8s 和架构守卫已加入缓存参数位。
- 验收证据：
  - `npm test --workspace @infinitech/notification-worker`
  - `npm run verify:architecture`
  - `git diff --check`
  - `npm run verify`
- 当前边界：这是单 worker 进程内短 TTL 缓存和 stale-if-error 失败关闭首版。偏好变更事件驱动主动失效已由后续 `DONE-20260525-143` 承接，批量保存和审批应用已由后续 `DONE-20260525-144`/`DONE-20260525-145` 承接；真实渠道账号、模板审批、provider sandbox/生产字段联调、升级策略和跨端消息中心仍待补。
- 文件：
  - `services/notification-worker/src/index.mjs`
  - `services/notification-worker/src/index.test.mjs`
  - `infra/docker/compose.yml`
  - `infra/k8s/base/app-stack.yaml`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `PROJECT_STATUS.md`
  - `docs/product/recent-progress-roadmap.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260525-141 用户端通知偏好设置首版

- 日期：2026-05-25
- 结果：通知偏好链路从“商户和运营可配置”继续推进到“用户可在小程序端自助配置”。后端新增 `GET/PUT /api/user/notification-preferences`，用户角色只能读写自己的 `target_role=user` 偏好，管理员可按授权目标查询；保存时服务端会强制覆盖请求里的 `target_role/target_id`，避免用户越权写入商户或其他用户偏好。BFF 已放行用户端 GET/PUT 路由，用户小程序 API client 新增 `getUserNotificationPreferences` 和 `saveUserNotificationPreference`，首页新增“通知偏好”入口，`pages/notification-preferences/index` 可按 `order.status_changed`、`after_sales.updated`、`coupon.campaign` 配置微信订阅、短信、App Push 开关、静默时段和静默渠道。
- 验收证据：
  - `cd services/api-go && go test -count=1 ./internal/httpapi -run 'TestMerchantNotificationsHTTPFlow'`
  - `npm test --workspace @infinitech/bff`
  - `npm run verify:architecture`
  - `git diff --check`
  - `npm run verify`
- 当前边界：这是用户端通知偏好 UI 与 API 首版，不等于真实触达渠道生产可用。偏好缓存与失败关闭首版已由后续 `DONE-20260525-142` 承接，偏好变更主动失效已由后续 `DONE-20260525-143` 承接，批量保存和审批应用已由后续 `DONE-20260525-144`/`DONE-20260525-145` 承接；真实短信/企业微信/订阅消息/push 生产账号、模板审批、provider sandbox/生产字段联调、升级策略和跨端消息中心仍待补。
- 文件：
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `apps/user-wechat-miniprogram/utils/api.ts`
  - `apps/user-wechat-miniprogram/app.json`
  - `apps/user-wechat-miniprogram/pages/index/index.ts`
  - `apps/user-wechat-miniprogram/pages/notification-preferences/index.ts`
  - `apps/user-wechat-miniprogram/pages/notification-preferences/index.wxml`
  - `apps/user-wechat-miniprogram/pages/notification-preferences/index.wxss`
  - `apps/user-wechat-miniprogram/pages/notification-preferences/index.json`
  - `apps/user-wechat-miniprogram/README.md`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `PROJECT_STATUS.md`
  - `docs/product/recent-progress-roadmap.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260525-140 管理端通知偏好操作入口首版

- 日期：2026-05-25
- 结果：通知链路从“商户可自助配置偏好”继续推进到“运营可在 Admin Web 读取和维护目标通知偏好”。Admin Web 操作目录新增 `notification-preferences` 和 `notification-preference-save`，可调用运营侧 `GET/PUT /api/admin/notification-preferences`；通知运营页新增偏好查询/保存快捷动作，通知详情抽屉会按目标商户、来源 topic 和失败渠道预填 `target_role`、`target_id`、`notification_type`、`disabled_channels` 和 `quiet_hours` JSON。保存通知偏好会进入高风险二次确认，理由明确提示会改变外部触达渠道和静默窗口；请求构造支持 `json` 字段类型，可把表单 JSON 解析成 `quiet_hours` 对象，非法 JSON 会阻断请求。
- 验收证据：
  - `npm test --workspace @infinitech/admin-web`
  - `npm run verify:architecture`
  - `git diff --check`
  - `npm run verify`
- 当前边界：这是管理端通知偏好操作入口首版，不是完整通知策略中心。用户端通知偏好 UI 已由后续 `DONE-20260525-141` 承接，偏好缓存与失败关闭首版已由后续 `DONE-20260525-142` 承接，偏好变更主动失效已由后续 `DONE-20260525-143` 承接，批量保存和审批应用已由后续 `DONE-20260525-144`/`DONE-20260525-145` 承接；真实短信/企业微信/订阅消息/push 生产账号、模板审批、provider sandbox/生产字段联调、升级策略和跨端消息中心仍待补。
- 文件：
  - `apps/admin-web/src/adminApi.mjs`
  - `apps/admin-web/src/main.js`
  - `apps/admin-web/src/adminOperations.mjs`
  - `apps/admin-web/src/adminViews.mjs`
  - `apps/admin-web/src/adminDetails.mjs`
  - `apps/admin-web/src/adminApi.test.mjs`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `apps/admin-web/README.md`
  - `PROJECT_STATUS.md`
  - `docs/product/recent-progress-roadmap.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260525-139 商户端通知偏好设置首版

- 日期：2026-05-25
- 结果：通知链路从“后端有偏好账本、worker 会读取”继续推进到“商户能在端侧自助配置偏好”。`apps/merchant-flutter` 新增 `lib/features/notifications/merchant_notification_preferences_page.dart` 和首页“通知”入口，商户可按 `order.status_changed`、`merchant.qualification_reviewed` 配置微信订阅、短信、企业微信和端内 Push 的外部渠道开关，也可开启静默时段，设置开始/结束时间、时区和静默渠道。商户端 API client 新增 `getMerchantNotificationPreferences` 与 `saveMerchantNotificationPreference`，直接调用 `GET/PUT /api/merchant/notification-preferences`，保存后的 `disabled_channels` 和 `quiet_hours` 进入后端偏好账本，供 `notification-worker` 投递前读取。架构守卫已覆盖页面路径、首页入口、API client、通知类型、渠道字段和静默 payload。
- 验收证据：
  - `npm run verify:architecture`
  - `git diff --check`
  - `npm run verify`
- 当前边界：这是商户端通知偏好 UI 首版，不等于真实触达渠道生产可用。管理端偏好操作入口已由后续 `DONE-20260525-140` 承接，用户端偏好 UI 已由后续 `DONE-20260525-141` 承接，偏好变更主动失效已由后续 `DONE-20260525-143` 承接，批量保存和审批应用已由后续 `DONE-20260525-144`/`DONE-20260525-145` 承接；真实短信/企业微信/订阅消息/push 生产账号、渠道模板审批、provider 字段映射/sandbox 联调、升级策略和跨端消息中心仍待补。
- 文件：
  - `apps/merchant-flutter/lib/features/notifications/merchant_notification_preferences_page.dart`
  - `apps/merchant-flutter/lib/features/home/merchant_home_page.dart`
  - `apps/merchant-flutter/lib/router.dart`
  - `apps/merchant-flutter/lib/api/merchant_api.dart`
  - `apps/merchant-flutter/README.md`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `PROJECT_STATUS.md`
  - `docs/product/recent-progress-roadmap.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260525-138 通知静默到期自动扫描调度首版

- 日期：2026-05-25
- 结果：通知链路从“运营可手动把静默 queued 回执调度成延迟重投”继续推进到“静默窗口结束后可自动扫描到期回执并调度可靠重投”。quiet-window provider 抑制回执现在会记录静默结束 `retry_at`；`GET /api/admin/notification-deliveries` 支持 `retry_at_before`，后台新增 `POST /api/admin/notification-deliveries/quiet-window-retries/schedule`，按 `status=queued`、`error_code=notification_quiet_window` 和 `retry_at<=now` 扫描到期回执，再生成 `notification.delivery_retries` 延迟 outbox。`notification-worker` 新增静默重试调度器和周期循环，可通过 `NOTIFICATION_QUIET_RETRY_AUTO_SCHEDULE`、`NOTIFICATION_QUIET_RETRY_INTERVAL_MS` 和 `NOTIFICATION_QUIET_RETRY_LIMIT` 在部署侧开启；BFF allowlist、Admin Web 高风险操作目录/通知详情抽屉、Docker Compose、K8s 和架构守卫已接入。
- 验收证据：
  - `npm test --workspace @infinitech/notification-worker`
  - `npm test --workspace @infinitech/admin-web`
  - `npm test --workspace @infinitech/bff`
  - `go test -count=1 ./internal/platform -run 'TestMerchantNotificationCenterStoresAndReadsQualificationReview'`
  - `go test -count=1 ./internal/httpapi -run 'TestMerchantNotificationsHTTPFlow'`
  - `cd services/api-go && go test -count=1 ./...`
  - `npm run verify:architecture`
  - `npm run verify`
- 当前边界：这是静默到期扫描与可靠重投调度首版，不等于真实通知渠道已经生产可用。商户端通知偏好 UI 已由后续 `DONE-20260525-139` 承接，管理端偏好操作入口已由后续 `DONE-20260525-140` 承接，用户端偏好 UI 已由后续 `DONE-20260525-141` 承接，批量保存和审批应用已由后续 `DONE-20260525-144`/`DONE-20260525-145` 承接；真实短信/企业微信/订阅消息/push 生产账号、渠道模板审批、provider 字段映射/sandbox 联调、升级策略和跨端消息中心仍待补。
- 文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `services/notification-worker/src/index.mjs`
  - `services/notification-worker/src/index.test.mjs`
  - `apps/admin-web/src/adminApi.mjs`
  - `apps/admin-web/src/adminApi.test.mjs`
  - `apps/admin-web/src/adminViews.mjs`
  - `apps/admin-web/src/adminOperations.mjs`
  - `apps/admin-web/src/adminDetails.mjs`
  - `infra/docker/compose.yml`
  - `infra/k8s/base/app-stack.yaml`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `PROJECT_STATUS.md`
  - `docs/product/recent-progress-roadmap.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260525-137 通知静默 queued 再投递调度首版

- 日期：2026-05-25
- 结果：通知链路从“静默窗口只会生成 queued 回执”继续推进到“静默 queued 回执可被运营调度为延迟再投递”。`NotificationDeliveryRetryScheduleRequest` 新增 `status`、`error_code` 和 `retry_at`，默认仍兼容 failed 回执退避重试；`GET /api/admin/notification-deliveries` 支持 `error_code` 筛选。内存 Store 与 PostgreSQL-backed Store 可按 `status=queued`、`error_code=notification_quiet_window` 选出静默窗口回执，并生成 `notification.delivery_retries` outbox 事件，事件 `available_at` 对齐指定 `retry_at`，payload 继续携带原通知快照，worker 到点可按原通知重发 provider。Admin Web 通知回执查询新增错误码筛选，重试表单新增 queued/failed 状态、错误码和指定重试时间；重试计划通知正文会标出回执状态和错误码，方便运营辨认这是静默补发。
- 验收证据：
  - `npm test --workspace @infinitech/notification-worker`
  - `go test -count=1 ./internal/platform -run 'TestMerchantNotificationCenterStoresAndReadsQualificationReview'`
  - `go test -count=1 ./internal/httpapi -run 'TestMerchantNotificationsHTTPFlow'`
  - `npm test --workspace @infinitech/admin-web`
  - `npm run verify:architecture`
- 当前边界：这是静默回执到延迟重投的可运营调度首版；自动扫描已由后续 `DONE-20260525-138` 承接，商户端偏好 UI 已由后续 `DONE-20260525-139` 承接，管理端偏好操作入口已由后续 `DONE-20260525-140` 承接，用户端偏好 UI 已由后续 `DONE-20260525-141` 承接，批量保存和审批应用已由后续 `DONE-20260525-144`/`DONE-20260525-145` 承接；仍缺真实短信/企业微信/订阅消息/push 生产账号、渠道模板审批和 provider 字段联调。
- 文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/notification-worker/src/index.mjs`
  - `services/notification-worker/src/index.test.mjs`
  - `apps/admin-web/src/adminApi.mjs`
  - `apps/admin-web/src/adminApi.test.mjs`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `PROJECT_STATUS.md`
  - `docs/product/recent-progress-roadmap.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260525-136 通知 worker 后端偏好读取首版

- 日期：2026-05-25
- 结果：通知链路从“偏好可通过 API 持久化”继续推进到“worker provider 投递前实际读取后端偏好账本并执行”。后端通知偏好列表新增 `preference_key` 精确查询，便于按 `default`、目标角色、目标 ID、通知类型和目标+类型 key 读取规则。`notification-worker` 新增 `deliveryPreferencesFromRecords` 和 `createNotificationPreferenceResolver`，会调用 `/api/admin/notification-preferences?preference_key=...` 拉取后端偏好记录，并与静态 `NOTIFICATION_DELIVERY_PREFERENCES` 合并后进入 provider 投递决策。若偏好读取失败，外部渠道不调用 provider，而是生成 `queued` 回执并记录 `notification_preference_lookup_failed`，避免在偏好不确定时误触达。
- 验收证据：
  - `npm test --workspace @infinitech/notification-worker`
  - `go test -count=1 ./internal/httpapi -run 'TestMerchantNotificationsHTTPFlow'`
  - `go test -count=1 ./internal/platform -run 'TestMerchantNotificationCenterStoresAndReadsQualificationReview'`
  - `npm run verify:architecture`
  - `git diff --check`
  - `cd services/api-go && go test -count=1 ./...`
  - `npm run verify`
- 当前边界：这是 worker 动态读取偏好首版；quiet-window 到期自动扫描已由后续 `DONE-20260525-138` 承接，商户端偏好 UI 已由后续 `DONE-20260525-139` 承接，管理端偏好操作入口已由后续 `DONE-20260525-140` 承接，用户端偏好 UI 已由后续 `DONE-20260525-141` 承接，批量保存和审批应用已由后续 `DONE-20260525-144`/`DONE-20260525-145` 承接；仍缺真实短信/企业微信/订阅消息/push 生产账号、渠道模板审批和 provider 字段联调。
- 文件：
  - `services/notification-worker/src/index.mjs`
  - `services/notification-worker/src/index.test.mjs`
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `PROJECT_STATUS.md`
  - `docs/product/recent-progress-roadmap.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260525-135 通知偏好后端账本与 API 首版

- 日期：2026-05-25
- 结果：通知链路从“worker 可按环境配置执行偏好”继续推进到“偏好可持久化、可审计、可由商户/运营 API 维护”。新增 `NotificationPreference`、`NotificationQuietHours`、`SaveNotificationPreferenceRequest` 和 `NotificationPreferenceListRequest` 合约；内存 Store 与 PostgreSQL-backed Store 支持 `NotificationPreferences`、`SaveNotificationPreference` 和 `SaveNotificationPreferenceWithAudit`，PostgreSQL 路径新增 `platform_notification_preferences` 规范表、`preference_key` 唯一约束和目标/类型索引。HTTP 新增商户 `GET/PUT /api/merchant/notification-preferences` 与运营 `GET/PUT /api/admin/notification-preferences`，商户只能维护自身通知类型偏好，运营写入由 `notification:write` 守护并记录 `admin.notification_preferences.saved` 审计；BFF 已代理对应路径。
- 验收证据：
  - `go test -count=1 ./internal/platform -run 'TestMerchantNotificationCenterStoresAndReadsQualificationReview'`
  - `go test -count=1 ./internal/httpapi -run 'TestMerchantNotificationsHTTPFlow'`
  - `npm test --workspace @infinitech/bff`
  - `npm run verify:architecture`
  - `git diff --check`
  - `cd services/api-go && go test -count=1 ./...`
  - `npm run verify`
- 当前边界：这是通知偏好后端账本与 API 首版；notification-worker 动态读取、静默窗口到期自动再投递、商户端 UI、管理端偏好操作入口、用户端偏好 UI、偏好变更主动失效、批量保存和审批应用已由后续增量承接，仍缺真实短信/企业微信/订阅消息/push 生产账号和渠道联调。
- 文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/auth.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `PROJECT_STATUS.md`
  - `docs/product/recent-progress-roadmap.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260525-134 通知偏好与静默窗口首版

- 日期：2026-05-25
- 结果：通知链路从“provider dispatch 可携带真实渠道模板”继续推进到“外部触达可按偏好和静默窗口控制”。`notification-worker` 新增 `normalizeDeliveryPreferences` 和 `notificationDeliveryPreferenceDecision`，可通过 `NOTIFICATION_DELIVERY_PREFERENCES` 解析偏好规则；规则按 `default`、`target_role`、`target_role:target_id`、`type:{notification_type}` 和 `target_role:target_id:{notification_type}` 合并，支持 `enabled_channels`、`disabled_channels` 与 `quiet_hours`。被偏好禁用或落入静默窗口的外部 provider 投递不会调用 provider endpoint/adapter，而是生成 `queued` 回执，分别写入 `notification_preference_disabled` 或 `notification_quiet_window` 原因，避免误触达又保留运营证据。Docker Compose 和 K8s 部署骨架已预留 `NOTIFICATION_DELIVERY_PREFERENCES`。
- 验收证据：
  - `npm test --workspace @infinitech/notification-worker`
  - `npm run verify:architecture`
  - `git diff --check`
  - `cd services/api-go && go test -count=1 ./...`
  - `npm run verify`
- 当前边界：这是 worker 侧偏好过滤和静默窗口首版；偏好后端账本、quiet-window 到期自动扫描、端侧偏好设置 UI、批量保存和审批应用已由后续增量承接，仍缺策略灰度、升级策略、跨端消息中心和真实渠道生产联调。
- 文件：
  - `services/notification-worker/src/index.mjs`
  - `services/notification-worker/src/index.test.mjs`
  - `infra/docker/compose.yml`
  - `infra/k8s/base/app-stack.yaml`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `PROJECT_STATUS.md`
  - `docs/product/recent-progress-roadmap.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260525-133 通知 provider 模板映射与渠道 payload 规范首版

- 日期：2026-05-25
- 结果：通知链路从“provider 可发送、回调可验签入账”继续推进到“provider dispatch 可携带真实渠道模板和结构化 payload”。`notification-worker` 新增 `normalizeProviderTemplates`、`applyProviderTemplate` 和渠道 payload 构建逻辑，可通过 `NOTIFICATION_PROVIDER_TEMPLATES` 按 notification type/template_key、channel 和 provider 匹配模板配置。provider dispatch 现在会保留原有兼容字段，同时新增 `template_id`、`template_params` 和 `provider_payload`：微信订阅消息覆盖 `touser/template_id/page/data/lang`，短信覆盖 `phone/template_code/sign_name/params`，企业微信覆盖 `touser/agentid/msgtype/template_id/text/params`，push 覆盖 `audience/title/body/template_id/extras`。Docker Compose 和 K8s 部署骨架已预留 `NOTIFICATION_PROVIDER_TEMPLATES`，后续可注入审核后的渠道模板映射进入 sandbox/生产联调。
- 验收证据：
  - `npm test --workspace @infinitech/notification-worker`
  - `npm run verify:architecture`
  - `git diff --check`
  - `cd services/api-go && go test -count=1 ./...`
  - `npm run verify`
- 当前边界：这是 provider 模板配置与渠道 payload 结构首版；偏好过滤、静默窗口和端侧偏好设置 UI 已由后续增量承接，仍缺真实短信、企业微信、微信订阅消息、push 生产账号、渠道模板审批、各 provider sandbox 字段联调、供应商配置审批、升级策略和跨端消息中心。
- 文件：
  - `services/notification-worker/src/index.mjs`
  - `services/notification-worker/src/index.test.mjs`
  - `infra/docker/compose.yml`
  - `infra/k8s/base/app-stack.yaml`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `PROJECT_STATUS.md`
  - `docs/product/recent-progress-roadmap.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260525-132 通知 provider 回调验签入账首版

- 日期：2026-05-25
- 结果：通知链路从“provider 可发送并写即时回执”继续推进到“外部 provider 异步回调可验签、可幂等入账”。API 新增公开回调入口 `POST /api/notifications/provider-callback`，接收 notification/channel/provider/status/provider message id、错误码、错误信息和 callback 时间；生产配置 `NOTIFICATION_PROVIDER_CALLBACK_SECRET` 后按 HMAC-SHA256 canonical lines 校验签名，缺失或错误签名返回 `INVALID_NOTIFICATION_PROVIDER_SIGNATURE`。回调入账会归一化 delivered/failed/queued 与时间字段，按 callback idempotency key、provider message id 或 notification/callback 时间生成幂等键，写入 `platform_notification_deliveries`，重复回调返回同一条投递回执。BFF 已代理该路径；`notification-worker` 新增 `signProviderCallback` 与 `normalizeProviderCallbackPayload`，便于 provider sandbox/mock 生成与后端一致的签名 payload；Docker Compose 与 K8s 部署骨架已预留 `NOTIFICATION_PROVIDER_CALLBACK_SECRET`/secretKeyRef。
- 验收证据：
  - `cd services/api-go && go test -count=1 ./internal/httpapi -run 'TestMerchantNotificationsHTTPFlow|TestNotificationProviderCallbackHTTPFlow'`
  - `npm test --workspace @infinitech/notification-worker`
  - `npm test --workspace @infinitech/bff`
  - `npm run verify:architecture`
  - `git diff --check`
  - `cd services/api-go && go test -count=1 ./...`
  - `npm run verify`
- 当前边界：这是通用 provider 回调安全边界和投递台账入账首版；模板 payload 规范和通知偏好 UI 已由后续增量承接，仍缺真实短信、企业微信、微信订阅消息、push 的生产账号、渠道模板审批、各 provider 字段映射/sandbox 联调、供应商配置审批、升级策略和跨端消息中心。
- 文件：
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/api-go/cmd/api/main.go`
  - `services/notification-worker/src/index.mjs`
  - `services/notification-worker/src/index.test.mjs`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `infra/docker/compose.yml`
  - `infra/k8s/base/app-stack.yaml`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `PROJECT_STATUS.md`
  - `docs/product/recent-progress-roadmap.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260525-131 通知 provider 执行器首版

- 日期：2026-05-25
- 结果：通知链路从“可编排重试”继续推进到“可生成真实渠道发送请求、可写回 provider 回执”。`notification.delivery_retries` outbox payload 新增原始 `notifications` 快照，重试任务不再只有失败回执 ID，而是携带原通知标题、正文、目标和类型，便于真实渠道重发。`notification-worker` 新增 provider dispatcher，可按 `NOTIFICATION_PROVIDER_CHANNELS` 处理初始通知，也可从 `notification.delivery_retries` 事件按失败回执生成短信、企业微信、微信订阅消息、push 的 provider dispatch；执行器会调用配置 endpoint/adapter，并把 delivered/failed、provider message id、错误码和错误信息写回 `/api/notifications/{notificationID}/deliveries`。未配置 provider endpoint 时会写回 `provider_not_configured` failed 回执，避免外部渠道还没触达却被误标成功。Docker Compose 和 K8s 部署骨架新增 `notification-worker` 及 provider endpoint/token 环境变量位。
- 验收证据：
  - `npm test --workspace @infinitech/notification-worker`
  - `cd services/api-go && go test -count=1 ./internal/platform -run 'TestMerchantNotificationCenterStoresAndReadsQualificationReview'`
  - `cd services/api-go && go test -count=1 ./internal/httpapi -run 'TestMerchantNotificationsHTTPFlow'`
  - `npm run verify:architecture`
  - `git diff --check`
  - `cd services/api-go && go test -count=1 ./...`
  - `npm run verify`
- 当前边界：这是 provider dispatch、endpoint/adapter 调用、失败回执写回和部署配置骨架首版；provider 回调验签入账和通知偏好 UI 已由后续增量承接，仍缺真实短信、企业微信、微信订阅消息、push 的生产账号、模板审批/映射、供应商配置审批、升级策略和跨端消息中心。
- 文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/notification-worker/src/index.mjs`
  - `services/notification-worker/src/index.test.mjs`
  - `infra/docker/compose.yml`
  - `infra/k8s/base/app-stack.yaml`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `PROJECT_STATUS.md`
  - `apps/admin-web/README.md`
  - `docs/product/recent-progress-roadmap.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260525-130 通知失败重试编排首版

- 日期：2026-05-25
- 结果：通知链路从“失败可告警”继续推进到“失败可按退避窗口可靠编排重试、可审计”。新增 `NotificationDeliveryRetrySchedule` 和 `POST /api/admin/notification-deliveries/retries/schedule`，由 `notification:write` scope 守护，运营可按 `target_role`、`target_id`、`channel`、`provider`、`limit`、`retry_after_seconds` 和 `now` 汇总 failed 投递回执并安排重试。内存 Store 与 PostgreSQL-backed Store 均会生成 `notification.delivery_retries` outbox 事件，事件类型为 `notification.delivery_retries.scheduled`，且 `available_at` 对齐 `retry_at`，让该事件只在 provider 退避窗口后进入 ready 队列；没有匹配失败回执时返回 skipped 并保留审计证据。调度动作写入 `admin.notification_delivery_retries.scheduled` 审计，payload 记录计划数量、渠道、provider、退避秒数、`retry_at`、重试策略、outbox event id 和幂等键。BFF 已代理该后台路径；Admin Web `notifications` 模块、通知详情抽屉和操作目录新增“安排投递重试”，可从失败行预填 `merchant_19`、`wechat_subscribe`、provider、limit 和默认退避秒数，并作为高风险动作进入二次确认。`notification-worker` 已订阅 `notification.delivery_retries`，生成面向安全/通知运营目标的重试计划通知 payload；outbox relay 默认 topic、Docker Compose 和 K8s 部署骨架已包含该 topic。
- 验收证据：
  - `cd services/api-go && go test -count=1 ./internal/platform -run 'TestMerchantNotificationCenterStoresAndReadsQualificationReview'`
  - `cd services/api-go && go test -count=1 ./internal/httpapi -run 'TestMerchantNotificationsHTTPFlow'`
  - `npm test --workspace @infinitech/admin-web`
  - `npm test --workspace @infinitech/notification-worker`
  - `npm test --workspace @infinitech/bff`
  - `npm test --workspace @infinitech/outbox-relay-worker`
  - `npm run verify:architecture`
  - `git diff --check`
  - `cd services/api-go && go test -count=1 ./...`
  - `npm run verify`
  - 本地 Playwright/Chromium 打开 `http://127.0.0.1:4177/apps/admin-web/index.html`，确认通知运营页展示“安排投递重试”，失败通知 `ntf_2` 详情抽屉包含该动作，点击后预填 `target_id=merchant_19`、`channel=wechat_subscribe`、`provider=wechat_subscribe`、`limit=20`、`retry_after_seconds=300`，并进入 `pending_confirmation` 高风险确认；浏览器 console error 和 page error 为空。静态服务器已关闭并确认端口不可连接。
- 当前边界：这是可靠重试编排、延迟 outbox 和审计证据首版；provider 执行器骨架和通知偏好 UI 已由后续增量承接，仍缺真实短信、企业微信、微信小程序订阅消息、商户端 push 生产账号、模板映射、回调验签、供应商配置审批、跨端消息中心、升级策略和静默窗口。
- 文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `apps/admin-web/src/adminApi.mjs`
  - `apps/admin-web/src/adminApi.test.mjs`
  - `apps/admin-web/src/adminAudit.mjs`
  - `apps/admin-web/src/adminDetails.mjs`
  - `apps/admin-web/src/adminOperations.mjs`
  - `apps/admin-web/src/adminViews.mjs`
  - `services/notification-worker/src/index.mjs`
  - `services/notification-worker/src/index.test.mjs`
  - `services/outbox-relay-worker/src/index.mjs`
  - `services/outbox-relay-worker/src/index.test.mjs`
  - `infra/docker/compose.yml`
  - `infra/k8s/base/app-stack.yaml`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `PROJECT_STATUS.md`
  - `apps/admin-web/README.md`
  - `docs/product/recent-progress-roadmap.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260525-129 通知失败回执告警首版

- 日期：2026-05-25
- 结果：通知链路从“失败可查询”继续推进到“失败可可靠告警、可审计”。新增 `NotificationFailureAlertEmission` 和 `POST /api/admin/notification-deliveries/failure-alerts/emit`，由 `notification:write` scope 守护，运营可按 `target_role`、`target_id`、`channel`、`provider`、`limit` 和 `now` 汇总 failed 投递回执。内存 Store 与 PostgreSQL-backed Store 均会把匹配到的失败回执生成 `notification.delivery_failed_alerts` outbox 事件，事件类型为 `notification.delivery_failed_alerts.emitted`，并写入 `admin.notification_delivery_failure_alerts.emitted` 审计；无失败回执时返回 skipped，避免空告警污染队列。BFF 已代理该后台路径；Admin Web `notifications` 模块、通知详情抽屉和操作目录新增“投递失败告警”，可从失败行预填 `merchant_19`、`wechat_subscribe` 和 provider，并作为高风险动作进入二次确认。`notification-worker` 已订阅 `notification.delivery_failed_alerts`，生成面向安全目标的失败告警通知 payload；outbox relay 默认 topic、Docker Compose 和 K8s 部署骨架已包含该 topic。
- 验收证据：
  - `npm test --workspace @infinitech/admin-web`
  - `npm test --workspace @infinitech/notification-worker`
  - `npm test --workspace @infinitech/outbox-relay-worker`
  - `npm test --workspace @infinitech/bff`
  - `cd services/api-go && go test -count=1 ./internal/httpapi -run 'TestMerchantNotificationsHTTPFlow'`
  - `cd services/api-go && go test -count=1 ./internal/platform -run 'TestMerchantNotificationCenterStoresAndReadsQualificationReview'`
  - `npm run verify:architecture`
  - `git diff --check`
  - `cd services/api-go && go test -count=1 ./...`
  - `npm run verify`
  - 本地浏览器打开 `http://127.0.0.1:4176/apps/admin-web/index.html`，确认通知运营页展示“投递失败告警”，失败通知 `ntf_2` 详情抽屉包含该动作，点击后预填 `target_id=merchant_19`、`channel=wechat_subscribe`、`provider=wechat_subscribe`、`limit=20`，并进入 `pending_confirmation` 高风险确认；浏览器 console error 为空。静态服务器已关闭并确认端口不可连接。
- 当前边界：这是失败回执到可靠 outbox 告警和审计证据的首版；失败重试编排已由后续 `DONE-20260525-130` 承接，仍缺真实短信、企业微信、微信小程序订阅消息、商户端 push provider 执行器，升级策略、静默窗口、通知偏好和跨端消息中心 UI。
- 文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `apps/admin-web/src/adminApi.mjs`
  - `apps/admin-web/src/adminApi.test.mjs`
  - `apps/admin-web/src/adminAudit.mjs`
  - `apps/admin-web/src/adminDetails.mjs`
  - `apps/admin-web/src/adminOperations.mjs`
  - `apps/admin-web/src/adminViews.mjs`
  - `services/notification-worker/src/index.mjs`
  - `services/notification-worker/src/index.test.mjs`
  - `services/outbox-relay-worker/src/index.mjs`
  - `services/outbox-relay-worker/src/index.test.mjs`
  - `infra/docker/compose.yml`
  - `infra/k8s/base/app-stack.yaml`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `PROJECT_STATUS.md`
  - `apps/admin-web/README.md`
  - `docs/product/recent-progress-roadmap.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260525-128 Admin Web 通知运营页首版

- 日期：2026-05-25
- 结果：通知链路从“后端可查、回执可记”继续推进到“运营后台可用”。Admin Web `notifications` 模块已从 `planned` 调整为 `wired`，新增“通知运营”视图，展示未读站内信、失败回执、渠道和待接 provider 状态；操作目录新增 `notifications`、`notification-deliveries` 和 `notification-delivery-record`，分别调用 `GET /api/admin/notifications`、`GET /api/admin/notification-deliveries` 和 `POST /api/notifications/{notificationID}/deliveries`。通知详情抽屉可按通知 ID、商户目标、来源 topic、失败状态和渠道预填通知台账、回执查询、补录回执、通知审计和来源 outbox 查询动作；补录回执纳入高风险二次确认。Admin Web RBAC 同步补齐 `notification:read` 和 `notification:write`，`ops_admin` 可读写通知账本/回执，`support_admin` 可只读排查通知争议。
- 验收证据：
  - `npm test --workspace @infinitech/admin-web`
  - `npm run verify:architecture`
  - `git diff --check`
  - `npm run verify`
- 当前边界：这是运营可视化和人工排查入口首版；失败回执告警和失败重试编排已由 `DONE-20260525-129` 与 `DONE-20260525-130` 承接，仍缺真实短信、企业微信、微信小程序订阅消息、商户端 push provider 执行器，通知偏好和跨端消息中心 UI。
- 文件：
  - `apps/admin-web/src/adminApi.mjs`
  - `apps/admin-web/src/adminApi.test.mjs`
  - `apps/admin-web/src/adminAudit.mjs`
  - `apps/admin-web/src/adminDetails.mjs`
  - `apps/admin-web/src/adminOperations.mjs`
  - `apps/admin-web/src/adminViews.mjs`
  - `apps/admin-web/src/config.mjs`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `PROJECT_STATUS.md`
  - `apps/admin-web/README.md`
  - `docs/product/recent-progress-roadmap.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260525-127 通知投递回执台账首版

- 日期：2026-05-25
- 结果：通知链路从“可写入、可查询”继续推进到“可证明投递、可排查失败”。新增 `PlatformNotificationDelivery` 回执账本，支持 `queued`、`delivered`、`failed` 状态，记录通知 ID、目标角色/对象、channel、provider、provider message id、错误码、错误信息、幂等键、attempted_at 和 delivered_at。内存 Store 与 PostgreSQL-backed Store 均新增 `RecordNotificationDelivery` 和 `NotificationDeliveries`；PostgreSQL 路径新增 `platform_notification_deliveries` 规范表、幂等唯一约束、通知/目标/渠道状态索引和快照恢复。HTTP 新增 `POST /api/notifications/{notificationID}/deliveries` 回执写入入口和 `GET /api/admin/notification-deliveries` 后台查询入口，BFF 已代理；notification-worker 写入 in-app 通知成功后会继续写入 delivered 回执，失败回执也可由后续外部渠道 worker 幂等记录。
- 验收证据：
  - `cd services/api-go && go test -count=1 ./internal/platform -run 'TestMerchantNotificationCenterStoresAndReadsQualificationReview'`
  - `cd services/api-go && go test -count=1 ./internal/httpapi -run 'TestMerchantNotificationsHTTPFlow|TestBackofficeRBACScopeMatrix'`
  - `npm test --workspace @infinitech/notification-worker`
  - `npm test --workspace @infinitech/bff`
  - `npm run verify:architecture`
- 当前边界：这是回执账本和 API 首版；Admin Web 通知运营页、失败回执告警和失败重试编排已由 `DONE-20260525-128`、`DONE-20260525-129` 与 `DONE-20260525-130` 承接，仍缺真实短信、企业微信、微信小程序订阅消息、商户端 push provider 执行器，通知偏好和跨端消息中心 UI。
- 文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `services/notification-worker/src/index.mjs`
  - `services/notification-worker/src/index.test.mjs`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `PROJECT_STATUS.md`
  - `docs/product/recent-progress-roadmap.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260525-126 通知运营查询接口首版

- 日期：2026-05-25
- 结果：商户站内通知从“商户可见”继续推进到“运营可查”。新增 `notification:read` scope 和 `CanReadNotifications` 权限判断，`ops_admin` 可读写通知账本，`support_admin` 可只读排查，`security_auditor` 不能读取运营通知；新增 `GET /api/admin/notifications`，支持按 target role/id、状态、source topic/event 和 limit 查询通知账本。内存 Store 与 PostgreSQL 查询路径均支持来源过滤，并保持 target role/id 成对校验；BFF 已代理后台通知查询路径，架构守卫固定 RBAC、路由和 source 过滤。
- 验收证据：
  - `cd services/api-go && go test -count=1 ./internal/httpapi -run 'TestMerchantNotificationsHTTPFlow|TestBackofficeRBACScopeMatrix'`
  - `cd services/api-go && go test -count=1 ./internal/platform -run 'TestMerchantNotificationCenterStoresAndReadsQualificationReview'`
  - `npm test --workspace @infinitech/bff`
  - `npm run verify:architecture`
- 当前边界：这是后台通知账本查询 API 首版；投递回执、Admin Web 通知运营页和失败回执告警已由 `DONE-20260525-127`、`DONE-20260525-128` 和 `DONE-20260525-129` 承接，仍缺通知偏好、真实短信/企业微信/订阅消息/push 渠道和完整跨端消息中心。
- 文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/httpapi/auth.go`
  - `services/api-go/internal/httpapi/auth_test.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `PROJECT_STATUS.md`
  - `docs/product/recent-progress-roadmap.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260525-125 商户站内通知中心首版

- 日期：2026-05-25
- 结果：商户资质审核结果从“可靠事件 + worker payload”继续推进到可见的站内通知账本。新增 `PlatformNotification`、创建/列表/已读请求、`in_app` 渠道和 `unread/read` 状态；内存 Store 与 PostgreSQL-backed Store 均支持通知幂等写入、按目标与状态查询和标记已读。PostgreSQL 路径已新增 `platform_notifications` 规范表、幂等键唯一约束、目标/状态/时间索引和 source 索引，并能从规范表恢复快照。HTTP 新增 `POST /api/notifications` worker/后台写入入口、`GET /api/merchant/notifications` 商户通知列表和 `POST /api/merchant/notifications/{notificationID}/read` 已读入口，BFF 已代理商户通知路径；`notification-worker` 可把 `merchant.qualification_reviewed` 等可靠事件规范化为站内通知创建请求并携带 worker token 写入 API，架构守卫固定链路。
- 验收证据：
  - `npm test --workspace @infinitech/notification-worker`
  - `npm test --workspace @infinitech/bff`
  - `cd services/api-go && go test -count=1 ./internal/platform -run 'TestMerchantNotificationCenterStoresAndReadsQualificationReview|TestAdminMerchantQualificationQueueAndDetail'`
  - `cd services/api-go && go test -count=1 ./internal/httpapi -run 'TestMerchantNotificationsHTTPFlow|TestBackofficeRBACScopeMatrix'`
  - `cd services/api-go && go test -count=1 ./internal/httpapi -run 'TestAdminMerchantQualification|TestMerchantInviteRegisterAndQualificationHTTPFlow'`
  - `npm run verify:architecture`
  - `git diff --check`
  - `cd services/api-go && go test -count=1 ./...`
  - `npm run verify`
- 当前边界：这是站内通知账本、商户未读列表和已读标记首版，仍不是完整触达系统；投递回执、Admin Web 通知运营页和失败回执告警已由后续增量承接，还需要真实短信、企业微信、微信小程序订阅消息、商户端 push、通知偏好和跨端消息中心 UI。
- 文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/auth.go`
  - `services/api-go/internal/httpapi/auth_test.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `services/notification-worker/src/index.mjs`
  - `services/notification-worker/src/index.test.mjs`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `PROJECT_STATUS.md`
  - `docs/product/recent-progress-roadmap.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260525-124 商户资质审核结果可靠通知首版

- 日期：2026-05-25
- 结果：商户资质审核从“后台更新状态并写审计”继续推进到可靠通知链路。`ReviewMerchantQualificationWithAudit` 在审核通过/驳回后生成 `merchant.qualification_reviewed` outbox 事件，payload 包含商户、资质、审核结论、原因、有效期、目标角色和通知文案；HTTP 审核响应返回 `outbox_event`，运营可继续从 Outbox 查询、恢复或人工处置该事件。`notification-worker` 已订阅该 topic 并生成面向商户的审核结果通知 payload，outbox relay 默认 topic、Docker Compose 和 K8s 部署骨架均已加入该事件，架构守卫固定链路。
- 验收证据：
  - `go test -count=1 ./internal/platform -run 'TestMerchantInviteRegistrationAndQualificationGate|TestAdminMerchantQualificationQueueAndDetail'`
  - `go test -count=1 ./internal/httpapi -run 'TestAdminMerchantQualification'`
  - `npm test --workspace @infinitech/notification-worker`
  - `npm test --workspace @infinitech/outbox-relay-worker`
  - `npm test --workspace @infinitech/bff`
  - `npm run verify:architecture`
  - `git diff --check`
  - `cd services/api-go && go test -count=1 ./...`
  - `npm run verify`
- 当前边界：这是可靠事件、relay 默认投递和通知 worker payload 首版，仍不是完整真实触达；站内信落库、投递回执、运营可见投递状态和失败回执告警已由后续增量承接，还需要接短信、企业微信、微信小程序订阅消息和商户端 push。
- 文件：
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/runtime.test.mjs`
  - `services/notification-worker/src/index.mjs`
  - `services/notification-worker/src/index.test.mjs`
  - `services/outbox-relay-worker/src/index.mjs`
  - `services/outbox-relay-worker/src/index.test.mjs`
  - `infra/docker/compose.yml`
  - `infra/k8s/base/app-stack.yaml`
  - `scripts/check-architecture.mjs`
  - `README.md`
  - `PROJECT_STATUS.md`
  - `docs/product/recent-progress-roadmap.md`
  - `docs/product/commercial-readiness-checklist.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260527-147 前端交付状态盘点与用户端参考图留痕

- 日期：2026-05-27
- 结果：已完成当前仓库全量代码和前端交付状态盘点。确认用户端效果图 `01-home.png` 到 `35-medicine-order-detail.png` 均已生成并放入 `docs/product/user-miniprogram-reference/`，用户端原生微信小程序当前只注册 12 个页面，仍缺 13-35 的原生代码页面；已有 12 页也需要按最新参考图重新对齐。确认实际仓库没有 `apps/user-flutter`、`apps/merchant-flutter`、`apps/rider-flutter`、`apps/admin-flutter`，当前 `merchant-uni`、`rider-uni`、`admin-uni` 是历史目录，后续应迁移到 Flutter/Dart。新增 `docs/product/frontend-delivery-status.md` 作为后续每一步前端推进的基线留痕，并补齐 `docs/product/user-miniprogram-page-prompts.md` 的第 35 张“药品订单详情”提示词。
- 验收证据：
  - `find docs/product/user-miniprogram-reference -maxdepth 1 -type f`
  - `find apps/user-wechat-miniprogram/pages -mindepth 2 -maxdepth 3 -name 'index.wxml'`
  - `npm run verify`
- 当前边界：本次只做盘点和 Markdown 留痕，尚未开始生成 13-35 原生微信小程序页面，也尚未创建 Flutter 工程。
- 文件：
  - `docs/product/frontend-delivery-status.md`
  - `docs/product/user-miniprogram-page-prompts.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260527-148 用户端登录注册效果图

- 日期：2026-05-27
- 结果：按生图流程新增用户端微信小程序登录/注册页参考图。页面使用悦享e食风格、白色原生导航栏、主色 `#009bf5`、浅灰背景、登录/注册分段、手机号验证码、微信一键登录、协议勾选和安全提示卡，作为后续原生微信小程序 `pages/auth/login/index` 还原依据。
- 验收证据：
  - 使用内置 `image_gen` 生成 UI mockup。
  - 已查看生成图并确认中文、布局、品牌和安全提示可用。
  - `file docs/product/user-miniprogram-reference/00-auth-login-register.png`
- 当前边界：本次只生成登录/注册 UI 效果图和 Markdown 留痕，尚未生成原生微信小程序登录/注册代码。
- 文件：
  - `docs/product/user-miniprogram-reference/00-auth-login-register.png`
  - `docs/product/user-miniprogram-reference/README.md`
  - `docs/product/user-miniprogram-page-prompts.md`
  - `docs/product/frontend-delivery-status.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260527-149 用户端 1.0 欢迎/登录/注册效果图复刻

- 日期：2026-05-27
- 来源：`HCRXchenghong/infinitech` 1.0 仓库的 `user-vue/pages.json`、`pages/welcome/welcome/index.vue`、`pages/auth/login/index.vue`、`pages/auth/register/index.vue`、`packages/mobile-core/src/WelcomeLandingPage.vue`、`AuthLoginPage.vue`、`AuthRegisterPage.vue`、`auth-login-page.scss` 和 `auth-register-page.scss`。
- 结果：按 1.0 用户端小程序真实欢迎页、登录页、注册页重新生成三张参考效果图。欢迎页复刻蓝色渐变启动页、蓝底白色链路/S logo、主标题「欢迎来到悦享e食」、副标题「附近美食，一键下单，准时送达」、底部「登录 / 注册」和「游客访问」卡片；登录页复刻白底极简认证页、验证码/密码登录 tabs、下划线输入、验证码按钮、微信登录入口；注册页复刻昵称、手机号、邀请码、验证码、密码、确认密码和微信注册/登录入口。
- 验收证据：
  - 已 clone 并检查 1.0 仓库源码：`/tmp/infinitech-legacy/user-vue`
  - 使用内置 `image_gen` 生成三张 UI mockup。
  - 已查看生成图并确认布局、核心文案和 1.0 视觉方向可用。
  - `file docs/product/user-miniprogram-reference/00-welcome-legacy.png docs/product/user-miniprogram-reference/00-login-legacy.png docs/product/user-miniprogram-reference/00-register-legacy.png`
- 当前边界：本次只生成 1.0 风格参考图和 Markdown 留痕，尚未生成原生微信小程序代码。
- 文件：
  - `docs/product/user-miniprogram-reference/00-welcome-legacy.png`
  - `docs/product/user-miniprogram-reference/00-login-legacy.png`
  - `docs/product/user-miniprogram-reference/00-register-legacy.png`
  - `docs/product/user-miniprogram-reference/README.md`
  - `docs/product/user-miniprogram-page-prompts.md`
  - `docs/product/frontend-delivery-status.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260527-150 用户端 1.0 欢迎页动态效果移植

- 日期：2026-05-27
- 来源：`HCRXchenghong/infinitech` 1.0 仓库的 `packages/mobile-core/src/WelcomeLandingPage.vue`。
- 结果：已把 1.0 用户端欢迎页启动动画移植为原生微信小程序页面，并设为 `apps/user-wechat-miniprogram/app.json` 第一启动页。动画保留蓝色渐变背景、logo 脉冲、大 logo 初始放大、`悦享e食` 150ms 打字机、2.5s 后完整标题「欢迎来到悦享e食」缩放归位、3.7s 底部登录/游客卡片渐入；已登录用户会跳过欢迎页进入首页，游客访问写入 `authMode=guest` 后进入首页。
- 验收证据：
  - `node -e "JSON.parse(require('fs').readFileSync('apps/user-wechat-miniprogram/app.json','utf8')); console.log('app.json ok')"`
  - `git diff --check`
  - `npm run verify`
- 当前边界：本次只移植欢迎页动态效果；登录页和注册页仍停留在参考图阶段，欢迎页「登录 / 注册」按钮暂时保留 1.0 路径并在页面未接入时提示“登录页下一步接入”。
- 文件：
  - `apps/user-wechat-miniprogram/app.json`
  - `apps/user-wechat-miniprogram/pages/welcome/welcome/index.json`
  - `apps/user-wechat-miniprogram/pages/welcome/welcome/index.wxml`
  - `apps/user-wechat-miniprogram/pages/welcome/welcome/index.wxss`
  - `apps/user-wechat-miniprogram/pages/welcome/welcome/index.ts`
  - `apps/user-wechat-miniprogram/README.md`
  - `docs/product/frontend-delivery-status.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260527-151 用户端小程序 01-35 页面补齐与后端对接

- 日期：2026-05-27
- 结果：用户端原生微信小程序已补齐欢迎/登录/注册和 01-35 业务参考图对应页面，`apps/user-wechat-miniprogram/app.json` 当前注册 38 个页面，每个页面都有 `index.json`、`index.wxml`、`index.ts`。已扩展 `utils/api.ts` 并优先接入当前 BFF/API：微信登录、首页配置、店铺、商品、团购、地址、购物车、下单、订单、售后、售后事件/证据、通知偏好、钱包充值、支付密码、余额支付、微信预支付和团购券。消息/群聊/红包/评价/钱包明细/优惠券/积分/邀请/找饭搭/反馈/处方等后端仍待补，页面已明确保留待接状态。
- 验收证据：
  - `node - <<'NODE' ... NODE` 路由完整性检查：`pages=38`，所有页面 `json/wxml/ts` 存在。
  - `git diff --check`
  - `npm run verify`
- 当前边界：本次完成用户端 UI 页面覆盖和现有后端优先接入，不代表所有缺口后端都已生产化；缺口清单见 `docs/product/user-miniprogram-backend-integration.md`。
- 文件：
  - `apps/user-wechat-miniprogram/app.json`
  - `apps/user-wechat-miniprogram/app.wxss`
  - `apps/user-wechat-miniprogram/utils/api.ts`
  - `apps/user-wechat-miniprogram/pages/**/index.json`
  - `apps/user-wechat-miniprogram/pages/**/index.wxml`
  - `apps/user-wechat-miniprogram/pages/**/index.ts`
  - `apps/user-wechat-miniprogram/README.md`
  - `docs/product/user-miniprogram-backend-integration.md`
  - `docs/product/frontend-delivery-status.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260527-152 用户端 01-07 精修与首批待接后端闭环

- 日期：2026-05-27
- 结果：已按参考图精修用户端小程序 `01` 到 `07` 高频交易页：首页补 8 宫格、精选团购、猜你喜欢、饭搭入口和底部 tab；附近商家改为简洁列表；店铺详情补头图、活动、外卖/团购 tab、分类点餐和购物车栏；购物车改为底部抽屉；确认订单补 ETA、费用层级和提交后支付选择；订单列表补状态 tab 和底部 tabbar；订单详情补地图式配送卡、时间线、地址和联系操作。已补首批待接后端：评价、钱包流水、反馈记录、找饭搭资料、红包发送/详情在 API-Go、BFF 和小程序 API 工具中打通。
- 验收证据：
  - `cd services/api-go && go test ./...`
  - `node - <<'NODE' ... NODE` 路由完整性检查：`pages=38`，所有页面 `json/wxml/ts` 存在。
  - `git diff --check`
  - `npm run verify`
- 当前边界：红包当前仅创建/详情和拆分记录首版，尚未做领取、退款、资金冻结/扣减、风控和 PostgreSQL 规范化账本；钱包提现、IM/群聊/私聊、优惠券、积分邀请、处方、跑腿专属费用和轨迹仍待后续批次。
- 文件：
  - `apps/user-wechat-miniprogram/pages/index/index.ts`
  - `apps/user-wechat-miniprogram/pages/index/index.wxml`
  - `apps/user-wechat-miniprogram/pages/index/index.wxss`
  - `apps/user-wechat-miniprogram/pages/shop/list/index.ts`
  - `apps/user-wechat-miniprogram/pages/shop/list/index.wxml`
  - `apps/user-wechat-miniprogram/pages/shop/list/index.wxss`
  - `apps/user-wechat-miniprogram/pages/shop/detail/index.ts`
  - `apps/user-wechat-miniprogram/pages/shop/detail/index.wxml`
  - `apps/user-wechat-miniprogram/pages/shop/detail/index.wxss`
  - `apps/user-wechat-miniprogram/pages/cart/index.ts`
  - `apps/user-wechat-miniprogram/pages/cart/index.wxml`
  - `apps/user-wechat-miniprogram/pages/cart/index.wxss`
  - `apps/user-wechat-miniprogram/pages/order/confirm/index.ts`
  - `apps/user-wechat-miniprogram/pages/order/confirm/index.wxml`
  - `apps/user-wechat-miniprogram/pages/order/confirm/index.wxss`
  - `apps/user-wechat-miniprogram/pages/order/list/index.ts`
  - `apps/user-wechat-miniprogram/pages/order/list/index.wxml`
  - `apps/user-wechat-miniprogram/pages/order/list/index.wxss`
  - `apps/user-wechat-miniprogram/pages/order/detail/index.ts`
  - `apps/user-wechat-miniprogram/pages/order/detail/index.wxml`
  - `apps/user-wechat-miniprogram/pages/order/detail/index.wxss`
  - `apps/user-wechat-miniprogram/pages/order/review/index.ts`
  - `apps/user-wechat-miniprogram/pages/wallet/index.ts`
  - `apps/user-wechat-miniprogram/pages/wallet/index.wxml`
  - `apps/user-wechat-miniprogram/pages/feedback/complaint/index.ts`
  - `apps/user-wechat-miniprogram/pages/feedback/records/index.ts`
  - `apps/user-wechat-miniprogram/pages/meal-match/index.ts`
  - `apps/user-wechat-miniprogram/pages/red-packet/send/index.ts`
  - `apps/user-wechat-miniprogram/pages/red-packet/send/index.wxml`
  - `apps/user-wechat-miniprogram/pages/red-packet/detail/index.ts`
  - `apps/user-wechat-miniprogram/pages/red-packet/detail/index.wxml`
  - `apps/user-wechat-miniprogram/utils/api.ts`
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/bff/src/server.mjs`
  - `docs/product/user-miniprogram-backend-integration.md`
  - `docs/product/frontend-delivery-status.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260527-153 用户端 08-15 精修与圈子消息后端闭环

- 日期：2026-05-27
- 结果：已按参考图精修用户端小程序 `08` 到 `15`：通知偏好补三类偏好卡、渠道开关和静默时间；支付密码改为首次输入/再次确认两步流程并补安全提示；地址列表补三张地址卡、默认标记和编辑/删除/设默认操作；圈子补推荐/附近/商户群/官方 tabs、发布入口、动态流、悬浮发布和底部 tab；找饭搭补介绍卡、检查清单、饭搭预览和安全提示；申请售后补订单摘要、类型、金额、原因、凭证位和处理说明；评价订单补星级、标签、匿名开关、菜品评价和图片位；消息中心补搜索、分段、通知入口、会话列表和底部 tab。已补首批圈子/消息待接后端：圈子动态列表/发布、消息会话列表、聊天记录和发送消息在 API-Go、BFF 和小程序 API 工具中打通。
- 验收证据：
  - `cd services/api-go && go test ./...`
  - `node - <<'NODE' ... NODE` 路由完整性检查：`pages=38`，所有页面 `json/wxml/ts` 存在。
  - `git diff --check`
  - `npm run verify`
- 当前边界：消息当时是内存 Store 首版会话/消息接口，不等于完整生产 IM；离线补偿/已读由 `DONE-20260527-164` 承接，WebSocket 投递由 `DONE-20260527-166` 承接，消息 PostgreSQL 规范化首版由 `DONE-20260529-175` 承接，会话成员权限校验首版由 `DONE-20260529-178` 承接。后续仍需动态群成员/静默设置、客服坐席工作台多端一致性和生产级 IM 网关。圈子发布当前为轻量发布/列表，举报、审核队列、置顶、灰度和风控待补。
- 文件：
  - `apps/user-wechat-miniprogram/pages/notification-preferences/index.ts`
  - `apps/user-wechat-miniprogram/pages/wallet/payment-password/index.ts`
  - `apps/user-wechat-miniprogram/pages/wallet/payment-password/index.wxml`
  - `apps/user-wechat-miniprogram/pages/wallet/payment-password/index.wxss`
  - `apps/user-wechat-miniprogram/pages/address/list/index.ts`
  - `apps/user-wechat-miniprogram/pages/address/list/index.wxml`
  - `apps/user-wechat-miniprogram/pages/address/list/index.wxss`
  - `apps/user-wechat-miniprogram/pages/circle/index.ts`
  - `apps/user-wechat-miniprogram/pages/circle/index.wxml`
  - `apps/user-wechat-miniprogram/pages/circle/index.wxss`
  - `apps/user-wechat-miniprogram/pages/meal-match/index.ts`
  - `apps/user-wechat-miniprogram/pages/meal-match/index.wxml`
  - `apps/user-wechat-miniprogram/pages/meal-match/index.wxss`
  - `apps/user-wechat-miniprogram/pages/after-sales/index.ts`
  - `apps/user-wechat-miniprogram/pages/after-sales/index.wxml`
  - `apps/user-wechat-miniprogram/pages/after-sales/index.wxss`
  - `apps/user-wechat-miniprogram/pages/order/review/index.ts`
  - `apps/user-wechat-miniprogram/pages/order/review/index.wxml`
  - `apps/user-wechat-miniprogram/pages/order/review/index.wxss`
  - `apps/user-wechat-miniprogram/pages/messages/index.ts`
  - `apps/user-wechat-miniprogram/pages/messages/index.wxml`
  - `apps/user-wechat-miniprogram/pages/messages/index.wxss`
  - `apps/user-wechat-miniprogram/pages/messages/merchant-group/index.ts`
  - `apps/user-wechat-miniprogram/pages/messages/merchant-group/index.wxml`
  - `apps/user-wechat-miniprogram/pages/messages/merchant-group/index.wxss`
  - `apps/user-wechat-miniprogram/utils/api.ts`
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/bff/src/server.mjs`
  - `docs/product/user-miniprogram-backend-integration.md`
  - `docs/product/frontend-delivery-status.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260527-154 用户端 16-24 精修与资产服务后端闭环

- 日期：2026-05-27
- 结果：已按参考图精修用户端小程序 `16` 到 `24`：我的补蓝色头图、会员资料卡、资产区、订单快捷、服务宫格、账户安全和底部 tab；钱包补余额蓝卡、充值/提现、资产统计、安全说明、账单筛选和操作列表；红包优惠补优惠摘要、tabs、分类筛选、券卡和规则说明；会员积分补等级、权益、任务、兑换、明细和退款提示；邀请好友补活动卡、邀请码复制、分享和邀请记录；搜索补分类 tabs、猜你想搜、排序条和混合结果卡；买药补校医务室、用药提示、药品分类、AI 问诊、药品列表、处方入口和底部结算栏；跑腿补服务类型、取送信息、物品要求、费用预估和确认下单；跑腿订单详情补状态、地图轨迹、骑手、进度、取送信息、费用和底部操作。
- 后端对接：已补用户概览、钱包总览/提现申请、优惠券列表/兑换、会员积分/签到、邀请摘要、混合搜索、买药首页、跑腿下单/详情在 API-Go、BFF 和小程序 API 工具中的首批闭环；新增 Store 覆盖测试 `TestUserAssetCatalogAndErrandAPIs`。
- 验收证据：
  - `cd services/api-go && go test ./...`
  - `node - <<'NODE' ... NODE` 路由完整性检查：`pages=38`，所有页面 `json/wxml/ts` 存在。
- 当前边界：本轮是内存 Store/首版聚合接口，不等于生产资金和医药能力；真实提现打款回调、优惠券锁定核销、积分邀请反作弊、处方审核、药品订单规范化、跑腿轨迹/补差价和 IM 实时推送仍待后续生产化。
- 文件：
  - `apps/user-wechat-miniprogram/pages/profile/index.ts`
  - `apps/user-wechat-miniprogram/pages/profile/index.wxml`
  - `apps/user-wechat-miniprogram/pages/profile/index.wxss`
  - `apps/user-wechat-miniprogram/pages/wallet/index.ts`
  - `apps/user-wechat-miniprogram/pages/wallet/index.wxml`
  - `apps/user-wechat-miniprogram/pages/wallet/index.wxss`
  - `apps/user-wechat-miniprogram/pages/coupons/index.ts`
  - `apps/user-wechat-miniprogram/pages/coupons/index.wxml`
  - `apps/user-wechat-miniprogram/pages/coupons/index.wxss`
  - `apps/user-wechat-miniprogram/pages/member-points/index.ts`
  - `apps/user-wechat-miniprogram/pages/member-points/index.wxml`
  - `apps/user-wechat-miniprogram/pages/member-points/index.wxss`
  - `apps/user-wechat-miniprogram/pages/invite-friends/index.ts`
  - `apps/user-wechat-miniprogram/pages/invite-friends/index.wxml`
  - `apps/user-wechat-miniprogram/pages/invite-friends/index.wxss`
  - `apps/user-wechat-miniprogram/pages/search/index.ts`
  - `apps/user-wechat-miniprogram/pages/search/index.wxml`
  - `apps/user-wechat-miniprogram/pages/search/index.wxss`
  - `apps/user-wechat-miniprogram/pages/medicine/home/index.ts`
  - `apps/user-wechat-miniprogram/pages/medicine/home/index.wxml`
  - `apps/user-wechat-miniprogram/pages/medicine/home/index.wxss`
  - `apps/user-wechat-miniprogram/pages/errand/home/index.ts`
  - `apps/user-wechat-miniprogram/pages/errand/home/index.wxml`
  - `apps/user-wechat-miniprogram/pages/errand/home/index.wxss`
  - `apps/user-wechat-miniprogram/pages/errand/order-detail/index.ts`
  - `apps/user-wechat-miniprogram/pages/errand/order-detail/index.wxml`
  - `apps/user-wechat-miniprogram/pages/errand/order-detail/index.wxss`
  - `apps/user-wechat-miniprogram/utils/api.ts`
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/bff/src/server.mjs`
  - `docs/product/user-miniprogram-backend-integration.md`
  - `docs/product/frontend-delivery-status.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260527-155 用户端 25-35 精修与客服处方药品闭环

- 日期：2026-05-27
- 结果：已按参考图精修用户端小程序 `25` 到 `35`：商户群聊补群资料、公告、券/订单卡/红包卡、输入栏和快捷面板；发红包补红色头图、拼手气/普通切换、金额/个数/祝福语和塞钱按钮；红包详情补进度、领取记录、钱包流水、领取和退回入口；在线客服补客服状态、场景 tabs、关联订单、聊天、处理建议和快捷入口；工单详情补状态、关联订单、问题描述、处理进度、方案和材料；投诉建议补反馈类型、关联订单、问题说明、影响程度、联系方式和凭证；反馈记录补统计、状态 tabs、搜索筛选、记录卡和帮助卡；上传处方、处方审核结果、药品订单确认、药品订单详情补成校医审核与药品履约链路。
- 后端对接：新增客服工单、红包领取/退回、处方审核、药品订单确认/详情 API，并接入 BFF 白名单和小程序 `utils/api.ts`。新增接口包括 `GET/POST /api/service-tickets`、`GET /api/service-tickets/{ticketID}`、`POST /api/service-tickets/{ticketID}/events`、`POST /api/red-packets/{packetID}/claim`、`POST /api/red-packets/{packetID}/refund`、`POST /api/prescriptions`、`GET /api/prescriptions/{reviewID}`、`POST /api/medicine/orders`、`GET /api/medicine/orders/{orderID}`。
- 验证：
  - `cd services/api-go && go test ./...`
  - `git diff --check`
- 当前边界：本轮是内存 Store/API-Go/BFF/小程序首版闭环，不等于生产资金和医药合规完成。红包资金冻结/扣减、24 小时自动退回、红包风控、客服工作台、处方影像对象存储/OCR/药师审核、药品库存锁定、PostgreSQL 规范化和 IM 实时推送仍待后续生产化。
- 涉及文件：
  - `apps/user-wechat-miniprogram/pages/messages/merchant-group/`
  - `apps/user-wechat-miniprogram/pages/red-packet/`
  - `apps/user-wechat-miniprogram/pages/customer-service/`
  - `apps/user-wechat-miniprogram/pages/service-ticket/`
  - `apps/user-wechat-miniprogram/pages/feedback/`
  - `apps/user-wechat-miniprogram/pages/prescription/`
  - `apps/user-wechat-miniprogram/pages/medicine/order-confirm/`
  - `apps/user-wechat-miniprogram/pages/medicine/order-detail/`
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/bff/src/server.mjs`
  - `docs/product/user-miniprogram-backend-integration.md`
  - `docs/product/frontend-delivery-status.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260527-156 用户端手机号登录注册后端闭环

- 日期：2026-05-27
- 结果：继续按参考图精修与待接后端收口，把登录/注册页从“验证码服务待接入”推进到可测链路。API-Go 新增手机号验证码、手机号验证码/密码登录、手机号注册接口；BFF 白名单放行；小程序登录页和注册页已接真实请求，开发环境会回填 `dev_code` 方便微信开发者工具直接验证，注册页补协议确认状态。
- 后端对接：新增 `POST /api/auth/phone/code`、`POST /api/auth/phone/login`、`POST /api/auth/phone/register`，内存 Store 和 PostgreSQL-backed Store 快照路径均保存手机号绑定、验证码票据和用户密码哈希；手机号注册/登录成功后签发用户 Bearer token，并让 `GET /api/user/profile` 返回绑定手机号。
- 验证：
  - `cd services/api-go && go test ./...`
- 当前边界：这是手机号认证的开发验证码/API 首版；生产短信 provider 与基础频控已由 `DONE-20260527-167` 承接。设备/图形风控、黑名单、验证码审计和生产模板审批仍待后续生产化。
- 涉及文件：
  - `apps/user-wechat-miniprogram/pages/auth/login/`
  - `apps/user-wechat-miniprogram/pages/auth/register/`
  - `apps/user-wechat-miniprogram/utils/api.ts`
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `docs/product/user-miniprogram-backend-integration.md`
  - `docs/product/frontend-delivery-status.md`
  - `apps/user-wechat-miniprogram/README.md`

### DONE-20260527-157 用户端找饭搭候选与安全动作闭环

- 日期：2026-05-27
- 结果：继续按参考图逐页精修与后端补齐，把 `12` 找饭搭从静态“推荐饭搭预览”推进到真实候选列表。页面会在资料满足真实性承诺、免责承诺和问卷后读取候选，展示匹配分、共同饮食/性格标签、距离文案和隐私提示；每张候选卡补“不感兴趣”和“举报”动作。
- 后端对接：新增 `GET /api/meal-match/candidates`、`POST /api/meal-match/reports`、`POST /api/meal-match/blocks`，复用 `GET/PUT /api/meal-match/profile` 前置资料。API-Go 内存 Store 根据饮食习惯和性格交集计算匹配分，拉黑后候选会隐藏；举报生成待审核记录。BFF 已放行新增路由，小程序 `utils/api.ts` 已接候选、举报和拉黑。
- 验证：
  - `cd services/api-go && go test ./...`
  - `npm run test --workspace @infinitech/bff`
- 当前边界：这是找饭搭用户端候选与安全动作首版，不等于生产社交风控完成。资料人工审核和举报处置首版已由 `DONE-20260528-168` 承接，同校/同楼隐私与设备风控首版已由 `DONE-20260528-173` 承接；举报分级策略、真实设备指纹 provider、精准地理隐私 provider 和 PostgreSQL 规范化仍待生产化。
- 涉及文件：
  - `apps/user-wechat-miniprogram/pages/meal-match/index.ts`
  - `apps/user-wechat-miniprogram/pages/meal-match/index.wxml`
  - `apps/user-wechat-miniprogram/pages/meal-match/index.wxss`
  - `apps/user-wechat-miniprogram/utils/api.ts`
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `docs/product/user-miniprogram-backend-integration.md`
  - `docs/product/frontend-delivery-status.md`
  - `apps/user-wechat-miniprogram/README.md`

### DONE-20260527-158 用户端红包资金冻结与过期退回闭环

- 日期：2026-05-27
- 结果：继续按参考图逐页精修与后端补齐，把 `26` 发红包和 `27` 红包详情从“拆分/领取记录首版”推进到资金闭环首版。发红包时服务端会从发包人余额扣减并冻结总金额；领取时释放对应冻结金额并把红包金额入账到领取人余额；发包人手动退回或过期批量退回时，未领取金额回到发包人余额。小程序红包详情页已展示服务端过期状态和退回文案，余额不足时提示先充值。
- 后端对接：扩展 `RedPacket` 领域状态，新增 `expired_refunded`、`claimed_amount_fen`、`refunded_amount_fen`、`expires_at`、`refunded_at`；新增 `AutoRefundExpiredRedPackets` 仓储能力和 `POST /api/admin/red-packets/expire` 管理接口，BFF 已放行。PostgreSQL-backed Store 会持久化红包快照，并通过既有钱包同步把红包冻结、领取和退回流水同步到钱包账本。
- 验证：
  - `cd services/api-go && go test ./...`
  - `npm run test --workspace @infinitech/bff`
- 当前边界：这是红包余额资金首版闭环，不是完整金融生产化。领取风控已由后续 `DONE-20260527-165` 承接；仍需真实 24 小时调度器、支付密码校验、群成员资格、红包 PostgreSQL 规范化表、财务对账和异常补偿。
- 涉及文件：
  - `apps/user-wechat-miniprogram/pages/red-packet/send/index.ts`
  - `apps/user-wechat-miniprogram/pages/red-packet/detail/index.ts`
  - `apps/user-wechat-miniprogram/pages/red-packet/detail/index.wxml`
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/platform/contracts_test.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `docs/product/user-miniprogram-backend-integration.md`
  - `docs/product/frontend-delivery-status.md`
  - `apps/user-wechat-miniprogram/README.md`

### DONE-20260527-159 用户端客服工单分派关闭与回访闭环

- 日期：2026-05-27
- 结果：继续按参考图逐页精修与后端补齐，把 `28` 在线客服、`29` 工单详情、`30` 投诉建议、`31` 反馈记录从“用户提交/补充说明首版”推进到客服处理闭环首版。工单现在支持后台客服列表、分派客服、客服给出处理方案、用户确认关闭和回访评分；小程序工单详情页的“接受方案/关闭工单”已接真实关闭与回访接口。
- 后端对接：扩展 `ServiceTicket` 字段，新增分派客服、处理时间、关闭时间、回访评分/评价；新增 `GET /api/admin/service-tickets`、`POST /api/admin/service-tickets/{ticketID}/assign`、`POST /api/admin/service-tickets/{ticketID}/resolve`、`POST /api/service-tickets/{ticketID}/close`、`POST /api/service-tickets/{ticketID}/follow-up`。BFF 已放行新增路由，PostgreSQL-backed Store 已把反馈/客服工单/工单事件纳入 snapshot 持久化。
- 验证：
  - `cd services/api-go && go test ./internal/platform ./internal/httpapi`
  - `npm run test --workspace @infinitech/bff`
- 当前边界：这是客服工单状态流转和用户侧确认首版，不是完整客服中心。客服工作台可视化已由 `DONE-20260528-169` 承接，SLA 状态与超时升级已由 `DONE-20260528-170` 承接，质检抽检和客服绩效已由 `DONE-20260528-171` 承接；IM 实时推送、PostgreSQL 规范化表和审计同事务仍待后续生产化。
- 涉及文件：
  - `apps/user-wechat-miniprogram/pages/service-ticket/detail/index.ts`
  - `apps/user-wechat-miniprogram/utils/api.ts`
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `docs/product/user-miniprogram-backend-integration.md`
  - `docs/product/frontend-delivery-status.md`
  - `apps/user-wechat-miniprogram/README.md`

### DONE-20260527-160 用户端处方影像上传票据与审核绑定闭环

- 日期：2026-05-27
- 结果：继续按参考图逐页精修与待接后端补齐，把 `32` 上传处方从“本地图片 URL + 直接创建审核单”推进到处方影像上传票据、上传确认和审核单绑定首版。用户端小程序选择处方图片后会先申请对象存储上传票据，再确认对象 key/类型/大小/hash，提交审核时把票据 ID、对象 key 和 hash 一并传给服务端；审核结果保留对象元数据，药品订单继续使用审核通过的处方 ID。
- 后端对接：新增 `POST /api/prescriptions/upload-ticket` 和 `POST /api/prescriptions/upload-confirm`，扩展 `PrescriptionReview` 保存 `image_upload_ticket_id`、`image_object_key`、`image_content_sha`。API-Go 内存 Store 支持处方影像票据签发、确认、防串用校验和审核单绑定；BFF 已放行新增路由；PostgreSQL-backed Store 已把处方影像票据、处方审核单、药品订单详情和处方序列纳入 snapshot 持久化。
- 验证：
  - `cd services/api-go && go test ./internal/platform -run 'TestPrescriptionImageUploadReviewAndMedicineOrder|TestUserAssetCatalogAndErrandAPIs'`
  - `cd services/api-go && go test ./internal/httpapi -run TestPrescriptionImageUploadReviewHTTPFlow`
  - `npm --workspace @infinitech/bff test -- --test-name-pattern "prescription image"`
- 当前边界：这是用户端处方影像提交的对象存储票据首版，不是完整医药合规闭环。OCR 结构化识别、药师/校医工作台、处方留档、药品库存锁定、对象扫描统一接管、处方 PostgreSQL 规范化表和合规审计仍待后续生产化。
- 涉及文件：
  - `apps/user-wechat-miniprogram/pages/prescription/upload/index.ts`
  - `apps/user-wechat-miniprogram/pages/prescription/upload/index.wxml`
  - `apps/user-wechat-miniprogram/pages/prescription/upload/index.wxss`
  - `apps/user-wechat-miniprogram/utils/api.ts`
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `docs/product/user-miniprogram-backend-integration.md`
  - `docs/product/frontend-delivery-status.md`
  - `apps/user-wechat-miniprogram/README.md`
  - `README.md`

### DONE-20260527-161 用户端药品订单库存锁定闭环

- 日期：2026-05-27
- 结果：继续按参考图逐页精修与待接后端补齐，把 `22` 买药首页、`34` 药品订单确认、`35` 药品订单详情的库存能力补成首版闭环。买药首页从 Store 库存读取剩余库存；药品订单创建会校验并锁定库存，订单项返回 `stock_locked` 和 `stock_remaining`；库存不足返回 HTTP 409 `INSUFFICIENT_STOCK`，小程序确认页会提示用户返回调整药品。
- 后端对接：API-Go 内存 Store 新增药品库存账本和锁定逻辑，PostgreSQL-backed Store snapshot 纳入 `medicine_stock`；HTTP 错误映射新增 `INSUFFICIENT_STOCK`；已补 Store 与 HTTP 库存锁定测试。
- 验证：
  - `cd services/api-go && go test ./internal/platform -run 'TestMedicineOrderLocksStockAndRejectsInsufficientInventory|TestPrescriptionImageUploadReviewAndMedicineOrder|TestUserAssetCatalogAndErrandAPIs'`
  - `cd services/api-go && go test ./internal/httpapi -run 'TestMedicineOrderInventoryHTTPFlow|TestPrescriptionImageUploadReviewHTTPFlow'`
- 当前边界：这是药品库存锁定首版，不是完整医药库存/药房系统。药品批次、效期、处方留档、OCR/药师工作台、对象扫描统一接管、药品 PostgreSQL 规范化表和合规审计仍待后续生产化。
- 涉及文件：
  - `apps/user-wechat-miniprogram/pages/medicine/order-confirm/index.ts`
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `docs/product/user-miniprogram-backend-integration.md`
  - `docs/product/frontend-delivery-status.md`
  - `apps/user-wechat-miniprogram/README.md`
  - `README.md`

### DONE-20260527-162 用户端处方 OCR 留档与药师复核闭环

- 日期：2026-05-27
- 结果：继续按参考图逐页精修与待接后端补齐，把 `32` 上传处方、`33` 处方审核结果、`34-35` 药品履约依赖的处方审核能力补成 OCR/留档/药师复核首版。创建处方审核单时会生成 OCR 识别摘要、识别剂量、置信度、处方留档编号和 6 年留档信息；小程序审核结果页展示 OCR 状态、剂量和留档编号；药师后台可按状态查询处方队列并复核通过/驳回，用户侧查询会同步看到复核结果。
- 后端对接：新增 `GET /api/admin/prescriptions`、`POST /api/admin/prescriptions/{reviewID}/review`，BFF 已放行；API-Go `PrescriptionReview` 扩展 `ocr_result` 和 `archive`，Store 支持药师复核状态流转，PostgreSQL-backed Store 通过 snapshot 持久化新增字段。
- 验证：
  - `cd services/api-go && go test ./internal/platform -run 'TestPrescriptionImageUploadReviewAndMedicineOrder|TestMedicineOrderLocksStockAndRejectsInsufficientInventory'`
  - `cd services/api-go && go test ./internal/httpapi -run 'TestPrescriptionImageUploadReviewHTTPFlow'`
  - `npm run test --workspace @infinitech/bff -- --test-name-pattern "prescription"`
- 当前边界：这是 OCR 摘要、处方留档和药师复核工作台首版，不是完整医药合规系统。对象扫描统一接管、OCR 真实 provider、药品批次效期、处方 PostgreSQL 规范化表、合规审计同事务和药房工作台可视化仍待后续生产化。
- 涉及文件：
  - `apps/user-wechat-miniprogram/pages/prescription/review-result/index.ts`
  - `apps/user-wechat-miniprogram/pages/prescription/review-result/index.wxml`
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/httpapi/auth.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `docs/product/user-miniprogram-backend-integration.md`
  - `docs/product/frontend-delivery-status.md`
  - `apps/user-wechat-miniprogram/README.md`
  - `README.md`

### DONE-20260527-163 用户端处方影像对象扫描门禁统一接管

- 日期：2026-05-27
- 结果：继续按参考图逐页精修与待接后端补齐，把 `32` 上传处方的处方影像从“票据确认”推进到和售后凭证共用对象存储 upload callback + scan result 门禁。启用 `RequireUploadCallbackForConfirm` / `RequireScanApprovalForConfirm` 后，处方影像必须先收到对象存储上传回调，再收到扫描通过结果，才允许 `upload-confirm` 和后续处方审核；扫描驳回或未扫描会阻断确认。小程序上传处方页新增“安全扫描中/安全扫描通过/扫描未通过”状态展示。
- 后端对接：`PrescriptionImageUploadTicket` 新增 `scan_status`、`scan_result` 和 `scan_checked_at`；通用 `POST /api/object-storage/upload-callback`、`POST /api/object-storage/scan-result` 现在可同时更新售后凭证和处方影像票据；BFF 已放行两个对象存储回调路由；默认开发模式仍可直接确认，生产开启扫描门禁后按统一流程拦截。
- 验证：
  - `cd services/api-go && go test ./internal/platform -run 'TestPrescriptionImageUploadReviewAndMedicineOrder|TestPrescriptionImageUploadRequiresUnifiedObjectScanApproval|TestAfterSalesEvidenceRequiresUploadCallbackAndScanApproval'`
  - `npm run test --workspace @infinitech/bff -- --test-name-pattern "prescription image upload"`
- 当前边界：这是处方影像对象扫描门禁的统一接管首版；真实 OCR provider、药品批次效期、处方 PostgreSQL 规范化表、合规审计同事务和药房工作台可视化仍待后续生产化。
- 涉及文件：
  - `apps/user-wechat-miniprogram/pages/prescription/upload/index.ts`
  - `apps/user-wechat-miniprogram/pages/prescription/upload/index.wxml`
  - `apps/user-wechat-miniprogram/pages/prescription/upload/index.wxss`
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `docs/product/user-miniprogram-backend-integration.md`
  - `docs/product/frontend-delivery-status.md`
  - `apps/user-wechat-miniprogram/README.md`
  - `README.md`

### DONE-20260527-164 用户端消息离线补偿与已读回执闭环

- 日期：2026-05-27
- 结果：继续按参考图逐页精修与待接后端补齐，把用户端消息中心/商户群聊从“列表、聊天、发送”推进到可补离线消息和可回写已读的首版闭环。API-Go 新增 `GET /api/messages/{threadID}/sync?since_id=&mark_read=` 和 `POST /api/messages/{threadID}/read`，会话摘要返回 `last_message_id`、`last_read_message_id`、`last_read_at`，打开会话可按游标拉取增量消息并清未读；发送聊天消息会写入 `message.sent` outbox 事件，给后续 realtime gateway/WebSocket 投递承接。PostgreSQL-backed Store 快照路径已持久化聊天消息与已读状态，避免本地重启丢失首版演示数据。
- 小程序对接：`pages/messages/merchant-group/index` 改为优先调用同步接口，按 `lastMessageId` 合并增量消息，打开会话时自动标记已读，发送成功后更新本地游标；`utils/api.ts` 新增 `syncChatMessages` 和 `markChatThreadRead`。
- BFF 对接：已放行消息同步和已读回执路由，覆盖 `GET /api/messages/{threadID}/sync` 与 `POST /api/messages/{threadID}/read`。
- 验证：
  - `cd services/api-go && go test ./internal/platform -run 'TestChatMessageSyncMarksReadAndQueuesRealtimeOutbox'`
  - `cd services/api-go && go test ./internal/httpapi -run 'TestChatSyncReadAndRealtimeOutboxHTTPFlow'`
  - `npm run test --workspace @infinitech/bff -- --test-name-pattern "chat sync"`
- 当前边界：这是用户端消息离线补偿、已读回执和实时出站事件的首版；仍不是完整生产 IM。WebSocket 投递已由 `DONE-20260527-166` 承接，消息 PostgreSQL 规范化首版已由 `DONE-20260529-175` 承接，会话成员权限校验首版已由 `DONE-20260529-178` 承接；后续还需要动态群成员/静默设置、客服坐席工作台多端一致性和跨端消息一致性。
- 涉及文件：
  - `apps/user-wechat-miniprogram/pages/messages/merchant-group/index.ts`
  - `apps/user-wechat-miniprogram/utils/api.ts`
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `docs/product/user-miniprogram-backend-integration.md`
  - `docs/product/frontend-delivery-status.md`
  - `apps/user-wechat-miniprogram/README.md`
  - `README.md`

### DONE-20260527-165 用户端红包领取风控闭环

- 日期：2026-05-27
- 结果：继续按参考图逐页精修与待接后端补齐，把 `27` 红包详情从“可领取/退回/过期退回”推进到领取风控首版。API-Go 在 `POST /api/red-packets/{packetID}/claim` 领取前执行同群短时频次和 24 小时累计金额校验：同一用户同一群 10 分钟内最多领取 3 次，24 小时累计领取金额超过上限会拦截；同一个红包重复点击领取保持幂等返回，不因频次规则误伤。风控拦截通过 `RISK_CONTROL_REJECTED` 和 429 返回，红包详情会返回 `risk` 状态与原因。
- 小程序对接：`pages/red-packet/detail/index` 新增“领取校验”提示卡，展示服务端返回的风控状态；领取失败时会把后端原因写入页面提示，并用 toast 告知用户。
- 验证：
  - `cd services/api-go && go test ./internal/platform -run 'TestMerchantRefundGroupAndRedPacketConstants|TestRedPacketWalletFreezeClaimAndAutoRefund|TestRedPacketClaimRiskControlsFrequencyAndIdempotency'`
  - `cd services/api-go && go test ./internal/httpapi -run 'TestRedPacketWalletAndExpiryHTTPFlow|TestRedPacketClaimRiskHTTPFlow'`
- 当前边界：这是红包领取频次和金额风控首版，不是完整资金风控系统；真实设备指纹、IP/地理异常、群成员资格、支付密码校验、人工复核队列、红包 PostgreSQL 规范化表、财务对账和异常补偿仍待后续生产化。
- 涉及文件：
  - `apps/user-wechat-miniprogram/pages/red-packet/detail/index.ts`
  - `apps/user-wechat-miniprogram/pages/red-packet/detail/index.wxml`
  - `apps/user-wechat-miniprogram/pages/red-packet/detail/index.wxss`
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/platform/contracts_test.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `docs/product/user-miniprogram-backend-integration.md`
  - `docs/product/frontend-delivery-status.md`
  - `apps/user-wechat-miniprogram/README.md`
  - `README.md`

### DONE-20260527-166 用户端消息 WebSocket 实时投递闭环

- 日期：2026-05-27
- 结果：继续按参考图逐页精修与待接后端补齐，把用户端商户群聊从“发送写 outbox”推进到 WebSocket 首版实时投递。`services/realtime-gateway` 新增 `/ws?thread_id=&user_id=` WebSocket 升级、连接注册、订阅过滤和 `/internal/realtime/publish` 内部发布入口；`message.sent` 事件会按 `thread_id` 投递给订阅同会话的客户端。
- Outbox 对接：`outbox-relay-worker` 默认包含 `message.sent` topic，新增 realtime gateway publisher 和 topic routing publisher；配置 `REALTIME_GATEWAY_URL` 后会把 `message.sent` POST 到实时网关，其他 topic 继续走 Kafka REST 或 console fallback。
- 小程序对接：`utils/api.ts` 新增 realtime base URL 与 socket URL helper；`pages/messages/merchant-group/index` 打开会话时建立 WebSocket，收到同会话 `message.sent` 后合并到消息流、更新游标和实时状态。
- 部署对接：Docker Compose 与 K8s base 已新增 realtime-gateway 服务/部署环境变量，outbox relay 已配置 `REALTIME_GATEWAY_URL`、`REALTIME_INTERNAL_TOKEN` 和 `message.sent` topic。
- 验证：
  - `npm run test --workspace @infinitech/realtime-gateway`
  - `npm run test --workspace @infinitech/outbox-relay-worker`
  - `npm run verify:architecture`
- 当前边界：这是单机/轻量 WebSocket 投递首版，不是完整生产 IM 网关；消息 PostgreSQL 规范化首版已由 `DONE-20260529-175` 承接，Redis adapter 多副本 fanout 首版已由 `DONE-20260529-176` 承接，WebSocket 签名 token 鉴权首版已由 `DONE-20260529-177` 承接。断线重连/心跳策略、消息顺序保障和压测仍待后续生产化。
- 涉及文件：
  - `services/realtime-gateway/src/server.mjs`
  - `services/realtime-gateway/src/runtime.mjs`
  - `services/realtime-gateway/src/runtime.test.mjs`
  - `services/outbox-relay-worker/src/index.mjs`
  - `services/outbox-relay-worker/src/index.test.mjs`
  - `apps/user-wechat-miniprogram/utils/api.ts`
  - `apps/user-wechat-miniprogram/pages/messages/merchant-group/index.ts`
  - `apps/user-wechat-miniprogram/pages/messages/merchant-group/index.wxml`
  - `infra/docker/compose.yml`
  - `infra/k8s/base/app-stack.yaml`
  - `scripts/check-architecture.mjs`
  - `docs/product/user-miniprogram-backend-integration.md`
  - `docs/product/frontend-delivery-status.md`
  - `apps/user-wechat-miniprogram/README.md`
  - `README.md`

### DONE-20260527-167 用户端登录注册短信验证码生产化首版

- 日期：2026-05-27
- 结果：继续按参考图逐页精修与待接后端补齐，把用户端登录/注册手机号验证码从“开发验证码回填”推进到生产短信 provider 首版。API-Go 新增 `PhoneVerificationConfig`、provider dispatch request/result、`PHONE_VERIFICATION_MODE=provider`、`SMS_PROVIDER_ENDPOINT`、`SMS_PROVIDER_TOKEN`、`SMS_TEMPLATE_PHONE_CODE` 等配置；生产模式下验证码不会返回给小程序，只发送给 provider，开发模式仍保留 `dev_code` 回填。
- 风控对接：手机号验证码新增重发冷却、单手机号小时/日频控和 `RATE_LIMITED` 429 响应；PostgreSQL snapshot 已保留验证码请求节流状态，避免重启后短时绕过首版频控。
- 小程序对接：登录页和注册页新增验证码发送状态提示，展示重发冷却；生产短信模式不回填验证码，开发模式继续自动填入验证码便于本地预览。
- 部署对接：K8s API-Go 部署骨架已加入短信 provider 环境变量，架构守卫新增“生产 SMS provider 模式”断言。
- 验证：
  - `cd services/api-go && go test ./internal/platform -run 'TestPhoneCodeRegisterAndLogin|TestPhoneCodeProviderModeHidesCodeAndRateLimits'`
  - `cd services/api-go && go test ./internal/httpapi -run 'TestPhoneRegisterAndLoginIssuesSignedToken|TestPhoneCodeProviderModeHTTPFlow'`
  - `cd services/api-go && go test ./...`
  - `npm run verify:architecture`
  - `git diff --check`
- 当前边界：这是短信验证码 provider 调用和基础频控首版，不是完整短信风控系统；真实短信模板审批、图形验证码/设备指纹、IP 黑名单、验证码审计归档、provider 回执重试和异常告警仍待后续生产化。
- 涉及文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/cmd/api/main.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `apps/user-wechat-miniprogram/utils/api.ts`
  - `apps/user-wechat-miniprogram/pages/auth/login/index.ts`
  - `apps/user-wechat-miniprogram/pages/auth/login/index.wxml`
  - `apps/user-wechat-miniprogram/pages/auth/login/index.wxss`
  - `apps/user-wechat-miniprogram/pages/auth/register/index.ts`
  - `apps/user-wechat-miniprogram/pages/auth/register/index.wxml`
  - `infra/k8s/base/app-stack.yaml`
  - `scripts/check-architecture.mjs`
  - `docs/product/user-miniprogram-backend-integration.md`
  - `docs/product/frontend-delivery-status.md`
  - `apps/user-wechat-miniprogram/README.md`
  - `README.md`

### DONE-20260528-168 用户端找饭搭人工审核与举报处置闭环

- 日期：2026-05-28
- 结果：继续按参考图逐页精修与待接后端补齐，把 `12` 找饭搭从“资料完成后直接可用”推进到资料提交后进入人工审核。小程序会展示“待完善 / 审核中 / 审核未通过 / 可开启”，后端未放行时不再保留种子候选卡，而是显示审核或资料缺口提示；审核通过后才返回候选，举报成立后目标资料会暂停展示。
- 后端对接：`MealMatchProfile` 新增审核状态、审核原因、审核记录 ID、审核人和审核时间；`SaveMealMatchProfile` 在前置资料完整后自动创建 `profile_review` 待审记录；新增 `GET /api/admin/meal-match/moderation` 和 `POST /api/admin/meal-match/moderation/{recordID}/review`，支持按状态/动作/用户筛选并审批资料或举报。BFF 已放行管理端审核路由，服务端 RBAC 新增 `meal_match:read` 与 `meal_match:review` scope，`ops_admin` 与 `support_admin` 可处理队列。
- 验证：
  - `cd services/api-go && go test ./internal/platform -run 'Test(CircleMealMatchCouponAndHomeCards|MealMatchCandidatesReportAndBlock)$'`
  - `cd services/api-go && go test ./internal/httpapi -run 'Test(BackofficeRBACScopeMatrix|BackofficeRBACPolicyCatalog|MealMatchCandidatesReportAndBlockHTTPFlow)$'`
  - `npm run test --workspace @infinitech/bff -- --test-name-pattern "meal match"`
  - `npm run verify:architecture -- --test-name-pattern "meal match moderation"`
- 当前边界：这是找饭搭资料人工审核和举报处置首版，不是完整社交风控系统。真实同校/同楼隐私策略、举报分级策略、设备指纹、精准地理隐私、审核工作台可视化和 PostgreSQL 规范化仍待后续生产化。
- 涉及文件：
  - `apps/user-wechat-miniprogram/pages/meal-match/index.ts`
  - `apps/user-wechat-miniprogram/pages/meal-match/index.wxml`
  - `apps/user-wechat-miniprogram/pages/meal-match/index.wxss`
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/httpapi/auth.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/platform/contracts_test.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/auth_test.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `scripts/check-architecture.mjs`
  - `docs/product/user-miniprogram-backend-integration.md`
  - `docs/product/frontend-delivery-status.md`
  - `apps/user-wechat-miniprogram/README.md`
  - `README.md`

### DONE-20260528-169 管理端客服工单工作台可视化首版

- 日期：2026-05-28
- 结果：继续按参考图逐页精修与后端能力补齐，把用户端反馈/客服工单的“后台可处理面”补到 Admin Web。管理端 `support` 模块从计划态推进到可打开的客服工作台，新增工单 KPI、队列、工单表格、详情抽屉核查清单和下一步动作。
- 后端对接：Admin Web 操作目录已接 `GET /api/admin/service-tickets`、`POST /api/admin/service-tickets/{ticketID}/assign`、`POST /api/admin/service-tickets/{ticketID}/resolve`；客服工作台详情抽屉可预填工单状态、客服分派和处理方案。客服分派、提交方案都纳入高风险二次确认和失败回放。
- 权限与守卫：`support_admin` 前端角色矩阵补 `service_ticket:read`、`service_ticket:write`；架构守卫固定 Admin Web 配置、视图、详情、操作、BFF 代理和测试覆盖，避免客服工作台回退成纯占位。
- 验证：
  - `npm run test --workspace @infinitech/admin-web`
  - `npm run test --workspace @infinitech/bff -- --test-name-pattern "service ticket"`
  - `npm run verify:architecture -- --test-name-pattern "admin web|bff keeps"`
  - `git diff --check`
  - `npm run verify`
  - 本地 Admin Web 冒烟：打开 `http://127.0.0.1:4173/apps/admin-web/index.html`，进入“客服工作台”，确认 `待质检`、`绩效风险`、`抽检客服工单`、`客服质检记录`、`客服绩效汇总`可见且控制台 0 error。
  - 本地 Admin Web 冒烟：打开 `http://127.0.0.1:4173/apps/admin-web/index.html`，进入“客服工作台”，确认 `SLA 超时`、`10 分钟首响/超时升级`、`升级客服工单`可见且控制台 0 error。
- 当前边界：本轮补的是管理端可视化和已有服务工单 API 的操作承接，不等于完整客服中心生产化。SLA 自动升级已由 `DONE-20260528-170` 承接，质检抽检和客服绩效已由 `DONE-20260528-171` 承接，客服消息风控已由 `DONE-20260528-172` 承接；后续仍需服务工单 PostgreSQL 规范化表和审计同事务强化。
- 涉及文件：
  - `apps/admin-web/src/config.mjs`
  - `apps/admin-web/src/adminApi.mjs`
  - `apps/admin-web/src/adminOperations.mjs`
  - `apps/admin-web/src/adminViews.mjs`
  - `apps/admin-web/src/adminDetails.mjs`
  - `apps/admin-web/src/adminApi.test.mjs`
  - `apps/admin-web/README.md`
  - `scripts/check-architecture.mjs`
  - `docs/product/user-miniprogram-backend-integration.md`
  - `docs/product/frontend-delivery-status.md`
  - `docs/product/recent-progress-roadmap.md`
  - `README.md`

### DONE-20260528-170 用户端客服工单 SLA 状态与超时升级闭环

- 日期：2026-05-28
- 结果：继续按参考图逐页精修与后端能力补齐，把客服工单从“人工处理首版”推进到 SLA 状态/超时升级首版。服务工单新增 `sla_status`、升级级别、升级原因、升级时间；后台可按 SLA 状态筛选工单，小程序工单详情页展示预计更新时间、SLA 状态和升级说明。
- 后端对接：新增 `POST /api/admin/service-tickets/{ticketID}/escalate`；`GET /api/admin/service-tickets` 新增 `sla_status` 和 `now` 查询参数；Store 会按 `reply_due_at` 自动计算 `normal`、`due_soon`、`overdue`、`escalated`、`completed`，分派后重置首响 SLA，处理方案/关闭后标记 completed。BFF 已放行升级路由。
- 管理端对接：Admin Web 客服工作台新增 `support-ticket-escalate` 操作、SLA 筛选项、超时队列指标、详情抽屉预填升级原因；客服分派、SLA 升级、处理方案都进入高风险二次确认。
- 验证：
  - `cd services/api-go && go test ./internal/platform ./internal/httpapi`
  - `npm run test --workspace @infinitech/bff -- --test-name-pattern "service ticket"`
  - `npm run test --workspace @infinitech/admin-web`
  - `npm run verify:architecture -- --test-name-pattern "admin web|bff keeps"`
  - `git diff --check`
  - `npm run verify`
  - `npm run verify`
- 当前边界：这是 SLA 状态和人工升级首版，不是完整客服中心。质检抽检和客服绩效已由 `DONE-20260528-171` 承接，客服消息风控已由 `DONE-20260528-172` 承接；后续仍需 SLA 自动巡检 worker、服务工单 PostgreSQL 规范化表和审计同事务强化。
- 涉及文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `apps/admin-web/src/adminApi.mjs`
  - `apps/admin-web/src/adminOperations.mjs`
  - `apps/admin-web/src/adminViews.mjs`
  - `apps/admin-web/src/adminDetails.mjs`
  - `apps/admin-web/src/adminApi.test.mjs`
  - `apps/user-wechat-miniprogram/pages/service-ticket/detail/index.ts`
  - `apps/user-wechat-miniprogram/pages/service-ticket/detail/index.wxml`
  - `apps/user-wechat-miniprogram/pages/service-ticket/detail/index.wxss`
  - `scripts/check-architecture.mjs`
  - `docs/product/user-miniprogram-backend-integration.md`
  - `docs/product/frontend-delivery-status.md`
  - `docs/product/recent-progress-roadmap.md`
	  - `apps/admin-web/README.md`
	  - `README.md`

### DONE-20260528-171 客服质检抽检与客服绩效首版

- 日期：2026-05-28
- 结果：继续按参考图逐页精修与后端能力补齐，把客服工单从“SLA/处理闭环”推进到内部运营闭环。新增客服质检记录、质检结果、质检分、辅导标记和客服绩效汇总，Admin Web 客服工作台可直接抽检工单、查看质检记录和查看客服绩效。
- 后端对接：新增 `POST /api/admin/service-tickets/{ticketID}/quality-review`、`GET /api/admin/service-ticket-quality-reviews`、`GET /api/admin/service-ticket-performance`；Store 会从服务工单、SLA 状态、回访评分和质检记录计算客服绩效，PostgreSQL-backed Store 已把质检记录纳入 snapshot 持久化。BFF 已放行质检和绩效路由。
- 管理端对接：Admin Web 操作目录新增 `support-quality-review`、`support-quality-reviews`、`support-performance`；客服工作台指标从“待确认”推进到“待质检/绩效风险”，详情抽屉可预填抽检分数、结果、辅导标记、质检记录筛选和客服绩效查询。质检写动作进入高风险二次确认。
- 验证：
  - `cd services/api-go && go test ./internal/platform ./internal/httpapi`
  - `npm run test --workspace @infinitech/bff -- --test-name-pattern "service ticket"`
  - `npm run test --workspace @infinitech/admin-web`
  - `npm run verify:architecture -- --test-name-pattern "admin web|bff keeps"`
  - `git diff --check`
  - `npm run verify`
  - 本地 Admin Web 冒烟：打开 `http://127.0.0.1:4173/apps/admin-web/index.html`，进入“客服工作台”，确认 `待质检`、`绩效风险`、`抽检客服工单`、`客服质检记录`、`客服绩效汇总`可见且控制台 0 error。
- 当前边界：这是客服质检/绩效首版，不是完整客服中心生产化。客服消息风控已由 `DONE-20260528-172` 承接；后续仍需 SLA 自动巡检 worker、服务工单/质检 PostgreSQL 规范化表、质检审计同事务强化、客服绩效产品化配置和真实 IM 生产化。
- 涉及文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/server.mjs`
  - `services/bff/src/runtime.test.mjs`
  - `apps/admin-web/src/adminApi.mjs`
  - `apps/admin-web/src/adminOperations.mjs`
  - `apps/admin-web/src/adminViews.mjs`
  - `apps/admin-web/src/adminDetails.mjs`
  - `apps/admin-web/src/adminApi.test.mjs`
  - `apps/admin-web/src/main.js`
  - `apps/admin-web/index.html`
  - `scripts/check-architecture.mjs`
  - `docs/product/user-miniprogram-backend-integration.md`
  - `docs/product/frontend-delivery-status.md`
  - `docs/product/recent-progress-roadmap.md`
  - `apps/admin-web/README.md`
  - `apps/user-wechat-miniprogram/README.md`
  - `README.md`

### DONE-20260528-172 客服消息敏感信息风控首版

- 日期：2026-05-28
- 结果：继续按参考图逐页精修与后端能力补齐，把“客服不会索要支付密码、验证码或银行卡信息”的页面提示补成可执行能力。IM 聊天消息、客服工单创建内容和客服工单补充事件都会执行敏感信息风控；正常内容标记 `passed`，敏感提及标记 `flagged`，疑似泄露验证码、支付密码或银行卡号时返回 `RISK_CONTROL_REJECTED` 并不写入消息/outbox/工单事件。
- 后端对接：`ChatMessage`、`ServiceTicket`、`ServiceTicketEvent` 新增 `risk_state`、`risk_reason_code`、`risk_reason`、`risk_checked_at`；`SendChatMessage`、`CreateServiceTicket`、`AddServiceTicketEvent` 共用风控判定，`message.sent` outbox payload 会带风险结果，BFF 继续放行聊天发送/同步/已读路由。
- 用户端对接：`pages/customer-service/chat/index` 新增发送前本地提示与拦截卡片，命中敏感数字内容时清空输入并显示“已拦截”，敏感词提及但未泄露时允许发送并提醒用户不要提供敏感信息。
- 管理端对接：Admin Web 客服工作台说明和防护清单已纳入“消息风控”，架构守卫固定后端、BFF、小程序和管理端路径。
- 验证：
  - `cd services/api-go && go test ./internal/platform ./internal/httpapi`
  - `npm run test --workspace @infinitech/bff -- --test-name-pattern "chat sync"`
  - `npm run test --workspace @infinitech/admin-web`
  - `npm run verify:architecture -- --test-name-pattern "customer service message risk"`
  - `git diff --check`
  - `npm run verify`
- 当前边界：这是敏感信息关键字/数字披露风控首版，不是完整客服中心生产化。消息 PostgreSQL 规范化首版已由 `DONE-20260529-175` 承接；后续仍需模型化内容安全、客服质检审计同事务、服务工单 PostgreSQL 规范化表、真实 IM 多端会话权限和多副本实时网关压测。
- 涉及文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/runtime.test.mjs`
  - `apps/user-wechat-miniprogram/pages/customer-service/chat/index.ts`
  - `apps/user-wechat-miniprogram/pages/customer-service/chat/index.wxml`
  - `apps/user-wechat-miniprogram/pages/customer-service/chat/index.wxss`
  - `apps/admin-web/src/adminViews.mjs`
  - `scripts/check-architecture.mjs`
  - `docs/product/user-miniprogram-backend-integration.md`
  - `docs/product/frontend-delivery-status.md`
  - `docs/product/recent-progress-roadmap.md`
  - `apps/user-wechat-miniprogram/README.md`
  - `README.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260528-173 用户端找饭搭同校隐私与设备风控首版

- 日期：2026-05-28
- 结果：继续按参考图逐页精修与待接后端补齐，把找饭搭从“人工审核后可见候选”推进到同校/同楼隐私放行和设备环境风控。资料必须携带学校、楼栋、隐私范围、位置精度和设备 ID；候选只在同校范围展示，双方任一方选择同楼保护时必须同楼才展示；候选位置不再暴露精确距离，只展示“同楼可约 / 同校范围 / 已隐藏位置”。
- 后端对接：`MealMatchProfile` 新增学校、校区、楼栋、隐私范围、位置精度、设备 ID 和设备风险字段；`MealMatchCandidate` 与候选列表返回同校/同楼、隐私提示和设备风控结果。`SaveMealMatchProfile` 会执行设备风控，缺失设备或共享设备进入人工复核，已知风险设备返回 `RISK_CONTROL_REJECTED`；审核通过后可把待复核设备标记为 `passed`。
- 用户端对接：`pages/meal-match/index` 新增本地设备 ID 生成与持久化，提交资料时带同校/同楼隐私字段；页面新增“隐私与设备安全”卡片，候选卡显示模糊位置和隐私范围，不再展示精确米数。
- 架构守卫：固定 API contracts、Store、HTTP 测试、BFF 测试和小程序页面的隐私/设备风控字段，防止后续回退成裸候选列表。
- 验证：
  - `cd services/api-go && go test ./internal/platform -run 'Test(CircleMealMatchCouponAndHomeCards|MealMatchCandidatesReportAndBlock)$'`
  - `cd services/api-go && go test ./internal/httpapi -run 'TestMealMatchCandidatesReportAndBlockHTTPFlow$'`
  - `cd services/api-go && go test ./internal/platform ./internal/httpapi`
  - `npm run test --workspace @infinitech/bff -- --test-name-pattern "meal match"`
  - `npm run verify:architecture -- --test-name-pattern "meal match moderation"`
- 当前边界：这是同校/同楼隐私和轻量设备风控首版，不是完整社交风控系统。后续仍需真实设备指纹 provider、IP/地理异常、举报分级策略、精准地理隐私 provider、找饭搭 PostgreSQL 规范化表和审核工作台更细颗粒度可视化。
- 涉及文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/contracts_test.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/runtime.test.mjs`
  - `apps/user-wechat-miniprogram/pages/meal-match/index.ts`
  - `apps/user-wechat-miniprogram/pages/meal-match/index.wxml`
  - `apps/user-wechat-miniprogram/pages/meal-match/index.wxss`
  - `scripts/check-architecture.mjs`
  - `docs/product/user-miniprogram-backend-integration.md`
  - `docs/product/frontend-delivery-status.md`
  - `docs/product/recent-progress-roadmap.md`
  - `apps/user-wechat-miniprogram/README.md`
  - `README.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260528-174 用户端真实微信 wx.login 页面流程首版

- 日期：2026-05-28
- 结果：继续按参考图逐页精修与待接后端补齐，把登录/注册页从“接口失败直接进入预览 token”推进到真实微信 `wx.login` 页面流程。登录页和注册页共用 `getWechatLoginCode()` 获取真实 code，再通过 `POST /api/auth/wechat-mini/login` 换取服务端 token；生产接口失败时停留当前页并提示，不再自动写入开发 token。
- 小程序对接：`utils/api.ts` 新增 `getWechatLoginCode()`、`isPreviewAuthAllowed()` 和 `activatePreviewAuth()`；开发 token 兜底仅允许本地 API 或显式 `allowPreviewAuth=true`。登录/注册页面新增微信登录状态提示，手机号登录/注册接口失败也不再在生产环境自动伪登录。
- 后端边界：API-Go 的 `code2session` provider resolver 和 BFF 白名单沿用既有实现；本轮重点收口页面生产行为，避免真实环境登录失败被本地预览 token 掩盖。
- 验证：
  - `npm run verify:architecture -- --test-name-pattern "phone verification"`
  - `git diff --check`
- 当前边界：这是微信登录页面生产兜底收口首版；仍需微信开发者工具真机走查、真实小程序 AppID/secret 联调、手机号一键登录/头像昵称授权策略和登录失败埋点。
- 涉及文件：
  - `apps/user-wechat-miniprogram/utils/api.ts`
  - `apps/user-wechat-miniprogram/pages/auth/login/index.ts`
  - `apps/user-wechat-miniprogram/pages/auth/login/index.wxml`
  - `apps/user-wechat-miniprogram/pages/auth/login/index.wxss`
  - `apps/user-wechat-miniprogram/pages/auth/register/index.ts`
  - `apps/user-wechat-miniprogram/pages/auth/register/index.wxml`
  - `scripts/check-architecture.mjs`
  - `docs/product/user-miniprogram-backend-integration.md`
  - `docs/product/frontend-delivery-status.md`
  - `docs/product/recent-progress-roadmap.md`
  - `apps/user-wechat-miniprogram/README.md`
  - `README.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260529-175 用户端消息 PostgreSQL 规范化首版

- 日期：2026-05-29
- 结果：继续按参考图逐页精修与待接后端补齐，把用户端消息/商户群聊从快照恢复推进到 PostgreSQL 规范化消息表和已读状态表。`PostgresStore` 启动会先把快照消息同步到 `conversations`、`conversation_members`、`messages` 和 `conversation_read_states`，再从规范化表恢复会话消息与已读游标，避免只依赖单个 JSON snapshot。
- 后端对接：`messages` 新增 `sender_name`、`risk_state`、`risk_reason_code`、`risk_reason`、`risk_checked_at`；`conversation_read_states` 存储 user/thread 已读游标。`persistAfter` 会同步消息、已读状态和 outbox，保持 `message.sent` 实时投递事件和风险字段可恢复。
- 数据库对接：新增 `infra/db/migrations/0005_chat_messages.sql`，补消息风险字段、发送人索引、会话时间索引和已读状态表；架构守卫固定消息 PostgreSQL 规范化、读回执表、前端同步和实时 Socket 接入。
- 验证：
  - `cd services/api-go && go test ./internal/platform -run 'TestChatMessageSyncMarksReadAndQueuesRealtimeOutbox|TestStoreSnapshotRoundTripRestoresStateAndIndexes'`
  - `cd services/api-go && go test ./internal/platform ./internal/httpapi`
  - `npm run verify:architecture -- --test-name-pattern "chat messages"`
- 当前边界：这是消息 PostgreSQL 规范化首版，不是完整生产 IM。Redis adapter 多副本 fanout 首版已由 `DONE-20260529-176` 承接，WebSocket 签名 token 鉴权首版已由 `DONE-20260529-177` 承接，会话成员权限校验首版已由 `DONE-20260529-178` 承接；后续仍需动态群成员/静默策略产品化、断线重连/心跳、消息顺序保障、压测和客服坐席多端一致性。
- 涉及文件：
  - `services/api-go/internal/platform/postgres_store.go`
  - `infra/db/migrations/0005_chat_messages.sql`
  - `scripts/check-architecture.mjs`
  - `docs/product/user-miniprogram-backend-integration.md`
  - `docs/product/frontend-delivery-status.md`
  - `docs/product/recent-progress-roadmap.md`
  - `apps/user-wechat-miniprogram/README.md`
  - `README.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260529-176 Realtime Gateway Redis 多副本 fanout 首版

- 日期：2026-05-29
- 结果：继续按参考图逐页精修与待接后端补齐，把用户端商户群聊实时投递从单进程 WebSocket client map 推进到 Redis Pub/Sub cluster fanout 首版。outbox relay 打到任意 realtime-gateway 副本后，该副本会先投递本地连接，再把标准化 `message.sent` envelope 发布到 Redis channel，其他副本收到后按 `thread_id` 投递本机连接。
- 网关对接：`createRealtimeServer` 新增 `instanceID`、`clusterChannel` 和可注入 `clusterAdapter`；Redis envelope 带 `source_id`，本副本收到自己发布的消息会跳过，避免回环重复投递。`/readyz` 新增 cluster 状态，便于查看 adapter、channel、发布/接收计数和最近错误。
- 部署对接：Docker Compose 给 realtime-gateway 接入 `REALTIME_REDIS_URL=redis://redis:6379/0` 和 `REALTIME_REDIS_CHANNEL`；K8s base 增加 `realtime-redis-url` secret 引用和 channel 配置，保留 6 副本网关部署口径。
- 验证：
  - `npm run test --workspace @infinitech/realtime-gateway`
  - `npm run verify:architecture -- --test-name-pattern "realtime gateway supports redis"`
- 当前边界：这是 Redis Pub/Sub 多副本 fanout 首版，不是完整生产 IM 网关。WebSocket 签名 token 鉴权首版已由 `DONE-20260529-177` 承接，会话成员权限校验首版已由 `DONE-20260529-178` 承接；后续仍需动态群成员/静默策略产品化、断线重连/心跳、消息顺序保障、Redis Cluster failover 演练、连接容量压测和客户端弱网重连体验。
- 涉及文件：
  - `services/realtime-gateway/src/runtime.mjs`
  - `services/realtime-gateway/src/server.mjs`
  - `services/realtime-gateway/src/runtime.test.mjs`
  - `infra/docker/compose.yml`
  - `infra/k8s/base/app-stack.yaml`
  - `scripts/check-architecture.mjs`
  - `docs/product/user-miniprogram-backend-integration.md`
  - `docs/product/frontend-delivery-status.md`
  - `docs/product/recent-progress-roadmap.md`
  - `apps/user-wechat-miniprogram/README.md`
  - `README.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260529-177 Realtime Gateway WebSocket 签名鉴权首版

- 日期：2026-05-29
- 结果：继续按参考图逐页精修与待接后端补齐，把用户端商户群聊实时连接从“只靠 `thread_id` 订阅”推进到 WebSocket upgrade 签名 token 鉴权首版。生产可通过 `REALTIME_WS_AUTH_REQUIRED=true` 强制校验 API-Go 同款 HMAC access token；连接上的 `user_id` 不能冒充其他用户，管理员/客服角色保留代看能力。
- 网关对接：realtime-gateway 新增 `authorizeWebSocketUpgrade`、`verifySignedRealtimeToken`、dev token 开关和 user_id spoofing 拦截；token 使用 `AUTH_TOKEN_SECRET` / `REALTIME_AUTH_TOKEN_SECRET` 校验 `sub`、`role`、`exp`。未开启强制鉴权时保持本地调试兼容。
- 小程序对接：登录/注册结果会保存 `userId`；开发 token 会同步写当前用户 ID；商户群 WebSocket URL 不再硬编码 `user_1`，连接时继续带 `Authorization: Bearer <token>`。
- 部署对接：Docker Compose 增加 `REALTIME_WS_AUTH_REQUIRED` 和 `AUTH_TOKEN_SECRET`；K8s base 对 realtime-gateway 开启 `REALTIME_WS_AUTH_REQUIRED=true` 并读取 `auth-token-secret`。
- 验证：
  - `npm run test --workspace @infinitech/realtime-gateway`
  - `npm run verify:architecture -- --test-name-pattern "realtime websocket"`
- 当前边界：这是 WebSocket 签名 token 鉴权首版，不是完整 IM 权限系统。会话成员服务端校验首版已由 `DONE-20260529-178` 承接；后续仍需动态群成员/静默策略产品化、token 撤销/会话状态联动、断线重连/心跳、消息顺序保障、Redis failover 演练和容量压测。
- 涉及文件：
  - `services/realtime-gateway/src/runtime.mjs`
  - `services/realtime-gateway/src/server.mjs`
  - `services/realtime-gateway/src/runtime.test.mjs`
  - `apps/user-wechat-miniprogram/utils/api.ts`
  - `apps/user-wechat-miniprogram/pages/messages/merchant-group/index.ts`
  - `infra/docker/compose.yml`
  - `infra/k8s/base/app-stack.yaml`
  - `scripts/check-architecture.mjs`
  - `docs/product/user-miniprogram-backend-integration.md`
  - `docs/product/frontend-delivery-status.md`
  - `docs/product/recent-progress-roadmap.md`
  - `apps/user-wechat-miniprogram/README.md`
  - `README.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260529-178 用户端消息会话成员权限校验首版

- 日期：2026-05-29
- 结果：继续按参考图逐页精修与待接后端补齐，把用户端消息/商户群聊从“已签名连接”推进到会话成员服务端校验首版。用户端消息列表、聊天记录、离线同步、已读回执和发送消息都会按会话成员关系隐藏非成员访问；realtime-gateway 在 WebSocket upgrade 通过签名 token 后，会用 `thread_id`、主体类型和主体 ID 调 API-Go 内部授权口确认成员资格。
- 后端对接：`ChatThreadMember`、`ChatThreadAccessRequest` 和 `ChatThreadAccessResult` 已纳入平台契约；`Store.AuthorizeChatThreadAccess` 按 `conversation_members` 默认成员规则校验 `user`、`merchant`、`rider`、`support_admin` 与系统主体，非成员返回隐藏式 `ErrNotFound`。API-Go 新增 `POST /internal/realtime/authorize`，通过 `REALTIME_INTERNAL_TOKEN` 保护，给 realtime-gateway 返回 `allowed` 或 `not_member`。
- 数据库/部署对接：`PostgresStore.AuthorizeChatThreadAccess` 读取 `conversation_members`，默认会话同步会写入完整成员集合而不是单成员；Docker Compose 和 K8s base 已补 `REALTIME_MEMBERSHIP_AUTH_URL`，API-Go 与 realtime-gateway 共享 `realtime-internal-token`。
- 网关对接：realtime-gateway 新增 `REALTIME_MEMBERSHIP_AUTH_URL`、`createHTTPMembershipAuthorizer` 和握手期成员校验；非成员连接会返回 `403 websocket thread membership denied`，授权服务不可用时返回 `503`，避免只靠前端 `thread_id` 进入陌生会话。
- 验证：
  - `npm run test --workspace @infinitech/realtime-gateway`
  - `go test ./internal/platform -run TestChatMessageSyncMarksReadAndQueuesRealtimeOutbox`
  - `go test ./internal/httpapi -run TestChatSyncReadAndRealtimeOutboxHTTPFlow`
  - `npm run verify:architecture -- --test-name-pattern "chat thread membership"`
- 当前边界：这是静态/默认会话成员权限校验首版，不是完整 IM 成员体系。动态群成员入退群、群静默/免打扰产品化、token 撤销/会话状态联动、客户端断线重连/心跳、消息顺序保障、Redis failover 演练和 10 万在线压测仍需后续补齐。
- 涉及文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/realtime-gateway/src/runtime.mjs`
  - `services/realtime-gateway/src/server.mjs`
  - `services/realtime-gateway/src/runtime.test.mjs`
  - `infra/docker/compose.yml`
  - `infra/k8s/base/app-stack.yaml`
  - `scripts/check-architecture.mjs`
  - `docs/product/user-miniprogram-backend-integration.md`
  - `docs/product/frontend-delivery-status.md`
  - `docs/product/recent-progress-roadmap.md`
  - `apps/user-wechat-miniprogram/README.md`
  - `README.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260529-179 用户端消息会话免打扰偏好首版

- 日期：2026-05-29
- 结果：继续按参考图逐页精修与待接后端补齐，把用户端消息/商户群聊从“能看、能发、能鉴权”再推进到“能按会话切换免打扰”。API-Go 新增 `GET /api/messages/{threadID}/preference` 和 `PUT /api/messages/{threadID}/preference`，用户可读取与切换 `muted`；消息列表会把当前会话的免打扰状态直接返回给小程序，商户群页右上角可切换“已静默/消息提醒”。
- 后端对接：`ChatThreadPreference`、`UpdateChatThreadPreferenceRequest` 已纳入平台契约；`Store` 新增会话成员偏好读写，官方群默认沿用 `user:* muted=true`，商户群/客服/骑手会话可按用户保存精确 `conversation_members` 记录。`AuthorizeChatThreadAccess` 和会话列表现在会带出最新 `muted`，让 realtime-gateway 内部授权和小程序 UI 看到同一状态。
- 数据库对接：`PostgresStore` 会把 `chat_thread_members` 快照同步回 `conversation_members`，启动时再从表恢复成员静默状态；这样会话免打扰不会只留在 JSON snapshot 里。
- 小程序对接：消息中心列表新增“免打扰”标签，群聊页调用新 preference API 并支持即时切换；分段筛选也开始按群聊/私聊/通知真正过滤会话列表。
- 验证：
  - `go test ./internal/platform -run TestChatMessageSyncMarksReadAndQueuesRealtimeOutbox`
  - `go test ./internal/httpapi -run TestChatSyncReadAndRealtimeOutboxHTTPFlow`
  - `npm run verify:architecture -- --test-name-pattern "chat thread mute"`
- 当前边界：这是会话免打扰偏好首版，不是完整 IM 产品化。动态群成员、跨端统一静默策略、token 撤销/会话状态联动、客户端断线重连/心跳、消息顺序保障、Redis failover 演练和 10 万在线压测仍需后续补齐。
- 涉及文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `apps/user-wechat-miniprogram/utils/api.ts`
  - `apps/user-wechat-miniprogram/pages/messages/index.ts`
  - `apps/user-wechat-miniprogram/pages/messages/index.wxml`
  - `apps/user-wechat-miniprogram/pages/messages/index.wxss`
  - `apps/user-wechat-miniprogram/pages/messages/merchant-group/index.ts`
  - `apps/user-wechat-miniprogram/pages/messages/merchant-group/index.wxml`
  - `scripts/check-architecture.mjs`
  - `docs/product/user-miniprogram-backend-integration.md`
  - `docs/product/frontend-delivery-status.md`
  - `docs/product/recent-progress-roadmap.md`
  - `apps/user-wechat-miniprogram/README.md`
  - `README.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260529-180 用户端商户群资料与成员预览首版

- 日期：2026-05-29
- 结果：继续按参考图逐页精修与待接后端补齐，把用户端商户群聊从“能收发消息和切免打扰”再推进到“群资料、群公告、群设置和成员预览都走后端”。API-Go 新增 `GET /api/messages/{threadID}/overview` 和 `GET /api/messages/{threadID}/members`，小程序商户群页会读取群摘要、群公告、活跃成员预览和当前静默状态；新增 `pages/messages/group-settings/index` 子页承接群设置入口、免打扰开关和成员列表预览。
- 后端对接：平台契约新增 `ChatThreadOverview` 和 `ChatThreadMemberProfile`，Repository/Store 新增群资料与成员查询；商户群默认成员补入 `user_group_xiaolin`、`user_group_ajie`，并通过会话成员视图输出“蓝海餐厅 / 小林 / 阿杰 / 我”等展示数据。商户群种子消息也改成更贴参考图的商户消息与拼单对话。
- 数据库对接：新能力继续复用现有 `conversation_members` 规范化成员表，不新增旁路快照字段；`PostgresStore` 已有的成员同步/恢复能力可直接承接群资料与成员预览查询。
- 小程序对接：商户群页顶部信息条改为“群资料 + 群设置”入口，聊天区补了聊天内优惠券卡；群设置页可读取群资料、公告、成员预览并切换免打扰。这样参考图里的“群设置”“326 人已加入 · 新用户默认静音”“群公告”不再只是硬编码占位。
- 验证：
  - `go test ./internal/platform -run TestChatMessageSyncMarksReadAndQueuesRealtimeOutbox`
  - `go test ./internal/httpapi -run TestChatSyncReadAndRealtimeOutboxHTTPFlow`
  - `npm run verify:architecture -- --test-name-pattern "chat thread mute|architecture directories"`
- 当前边界：这是群资料与成员预览首版，不是完整动态群成员体系。真正的入群/退群、群主/管理员操作、跨端成员一致性、token 撤销/会话状态联动、断线重连/心跳、消息顺序保障、Redis failover 演练和 10 万在线压测仍需后续补齐。
- 涉及文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `apps/user-wechat-miniprogram/app.json`
  - `apps/user-wechat-miniprogram/utils/api.ts`
  - `apps/user-wechat-miniprogram/pages/messages/merchant-group/index.ts`
  - `apps/user-wechat-miniprogram/pages/messages/merchant-group/index.wxml`
  - `apps/user-wechat-miniprogram/pages/messages/merchant-group/index.wxss`
  - `apps/user-wechat-miniprogram/pages/messages/group-settings/index.json`
  - `apps/user-wechat-miniprogram/pages/messages/group-settings/index.ts`
  - `apps/user-wechat-miniprogram/pages/messages/group-settings/index.wxml`
  - `apps/user-wechat-miniprogram/pages/messages/group-settings/index.wxss`
  - `scripts/check-architecture.mjs`
  - `docs/product/frontend-delivery-status.md`
  - `docs/product/user-miniprogram-backend-integration.md`
  - `docs/product/recent-progress-roadmap.md`
  - `apps/user-wechat-miniprogram/README.md`
  - `README.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260529-181 用户端商户群自助入退群与群券资格首版

- 日期：2026-05-29
- 结果：继续按参考图逐页精修与待接后端补齐，把用户端商户群从“能看群资料”推进到“能自助入群、退群并按资格领券”。API-Go 新增 `GET /api/messages/{threadID}/membership`、`POST /api/messages/{threadID}/join`、`POST /api/messages/{threadID}/leave`；商户群成员数会随当前成员快照动态变化，群券 `GROUP8` 只有群成员可领取。
- 后端对接：平台契约新增 `ChatThreadMembership`、`ChatThreadJoinRequest`、`ChatThreadLeaveRequest`；`Store` / `PostgresStore` 新增自助入群/退群读写与群券资格校验，`ClaimUserCoupon` 已识别商户群券领取逻辑并保持幂等。
- 数据库对接：`conversation_members` 现在按当前成员快照整表同步，不再在恢复期盲目回填默认成员；这样用户退群后，PostgreSQL-backed Store 重启也不会把已退出成员自动加回去。
- 小程序对接：店铺详情新增“商户群福利”卡片，支持从商家页直接入群；商户群页聊天内优惠券卡可直接领取群券；群设置页支持退出商户群；红包优惠页新增商户群福利卡，支持“入群并领券”一条链路。
- 验证：
  - `go test ./internal/platform -run 'TestChatMessageSyncMarksReadAndQueuesRealtimeOutbox|TestMerchantGroupMembershipJoinLeaveAndCouponEligibility|TestUserAssetCatalogAndErrandAPIs'`
  - `go test ./internal/httpapi -run 'TestChatSyncReadAndRealtimeOutboxHTTPFlow|TestMerchantGroupMembershipAndCouponHTTPFlow'`
  - `npm run verify:architecture -- --test-name-pattern 'chat thread mute|merchant group self-serve membership'`
- 当前边界：这是商户群自助入退群与群券资格首版，不含群主/管理员角色、批量拉群、被踢出群通知、断线重连/顺序保障和生产压测。
- 涉及文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `apps/user-wechat-miniprogram/utils/api.ts`
  - `apps/user-wechat-miniprogram/pages/messages/merchant-group/index.ts`
  - `apps/user-wechat-miniprogram/pages/messages/merchant-group/index.wxml`
  - `apps/user-wechat-miniprogram/pages/messages/merchant-group/index.wxss`
  - `apps/user-wechat-miniprogram/pages/messages/group-settings/index.ts`
  - `apps/user-wechat-miniprogram/pages/messages/group-settings/index.wxml`
  - `apps/user-wechat-miniprogram/pages/messages/group-settings/index.wxss`
  - `apps/user-wechat-miniprogram/pages/shop/detail/index.ts`
  - `apps/user-wechat-miniprogram/pages/shop/detail/index.wxml`
  - `apps/user-wechat-miniprogram/pages/shop/detail/index.wxss`
  - `apps/user-wechat-miniprogram/pages/coupons/index.ts`
  - `apps/user-wechat-miniprogram/pages/coupons/index.wxml`
  - `apps/user-wechat-miniprogram/pages/coupons/index.wxss`
  - `scripts/check-architecture.mjs`
  - `docs/product/frontend-delivery-status.md`
  - `docs/product/user-miniprogram-backend-integration.md`
  - `docs/product/recent-progress-roadmap.md`
  - `apps/user-wechat-miniprogram/README.md`
  - `README.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260529-182 用户端店铺详情评价与商家信息聚合首版

- 日期：2026-05-29
- 结果：继续按参考图逐页精修与待接后端补齐，把用户端店铺详情页里仍是占位文案的“评价 / 商家”两栏推进到真实数据闭环。API-Go 新增 `GET /api/shops/{shopID}/detail`，统一返回头图、公告、活动标签、评分摘要、评价列表、营业时间、联系电话、地址、服务承诺和资质信息。
- 后端对接：平台契约新增 `ShopDetail`、`ShopReviewSummary`、`ShopReviewEntry` 和 `ShopMerchantInfo`；`Store.ShopDetail` 会把默认评价种子、动态新增评价、商户资质与员工联系电话聚合成一个公开详情结果。
- 小程序对接：店铺详情页已补真实头图、评价摘要、评价卡片、商家信息卡、资质清单、配送说明以及“联系商家 / 复制地址”操作，`评价` 和 `商家` tabs 不再停留在占位提示。
- 验证：
  - `go test ./internal/platform -run TestShopDetailAggregatesReviewsAndMerchantInfo`
  - `go test ./internal/httpapi -run TestShopDetailHTTPFlow`
- 当前边界：这是店铺详情评价与商家信息聚合首版，商品分类映射、真实评价图片上传、商家电话脱敏策略和门店多店地址体系仍待后续补齐。
- 涉及文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `apps/user-wechat-miniprogram/utils/api.ts`
  - `apps/user-wechat-miniprogram/pages/shop/detail/index.ts`
  - `apps/user-wechat-miniprogram/pages/shop/detail/index.wxml`
  - `apps/user-wechat-miniprogram/pages/shop/detail/index.wxss`
  - `docs/product/frontend-delivery-status.md`
  - `docs/product/user-miniprogram-backend-integration.md`
  - `docs/product/recent-progress-roadmap.md`
  - `apps/user-wechat-miniprogram/README.md`
  - `README.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260529-183 用户端下单链路购物车与地址回填首版

- 日期：2026-05-29
- 结果：继续按参考图逐页精修与待接后端补齐，把用户端“购物车 -> 确认订单 -> 订单详情”这条高频下单链路从半静态预览推进到真实回填首版。购物车页已接真实商品增减、清空和费用汇总；确认订单页会按当前用户地址列表和购物车摘要回填地址与费用；订单详情页可直接展示真实商家名、配送地址快照和事件时间。
- 后端对接：平台契约为 `CartSummary` 新增 `shop_name`，为 `Order` 新增 `shop_name` 和 `address_snapshot`；`Store.CheckoutCart` 会在下单时固化商家名和地址快照，`GET /api/cart` 与 `GET /api/orders/{orderID}` 都能直接给小程序渲染所需字段。
- 小程序对接：地址列表页支持从确认订单页进入选择模式，并可通过同一个地址保存接口回写默认地址；确认订单页已补默认地址回填、地址选择返回和真实购物车费用；订单详情页已补真实状态文案与时间线时间。
- 验证：
  - `go test ./internal/platform -run TestShopAddressCartCheckoutPaymentAndGrabFlow`
  - `go test ./internal/httpapi -run TestShopAddressCartCheckoutHTTPFlow`
- 当前边界：这是下单链路真实回填首版，还没有地址删除接口、订单选项编辑面板、优惠券核销联动和微信支付真机调起验收。
- 涉及文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `apps/user-wechat-miniprogram/pages/cart/index.ts`
  - `apps/user-wechat-miniprogram/pages/cart/index.wxml`
  - `apps/user-wechat-miniprogram/pages/cart/index.wxss`
  - `apps/user-wechat-miniprogram/pages/address/list/index.ts`
  - `apps/user-wechat-miniprogram/pages/address/list/index.wxml`
  - `apps/user-wechat-miniprogram/pages/address/list/index.wxss`
  - `apps/user-wechat-miniprogram/pages/order/confirm/index.ts`
  - `apps/user-wechat-miniprogram/pages/order/confirm/index.wxml`
  - `apps/user-wechat-miniprogram/pages/order/confirm/index.wxss`
  - `apps/user-wechat-miniprogram/pages/order/detail/index.ts`
  - `apps/user-wechat-miniprogram/pages/order/detail/index.wxml`
  - `apps/user-wechat-miniprogram/pages/order/detail/index.wxss`
  - `apps/user-wechat-miniprogram/README.md`
  - `docs/product/frontend-delivery-status.md`
  - `docs/product/user-miniprogram-backend-integration.md`
  - `docs/product/recent-progress-roadmap.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260529-184 用户端订单评价真实回填与更新首版

- 日期：2026-05-29
- 结果：继续按参考图逐页精修与待接后端补齐，把用户端“订单列表 -> 订单详情 -> 评价订单”这条评价链路从静态壳推进到真实回填首版。订单列表现在能识别待评价状态并直接跳评价页；订单详情能展示 `reviewed` 状态并提供去评价/修改评价入口；评价页会读取真实订单摘要、回填既有评价、支持匿名开关和星级交互，并对同一订单执行更新覆盖。
- 后端对接：平台契约为 `Review` 新增 `anonymous`，为 `Order` 新增 `reviewed`；`GET /api/reviews` 支持按 `order_id` 过滤，同一用户对同一订单再次提交会复用原评价 ID 原地更新；店铺评价聚合会对匿名评价做用户名和头像脱敏。
- 小程序对接：订单列表改用真实 `shop_name`、时间和 `reviewed` 状态渲染操作按钮；订单详情增加去评价/修改评价入口；评价页已接 `GET /api/orders/{orderID}` 与 `GET /api/reviews?order_id=`，能回填内容、标签、匿名状态并提交更新。
- 验证：
  - `go test ./internal/platform -run 'TestShopDetailAggregatesReviewsAndMerchantInfo|TestOrderReviewStateAndOrderScopedLookup'`
  - `go test ./internal/httpapi -run 'TestShopAddressCartCheckoutHTTPFlow|TestReviewHTTPFlowSupportsOrderScopedLookupAndUpdate'`
- 当前边界：图片上传位仍是占位交互，评价维度目前仍聚合成单条订单评价，尚未拆分到菜品逐项评分或骑手独立评分账本。
- 涉及文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `apps/user-wechat-miniprogram/utils/api.ts`
  - `apps/user-wechat-miniprogram/pages/order/list/index.ts`
  - `apps/user-wechat-miniprogram/pages/order/list/index.wxml`
  - `apps/user-wechat-miniprogram/pages/order/detail/index.ts`
  - `apps/user-wechat-miniprogram/pages/order/detail/index.wxml`
  - `apps/user-wechat-miniprogram/pages/order/review/index.ts`
  - `apps/user-wechat-miniprogram/pages/order/review/index.wxml`
  - `apps/user-wechat-miniprogram/pages/order/review/index.wxss`
  - `apps/user-wechat-miniprogram/README.md`
  - `docs/product/frontend-delivery-status.md`
  - `docs/product/user-miniprogram-backend-integration.md`
  - `docs/product/recent-progress-roadmap.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260529-185 用户端售后页订单上下文与进度预览首版

- 日期：2026-05-29
- 结果：继续按参考图逐页精修与待接后端补齐，把用户端“申请售后”页从静态表单推进到真实订单上下文首版。页面现在会按选中的真实订单回填商家、商品摘要、订单状态和可退金额；售后记录会按当前订单筛选，并可直接查看最近处理进度、时间线和已上传凭证预览。顺手也修掉了原先“联系客服”按钮误跳返回的问题，不再把售后单错误导向客服工单详情页。
- 后端对接：`AfterSalesRequest` 新增 `shop_name`、`order_status`、`order_item_summary`、`latest_event_message` 和 `latest_event_at`；`GET /api/after-sales` 支持按 `order_id` 过滤，便于小程序按订单拉取售后记录并展示最近用户可见进度。
- 小程序对接：售后页现在会联合读取 `GET /api/orders`、`GET /api/orders/{orderID}`、`GET /api/after-sales?order_id=`、`GET /api/after-sales/{requestID}/events` 和 `GET /api/after-sales/{requestID}/evidence`；当当前订单已有处理中售后时，提交按钮会切到“查看进度”，前端会优先展示已有申请，减少重复提交。
- 验证：
  - `go test ./internal/platform -run 'TestAfterSalesReviewApprovesAndRefundsBalance|TestUserAfterSalesRequestsExposeOrderContextAndFilter'`
  - `go test ./internal/httpapi -run 'TestAfterSalesHTTPFlow|TestReviewHTTPFlowSupportsOrderScopedLookupAndUpdate'`
- 当前边界：这一段还没有把售后图片上传入口真正放到用户手里，当前只支持预览已有凭证；对象存储上传票据和确认链路仍主要由接口和测试覆盖，前端补传会在下一段接入。
- 涉及文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `apps/user-wechat-miniprogram/utils/api.ts`
  - `apps/user-wechat-miniprogram/pages/after-sales/index.ts`
  - `apps/user-wechat-miniprogram/pages/after-sales/index.wxml`
  - `apps/user-wechat-miniprogram/pages/after-sales/index.wxss`
  - `apps/user-wechat-miniprogram/README.md`
  - `docs/product/frontend-delivery-status.md`
  - `docs/product/user-miniprogram-backend-integration.md`
  - `docs/product/recent-progress-roadmap.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260529-186 用户端售后页补充凭证上传首版

- 日期：2026-05-29
- 结果：继续按参考图逐页精修与待接后端补齐，把用户端售后页里原本还停留在“下一段接入”的凭证上传补成首版可用。现在用户在处理中售后单里可以直接补传图片凭证，凭证区会展示上传中状态、历史凭证卡和点击预览，不再只是静态加号占位。
- 后端对接：本轮未新增后端契约，但把既有 `POST /api/after-sales/{requestID}/evidence/upload-ticket`、`POST /api/after-sales/{requestID}/evidence/confirm`、`POST /api/object-storage/upload-callback` 和 `POST /api/object-storage/scan-result` 真正串到用户端；前端会先尝试直接确认，若环境要求对象存储上传回调或扫描通过，再按严格门禁补回调和扫描结果后重试确认。
- 小程序对接：售后页已新增补充凭证上传交互、上传中卡片态、严格环境兜底确认链路和凭证区文案收口；`utils/api.ts` 也补了售后凭证确认、对象存储上传回调和扫描结果提交 helper，后续评价页图片上传可以直接复用这组模式。
- 验证：
  - `npm run verify`
  - `git diff --check`
- 当前边界：小程序侧目前仍按参考图控制在最多 3 张售后凭证展示位；若生产环境使用真实对象存储签名且拒绝前端模拟回调，最终仍需依赖正式对象存储回调与扫描 worker 才能完成确认。
- 涉及文件：
  - `apps/user-wechat-miniprogram/utils/api.ts`
  - `apps/user-wechat-miniprogram/pages/after-sales/index.ts`
  - `apps/user-wechat-miniprogram/pages/after-sales/index.wxml`
  - `apps/user-wechat-miniprogram/pages/after-sales/index.wxss`
  - `apps/user-wechat-miniprogram/README.md`
  - `docs/product/frontend-delivery-status.md`
  - `docs/product/user-miniprogram-backend-integration.md`
  - `docs/product/recent-progress-roadmap.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260529-187 用户端评价页图片上传与逐项菜品评分首版

- 日期：2026-05-29
- 结果：继续按参考图逐页精修与待接后端补齐，把用户端“评价订单”页从只支持单条订单评价，推进到支持评价图片上传和逐项菜品评分的首版。现在评价页会按订单商品回填菜品行，支持逐项点星、评价图片上传/预览、匿名状态回填，并把图片和 `item_ratings` 一起提交给后端。
- 后端对接：`Review` 新增 `item_ratings`；API-Go 新增 `POST /api/reviews/upload-ticket` 与 `POST /api/reviews/upload-confirm`，复用对象存储上传票据、上传回调和扫描门禁能力。评论更新仍保持同一用户同一订单原地覆盖，但现在会一并更新 `image_urls` 和 `item_ratings`。
- 小程序对接：评价页已接 `GET /api/orders/{orderID}`、`GET /api/reviews?order_id=`、`POST /api/reviews/upload-ticket`、`POST /api/reviews/upload-confirm`，并复用对象存储上传回调/扫描结果 helper，在严格环境下先补对象存储回调，再按需要补扫描通过结果后重试确认。
- 验证：
  - `go test ./internal/platform -run 'TestOrderReviewStateAndOrderScopedLookup|TestReviewImageUploadSupportsOrderReviewAssets|TestReviewImageUploadRequiresUnifiedObjectScanApproval'`
  - `go test ./internal/httpapi -run 'TestReviewHTTPFlowSupportsOrderScopedLookupAndUpdate'`
  - `npm run verify`
  - `git diff --check`
- 当前边界：评价页现在仍把“总体评价”和“配送服务”聚合成单条 `rating` 计算后提交，尚未拆分成独立骑手评分账本；图片上传先服务于评价详情回填与商家侧展示，店铺评价聚合里的图片瀑布流仍待后续再做。
- 涉及文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `apps/user-wechat-miniprogram/utils/api.ts`
  - `apps/user-wechat-miniprogram/pages/order/review/index.ts`
  - `apps/user-wechat-miniprogram/pages/order/review/index.wxml`
  - `apps/user-wechat-miniprogram/pages/order/review/index.wxss`
  - `apps/user-wechat-miniprogram/README.md`
  - `docs/product/frontend-delivery-status.md`
  - `docs/product/user-miniprogram-backend-integration.md`
  - `docs/product/recent-progress-roadmap.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260529-188 用户端店铺详情评价图片区与配送服务评分首版

- 日期：2026-05-29
- 结果：继续按参考图逐页精修与待接后端补齐，把用户端“店铺详情 -> 评价订单 -> 再回店铺评价区”的链路补成了一条线。现在店铺详情评价卡会展示真实评价图片、菜品评分摘要和配送服务评分，评价页里的“配送服务”也不再只停留在前端，而是会独立落到后端。
- 后端对接：`Review` 新增 `rider_rating`；`CreateReview` 会独立保存配送服务评分；`ShopDetail` 评价聚合新增 `image_urls`、`item_highlights` 和 `rider_stars_text`，默认评价种子与动态订单评价都会回填这几类字段。
- 小程序对接：`pages/order/review/index` 提交时改成商家总体评分走 `rating`、配送服务走 `rider_rating`；`pages/shop/detail/index` 的评价区已补图片预览、菜品评分摘要 chip 和配送服务评分展示。
- 验证：
  - `go test ./internal/platform -run 'TestShopDetailAggregatesReviewsAndMerchantInfo|TestOrderReviewStateAndOrderScopedLookup'`
  - `go test ./internal/httpapi -run 'TestShopDetailHTTPFlow|TestReviewHTTPFlowSupportsOrderScopedLookupAndUpdate'`
  - `npm run verify`
  - `git diff --check`
- 当前边界：店铺详情评价区暂时还是轻量图片宫格，没有继续做更重的瀑布流/视频态；`rider_rating` 已经落在订单评价里，但还没继续接到骑手端绩效账本和聚合看板。
- 涉及文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `apps/user-wechat-miniprogram/pages/order/review/index.ts`
  - `apps/user-wechat-miniprogram/pages/shop/detail/index.ts`
  - `apps/user-wechat-miniprogram/pages/shop/detail/index.wxml`
  - `apps/user-wechat-miniprogram/pages/shop/detail/index.wxss`
  - `apps/user-wechat-miniprogram/README.md`
  - `docs/product/frontend-delivery-status.md`
  - `docs/product/user-miniprogram-backend-integration.md`
  - `docs/product/recent-progress-roadmap.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260529-189 骑手配送评分聚合与站长绩效展示首版

- 日期：2026-05-29
- 结果：继续按参考图逐页精修与待接后端补齐，把已经落在订单评价里的 `rider_rating` 真正接到了骑手绩效聚合里，并让站长工作台、管理端绩效视图和用户端店铺评价图区都能看见这条数据。
- 后端对接：`RiderPerformance` 新增 `rider_average_rating`、`rider_review_count`；`StationRiderPerformance` 现在会从订单评价和骑手直评里聚合配送评分样本，返回骑手平均配送分和样本数。
- 前端对接：站长工作台骑手卡已展示配送评分与样本数；Admin Web 骑手绩效视图已把原来的占位列换成配送评分；用户端店铺详情评价图区补成稳定三宫格预览，多图时最后一格会提示余量。
- 验证：
  - `go test ./internal/platform -run 'TestStationManagerManualAssignUsesStationScope|TestShopDetailAggregatesReviewsAndMerchantInfo|TestOrderReviewStateAndOrderScopedLookup'`
  - `go test ./internal/httpapi -run 'TestStationManagerManualDispatchHTTPFlow|TestShopDetailHTTPFlow|TestReviewHTTPFlowSupportsOrderScopedLookupAndUpdate'`
  - `npm run test --workspace @infinitech/admin-web`
  - `npm run verify`
  - `git diff --check`
- 当前边界：配送评分目前只进入绩效展示口径，还没有继续参与派单优先级权重计算；骑手端和管理端也还没补更细的评价明细钻取。
- 涉及文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `apps/admin-web/src/adminSnapshot.mjs`
  - `apps/admin-web/src/adminViews.mjs`
  - `apps/admin-web/src/adminDetails.mjs`
  - `apps/admin-web/src/adminApi.test.mjs`
  - `apps/rider-uni/pages/station/index.vue`
  - `apps/rider-uni/README.md`
  - `apps/admin-web/README.md`
  - `apps/user-wechat-miniprogram/pages/shop/detail/index.ts`
  - `apps/user-wechat-miniprogram/pages/shop/detail/index.wxml`
  - `apps/user-wechat-miniprogram/pages/shop/detail/index.wxss`
  - `apps/user-wechat-miniprogram/README.md`
  - `docs/product/frontend-delivery-status.md`
  - `docs/product/recent-progress-roadmap.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260529-190 骑手配送评分进入派单分与优先级首版

- 日期：2026-05-29
- 结果：继续按参考图逐页精修与待接后端补齐，把骑手配送评分从“只展示”推进到“参与派单”。`RiderPerformance.score` 现已纳入配送评分样本折算结果，站点内骑手 `dispatch_priority` 会在绩效快照和自动派单前刷新，新的配送评价会直接影响下一次自动派单候选排序。
- 后端对接：`RiderPerformance` 新增 `rider_average_rating`、`rider_review_count` 后，绩效评分会把配送评分按样本数置信度折算进总分；`dispatchDecisionLocked` 在挑选候选骑手前会刷新站点骑手优先级；新增自动派单测试覆盖“同基础指标下高配送评分骑手优先”。
- 前端对接：站长工作台骑手卡已从“配送评分 + 样本数”补成“配送评分 + 样本数 + 派单分”；Admin Web 骑手绩效视图也会在评分列里附带派单分，便于解释排序依据。
- 验证：
  - `go test ./internal/platform -run 'TestStationManagerManualAssignUsesStationScope|TestAutoAssignRefreshesPriorityFromRiderRatings|TestShopDetailAggregatesReviewsAndMerchantInfo|TestOrderReviewStateAndOrderScopedLookup'`
  - `go test ./internal/httpapi -run 'TestStationManagerManualDispatchHTTPFlow|TestShopDetailHTTPFlow|TestReviewHTTPFlowSupportsOrderScopedLookupAndUpdate'`
  - `npm run test --workspace @infinitech/admin-web`
  - `npm run verify`
  - `git diff --check`
- 当前边界：配送评分已经进入派单分和优先级，但还没有做骑手评分明细钻取、评分趋势图或更复杂的派单权重解释面板。
- 涉及文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `apps/admin-web/src/adminSnapshot.mjs`
  - `apps/admin-web/README.md`
  - `apps/rider-uni/pages/station/index.vue`
  - `apps/rider-uni/README.md`
  - `docs/product/recent-progress-roadmap.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260529-191 骑手绩效拆解与站长可解释展示首版

- 日期：2026-05-29
- 结果：继续按参考图逐页精修与待接后端补齐，把骑手绩效从“有分数”推进到“能解释分数”。`RiderPerformance` 现已返回 `score_breakdown`，拆出接单加分、单量加分、履约加分、评分加分、评分置信度和团队均值；站长工作台新增绩效拆解卡，后台骑手绩效表和详情抽屉也同步展示派单分与拆解。
- 后端对接：`evaluateRiderPerformanceLevel` 现在会返回完整拆解结构；`/api/station-manager/rider-performance` 会把真实 `score_breakdown` 返回给站长端和后台快照消费方。
- 前端对接：站长工作台新增骑手绩效拆解区；Admin Web 骑手绩效视图新增“派单分”“评分拆解”列，详情抽屉也能直接查看整行拆解事实。
- 验证：
  - `go test ./...`
  - `npm run test --workspace @infinitech/admin-web`
  - `npm run verify`
  - `git diff --check`
- 当前边界：现在已经能解释派单分来源，但还没有补骑手评分趋势图、最近评价摘录或更细的异常履约钻取。
- 涉及文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `apps/admin-web/src/adminSnapshot.mjs`
  - `apps/admin-web/src/adminViews.mjs`
  - `apps/admin-web/src/adminDetails.mjs`
  - `apps/admin-web/src/adminApi.test.mjs`
  - `apps/rider-uni/pages/station/index.vue`
  - `apps/rider-uni/README.md`
  - `apps/admin-web/README.md`
  - `docs/product/recent-progress-roadmap.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260601-192 骑手绩效趋势与异常履约钻取首版

- 日期：2026-06-01
- 结果：继续按参考图逐页精修与待接后端补齐，把骑手绩效从“能解释分数”推进到“能解释最近为什么变了”。`RiderPerformance` 现已补出 `recent_trend`、`recent_reviews` 和 `exception_summary`，站长工作台与管理端详情抽屉都能直接看到最近 3 天趋势、评价摘录和异常履约摘要。
- 后端对接：`/api/station-manager/rider-performance` 现在会返回最近 3 天绩效趋势点、最新骑手评价摘录，以及 7 天窗口内的超时、拒单、售后、低分异常摘要；绩效快照生成时会同步刷新这些派单解释数据。
- 前端对接：站长工作台绩效拆解卡新增趋势、最近评价和异常履约区；Admin Web 骑手绩效详情抽屉会把这三组事实拼进详情面板，方便运营做派单复盘和追责。
- 验证：
  - `go test ./...`
  - `npm run test --workspace @infinitech/admin-web`
  - `npm run verify`
  - `git diff --check`
- 当前边界：现在已经能看到短期趋势和异常摘要，但还没有把趋势图产品化成可交互筛选，也还没有补“异常事件 -> 订单/售后/派单审计”的逐条深链钻取。
- 涉及文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `apps/admin-web/src/adminSnapshot.mjs`
  - `apps/admin-web/src/adminDetails.mjs`
  - `apps/admin-web/src/adminApi.test.mjs`
  - `apps/admin-web/README.md`
  - `apps/rider-uni/pages/station/index.vue`
  - `apps/rider-uni/README.md`
  - `docs/product/recent-progress-roadmap.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260603-193 骑手异常履约明细深链首版

- 日期：2026-06-03
- 结果：继续按参考图逐页精修与待接后端补齐，把骑手异常履约从“摘要可见”推进到“明细可追”。`RiderPerformance` 现已补出 `exception_details`，会把派单超时、骑手拒单、售后介入和低分评价整理成可排序的异常明细，并带上订单、派单事件、售后单或评价单的关联 ID。
- 后端对接：`/api/station-manager/rider-performance` 现在会返回 `exception_details`；后台售后列表补了 `request_id / order_id / status` 过滤，方便异常明细动作真正落到后端查询；相关 Store / PostgreSQL 代理签名也已同步。
- 前端对接：站长工作台异常履约区新增异常明细列表；Admin Web 骑手绩效详情抽屉会展示逐条异常明细，并预填“查看派单事件 / 查看售后 / 查看审计”动作。
- 验证：
  - `go test ./internal/platform -run 'TestRiderPerformanceExceptionDetailsExposeDrilldownContext|TestStationManagerManualAssignUsesStationScope'`
  - `go test ./internal/httpapi -run 'TestAfterSalesHTTPFlow|TestStationManagerManualDispatchHTTPFlow|TestStationManagerRiderPerformanceExceptionDrilldownHTTPFlow'`
  - `npm run test --workspace @infinitech/admin-web`
  - `go test ./...`
  - `npm run verify`
  - `git diff --check`
- 当前边界：现在已经能从绩效详情跳到派单事件、售后和审计，但还没有把异常明细继续钻到专门的订单详情/售后详情页，也还没有做跨时间范围筛选。
- 涉及文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/platform/postgres_store_test.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `apps/admin-web/src/adminApi.mjs`
  - `apps/admin-web/src/adminSnapshot.mjs`
  - `apps/admin-web/src/adminDetails.mjs`
  - `apps/admin-web/src/adminApi.test.mjs`
  - `apps/admin-web/README.md`
  - `apps/rider-uni/pages/station/index.vue`
  - `apps/rider-uni/README.md`
  - `docs/product/recent-progress-roadmap.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260603-194 售后时间线与凭证钻取首版

- 日期：2026-06-03
- 结果：继续按参考图逐页精修与待接后端补齐，把后台售后排查从“能筛到单子”推进到“能直接回看时间线和凭证”。管理端控制台现已补出 `after-sales-events`、`after-sales-evidence` 两个只读查询操作，售后详情抽屉和骑手异常履约明细都能直接预填跳过去。
- 后端对接：这轮没有新增服务端业务接口，直接复用了现有 `GET /api/after-sales/{requestID}/events`、`GET /api/after-sales/{requestID}/evidence` 和 `GET /api/dispatch/orders/{orderID}/events`。
- 前端对接：Admin Web 新增售后时间线/凭证操作定义；售后审核详情抽屉补出“查看售后列表 / 时间线 / 凭证 / 订单派单事件”动作；骑手异常履约明细中的售后类异常会优先给出时间线和凭证动作。
- 验证：
  - `npm run test --workspace @infinitech/admin-web`
  - `npm run verify`
  - `git diff --check`
- 当前边界：现在已经能直接查售后时间线和凭证，但结果仍主要通过操作台 JSON 展示，还没有做成专门的可视化时间线/附件预览面板。
- 涉及文件：
  - `apps/admin-web/src/adminApi.mjs`
  - `apps/admin-web/src/adminDetails.mjs`
  - `apps/admin-web/src/adminSnapshot.mjs`
  - `apps/admin-web/src/adminApi.test.mjs`
  - `apps/admin-web/README.md`
  - `docs/product/recent-progress-roadmap.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260603-195 售后与派单结果可视化首版

- 日期：2026-06-03
- 结果：继续按参考图逐页精修与待接后端补齐，把管理端售后/派单只读查询从“看原始 JSON”推进到“先看结构化结果，再决定要不要钻底层返回”。接口操作台现已对 `after-sales-events`、`after-sales-evidence` 和 `dispatch-order-events` 渲染可视化预览。
- 后端对接：这轮没有新增服务端接口，继续复用已有 `GET /api/after-sales/{requestID}/events`、`GET /api/after-sales/{requestID}/evidence`、`GET /api/dispatch/orders/{orderID}/events` 返回结构。
- 前端对接：Admin Web 新增 `adminResultPreview` 结果适配层；操作台为售后时间线补出事件卡、为售后凭证补出凭证卡和图片预览、为派单事件补出事件卡与原因/候选摘要；原始 JSON 继续保留在折叠区，方便排障对照。
- 验证：
  - `npm run test --workspace @infinitech/admin-web`
  - `npm run verify`
  - `git diff --check`
- 当前边界：现在的可视化结果仍停留在接口操作台范围内，尚未把售后时间线/凭证/派单事件进一步做成独立详情页或支持时间范围筛选。
- 涉及文件：
  - `apps/admin-web/src/adminResultPreview.mjs`
  - `apps/admin-web/src/adminResultPreview.test.mjs`
  - `apps/admin-web/src/main.js`
  - `apps/admin-web/src/styles.css`
  - `apps/admin-web/README.md`
  - `docs/product/recent-progress-roadmap.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260603-196 售后聚合详情首版

- 日期：2026-06-03
- 结果：继续按参考图逐页精修与待接后端补齐，把后台售后排查从“列表/时间线/凭证/派单四处跳”推进到“一个聚合详情先把上下文看全”。后端现已新增 `GET /api/admin/after-sales/{requestID}`，前端售后抽屉、骑手异常履约明细和接口操作台都已经接上。
- 后端对接：新增 `AdminAfterSalesDetail` 聚合契约与仓储方法，一次性返回售后工单、事件时间线、凭证列表、派单事件，以及事件/凭证/派单摘要统计；HTTP 路由新增管理员只读详情口并沿用现有售后读取权限。
- 前端对接：Admin Web 新增 `after-sales-detail` 操作定义；售后详情抽屉和骑手异常履约里的售后异常优先给出“查看售后详情”；结果预览会把聚合结果先渲染成工单概览、时间线概览、凭证概览、派单概览四张卡。
- 验证：
  - `cd services/api-go && go test ./internal/platform -run 'TestAdminAfterSalesDetailAggregatesTimelineEvidenceAndDispatch|TestAfterSalesEventsEscalateToAdminReviewAndAuditTimeline'`
  - `cd services/api-go && go test ./internal/httpapi -run 'TestShopAddressCartCheckoutHTTPFlow|TestAfterSalesHTTPFlow'`
  - `npm run test --workspace @infinitech/admin-web`
  - `npm run verify`
  - `git diff --check`
- 当前边界：聚合详情已经把售后主上下文收口，但还没有把订单审计、退款流水和客服工单也一起并进同一详情面板；如果继续往前推，下一段值得做的是“售后详情深链到退款/审计/客服”的统一钻取。
- 涉及文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `apps/admin-web/src/adminApi.mjs`
  - `apps/admin-web/src/adminViews.mjs`
  - `apps/admin-web/src/adminDetails.mjs`
  - `apps/admin-web/src/adminSnapshot.mjs`
  - `apps/admin-web/src/adminResultPreview.mjs`
  - `apps/admin-web/src/adminApi.test.mjs`
  - `apps/admin-web/src/adminResultPreview.test.mjs`
  - `apps/admin-web/README.md`
  - `docs/product/recent-progress-roadmap.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260603-197 售后资金与客服深链首版

- 日期：2026-06-03
- 结果：继续按参考图逐页精修与待接后端补齐，把售后聚合详情从“只看售后自身上下文”推进到“顺着同一单把退款和客服也追清楚”。售后聚合详情现在会一并返回退款摘要、关联退款记录、关联客服工单和客服工单摘要；客服工单列表也支持按 `related_order_id` 过滤。
- 后端对接：`AdminAfterSalesDetail` 新增退款与客服工单聚合字段；`AdminServiceTickets` 新增 `related_order_id` 过滤能力并透传到 HTTP 查询口。
- 前端对接：Admin Web 售后详情抽屉新增“查看客服工单”动作；骑手异常履约里的售后异常也会带着订单号直跳客服工单列表；售后聚合详情预览继续补出“退款概览 / 客服工单”两张卡。
- 验证：
  - `cd services/api-go && go test ./internal/platform -run 'TestAdminAfterSalesDetailAggregatesTimelineEvidenceAndDispatch|TestAfterSalesEventsEscalateToAdminReviewAndAuditTimeline'`
  - `cd services/api-go && go test ./internal/httpapi -run 'TestAfterSalesHTTPFlow|TestServiceTicketAdminAndUserClosureHTTPFlow'`
  - `npm run test --workspace @infinitech/admin-web`
  - `npm run verify`
  - `git diff --check`
- 当前边界：现在已经能从售后详情追到退款和客服，但还没有把订单审计、退款审计和客服工单详情继续并到同一聚合口；下一段更值当的是把“售后 -> 退款流水 -> 审计日志 -> 客服详情”再做成统一深链。
- 涉及文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `apps/admin-web/src/adminApi.mjs`
  - `apps/admin-web/src/adminDetails.mjs`
  - `apps/admin-web/src/adminSnapshot.mjs`

### DONE-20260603-198 售后审计与工单详情深链首版

- 日期：2026-06-03
- 结果：继续按参考图逐页精修与待接后端补齐，把售后聚合详情再往“排查一屏收口”推进了一截。现在售后聚合详情会继续带回关联审计摘要和最近审计记录，管理端也补出了订单审计、退款审计和客服工单详情的直达入口。
- 后端对接：`AdminAfterSalesDetail` 新增 `related_audits` 与 `audit_summary`，会聚合订单、售后工单和客服工单的关联审计记录，并返回验证状态和最新动作。
- 前端对接：Admin Web 售后详情抽屉与骑手异常履约明细新增“查看订单审计 / 查看退款审计”；客服工作台新增 `support-ticket-detail` 操作与可视化预览，能直接看工单状态、SLA、时间线和最近处理进展。
- 验证：
  - `cd services/api-go && go test ./internal/platform -run TestAdminAfterSalesDetailAggregatesTimelineEvidenceAndDispatch`
  - `cd services/api-go && go test ./internal/httpapi -run 'TestAfterSalesHTTPFlow|TestServiceTicketAdminAndUserClosureHTTPFlow'`
  - `npm run test --workspace @infinitech/admin-web`
  - `npm run verify`
  - `git diff --check`
- 当前边界：售后聚合详情里已经能看到关联审计摘要，但还没把审计原文、退款流水明细和客服详情做成同一面板内的交互切换；下一段更值当的是把售后详情里的退款/审计/客服卡片继续做成真正可钻的联动面板。
- 涉及文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `apps/admin-web/src/adminApi.mjs`
  - `apps/admin-web/src/adminApi.test.mjs`
  - `apps/admin-web/src/adminDetails.mjs`
  - `apps/admin-web/src/adminSnapshot.mjs`
  - `apps/admin-web/src/adminResultPreview.mjs`
  - `apps/admin-web/src/adminResultPreview.test.mjs`
  - `apps/admin-web/README.md`
  - `docs/product/recent-progress-roadmap.md`
  - `apps/admin-web/src/adminResultPreview.mjs`
  - `apps/admin-web/src/adminApi.test.mjs`
  - `apps/admin-web/src/adminResultPreview.test.mjs`
  - `apps/admin-web/README.md`
  - `docs/product/recent-progress-roadmap.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260603-199 退款流水与管理员工单详情首版

- 日期：2026-06-03
- 结果：继续按参考图逐页精修与待接后端补齐，把售后深链里还偏“手工拼”的两段补成后台自己的正式能力。现在后台已经有管理员退款流水列表和管理员客服工单详情接口，订单监控、售后审核和异常履约可以直接顺着订单追退款，客服工作台也能直接走管理员详情口看工单。
- 后端对接：新增 `GET /api/admin/refunds`，支持按 `order_id / user_id / destination / status` 筛选退款流水；新增 `GET /api/admin/service-tickets/{ticketID}`，管理员可直接读取工单详情，不再复用用户查询口。
- 前端对接：Admin Web 新增 `refund-transactions` 操作；订单监控、售后审核和骑手异常履约明细新增“查看退款流水”动作；`support-ticket-detail` 改走管理员路径；退款流水补出首版结果预览卡片。
- 验证：
  - `cd services/api-go && go test ./internal/platform -run 'TestAdminRefundTransactionsFiltersByOrderAndDestination|TestServiceTicketAssignmentResolveCloseAndFollowUp|TestAdminAfterSalesDetailAggregatesTimelineEvidenceAndDispatch'`
  - `cd services/api-go && go test ./internal/httpapi -run 'TestAfterSalesHTTPFlow|TestServiceTicketAdminAndUserClosureHTTPFlow'`
  - `npm run test --workspace @infinitech/admin-web`
  - `npm run verify`
  - `git diff --check`
- 当前边界：退款流水现在已经能查、能筛、能预览，但还没有把售后聚合详情里的退款卡片和后台操作台做成面板内直跳联动；下一段更值当的是把售后详情里的退款/客服/审计卡片变成同屏切换的联动工作面。
- 涉及文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `apps/admin-web/src/adminApi.mjs`
  - `apps/admin-web/src/adminApi.test.mjs`
  - `apps/admin-web/src/adminDetails.mjs`
  - `apps/admin-web/src/adminSnapshot.mjs`
  - `apps/admin-web/src/adminViews.mjs`
  - `apps/admin-web/src/adminResultPreview.mjs`
  - `apps/admin-web/src/adminResultPreview.test.mjs`
  - `apps/admin-web/README.md`
  - `docs/product/recent-progress-roadmap.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260603-200 售后预览卡同屏联动首版

- 日期：2026-06-03
- 结果：继续按参考图逐页精修与待接后端补齐，把后台售后深链里最后一段“看完摘要还要手动填参数”的阻力先拿掉。现在售后聚合详情和退款流水预览卡已经能直接把操作台预填到时间线、凭证、派单事件、退款流水、客服工单详情和审计检索，后台排查开始更像一个同屏工作面。
- 后端对接：本轮没有新增接口，继续复用既有 `after-sales-detail`、`after-sales-events`、`after-sales-evidence`、`dispatch-order-events`、`refund-transactions`、`support-ticket-detail` 和 `audit-logs`。
- 前端对接：`adminResultPreview` 为售后聚合详情与退款流水卡补出预填动作；`main.js` 新增预览卡按钮事件绑定，可直接切换当前操作并自动带参；`styles.css` 补了预览卡动作区样式。
- 验证：
  - `node --test apps/admin-web/src/adminResultPreview.test.mjs`
  - `node --check apps/admin-web/src/main.js`
  - `npm run test --workspace @infinitech/admin-web`
  - `npm run verify`
  - `git diff --check`
- 当前边界：现在已经能从摘要直接跳到操作台，但还没有把这些联动结果收成真正的多面板并排视图；下一段更值当的是把退款/审计/客服详情继续做成可并排对照的工作区。
- 涉及文件：
  - `apps/admin-web/src/adminResultPreview.mjs`
  - `apps/admin-web/src/adminResultPreview.test.mjs`
  - `apps/admin-web/src/main.js`
  - `apps/admin-web/src/styles.css`
  - `apps/admin-web/README.md`
  - `docs/product/recent-progress-roadmap.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260603-201 订单聚合详情首版

- 日期：2026-06-03
- 结果：继续按参考图逐页精修与待接后端补齐，把后台“订单其实才是总入口”这条线补成首版。现在管理端已经有管理员订单聚合详情接口，订单监控、售后详情和退款/客服预览都能顺着同一个订单把售后、退款、工单、派单和审计一起拉回来看。
- 后端对接：新增 `GET /api/admin/orders/{orderID}`，返回订单主信息、售后记录摘要、退款摘要、客服工单摘要、派单摘要和关联审计摘要；聚合体里同时带回对应明细列表，方便后台继续深钻。
- 前端对接：Admin Web 新增 `order-detail` 操作；订单监控详情抽屉和售后详情抽屉新增“查看订单总览”；结果预览新增订单聚合详情卡组，并把售后/退款/客服预览顺手接回订单总入口。
- 验证：
  - `cd services/api-go && go test ./internal/platform -run 'TestAdminAfterSalesDetailAggregatesTimelineEvidenceAndDispatch|TestAdminOrderDetailAggregatesAfterSalesRefundsSupportDispatchAndAudits'`
  - `cd services/api-go && go test ./internal/httpapi -run 'TestAfterSalesHTTPFlow|TestAdminOrderDetailHTTPFlow'`
  - `node --test apps/admin-web/src/adminResultPreview.test.mjs apps/admin-web/src/adminApi.test.mjs`
  - `npm run test --workspace @infinitech/admin-web`
  - `npm run verify`
  - `git diff --check`
- 当前边界：订单聚合详情现在已经能做后台排查总入口，但还没有把退款、工单、审计做成真正并排对照的多面板工作区；下一段更值当的是继续把订单总览下的三个深链结果收成同屏协作面。
- 涉及文件：
  - `services/api-go/internal/platform/contracts.go`
  - `services/api-go/internal/platform/repository.go`
  - `services/api-go/internal/platform/store.go`
  - `services/api-go/internal/platform/postgres_store.go`
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `apps/admin-web/src/adminApi.mjs`
  - `apps/admin-web/src/adminViews.mjs`
  - `apps/admin-web/src/adminDetails.mjs`
  - `apps/admin-web/src/adminResultPreview.mjs`
  - `apps/admin-web/src/adminApi.test.mjs`
  - `apps/admin-web/src/adminResultPreview.test.mjs`
  - `apps/admin-web/README.md`
  - `docs/product/recent-progress-roadmap.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260603-202 订单总览联动结果区首版

- 日期：2026-06-03
- 结果：继续按参考图逐页精修与待接后端补齐，把后台“总览点进去还要反复抄参数”的最后一层阻力再削掉一截。现在 Admin Web 的预览卡只读动作会直接拉结果，并把退款、工单、派单、审计这些深链结果保留到同屏联动面板里，后台排查开始真正支持并排对照。
- 前端对接：新增 `adminLinkedWorkspace` helper，统一处理联动结果去重、保留最近 3 个面板和只读动作直拉；`main.js` 补出联动结果区、关闭/提主结果动作和卡片内继续联动；`styles.css` 补齐多面板工作区样式；`admin-web` README 和路线图同步记录。
- 测试：新增 `apps/admin-web/src/adminLinkedWorkspace.test.mjs`，覆盖只读动作直拉判断、请求 URL 去重和最近结果上限；并回归 `adminResultPreview`、`adminApi` 现有测试。
- 验证：
  - `node --check apps/admin-web/src/main.js`
  - `node --check apps/admin-web/src/adminLinkedWorkspace.mjs`
  - `node --test apps/admin-web/src/adminLinkedWorkspace.test.mjs apps/admin-web/src/adminResultPreview.test.mjs apps/admin-web/src/adminApi.test.mjs`
- 当前边界：现在已经能把联动结果并排挂起来，但还没有把退款/客服/审计的多面板数据再做成共享筛选或同步定位；下一段更值当的是把“同一订单上下文”继续做成更强的联动工作区，而不是只停在并排展示。
- 涉及文件：
  - `apps/admin-web/src/adminLinkedWorkspace.mjs`
  - `apps/admin-web/src/adminLinkedWorkspace.test.mjs`
  - `apps/admin-web/src/main.js`
  - `apps/admin-web/src/styles.css`
  - `apps/admin-web/README.md`
  - `docs/product/recent-progress-roadmap.md`
  - `EXECUTION_LEDGER.md`

### DONE-20260603-203 同单工作区首版

- 日期：2026-06-03
- 结果：继续按参考图逐页精修与待接后端补齐，把后台“有了订单上下文但下一步还得自己想挂哪些面板”的那层心智负担再往下压一截。现在订单聚合详情、售后聚合详情和工单详情这类主结果，会直接长出“同单工作区”工具条，能一键展开订单工作区或售后工作区，也能单独挂退款、工单、派单、审计、时间线、凭证这些常用面板。
- 前端对接：`adminLinkedWorkspace` 新增上下文提取、工作区动作和默认 bundle；`main.js` 新增同单工作区工具条、一键挂板和多面板批量拉取；`styles.css` 补齐工具条样式；README、路线图同步记录。
- 测试：扩充 `apps/admin-web/src/adminLinkedWorkspace.test.mjs`，覆盖订单/售后聚合结果的上下文提取、同单工作区动作和默认 bundle。
- 验证：
  - `node --check apps/admin-web/src/main.js`
  - `node --test apps/admin-web/src/adminLinkedWorkspace.test.mjs apps/admin-web/src/adminResultPreview.test.mjs apps/admin-web/src/adminApi.test.mjs`
- 当前边界：现在已经能基于上下文一键挂起常用面板，但还没有做到多面板共享筛选、同步定位和联动高亮；下一段更值当的是把“同一订单”的退款/工单/审计做成真正互相对齐的工作区，而不是只停在一起展示。
- 涉及文件：
  - `apps/admin-web/src/adminLinkedWorkspace.mjs`
  - `apps/admin-web/src/adminLinkedWorkspace.test.mjs`
  - `apps/admin-web/src/main.js`
  - `apps/admin-web/src/styles.css`
  - `apps/admin-web/README.md`
  - `docs/product/recent-progress-roadmap.md`
  - `EXECUTION_LEDGER.md`

## 进行中

### TASK-USER-MP-001 用户端原生微信小程序

- 状态：进行中
- 目标：完成首版用户端原生微信小程序。
- 已完成：首页、圈子、找饭搭、商家列表、店铺详情、团购套餐购买入口、购物车、确认订单、订单列表、订单详情、余额支付密码、地址预览页；01-35 已按参考图首轮精修；小程序 API 客户端和微信登录、真实微信 `wx.login` 页面流程、手机号验证码/密码登录注册/生产短信 provider 首版、店铺/商品/团购/购物车/结算/订单查询/评价/售后/反馈/客服工单分派关闭回访/客服消息敏感信息风控/圈子动态/消息会话/聊天发送/消息离线补偿/已读回执/消息 WebSocket 实时投递/消息 PostgreSQL 规范化首版/realtime gateway Redis adapter 多副本 fanout 首版/realtime gateway WebSocket 签名鉴权首版/realtime gateway 会话成员权限校验首版/realtime gateway 会话免打扰偏好首版/realtime gateway 群资料与成员预览首版/饭搭资料人工审核候选举报拉黑举报处置/找饭搭同校同楼隐私和设备风控/红包发送详情领取退回过期退回/红包领取风控/钱包账单/钱包总览/提现申请/优惠券/会员积分/邀请好友/搜索/买药首页/处方影像上传确认/处方影像对象扫描门禁/处方 OCR 留档/处方药师复核/处方审核/药品订单确认详情/药品库存锁定/跑腿下单详情/支付密码/微信支付预下单接入点；默认访问 BFF。
- 下一步：做微信开发者工具真机视觉走查，继续补售后凭证上传入口、评价图片上传/逐项菜品评价、订单选项编辑、地址删除/编辑、优惠券核销联动，以及群主/管理员角色、被踢出群/成员变更通知、断线重连、消息顺序保障和压测生产化。
- 范围：登录、首页、后台可控卡片、圈子/小微墙、找饭搭、分类、商家、搜索、今日推荐、商品详情、购物车、地址、订单备注、餐具数量、下单、微信支付、余额支付密码、订单、售后、评价、收藏、积分会员、钱包充值/提现/账单、消息、官方群/商户群、红包、客服、邀请、买药、快递/跑腿、公益。
- 验收：
  - 微信开发者工具可运行。
  - 真机可跑通首单闭环。
  - UI 使用 `#009bf5` 和旧版 logo。
  - 找饭搭未满足性别、真实性承诺、免责承诺、问卷和人工审核前不可使用。

### TASK-BACKEND-001 核心业务 API

- 状态：进行中
- 目标：建立 Go 核心业务 API。
- 已完成：微信登录签名 token、真实微信 `code2session` provider resolver、auth session 持久化与 logout 撤销、生产默认关闭开发 token、商户邀约/资质、商户员工健康证、商户补充资料、商户主体登录、管理员 bootstrap 登录、骑手/站长邀约注册、骑手/站长主体登录、商户资料、商户订单列表、商户接单/出餐状态机、商户商品管理、商户/骑手保证金、店铺接单门槛、团购下单发券与扫码核销、骑手在线状态、10 分钟后自动派单、拒单顺延派单、派单确认超时自动转派、订单状态机补偿首版、管理端运营快照聚合首版、管理端操作审计日志首版、审计日志 PostgreSQL `audit_logs` 规范化表首版、管理端审计服务端安全边界首版、管理端审计完整性证明首版、管理端服务端 RBAC 策略矩阵首版、平台 outbox 事件首版、outbox relay worker 首版、outbox relay 可运行化与部署骨架、outbox 积压观测首版、outbox 手动恢复/重放首版、outbox 批量恢复/重放首版、outbox 死信隔离首版、outbox relay 租约领取首版、outbox relay 租约续租首版、PostgreSQL outbox 规范化 relay 路径首版、outbox 租约健康观测首版、消费端幂等落库首版、支付/钱包 PostgreSQL 规范化恢复首版、订单创建 PostgreSQL 事务化首版、购物车结算 PostgreSQL 事务化首版、余额支付 PostgreSQL 事务扣减首版、退款策略与余额退款核心闭环首版、payment-worker 原路退款事件规范化首版、退款 PostgreSQL 事务化首版、售后申请与审核核心闭环首版、售后 PostgreSQL 规范化恢复首版、售后审核 PostgreSQL 事务化首版、售后部分退款资金账本首版、售后仲裁与客服介入处理日志首版、售后可退金额与证据上传票据首版、售后证据确认与附件元数据首版、对象存储上传签名配置化首版、售后上传票据账本与确认防伪首版、售后对象存在性 HEAD 校验开关首版、售后上传回调验签与扫描门禁首版、对象扫描 worker 首版、对象扫描 worker ClamAV 适配与下载首版、对象生命周期清理 worker 首版、对象清理失败账本首版、对象清理统计接口首版、派单审计事件 PostgreSQL 规范化恢复首版、商家订单流转 PostgreSQL 事务化首版、骑手取货/送达完成、固定单量完成后免责拒派决策、站长站点骑手/订单视图、站长手动派单、站长任务时长/固定单量配置、站点骑手绩效等级快照、派单事件持久化与查询审计、站点区域匹配首版、店铺、商品、地址、购物车、结算订单、订单列表、订单详情、余额支付、余额支付密码、微信支付预下单/回调验签、骑手抢单、每日一次免责取消的领域实现和 HTTP 接口；用户/商户/骑手/站长/管理员/安全审计员及后台分权角色鉴权骨架；核心 PostgreSQL 迁移；Repository 边界；PostgreSQL-backed Store 快照持久化第一阶段。
- 补充进展：`DONE-20260523-091` 已完成退款策略配置与 `admin.refund_settings.updated` 审计同事务首版；`DONE-20260523-092` 已完成管理端订单退款与 `admin.order.refunded` 审计同事务首版；`DONE-20260523-093` 已完成售后审核与 `after_sales.reviewed` 审计同事务首版；`DONE-20260523-094` 已完成订单状态补偿与 `admin.order_state.compensated` 审计同事务首版；`DONE-20260524-095` 已完成 outbox 运维 claim/lease renew/publish/fail/replay/batch replay 与审计同事务首版；`DONE-20260524-096` 已完成商户/骑手/站长邀约与审计同事务首版；`DONE-20260524-097` 已完成管理端服务端 RBAC 策略矩阵首版；`DONE-20260524-098` 已完成 RBAC 权限治理查询与变更申请审计首版；`DONE-20260524-099` 已完成 RBAC 权限申请审批/驳回台账首版；`DONE-20260524-100` 已完成 RBAC 权限变更手动应用首版；`DONE-20260524-101` 已完成 RBAC 权限变更审计回滚首版；`DONE-20260524-102` 已完成管理端审计导出首版；`DONE-20260524-103` 已完成管理端审计留存/告警健康报告首版；`DONE-20260524-104` 已完成管理端审计留存告警 outbox 投递首版；`DONE-20260524-105` 已完成管理端审计 WORM/冷归档请求首版；`DONE-20260524-106` 已完成管理端审计归档 worker 首版；`DONE-20260524-107` 已完成审计归档完成回写与记录查询首版；`DONE-20260524-108` 已完成审计归档对象下载校验与回查首版；`DONE-20260524-109` 已完成审计归档校验历史查询首版；`DONE-20260524-110` 已完成审计归档校验历史可视化面板首版；`DONE-20260524-111` 已完成管理端 P0 业务详情面板首版；`DONE-20260524-112` 已完成管理端高风险操作二次确认与结果追踪首版；`DONE-20260524-113` 已完成管理端失败回放入口首版；`DONE-20260524-114` 已完成管理端 P0 业务筛选分页首版；`DONE-20260525-115` 已完成管理端售后审核表单首版；`DONE-20260525-116` 已完成管理端订单退款表单首版；`DONE-20260525-117` 已完成管理端 Outbox 单事件恢复表单首版；`DONE-20260525-118` 已完成管理端 Outbox 发布/失败人工处置表单首版；`DONE-20260525-119` 已完成管理端 Outbox 领取/续租表单首版；`DONE-20260525-120` 已完成管理端 Outbox 死信分诊/解封表单首版；`DONE-20260525-121` 已完成商户资质审核后端与管理端表单首版；`DONE-20260525-122` 已完成 Outbox 单事件事故辅助明细首版；`DONE-20260525-123` 已完成商户资质待审列表与明细接口首版；`DONE-20260525-124` 已完成商户资质审核结果可靠通知首版；`DONE-20260525-125` 已完成商户站内通知中心首版；`DONE-20260525-126` 已完成通知运营查询接口首版；`DONE-20260525-127` 已完成通知投递回执台账首版；`DONE-20260525-128` 已完成 Admin Web 通知运营页首版；`DONE-20260525-129` 已完成通知失败回执告警首版。已迁移的审计同事务路径在 PostgreSQL-backed Store 下均在同一数据库事务内写入业务表、outbox 表、资质表或邀约快照与 `audit_logs`。
- 下一步：继续扫描后台配置、运营处置、资金和风控写路径，把剩余关键写操作迁移到业务写入与审计写入同事务强制提交；补字段级/租户级权限、策略版本回滚和菜单隐藏策略；补生产 WORM bucket 保留策略、归档校验生产演练回查、保留期删除审批；补真实 Kafka/NATS broker 运维和 relay 积压恢复；继续推进微信原路退款 API 调用、售后对象存储真实签名/回调和提现/结算资金链路。
- 范围：认证、用户、商户、骑手、店铺、商品、订单、团购、买药、跑腿、钱包、支付、消息、客服、RTC、邀请、配置。
- 验收：
  - 单元测试和 API contract 测试通过。
  - 订单、支付、钱包关键链路具备幂等和并发测试。

### TASK-20260521-002 生产化基础设施细化

- 状态：进行中
- 目标：把当前架构骨架推进到可部署的自建/混合云生产底座。
- 范围：镜像构建、K8s 配置拆分、Ingress、Secrets/Vault、PostgreSQL HA、Redis Cluster、Kafka 多 Broker、MinIO、观测与告警。
- 验收：
  - 本地 Docker 依赖可启动。
  - K8s base 可通过 dry-run 校验。
  - 生产 secrets 不进入仓库。
  - 观测指标和告警规则有基线。
- 证据：
  - Docker 启动日志。
  - K8s dry-run 输出。
  - 基础健康检查结果。
  - Grafana/Prometheus 告警截图或导出。

## 端侧与后续任务

### TASK-MERCHANT-FLUTTER-001 商户端 Flutter

- 状态：进行中
- 目标：完成首版商户端。
- 已完成：后端邀约注册、商户 token、商户邀请注册带密码、商户账号密码登录 API 工具、匿名注册/登录不带开发 token、资质上传、资质缺失检查、员工健康证、补充资料、商户保证金查询/缴纳、商户订单列表、接单、出餐、商品列表、商品创建/编辑、商品状态调整、团购券扫码核销；商户端经营概况页面、订单处理页面、商品管理页面、团购核销页面、资质资料页面和 API 客户端。
- 下一步：补商户端登录/邀请注册页面、店铺装修、商户钱包、消息页和资质过期强弹窗。
- 范围：管理员邀请链接注册、登录、经营、订单、商品图片/描述/配料表、店铺详情页、团购二维码验券、商户自发券、平台活动券确认参与、钱包、消息、商户群、红包、营业执照/健康证有效期、员工信息、补充资料、资质过期弹窗、售后沟通、评价查看。
- 验收：
  - Flutter 可运行。
  - 商户能接单、拒单、出餐、联系用户/骑手。
  - 资质过期或未审核通过时店铺暂时关闭并提示补资料。
  - 团购券二维码扫码验券可用。
  - 商户自发券和商户承担活动券能进入结算审计。

### TASK-RIDER-FLUTTER-001 骑手端 Flutter

- 状态：进行中
- 目标：完成首版骑手端。
- 已完成：抢单大厅首屏、骑手上线/离线、保证金缴纳/微信免押/退押申请、取货、送达完成、每日一次免责取消、拒绝当前派单、站长工作台首屏、站点骑手列表、骑手等级/接单耗时、待调度订单列表、站长手动派单、每日任务时长和固定单量配置、站长创建骑手邀约、骑手接受邀约带密码和骑手/站长账号密码登录的 API 调用入口。
- 范围：站长/骑手邀约注册、分权登录、在线/离线、抢单、10 分钟后自动派单、拒绝顺延下一位在线骑手、等级优先派单、站长手动派单、站长每日任务时长配置、每日固定订单数后免责不接、任务、钱包、充值/提现/账单、收入、数据统计、接单设置、健康证、保险、违规申诉、骑手之家、历史订单、保证金/微信免押、退押金、客服、红包。
- 验收：
  - 骑手可抢单和接派单。
  - 在线派单每日一次免责取消规则生效。
  - 完成后台配置固定订单数后，后续派单免责不接规则生效。
  - 后台可查看取消审计。
  - 后台可查看骑手接单耗时、团队均值、等级和派单优先级。
  - 站长可查看站点全部数据并手动派单。
  - 未缴纳保证金且未通过微信免押的骑手不可接单。

### TASK-ADMIN-WEB-001 管理端 Web

- 状态：进行中
- 目标：完成桌面管理端。
- 已完成：最小运营控制台首版，包含登录操作台、商户/站长/骑手邀约、运营快照、操作审计、退款策略、售后列表、客服工单列表、对象存储清理、outbox 运维、订单状态补偿、P0 运营指标位、今日待办、模块状态和 RBAC 首版矩阵；已补订单监控、售后审核、商户资质、骑手/站长、骑手绩效、派单审计、审计检索、退款策略、客服工作台、通知运营和权限治理等业务视图首版；P0 KPI、队列和表格已可由运营快照生成，并对展示字段做 HTML 转义；业务表格行已新增详情面板首版，可展示字段、核查清单并跳到补偿、订单退款、商户资质审核、售后审核、客服分派、工单方案、审计、outbox、对象清理、通知回执和 RBAC 等下一步操作；通知运营模块已可查询通知台账、投递回执并补录高风险回执；客服工作台已可查询服务工单、预填客服分派和处理方案并进入高风险二次确认；商户模块和详情抽屉已新增资质审核表单首版，可预填 `merchant_id`、`qualification_id`、`decision`、`reason` 和 `reviewed_at` 并进入高风险二次确认；订单模块和详情抽屉已新增退款表单首版，可预填 `order_id`、`reason`、`idempotency_key`、`amount_fen` 和 `destination` 并进入高风险二次确认；售后模块和详情抽屉已新增审核表单首版，可预填 `request_id`、`decision`、`reason`、`refund_destination` 和 `refund_idempotency_key` 并进入高风险二次确认；运营首页和 Outbox 队列详情抽屉已新增事件明细、死信分诊、领取租约、续租、死信解封、单事件恢复、标记失败和标记已发布表单首版，可预填 `status=dead_letter`、`event_id`、失败原因、重试延迟、最大尝试次数和租约参数，高风险写动作进入二次确认；后端 Outbox 单事件明细已返回事件状态、payload 摘要、关联目标、最近审计、推荐操作和处置核查清单；审计中心已支持 actor/action/target/after/before/limit 筛选、before 游标翻页、保存筛选、详情抽屉、跨模块跳转、脱敏 payload 摘要、完整性状态展示、CSV 导出、留存告警报告、告警 outbox 投递、WORM/冷归档请求和归档校验历史可视化面板；后端已支持 `security_auditor` 只读审计角色、后台分权 RBAC 策略矩阵、RBAC 策略查询、权限变更申请审计、权限申请审批/驳回台账、权限变更手动应用、权限变更审计回滚、审计 CSV 导出、审计留存/告警健康报告、审计留存告警 outbox 投递、审计 WORM/冷归档请求、审计归档 worker、归档完成记录查询、归档对象校验、归档校验历史查询、审计 payload 服务端白名单/敏感字段掩码和审计完整性证明首版。
- 补充进展：退款策略保存、管理端订单退款、售后审核、商户资质审核、订单状态补偿、对象清理完成/失败、outbox 运维和商户/骑手邀约后端已改为同事务写入业务结果与审计；管理端服务端 RBAC 策略矩阵首版已落地，Admin Web 审计检索和权限治理页可继续查看完整性状态、服务端规范化 payload、权限申请审计、审批/驳回审计、应用审计、回滚审计、审计 CSV 导出审计、留存告警健康报告、告警 outbox 投递结果、归档请求结果和归档校验历史可视化面板。
- 下一步：把 P0 详情面板继续接更多后端业务明细接口、更完整的审核辅助信息、字段级/租户级 RBAC、剩余后台配置/运营处置/资金风控写路径同事务审计、生产 WORM bucket 保留策略、归档校验生产演练回查、保留期删除审批、真实告警渠道投递、KMS/链式不可抵赖签名和首页卡片/优惠券/圈子饭搭配置页。
- 范围：订单、售后、用户、商户、骑手绩效等级、商品、精选商品、首页卡片、首页活动、优惠券、圈子/饭搭、团购、买药、跑腿、支付中心、钱包、提现、退款策略、积分会员、通知推送、评价、风控、数据备份恢复、派单规则、骑手计价、群聊红包、客服、RTC、电话联系审计、OAuth/API 管理、系统日志。
- 验收：
  - 桌面浏览器可完成运营配置和订单处理。
  - 权限边界清楚。
  - 首页卡片、圈子开关、饭搭协议/问卷、优惠券资金责任可配置可审计。

### TASK-ADMIN-FLUTTER-001 移动管理端 Flutter

- 状态：未开始
- 目标：完成手机运营高频处理端。
- 范围：订单监控、调度、客服、退款审核、告警处理。
- 验收：
  - Flutter 可运行。
  - 移动端可处理紧急运营任务。

### TASK-BFF-001 多端 BFF

- 状态：进行中
- 目标：建立面向用户小程序、商户端、骑手端、管理端的聚合层。
- 已完成：运行时配置、首页模块/卡片、用户端微信登录、商户邀约与主体登录、管理员 bootstrap 登录、骑手/站长邀约与主体登录、店铺/商品/地址/购物车/结算/订单/钱包/微信支付预下单、退款策略/订单退款、商户资质审核、商户资质待审列表/明细、商户订单/商品/保证金、骑手调度/保证金、派单事件查询、派单确认超时转派、订单状态机补偿、管理端运营快照代理、管理端操作审计代理、管理端审计导出代理、管理端审计留存/告警健康报告代理、管理端审计留存告警 outbox 投递代理、管理端 RBAC 策略查询、权限变更申请、申请列表、审批/驳回、手动应用和审计回滚代理、对象存储清理候选/统计/完成/失败代理、站长任务/绩效核心接口代理、浏览器来源 CORS 白名单与预检处理首版。
- 范围：运行时配置、聚合接口、灰度、端侧兼容、错误标准化。
- 验收：
  - 各端不直接依赖内部领域 API。
  - BFF 有基础缓存、限流和观测。

### TASK-REALTIME-001 IM 与实时通知

- 状态：进行中
- 目标：建立实时网关。
- 已完成：`message.sent` outbox 到 realtime-gateway 的 WebSocket 首版投递，Redis Pub/Sub 多副本 fanout 首版，WebSocket 签名 token 鉴权首版，会话成员服务端校验首版，小程序商户群页已接同会话实时消息接收。
- 范围：消息、订单通知、骑手位置、抢单/派单事件、官方客服。
- 验收：
  - 用户/商户/骑手/客服消息实时送达。
  - 消息落库、已读、离线补偿可用。

### TASK-RTC-001 网络语音通话

- 状态：未开始
- 目标：支持平台内网络语音通话。
- 范围：WebRTC 信令、呼叫状态、通话审计、订单/会话关联。
- 验收：
  - 呼叫、接听、拒绝、取消、挂断、超时全流程可用。
  - RTC 信令压测达标。

### TASK-DISPATCH-001 抢单加派单

- 状态：进行中
- 目标：建立骑手调度系统。
- 已完成：并发抢单单成功约束、10 分钟后自动派单、拒单顺延、派单确认超时自动转派、站长手动派单、每日一次免责取消、固定单量后免责拒派决策、站点区域匹配首版、派单审计事件持久化与查询。
- 范围：抢单池、派单策略、在线状态、位置、候选评分、超时转派、免责取消。
- 验收：
  - 并发抢同一订单只能成功一次。
  - 派单超时可自动转派。
  - 每日一次免责取消按配置生效。

### TASK-PRICING-001 骑手计价规则

- 状态：未开始
- 目标：后台可配置骑手订单计价。
- 范围：基础费、距离、重量、时段、天气、高峰、跨区、等待、补贴、抽佣、结算。
- 验收：
  - 订单保存计价版本和明细。
  - 骑手端和后台看到一致明细。

### TASK-PAYMENT-001 微信支付与余额支付

- 状态：进行中
- 目标：支付链路可生产化。
- 已完成：微信预支付/回调验签骨架、余额支付密码、余额支付幂等、支付/钱包 PostgreSQL 规范化同步与恢复首版、余额支付 PostgreSQL 事务扣减首版、退款策略与余额退款核心闭环首版、payment-worker 原路退款事件规范化首版、退款 PostgreSQL 事务化首版、售后部分退款资金账本首版、售后仲裁与客服介入处理日志首版、售后可退金额与证据上传票据首版、售后证据确认与附件元数据首版、对象存储上传签名配置化首版、售后上传票据账本与确认防伪首版、售后对象存在性 HEAD 校验开关首版、售后上传回调验签与扫描门禁首版、对象扫描 worker 首版、对象扫描 worker ClamAV 适配与下载首版、对象生命周期清理 worker 首版、对象清理失败账本首版、对象清理统计接口首版。
- 范围：微信预支付、回调、退款、对账、余额扣减、钱包流水、充值跳转。
- 验收：
  - 回调幂等。
  - 钱包并发扣减安全。
  - 充值 URL 跳转可配置。

### TASK-INTEGRATION-001 OAuth 和第三方 API

- 状态：未开始
- 目标：支持外部平台 OAuth 和 API 数据调用。
- 范围：provider registry、授权、回调、凭据、同步任务、审计、限流、熔断。
- 验收：
  - 微信 provider 可用。
  - 新 provider 可通过配置接入。

### TASK-INVITE-001 邀请用户页

- 状态：未开始
- 目标：完成邀请页和邀请数据闭环。
- 范围：邀请链接、落地页、接受邀请、奖励、邀请统计、商户/骑手入驻邀请扩展。
- 验收：
  - 用户可分享邀请。
  - 被邀请用户可注册并绑定邀请关系。
  - 后台可查看数据。

### TASK-CAPACITY-001 10 万同时在线压测

- 状态：未开始
- 目标：验证平台支撑 10 万同时在线。
- 范围：10k、30k、60k、100k 四档压测，IM、订单、骑手位置、抢单、RTC 信令混合流量。
- 验收：
  - 四档压测报告归档。
  - 指标达到 `PLATFORM_MASTER_PLAN.md` 的容量阈值。
  - 不能仅凭架构设计宣称完成。

### TASK-DR-001 容灾演练

- 状态：未开始
- 目标：建立并验证容灾能力。
- 范围：API/BFF/实时网关节点故障、Redis failover、数据库只读/主从切换、队列积压恢复、支付回调重放。
- 验收：
  - RPO/RTO 结果有证据。
  - 发布回滚演练通过。

### DONE-20260603-204 同单工作区同步刷新

- 时间：2026-06-03
- 范围：`apps/admin-web/src/adminLinkedWorkspace.mjs`、`apps/admin-web/src/adminLinkedWorkspace.test.mjs`、`apps/admin-web/src/main.js`、`apps/admin-web/src/styles.css`、`apps/admin-web/README.md`、`docs/product/recent-progress-roadmap.md`
- 内容：
  - Admin Web 同单工作区新增共享上下文同步首版，联动卡片会标记“已对齐 / 待同步”，并显示同步后的订单/售后/工单/用户目标 token。
  - 联动结果区新增“同步当前筛选 / 同步全部”和单卡“同步当前上下文”，主结果切换后，退款、客服工单、派单事件、审计等已挂载卡片可以直接批量刷新到当前上下文，不必手填参数重挂。
  - 同步逻辑统一收口到 `adminLinkedWorkspace.mjs`，对 `order-detail`、`refund-transactions`、`support-tickets`、`support-ticket-detail`、`dispatch-order-events`、`after-sales-*`、`audit-logs` 等操作做共享参数重绑。
- 验证：
  - `node --check apps/admin-web/src/main.js`
  - `node --test apps/admin-web/src/adminLinkedWorkspace.test.mjs apps/admin-web/src/adminResultPreview.test.mjs apps/admin-web/src/adminApi.test.mjs`
  - `npm run test --workspace @infinitech/admin-web`
  - `npm run verify`
  - `git diff --check`

### DONE-20260603-205 同单工作区上下文切换

- 时间：2026-06-03
- 范围：`apps/admin-web/src/adminLinkedWorkspace.mjs`、`apps/admin-web/src/adminLinkedWorkspace.test.mjs`、`apps/admin-web/src/main.js`、`apps/admin-web/src/styles.css`、`apps/admin-web/README.md`、`docs/product/recent-progress-roadmap.md`
- 内容：
  - Admin Web 同单工作区新增“切到此上下文”和候选上下文区，运营不必再先把某张联动卡提升为主结果，就能直接围绕该卡的订单/售后/工单/用户上下文继续排查。
  - 工作区切换后会保留主结果，并补出“跟随主结果”回退入口；焦点定位会自动跟到新上下文，让“筛选 + 同步当前筛选/全部 + 带参刷新”继续围绕同一组 token 运作。
  - `adminLinkedWorkspace.mjs` 新增上下文归一化、上下文候选聚合、上下文相等判断和主焦点推导；`adminLinkedWorkspace.test.mjs` 补了对应断言。
- 验证：
  - `node --check apps/admin-web/src/main.js`
  - `node --test apps/admin-web/src/adminLinkedWorkspace.test.mjs apps/admin-web/src/adminResultPreview.test.mjs apps/admin-web/src/adminApi.test.mjs`
  - `npm run test --workspace @infinitech/admin-web`
  - `npm run verify`
  - `git diff --check`

### DONE-20260603-206 同单工作区分组同步与表单预填

- 时间：2026-06-03
- 范围：`apps/admin-web/src/adminLinkedWorkspace.mjs`、`apps/admin-web/src/adminLinkedWorkspace.test.mjs`、`apps/admin-web/src/main.js`、`apps/admin-web/src/styles.css`、`apps/admin-web/README.md`、`docs/product/recent-progress-roadmap.md`
- 内容：
  - Admin Web 同单工作区新增按退款、客服、派单、审计分组的同步按钮，运营可以只刷新当前关注的那组联动面板，不必总是整批重跑。
  - 操作台当前操作会跟随工作区上下文自动预填 `order_id / related_order_id / request_id / ticket_id / user_id / target_type / target_id` 等参数；切换上下文、切换操作和回到“跟随主结果”时都会同步更新。
  - 顺手修掉了联动结果卡分组 badge 的显示问题，退款/客服/派单/审计卡现在会显示正确分组，不再一律落成“其他”。
- 验证：
  - `node --check apps/admin-web/src/main.js`
  - `node --test apps/admin-web/src/adminLinkedWorkspace.test.mjs apps/admin-web/src/adminResultPreview.test.mjs apps/admin-web/src/adminApi.test.mjs`
  - `npm run test --workspace @infinitech/admin-web`
  - `npm run verify`
  - `git diff --check`

### DONE-20260603-207 同单工作区上下文自动刷新

- 时间：2026-06-03
- 范围：`apps/admin-web/src/adminLinkedWorkspace.mjs`、`apps/admin-web/src/adminLinkedWorkspace.test.mjs`、`apps/admin-web/src/main.js`、`apps/admin-web/README.md`、`docs/product/recent-progress-roadmap.md`
- 内容：
  - Admin Web 同单工作区新增上下文自动刷新：切到联动卡上下文、切候选上下文、回到“跟随主结果”，以及把联动卡提升为主结果后，当前筛选下待同步的联动卡会自动刷新，不再需要再手动补一轮“同步当前筛选”。
  - 焦点态下的同步匹配口径也一起修正，除了识别联动卡当前的订单/售后/工单/用户 token，还会识别同步后的目标 token，避免刚切完上下文时因为焦点已切到新 token 而漏掉最该刷新的卡。
  - `main.js` 补了异步上下文切换与自动刷新串联；`adminLinkedWorkspace.mjs` 抽出 `linkedWorkspaceSyncEntries` / `linkedWorkspaceSyncActions`，把分组同步、焦点同步和目标 token 判断统一收口。
- 验证：
  - `node --check apps/admin-web/src/main.js`
  - `node --test apps/admin-web/src/adminLinkedWorkspace.test.mjs apps/admin-web/src/adminResultPreview.test.mjs apps/admin-web/src/adminApi.test.mjs`
  - `npm run test --workspace @infinitech/admin-web`
  - `npm run verify`
  - `git diff --check`

### DONE-20260603-208 同单工作区同步反馈

- 时间：2026-06-03
- 范围：`apps/admin-web/src/adminLinkedWorkspace.mjs`、`apps/admin-web/src/adminLinkedWorkspace.test.mjs`、`apps/admin-web/src/main.js`、`apps/admin-web/src/styles.css`、`apps/admin-web/README.md`、`docs/product/recent-progress-roadmap.md`
- 内容：
  - Admin Web 同单工作区新增同步反馈层：工具条会显示“同步中 / 同步反馈”提示，并把退款、客服、派单、审计四组的待同步/已对齐状态汇总展示出来。
  - 当前筛选、全部同步和分组同步按钮在刷新期间会进入对应“刷新中”态；刷新完成后会直接回写“已刷新当前筛选/全部/某分组 N 张卡”，减少运营对同步结果的猜测。
  - `adminLinkedWorkspace.mjs` 新增分组汇总能力 `linkedWorkspaceSyncOverview`；`main.js` 接入同步活动状态、分组状态摘要和按钮 loading 文案；`styles.css` 补齐同步反馈区样式。
- 验证：
  - `node --check apps/admin-web/src/main.js`
  - `node --test apps/admin-web/src/adminLinkedWorkspace.test.mjs apps/admin-web/src/adminResultPreview.test.mjs apps/admin-web/src/adminApi.test.mjs`
  - `npm run test --workspace @infinitech/admin-web`
  - `npm run verify`
  - `git diff --check`

### DONE-20260603-209 同单工作区同步时间与分组完成感

- 时间：2026-06-03
- 范围：`apps/admin-web/src/adminLinkedWorkspace.mjs`、`apps/admin-web/src/adminLinkedWorkspace.test.mjs`、`apps/admin-web/src/main.js`、`apps/admin-web/src/styles.css`、`apps/admin-web/README.md`、`docs/product/recent-progress-roadmap.md`
- 内容：
  - Admin Web 同单工作区同步反馈继续补强：现在除了“同步中 / 已刷新”，还会带上最近一次完成时间和涉及分组摘要，方便运营确认刚刚刷新的是哪几条线。
  - 分组状态摘要会在同步完成后显示“刚刷新 N”，退款/客服/派单/审计几组不再只剩统一的“已对齐”，而是能把刚刚发生的动作保留一层可读痕迹。
  - `adminLinkedWorkspace.mjs` 新增同步动作分组统计 `linkedWorkspaceSyncActionGroups`；`main.js` 接入完成时间、分组摘要和“刚刷新”态；`styles.css` 补齐状态文案排布。
- 验证：
  - `node --check apps/admin-web/src/main.js`
  - `node --test apps/admin-web/src/adminLinkedWorkspace.test.mjs apps/admin-web/src/adminResultPreview.test.mjs apps/admin-web/src/adminApi.test.mjs`
  - `npm run test --workspace @infinitech/admin-web`
  - `npm run verify`
  - `git diff --check`

### DONE-20260603-210 同单工作区局部失败反馈

- 时间：2026-06-03
- 范围：`apps/admin-web/src/main.js`、`apps/admin-web/README.md`、`docs/product/recent-progress-roadmap.md`、`EXECUTION_LEDGER.md`
- 内容：
  - Admin Web 同单工作区同步反馈新增局部失败感知：如果某一轮刷新只成功了一部分卡，状态文案会写明“其中失败 N 张”，不再把局部失败伪装成完全成功。
  - 分组状态摘要也会同步标出失败分组和失败张数，退款/客服/派单/审计里哪一条线掉队，工作区会直接用红色 badge 提醒。
  - 复用 `linkedWorkspaceSyncActionGroups` 对成功/失败动作按组计数，`main.js` 把失败数、失败分组和完成时间一起回写到同步状态里。
- 验证：
  - `node --check apps/admin-web/src/main.js`
  - `node --test apps/admin-web/src/adminLinkedWorkspace.test.mjs apps/admin-web/src/adminResultPreview.test.mjs apps/admin-web/src/adminApi.test.mjs`
  - `npm run test --workspace @infinitech/admin-web`
  - `npm run verify`
  - `git diff --check`

### DONE-20260603-211 同单工作区失败项重试

- 时间：2026-06-03
- 范围：`apps/admin-web/src/adminLinkedWorkspace.mjs`、`apps/admin-web/src/adminLinkedWorkspace.test.mjs`、`apps/admin-web/src/main.js`、`apps/admin-web/src/styles.css`、`apps/admin-web/README.md`、`docs/product/recent-progress-roadmap.md`
- 内容：
  - Admin Web 同单工作区新增失败项重试入口：工作区会直接给出“重试失败项”以及按退款/客服分组的失败重试按钮，局部失败后不必手动筛卡重跑。
  - 重试链路会沿用当前工作区上下文、当前筛选和当前焦点，把失败卡重新组装成同一批 GET 动作重跑；成功/失败结果继续回写到现有同步反馈区。
  - `adminLinkedWorkspace.mjs` 新增失败项识别与分组统计 helper；`main.js` 接入失败项重试入口、失败项重试状态和重试反馈；`styles.css` 补齐失败重试按钮样式。
- 验证：
  - `node --check apps/admin-web/src/main.js`
  - `node --test apps/admin-web/src/adminLinkedWorkspace.test.mjs apps/admin-web/src/adminResultPreview.test.mjs apps/admin-web/src/adminApi.test.mjs`
  - `npm run test --workspace @infinitech/admin-web`
  - `npm run verify`
  - `git diff --check`

### DONE-20260603-212 同单工作区失败定位与单卡重试

- 时间：2026-06-03
- 范围：`apps/admin-web/src/adminLinkedWorkspace.mjs`、`apps/admin-web/src/adminLinkedWorkspace.test.mjs`、`apps/admin-web/src/main.js`、`apps/admin-web/src/styles.css`、`apps/admin-web/README.md`、`docs/product/recent-progress-roadmap.md`、`EXECUTION_LEDGER.md`
- 内容：
  - Admin Web 联动结果区新增“失败”筛选，失败卡会直接显示失败原因，先让运营知道是超时、网关错误还是接口失败，再决定怎么补救。
  - 失败卡片本身补了“重试此卡”，工作区工具条补了“仅看失败项”；现在既可以集中看红卡，也可以逐张把掉队卡补回来。
  - `adminLinkedWorkspace.mjs` 新增失败原因摘要和失败筛选口径；`main.js` 接入失败原因展示、单卡重试和失败筛选捷径；`styles.css` 补了失败提示样式。
- 验证：
  - `node --check apps/admin-web/src/main.js`
  - `node --test apps/admin-web/src/adminLinkedWorkspace.test.mjs`
  - `git diff --check -- apps/admin-web/src/adminLinkedWorkspace.mjs apps/admin-web/src/adminLinkedWorkspace.test.mjs apps/admin-web/src/main.js apps/admin-web/src/styles.css`
  - `npm run test --workspace @infinitech/admin-web`
  - `npm run verify`
  - `git diff --check`

### DONE-20260603-213 同单工作区失败卡请求事实

- 时间：2026-06-03
- 范围：`apps/admin-web/src/adminLinkedWorkspace.mjs`、`apps/admin-web/src/adminLinkedWorkspace.test.mjs`、`apps/admin-web/src/main.js`、`apps/admin-web/src/styles.css`、`apps/admin-web/README.md`、`docs/product/recent-progress-roadmap.md`、`EXECUTION_LEDGER.md`
- 内容：
  - Admin Web 失败卡补出结构化失败事实：现在红卡会直接显示 HTTP 状态、最后一次请求路径和原参数摘要，定位失败时不必先去翻原始 JSON。
  - 失败卡同时会展示“按当前工作区上下文重试后会改写成什么参数”，让运营在补跑前先确认上下文切换是否符合预期。
  - `adminLinkedWorkspace.mjs` 新增失败事实与重试参数摘要 helper；`main.js` 接入失败事实面板；`styles.css` 补齐失败事实样式；`adminLinkedWorkspace.test.mjs` 补覆盖。
- 验证：
  - `node --check apps/admin-web/src/main.js`
  - `node --test apps/admin-web/src/adminLinkedWorkspace.test.mjs`
  - `git diff --check -- apps/admin-web/src/adminLinkedWorkspace.mjs apps/admin-web/src/adminLinkedWorkspace.test.mjs apps/admin-web/src/main.js apps/admin-web/src/styles.css`
  - `npm run test --workspace @infinitech/admin-web`
  - `npm run verify`
  - `git diff --check`

### DONE-20260603-214 同单工作区失败卡回填与对照痕迹

- 时间：2026-06-03
- 范围：`apps/admin-web/src/adminLinkedWorkspace.mjs`、`apps/admin-web/src/adminLinkedWorkspace.test.mjs`、`apps/admin-web/src/main.js`、`apps/admin-web/README.md`、`docs/product/recent-progress-roadmap.md`、`EXECUTION_LEDGER.md`
- 内容：
  - Admin Web 失败卡补了“回填到操作台”，失败后可以直接把当前工作区上下文下的重试参数带回操作台，不必手工抄参数。
  - 联动结果 entry 开始保留最近一次失败和最近一次成功的痕迹；红卡会把两次状态并排给出来，方便判断故障是持续的还是回归造成的。
  - `adminLinkedWorkspace.mjs` 新增失败/成功痕迹合并与读取逻辑；`main.js` 接入卡片回填和痕迹展示；`adminLinkedWorkspace.test.mjs` 补了痕迹保留断言。
- 验证：
  - `node --check apps/admin-web/src/main.js`
  - `node --test apps/admin-web/src/adminLinkedWorkspace.test.mjs`
  - `git diff --check -- apps/admin-web/src/adminLinkedWorkspace.mjs apps/admin-web/src/adminLinkedWorkspace.test.mjs apps/admin-web/src/main.js`
  - `npm run test --workspace @infinitech/admin-web`
  - `npm run verify`
  - `git diff --check`

### DONE-20260603-215 同单工作区回填确认与局部历史

- 时间：2026-06-03
- 范围：`apps/admin-web/src/adminLinkedWorkspace.mjs`、`apps/admin-web/src/adminLinkedWorkspace.test.mjs`、`apps/admin-web/src/main.js`、`apps/admin-web/src/styles.css`、`apps/admin-web/README.md`、`docs/product/recent-progress-roadmap.md`、`EXECUTION_LEDGER.md`
- 内容：
  - Admin Web 失败卡点“回填到操作台”后，操作台会显示专门的回填确认面板，直接给出来源、当前参数、失败请求事实，以及“执行当前查询 / 收起提示”两个下一步。
  - 红卡里的最近失败/最近成功痕迹改成了局部历史列表，带时间戳、状态 badge 和参数摘要，判断是持续失败还是刚刚恢复会更直观。
  - `adminLinkedWorkspace.mjs` 新增预填 action helper；`main.js` 接入回填确认面板和局部历史列表；`styles.css` 补齐对应样式；`adminLinkedWorkspace.test.mjs` 补了预填 action 断言。
- 验证：
  - `node --check apps/admin-web/src/main.js`
  - `node --test apps/admin-web/src/adminLinkedWorkspace.test.mjs`
  - `git diff --check -- apps/admin-web/src/adminLinkedWorkspace.mjs apps/admin-web/src/adminLinkedWorkspace.test.mjs apps/admin-web/src/main.js apps/admin-web/src/styles.css`
  - `npm run test --workspace @infinitech/admin-web`
  - `npm run verify`
  - `git diff --check`
