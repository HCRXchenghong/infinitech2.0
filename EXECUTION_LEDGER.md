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
  - `apps/merchant-uni/README.md`

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
  - `apps/rider-uni/README.md`

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
  - `apps/merchant-uni/README.md`
  - `apps/rider-uni/README.md`
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
  - `apps/rider-uni/README.md`
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
- 结果：已把外卖/买药类店铺订单从“支付成功直接进入骑手调度”调整为“商户待接单 -> 备餐中 -> 出餐后进入骑手调度”；骑手在商户出餐前不能抢单；商户订单列表、接单、出餐接口已接入 BFF；商户端 uni-app 已补经营概况和订单处理首版页面。
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
  - `apps/merchant-uni/pages.json`
  - `apps/merchant-uni/utils/api.js`
  - `apps/merchant-uni/pages/index/index.vue`
  - `apps/merchant-uni/pages/orders/index.vue`

### DONE-20260521-020 商户商品管理闭环推进

- 日期：2026-05-21
- 结果：已补商户商品列表、创建/编辑、上架/售罄/下架接口；商户只能管理自己店铺商品；用户端公开商品列表只展示上架商品，售罄/下架会从公开列表隐藏；BFF 已代理商户商品接口；商户端 uni-app 已新增商品管理页。
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
  - `apps/merchant-uni/pages.json`
  - `apps/merchant-uni/utils/api.js`
  - `apps/merchant-uni/pages/index/index.vue`
  - `apps/merchant-uni/pages/products/index.vue`

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
- 结果：已补店铺团购套餐公开列表、团购下单、支付成功后自动发券、用户团购券列表、商户扫码核销团购券；核销必须使用 `qr_scan`，必须由本店商户执行，重复核销会被拒绝；BFF 已代理团购相关接口；用户小程序店铺页接入团购套餐和购买入口；商户端 uni-app 已新增团购核销页。
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
  - `apps/merchant-uni/pages.json`
  - `apps/merchant-uni/utils/api.js`
  - `apps/merchant-uni/pages/index/index.vue`
  - `apps/merchant-uni/pages/groupbuy/index.vue`

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
- 结果：已补站长查看站点骑手、站长查看待调度订单和站长手动派单接口；手动派单要求同站点在线骑手且保证金/免押状态可接单，跨站点骑手对站长不可见不可派；BFF 已代理站长调度接口；骑手端 uni-app 已新增抢单大厅首屏和站长工作台首屏。
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
  - `apps/rider-uni/pages.json`
  - `apps/rider-uni/utils/api.js`
  - `apps/rider-uni/pages/index/index.vue`
  - `apps/rider-uni/pages/station/index.vue`

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
  - `apps/rider-uni/utils/api.js`
  - `apps/rider-uni/pages/station/index.vue`

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
  - `apps/rider-uni/utils/api.js`
  - `apps/rider-uni/pages/station/index.vue`

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
  - `apps/rider-uni/pages/index/index.vue`

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
  - `apps/rider-uni/utils/api.js`
  - `apps/rider-uni/pages/index/index.vue`

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
  - `apps/rider-uni/utils/api.js`
  - `apps/rider-uni/pages/index/index.vue`

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
  - `apps/rider-uni/utils/api.js`
  - `apps/rider-uni/README.md`

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
- 结果：已把骑手/站长邀约注册升级为设置账号密码，密码使用 bcrypt 哈希保存；新增 `POST /api/auth/rider/login`，骑手和站长可用 `account_id + password` 登录并签发 session-backed token；错误密码返回 `INVALID_CREDENTIALS` 401；PostgreSQL snapshot 已持久化 rider password hash，重启后主体登录不丢；BFF 已代理骑手登录；骑手端 uni-app API 工具已支持接受邀约带密码、账号密码登录，并按账号类型保存 rider/station manager token。
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
  - `apps/rider-uni/utils/api.js`
  - `services/api-go/go.mod`
  - `services/api-go/go.sum`

### DONE-20260522-035 商户/管理员主体登录闭环

- 日期：2026-05-22
- 结果：已把商户邀约注册升级为必须设置账号密码，密码使用 bcrypt 哈希保存；新增 `POST /api/auth/merchant/login`，商户可用 `account_id + password` 登录并签发 session-backed merchant token，错误密码返回 `INVALID_CREDENTIALS` 401；PostgreSQL snapshot 已持久化 merchant password hash，重启后商户主体登录不丢；新增 `POST /api/auth/admin/login`，管理员 bootstrap 密码登录默认关闭、无内置默认密码，仅同时配置 `ADMIN_BOOTSTRAP_ACCOUNT_ID` 和 `ADMIN_BOOTSTRAP_PASSWORD` 且密码长度 8-72 字节时启用；BFF 已代理商户登录和管理员登录；商户端 uni-app API 工具已支持邀请注册带密码、账号密码登录，并在匿名注册/登录时不携带开发 token。
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
  - `apps/merchant-uni/utils/api.js`
  - `apps/merchant-uni/README.md`
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
- 结果：已把商户员工健康证和补充资料从后端模型推进到商户端首批闭环：Store/HTTP 测试覆盖员工资料读取、保存、健康证失效日期校验、跨商户店铺写入隐藏，以及补充资料读取/提交和跨商户写入隐藏；BFF 已代理 `merchant/me`、资质、员工、补充资料路径并覆盖 Authorization 转发；商户端 uni-app 新增“资质资料”页面，可查看接单门槛、资质缺失、保证金状态、员工健康证和补充资料，并提交营业执照、健康证、员工信息和门头/后厨等资料。
- 验收证据：
  - `cd services/api-go && go test ./internal/platform ./internal/httpapi`
  - `npm run test --workspace @infinitech/bff`
  - `npm run verify:architecture`
- 文件：
  - `services/api-go/internal/platform/store_test.go`
  - `services/api-go/internal/httpapi/router_test.go`
  - `services/bff/src/runtime.test.mjs`
  - `apps/merchant-uni/utils/api.js`
  - `apps/merchant-uni/pages.json`
  - `apps/merchant-uni/pages/index/index.vue`
  - `apps/merchant-uni/pages/compliance/index.vue`
  - `apps/merchant-uni/README.md`
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

## 进行中

### TASK-USER-MP-001 用户端原生微信小程序

- 状态：进行中
- 目标：完成首版用户端原生微信小程序。
- 已完成：首页、圈子、找饭搭、商家列表、店铺详情、团购套餐购买入口、购物车、确认订单、订单列表、订单详情、余额支付密码、地址预览页；小程序 API 客户端和微信登录、店铺/商品/团购/购物车/结算/订单查询/支付密码/微信支付预下单接入点；默认访问 BFF。
- 下一步：接入定位、消息、邀请页、售后、评价和真实微信 `wx.login` 页面流程。
- 范围：登录、首页、后台可控卡片、圈子/小微墙、找饭搭、分类、商家、搜索、今日推荐、商品详情、购物车、地址、订单备注、餐具数量、下单、微信支付、余额支付密码、订单、售后、评价、收藏、积分会员、钱包充值/提现/账单、消息、官方群/商户群、红包、客服、邀请、买药、快递/跑腿、公益。
- 验收：
  - 微信开发者工具可运行。
  - 真机可跑通首单闭环。
  - UI 使用 `#009bf5` 和旧版 logo。
  - 找饭搭未满足性别、真实性承诺、免责承诺和问卷前不可使用。

### TASK-BACKEND-001 核心业务 API

- 状态：进行中
- 目标：建立 Go 核心业务 API。
- 已完成：微信登录签名 token、真实微信 `code2session` provider resolver、auth session 持久化与 logout 撤销、生产默认关闭开发 token、商户邀约/资质、商户员工健康证、商户补充资料、商户主体登录、管理员 bootstrap 登录、骑手/站长邀约注册、骑手/站长主体登录、商户资料、商户订单列表、商户接单/出餐状态机、商户商品管理、商户/骑手保证金、店铺接单门槛、团购下单发券与扫码核销、骑手在线状态、10 分钟后自动派单、拒单顺延派单、派单确认超时自动转派、订单状态机补偿首版、平台 outbox 事件首版、outbox relay worker 首版、outbox relay 可运行化与部署骨架、outbox 积压观测首版、outbox 手动恢复/重放首版、outbox 批量恢复/重放首版、outbox 死信隔离首版、outbox relay 租约领取首版、outbox relay 租约续租首版、PostgreSQL outbox 规范化 relay 路径首版、outbox 租约健康观测首版、消费端幂等落库首版、支付/钱包 PostgreSQL 规范化恢复首版、订单创建 PostgreSQL 事务化首版、购物车结算 PostgreSQL 事务化首版、余额支付 PostgreSQL 事务扣减首版、退款策略与余额退款核心闭环首版、payment-worker 原路退款事件规范化首版、退款 PostgreSQL 事务化首版、售后申请与审核核心闭环首版、售后 PostgreSQL 规范化恢复首版、售后审核 PostgreSQL 事务化首版、售后部分退款资金账本首版、售后仲裁与客服介入处理日志首版、售后可退金额与证据上传票据首版、售后证据确认与附件元数据首版、对象存储上传签名配置化首版、售后上传票据账本与确认防伪首版、售后对象存在性 HEAD 校验开关首版、售后上传回调验签与扫描门禁首版、对象扫描 worker 首版、对象扫描 worker ClamAV 适配与下载首版、对象生命周期清理 worker 首版、对象清理失败账本首版、对象清理统计接口首版、派单审计事件 PostgreSQL 规范化恢复首版、商家订单流转 PostgreSQL 事务化首版、骑手取货/送达完成、固定单量完成后免责拒派决策、站长站点骑手/订单视图、站长手动派单、站长任务时长/固定单量配置、站点骑手绩效等级快照、派单事件持久化与查询审计、站点区域匹配首版、店铺、商品、地址、购物车、结算订单、订单列表、订单详情、余额支付、余额支付密码、微信支付预下单/回调验签、骑手抢单、每日一次免责取消的领域实现和 HTTP 接口；用户/商户/骑手/站长/管理员角色鉴权骨架；核心 PostgreSQL 迁移；Repository 边界；PostgreSQL-backed Store 快照持久化第一阶段。
- 下一步：补真实 Kafka/NATS broker 运维和 relay 积压恢复；继续推进微信原路退款 API 调用、售后对象存储真实签名/回调和提现/结算资金链路。
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

### TASK-MERCHANT-UNI-001 商户端 uni-app

- 状态：进行中
- 目标：完成首版商户端。
- 已完成：后端邀约注册、商户 token、商户邀请注册带密码、商户账号密码登录 API 工具、匿名注册/登录不带开发 token、资质上传、资质缺失检查、员工健康证、补充资料、商户保证金查询/缴纳、商户订单列表、接单、出餐、商品列表、商品创建/编辑、商品状态调整、团购券扫码核销；商户端经营概况页面、订单处理页面、商品管理页面、团购核销页面、资质资料页面和 API 客户端。
- 下一步：补商户端登录/邀请注册页面、店铺装修、商户钱包、消息页和资质过期强弹窗。
- 范围：管理员邀请链接注册、登录、经营、订单、商品图片/描述/配料表、店铺详情页、团购二维码验券、商户自发券、平台活动券确认参与、钱包、消息、商户群、红包、营业执照/健康证有效期、员工信息、补充资料、资质过期弹窗、售后沟通、评价查看。
- 验收：
  - HBuilderX/uni-app 可运行。
  - 商户能接单、拒单、出餐、联系用户/骑手。
  - 资质过期或未审核通过时店铺暂时关闭并提示补资料。
  - 团购券二维码扫码验券可用。
  - 商户自发券和商户承担活动券能进入结算审计。

### TASK-RIDER-UNI-001 骑手端 uni-app

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

- 状态：未开始
- 目标：完成桌面管理端。
- 范围：订单、售后、用户、商户、骑手绩效等级、商品、精选商品、首页卡片、首页活动、优惠券、圈子/饭搭、团购、买药、跑腿、支付中心、钱包、提现、退款策略、积分会员、通知推送、评价、风控、数据备份恢复、派单规则、骑手计价、群聊红包、客服、RTC、电话联系审计、OAuth/API 管理、系统日志。
- 验收：
  - 桌面浏览器可完成运营配置和订单处理。
  - 权限边界清楚。
  - 首页卡片、圈子开关、饭搭协议/问卷、优惠券资金责任可配置可审计。

### TASK-ADMIN-UNI-001 移动管理端 uni-app

- 状态：未开始
- 目标：完成手机运营高频处理端。
- 范围：订单监控、调度、客服、退款审核、告警处理。
- 验收：
  - HBuilderX/uni-app 可运行。
  - 移动端可处理紧急运营任务。

### TASK-BFF-001 多端 BFF

- 状态：进行中
- 目标：建立面向用户小程序、商户端、骑手端、管理端的聚合层。
- 已完成：运行时配置、首页模块/卡片、用户端微信登录、商户邀约与主体登录、管理员 bootstrap 登录、骑手/站长邀约与主体登录、店铺/商品/地址/购物车/结算/订单/钱包/微信支付预下单、退款策略/订单退款、商户订单/商品/保证金、骑手调度/保证金、派单事件查询、派单确认超时转派、订单状态机补偿、对象存储清理候选/统计/完成/失败代理、站长任务/绩效核心接口代理。
- 范围：运行时配置、聚合接口、灰度、端侧兼容、错误标准化。
- 验收：
  - 各端不直接依赖内部领域 API。
  - BFF 有基础缓存、限流和观测。

### TASK-REALTIME-001 IM 与实时通知

- 状态：未开始
- 目标：建立实时网关。
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
