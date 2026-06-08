export const DEFAULT_BFF_BASE_URL = "http://localhost:25500";

export const ADMIN_API_OPERATIONS = Object.freeze([
  {
    key: "admin-login",
    title: "管理员登录",
    method: "POST",
    path: "/api/auth/admin/login",
    authRequired: false,
    area: "auth",
    fields: [
      { key: "account_id", label: "账号", type: "text", defaultValue: "admin_1", required: true },
      { key: "password", label: "密码", type: "password", required: true }
    ]
  },
  {
    key: "merchant-invite",
    title: "创建商户邀请",
    method: "POST",
    path: "/api/admin/merchant-invites",
    authRequired: true,
    area: "merchant",
    fields: [
      { key: "expires_in_hours", label: "有效小时", type: "number", defaultValue: 72 },
      { key: "note", label: "备注", type: "text", defaultValue: "merchant onboarding" }
    ]
  },
  {
    key: "merchant-qualifications",
    title: "商户资质待审列表",
    method: "GET",
    path: "/api/admin/merchant-qualifications",
    authRequired: true,
    area: "merchant",
    queryFields: [
      { key: "status", label: "状态", type: "select", defaultValue: "pending_review", options: ["pending_review", "approved", "rejected", "expired", "all"] },
      { key: "merchant_id", label: "商户 ID", type: "text", defaultValue: "" },
      { key: "type", label: "资质类型", type: "select", defaultValue: "", options: ["", "business_license", "health_certificate", "supplemental_document"] },
      { key: "limit", label: "条数", type: "number", defaultValue: 20 },
      { key: "now", label: "查询时间", type: "text", defaultValue: "" }
    ],
    fields: []
  },
  {
    key: "merchant-qualification-detail",
    title: "商户资质明细",
    method: "GET",
    path: "/api/admin/merchant-qualifications/:qualification_id",
    authRequired: true,
    area: "merchant",
    pathFields: [
      { key: "qualification_id", label: "资质 ID", type: "text", defaultValue: "mq_merchant_19_health_certificate", required: true }
    ],
    queryFields: [
      { key: "audit_limit", label: "审计条数", type: "number", defaultValue: 20 },
      { key: "now", label: "诊断时间", type: "text", defaultValue: "" }
    ],
    fields: []
  },
  {
    key: "merchant-qualification-review",
    title: "审核商户资质",
    method: "POST",
    path: "/api/admin/merchant-qualifications/:qualification_id/review",
    authRequired: true,
    area: "merchant",
    pathFields: [
      { key: "qualification_id", label: "资质 ID", type: "text", defaultValue: "mq_merchant_19_health_certificate", required: true }
    ],
    fields: [
      { key: "merchant_id", label: "商户 ID", type: "text", defaultValue: "merchant_19", required: true },
      { key: "decision", label: "审核结果", type: "select", defaultValue: "approve", options: ["approve", "reject"], required: true },
      { key: "reason", label: "审核原因", type: "text", defaultValue: "资质原件核验通过", required: true },
      { key: "reviewed_at", label: "审核时间", type: "text", defaultValue: "" }
    ]
  },
  {
    key: "station-manager-invite",
    title: "创建站长邀请",
    method: "POST",
    path: "/api/admin/rider-invites",
    authRequired: true,
    area: "rider",
    fields: [
      { key: "type", label: "类型", type: "select", defaultValue: "station_manager", options: ["station_manager"] },
      { key: "station_id", label: "站点", type: "text", defaultValue: "station_1", required: true }
    ]
  },
  {
    key: "rider-invite",
    title: "创建骑手邀请",
    method: "POST",
    path: "/api/admin/rider-invites",
    authRequired: true,
    area: "rider",
    fields: [
      { key: "type", label: "类型", type: "select", defaultValue: "rider", options: ["rider"] },
      { key: "station_id", label: "站点", type: "text", defaultValue: "station_1", required: true }
    ]
  },
  {
    key: "refund-settings-read",
    title: "读取退款策略",
    method: "GET",
    path: "/api/admin/refund-settings",
    authRequired: true,
    area: "finance",
    fields: []
  },
  {
    key: "refund-settings-save",
    title: "保存退款策略",
    method: "PUT",
    path: "/api/admin/refund-settings",
    authRequired: true,
    area: "finance",
    fields: [
      { key: "default_refund_strategy", label: "默认策略", type: "select", defaultValue: "balance_first", options: ["balance_first", "original_route_first"] }
    ]
  },
  {
    key: "order-refund",
    title: "订单退款",
    method: "POST",
    path: "/api/orders/:order_id/refund",
    authRequired: true,
    area: "finance",
    pathFields: [
      { key: "order_id", label: "订单号", type: "text", defaultValue: "ord_10031", required: true }
    ],
    fields: [
      { key: "reason", label: "退款原因", type: "text", defaultValue: "客服确认退款", required: true },
      { key: "idempotency_key", label: "退款幂等键", type: "text", defaultValue: "refund_ord_10031", required: true },
      { key: "amount_fen", label: "退款金额分", type: "number" },
      { key: "destination", label: "退款去向", type: "select", defaultValue: "balance", options: ["", "balance", "original_route"] }
    ]
  },
  {
    key: "order-detail",
    title: "订单聚合详情",
    method: "GET",
    path: "/api/admin/orders/:order_id",
    authRequired: true,
    area: "order",
    pathFields: [
      { key: "order_id", label: "订单号", type: "text", defaultValue: "ord_10031", required: true }
    ],
    fields: []
  },
  {
    key: "refund-transactions",
    title: "退款流水",
    method: "GET",
    path: "/api/admin/refunds",
    authRequired: true,
    area: "finance",
    queryFields: [
      { key: "order_id", label: "订单号", type: "text", defaultValue: "" },
      { key: "user_id", label: "用户 ID", type: "text", defaultValue: "" },
      { key: "destination", label: "退款去向", type: "select", defaultValue: "", options: ["", "balance", "original_route"] },
      { key: "status", label: "状态", type: "select", defaultValue: "", options: ["", "success", "pending", "failed"] },
      { key: "limit", label: "条数", type: "number", defaultValue: 20 }
    ],
    fields: []
  },
  {
    key: "after-sales-list",
    title: "售后审核列表",
    method: "GET",
    path: "/api/admin/after-sales",
    authRequired: true,
    area: "support",
    queryFields: [
      { key: "request_id", label: "售后工单", type: "text", defaultValue: "" },
      { key: "order_id", label: "订单号", type: "text", defaultValue: "" },
      { key: "status", label: "状态", type: "select", defaultValue: "", options: ["", "pending_merchant", "admin_review", "approved", "rejected", "refunded"] }
    ],
    fields: []
  },
  {
    key: "after-sales-detail",
    title: "售后聚合详情",
    method: "GET",
    path: "/api/admin/after-sales/:request_id",
    authRequired: true,
    area: "support",
    pathFields: [
      { key: "request_id", label: "售后工单", type: "text", defaultValue: "asr_231", required: true }
    ],
    fields: []
  },
  {
    key: "after-sales-events",
    title: "售后时间线",
    method: "GET",
    path: "/api/after-sales/:request_id/events",
    authRequired: true,
    area: "support",
    pathFields: [
      { key: "request_id", label: "售后工单", type: "text", defaultValue: "asr_231", required: true }
    ],
    fields: []
  },
  {
    key: "after-sales-evidence",
    title: "售后凭证",
    method: "GET",
    path: "/api/after-sales/:request_id/evidence",
    authRequired: true,
    area: "support",
    pathFields: [
      { key: "request_id", label: "售后工单", type: "text", defaultValue: "asr_231", required: true }
    ],
    fields: []
  },
  {
    key: "after-sales-review",
    title: "审核售后",
    method: "POST",
    path: "/api/after-sales/:request_id/review",
    authRequired: true,
    area: "support",
    pathFields: [
      { key: "request_id", label: "售后工单", type: "text", defaultValue: "asr_231", required: true }
    ],
    fields: [
      { key: "decision", label: "审核结果", type: "select", defaultValue: "approve", options: ["approve", "reject", "escalate"], required: true },
      { key: "reason", label: "审核原因", type: "text", defaultValue: "证据核验通过", required: true },
      { key: "refund_destination", label: "退款去向", type: "select", defaultValue: "balance", options: ["", "balance", "original_route"] },
      { key: "refund_idempotency_key", label: "退款幂等键", type: "text", defaultValue: "after_sales:asr_231" }
    ]
  },
  {
    key: "support-tickets",
    title: "客服工单列表",
    method: "GET",
    path: "/api/admin/service-tickets",
    authRequired: true,
    area: "support",
    queryFields: [
      { key: "user_id", label: "用户 ID", type: "text", defaultValue: "" },
      { key: "related_order_id", label: "关联订单", type: "text", defaultValue: "" },
      { key: "status", label: "状态", type: "select", defaultValue: "", options: ["", "processing", "waiting_confirm", "resolved", "closed"] },
      { key: "sla_status", label: "SLA", type: "select", defaultValue: "", options: ["", "normal", "due_soon", "overdue", "escalated", "completed"] },
      { key: "assigned_support_id", label: "客服 ID", type: "text", defaultValue: "" },
      { key: "limit", label: "条数", type: "number", defaultValue: 20 },
      { key: "now", label: "查询时间", type: "text", defaultValue: "" }
    ],
    fields: []
  },
  {
    key: "support-ticket-detail",
    title: "客服工单详情",
    method: "GET",
    path: "/api/admin/service-tickets/:ticket_id",
    authRequired: true,
    area: "support",
    pathFields: [
      { key: "ticket_id", label: "工单 ID", type: "text", defaultValue: "st_1", required: true }
    ],
    fields: []
  },
  {
    key: "support-ticket-assign",
    title: "分派客服工单",
    method: "POST",
    path: "/api/admin/service-tickets/:ticket_id/assign",
    authRequired: true,
    area: "support",
    pathFields: [
      { key: "ticket_id", label: "工单 ID", type: "text", defaultValue: "st_1", required: true }
    ],
    fields: [
      { key: "support_id", label: "客服 ID", type: "text", defaultValue: "support_1" },
      { key: "support_name", label: "客服姓名", type: "text", defaultValue: "客服小悦" },
      { key: "actor_id", label: "操作人", type: "text", defaultValue: "" }
    ]
  },
  {
    key: "support-ticket-escalate",
    title: "升级客服工单",
    method: "POST",
    path: "/api/admin/service-tickets/:ticket_id/escalate",
    authRequired: true,
    area: "support",
    pathFields: [
      { key: "ticket_id", label: "工单 ID", type: "text", defaultValue: "st_1", required: true }
    ],
    fields: [
      { key: "reason", label: "升级原因", type: "text", defaultValue: "超过 10 分钟未更新，升级给客服主管处理", required: true },
      { key: "escalation_level", label: "升级级别", type: "select", defaultValue: "support_lead", options: ["support_lead", "ops_manager", "risk_review"], required: true },
      { key: "actor_id", label: "操作人", type: "text", defaultValue: "" }
    ]
  },
  {
    key: "support-ticket-resolve",
    title: "提交工单方案",
    method: "POST",
    path: "/api/admin/service-tickets/:ticket_id/resolve",
    authRequired: true,
    area: "support",
    pathFields: [
      { key: "ticket_id", label: "工单 ID", type: "text", defaultValue: "st_1", required: true }
    ],
    fields: [
      { key: "solution", label: "处理方案", type: "text", defaultValue: "已发放 5 元延误券，请用户确认处理结果", required: true },
      { key: "actor_id", label: "操作人", type: "text", defaultValue: "" }
    ]
  },
  {
    key: "support-quality-reviews",
    title: "客服质检记录",
    method: "GET",
    path: "/api/admin/service-ticket-quality-reviews",
    authRequired: true,
    area: "support",
    queryFields: [
      { key: "ticket_id", label: "工单 ID", type: "text", defaultValue: "" },
      { key: "support_id", label: "客服 ID", type: "text", defaultValue: "" },
      { key: "reviewer_id", label: "质检员", type: "text", defaultValue: "" },
      { key: "result", label: "质检结果", type: "select", defaultValue: "", options: ["", "passed", "needs_coaching", "critical"] },
      { key: "coaching_required", label: "需辅导", type: "select", defaultValue: "", options: ["", "true", "false"] },
      { key: "limit", label: "条数", type: "number", defaultValue: 20 }
    ],
    fields: []
  },
  {
    key: "support-performance",
    title: "客服绩效汇总",
    method: "GET",
    path: "/api/admin/service-ticket-performance",
    authRequired: true,
    area: "support",
    queryFields: [
      { key: "support_id", label: "客服 ID", type: "text", defaultValue: "" },
      { key: "limit", label: "条数", type: "number", defaultValue: 20 },
      { key: "now", label: "查询时间", type: "text", defaultValue: "" }
    ],
    fields: []
  },
  {
    key: "support-quality-review",
    title: "抽检客服工单",
    method: "POST",
    path: "/api/admin/service-tickets/:ticket_id/quality-review",
    authRequired: true,
    area: "support",
    pathFields: [
      { key: "ticket_id", label: "工单 ID", type: "text", defaultValue: "st_1", required: true }
    ],
    fields: [
      { key: "score", label: "质检分", type: "number", defaultValue: 88, required: true },
      { key: "result", label: "质检结果", type: "select", defaultValue: "", options: ["", "passed", "needs_coaching", "critical"] },
      { key: "notes", label: "质检备注", type: "text", defaultValue: "处理方案完整，话术符合规范", required: true },
      { key: "coaching_required", label: "需要辅导", type: "boolean", defaultValue: "false" },
      { key: "reviewer_id", label: "质检员", type: "text", defaultValue: "" },
      { key: "reviewer_name", label: "质检员姓名", type: "text", defaultValue: "质检主管" }
    ]
  },
  {
    key: "operations-snapshot",
    title: "运营快照",
    method: "GET",
    path: "/api/admin/operations/snapshot",
    authRequired: true,
    area: "ops",
    queryFields: [
      { key: "limit", label: "条数", type: "number", defaultValue: 20 },
      { key: "lease_expiring_within_seconds", label: "租约预警秒", type: "number", defaultValue: 60 },
      { key: "object_cleanup_grace_seconds", label: "对象清理宽限秒", type: "number", defaultValue: 3600 }
    ],
    fields: []
  },
  {
    key: "audit-logs",
    title: "操作审计",
    method: "GET",
    path: "/api/admin/audit-logs",
    authRequired: true,
    area: "ops",
    queryFields: [
      { key: "actor_type", label: "操作者类型", type: "select", defaultValue: "", options: ["", "admin", "merchant", "station_manager", "rider"] },
      { key: "actor_id", label: "操作者 ID", type: "text", defaultValue: "" },
      { key: "target_type", label: "目标类型", type: "text", defaultValue: "" },
      { key: "target_id", label: "目标 ID", type: "text", defaultValue: "" },
      { key: "action", label: "动作", type: "text", defaultValue: "" },
      { key: "after", label: "晚于时间", type: "text", defaultValue: "" },
      { key: "before", label: "早于时间", type: "text", defaultValue: "" },
      { key: "limit", label: "条数", type: "number", defaultValue: 20 }
    ],
    fields: []
  },
  {
    key: "audit-logs-export",
    title: "导出审计 CSV",
    method: "GET",
    path: "/api/admin/audit-logs/export",
    authRequired: true,
    area: "ops",
    queryFields: [
      { key: "actor_type", label: "操作者类型", type: "select", defaultValue: "", options: ["", "admin", "merchant", "station_manager", "rider"] },
      { key: "actor_id", label: "操作者 ID", type: "text", defaultValue: "" },
      { key: "target_type", label: "目标类型", type: "text", defaultValue: "" },
      { key: "target_id", label: "目标 ID", type: "text", defaultValue: "" },
      { key: "action", label: "动作", type: "text", defaultValue: "" },
      { key: "after", label: "晚于时间", type: "text", defaultValue: "" },
      { key: "before", label: "早于时间", type: "text", defaultValue: "" },
      { key: "limit", label: "条数", type: "number", defaultValue: 100 }
    ],
    fields: []
  },
  {
    key: "audit-retention-report",
    title: "审计留存告警",
    method: "GET",
    path: "/api/admin/audit-logs/retention-report",
    authRequired: true,
    area: "ops",
    queryFields: [
      { key: "retention_days", label: "留存天数", type: "number", defaultValue: 2555 },
      { key: "hot_days", label: "热存天数", type: "number", defaultValue: 180 },
      { key: "integrity_sample_limit", label: "完整性样本", type: "number", defaultValue: 500 }
    ],
    fields: []
  },
  {
    key: "audit-retention-alert-emit",
    title: "投递审计告警",
    method: "POST",
    path: "/api/admin/audit-logs/retention-alerts/emit",
    authRequired: true,
    area: "ops",
    fields: [
      { key: "retention_days", label: "留存天数", type: "number", defaultValue: 2555 },
      { key: "hot_days", label: "热存天数", type: "number", defaultValue: 180 },
      { key: "integrity_sample_limit", label: "完整性样本", type: "number", defaultValue: 500 }
    ]
  },
  {
    key: "audit-archive-request",
    title: "请求 WORM 归档",
    method: "POST",
    path: "/api/admin/audit-logs/archive/request",
    authRequired: true,
    area: "ops",
    fields: [
      { key: "hot_days", label: "热存天数", type: "number", defaultValue: 180 },
      { key: "limit", label: "归档条数", type: "number", defaultValue: 500 },
      { key: "storage_prefix", label: "归档前缀", type: "text", defaultValue: "worm://audit-logs" }
    ]
  },
  {
    key: "audit-archive-records",
    title: "归档完成记录",
    method: "GET",
    path: "/api/admin/audit-logs/archive/records",
    authRequired: true,
    area: "ops",
    queryFields: [
      { key: "archive_id", label: "归档 ID", type: "text", defaultValue: "" },
      { key: "limit", label: "条数", type: "number", defaultValue: 50 }
    ],
    fields: []
  },
  {
    key: "audit-archive-verify",
    title: "校验归档对象",
    method: "POST",
    path: "/api/admin/audit-logs/archive/verify",
    authRequired: true,
    area: "ops",
    fields: [
      { key: "archive_id", label: "归档 ID", type: "text", defaultValue: "audit_archive_1", required: true }
    ]
  },
  {
    key: "audit-archive-verifications",
    title: "归档校验历史",
    method: "GET",
    path: "/api/admin/audit-logs/archive/verifications",
    authRequired: true,
    area: "ops",
    queryFields: [
      { key: "archive_id", label: "归档 ID", type: "text", defaultValue: "" },
      { key: "status", label: "状态", type: "select", defaultValue: "", options: ["", "verified", "failed"] },
      { key: "limit", label: "条数", type: "number", defaultValue: 50 }
    ],
    fields: []
  },
  {
    key: "notifications",
    title: "通知台账",
    method: "GET",
    path: "/api/admin/notifications",
    authRequired: true,
    area: "notifications",
    queryFields: [
      { key: "target_role", label: "目标角色", type: "select", defaultValue: "merchant", options: ["", "merchant", "user", "rider", "security"] },
      { key: "target_id", label: "目标 ID", type: "text", defaultValue: "merchant_1" },
      { key: "status", label: "状态", type: "select", defaultValue: "unread", options: ["", "all", "unread", "read"] },
      { key: "source_topic", label: "来源主题", type: "text", defaultValue: "merchant.qualification_reviewed" },
      { key: "source_event_id", label: "来源事件", type: "text", defaultValue: "" },
      { key: "limit", label: "条数", type: "number", defaultValue: 20 }
    ],
    fields: []
  },
  {
    key: "notification-deliveries",
    title: "通知回执",
    method: "GET",
    path: "/api/admin/notification-deliveries",
    authRequired: true,
    area: "notifications",
    queryFields: [
      { key: "notification_id", label: "通知 ID", type: "text", defaultValue: "" },
      { key: "target_role", label: "目标角色", type: "select", defaultValue: "merchant", options: ["", "merchant", "user", "rider", "security"] },
      { key: "target_id", label: "目标 ID", type: "text", defaultValue: "merchant_1" },
      { key: "channel", label: "渠道", type: "select", defaultValue: "", options: ["", "in_app", "wechat_subscribe", "sms", "enterprise_wechat", "push"] },
      { key: "provider", label: "Provider", type: "text", defaultValue: "" },
      { key: "status", label: "状态", type: "select", defaultValue: "failed", options: ["", "all", "queued", "delivered", "failed"] },
      { key: "error_code", label: "错误码", type: "text", defaultValue: "" },
      { key: "retry_at_before", label: "重试到期前", type: "text", defaultValue: "" },
      { key: "limit", label: "条数", type: "number", defaultValue: 20 }
    ],
    fields: []
  },
  {
    key: "notification-preferences",
    title: "通知偏好",
    method: "GET",
    path: "/api/admin/notification-preferences",
    authRequired: true,
    area: "notifications",
    queryFields: [
      { key: "preference_key", label: "偏好 Key", type: "text", defaultValue: "" },
      { key: "target_role", label: "目标角色", type: "select", defaultValue: "merchant", options: ["", "merchant", "user", "rider", "security"] },
      { key: "target_id", label: "目标 ID", type: "text", defaultValue: "merchant_1" },
      { key: "notification_type", label: "通知类型", type: "text", defaultValue: "merchant.qualification_reviewed" },
      { key: "limit", label: "条数", type: "number", defaultValue: 20 }
    ],
    fields: []
  },
  {
    key: "notification-preference-save",
    title: "保存通知偏好",
    method: "PUT",
    path: "/api/admin/notification-preferences",
    authRequired: true,
    area: "notifications",
    fields: [
      { key: "target_role", label: "目标角色", type: "select", defaultValue: "merchant", options: ["merchant", "user", "rider", "security"], required: true },
      { key: "target_id", label: "目标 ID", type: "text", defaultValue: "merchant_1", required: true },
      { key: "notification_type", label: "通知类型", type: "text", defaultValue: "merchant.qualification_reviewed", required: true },
      { key: "enabled_channels", label: "启用渠道", type: "csv", defaultValue: "wechat_subscribe,push" },
      { key: "disabled_channels", label: "禁用渠道", type: "csv", defaultValue: "sms" },
      { key: "quiet_hours", label: "静默规则 JSON", type: "json", defaultValue: "{\"enabled\":true,\"start\":\"22:00\",\"end\":\"08:00\",\"timezone_offset\":\"+08:00\",\"channels\":[\"wechat_subscribe\",\"push\"]}" },
      { key: "updated_at", label: "更新时间", type: "text", defaultValue: "" }
    ]
  },
  {
    key: "notification-preference-batch-save",
    title: "批量保存通知偏好",
    method: "POST",
    path: "/api/admin/notification-preferences/batch",
    authRequired: true,
    area: "notifications",
    fields: [
      { key: "preferences", label: "偏好策略 JSON", type: "json", defaultValue: "[{\"target_role\":\"merchant\",\"target_id\":\"merchant_1\",\"notification_type\":\"merchant.qualification_reviewed\",\"disabled_channels\":[\"sms\"],\"quiet_hours\":{\"enabled\":true,\"start\":\"22:00\",\"end\":\"08:00\",\"timezone_offset\":\"+08:00\",\"channels\":[\"wechat_subscribe\",\"push\"]}},{\"target_role\":\"merchant\",\"target_id\":\"merchant_1\",\"notification_type\":\"order.status_changed\",\"enabled_channels\":[\"wechat_subscribe\",\"push\"],\"disabled_channels\":[\"sms\"]}]", required: true },
      { key: "reason", label: "变更原因", type: "text", defaultValue: "批量更新关键通知触达策略", required: true },
      { key: "updated_at", label: "更新时间", type: "text", defaultValue: "" }
    ]
  },
  {
    key: "notification-preference-change-requests",
    title: "通知偏好变更申请",
    method: "GET",
    path: "/api/admin/notification-preferences/change-requests",
    authRequired: true,
    area: "notifications",
    queryFields: [
      { key: "status", label: "状态", type: "select", defaultValue: "pending_approval", options: ["", "pending_approval", "approved", "rejected", "applied"] },
      { key: "limit", label: "条数", type: "number", defaultValue: 20 }
    ],
    fields: []
  },
  {
    key: "notification-preference-change-request",
    title: "提交通知偏好变更",
    method: "POST",
    path: "/api/admin/notification-preferences/change-requests",
    authRequired: true,
    area: "notifications",
    fields: [
      { key: "preferences", label: "偏好策略 JSON", type: "json", defaultValue: "[{\"target_role\":\"merchant\",\"target_id\":\"merchant_1\",\"notification_type\":\"merchant.qualification_reviewed\",\"disabled_channels\":[\"sms\"]},{\"target_role\":\"user\",\"target_id\":\"user_1\",\"notification_type\":\"after_sales.updated\",\"disabled_channels\":[\"push\"]}]", required: true },
      { key: "rollout", label: "灰度策略 JSON", type: "json", defaultValue: "{\"mode\":\"target_ids\",\"target_ids\":[\"merchant_1\"],\"max_targets\":10}", required: true },
      { key: "reason", label: "申请原因", type: "text", defaultValue: "申请批量调整关键通知触达策略", required: true },
      { key: "updated_at", label: "计划更新时间", type: "text", defaultValue: "" }
    ]
  },
  {
    key: "notification-preference-change-review",
    title: "审批通知偏好变更",
    method: "POST",
    path: "/api/admin/notification-preferences/change-requests/:change_request_id/review",
    authRequired: true,
    area: "notifications",
    pathFields: [
      { key: "change_request_id", label: "申请 ID", type: "text", defaultValue: "ntfp_change_1", required: true }
    ],
    fields: [
      { key: "decision", label: "审批结果", type: "select", defaultValue: "approve", options: ["approve", "reject"], required: true },
      { key: "reason", label: "审批原因", type: "text", defaultValue: "策略范围和静默窗口已复核", required: true }
    ]
  },
  {
    key: "notification-preference-change-apply",
    title: "应用通知偏好变更",
    method: "POST",
    path: "/api/admin/notification-preferences/change-requests/:change_request_id/apply",
    authRequired: true,
    area: "notifications",
    pathFields: [
      { key: "change_request_id", label: "申请 ID", type: "text", defaultValue: "ntfp_change_1", required: true }
    ],
    fields: [
      { key: "reason", label: "应用原因", type: "text", defaultValue: "按已审批策略应用到生产偏好账本", required: true },
      { key: "updated_at", label: "应用时间", type: "text", defaultValue: "" }
    ]
  },
  {
    key: "notification-delivery-record",
    title: "补录通知回执",
    method: "POST",
    path: "/api/notifications/:notification_id/deliveries",
    authRequired: true,
    area: "notifications",
    pathFields: [
      { key: "notification_id", label: "通知 ID", type: "text", defaultValue: "ntf_1", required: true }
    ],
    fields: [
      { key: "channel", label: "渠道", type: "select", defaultValue: "in_app", options: ["in_app", "wechat_subscribe", "sms", "enterprise_wechat", "push"] },
      { key: "provider", label: "Provider", type: "text", defaultValue: "in_app" },
      { key: "status", label: "状态", type: "select", defaultValue: "delivered", options: ["queued", "delivered", "failed"], required: true },
      { key: "provider_message_id", label: "Provider 消息 ID", type: "text", defaultValue: "ntf_1" },
      { key: "error_code", label: "错误码", type: "text", defaultValue: "" },
      { key: "error_message", label: "错误信息", type: "text", defaultValue: "" },
      { key: "idempotency_key", label: "回执幂等键", type: "text", defaultValue: "delivery:manual:ntf_1:in_app", required: true },
      { key: "attempted_at", label: "尝试时间", type: "text", defaultValue: "" },
      { key: "delivered_at", label: "送达时间", type: "text", defaultValue: "" }
    ]
  },
  {
    key: "notification-failure-alert-emit",
    title: "投递失败告警",
    method: "POST",
    path: "/api/admin/notification-deliveries/failure-alerts/emit",
    authRequired: true,
    area: "notifications",
    fields: [
      { key: "target_role", label: "目标角色", type: "select", defaultValue: "merchant", options: ["", "merchant", "user", "rider", "security"] },
      { key: "target_id", label: "目标 ID", type: "text", defaultValue: "merchant_1" },
      { key: "channel", label: "渠道", type: "select", defaultValue: "wechat_subscribe", options: ["", "in_app", "wechat_subscribe", "sms", "enterprise_wechat", "push"] },
      { key: "provider", label: "Provider", type: "text", defaultValue: "" },
      { key: "limit", label: "失败条数", type: "number", defaultValue: 20 },
      { key: "now", label: "告警时间", type: "text", defaultValue: "" }
    ]
  },
  {
    key: "notification-delivery-retry-schedule",
    title: "安排投递重试",
    method: "POST",
    path: "/api/admin/notification-deliveries/retries/schedule",
    authRequired: true,
    area: "notifications",
    fields: [
      { key: "target_role", label: "目标角色", type: "select", defaultValue: "merchant", options: ["", "merchant", "user", "rider", "security"] },
      { key: "target_id", label: "目标 ID", type: "text", defaultValue: "merchant_1" },
      { key: "channel", label: "渠道", type: "select", defaultValue: "wechat_subscribe", options: ["", "in_app", "wechat_subscribe", "sms", "enterprise_wechat", "push"] },
      { key: "provider", label: "Provider", type: "text", defaultValue: "" },
      { key: "status", label: "回执状态", type: "select", defaultValue: "failed", options: ["failed", "queued"] },
      { key: "error_code", label: "错误码", type: "text", defaultValue: "" },
      { key: "limit", label: "回执条数", type: "number", defaultValue: 20 },
      { key: "retry_after_seconds", label: "重试延迟秒", type: "number", defaultValue: 300 },
      { key: "retry_at", label: "指定重试时间", type: "text", defaultValue: "" },
      { key: "now", label: "计划时间", type: "text", defaultValue: "" }
    ]
  },
  {
    key: "notification-quiet-window-retry-schedule",
    title: "扫描静默重试",
    method: "POST",
    path: "/api/admin/notification-deliveries/quiet-window-retries/schedule",
    authRequired: true,
    area: "notifications",
    fields: [
      { key: "target_role", label: "目标角色", type: "select", defaultValue: "merchant", options: ["", "merchant", "user", "rider", "security"] },
      { key: "target_id", label: "目标 ID", type: "text", defaultValue: "" },
      { key: "channel", label: "渠道", type: "select", defaultValue: "", options: ["", "wechat_subscribe", "sms", "enterprise_wechat", "push"] },
      { key: "provider", label: "Provider", type: "text", defaultValue: "" },
      { key: "limit", label: "扫描条数", type: "number", defaultValue: 50 },
      { key: "retry_after_seconds", label: "到期后延迟秒", type: "number", defaultValue: 0 },
      { key: "now", label: "扫描时间", type: "text", defaultValue: "" }
    ]
  },
  {
    key: "rbac-policy",
    title: "RBAC 策略矩阵",
    method: "GET",
    path: "/api/admin/rbac/policy",
    authRequired: true,
    area: "security",
    fields: []
  },
  {
    key: "rbac-change-requests",
    title: "权限申请列表",
    method: "GET",
    path: "/api/admin/rbac/change-requests",
    authRequired: true,
    area: "security",
    queryFields: [
      { key: "status", label: "状态", type: "select", defaultValue: "pending_approval", options: ["", "pending_approval", "approved", "rejected", "applied", "rolled_back"] },
      { key: "limit", label: "条数", type: "number", defaultValue: 20 }
    ],
    fields: []
  },
  {
    key: "rbac-change-request",
    title: "提交权限变更申请",
    method: "POST",
    path: "/api/admin/rbac/change-requests",
    authRequired: true,
    area: "security",
    fields: [
      { key: "role", label: "目标角色", type: "select", defaultValue: "support_admin", options: ["ops_admin", "finance_admin", "dispatch_admin", "support_admin", "security_auditor"], required: true },
      { key: "requested_scopes", label: "申请 scopes", type: "csv", defaultValue: "after_sales:read,after_sales:event,rbac:read", required: true },
      { key: "reason", label: "申请原因", type: "text", defaultValue: "least privilege recertification", required: true }
    ]
  },
  {
    key: "rbac-review-request",
    title: "审批权限申请",
    method: "POST",
    path: "/api/admin/rbac/change-requests/:change_request_id/review",
    authRequired: true,
    area: "security",
    pathFields: [
      { key: "change_request_id", label: "申请 ID", type: "text", defaultValue: "rbac_change_1", required: true }
    ],
    fields: [
      { key: "decision", label: "审批结果", type: "select", defaultValue: "approve", options: ["approve", "reject"], required: true },
      { key: "reason", label: "审批原因", type: "text", defaultValue: "least privilege reviewed", required: true }
    ]
  },
  {
    key: "rbac-apply-request",
    title: "应用已审批权限",
    method: "POST",
    path: "/api/admin/rbac/change-requests/:change_request_id/apply",
    authRequired: true,
    area: "security",
    pathFields: [
      { key: "change_request_id", label: "申请 ID", type: "text", defaultValue: "rbac_change_1", required: true }
    ],
    fields: [
      { key: "reason", label: "应用原因", type: "text", defaultValue: "approved scope change applied to runtime policy", required: true }
    ]
  },
  {
    key: "rbac-rollback-request",
    title: "回滚已应用权限",
    method: "POST",
    path: "/api/admin/rbac/change-requests/:change_request_id/rollback",
    authRequired: true,
    area: "security",
    pathFields: [
      { key: "change_request_id", label: "申请 ID", type: "text", defaultValue: "rbac_change_1", required: true }
    ],
    fields: [
      { key: "reason", label: "回滚原因", type: "text", defaultValue: "restore previous audited runtime policy", required: true }
    ]
  },
  {
    key: "object-cleanup-stats",
    title: "对象清理统计",
    method: "GET",
    path: "/api/admin/object-storage/cleanup-stats",
    authRequired: true,
    area: "ops",
    queryFields: [
      { key: "grace_seconds", label: "宽限秒", type: "number", defaultValue: 3600 }
    ],
    fields: []
  },
  {
    key: "object-cleanup-candidates",
    title: "对象清理候选",
    method: "GET",
    path: "/api/admin/object-storage/cleanup-candidates",
    authRequired: true,
    area: "ops",
    queryFields: [
      { key: "limit", label: "条数", type: "number", defaultValue: 20 },
      { key: "grace_seconds", label: "宽限秒", type: "number", defaultValue: 3600 }
    ],
    fields: []
  },
  {
    key: "outbox-stats",
    title: "Outbox 健康",
    method: "GET",
    path: "/api/admin/outbox/stats",
    authRequired: true,
    area: "ops",
    queryFields: [
      { key: "topic", label: "主题", type: "text", defaultValue: "order.paid" },
      { key: "lease_expiring_within_seconds", label: "租约预警秒", type: "number", defaultValue: 60 }
    ],
    fields: []
  },
  {
    key: "outbox-events",
    title: "Outbox 事件",
    method: "GET",
    path: "/api/admin/outbox/events",
    authRequired: true,
    area: "ops",
    queryFields: [
      { key: "topic", label: "主题", type: "text", defaultValue: "order.paid" },
      { key: "status", label: "状态", type: "select", defaultValue: "pending", options: ["pending", "failed", "dead_letter", "published"] },
      { key: "limit", label: "条数", type: "number", defaultValue: 20 }
    ],
    fields: []
  },
  {
    key: "outbox-event-detail",
    title: "Outbox 事件明细",
    method: "GET",
    path: "/api/admin/outbox/events/:event_id",
    authRequired: true,
    area: "ops",
    pathFields: [
      { key: "event_id", label: "事件 ID", type: "text", defaultValue: "obe_1", required: true }
    ],
    queryFields: [
      { key: "now", label: "诊断时间", type: "text", defaultValue: "" },
      { key: "audit_limit", label: "审计条数", type: "number", defaultValue: 20 }
    ],
    fields: []
  },
  {
    key: "outbox-dead-letter-triage",
    title: "Outbox 死信分诊",
    method: "GET",
    path: "/api/admin/outbox/events",
    authRequired: true,
    area: "ops",
    queryFields: [
      { key: "topic", label: "主题", type: "text", defaultValue: "order.paid" },
      { key: "status", label: "状态", type: "select", defaultValue: "dead_letter", options: ["dead_letter", "failed"] },
      { key: "limit", label: "条数", type: "number", defaultValue: 20 },
      { key: "now", label: "查询时间", type: "text", defaultValue: "" }
    ],
    fields: []
  },
  {
    key: "outbox-replay-batch",
    title: "批量恢复 Outbox",
    method: "POST",
    path: "/api/admin/outbox/events/replay",
    authRequired: true,
    area: "ops",
    fields: [
      { key: "topic", label: "主题", type: "text", defaultValue: "order.paid", required: true },
      { key: "limit", label: "条数", type: "number", defaultValue: 10 }
    ]
  },
  {
    key: "outbox-claim-events",
    title: "领取 Outbox 租约",
    method: "POST",
    path: "/api/admin/outbox/events/claim",
    authRequired: true,
    area: "ops",
    fields: [
      { key: "topic", label: "主题", type: "text", defaultValue: "order.paid", required: true },
      { key: "limit", label: "条数", type: "number", defaultValue: 10 },
      { key: "lease_owner", label: "租约持有人", type: "text", defaultValue: "relay-admin", required: true },
      { key: "lease_seconds", label: "租约秒", type: "number", defaultValue: 60 },
      { key: "now", label: "领取时间", type: "text", defaultValue: "" }
    ]
  },
  {
    key: "outbox-replay-event",
    title: "恢复单个 Outbox",
    method: "POST",
    path: "/api/admin/outbox/events/:event_id/replay",
    authRequired: true,
    area: "ops",
    pathFields: [
      { key: "event_id", label: "事件 ID", type: "text", defaultValue: "obe_1", required: true }
    ],
    fields: [
      { key: "now", label: "恢复时间", type: "text", defaultValue: "" }
    ]
  },
  {
    key: "outbox-release-dead-letter",
    title: "解封 Outbox 死信",
    method: "POST",
    path: "/api/admin/outbox/events/:event_id/replay",
    authRequired: true,
    area: "ops",
    pathFields: [
      { key: "event_id", label: "事件 ID", type: "text", defaultValue: "obe_dead_1", required: true }
    ],
    fields: [
      { key: "now", label: "解封时间", type: "text", defaultValue: "" }
    ]
  },
  {
    key: "outbox-renew-lease",
    title: "续租 Outbox 租约",
    method: "POST",
    path: "/api/admin/outbox/events/:event_id/lease/renew",
    authRequired: true,
    area: "ops",
    pathFields: [
      { key: "event_id", label: "事件 ID", type: "text", defaultValue: "obe_1", required: true }
    ],
    fields: [
      { key: "lease_owner", label: "租约持有人", type: "text", defaultValue: "relay-admin", required: true },
      { key: "lease_seconds", label: "租约秒", type: "number", defaultValue: 60 },
      { key: "now", label: "续租时间", type: "text", defaultValue: "" }
    ]
  },
  {
    key: "outbox-mark-failed",
    title: "标记 Outbox 失败",
    method: "POST",
    path: "/api/admin/outbox/events/:event_id/failed",
    authRequired: true,
    area: "ops",
    pathFields: [
      { key: "event_id", label: "事件 ID", type: "text", defaultValue: "obe_1", required: true }
    ],
    fields: [
      { key: "error", label: "失败原因", type: "text", defaultValue: "relay down", required: true },
      { key: "retry_after_seconds", label: "重试延迟秒", type: "number", defaultValue: 120 },
      { key: "max_attempts", label: "最大尝试次数", type: "number", defaultValue: 10 },
      { key: "now", label: "失败时间", type: "text", defaultValue: "" }
    ]
  },
  {
    key: "outbox-mark-published",
    title: "标记 Outbox 已发布",
    method: "POST",
    path: "/api/admin/outbox/events/:event_id/published",
    authRequired: true,
    area: "ops",
    pathFields: [
      { key: "event_id", label: "事件 ID", type: "text", defaultValue: "obe_1", required: true }
    ],
    fields: [
      { key: "published_at", label: "发布时间", type: "text", defaultValue: "" }
    ]
  },
  {
    key: "order-compensate",
    title: "订单状态补偿",
    method: "POST",
    path: "/api/admin/orders/:order_id/state/compensate",
    authRequired: true,
    area: "order",
    pathFields: [
      { key: "order_id", label: "订单号", type: "text", defaultValue: "ord_1", required: true }
    ],
    fields: []
  },
  {
    key: "dispatch-order-events",
    title: "订单派单事件",
    method: "GET",
    path: "/api/dispatch/orders/:order_id/events",
    authRequired: true,
    area: "dispatch",
    pathFields: [
      { key: "order_id", label: "订单号", type: "text", defaultValue: "ord_1", required: true }
    ],
    queryFields: [
      { key: "station_manager_id", label: "站长范围", type: "text", defaultValue: "" }
    ],
    fields: []
  },
  {
    key: "station-riders",
    title: "站点骑手列表",
    method: "GET",
    path: "/api/station-manager/riders",
    authRequired: true,
    area: "rider",
    fields: []
  },
  {
    key: "station-orders",
    title: "站点待调度订单",
    method: "GET",
    path: "/api/station-manager/orders",
    authRequired: true,
    area: "dispatch",
    fields: []
  },
  {
    key: "station-performance",
    title: "站点骑手绩效",
    method: "GET",
    path: "/api/station-manager/rider-performance",
    authRequired: true,
    area: "rider",
    fields: []
  },
  {
    key: "station-task-config",
    title: "站点任务配置",
    method: "GET",
    path: "/api/station-manager/task-duration",
    authRequired: true,
    area: "dispatch",
    fields: []
  }
]);

export function getAdminOperation(key) {
  return ADMIN_API_OPERATIONS.find((operation) => operation.key === key) || null;
}

export function fieldsForOperation(operation) {
  return [
    ...(operation.pathFields || []),
    ...(operation.queryFields || []),
    ...(operation.fields || [])
  ];
}

function normalizeValue(field, value) {
  if (field.type === "csv") {
    const rawItems = Array.isArray(value) ? value : String(value ?? "").split(",");
    const seen = new Set();
    return rawItems.map((item) => String(item).trim()).filter((item) => {
      if (!item || seen.has(item)) {
        return false;
      }
      seen.add(item);
      return true;
    });
  }
  if (field.type === "json") {
    if (value && typeof value === "object" && !Array.isArray(value)) {
      return value;
    }
    const normalized = String(value ?? "").trim();
    if (!normalized) {
      return undefined;
    }
    try {
      return JSON.parse(normalized);
    } catch {
      throw new Error(`${field.label}必须是有效 JSON`);
    }
  }
  if (field.type === "number") {
    const numberValue = Number(value);
    return Number.isFinite(numberValue) ? numberValue : undefined;
  }
  if (field.type === "boolean") {
    if (typeof value === "boolean") {
      return value;
    }
    const normalized = String(value ?? "").trim().toLowerCase();
    if (!normalized) {
      return undefined;
    }
    return normalized === "true" || normalized === "1" || normalized === "yes";
  }
  if (typeof value === "string") {
    return value.trim();
  }
  return value;
}

function isEmptyValue(value) {
  if (Array.isArray(value)) {
    return value.length === 0;
  }
  return value === undefined || value === "";
}

export function buildAdminRequest(operation, values = {}, token = "") {
  if (!operation) {
    throw new Error("unknown admin operation");
  }
  let path = operation.path;
  for (const field of operation.pathFields || []) {
    const value = normalizeValue(field, values[field.key] ?? field.defaultValue ?? "");
    if (field.required && !value) {
      throw new Error(`${field.label}不能为空`);
    }
    path = path.replace(`:${field.key}`, encodeURIComponent(String(value)));
  }

  const query = new URLSearchParams();
  for (const field of operation.queryFields || []) {
    const value = normalizeValue(field, values[field.key] ?? field.defaultValue ?? "");
    if (value !== undefined && value !== "") {
      query.set(field.key, String(value));
    }
  }
  const body = {};
  for (const field of operation.fields || []) {
    const value = normalizeValue(field, values[field.key] ?? field.defaultValue ?? "");
    if (field.required && isEmptyValue(value)) {
      throw new Error(`${field.label}不能为空`);
    }
    if (!isEmptyValue(value)) {
      body[field.key] = value;
    }
  }

  const headers = { "Content-Type": "application/json", "X-Client-Kind": "admin-web" };
  if (operation.authRequired && token) {
    headers.Authorization = token.startsWith("Bearer ") ? token : `Bearer ${token}`;
  }

  const url = `${path}${query.toString() ? `?${query.toString()}` : ""}`;
  return {
    method: operation.method,
    url,
    headers,
    body: operation.method === "GET" ? undefined : JSON.stringify(body)
  };
}

export async function executeAdminOperation({ baseUrl = DEFAULT_BFF_BASE_URL, token = "", operationKey, values = {}, fetchImpl = fetch }) {
  const operation = getAdminOperation(operationKey);
  const request = buildAdminRequest(operation, values, token);
  const response = await fetchImpl(`${baseUrl.replace(/\/$/, "")}${request.url}`, request);
  const text = await response.text();
  let payload;
  try {
    payload = text ? JSON.parse(text) : null;
  } catch {
    payload = { raw: text };
  }
  return {
    ok: response.ok,
    status: response.status,
    operation,
    request: { method: request.method, url: request.url },
    payload
  };
}
