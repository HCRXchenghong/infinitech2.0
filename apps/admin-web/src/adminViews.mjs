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
    actions: ["operations-snapshot", "after-sales-list", "refund-settings-read", "outbox-stats", "object-cleanup-stats"],
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
    actions: ["operations-snapshot", "order-compensate", "outbox-events", "outbox-stats"],
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
    actions: ["operations-snapshot", "after-sales-list", "object-cleanup-candidates", "object-cleanup-stats"],
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
    actions: ["operations-snapshot", "merchant-invite"],
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
    subtitle: "接单耗时、团队均值、完成率和派单优先级。",
    metrics: [
      { label: "团队均值", value: "42s", tone: "blue" },
      { label: "S 级骑手", value: "38", tone: "green" },
      { label: "超时高风险", value: "11", tone: "red" },
      { label: "固定单量达成", value: "72%", tone: "amber" }
    ],
    actions: ["operations-snapshot", "station-performance", "station-riders"],
    columns: ["骑手", "平均接单", "完成率", "取消率", "等级", "派单优先级"],
    rows: [
      ["rider_71", "19s", "98%", "0.4%", "S", "最高"],
      ["rider_88", "36s", "94%", "1.2%", "A", "高"],
      ["rider_93", "75s", "81%", "4.8%", "C", "低"],
      ["rider_102", "44s", "91%", "2.1%", "B", "中"]
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
    actions: ["refund-settings-read", "refund-settings-save"],
    columns: ["策略", "资金去向", "适用场景", "审计", "风险"],
    rows: [
      ["balance_first", "平台余额", "团购售罄/普通售后", "钱包流水", "余额风控"],
      ["original_route_first", "微信原路", "后台指定策略", "支付事件", "对账补偿"],
      ["partial_refund", "平台余额", "售后部分退款", "售后账本", "防超退"],
      ["red_packet_return", "平台余额", "红包未领取退回", "红包流水", "资金冻结"]
    ],
    safeguards: ["钱包余额只经流水变化", "退款幂等键必填", "原路退款先入 outbox"]
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
