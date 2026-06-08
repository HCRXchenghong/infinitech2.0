import assert from "node:assert/strict";
import test from "node:test";
import { previewAdminResult } from "./adminResultPreview.mjs";

test("previewAdminResult builds after-sales timeline cards", () => {
  const preview = previewAdminResult({
    ok: true,
    operation: { key: "after-sales-events" },
    request: { url: "/api/after-sales/asr_1/events" },
    payload: {
      success: true,
      data: [
        {
          id: "ase_1",
          request_id: "asr_1",
          actor_id: "user_1",
          actor_role: "user",
          action: "created",
          message: "用户提交售后并上传首张图片",
          visible_to_user: true,
          attachments: ["https://cdn.test/after-sales/asr_1/p1.jpg"],
          created_at: "2026-06-03T01:05:00Z"
        },
        {
          id: "ase_2",
          request_id: "asr_1",
          actor_id: "admin_1",
          actor_role: "admin",
          action: "internal_note",
          message: "客服内部备注",
          visible_to_user: false,
          attachments: [],
          created_at: "2026-06-03T01:15:00Z"
        }
      ]
    }
  });

  assert.ok(preview);
  assert.equal(preview.title, "售后时间线");
  assert.equal(preview.subtitle, "工单 asr_1 · 2 条事件");
  assert.equal(preview.stats[0].value, "2 条");
  assert.equal(preview.stats[1].value, "1 条");
  assert.equal(preview.stats[2].value, "1 条");
  assert.equal(preview.items[0].title, "用户提交售后");
  assert.equal(preview.items[0].chips[0], "附件 1");
  assert.equal(preview.items[0].links[0].href, "https://cdn.test/after-sales/asr_1/p1.jpg");
  assert.equal(preview.items[1].badge, "内部备注");
  assert.equal(preview.items[1].tone, "slate");
});

test("previewAdminResult builds after-sales evidence cards", () => {
  const preview = previewAdminResult({
    ok: true,
    operation: { key: "after-sales-evidence" },
    request: { url: "/api/after-sales/asr_2/evidence" },
    payload: {
      success: true,
      data: [
        {
          id: "asev_1",
          request_id: "asr_2",
          file_name: "leak.jpg",
          public_url: "https://cdn.test/after-sales/asr_2/leak.jpg",
          object_key: "after-sales/asr_2/leak.jpg",
          content_type: "image/jpeg",
          size_bytes: 524288,
          uploaded_by_role: "user",
          uploaded_by_id: "user_9",
          status: "uploaded",
          content_sha: "sha256:1234567890abcdef1234567890abcdef",
          created_at: "2026-06-03T01:05:00Z",
          confirmed_at: "2026-06-03T01:06:00Z"
        },
        {
          id: "asev_2",
          request_id: "asr_2",
          file_name: "receipt.pdf",
          public_url: "https://cdn.test/after-sales/asr_2/receipt.pdf",
          object_key: "after-sales/asr_2/receipt.pdf",
          content_type: "application/pdf",
          size_bytes: 1048576,
          uploaded_by_role: "merchant",
          uploaded_by_id: "merchant_1",
          status: "uploaded",
          created_at: "2026-06-03T01:08:00Z",
          confirmed_at: "2026-06-03T01:09:00Z"
        }
      ]
    }
  });

  assert.ok(preview);
  assert.equal(preview.title, "售后凭证");
  assert.equal(preview.subtitle, "工单 asr_2 · 2 份凭证");
  assert.equal(preview.stats[0].value, "2 份");
  assert.equal(preview.stats[1].value, "1 张");
  assert.equal(preview.stats[2].value, "1.5 MB");
  assert.equal(preview.items[0].previewImageUrl, "https://cdn.test/after-sales/asr_2/leak.jpg");
  assert.equal(preview.items[0].links[0].label, "打开凭证");
  assert.ok(preview.items[0].chips.some((chip) => chip.startsWith("SHA ")));
  assert.equal(preview.items[1].previewImageUrl, "");
});

test("previewAdminResult builds dispatch event cards", () => {
  const preview = previewAdminResult({
    ok: true,
    operation: { key: "dispatch-order-events" },
    request: { url: "/api/dispatch/orders/ord_9/events" },
    payload: {
      success: true,
      data: [
        {
          id: "dpe_1",
          order_id: "ord_9",
          station_id: "station_1",
          mode: "auto_assign",
          type: "dispatch.auto_assign",
          rider_id: "rider_1",
          idempotency_key: "dispatch:auto:1",
          online_candidate_size: 3,
          rejected_rider_ids: [],
          can_decline_without_penalty: false,
          created_at: "2026-06-03T01:10:00Z"
        },
        {
          id: "dpe_2",
          order_id: "ord_9",
          station_id: "station_1",
          mode: "auto_assign",
          type: "dispatch.timeout",
          rider_id: "rider_1",
          reason: "assignment_timeout",
          idempotency_key: "dispatch:timeout:1",
          online_candidate_size: 2,
          rejected_rider_ids: ["rider_1"],
          can_decline_without_penalty: true,
          created_at: "2026-06-03T01:15:00Z"
        }
      ]
    }
  });

  assert.ok(preview);
  assert.equal(preview.title, "订单派单事件");
  assert.equal(preview.subtitle, "订单 ord_9 · 2 条派单记录");
  assert.equal(preview.stats[1].value, "2 条");
  assert.equal(preview.stats[3].value, "1 次");
  assert.equal(preview.items[0].title, "自动派单");
  assert.equal(preview.items[1].tone, "amber");
  assert.ok(preview.items[1].chips.includes("免责拒派"));
  assert.ok(preview.items[1].chips.includes("已拒 1"));
});

test("previewAdminResult builds refund transaction cards", () => {
  const preview = previewAdminResult({
    ok: true,
    operation: { key: "refund-transactions" },
    request: { url: "/api/admin/refunds?order_id=ord_9" },
    payload: {
      success: true,
      data: [
        {
          id: "rfd_9",
          order_id: "ord_9",
          user_id: "user_9",
          amount_fen: 800,
          destination: "balance",
          status: "success",
          reason: "漏送补退",
          idempotency_key: "after_sales:asr_9",
          created_at: "2026-06-03T01:28:00Z"
        },
        {
          id: "rfd_8",
          order_id: "ord_9",
          user_id: "user_9",
          amount_fen: 1200,
          destination: "original_route",
          status: "pending",
          reason: "后台指定原路退",
          idempotency_key: "manual:ord_9",
          created_at: "2026-06-03T01:18:00Z"
        }
      ]
    }
  });

  assert.ok(preview);
  assert.equal(preview.title, "退款流水");
  assert.equal(preview.subtitle, "订单 ord_9 · 2 笔退款");
  assert.equal(preview.stats[1].value, "¥20.00");
  assert.equal(preview.items[0].body, "¥8.00 · 退平台余额");
  assert.equal(preview.items[1].badge, "处理中");
  assert.ok(preview.items[0].chips.some((chip) => chip.includes("after_sales:asr_9")));
  assert.equal(preview.items[0].actions[0].operationKey, "audit-logs");
  assert.equal(preview.items[0].actions[0].values.target_id, "ord_9");
  assert.equal(preview.items[0].actions[0].values.action, "admin.order.refunded");
  assert.equal(preview.items[0].actions[1].operationKey, "order-detail");
});

test("previewAdminResult builds aggregated order detail cards", () => {
  const preview = previewAdminResult({
    ok: true,
    operation: { key: "order-detail" },
    request: { url: "/api/admin/orders/ord_9" },
    payload: {
      success: true,
      data: {
        order: {
          id: "ord_9",
          user_id: "user_9",
          shop_name: "蓝海餐厅",
          reviewed: true,
          address_snapshot: {
            detail: "望京SOHO"
          },
          type: "takeout",
          status: "dispatching",
          amount_fen: 1800,
          payment_method: "balance",
          rider_id: "rider_1",
          created_at: "2026-06-03T01:05:00Z",
          items: [
            {
              product_name: "招牌牛肉饭",
              quantity: 1
            }
          ]
        },
        after_sales_summary: {
          total: 1,
          open_count: 1,
          refunded_count: 0,
          latest_status: "admin_review",
          latest_updated_at: "2026-06-03T01:25:00Z"
        },
        refund_summary: {
          total: 1,
          success_count: 1,
          total_amount_fen: 800,
          latest_destination: "balance",
          latest_created_at: "2026-06-03T01:28:00Z"
        },
        service_ticket_summary: {
          total: 1,
          open_count: 1,
          escalated_count: 0,
          latest_status: "processing",
          latest_updated_at: "2026-06-03T01:27:00Z"
        },
        dispatch_summary: {
          total: 2,
          auto_assign_count: 2,
          manual_assign_count: 0,
          reject_count: 0,
          timeout_count: 1,
          latest_type: "dispatch.timeout",
          latest_event_at: "2026-06-03T01:15:00Z"
        },
        audit_summary: {
          total: 3,
          verified_count: 3,
          order_count: 1,
          after_sales_count: 1,
          service_ticket_count: 1,
          latest_action: "admin.order.refunded",
          latest_created_at: "2026-06-03T01:29:00Z"
        },
        after_sales_requests: [
          {
            id: "asr_9",
            status: "admin_review",
            reason: "餐品漏送",
            requested_amount_fen: 800,
            latest_event_message: "平台已介入核实"
          }
        ],
        refunds: [
          {
            id: "rfd_9",
            amount_fen: 800,
            destination: "balance",
            status: "success",
            created_at: "2026-06-03T01:28:00Z"
          }
        ],
        service_tickets: [
          {
            id: "st_9",
            status: "processing",
            title: "配送问题 · 预计送达未更新",
            assigned_support_name: "客服小悦"
          }
        ],
        related_audits: [
          {
            id: "aud_3",
            action: "admin.order.refunded",
            target_type: "order",
            target_id: "ord_9",
            request_id: "req_3",
            created_at: "2026-06-03T01:29:00Z"
          }
        ]
      }
    }
  });

  assert.ok(preview);
  assert.equal(preview.title, "订单聚合详情");
  assert.equal(preview.subtitle, "订单 ord_9 · 蓝海餐厅");
  assert.equal(preview.stats[0].value, "待配送");
  assert.equal(preview.items[0].chips[0], "支付 余额支付");
  assert.equal(preview.items[1].body, "平台仲裁 · 餐品漏送");
  assert.equal(preview.items[2].body, "最近退款 ¥8.00，退平台余额。");
  assert.equal(preview.items[3].body, "处理中 · 配送问题 · 预计送达未更新");
  assert.equal(preview.items[4].body, "拒单 0 次，超时 1 次。");
  assert.equal(preview.items[5].body, "最近审计 订单退款，目标 order:ord_9");
  assert.equal(preview.items[0].actions[0].operationKey, "refund-transactions");
  assert.equal(preview.items[1].actions[0].operationKey, "after-sales-detail");
  assert.equal(preview.items[3].actions[0].operationKey, "support-ticket-detail");
  assert.equal(preview.items[5].actions[0].operationKey, "audit-logs");
});

test("previewAdminResult builds aggregated after-sales detail cards", () => {
  const preview = previewAdminResult({
    ok: true,
    operation: { key: "after-sales-detail" },
    request: { url: "/api/admin/after-sales/asr_9" },
    payload: {
      success: true,
      data: {
        request: {
          id: "asr_9",
          order_id: "ord_9",
          user_id: "user_9",
          shop_name: "蓝海餐厅",
          order_status: "merchant_pending",
          order_item_summary: "招牌牛肉饭 x 1",
          latest_event_message: "平台已介入核实",
          latest_event_at: "2026-06-03T01:25:00Z",
          reason: "餐品漏送",
          requested_amount_fen: 800,
          refunded_amount_fen: 0,
          refundable_fen: 1800,
          status: "admin_review"
        },
        event_summary: {
          total: 4,
          user_visible: 3,
          internal_only: 1,
          attachment_count: 2,
          latest_action: "internal_note",
          latest_event_at: "2026-06-03T01:26:00Z"
        },
        evidence_summary: {
          total: 2,
          image_count: 1,
          confirmed_count: 2,
          total_size_bytes: 1572864,
          latest_confirmed_at: "2026-06-03T01:18:00Z"
        },
        dispatch_summary: {
          total: 2,
          auto_assign_count: 2,
          manual_assign_count: 0,
          reject_count: 0,
          timeout_count: 1,
          latest_type: "dispatch.timeout",
          latest_event_at: "2026-06-03T01:15:00Z"
        },
        refund_summary: {
          total: 1,
          success_count: 1,
          total_amount_fen: 800,
          latest_destination: "balance",
          latest_created_at: "2026-06-03T01:28:00Z"
        },
        service_ticket_summary: {
          total: 1,
          open_count: 1,
          escalated_count: 0,
          latest_status: "processing",
          latest_updated_at: "2026-06-03T01:27:00Z"
        },
        audit_summary: {
          total: 3,
          verified_count: 3,
          order_count: 1,
          after_sales_count: 1,
          service_ticket_count: 1,
          latest_action: "admin.order.refunded",
          latest_created_at: "2026-06-03T01:29:00Z"
        },
        refunds: [
          {
            id: "rfd_9",
            amount_fen: 800,
            destination: "balance",
            status: "success",
            created_at: "2026-06-03T01:28:00Z"
          }
        ],
        service_tickets: [
          {
            id: "st_9",
            status: "processing",
            title: "配送问题 · 预计送达未更新",
            assigned_support_name: "客服小悦"
          }
        ],
        related_audits: [
          {
            id: "aud_3",
            action: "admin.order.refunded",
            target_type: "order",
            target_id: "ord_9",
            request_id: "req_3",
            created_at: "2026-06-03T01:29:00Z"
          }
        ],
        evidence: [
          {
            file_name: "evidence.jpg",
            public_url: "https://cdn.test/after-sales/asr_9/evidence.jpg",
            content_type: "image/jpeg"
          }
        ]
      }
    }
  });

  assert.ok(preview);
  assert.equal(preview.title, "售后聚合详情");
  assert.equal(preview.subtitle, "工单 asr_9 · 订单 ord_9");
  assert.equal(preview.stats[0].value, "平台仲裁");
  assert.equal(preview.items[0].chips[0], "申请 ¥8.00");
  assert.equal(preview.items[1].body, "用户可见 3 条，内部备注 1 条。");
  assert.equal(preview.items[2].previewImageUrl, "https://cdn.test/after-sales/asr_9/evidence.jpg");
  assert.equal(preview.items[3].body, "拒单 0 次，超时 1 次。");
  assert.equal(preview.items[4].body, "最近退款 ¥8.00，退平台余额。");
  assert.equal(preview.items[5].body, "处理中 · 配送问题 · 预计送达未更新");
  assert.equal(preview.items[6].body, "最近审计 订单退款，目标 order:ord_9");
  assert.equal(preview.items[0].actions[0].operationKey, "after-sales-list");
  assert.equal(preview.items[0].actions[1].operationKey, "order-detail");
  assert.equal(preview.items[0].actions[0].values.request_id, "asr_9");
  assert.equal(preview.items[1].actions[0].operationKey, "after-sales-events");
  assert.equal(preview.items[2].actions[0].operationKey, "after-sales-evidence");
  assert.equal(preview.items[3].actions[0].operationKey, "dispatch-order-events");
  assert.equal(preview.items[3].actions[0].values.order_id, "ord_9");
  assert.equal(preview.items[4].actions[0].operationKey, "refund-transactions");
  assert.equal(preview.items[5].actions[0].operationKey, "support-ticket-detail");
  assert.equal(preview.items[5].actions[0].values.ticket_id, "st_9");
  assert.equal(preview.items[6].actions[0].operationKey, "audit-logs");
  assert.equal(preview.items[6].actions[1].values.target_type, "after_sales");
});

test("previewAdminResult builds support ticket detail cards", () => {
  const preview = previewAdminResult({
    ok: true,
    operation: { key: "support-ticket-detail" },
    request: { url: "/api/service-tickets/st_9?user_id=user_9" },
    payload: {
      success: true,
      data: {
        ticket: {
          id: "st_9",
          user_id: "user_9",
          related_order_id: "ord_9",
          category: "配送问题",
          title: "配送问题 · 预计送达未更新",
          content: "骑手到店很久了",
          status: "processing",
          sla_status: "overdue",
          assigned_support_name: "客服小悦",
          severity: "high",
          reply_due_at: "2026-06-03T01:20:00Z",
          assigned_at: "2026-06-03T01:08:00Z",
          updated_at: "2026-06-03T01:27:00Z"
        },
        events: [
          {
            id: "ste_1",
            actor_id: "system",
            actor_role: "system",
            title: "已提交",
            message: "问题已同步到客服工单",
            status: "done",
            attachments: [],
            created_at: "2026-06-03T01:05:00Z"
          },
          {
            id: "ste_2",
            actor_id: "support_1",
            actor_role: "support",
            title: "处理方案",
            message: "已发放 5 元延误券，请确认处理结果",
            status: "active",
            attachments: ["https://cdn.test/support/st_9/solution.jpg"],
            created_at: "2026-06-03T01:27:00Z"
          }
        ]
      }
    }
  });

  assert.ok(preview);
  assert.equal(preview.title, "客服工单详情");
  assert.equal(preview.subtitle, "工单 st_9 · 订单 ord_9");
  assert.equal(preview.stats[0].value, "处理中");
  assert.equal(preview.stats[1].value, "overdue");
  assert.equal(preview.items[0].chips[0], "配送问题");
  assert.equal(preview.items[1].body, "当前由 客服小悦 跟进。");
  assert.equal(preview.items[2].body, "已发放 5 元延误券，请确认处理结果");
  assert.equal(preview.items[2].links[0].href, "https://cdn.test/support/st_9/solution.jpg");
  assert.equal(preview.items[3].body, "首条 已提交，最新 处理方案。");
});

test("previewAdminResult ignores non-previewable results", () => {
  assert.equal(previewAdminResult({
    ok: true,
    operation: { key: "refund-settings-read" },
    payload: { success: true, data: { default_refund_strategy: "balance_first" } }
  }), null);

  assert.equal(previewAdminResult({
    ok: false,
    operation: { key: "after-sales-events" },
    payload: { success: false }
  }), null);
});
