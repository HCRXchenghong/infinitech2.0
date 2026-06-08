export const ADMIN_WEB_VIEWS = Object.freeze({
  dashboard: {
    key: "dashboard",
    title: "运营首页",
    subtitle: "P0 风险、待办队列、资金和异步任务聚合。",
    metrics: [
      { label: "售后首响", value: "30m", tone: "red" },
      { label: "资质复核", value: "19", tone: "amber" },
      { label: "调度阻塞", value: "0", tone: "green" },
      { label: "队列积压", value: "0", tone: "blue" }
    ],
    actions: ["operations-snapshot", "audit-logs", "after-sales-list", "merchant-qualifications", "merchant-qualification-detail", "refund-settings-read", "outbox-stats", "outbox-events", "outbox-event-detail", "outbox-dead-letter-triage", "outbox-claim-events", "outbox-renew-lease", "outbox-release-dead-letter", "outbox-replay-event", "outbox-mark-failed", "outbox-mark-published", "object-cleanup-stats"],
    columns: ["队列", "归属", "SLA", "处理入口"],
    rows: [
      ["售后审核", "客服", "30 分钟首响", "售后列表"],
      ["商户资质", "招商", "当天复核", "资质台"],
      ["骑手调度", "配送", "实时处理", "派单台"],
      ["Outbox", "运维", "5 分钟内恢复", "事件台"]
    ],
    safeguards: ["只读总览默认安全", "P0 队列优先显示", "资金和订单动作保留审计"]
  },
  orders: {
    key: "orders",
    title: "订单监控",
    subtitle: "外卖、团购、买药、跑腿订单统一监控。",
    metrics: [
      { label: "待商家接单", value: "24", tone: "amber" },
      { label: "待派单", value: "17", tone: "red" },
      { label: "配送中", value: "312", tone: "blue" },
      { label: "异常单", value: "6", tone: "red" }
    ],
    actions: ["operations-snapshot", "order-detail", "order-refund", "refund-transactions", "order-compensate", "audit-logs", "outbox-events", "outbox-stats"],
    columns: ["订单", "类型", "状态", "商户", "骑手", "风险"],
    rows: [
      ["ord_10031", "外卖", "待派单", "蓝湾轻食", "未分配", "超 8 分钟"],
      ["ord_10032", "团购", "已发券", "星火烤肉", "到店核销", "无"],
      ["ord_10033", "跑腿", "配送中", "平台服务", "rider_71", "天气补贴"],
      ["ord_10034", "买药", "商户备货", "康宁药房", "待取货", "资质敏感"]
    ],
    safeguards: ["订单状态补偿必须写审计", "退款不直接改余额", "异常订单进入客服复核"]
  },
  "after-sales": {
    key: "after-sales",
    title: "售后审核",
    subtitle: "退款、证据、仲裁和客服介入统一处理。",
    metrics: [
      { label: "待审核", value: "37", tone: "red" },
      { label: "证据待扫", value: "5", tone: "amber" },
      { label: "仲裁中", value: "9", tone: "blue" },
      { label: "超时风险", value: "3", tone: "red" }
    ],
    actions: ["operations-snapshot", "after-sales-list", "after-sales-detail", "after-sales-review", "refund-transactions", "audit-logs", "object-cleanup-candidates", "object-cleanup-stats"],
    columns: ["工单", "订单", "申请人", "状态", "可退金额", "证据"],
    rows: [
      ["asr_231", "ord_10031", "user_18", "商户待处理", "3200 分", "2 个附件"],
      ["asr_232", "ord_10028", "user_44", "平台仲裁", "1800 分", "扫描通过"],
      ["asr_233", "ord_10011", "user_09", "等待补证", "900 分", "未确认"],
      ["asr_234", "ord_10009", "merchant_2", "客服介入", "5200 分", "1 个附件"]
    ],
    safeguards: ["部分退款累计不得超额", "证据确认需票据和扫描通过", "内部备注不对用户展示"]
  },
  merchants: {
    key: "merchants",
    title: "商户资质",
    subtitle: "邀请准入、保证金、营业执照、健康证和员工资料。",
    metrics: [
      { label: "待入驻", value: "12", tone: "blue" },
      { label: "待审核", value: "21", tone: "amber" },
      { label: "资质过期", value: "7", tone: "red" },
      { label: "未缴保证金", value: "4", tone: "red" }
    ],
    actions: ["operations-snapshot", "merchant-qualifications", "merchant-qualification-detail", "merchant-invite", "merchant-qualification-review", "audit-logs"],
    columns: ["商户", "店铺", "能力", "保证金", "资质", "到期"],
    rows: [
      ["merchant_12", "蓝湾轻食", "外卖/团购", "已缴", "营业执照/健康证", "2026-11-30"],
      ["merchant_19", "康宁药房", "买药/外卖", "已缴", "药房资质待复核", "2026-07-12"],
      ["merchant_25", "星火烤肉", "团购", "未缴", "健康证缺失", "已关店"],
      ["merchant_31", "木木咖啡", "外卖", "已缴", "员工健康证临期", "2026-06-02"]
    ],
    safeguards: ["商户不能自助注册", "资质过期临时关店", "商户保证金无免押"]
  },
  riders: {
    key: "riders",
    title: "骑手/站长",
    subtitle: "邀约、在线、保证金、免押、任务时长和站点管理。",
    metrics: [
      { label: "在线骑手", value: "426", tone: "green" },
      { label: "待缴保证金", value: "18", tone: "amber" },
      { label: "免押审核", value: "6", tone: "blue" },
      { label: "纠纷退押", value: "3", tone: "red" }
    ],
    actions: ["operations-snapshot", "station-manager-invite", "rider-invite", "station-riders", "station-task-config"],
    columns: ["账号", "站点", "状态", "准入", "等级", "今日单量"],
    rows: [
      ["station_manager_2", "station_1", "在线", "站长", "A", "全部"],
      ["rider_71", "station_1", "在线", "保证金已缴", "S", "28/30"],
      ["rider_88", "station_1", "离线", "微信免押", "B", "12/30"],
      ["rider_93", "station_2", "在线", "待缴保证金", "C", "0/30"]
    ],
    safeguards: ["未缴保证金不可接单", "退押按最后一单/离职/纠纷时间顺延", "每日任务由站长配置"]
  },
  "rider-performance": {
    key: "rider-performance",
    title: "骑手绩效",
    subtitle: "接单耗时、完成率、配送评分、派单分拆解和优先级。",
    metrics: [
      { label: "团队均值", value: "42s", tone: "blue" },
      { label: "S 级骑手", value: "38", tone: "green" },
      { label: "有评分样本", value: "96", tone: "green" },
      { label: "固定单量达成", value: "72%", tone: "amber" }
    ],
    actions: ["operations-snapshot", "station-performance", "station-riders"],
    columns: ["骑手", "平均接单", "完成率", "配送评分", "派单分", "评分拆解", "等级", "派单优先级"],
    rows: [
      ["rider_71", "19s", "98%", "4.9 / 82", "118", "接单 45 / 单量 34 / 履约 15 / 评分 12", "S", "400"],
      ["rider_88", "36s", "94%", "4.8 / 45", "103", "接单 30 / 单量 33 / 履约 14 / 评分 11", "A", "300"],
      ["rider_93", "75s", "81%", "4.2 / 18", "72", "接单 18 / 单量 28 / 履约 12 / 评分 8", "C", "100"],
      ["rider_102", "44s", "91%", "4.6 / 26", "89", "接单 26 / 单量 32 / 履约 14 / 评分 10", "B", "200"]
    ],
    safeguards: ["等级按站点团队水平相对评估", "固定单量后可免责拒派", "异常取消进入审计"]
  },
  dispatch: {
    key: "dispatch",
    title: "派单审计",
    subtitle: "抢单大厅、十分钟后自动派单、拒单顺延和手动派单。",
    metrics: [
      { label: "抢单池", value: "14", tone: "blue" },
      { label: "待自动派单", value: "8", tone: "amber" },
      { label: "拒单顺延", value: "3", tone: "red" },
      { label: "手动派单", value: "2", tone: "blue" }
    ],
    actions: ["operations-snapshot", "station-orders", "station-performance", "order-compensate"],
    columns: ["订单", "阶段", "候选骑手", "策略", "下一步", "审计"],
    rows: [
      ["ord_10031", "抢单 08:12", "12", "抢单大厅", "等待", "已记录"],
      ["ord_10018", "超 10 分钟", "rider_71", "等级优先", "自动派单", "已记录"],
      ["ord_10021", "骑手拒单", "rider_88", "顺延下一位", "再派单", "已记录"],
      ["ord_10025", "站长介入", "rider_93", "手动派单", "确认中", "已记录"]
    ],
    safeguards: ["同单抢单只能一个成功", "派单超时自动转派", "站长手动派单必须留痕"]
  },
  "audit-logs": {
    key: "audit-logs",
    title: "审计检索",
    subtitle: "按操作者、动作、目标和时间游标追踪后台高风险操作。",
    metrics: [
      { label: "查询入口", value: "已接入", tone: "green" },
      { label: "过滤维度", value: "6", tone: "blue" },
      { label: "Payload", value: "脱敏", tone: "amber" },
      { label: "WORM 归档", value: "请求化", tone: "blue" }
    ],
    actions: ["audit-logs", "audit-logs-export", "audit-retention-report", "audit-retention-alert-emit", "audit-archive-request", "audit-archive-records", "audit-archive-verify", "audit-archive-verifications", "operations-snapshot"],
    columns: ["时间", "操作者", "动作", "目标", "请求", "摘要", "完整性"],
    rows: [
      ["05-22 12:00", "admin:admin_1", "admin.order.refunded", "order:ord_1", "req_1", "amount_fen: 1200", "已验证"],
      ["05-22 12:01", "admin:admin_1", "admin.refund_settings.updated", "refund_settings:default", "req_2", "default_refund_strategy: balance_first", "已验证"],
      ["05-22 12:02", "admin:admin_1", "admin.outbox.replayed", "outbox_event:obe_1", "req_3", "event_id: obe_1", "已验证"],
      ["05-22 12:03", "admin:admin_1", "after_sales.reviewed", "after_sales:asr_1", "req_4", "decision: approve", "已验证"]
    ],
    safeguards: ["默认不展示敏感 payload", "返回完整性验证状态", "留存告警与归档请求可进入可靠投递"]
  },
  "refund-settings": {
    key: "refund-settings",
    title: "退款策略",
    subtitle: "默认退余额或原路返回，按后台策略统一控制。",
    metrics: [
      { label: "默认策略", value: "余额优先", tone: "blue" },
      { label: "今日退款", value: "82", tone: "amber" },
      { label: "原路事件", value: "12", tone: "blue" },
      { label: "失败待补偿", value: "0", tone: "green" }
    ],
    actions: ["refund-settings-read", "refund-settings-save", "audit-logs"],
    columns: ["策略", "资金去向", "适用场景", "审计", "风险"],
    rows: [
      ["balance_first", "平台余额", "团购售罄/普通售后", "钱包流水", "余额风控"],
      ["original_route_first", "微信原路", "后台指定策略", "支付事件", "对账补偿"],
      ["partial_refund", "平台余额", "售后部分退款", "售后账本", "防超退"],
      ["red_packet_return", "平台余额", "红包未领取退回", "红包流水", "资金冻结"]
    ],
    safeguards: ["钱包余额只经流水变化", "退款幂等键必填", "原路退款先入 outbox"]
  },
  notifications: {
    key: "notifications",
    title: "通知运营",
    subtitle: "站内信、外部触达回执和失败原因统一追踪。",
    metrics: [
      { label: "未读站内信", value: "18", tone: "amber" },
      { label: "失败回执", value: "3", tone: "red" },
      { label: "已接渠道", value: "1", tone: "blue" },
      { label: "待接 Provider", value: "4", tone: "slate" }
    ],
    actions: ["notifications", "notification-deliveries", "notification-preferences", "notification-preference-save", "notification-preference-batch-save", "notification-preference-change-requests", "notification-preference-change-request", "notification-preference-change-review", "notification-preference-change-apply", "notification-delivery-record", "notification-failure-alert-emit", "notification-delivery-retry-schedule", "notification-quiet-window-retry-schedule", "audit-logs", "outbox-events"],
    columns: ["通知", "对象", "状态", "渠道", "来源", "回执"],
    rows: [
      ["ntf_1", "merchant:merchant_1", "unread", "in_app", "merchant.qualification_reviewed", "delivered"],
      ["ntf_2", "merchant:merchant_19", "unread", "wechat_subscribe", "merchant.qualification_reviewed", "failed invalid_openid"],
      ["ntf_3", "merchant:merchant_25", "read", "in_app", "merchant.qualification_reviewed", "delivered"],
      ["ntf_4", "merchant:merchant_31", "unread", "sms", "audit.retention_alerts", "queued"]
    ],
    safeguards: ["通知写入必须有幂等键", "失败回执保留 provider 错误码", "失败告警和重试计划必须写 outbox 与审计", "真实短信/企业微信/Push 仍需 provider 接入"]
  },
  support: {
    key: "support",
    title: "客服工作台",
    subtitle: "用户反馈、客服工单、消息风控、SLA 升级、分派、处理方案、回访、质检和绩效闭环。",
    metrics: [
      { label: "处理中", value: "18", tone: "red" },
      { label: "SLA 超时", value: "3", tone: "red" },
      { label: "待质检", value: "6", tone: "amber" },
      { label: "绩效风险", value: "2", tone: "red" }
    ],
    actions: ["support-tickets", "support-ticket-detail", "support-ticket-assign", "support-ticket-escalate", "support-ticket-resolve", "support-quality-review", "support-quality-reviews", "support-performance", "after-sales-list", "notifications", "audit-logs"],
    columns: ["工单", "用户", "类型", "状态", "客服", "SLA"],
    rows: [
      ["st_1", "user_1", "配送问题", "processing", "未分派", "SLA 超时"],
      ["st_2", "user_1", "商品质量", "waiting_confirm", "客服小悦", "待用户确认"],
      ["st_3", "user_2", "红包钱包", "resolved", "客服阿宁", "已解决"],
      ["st_4", "user_3", "功能建议", "closed", "客服小悦", "已回访"]
    ],
    safeguards: ["客服方案会展示给用户确认", "SLA 超时先升级再给出处理方案", "支付密码、验证码和银行卡消息由服务端风控拦截或标记", "分派、升级、方案提交和质检结果需要二次确认", "关闭和回访由用户侧完成，质检只影响内部绩效和辅导"]
  },
  permissions: {
    key: "permissions",
    title: "权限治理",
    subtitle: "查看服务端真实 RBAC 矩阵，追踪权限申请，并手动应用或回滚已审计的运行时权限变更。",
    metrics: [
      { label: "策略版本", value: "2026-05-24", tone: "blue" },
      { label: "后台角色", value: "7", tone: "green" },
      { label: "高危 scope", value: "6", tone: "red" },
      { label: "变更模式", value: "审批后应用", tone: "amber" }
    ],
    actions: ["rbac-policy", "rbac-change-requests", "rbac-change-request", "rbac-review-request", "rbac-apply-request", "rbac-rollback-request", "audit-logs"],
    columns: ["角色", "数据域", "可见权限", "高危动作", "审批方式"],
    rows: [
      ["super_admin", "platform", "全部", "rbac:write", "提交申请并审计"],
      ["ops_admin", "platform", "运营/邀约/售后/outbox", "order:compensate", "需超级管理员申请"],
      ["finance_admin", "finance", "退款/钱包/结算", "refund:write", "需超级管理员申请"],
      ["security_auditor", "security", "审计/RBAC 只读", "无写权限", "只读复核"]
    ],
    safeguards: ["普通分权角色只能读取策略矩阵", "申请人与审批人不能是同一管理员", "已应用变更可按审计快照回滚", "字段级和站点/商户数据域仍需继续补齐"]
  }
});

export const FALLBACK_ADMIN_VIEW = Object.freeze({
  key: "planned",
  title: "待实装模块",
  subtitle: "当前模块已进入路线图，后续按 P0/P1 接入真实页面和 API。",
  metrics: [
    { label: "状态", value: "计划中", tone: "slate" },
    { label: "权限", value: "待拆分", tone: "amber" },
    { label: "审计", value: "待接入", tone: "amber" },
    { label: "API", value: "待接入", tone: "slate" }
  ],
  actions: [],
  columns: ["能力", "状态", "下一步", "验收"],
  rows: [
    ["页面", "待实装", "接业务 API", "可操作"],
    ["权限", "待细化", "服务端策略", "可审计"],
    ["数据", "待接入", "真实列表", "可追踪"],
    ["告警", "待接入", "观测指标", "可恢复"]
  ],
  safeguards: ["先接 P0 资金/订单/资质/调度", "所有写操作留审计", "敏感字段默认脱敏"]
});

export function getAdminView(key) {
  return ADMIN_WEB_VIEWS[key] || { ...FALLBACK_ADMIN_VIEW, key };
}
