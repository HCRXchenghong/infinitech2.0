const EMPTY_ROW_LABEL = "暂无快照数据";

function compact(value, fallback = "-") {
  if (value === undefined || value === null || value === "") {
    return fallback;
  }
  return String(value);
}

function operationAction(label, operationKey, values = {}) {
  return { label, operationKey, values };
}

function auditAction(label, targetType, targetId) {
  return operationAction(label, "audit-logs", {
    target_type: targetType,
    target_id: targetId,
    limit: 50
  });
}

function rowId(row) {
  return compact(row?.[0], "");
}

function stationId(row) {
  return compact(row?.[1], "");
}

function notificationTarget(row) {
  const [role, id] = compact(row?.[1], "merchant:merchant_1").split(":");
  return {
    role: role || "merchant",
    id: id || "merchant_1"
  };
}

function supportAssigneeId(row) {
  const name = compact(row?.[4], "");
  if (!name || name === "未分派") return "";
  if (name === "客服小悦") return "support_1";
  if (name === "客服阿宁") return "support_2";
  return "support_1";
}

const DETAIL_CONFIGS = Object.freeze({
  dashboard: {
    title: (row) => `队列 ${rowId(row)}`,
    subtitle: "P0 队列详情、负责人和处理入口。",
    actions: (row) => [
      operationAction("刷新运营快照", "operations-snapshot", { limit: 20 }),
      operationAction("查看审计", "audit-logs", { action: "", target_type: "", limit: 50 }),
      operationAction("查看待审资质", "merchant-qualifications", { status: "pending_review", limit: 20 }),
      operationAction("查看资质明细", "merchant-qualification-detail", { qualification_id: "mq_merchant_19_health_certificate", audit_limit: 20 }),
      operationAction("查看 Outbox 健康", "outbox-stats", { lease_expiring_within_seconds: 60 }),
      operationAction("查看 Outbox 明细", "outbox-event-detail", { event_id: "obe_1", audit_limit: 20 }),
      operationAction("分诊 Outbox 死信", "outbox-dead-letter-triage", { topic: "order.paid", status: "dead_letter", limit: 20 }),
      operationAction("领取 Outbox 租约", "outbox-claim-events", { topic: "order.paid", limit: 10, lease_owner: "relay-admin", lease_seconds: 60 }),
      operationAction("续租 Outbox 租约", "outbox-renew-lease", { event_id: "obe_1", lease_owner: "relay-admin", lease_seconds: 60 }),
      operationAction("解封 Outbox 死信", "outbox-release-dead-letter", { event_id: "obe_dead_1" }),
      operationAction("恢复单个 Outbox", "outbox-replay-event", { event_id: "obe_1" }),
      operationAction("标记 Outbox 失败", "outbox-mark-failed", { event_id: "obe_1", error: "relay down", retry_after_seconds: 120, max_attempts: 10 }),
      operationAction("标记 Outbox 已发布", "outbox-mark-published", { event_id: "obe_1" })
    ],
    checklist: [
      "先确认 SLA 是否已越界",
      "高风险队列必须留下处理审计",
      "异步任务先看 ready、blocked 和 dead letter"
    ]
  },
  orders: {
    title: (row) => `订单 ${rowId(row)}`,
    subtitle: "订单状态、履约风险、补偿和审计入口。",
    actions: (row) => [
      operationAction("查看订单总览", "order-detail", { order_id: rowId(row) }),
      operationAction("订单退款", "order-refund", {
        order_id: rowId(row),
        reason: "客服确认退款",
        idempotency_key: `refund_${rowId(row)}`,
        destination: "balance"
      }),
      operationAction("查看退款流水", "refund-transactions", {
        order_id: rowId(row),
        user_id: "",
        destination: "",
        status: "",
        limit: 20
      }),
      operationAction("补偿订单状态", "order-compensate", { order_id: rowId(row) }),
      auditAction("查看订单审计", "order", rowId(row)),
      operationAction("查看订单事件", "outbox-events", { topic: "order.status_changed", status: "pending", limit: 20 })
    ],
    checklist: [
      "先核对支付流水、订单事件和派单事件",
      "状态补偿必须走服务端补偿接口",
      "退款或取消需要客服/财务复核"
    ]
  },
  "after-sales": {
    title: (row) => `售后 ${rowId(row)}`,
    subtitle: "售后工单、可退金额、证据和客服介入入口。",
    actions: (row) => [
      operationAction("查看订单总览", "order-detail", { order_id: compact(row?.[1], "") }),
      operationAction("审核售后", "after-sales-review", {
        request_id: rowId(row),
        decision: "approve",
        reason: "证据核验通过",
        refund_destination: "balance",
        refund_idempotency_key: `after_sales:${rowId(row)}`
      }),
      operationAction("查看售后详情", "after-sales-detail", { request_id: rowId(row) }),
      operationAction("查看售后列表", "after-sales-list", { request_id: rowId(row), order_id: compact(row?.[1], ""), status: "" }),
      operationAction("查看退款流水", "refund-transactions", {
        order_id: compact(row?.[1], ""),
        user_id: "",
        destination: "",
        status: "",
        limit: 20
      }),
      operationAction("查看客服工单", "support-tickets", { related_order_id: compact(row?.[1], ""), status: "", sla_status: "", assigned_support_id: "", limit: 20 }),
      operationAction("查看售后时间线", "after-sales-events", { request_id: rowId(row) }),
      operationAction("查看售后凭证", "after-sales-evidence", { request_id: rowId(row) }),
      operationAction("查看订单派单事件", "dispatch-order-events", { order_id: compact(row?.[1], ""), station_manager_id: "" }),
      auditAction("查看订单审计", "order", compact(row?.[1], "")),
      operationAction("查看退款审计", "audit-logs", {
        action: "admin.order.refunded",
        target_type: "order",
        target_id: compact(row?.[1], ""),
        actor_type: "",
        actor_id: "",
        after: "",
        before: "",
        limit: 50
      }),
      auditAction("查看售后审计", "after_sales", rowId(row)),
      operationAction("查看证据清理候选", "object-cleanup-candidates", { limit: 20, grace_seconds: 3600 })
    ],
    checklist: [
      "部分退款累计不得超过订单可退金额",
      "证据附件必须来自有效票据并通过扫描门禁",
      "内部备注和用户可见处理日志要分开"
    ]
  },
  merchants: {
    title: (row) => `商户 ${rowId(row)}`,
    subtitle: "商户准入、保证金、店铺能力和资质风险。",
    actions: (row) => [
      operationAction("查看待审资质", "merchant-qualifications", { status: "pending_review", merchant_id: rowId(row), limit: 20 }),
      operationAction("查看资质明细", "merchant-qualification-detail", { qualification_id: `mq_${rowId(row)}_health_certificate`, audit_limit: 20 }),
      operationAction("创建商户邀请", "merchant-invite", { expires_in_hours: 72 }),
      operationAction("审核商户资质", "merchant-qualification-review", {
        merchant_id: rowId(row),
        qualification_id: `mq_${rowId(row)}_health_certificate`,
        decision: "approve",
        reason: "资质原件核验通过"
      }),
      auditAction("查看商户审计", "merchant_account", rowId(row)),
      operationAction("刷新运营快照", "operations-snapshot", { limit: 20 })
    ],
    checklist: [
      "商户必须邀约注册，不能公开自助入驻",
      "营业执照、健康证和保证金共同控制接单",
      "资质过期或缺失时必须临时关店"
    ]
  },
  riders: {
    title: (row) => `骑手/站长 ${rowId(row)}`,
    subtitle: "骑手准入、在线状态、站点和任务配置。",
    actions: (row) => [
      operationAction("查看站点骑手", "station-riders", { station_id: stationId(row) }),
      operationAction("查看站点任务配置", "station-task-config", { station_id: stationId(row) }),
      auditAction("查看骑手审计", "rider", rowId(row))
    ],
    checklist: [
      "骑手未缴保证金且未免押通过时不能接单",
      "站长只能处理所属站点数据",
      "退押要按最后一单、离职和纠纷时间顺延"
    ]
  },
  "rider-performance": {
    title: (row) => `绩效 ${rowId(row)}`,
    subtitle: "接单耗时、完成率、配送评分、派单分拆解、等级和优先级。",
    actions: (row) => [
      operationAction("查看站点绩效", "station-performance", {}),
      auditAction("查看骑手审计", "rider", rowId(row)),
      operationAction("刷新运营快照", "operations-snapshot", { limit: 20 })
    ],
    checklist: [
      "等级必须按站点团队相对水平评估",
      "配送评分要来自真实完单评价，不得人工补录",
      "派单分拆解要能解释接单、单量、履约和评分加成",
      "异常取消和超时要回看派单审计",
      "固定单量达成后免责拒派要可审计"
    ]
  },
  dispatch: {
    title: (row) => `派单 ${rowId(row)}`,
    subtitle: "抢单、自动派单、拒单顺延和手动派单审计。",
    actions: (row) => [
      operationAction("查看站点待调度订单", "station-orders", {}),
      operationAction("补偿订单状态", "order-compensate", { order_id: rowId(row) }),
      auditAction("查看订单审计", "order", rowId(row))
    ],
    checklist: [
      "同一订单抢单只能成功一次",
      "派单确认超时后必须自动转派",
      "站长手动派单和拒单顺延必须留痕"
    ]
  },
  "audit-logs": {
    title: (row) => `审计 ${rowId(row)}`,
    subtitle: "审计完整性、导出、留存和归档入口。",
    actions: () => [
      operationAction("导出审计 CSV", "audit-logs-export", { limit: 100 }),
      operationAction("查看留存报告", "audit-retention-report", { retention_days: 2555, hot_days: 180 }),
      operationAction("查询归档校验历史", "audit-archive-verifications", { limit: 50 })
    ],
    checklist: [
      "导出行为本身必须写审计",
      "完整性失败要进入告警和归档复核",
      "冷归档必须保留 manifest 与校验证据"
    ]
  },
  "refund-settings": {
    title: (row) => `退款策略 ${rowId(row)}`,
    subtitle: "退款目的地、幂等、Outbox 和财务审计。",
    actions: () => [
      operationAction("读取退款策略", "refund-settings-read", {}),
      operationAction("保存退款策略", "refund-settings-save", { default_refund_strategy: "balance_first" }),
      auditAction("查看退款策略审计", "refund_settings", "default")
    ],
    checklist: [
      "钱包余额只能由流水驱动",
      "原路退款先进入 outbox 等待 worker",
      "退款策略变更必须同事务写审计"
    ]
  },
  notifications: {
    title: (row) => `通知 ${rowId(row)}`,
    subtitle: "通知账本、投递回执、失败原因和来源事件。",
    actions: (row) => {
      const target = notificationTarget(row);
      const receipt = compact(row?.[5], "");
      const isFailed = receipt.includes("failed");
      const channel = compact(row?.[3], "in_app");
      return [
        operationAction("查看通知台账", "notifications", {
          target_role: target.role,
          target_id: target.id,
          status: "all",
          source_topic: compact(row?.[4], ""),
          limit: 20
        }),
        operationAction("查看投递回执", "notification-deliveries", {
          notification_id: rowId(row),
          target_role: target.role,
          target_id: target.id,
          status: isFailed ? "failed" : "all",
          limit: 20
        }),
        operationAction("补录通知回执", "notification-delivery-record", {
          notification_id: rowId(row),
          channel,
          provider: channel,
          status: isFailed ? "failed" : "delivered",
          provider_message_id: rowId(row),
          error_code: isFailed ? "invalid_openid" : "",
          error_message: isFailed ? "provider returned invalid_openid" : "",
          idempotency_key: `delivery:manual:${rowId(row)}:${channel}`
        }),
        operationAction("投递失败告警", "notification-failure-alert-emit", {
          target_role: target.role,
          target_id: target.id,
          channel: isFailed ? channel : "",
          provider: isFailed ? channel : "",
          limit: 20
        }),
        operationAction("安排投递重试", "notification-delivery-retry-schedule", {
          target_role: target.role,
          target_id: target.id,
          channel: isFailed ? channel : "",
          provider: isFailed ? channel : "",
          limit: 20,
          retry_after_seconds: channel === "sms" ? 600 : 300
        }),
        operationAction("扫描静默重试", "notification-quiet-window-retry-schedule", {
          target_role: target.role,
          target_id: target.id,
          channel: isFailed ? "" : channel,
          provider: isFailed ? "" : channel,
          limit: 50
        }),
        operationAction("查看通知偏好", "notification-preferences", {
          target_role: target.role,
          target_id: target.id,
          notification_type: compact(row?.[4], "merchant.qualification_reviewed"),
          limit: 20
        }),
        operationAction("保存通知偏好", "notification-preference-save", {
          target_role: target.role,
          target_id: target.id,
          notification_type: compact(row?.[4], "merchant.qualification_reviewed"),
          disabled_channels: isFailed ? channel : "",
          quiet_hours: "{\"enabled\":true,\"start\":\"22:00\",\"end\":\"08:00\",\"timezone_offset\":\"+08:00\",\"channels\":[\"wechat_subscribe\",\"push\"]}"
        }),
        auditAction("查看通知审计", "platform_notification", rowId(row)),
        operationAction("查看来源事件", "outbox-events", { topic: compact(row?.[4], "merchant.qualification_reviewed"), status: "published", limit: 20 })
      ];
    },
    checklist: [
      "先核对通知账本与回执状态是否一致",
      "失败回执要保留 provider 错误码和原始错误摘要",
      "安排重试前确认 provider 故障已恢复或进入退避窗口",
      "补录回执必须使用幂等键，避免重复覆盖判断"
    ]
  },
  support: {
    title: (row) => `客服工单 ${rowId(row)}`,
    subtitle: "用户问题、SLA、客服分派和处理方案。",
    actions: (row) => [
      operationAction("查看客服工单", "support-tickets", {
        status: compact(row?.[3], ""),
        assigned_support_id: supportAssigneeId(row),
        limit: 20
      }),
      operationAction("查看工单详情", "support-ticket-detail", {
        ticket_id: rowId(row)
      }),
      operationAction("分派客服工单", "support-ticket-assign", {
        ticket_id: rowId(row),
        support_id: "support_1",
        support_name: compact(row?.[4], "") === "未分派" ? "客服小悦" : compact(row?.[4], "客服小悦")
      }),
      operationAction("升级客服工单", "support-ticket-escalate", {
        ticket_id: rowId(row),
        reason: compact(row?.[5], "").includes("超时") ? "超过 10 分钟未更新，升级给客服主管处理" : "用户问题需要主管复核",
        escalation_level: "support_lead"
      }),
      operationAction("提交处理方案", "support-ticket-resolve", {
        ticket_id: rowId(row),
        solution: compact(row?.[2], "").includes("红包") ? "已核对红包流水，未到账金额将退回余额，请用户确认。" : "已核实问题并给出补偿方案，请用户确认处理结果。"
      }),
      operationAction("抽检客服工单", "support-quality-review", {
        ticket_id: rowId(row),
        score: compact(row?.[5], "").includes("超时") ? 74 : 92,
        result: compact(row?.[5], "").includes("超时") ? "needs_coaching" : "passed",
        notes: compact(row?.[5], "").includes("超时") ? "首响超时后补偿方案完整，需复盘主动同步话术" : "处理方案完整，话术符合规范",
        coaching_required: compact(row?.[5], "").includes("超时") ? "true" : "false",
        reviewer_name: "质检主管"
      }),
      operationAction("查看质检记录", "support-quality-reviews", {
        ticket_id: rowId(row),
        support_id: supportAssigneeId(row),
        limit: 20
      }),
      operationAction("查看客服绩效", "support-performance", {
        support_id: supportAssigneeId(row),
        limit: 20
      }),
      operationAction("查看关联售后", "after-sales-list", {}),
      auditAction("查看工单审计", "service_ticket", rowId(row))
    ],
    checklist: [
      "先确认用户问题分类、关联订单和最近沟通记录",
      "分派客服后需要在 SLA 内给出处理方案",
      "处理方案会进入用户确认流程，话术不能包含支付密码或验证码要求"
    ]
  },
  permissions: {
    title: (row) => `权限 ${rowId(row)}`,
    subtitle: "角色数据域、scope、审批应用和回滚。",
    actions: (row) => [
      operationAction("读取 RBAC 策略", "rbac-policy", {}),
      operationAction("读取权限申请", "rbac-change-requests", { status: "", limit: 20 }),
      auditAction("查看角色审计", "admin_rbac_role", rowId(row))
    ],
    checklist: [
      "高危权限必须至少两人参与",
      "审批不会自动生效，必须手动应用",
      "回滚只能基于最新已应用的审计快照"
    ]
  }
});

const DEFAULT_DETAIL = Object.freeze({
  title: (row) => `详情 ${rowId(row)}`,
  subtitle: "该模块还在产品化推进中，先保留只读详情和审计入口。",
  actions: (row, view) => [
    operationAction("刷新运营快照", "operations-snapshot", { limit: 20 }),
    auditAction("查看相关审计", view.key, rowId(row))
  ],
  checklist: [
    "先补真实 API 和分页筛选",
    "所有写操作必须进入审计账本",
    "敏感字段默认脱敏展示"
  ]
});

export function buildAdminBusinessDetail(view, rowIndex) {
  const index = Number(rowIndex);
  if (!view || !Array.isArray(view.rows) || !Number.isInteger(index) || index < 0 || index >= view.rows.length) {
    return null;
  }
  const row = view.rows[index] || [];
  if (row[0] === EMPTY_ROW_LABEL) {
    return null;
  }
  const config = DETAIL_CONFIGS[view.key] || DEFAULT_DETAIL;
  const columns = Array.isArray(view.columns) ? view.columns : [];
  const detailMeta = Array.isArray(view.detailRows) ? view.detailRows[index] : null;
  return {
    id: `${view.key}:${index}`,
    moduleKey: view.key,
    rowIndex: index,
    title: config.title(row, view),
    subtitle: config.subtitle,
    facts: [
      ...columns.map((label, factIndex) => ({
        label: compact(label),
        value: compact(row[factIndex])
      })),
      ...((detailMeta?.facts || []).map((fact) => ({
        label: compact(fact?.label),
        value: compact(fact?.value)
      })))
    ],
    actions: [
      ...config.actions(row, view),
      ...((detailMeta?.actions || []).map((action) => ({
        label: compact(action?.label),
        operationKey: compact(action?.operationKey, ""),
        values: action?.values && typeof action.values === "object" ? action.values : {}
      }))).filter((action) => action.operationKey)
    ],
    checklist: [
      ...config.checklist,
      ...((detailMeta?.checklist || []).map((item) => compact(item)))
    ]
  };
}
