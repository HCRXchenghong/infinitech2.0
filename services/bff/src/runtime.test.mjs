import assert from "node:assert/strict";
import { once } from "node:events";
import http from "node:http";
import test from "node:test";
import { createRuntimeConfig, defaultHomeCards, defaultHomeModules } from "./runtime.mjs";
import { createBffServer } from "./server.mjs";

test("runtime config exposes all client kinds and brand", () => {
  const config = createRuntimeConfig({ API_BASE_URL: "http://api", REALTIME_URL: "ws://rt/ws" });
  assert.deepEqual(config.clientKinds, ["wechat-miniprogram", "merchant-uni", "rider-uni", "admin-web", "admin-uni"]);
  assert.equal(config.brand.primaryColor, "#009bf5");
  assert.equal(config.apiBaseUrl, "http://api");
  assert.equal(config.homeCardsMode, "admin_configurable");
});

test("home modules expose circle and keep social as configurable disabled module", () => {
  const modules = defaultHomeModules();
  const enabled = modules.filter((item) => item.enabled);
  assert.deepEqual(enabled.map((item) => item.key), ["takeout", "groupbuy", "medicine", "courier", "circle", "meal-match", "coupons", "points"]);
  assert.ok(enabled.every((item) => item.icon_url?.startsWith("/assets/generated/category-")));
  assert.equal(modules.find((item) => item.key === "charity").enabled, false);
  assert.equal(modules.find((item) => item.key === "social").enabled, false);
});

test("home cards are explicitly admin configurable", () => {
  const cards = defaultHomeCards();
  assert.equal(cards[0].type, "product");
  assert.equal(cards[1].type, "circle_post");
  assert.ok(cards.every((item) => item.enabled));
});

test("bff server answers runtime endpoint", async () => {
  const server = createBffServer({ env: { API_BASE_URL: "http://api" } });
  server.listen(0);
  await once(server, "listening");
  const { port } = server.address();
  const body = await getJSON(`http://127.0.0.1:${port}/api/runtime`);
  const cards = await getJSON(`http://127.0.0.1:${port}/api/home/cards`);
  server.close();
  assert.equal(body.success, true);
  assert.equal(body.data.apiBaseUrl, "http://api");
  assert.equal(cards.success, true);
  assert.equal(cards.data[0].type, "product");
});

test("bff proxies public search catalog route with query intact", async () => {
  const upstream = http.createServer((req, res) => {
    if (req.method === "GET" && req.url === "/api/search?keyword=%E7%89%9B%E8%82%89%E9%A5%AD&category=product") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({
        success: true,
        data: {
          keyword: "牛肉饭",
          category: "product",
          suggestions: ["招牌牛肉饭"],
          results: [{ id: "prod_beef_rice", type: "product", title: "招牌牛肉饭" }],
          authorization: req.headers.authorization || ""
        }
      }));
      return;
    }
    res.writeHead(404, { "Content-Type": "application/json" });
    res.end(JSON.stringify({ success: false }));
  });
  upstream.listen(0);
  await once(upstream, "listening");

  const server = createBffServer({ env: { API_BASE_URL: `http://127.0.0.1:${upstream.address().port}` } });
  server.listen(0);
  await once(server, "listening");
  const body = await getJSON(`http://127.0.0.1:${server.address().port}/api/search?keyword=%E7%89%9B%E8%82%89%E9%A5%AD&category=product`);

  server.close();
  upstream.close();

  assert.equal(body.success, true);
  assert.equal(body.data.keyword, "牛肉饭");
  assert.equal(body.data.category, "product");
  assert.equal(body.data.authorization, "");
  assert.equal(body.data.results[0].id, "prod_beef_rice");
  assert.deepEqual(body.data.suggestions, ["招牌牛肉饭"]);
});

test("bff allows configured browser origins for admin api preflight and proxy responses", async () => {
  const upstream = http.createServer((req, res) => {
    if (req.method === "GET" && req.url === "/api/admin/operations/snapshot?limit=1") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({
        success: true,
        data: {
          generated_at: "2026-05-22T12:00:00Z",
          counts: { total_orders: 1 },
          authorization: req.headers.authorization
        }
      }));
      return;
    }
    res.writeHead(404, { "Content-Type": "application/json" });
    res.end(JSON.stringify({ success: false }));
  });
  upstream.listen(0);
  await once(upstream, "listening");

  const server = createBffServer({
    env: {
      API_BASE_URL: `http://127.0.0.1:${upstream.address().port}`,
      BFF_ALLOWED_ORIGINS: "http://admin.test"
    }
  });
  server.listen(0);
  await once(server, "listening");
  const baseUrl = `http://127.0.0.1:${server.address().port}`;

  const preflight = await requestRaw("OPTIONS", `${baseUrl}/api/admin/operations/snapshot`, {
    Origin: "http://admin.test",
    "Access-Control-Request-Method": "GET",
    "Access-Control-Request-Headers": "Authorization,Content-Type"
  });
  const snapshot = await requestRaw("GET", `${baseUrl}/api/admin/operations/snapshot?limit=1`, {
    Origin: "http://admin.test",
    Authorization: "Bearer admin:admin_1"
  });
  const blocked = await requestRaw("GET", `${baseUrl}/api/runtime`, {
    Origin: "http://evil.test"
  });

  server.close();
  upstream.close();

  assert.equal(preflight.statusCode, 204);
  assert.equal(preflight.headers["access-control-allow-origin"], "http://admin.test");
  assert.match(preflight.headers["access-control-allow-methods"], /GET/);
  assert.match(preflight.headers["access-control-allow-headers"], /Authorization/);
  assert.equal(snapshot.statusCode, 200);
  assert.equal(snapshot.headers["access-control-allow-origin"], "http://admin.test");
  assert.equal(JSON.parse(snapshot.body).data.authorization, "Bearer admin:admin_1");
  assert.equal(blocked.statusCode, 403);
  assert.equal(blocked.headers["access-control-allow-origin"], undefined);
});

test("bff proxies prescription image upload routes", async () => {
  const upstream = http.createServer((req, res) => {
    if (req.method === "POST" && req.url === "/api/prescriptions/upload-ticket") {
      let body = "";
      req.on("data", (chunk) => {
        body += chunk;
      });
      req.on("end", () => {
        res.writeHead(201, { "Content-Type": "application/json" });
        res.end(JSON.stringify({
          success: true,
          data: {
            ticket_id: "rxu_1",
            object_key: "prescriptions/user/sig/prescription.jpg",
            upload_url: "https://object-storage.infinitech.local/upload/prescriptions/user/sig/prescription.jpg",
            public_url: "https://cdn.infinitech.local/prescriptions/user/sig/prescription.jpg",
            method: "PUT",
            request: JSON.parse(body),
            authorization: req.headers.authorization
          }
        }));
      });
      return;
    }
    if (req.method === "POST" && req.url === "/api/prescriptions/upload-confirm") {
      let body = "";
      req.on("data", (chunk) => {
        body += chunk;
      });
      req.on("end", () => {
        res.writeHead(201, { "Content-Type": "application/json" });
        res.end(JSON.stringify({
          success: true,
          data: {
            id: "rxu_1",
            status: "confirmed",
            object_key: JSON.parse(body).object_key,
            authorization: req.headers.authorization
          }
        }));
      });
      return;
    }
    if (req.method === "POST" && req.url === "/api/object-storage/upload-callback") {
      let body = "";
      req.on("data", (chunk) => {
        body += chunk;
      });
      req.on("end", () => {
        res.writeHead(200, { "Content-Type": "application/json" });
        res.end(JSON.stringify({
          success: true,
          data: {
            id: "rxu_1",
            status: "uploaded",
            scan_status: "pending",
            object_key: JSON.parse(body).object_key
          }
        }));
      });
      return;
    }
    if (req.method === "POST" && req.url === "/api/object-storage/scan-result") {
      let body = "";
      req.on("data", (chunk) => {
        body += chunk;
      });
      req.on("end", () => {
        res.writeHead(200, { "Content-Type": "application/json" });
        res.end(JSON.stringify({
          success: true,
          data: {
            id: "rxu_1",
            status: "uploaded",
            scan_status: JSON.parse(body).scan_status,
            scan_result: "clean",
            object_key: JSON.parse(body).object_key
          }
        }));
      });
      return;
    }
    res.writeHead(404, { "Content-Type": "application/json" });
    res.end(JSON.stringify({ success: false }));
  });
  upstream.listen(0);
  await once(upstream, "listening");
  const server = createBffServer({ env: { API_BASE_URL: `http://127.0.0.1:${upstream.address().port}` } });
  server.listen(0);
  await once(server, "listening");
  const baseUrl = `http://127.0.0.1:${server.address().port}`;

  const ticket = await postJSON(`${baseUrl}/api/prescriptions/upload-ticket`, "Bearer user:user_1", { product_id: "med_amoxicillin", file_name: "prescription.jpg", content_type: "image/jpeg", size_bytes: 2048 });
  const uploaded = await postJSON(`${baseUrl}/api/object-storage/upload-callback`, "", { ticket_id: "rxu_1", object_key: ticket.data.object_key, content_type: "image/jpeg", size_bytes: 2048, content_sha: "sha256:rx" });
  const scanned = await postJSON(`${baseUrl}/api/object-storage/scan-result`, "", { ticket_id: "rxu_1", object_key: ticket.data.object_key, scan_status: "passed" });
  const confirmed = await postJSON(`${baseUrl}/api/prescriptions/upload-confirm`, "Bearer user:user_1", { ticket_id: "rxu_1", object_key: ticket.data.object_key, content_type: "image/jpeg", size_bytes: 2048 });

  server.close();
  upstream.close();

  assert.equal(ticket.data.authorization, "Bearer user:user_1");
  assert.equal(ticket.data.request.file_name, "prescription.jpg");
  assert.equal(uploaded.data.scan_status, "pending");
  assert.equal(scanned.data.scan_status, "passed");
  assert.equal(confirmed.data.status, "confirmed");
  assert.equal(confirmed.data.object_key, "prescriptions/user/sig/prescription.jpg");
});

test("bff proxies prescription review workbench routes", async () => {
  const upstream = http.createServer((req, res) => {
    if (req.method === "GET" && req.url === "/api/admin/prescriptions?status=approved") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ success: true, data: [{ id: "rx_1", authorization: req.headers.authorization }] }));
      return;
    }
    if (req.method === "POST" && req.url === "/api/admin/prescriptions/rx_1/review") {
      let body = "";
      req.on("data", (chunk) => {
        body += chunk;
      });
      req.on("end", () => {
        res.writeHead(200, { "Content-Type": "application/json" });
        res.end(JSON.stringify({
          success: true,
          data: {
            id: "rx_1",
            status: JSON.parse(body).decision,
            authorization: req.headers.authorization
          }
        }));
      });
      return;
    }
    res.writeHead(404, { "Content-Type": "application/json" });
    res.end(JSON.stringify({ success: false }));
  });
  upstream.listen(0);
  await once(upstream, "listening");
  const server = createBffServer({ env: { API_BASE_URL: `http://127.0.0.1:${upstream.address().port}` } });
  server.listen(0);
  await once(server, "listening");
  const baseUrl = `http://127.0.0.1:${server.address().port}`;

  const queue = await getJSON(`${baseUrl}/api/admin/prescriptions?status=approved`, "Bearer support_admin:support_1");
  const reviewed = await postJSON(`${baseUrl}/api/admin/prescriptions/rx_1/review`, "Bearer support_admin:support_1", { decision: "rejected" });

  server.close();
  upstream.close();

  assert.equal(queue.data[0].authorization, "Bearer support_admin:support_1");
  assert.equal(reviewed.data.status, "rejected");
  assert.equal(reviewed.data.authorization, "Bearer support_admin:support_1");
});

test("bff proxies chat sync and read routes", async () => {
  const upstream = http.createServer((req, res) => {
    if (req.method === "GET" && req.url === "/api/messages/merchant_blue_sea/sync?since_id=msg_1&mark_read=true") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({
        success: true,
        data: {
          thread_id: "merchant_blue_sea",
          messages: [{ id: "msg_2", content: "离线消息", authorization: req.headers.authorization }],
          unread_count: 0,
          next_cursor: "msg_2"
        }
      }));
      return;
    }
    if (req.method === "POST" && req.url === "/api/messages/merchant_blue_sea/read") {
      let body = "";
      req.on("data", (chunk) => {
        body += chunk;
      });
      req.on("end", () => {
        res.writeHead(200, { "Content-Type": "application/json" });
        res.end(JSON.stringify({
          success: true,
          data: { thread_id: "merchant_blue_sea", unread_count: 0, body: JSON.parse(body), authorization: req.headers.authorization }
        }));
      });
      return;
    }
    if (req.method === "POST" && req.url === "/api/messages/merchant_blue_sea") {
      let body = "";
      req.on("data", (chunk) => {
        body += chunk;
      });
      req.on("end", () => {
        res.writeHead(201, { "Content-Type": "application/json" });
        res.end(JSON.stringify({
          success: true,
          data: { id: "msg_3", risk_state: "passed", body: JSON.parse(body), authorization: req.headers.authorization }
        }));
      });
      return;
    }
    res.writeHead(404, { "Content-Type": "application/json" });
    res.end(JSON.stringify({ success: false }));
  });
  upstream.listen(0);
  await once(upstream, "listening");
  const server = createBffServer({ env: { API_BASE_URL: `http://127.0.0.1:${upstream.address().port}` } });
  server.listen(0);
  await once(server, "listening");
  const baseUrl = `http://127.0.0.1:${server.address().port}`;

  const sync = await getJSON(`${baseUrl}/api/messages/merchant_blue_sea/sync?since_id=msg_1&mark_read=true`, "Bearer user:user_1");
  const read = await postJSON(`${baseUrl}/api/messages/merchant_blue_sea/read`, "Bearer user:user_1", { last_message_id: "msg_2" });
  const sent = await postJSON(`${baseUrl}/api/messages/merchant_blue_sea`, "Bearer user:user_1", { content: "请帮我催一下商家" });

  server.close();
  upstream.close();

  assert.equal(sync.data.next_cursor, "msg_2");
  assert.equal(sync.data.messages[0].authorization, "Bearer user:user_1");
  assert.equal(read.data.body.last_message_id, "msg_2");
  assert.equal(read.data.authorization, "Bearer user:user_1");
  assert.equal(sent.data.risk_state, "passed");
  assert.equal(sent.data.body.content, "请帮我催一下商家");
  assert.equal(sent.data.authorization, "Bearer user:user_1");
});

test("bff proxies user-facing api routes with authorization", async () => {
  const upstream = http.createServer((req, res) => {
    if (req.method === "POST" && req.url === "/api/auth/wechat-mini/login") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ success: true, data: { access_token: "signed.token", token_type: "Bearer" } }));
      return;
    }
    if (req.method === "POST" && req.url === "/api/auth/phone/code") {
      let body = "";
      req.on("data", (chunk) => {
        body += chunk;
      });
      req.on("end", () => {
        res.writeHead(200, { "Content-Type": "application/json" });
        res.end(JSON.stringify({ success: true, data: { ...JSON.parse(body), dev_code: "135790" } }));
      });
      return;
    }
    if (req.method === "POST" && req.url === "/api/auth/phone/login") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ success: true, data: { access_token: "phone.login.token", provider: "phone" } }));
      return;
    }
    if (req.method === "POST" && req.url === "/api/auth/phone/register") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ success: true, data: { access_token: "phone.register.token", provider: "phone" } }));
      return;
    }
    if (req.method === "POST" && req.url === "/api/auth/logout") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ success: true, data: { revoked: true, authorization: req.headers.authorization } }));
      return;
    }
    if (req.method === "POST" && req.url === "/api/admin/rider-invites") {
      let body = "";
      req.on("data", (chunk) => {
        body += chunk;
      });
      req.on("end", () => {
        res.writeHead(201, { "Content-Type": "application/json" });
        res.end(JSON.stringify({
          success: true,
          data: {
            token: "ri_admin_station_manager",
            type: JSON.parse(body).type,
            station_id: JSON.parse(body).station_id,
            authorization: req.headers.authorization
          }
        }));
      });
      return;
    }
    if (req.method === "GET" && req.url === "/api/admin/refund-settings") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ success: true, data: { default_refund_strategy: "balance_first", authorization: req.headers.authorization } }));
      return;
    }
    if (req.method === "PUT" && req.url === "/api/admin/refund-settings") {
      let body = "";
      req.on("data", (chunk) => {
        body += chunk;
      });
      req.on("end", () => {
        res.writeHead(200, { "Content-Type": "application/json" });
        res.end(JSON.stringify({ success: true, data: { ...JSON.parse(body), authorization: req.headers.authorization } }));
      });
      return;
    }
    if (req.method === "GET" && req.url === "/api/admin/after-sales") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ success: true, data: [{ id: "asr_1", status: "pending_merchant", authorization: req.headers.authorization }] }));
      return;
    }
    if (req.method === "GET" && req.url === "/api/admin/operations/snapshot?limit=5&lease_expiring_within_seconds=60&object_cleanup_grace_seconds=60") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({
        success: true,
        data: {
          generated_at: "2026-05-22T12:00:00Z",
          counts: { total_orders: 3, after_sales_pending: 1, outbox_ready: 1 },
          orders: [{ id: "ord_1", status: "dispatching" }],
          merchants: [{ account: { id: "merchant_1" }, can_accept_orders: true }],
          riders: [{ id: "rider_1", online: true }],
          rider_performance: [{ rider_id: "rider_1", level: "S" }],
          after_sales: [{ id: "asr_1", status: "pending_merchant" }],
          dispatch_events: [],
          refund_settings: { default_refund_strategy: "balance_first" },
          outbox_stats: { ready: 1, blocked: 0 },
          object_storage_cleanup_stats: { pending: 0 },
          authorization: req.headers.authorization
        }
      }));
      return;
    }
    if (req.method === "GET" && req.url === "/api/admin/audit-logs?target_type=order&limit=1") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({
        success: true,
        data: [{
          id: "aud_1",
          actor_type: "admin",
          actor_id: "admin_1",
          action: "admin.order.refunded",
          target_type: "order",
          target_id: "ord_1",
          authorization: req.headers.authorization
        }]
      }));
      return;
    }
    if (req.method === "GET" && req.url === "/api/admin/audit-logs/export?target_type=order&limit=1") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({
        success: true,
        data: {
          format: "csv",
          filename: "audit-logs-20260522T120000Z.csv",
          row_count: 1,
          csv: "id,action\\naud_1,admin.order.refunded\\n",
          authorization: req.headers.authorization
        }
      }));
      return;
    }
    if (req.method === "GET" && req.url === "/api/admin/audit-logs/retention-report?retention_days=2555&hot_days=180&integrity_sample_limit=500") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({
        success: true,
        data: {
          status: "warning",
          retention_days: 2555,
          hot_days: 180,
          total_logs: 4,
          integrity_sample_size: 4,
          integrity_failures: 0,
          export_events: 1,
          missing_critical_actions: ["after_sales.reviewed"],
          alerts: [{ code: "audit.missing_critical_action", severity: "warning", count: 1 }],
          authorization: req.headers.authorization
        }
      }));
      return;
    }
    if (req.method === "POST" && req.url === "/api/admin/audit-logs/retention-alerts/emit") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({
        success: true,
        data: {
          emission: {
            status: "emitted",
            alert_count: 1,
            critical_count: 0,
            warning_count: 1,
            topic: "audit.retention_alerts",
            outbox_event_id: "obe_audit_alert_1",
            authorization: req.headers.authorization
          },
          outbox_event: { id: "obe_audit_alert_1", topic: "audit.retention_alerts", event_type: "audit.retention_alerts.emitted" },
          audit_log: { id: "aud_alert_1", action: "admin.audit_retention_alerts.emitted" }
        }
      }));
      return;
    }
    if (req.method === "POST" && req.url === "/api/admin/audit-logs/archive/request") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({
        success: true,
        data: {
          archive: {
            archive_id: "audit_archive_1",
            status: "requested",
            topic: "audit.archive_requested",
            storage_key: "worm://audit-logs/2026/05/24/audit_archive_1.jsonl",
            log_count: 2,
            manifest_algorithm: "sha256:v1",
            manifest_hash: "abc",
            outbox_event_id: "obe_audit_archive_1",
            authorization: req.headers.authorization
          },
          outbox_event: { id: "obe_audit_archive_1", topic: "audit.archive_requested", event_type: "audit.archive_requested" },
          audit_log: { id: "aud_archive_1", action: "admin.audit_archive.requested" }
        }
      }));
      return;
    }
    if (req.method === "GET" && req.url === "/api/admin/audit-logs/archive/records?archive_id=audit_archive_1&limit=5") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({
        success: true,
        data: [{
          archive_id: "audit_archive_1",
          status: "archived",
          storage_key: "worm://audit-logs/2026/05/24/audit_archive_1.jsonl",
          manifest_hash: "abc",
          content_hash: "content_hash_bff",
          bytes: 1024,
          authorization: req.headers.authorization
        }]
      }));
      return;
    }
    if (req.method === "GET" && req.url === "/api/admin/audit-logs/archive/verifications?archive_id=audit_archive_1&status=verified&limit=5") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({
        success: true,
        data: [{
          archive_id: "audit_archive_1",
          status: "verified",
          storage_key: "worm://audit-logs/2026/05/24/audit_archive_1.jsonl",
          actual_content_hash: "content_hash_bff",
          content_hash_matched: true,
          manifest_hash_matched: true,
          authorization: req.headers.authorization
        }]
      }));
      return;
    }
    if (req.method === "POST" && req.url === "/api/admin/audit-logs/archive/complete") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({
        success: true,
        data: {
          archive: {
            archive_id: "audit_archive_1",
            status: "archived",
            content_hash: "content_hash_bff",
            authorization: req.headers.authorization
          },
          audit_log: { id: "aud_archive_complete_1", action: "admin.audit_archive.completed" }
        }
      }));
      return;
    }
    if (req.method === "POST" && req.url === "/api/admin/audit-logs/archive/verify") {
      let body = "";
      req.on("data", (chunk) => {
        body += chunk;
      });
      req.on("end", () => {
        res.writeHead(200, { "Content-Type": "application/json" });
        res.end(JSON.stringify({
          success: true,
          data: {
            verification: {
              archive_id: "audit_archive_1",
              status: "verified",
              actual_content_hash: "content_hash_bff",
              content_hash_matched: true,
              manifest_hash_matched: true,
              request: JSON.parse(body),
              authorization: req.headers.authorization
            },
            audit_log: { id: "aud_archive_verify_1", action: "admin.audit_archive.verified" }
          }
        }));
      });
      return;
    }
    if (req.method === "GET" && req.url === "/api/admin/rbac/policy") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({
        success: true,
        data: {
          version: "2026-05-24.rbac.v1",
          current_role: "admin",
          can_request_changes: true,
          roles: [{ role: "admin", scopes: ["*"] }],
          scopes: [{ key: "rbac:read", risk_level: "medium" }],
          authorization: req.headers.authorization
        }
      }));
      return;
    }
    if (req.method === "GET" && req.url === "/api/admin/rbac/change-requests?status=pending_approval&limit=5") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({
        success: true,
        data: {
          items: [{ id: "rbac_change_1", role: "support_admin", status: "pending_approval" }],
          pending_count: 1,
          authorization: req.headers.authorization
        }
      }));
      return;
    }
    if (req.method === "POST" && req.url === "/api/admin/rbac/change-requests") {
      let body = "";
      req.on("data", (chunk) => {
        body += chunk;
      });
      req.on("end", () => {
        res.writeHead(201, { "Content-Type": "application/json" });
        res.end(JSON.stringify({
          success: true,
          data: {
            id: "rbac_change_1",
            status: "pending_approval",
            request: JSON.parse(body),
            authorization: req.headers.authorization
          }
        }));
      });
      return;
    }
    if (req.method === "POST" && req.url === "/api/admin/rbac/change-requests/rbac_change_1/review") {
      let body = "";
      req.on("data", (chunk) => {
        body += chunk;
      });
      req.on("end", () => {
        res.writeHead(200, { "Content-Type": "application/json" });
        res.end(JSON.stringify({
          success: true,
          data: {
            change_request: { id: "rbac_change_1", status: "approved", review: JSON.parse(body) },
            auto_applied: false,
            authorization: req.headers.authorization
          }
        }));
      });
      return;
    }
    if (req.method === "POST" && req.url === "/api/admin/rbac/change-requests/rbac_change_1/apply") {
      let body = "";
      req.on("data", (chunk) => {
        body += chunk;
      });
      req.on("end", () => {
        res.writeHead(200, { "Content-Type": "application/json" });
        res.end(JSON.stringify({
          success: true,
          data: {
            change_request: { id: "rbac_change_1", status: "applied", apply: JSON.parse(body) },
            runtime_applied: true,
            authorization: req.headers.authorization
          }
        }));
      });
      return;
    }
    if (req.method === "POST" && req.url === "/api/admin/rbac/change-requests/rbac_change_1/rollback") {
      let body = "";
      req.on("data", (chunk) => {
        body += chunk;
      });
      req.on("end", () => {
        res.writeHead(200, { "Content-Type": "application/json" });
        res.end(JSON.stringify({
          success: true,
          data: {
            change_request: { id: "rbac_change_1", status: "rolled_back", rollback: JSON.parse(body) },
            runtime_applied: true,
            rolled_back: true,
            authorization: req.headers.authorization
          }
        }));
      });
      return;
    }
    if (req.method === "GET" && req.url === "/api/admin/merchant-qualifications?status=pending_review&limit=1") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({
        success: true,
        data: {
          status: "pending_review",
          counts: { total: 1, pending_review: 1, approved: 0, rejected: 0, expired: 0 },
          qualifications: [{
            qualification: { id: "mq_merchant_1_health", type: "health_certificate", status: "pending_review" },
            merchant: { id: "merchant_1", display_name: "蓝湾轻食" },
            recommended_operation: { key: "merchant-qualification-review" },
            authorization: req.headers.authorization
          }]
        }
      }));
      return;
    }
    if (req.method === "GET" && req.url === "/api/admin/merchant-qualifications/mq_merchant_1_health?audit_limit=5") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({
        success: true,
        data: {
          qualification: { id: "mq_merchant_1_health", type: "health_certificate", status: "pending_review" },
          merchant: { id: "merchant_1", display_name: "蓝湾轻食" },
          checklist: ["核验文件主体、证照编号和商户账号主体一致"],
          recommended_operation: { key: "merchant-qualification-review" },
          authorization: req.headers.authorization
        }
      }));
      return;
    }
    if (req.method === "POST" && req.url === "/api/admin/merchant-qualifications/mq_merchant_1_health/review") {
      let body = "";
      req.on("data", (chunk) => {
        body += chunk;
      });
      req.on("end", () => {
        res.writeHead(200, { "Content-Type": "application/json" });
        res.end(JSON.stringify({
          success: true,
          data: {
            profile: { account: { id: "merchant_1" }, missing_qualifications: [], can_accept_orders: true },
            qualification: { id: "mq_merchant_1_health", status: "approved", decision: JSON.parse(body).decision },
            audit_log: { id: "aud_merchant_qualification_review_1", action: "admin.merchant_qualification.reviewed" },
            outbox_event: { id: "obe_merchant_qualification_review_1", topic: "merchant.qualification_reviewed", event_type: "merchant.qualification.reviewed" },
            request: JSON.parse(body),
            authorization: req.headers.authorization
          }
        }));
      });
      return;
    }
    if (req.method === "GET" && req.url === "/api/admin/object-storage/cleanup-candidates?limit=1&grace_seconds=60") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({
        success: true,
        data: [{
          ticket_id: "aset_1",
          object_key: "after-sales/asr_1/sig/evidence.jpg",
          reason: "expired_unconfirmed",
          authorization: req.headers.authorization
        }]
      }));
      return;
    }
    if (req.method === "GET" && req.url === "/api/admin/object-storage/cleanup-stats?grace_seconds=60") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({
        success: true,
        data: {
          pending: 1,
          failed: 1,
          deleted: 0,
          cleanup_attempts: 1,
          authorization: req.headers.authorization
        }
      }));
      return;
    }
    if (req.method === "POST" && req.url === "/api/admin/object-storage/cleanup-complete") {
      let body = "";
      req.on("data", (chunk) => {
        body += chunk;
      });
      req.on("end", () => {
        res.writeHead(200, { "Content-Type": "application/json" });
        res.end(JSON.stringify({
          success: true,
          data: {
            ...JSON.parse(body),
            status: "deleted",
            authorization: req.headers.authorization
          }
        }));
      });
      return;
    }
    if (req.method === "POST" && req.url === "/api/admin/object-storage/cleanup-failed") {
      let body = "";
      req.on("data", (chunk) => {
        body += chunk;
      });
      req.on("end", () => {
        res.writeHead(200, { "Content-Type": "application/json" });
        res.end(JSON.stringify({
          success: true,
          data: {
            ...JSON.parse(body),
            cleanup_attempts: 1,
            authorization: req.headers.authorization
          }
        }));
      });
      return;
    }
    if (req.method === "POST" && req.url === "/api/admin/orders/ord_1/state/compensate") {
      let body = "";
      req.on("data", (chunk) => {
        body += chunk;
      });
      req.on("end", () => {
        res.writeHead(200, { "Content-Type": "application/json" });
        res.end(JSON.stringify({
          success: true,
          data: {
            changed: true,
            previous_status: "pending_payment",
            expected_status: "dispatching",
            order: { id: "ord_1", status: "dispatching" },
            request: JSON.parse(body),
            authorization: req.headers.authorization
          }
        }));
      });
      return;
    }
    if (req.method === "GET" && req.url === "/api/admin/outbox/events?topic=order.paid&limit=1") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({
        success: true,
        data: [{
          id: "obe_1",
          topic: "order.paid",
          aggregate_id: "ord_1",
          status: "pending",
          authorization: req.headers.authorization
        }]
      }));
      return;
    }
    if (req.method === "GET" && req.url === "/api/admin/outbox/events/obe_1?now=2026-05-22T12:00:30Z&audit_limit=5") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({
        success: true,
        data: {
          generated_at: "2026-05-22T12:00:30Z",
          event: {
            id: "obe_1",
            topic: "order.paid",
            aggregate_type: "order",
            aggregate_id: "ord_1",
            status: "failed",
            attempts: 1,
            last_error: "relay down"
          },
          incident_code: "outbox.retry_backoff",
          incident_severity: "warning",
          ready: false,
          blocked: true,
          payload_summary: [{ key: "order_id", value: "ord_1" }],
          related_targets: [{ target_type: "order", target_id: "ord_1", source: "aggregate", operation_key: "audit-logs" }],
          recommended_operation: { key: "outbox-replay-event", title: "恢复单个 Outbox", values: { event_id: "obe_1" } },
          recent_audits: [{ id: "aud_outbox_failed_1", action: "admin.outbox.failed" }],
          authorization: req.headers.authorization
        }
      }));
      return;
    }
    if (req.method === "POST" && req.url === "/api/admin/outbox/events/claim") {
      let body = "";
      req.on("data", (chunk) => {
        body += chunk;
      });
      req.on("end", () => {
        res.writeHead(200, { "Content-Type": "application/json" });
        res.end(JSON.stringify({
          success: true,
          data: {
            topic: "order.paid",
            claimed: 1,
            lease_owner: "relay-bff",
            lease_expires_at: "2026-05-22T12:01:00Z",
            events: [{
              id: "obe_1",
              topic: "order.paid",
              lease_owner: "relay-bff",
              lease_expires_at: "2026-05-22T12:01:00Z"
            }],
            request: JSON.parse(body),
            authorization: req.headers.authorization
          }
        }));
      });
      return;
    }
    if (req.method === "POST" && req.url === "/api/admin/outbox/events/obe_1/lease/renew") {
      let body = "";
      req.on("data", (chunk) => {
        body += chunk;
      });
      req.on("end", () => {
        res.writeHead(200, { "Content-Type": "application/json" });
        res.end(JSON.stringify({
          success: true,
          data: {
            id: "obe_1",
            status: "pending",
            lease_owner: "relay-bff",
            lease_expires_at: "2026-05-22T12:01:30Z",
            request: JSON.parse(body),
            authorization: req.headers.authorization
          }
        }));
      });
      return;
    }
    if (req.method === "GET" && req.url === "/api/admin/outbox/stats?topic=order.paid&now=2026-05-22T12:03:00Z&lease_expiring_within_seconds=60") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({
        success: true,
          data: {
            generated_at: "2026-05-22T12:03:00Z",
            topic: "order.paid",
            total: 1,
            pending: 0,
            failed: 1,
            dead_letter: 0,
            published: 0,
            leased: 0,
            lease_expiring_within_seconds: 60,
            lease_expiring_soon: 0,
            ready: 1,
            blocked: 0,
            oldest_ready_lag_seconds: 60,
            next_lease_expires_in_seconds: 0,
            lease_owners: [],
          topics: [{
            topic: "order.paid",
            total: 1,
            pending: 0,
            failed: 1,
            dead_letter: 0,
            published: 0,
            leased: 0,
            lease_expiring_soon: 0,
            ready: 1,
            blocked: 0,
            oldest_ready_lag_seconds: 60,
            next_lease_expires_in_seconds: 0
          }],
          authorization: req.headers.authorization
        }
      }));
      return;
    }
    if (req.method === "POST" && req.url === "/api/admin/outbox/events/obe_1/failed") {
      let body = "";
      req.on("data", (chunk) => {
        body += chunk;
      });
      req.on("end", () => {
        res.writeHead(200, { "Content-Type": "application/json" });
        res.end(JSON.stringify({
          success: true,
          data: {
            id: "obe_1",
            status: "failed",
            request: JSON.parse(body),
            authorization: req.headers.authorization
          }
        }));
      });
      return;
    }
    if (req.method === "POST" && req.url === "/api/admin/outbox/events/replay") {
      let body = "";
      req.on("data", (chunk) => {
        body += chunk;
      });
      req.on("end", () => {
        res.writeHead(200, { "Content-Type": "application/json" });
        res.end(JSON.stringify({
          success: true,
          data: {
            topic: "order.paid",
            limit: 10,
            replayed: 1,
            events: [{
              id: "obe_1",
              status: "pending",
              available_at: "2026-05-22T12:01:20Z"
            }],
            request: JSON.parse(body),
            authorization: req.headers.authorization
          }
        }));
      });
      return;
    }
    if (req.method === "POST" && req.url === "/api/admin/outbox/events/obe_1/replay") {
      let body = "";
      req.on("data", (chunk) => {
        body += chunk;
      });
      req.on("end", () => {
        res.writeHead(200, { "Content-Type": "application/json" });
        res.end(JSON.stringify({
          success: true,
          data: {
            id: "obe_1",
            status: "pending",
            available_at: "2026-05-22T12:01:30Z",
            request: JSON.parse(body),
            authorization: req.headers.authorization
          }
        }));
      });
      return;
    }
    if (req.method === "POST" && req.url === "/api/admin/outbox/events/obe_1/published") {
      let body = "";
      req.on("data", (chunk) => {
        body += chunk;
      });
      req.on("end", () => {
        res.writeHead(200, { "Content-Type": "application/json" });
        res.end(JSON.stringify({
          success: true,
          data: {
            id: "obe_1",
            status: "published",
            request: JSON.parse(body),
            authorization: req.headers.authorization
          }
        }));
      });
      return;
    }
    if (req.method === "POST" && req.url === "/api/station-manager/rider-invites") {
      let body = "";
      req.on("data", (chunk) => {
        body += chunk;
      });
      req.on("end", () => {
        res.writeHead(201, { "Content-Type": "application/json" });
        res.end(JSON.stringify({
          success: true,
          data: {
            token: "ri_station_rider",
            type: JSON.parse(body).type,
            station_id: JSON.parse(body).station_id,
            authorization: req.headers.authorization
          }
        }));
      });
      return;
    }
    if (req.method === "POST" && req.url === "/api/auth/rider/invite-register") {
      res.writeHead(201, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ success: true, data: { access_token: "rider.signed.token", rider: { id: "rider_3", type: "rider" } } }));
      return;
    }
    if (req.method === "POST" && req.url === "/api/auth/rider/login") {
      let body = "";
      req.on("data", (chunk) => {
        body += chunk;
      });
      req.on("end", () => {
        res.writeHead(200, { "Content-Type": "application/json" });
        res.end(JSON.stringify({
          success: true,
          data: {
            access_token: "rider.login.token",
            rider: { id: JSON.parse(body).account_id, type: "rider" }
          }
        }));
      });
      return;
    }
    if (req.method === "POST" && req.url === "/api/auth/merchant/invite-register") {
      let body = "";
      req.on("data", (chunk) => {
        body += chunk;
      });
      req.on("end", () => {
        const payload = JSON.parse(body);
        res.writeHead(201, { "Content-Type": "application/json" });
        res.end(JSON.stringify({
          success: true,
          data: {
            access_token: "merchant.invite.token",
            profile: { account: { id: "merchant_3", display_name: payload.display_name } }
          }
        }));
      });
      return;
    }
    if (req.method === "POST" && req.url === "/api/auth/merchant/login") {
      let body = "";
      req.on("data", (chunk) => {
        body += chunk;
      });
      req.on("end", () => {
        res.writeHead(200, { "Content-Type": "application/json" });
        res.end(JSON.stringify({
          success: true,
          data: {
            access_token: "merchant.login.token",
            profile: { account: { id: JSON.parse(body).account_id } }
          }
        }));
      });
      return;
    }
    if (req.method === "POST" && req.url === "/api/auth/admin/login") {
      let body = "";
      req.on("data", (chunk) => {
        body += chunk;
      });
      req.on("end", () => {
        res.writeHead(200, { "Content-Type": "application/json" });
        res.end(JSON.stringify({
          success: true,
          data: {
            access_token: "admin.login.token",
            admin: { id: JSON.parse(body).account_id, role: "admin" }
          }
        }));
      });
      return;
    }
    if (req.method === "GET" && req.url === "/api/shops") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ success: true, data: [{ id: "shop_1" }] }));
      return;
    }
    if (req.method === "GET" && req.url === "/api/shops/shop_1/groupbuy-deals") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ success: true, data: [{ id: "deal_two_person_set", status: "active" }] }));
      return;
    }
    if (req.method === "POST" && req.url === "/api/cart/items") {
      let body = "";
      req.on("data", (chunk) => {
        body += chunk;
      });
      req.on("end", () => {
        res.writeHead(200, { "Content-Type": "application/json" });
        res.end(JSON.stringify({
          success: true,
          data: {
            authorization: req.headers.authorization,
            body: JSON.parse(body)
          }
        }));
      });
      return;
    }
    if (req.method === "GET" && req.url === "/api/user/notification-preferences?notification_type=after_sales.updated&limit=10") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({
        success: true,
        data: [{ id: "ntfp_user_1", preference_key: "user:user_1:after_sales.updated", target_role: "user", target_id: "user_1", authorization: req.headers.authorization }]
      }));
      return;
    }
    if (req.method === "PUT" && req.url === "/api/user/notification-preferences") {
      let body = "";
      req.on("data", (chunk) => {
        body += chunk;
      });
      req.on("end", () => {
        res.writeHead(200, { "Content-Type": "application/json" });
        res.end(JSON.stringify({
          success: true,
          data: { id: "ntfp_user_1", preference_key: "user:user_1:after_sales.updated", authorization: req.headers.authorization, body }
        }));
      });
      return;
    }
    if (req.method === "GET" && req.url === "/api/orders") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ success: true, data: [{ id: "ord_1", status: "dispatching" }] }));
      return;
    }
    if (req.method === "POST" && req.url === "/api/orders/ord_1/refund") {
      let body = "";
      req.on("data", (chunk) => {
        body += chunk;
      });
      req.on("end", () => {
        res.writeHead(200, { "Content-Type": "application/json" });
        res.end(JSON.stringify({
          success: true,
          data: {
            refund: { id: "rfd_1", destination: "balance", status: "success", reason: JSON.parse(body).reason },
            order: { id: "ord_1", status: "refunded" },
            wallet_account: { user_id: "user_1", balance_fen: 1200 },
            authorization: req.headers.authorization
          }
        }));
      });
      return;
    }
    if (req.method === "GET" && req.url === "/api/after-sales") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ success: true, data: [{ id: "asr_1", order_id: "ord_1", status: "pending_merchant", authorization: req.headers.authorization }] }));
      return;
    }
    if (req.method === "POST" && req.url === "/api/after-sales") {
      let body = "";
      req.on("data", (chunk) => {
        body += chunk;
      });
      req.on("end", () => {
        res.writeHead(201, { "Content-Type": "application/json" });
        res.end(JSON.stringify({ success: true, data: { id: "asr_1", status: "pending_merchant", request: JSON.parse(body), authorization: req.headers.authorization } }));
      });
      return;
    }
    if (req.method === "GET" && req.url === "/api/after-sales/asr_1/events") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ success: true, data: [{ id: "asev_1", request_id: "asr_1", action: "created", authorization: req.headers.authorization }] }));
      return;
    }
    if (req.method === "POST" && req.url === "/api/after-sales/asr_1/events") {
      let body = "";
      req.on("data", (chunk) => {
        body += chunk;
      });
      req.on("end", () => {
        res.writeHead(201, { "Content-Type": "application/json" });
        res.end(JSON.stringify({
          success: true,
          data: {
            event: { id: "asev_2", request_id: "asr_1", action: JSON.parse(body).action },
            after_sales: { id: "asr_1", status: "admin_review" },
            authorization: req.headers.authorization
          }
        }));
      });
      return;
    }
    if (req.method === "POST" && req.url === "/api/after-sales/asr_1/evidence/upload-ticket") {
      let body = "";
      req.on("data", (chunk) => {
        body += chunk;
      });
      req.on("end", () => {
        res.writeHead(201, { "Content-Type": "application/json" });
        res.end(JSON.stringify({
          success: true,
          data: {
            object_key: "after-sales/asr_1/sig/evidence.jpg",
            upload_url: "https://object-storage.infinitech.local/upload/after-sales/asr_1/sig/evidence.jpg",
            public_url: "https://cdn.infinitech.local/after-sales/asr_1/sig/evidence.jpg",
            method: "PUT",
            request: JSON.parse(body),
            authorization: req.headers.authorization
          }
        }));
      });
      return;
    }
    if (req.method === "POST" && req.url === "/api/after-sales/asr_1/evidence/confirm") {
      let body = "";
      req.on("data", (chunk) => {
        body += chunk;
      });
      req.on("end", () => {
        res.writeHead(201, { "Content-Type": "application/json" });
        res.end(JSON.stringify({
          success: true,
          data: {
            evidence: { id: "ase_1", request_id: "asr_1", status: "uploaded", object_key: JSON.parse(body).object_key },
            event: { id: "asev_3", request_id: "asr_1", action: "evidence_uploaded" },
            after_sales: { id: "asr_1", evidence_urls: ["https://cdn.infinitech.local/after-sales/asr_1/sig/evidence.jpg"] },
            authorization: req.headers.authorization
          }
        }));
      });
      return;
    }
    if (req.method === "GET" && req.url === "/api/after-sales/asr_1/evidence") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({
        success: true,
        data: [{
          id: "ase_1",
          request_id: "asr_1",
          status: "uploaded",
          authorization: req.headers.authorization
        }]
      }));
      return;
    }
    if (req.method === "POST" && req.url === "/api/after-sales/asr_1/review") {
      let body = "";
      req.on("data", (chunk) => {
        body += chunk;
      });
      req.on("end", () => {
        res.writeHead(200, { "Content-Type": "application/json" });
        res.end(JSON.stringify({
          success: true,
          data: {
            after_sales: { id: "asr_1", status: "refunded", review_reason: JSON.parse(body).reason },
            refund: { id: "rfd_asr_1", status: "success" },
            order: { id: "ord_1", status: "refunded" },
            wallet_account: { user_id: "user_1", balance_fen: 1200 },
            authorization: req.headers.authorization
          }
        }));
      });
      return;
    }
    if (req.method === "GET" && req.url === "/api/merchant/orders") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ success: true, data: [{ id: "ord_1", status: "merchant_pending", authorization: req.headers.authorization }] }));
      return;
    }
    if (req.method === "GET" && req.url === "/api/merchant/after-sales") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ success: true, data: [{ id: "asr_1", status: "pending_merchant", authorization: req.headers.authorization }] }));
      return;
    }
    if (req.method === "GET" && req.url === "/api/merchant/me") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({
        success: true,
        data: {
          account: { id: "merchant_1", display_name: "蓝海餐厅", deposit_status: "paid" },
          can_accept_orders: true,
          missing_qualifications: [],
          staff: [{ id: "staff_1", shop_id: "shop_1", name: "张三" }],
          supplemental_materials: [{ id: "material_1", type: "storefront_photo" }],
          authorization: req.headers.authorization
        }
      }));
      return;
    }
    if (req.method === "GET" && req.url === "/api/merchant/notifications?status=unread&limit=10") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({
        success: true,
        data: [{ id: "ntf_1", target_role: "merchant", target_id: "merchant_1", status: "unread", source_event_id: "obe_mq_1", authorization: req.headers.authorization }]
      }));
      return;
    }
    if (req.method === "GET" && req.url === "/api/merchant/notification-preferences?notification_type=order.status_changed&limit=10") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({
        success: true,
        data: [{ id: "ntfp_merchant_1", preference_key: "merchant:merchant_1:order.status_changed", target_role: "merchant", target_id: "merchant_1", authorization: req.headers.authorization }]
      }));
      return;
    }
    if (req.method === "PUT" && req.url === "/api/merchant/notification-preferences") {
      let body = "";
      req.on("data", (chunk) => {
        body += chunk;
      });
      req.on("end", () => {
        res.writeHead(200, { "Content-Type": "application/json" });
        res.end(JSON.stringify({
          success: true,
          data: { id: "ntfp_merchant_1", preference_key: "merchant:merchant_1:order.status_changed", authorization: req.headers.authorization, body }
        }));
      });
      return;
    }
    if (req.method === "GET" && req.url === "/api/admin/notifications?target_role=merchant&target_id=merchant_1&status=unread&source_topic=merchant.qualification_reviewed&limit=10") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({
        success: true,
        data: [{ id: "ntf_1", target_role: "merchant", target_id: "merchant_1", status: "unread", source_topic: "merchant.qualification_reviewed", authorization: req.headers.authorization }]
      }));
      return;
    }
    if (req.method === "GET" && req.url === "/api/admin/notification-deliveries?target_role=merchant&target_id=merchant_1&status=failed&limit=10") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({
        success: true,
        data: [{ id: "ntfd_1", notification_id: "ntf_1", target_role: "merchant", target_id: "merchant_1", status: "failed", error_code: "invalid_openid", authorization: req.headers.authorization }]
      }));
      return;
    }
    if (req.method === "GET" && req.url === "/api/admin/notification-preferences?target_role=merchant&target_id=merchant_1&notification_type=merchant.qualification_reviewed&limit=10") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({
        success: true,
        data: [{ id: "ntfp_1", preference_key: "merchant:merchant_1:merchant.qualification_reviewed", target_role: "merchant", target_id: "merchant_1", notification_type: "merchant.qualification_reviewed", authorization: req.headers.authorization }]
      }));
      return;
    }
    if (req.method === "PUT" && req.url === "/api/admin/notification-preferences") {
      let body = "";
      req.on("data", (chunk) => {
        body += chunk;
      });
      req.on("end", () => {
        res.writeHead(200, { "Content-Type": "application/json" });
        res.end(JSON.stringify({
          success: true,
          data: {
            preference: { id: "ntfp_1", preference_key: "merchant:merchant_1:merchant.qualification_reviewed", authorization: req.headers.authorization, body },
            audit_log: { id: "aud_notification_preference_1", action: "admin.notification_preferences.saved" }
          }
        }));
      });
      return;
    }
    if (req.method === "POST" && req.url === "/api/admin/notification-preferences/batch") {
      let body = "";
      req.on("data", (chunk) => {
        body += chunk;
      });
      req.on("end", () => {
        res.writeHead(200, { "Content-Type": "application/json" });
        res.end(JSON.stringify({
          success: true,
          data: {
            batch: { batch_id: "ntfp_batch_1", saved: 2, preference_keys: ["merchant:merchant_1:merchant.qualification_reviewed", "merchant:merchant_1:order.status_changed"], authorization: req.headers.authorization, body },
            audit_log: { id: "aud_notification_preference_batch_1", action: "admin.notification_preferences.batch_saved" }
          }
        }));
      });
      return;
    }
    if (req.method === "GET" && req.url === "/api/admin/notification-preferences/change-requests?status=pending_approval&limit=5") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({
        success: true,
        data: { items: [{ id: "ntfp_change_1", status: "pending_approval", authorization: req.headers.authorization }], pending_count: 1 }
      }));
      return;
    }
    if (req.method === "POST" && req.url === "/api/admin/notification-preferences/change-requests") {
      let body = "";
      req.on("data", (chunk) => {
        body += chunk;
      });
      req.on("end", () => {
        res.writeHead(201, { "Content-Type": "application/json" });
        res.end(JSON.stringify({
          success: true,
          data: { id: "ntfp_change_1", status: "pending_approval", authorization: req.headers.authorization, body }
        }));
      });
      return;
    }
    if (req.method === "POST" && req.url === "/api/admin/notification-preferences/change-requests/ntfp_change_1/review") {
      let body = "";
      req.on("data", (chunk) => {
        body += chunk;
      });
      req.on("end", () => {
        res.writeHead(200, { "Content-Type": "application/json" });
        res.end(JSON.stringify({
          success: true,
          data: { change_request: { id: "ntfp_change_1", status: "approved", authorization: req.headers.authorization, body }, auto_applied: false }
        }));
      });
      return;
    }
    if (req.method === "POST" && req.url === "/api/admin/notification-preferences/change-requests/ntfp_change_1/apply") {
      let body = "";
      req.on("data", (chunk) => {
        body += chunk;
      });
      req.on("end", () => {
        res.writeHead(200, { "Content-Type": "application/json" });
        res.end(JSON.stringify({
          success: true,
          data: { change_request: { id: "ntfp_change_1", status: "applied", authorization: req.headers.authorization, skipped_count: 1 }, batch: { batch_id: "ntfp_batch_approval_1", saved: 1, body }, audit_log: { action: "admin.notification_preferences.change_applied" } }
        }));
      });
      return;
    }
    if (req.method === "POST" && req.url === "/api/admin/notification-deliveries/failure-alerts/emit") {
      let body = "";
      req.on("data", (chunk) => {
        body += chunk;
      });
      req.on("end", () => {
        res.writeHead(200, { "Content-Type": "application/json" });
        res.end(JSON.stringify({
          success: true,
          data: {
            emission: { status: "emitted", failed_count: 1, topic: "notification.delivery_failed_alerts", outbox_event_id: "obe_notification_failure_1", authorization: req.headers.authorization, body },
            outbox_event: { id: "obe_notification_failure_1", topic: "notification.delivery_failed_alerts" }
          }
        }));
      });
      return;
    }
    if (req.method === "POST" && req.url === "/api/admin/notification-deliveries/retries/schedule") {
      let body = "";
      req.on("data", (chunk) => {
        body += chunk;
      });
      req.on("end", () => {
        res.writeHead(200, { "Content-Type": "application/json" });
        res.end(JSON.stringify({
          success: true,
          data: {
            schedule: { status: "scheduled", scheduled_count: 1, topic: "notification.delivery_retries", retry_after_seconds: 300, outbox_event_id: "obe_notification_retry_1", authorization: req.headers.authorization, body },
            outbox_event: { id: "obe_notification_retry_1", topic: "notification.delivery_retries" }
          }
        }));
      });
      return;
    }
    if (req.method === "POST" && req.url === "/api/admin/notification-deliveries/quiet-window-retries/schedule") {
      let body = "";
      req.on("data", (chunk) => {
        body += chunk;
      });
      req.on("end", () => {
        res.writeHead(200, { "Content-Type": "application/json" });
        res.end(JSON.stringify({
          success: true,
          data: {
            schedule: { status: "scheduled", scheduled_count: 1, delivery_status: "queued", error_code: "notification_quiet_window", topic: "notification.delivery_retries", outbox_event_id: "obe_notification_quiet_retry_1", authorization: req.headers.authorization, body },
            outbox_event: { id: "obe_notification_quiet_retry_1", topic: "notification.delivery_retries" }
          }
        }));
      });
      return;
    }
    if (req.method === "POST" && req.url === "/api/notifications/provider-callback") {
      let body = "";
      req.on("data", (chunk) => {
        body += chunk;
      });
      req.on("end", () => {
        res.writeHead(201, { "Content-Type": "application/json" });
        res.end(JSON.stringify({
          success: true,
          data: {
            delivery: { id: "ntfd_provider_callback_1", status: "delivered", provider_message_id: JSON.parse(body).provider_message_id },
            authorization: req.headers.authorization || "",
            body
          }
        }));
      });
      return;
    }
    if (req.method === "POST" && req.url === "/api/merchant/notifications/ntf_1/read") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({
        success: true,
        data: { id: "ntf_1", status: "read", authorization: req.headers.authorization }
      }));
      return;
    }
    if (req.method === "POST" && req.url === "/api/merchant/qualifications") {
      let body = "";
      req.on("data", (chunk) => {
        body += chunk;
      });
      req.on("end", () => {
        res.writeHead(200, { "Content-Type": "application/json" });
        res.end(JSON.stringify({
          success: true,
          data: {
            missing_qualifications: [],
            request: JSON.parse(body),
            authorization: req.headers.authorization
          }
        }));
      });
      return;
    }
    if (req.method === "GET" && req.url === "/api/merchant/staff") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ success: true, data: [{ id: "staff_1", shop_id: "shop_1", name: "张三", authorization: req.headers.authorization }] }));
      return;
    }
    if (req.method === "POST" && req.url === "/api/merchant/staff") {
      let body = "";
      req.on("data", (chunk) => {
        body += chunk;
      });
      req.on("end", () => {
        res.writeHead(200, { "Content-Type": "application/json" });
        res.end(JSON.stringify({ success: true, data: { id: "staff_2", status: "active", ...JSON.parse(body), authorization: req.headers.authorization } }));
      });
      return;
    }
    if (req.method === "GET" && req.url === "/api/merchant/materials") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ success: true, data: [{ id: "material_1", type: "storefront_photo", authorization: req.headers.authorization }] }));
      return;
    }
    if (req.method === "POST" && req.url === "/api/merchant/materials") {
      let body = "";
      req.on("data", (chunk) => {
        body += chunk;
      });
      req.on("end", () => {
        res.writeHead(200, { "Content-Type": "application/json" });
        res.end(JSON.stringify({ success: true, data: { id: "material_2", status: "submitted", ...JSON.parse(body), authorization: req.headers.authorization } }));
      });
      return;
    }
    if (req.method === "GET" && req.url === "/api/merchant/deposit") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ success: true, data: { subject_type: "merchant", subject_id: "merchant_1", amount_fen: 5000, status: "paid", authorization: req.headers.authorization } }));
      return;
    }
    if (req.method === "POST" && req.url === "/api/merchant/deposit/pay") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ success: true, data: { subject_type: "merchant", subject_id: "merchant_1", amount_fen: 5000, status: "paid", authorization: req.headers.authorization } }));
      return;
    }
    if (req.method === "POST" && req.url === "/api/merchant/orders/ord_1/accept") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ success: true, data: { id: "ord_1", status: "preparing", authorization: req.headers.authorization } }));
      return;
    }
    if (req.method === "POST" && req.url === "/api/merchant/orders/ord_1/ready") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ success: true, data: { id: "ord_1", status: "dispatching", authorization: req.headers.authorization } }));
      return;
    }
    if (req.method === "GET" && req.url === "/api/merchant/products?shop_id=shop_1") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ success: true, data: [{ id: "prod_1", status: "active", authorization: req.headers.authorization }] }));
      return;
    }
    if (req.method === "POST" && req.url === "/api/merchant/products") {
      let body = "";
      req.on("data", (chunk) => {
        body += chunk;
      });
      req.on("end", () => {
        res.writeHead(200, { "Content-Type": "application/json" });
        res.end(JSON.stringify({ success: true, data: { id: "prod_2", ...JSON.parse(body), authorization: req.headers.authorization } }));
      });
      return;
    }
    if (req.method === "POST" && req.url === "/api/merchant/products/prod_2/status") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ success: true, data: { id: "prod_2", status: "sold_out", authorization: req.headers.authorization } }));
      return;
    }
    if (req.method === "POST" && req.url === "/api/groupbuy/orders") {
      let body = "";
      req.on("data", (chunk) => {
        body += chunk;
      });
      req.on("end", () => {
        res.writeHead(201, { "Content-Type": "application/json" });
        res.end(JSON.stringify({ success: true, data: { id: "ord_groupbuy_1", type: "groupbuy", ...JSON.parse(body), authorization: req.headers.authorization } }));
      });
      return;
    }
    if (req.method === "GET" && req.url === "/api/groupbuy/vouchers") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ success: true, data: [{ id: "gbv_1", status: "issued", authorization: req.headers.authorization }] }));
      return;
    }
    if (req.method === "POST" && req.url === "/api/merchant/groupbuy/vouchers/scan") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ success: true, data: { voucher: { id: "gbv_1", status: "redeemed" }, order: { id: "ord_groupbuy_1", status: "completed" }, authorization: req.headers.authorization } }));
      return;
    }
    if (req.method === "POST" && req.url === "/api/rider/online") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ success: true, data: { id: "rider_1", online: true, authorization: req.headers.authorization } }));
      return;
    }
    if (req.method === "GET" && req.url === "/api/rider/deposit") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ success: true, data: { subject_type: "rider", subject_id: "rider_1", amount_fen: 5000, status: "paid", authorization: req.headers.authorization } }));
      return;
    }
    if (req.method === "POST" && req.url === "/api/rider/deposit/pay") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ success: true, data: { subject_type: "rider", subject_id: "rider_1", amount_fen: 5000, status: "paid", authorization: req.headers.authorization } }));
      return;
    }
    if (req.method === "POST" && req.url === "/api/rider/deposit/wechat-exempt") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ success: true, data: { deposit: { subject_type: "rider", status: "wechat_exempt_approved" }, rider: { id: "rider_1", deposit_status: "wechat_exempt_approved" }, authorization: req.headers.authorization } }));
      return;
    }
    if (req.method === "POST" && req.url === "/api/rider/deposit/refund-request") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ success: true, data: { deposit: { subject_type: "rider", status: "refund_pending" }, rider: { id: "rider_1", deposit_status: "refund_pending" }, authorization: req.headers.authorization } }));
      return;
    }
    if (req.method === "POST" && req.url === "/api/rider/orders/ord_1/grab") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ success: true, data: { id: "ord_1", status: "rider_assigned", rider_id: "rider_1", authorization: req.headers.authorization } }));
      return;
    }
    if (req.method === "POST" && req.url === "/api/rider/orders/ord_1/pickup") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ success: true, data: { id: "ord_1", status: "picked_up", rider_id: "rider_1", authorization: req.headers.authorization } }));
      return;
    }
    if (req.method === "POST" && req.url === "/api/rider/orders/ord_1/delivered") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ success: true, data: { id: "ord_1", status: "completed", rider_id: "rider_1", authorization: req.headers.authorization } }));
      return;
    }
    if (req.method === "POST" && req.url === "/api/dispatch/orders/ord_1/auto-assign") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ success: true, data: { order: { id: "ord_1", rider_id: "rider_1" }, decision: { mode: "auto_assign" }, authorization: req.headers.authorization } }));
      return;
    }
    if (req.method === "POST" && req.url === "/api/dispatch/orders/ord_1/timeout-reassign") {
      let body = "";
      req.on("data", (chunk) => {
        body += chunk;
      });
      req.on("end", () => {
        res.writeHead(200, { "Content-Type": "application/json" });
        res.end(JSON.stringify({
          success: true,
          data: {
            order: { id: "ord_1", rider_id: "rider_2" },
            decision: { candidate_rider_id: "rider_2", reason: "assignment_timeout", timeout_seconds: JSON.parse(body).timeout_seconds },
            authorization: req.headers.authorization
          }
        }));
      });
      return;
    }
    if (req.method === "GET" && req.url === "/api/dispatch/orders/ord_1/events") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({
        success: true,
        data: [{
          id: "dpe_1",
          type: "dispatch.auto_assign",
          station_id: "station_1",
          authorization: req.headers.authorization
        }]
      }));
      return;
    }
    if (req.method === "POST" && req.url === "/api/rider/orders/ord_1/reject-assignment") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ success: true, data: { order: { id: "ord_1", rider_id: "rider_2" }, decision: { candidate_rider_id: "rider_2", can_decline_without_penalty: true, reason: "after_fixed_order_count" }, authorization: req.headers.authorization } }));
      return;
    }
    if (req.method === "GET" && req.url === "/api/station-manager/riders") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ success: true, data: [{ id: "rider_1", type: "rider", authorization: req.headers.authorization }] }));
      return;
    }
    if (req.method === "GET" && req.url === "/api/station-manager/orders") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ success: true, data: [{ id: "ord_1", status: "dispatching", authorization: req.headers.authorization }] }));
      return;
    }
    if (req.method === "POST" && req.url === "/api/station-manager/dispatch/ord_1/manual-assign") {
      let body = "";
      req.on("data", (chunk) => {
        body += chunk;
      });
      req.on("end", () => {
        res.writeHead(200, { "Content-Type": "application/json" });
        res.end(JSON.stringify({
          success: true,
          data: {
            order: { id: "ord_1", rider_id: JSON.parse(body).rider_id },
            decision: { mode: "manual_assign" },
            authorization: req.headers.authorization
          }
        }));
      });
      return;
    }
    if (req.method === "GET" && req.url === "/api/station-manager/task-duration") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ success: true, data: { station_id: "station_1", daily_task_duration_minutes: 480, daily_fixed_order_count: 30, authorization: req.headers.authorization } }));
      return;
    }
    if (req.method === "PUT" && req.url === "/api/station-manager/task-duration") {
      let body = "";
      req.on("data", (chunk) => {
        body += chunk;
      });
      req.on("end", () => {
        res.writeHead(200, { "Content-Type": "application/json" });
        res.end(JSON.stringify({ success: true, data: { station_id: "station_1", ...JSON.parse(body), authorization: req.headers.authorization } }));
      });
      return;
    }
    if (req.method === "GET" && req.url === "/api/station-manager/rider-performance") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ success: true, data: [{ rider_id: "rider_2", level: "A", dispatch_priority: 300, authorization: req.headers.authorization }] }));
      return;
    }
    if (req.method === "POST" && req.url === "/api/payments/wechat/prepay") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ success: true, data: { prepay: { out_trade_no: "wx_ord_1" } } }));
      return;
    }
    res.writeHead(404, { "Content-Type": "application/json" });
    res.end(JSON.stringify({ success: false }));
  });
  upstream.listen(0);
  await once(upstream, "listening");

  const apiBaseUrl = `http://127.0.0.1:${upstream.address().port}`;
  const server = createBffServer({ env: { API_BASE_URL: apiBaseUrl } });
  server.listen(0);
  await once(server, "listening");
  const baseUrl = `http://127.0.0.1:${server.address().port}`;

  const login = await postJSON(`${baseUrl}/api/auth/wechat-mini/login`, "", { code: "wx_code_1" });
  const phoneCode = await postJSON(`${baseUrl}/api/auth/phone/code`, "", { phone: "13900003333", purpose: "login" });
  const phoneLogin = await postJSON(`${baseUrl}/api/auth/phone/login`, "", { phone: "13900003333", code: "135790", mode: "code" });
  const phoneRegister = await postJSON(`${baseUrl}/api/auth/phone/register`, "", { phone: "13900003334", code: "135790", password: "Pass123", accepted_agreement: true });
  const logout = await postJSON(`${baseUrl}/api/auth/logout`, "Bearer signed.token", {});
  const adminRiderInvite = await postJSON(`${baseUrl}/api/admin/rider-invites`, "Bearer admin:admin_1", { type: "station_manager", station_id: "station_2" });
  const refundSettings = await getJSON(`${baseUrl}/api/admin/refund-settings`, "Bearer admin:admin_1");
  const savedRefundSettings = await putJSON(`${baseUrl}/api/admin/refund-settings`, "Bearer admin:admin_1", { default_refund_strategy: "balance_first" });
  const compensatedOrder = await postJSON(`${baseUrl}/api/admin/orders/ord_1/state/compensate`, "Bearer admin:admin_1", { now: "2026-05-22T12:00:00Z" });
  const outboxEvents = await getJSON(`${baseUrl}/api/admin/outbox/events?topic=order.paid&limit=1`, "Bearer admin:admin_1");
  const outboxEventDetail = await getJSON(`${baseUrl}/api/admin/outbox/events/obe_1?now=2026-05-22T12:00:30Z&audit_limit=5`, "Bearer admin:admin_1");
  const claimedOutbox = await postJSON(`${baseUrl}/api/admin/outbox/events/claim`, "Bearer admin:admin_1", { topic: "order.paid", limit: 1, lease_owner: "relay-bff", lease_seconds: 60, now: "2026-05-22T12:00:00Z" });
  const renewedOutboxLease = await postJSON(`${baseUrl}/api/admin/outbox/events/obe_1/lease/renew`, "Bearer admin:admin_1", { lease_owner: "relay-bff", lease_seconds: 60, now: "2026-05-22T12:00:30Z" });
  const failedOutbox = await postJSON(`${baseUrl}/api/admin/outbox/events/obe_1/failed`, "Bearer admin:admin_1", { error: "relay down", retry_after_seconds: 120, max_attempts: 10 });
  const outboxStats = await getJSON(`${baseUrl}/api/admin/outbox/stats?topic=order.paid&now=2026-05-22T12:03:00Z&lease_expiring_within_seconds=60`, "Bearer admin:admin_1");
  const replayedOutboxBatch = await postJSON(`${baseUrl}/api/admin/outbox/events/replay`, "Bearer admin:admin_1", { topic: "order.paid", limit: 10, now: "2026-05-22T12:01:20Z" });
  const replayedOutbox = await postJSON(`${baseUrl}/api/admin/outbox/events/obe_1/replay`, "Bearer admin:admin_1", { now: "2026-05-22T12:01:30Z" });
  const publishedOutbox = await postJSON(`${baseUrl}/api/admin/outbox/events/obe_1/published`, "Bearer admin:admin_1", { published_at: "2026-05-22T12:02:00Z" });
  const stationRiderInvite = await postJSON(`${baseUrl}/api/station-manager/rider-invites`, "Bearer station_manager:station_manager_1", { type: "rider", station_id: "station_1" });
  const riderInviteRegister = await postJSON(`${baseUrl}/api/auth/rider/invite-register`, "", { token: "ri_station_rider", password: "RiderPass123" });
  const riderLogin = await postJSON(`${baseUrl}/api/auth/rider/login`, "", { account_id: "rider_3", password: "RiderPass123" });
  const merchantInviteRegister = await postJSON(`${baseUrl}/api/auth/merchant/invite-register`, "", { token: "mi_merchant_3", display_name: "蓝海商户", account_type: "standard", password: "MerchantPass123" });
  const merchantLogin = await postJSON(`${baseUrl}/api/auth/merchant/login`, "", { account_id: "merchant_3", password: "MerchantPass123" });
  const adminLogin = await postJSON(`${baseUrl}/api/auth/admin/login`, "", { account_id: "admin_1", password: "AdminPass123" });
  const shops = await getJSON(`${baseUrl}/api/shops`);
  const groupbuyDeals = await getJSON(`${baseUrl}/api/shops/shop_1/groupbuy-deals`);
  const cart = await postJSON(`${baseUrl}/api/cart/items`, "Bearer user:user_1", { shop_id: "shop_1", product_id: "prod_beef_rice", quantity: 1 });
  const userNotificationPreferences = await getJSON(`${baseUrl}/api/user/notification-preferences?notification_type=after_sales.updated&limit=10`, "Bearer user:user_1");
  const savedUserNotificationPreference = await putJSON(`${baseUrl}/api/user/notification-preferences`, "Bearer user:user_1", { notification_type: "after_sales.updated", disabled_channels: ["sms"] });
  const orders = await getJSON(`${baseUrl}/api/orders`, "Bearer user:user_1");
  const refundedOrder = await postJSON(`${baseUrl}/api/orders/ord_1/refund`, "Bearer admin:admin_1", { reason: "商品售罄", idempotency_key: "refund_ord_1" });
  const userAfterSales = await getJSON(`${baseUrl}/api/after-sales`, "Bearer user:user_1");
  const createdAfterSales = await postJSON(`${baseUrl}/api/after-sales`, "Bearer user:user_1", { order_id: "ord_1", reason: "餐品漏送", requested_amount_fen: 1200 });
  const afterSalesEvents = await getJSON(`${baseUrl}/api/after-sales/asr_1/events`, "Bearer user:user_1");
  const addedAfterSalesEvent = await postJSON(`${baseUrl}/api/after-sales/asr_1/events`, "Bearer user:user_1", { action: "customer_service_intervention", message: "申请客服介入" });
  const afterSalesUploadTicket = await postJSON(`${baseUrl}/api/after-sales/asr_1/evidence/upload-ticket`, "Bearer user:user_1", { file_name: "evidence.jpg", content_type: "image/jpeg", size_bytes: 1024 });
  const confirmedAfterSalesEvidence = await postJSON(`${baseUrl}/api/after-sales/asr_1/evidence/confirm`, "Bearer user:user_1", { object_key: "after-sales/asr_1/sig/evidence.jpg", content_type: "image/jpeg", size_bytes: 1024 });
  const afterSalesEvidence = await getJSON(`${baseUrl}/api/after-sales/asr_1/evidence`, "Bearer merchant:merchant_1");
  const reviewedAfterSales = await postJSON(`${baseUrl}/api/after-sales/asr_1/review`, "Bearer merchant:merchant_1", { decision: "approve", reason: "确认漏送" });
  const adminAfterSales = await getJSON(`${baseUrl}/api/admin/after-sales`, "Bearer admin:admin_1");
  const adminOperationsSnapshot = await getJSON(`${baseUrl}/api/admin/operations/snapshot?limit=5&lease_expiring_within_seconds=60&object_cleanup_grace_seconds=60`, "Bearer admin:admin_1");
  const adminAuditLogs = await getJSON(`${baseUrl}/api/admin/audit-logs?target_type=order&limit=1`, "Bearer admin:admin_1");
  const adminAuditExport = await getJSON(`${baseUrl}/api/admin/audit-logs/export?target_type=order&limit=1`, "Bearer admin:admin_1");
  const adminAuditRetention = await getJSON(`${baseUrl}/api/admin/audit-logs/retention-report?retention_days=2555&hot_days=180&integrity_sample_limit=500`, "Bearer admin:admin_1");
  const adminAuditRetentionAlert = await postJSON(`${baseUrl}/api/admin/audit-logs/retention-alerts/emit`, "Bearer admin:admin_1", { retention_days: 2555, hot_days: 180, integrity_sample_limit: 500 });
  const adminAuditArchive = await postJSON(`${baseUrl}/api/admin/audit-logs/archive/request`, "Bearer admin:admin_1", { hot_days: 180, limit: 500, storage_prefix: "worm://audit-logs" });
  const adminAuditArchiveRecords = await getJSON(`${baseUrl}/api/admin/audit-logs/archive/records?archive_id=audit_archive_1&limit=5`, "Bearer admin:admin_1");
  const adminAuditArchiveVerifications = await getJSON(`${baseUrl}/api/admin/audit-logs/archive/verifications?archive_id=audit_archive_1&status=verified&limit=5`, "Bearer admin:admin_1");
  const adminAuditArchiveComplete = await postJSON(`${baseUrl}/api/admin/audit-logs/archive/complete`, "Bearer admin:admin_1", { archive_id: "audit_archive_1", storage_key: "worm://audit-logs/2026/05/24/audit_archive_1.jsonl", manifest_algorithm: "sha256:v1", manifest_hash: "abc", content_hash: "content_hash_bff", bytes: 1024 });
  const adminAuditArchiveVerify = await postJSON(`${baseUrl}/api/admin/audit-logs/archive/verify`, "Bearer admin:admin_1", { archive_id: "audit_archive_1" });
  const adminRBACPolicy = await getJSON(`${baseUrl}/api/admin/rbac/policy`, "Bearer admin:admin_1");
  const adminRBACChanges = await getJSON(`${baseUrl}/api/admin/rbac/change-requests?status=pending_approval&limit=5`, "Bearer admin:admin_1");
  const adminRBACChange = await postJSON(`${baseUrl}/api/admin/rbac/change-requests`, "Bearer admin:admin_1", { role: "support_admin", requested_scopes: ["after_sales:read", "rbac:read"], reason: "support recertification" });
  const adminRBACReview = await postJSON(`${baseUrl}/api/admin/rbac/change-requests/rbac_change_1/review`, "Bearer admin:admin_2", { decision: "approve", reason: "least privilege approved" });
  const adminRBACApply = await postJSON(`${baseUrl}/api/admin/rbac/change-requests/rbac_change_1/apply`, "Bearer admin:admin_2", { reason: "apply approved runtime policy" });
  const adminRBACRollback = await postJSON(`${baseUrl}/api/admin/rbac/change-requests/rbac_change_1/rollback`, "Bearer admin:admin_3", { reason: "rollback runtime policy" });
  const adminMerchantQualifications = await getJSON(`${baseUrl}/api/admin/merchant-qualifications?status=pending_review&limit=1`, "Bearer admin:ops_1");
  const adminMerchantQualificationDetail = await getJSON(`${baseUrl}/api/admin/merchant-qualifications/mq_merchant_1_health?audit_limit=5`, "Bearer admin:ops_1");
  const adminMerchantQualificationReview = await postJSON(`${baseUrl}/api/admin/merchant-qualifications/mq_merchant_1_health/review`, "Bearer admin:ops_1", { merchant_id: "merchant_1", decision: "approve", reason: "qualification verified" });
  const adminNotifications = await getJSON(`${baseUrl}/api/admin/notifications?target_role=merchant&target_id=merchant_1&status=unread&source_topic=merchant.qualification_reviewed&limit=10`, "Bearer admin:support_1");
  const adminNotificationDeliveries = await getJSON(`${baseUrl}/api/admin/notification-deliveries?target_role=merchant&target_id=merchant_1&status=failed&limit=10`, "Bearer admin:support_1");
  const adminNotificationPreferences = await getJSON(`${baseUrl}/api/admin/notification-preferences?target_role=merchant&target_id=merchant_1&notification_type=merchant.qualification_reviewed&limit=10`, "Bearer admin:support_1");
  const savedAdminNotificationPreference = await putJSON(`${baseUrl}/api/admin/notification-preferences`, "Bearer admin:ops_1", { target_role: "merchant", target_id: "merchant_1", notification_type: "merchant.qualification_reviewed", disabled_channels: ["sms"] });
  const savedAdminNotificationPreferenceBatch = await postJSON(`${baseUrl}/api/admin/notification-preferences/batch`, "Bearer admin:ops_1", { reason: "bulk notification policy rollout", preferences: [{ target_role: "merchant", target_id: "merchant_1", notification_type: "merchant.qualification_reviewed", disabled_channels: ["sms"] }, { target_role: "merchant", target_id: "merchant_1", notification_type: "order.status_changed", disabled_channels: ["push"] }] });
  const adminNotificationPreferenceChanges = await getJSON(`${baseUrl}/api/admin/notification-preferences/change-requests?status=pending_approval&limit=5`, "Bearer admin:support_1");
  const adminNotificationPreferenceChange = await postJSON(`${baseUrl}/api/admin/notification-preferences/change-requests`, "Bearer admin:ops_1", { reason: "approval needed", rollout: { mode: "target_ids", target_ids: ["merchant_1"], max_targets: 10 }, preferences: [{ target_role: "merchant", target_id: "merchant_1", notification_type: "order.status_changed", disabled_channels: ["sms"] }, { target_role: "user", target_id: "user_1", notification_type: "after_sales.updated", disabled_channels: ["push"] }] });
  const adminNotificationPreferenceReview = await postJSON(`${baseUrl}/api/admin/notification-preferences/change-requests/ntfp_change_1/review`, "Bearer admin:ops_2", { decision: "approve", reason: "reviewed" });
  const adminNotificationPreferenceApply = await postJSON(`${baseUrl}/api/admin/notification-preferences/change-requests/ntfp_change_1/apply`, "Bearer admin:ops_2", { reason: "apply approved policy" });
  const adminNotificationFailureAlert = await postJSON(`${baseUrl}/api/admin/notification-deliveries/failure-alerts/emit`, "Bearer admin:ops_1", { target_role: "merchant", target_id: "merchant_1", channel: "wechat_subscribe", limit: 10 });
  const adminNotificationRetrySchedule = await postJSON(`${baseUrl}/api/admin/notification-deliveries/retries/schedule`, "Bearer admin:ops_1", { target_role: "merchant", target_id: "merchant_1", channel: "wechat_subscribe", provider: "wechat_subscribe", limit: 10, retry_after_seconds: 300 });
  const adminNotificationQuietRetrySchedule = await postJSON(`${baseUrl}/api/admin/notification-deliveries/quiet-window-retries/schedule`, "Bearer admin:ops_1", { channel: "push", provider: "push", limit: 10, now: "2026-05-25T12:10:00Z" });
  const notificationProviderCallback = await postJSON(`${baseUrl}/api/notifications/provider-callback`, "", { notification_id: "ntf_1", channel: "wechat_subscribe", provider: "wechat_subscribe", status: "delivered", provider_message_id: "wx_msg_1", callback_at: "2026-05-25T12:09:00Z", signature: "signed" });
  const objectCleanupCandidates = await getJSON(`${baseUrl}/api/admin/object-storage/cleanup-candidates?limit=1&grace_seconds=60`, "Bearer admin:admin_1");
  const objectCleanupStats = await getJSON(`${baseUrl}/api/admin/object-storage/cleanup-stats?grace_seconds=60`, "Bearer admin:admin_1");
  const failedObjectCleanup = await postJSON(`${baseUrl}/api/admin/object-storage/cleanup-failed`, "Bearer admin:admin_1", { ticket_id: "aset_1", object_key: "after-sales/asr_1/sig/evidence.jpg", reason: "expired_unconfirmed", error: "delete denied" });
  const completedObjectCleanup = await postJSON(`${baseUrl}/api/admin/object-storage/cleanup-complete`, "Bearer admin:admin_1", { ticket_id: "aset_1", object_key: "after-sales/asr_1/sig/evidence.jpg", reason: "expired_unconfirmed" });
  const merchantProfile = await getJSON(`${baseUrl}/api/merchant/me`, "Bearer merchant:merchant_1");
  const merchantNotifications = await getJSON(`${baseUrl}/api/merchant/notifications?status=unread&limit=10`, "Bearer merchant:merchant_1");
  const merchantNotificationPreferences = await getJSON(`${baseUrl}/api/merchant/notification-preferences?notification_type=order.status_changed&limit=10`, "Bearer merchant:merchant_1");
  const savedMerchantNotificationPreference = await putJSON(`${baseUrl}/api/merchant/notification-preferences`, "Bearer merchant:merchant_1", { notification_type: "order.status_changed", disabled_channels: ["push"] });
  const readMerchantNotification = await postJSON(`${baseUrl}/api/merchant/notifications/ntf_1/read`, "Bearer merchant:merchant_1", { read_at: "2026-05-25T12:01:00Z" });
  const savedMerchantQualification = await postJSON(`${baseUrl}/api/merchant/qualifications`, "Bearer merchant:merchant_1", { type: "health_certificate", file_url: "https://cdn.test/health.jpg", expires_at: "2027-05-22T00:00:00Z" });
  const merchantStaff = await getJSON(`${baseUrl}/api/merchant/staff`, "Bearer merchant:merchant_1");
  const savedMerchantStaff = await postJSON(`${baseUrl}/api/merchant/staff`, "Bearer merchant:merchant_1", { shop_id: "shop_1", name: "李四", phone: "13900000000", health_certificate_url: "https://cdn.test/staff.jpg", health_certificate_expires_at: "2027-05-22T00:00:00Z" });
  const merchantMaterials = await getJSON(`${baseUrl}/api/merchant/materials`, "Bearer merchant:merchant_1");
  const savedMerchantMaterial = await postJSON(`${baseUrl}/api/merchant/materials`, "Bearer merchant:merchant_1", { shop_id: "shop_1", type: "kitchen_photo", file_url: "https://cdn.test/kitchen.jpg", description: "后厨照", expires_at: "2027-05-22T00:00:00Z" });
  const merchantOrders = await getJSON(`${baseUrl}/api/merchant/orders`, "Bearer merchant:merchant_1");
  const merchantAfterSales = await getJSON(`${baseUrl}/api/merchant/after-sales`, "Bearer merchant:merchant_1");
  const merchantDeposit = await getJSON(`${baseUrl}/api/merchant/deposit`, "Bearer merchant:merchant_1");
  const paidMerchantDeposit = await postJSON(`${baseUrl}/api/merchant/deposit/pay`, "Bearer merchant:merchant_1", { amount_fen: 5000 });
  const accepted = await postJSON(`${baseUrl}/api/merchant/orders/ord_1/accept`, "Bearer merchant:merchant_1", {});
  const ready = await postJSON(`${baseUrl}/api/merchant/orders/ord_1/ready`, "Bearer merchant:merchant_1", {});
  const merchantProducts = await getJSON(`${baseUrl}/api/merchant/products?shop_id=shop_1`, "Bearer merchant:merchant_1");
  const createdProduct = await postJSON(`${baseUrl}/api/merchant/products`, "Bearer merchant:merchant_1", { shop_id: "shop_1", name: "轻食鸡胸饭", price_fen: 2299, stock_count: 20 });
  const productStatus = await postJSON(`${baseUrl}/api/merchant/products/prod_2/status`, "Bearer merchant:merchant_1", { status: "sold_out" });
  const groupbuyOrder = await postJSON(`${baseUrl}/api/groupbuy/orders`, "Bearer user:user_1", { shop_id: "shop_1", deal_id: "deal_two_person_set", quantity: 1 });
  const groupbuyVouchers = await getJSON(`${baseUrl}/api/groupbuy/vouchers`, "Bearer user:user_1");
  const scannedVoucher = await postJSON(`${baseUrl}/api/merchant/groupbuy/vouchers/scan`, "Bearer merchant:merchant_1", { qr_payload: "infinitech://groupbuy/voucher/GBV1", method: "qr_scan" });
  const riderDeposit = await getJSON(`${baseUrl}/api/rider/deposit`, "Bearer rider:rider_1");
  const paidRiderDeposit = await postJSON(`${baseUrl}/api/rider/deposit/pay`, "Bearer rider:rider_1", { amount_fen: 5000 });
  const riderExempt = await postJSON(`${baseUrl}/api/rider/deposit/wechat-exempt`, "Bearer rider:rider_1", { application_id: "wx_exempt_1" });
  const riderRefund = await postJSON(`${baseUrl}/api/rider/deposit/refund-request`, "Bearer rider:rider_1", {});
  const riderOnline = await postJSON(`${baseUrl}/api/rider/online`, "Bearer rider:rider_1", { online: true, capacity: 2 });
  const grabbedOrder = await postJSON(`${baseUrl}/api/rider/orders/ord_1/grab`, "Bearer rider:rider_1", {});
  const pickedUpOrder = await postJSON(`${baseUrl}/api/rider/orders/ord_1/pickup`, "Bearer rider:rider_1", {});
  const deliveredOrder = await postJSON(`${baseUrl}/api/rider/orders/ord_1/delivered`, "Bearer rider:rider_1", {});
  const autoAssign = await postJSON(`${baseUrl}/api/dispatch/orders/ord_1/auto-assign`, "Bearer admin:admin_1", {});
  const timeoutReassign = await postJSON(`${baseUrl}/api/dispatch/orders/ord_1/timeout-reassign`, "Bearer station_manager:station_manager_1", { timeout_seconds: 60 });
  const dispatchEvents = await getJSON(`${baseUrl}/api/dispatch/orders/ord_1/events`, "Bearer admin:admin_1");
  const rejectAssign = await postJSON(`${baseUrl}/api/rider/orders/ord_1/reject-assignment`, "Bearer rider:rider_1", {});
  const stationRiders = await getJSON(`${baseUrl}/api/station-manager/riders`, "Bearer station_manager:station_manager_1");
  const stationOrders = await getJSON(`${baseUrl}/api/station-manager/orders`, "Bearer station_manager:station_manager_1");
  const manualAssign = await postJSON(`${baseUrl}/api/station-manager/dispatch/ord_1/manual-assign`, "Bearer station_manager:station_manager_1", { rider_id: "rider_2" });
  const stationTaskConfig = await getJSON(`${baseUrl}/api/station-manager/task-duration`, "Bearer station_manager:station_manager_1");
  const savedStationTaskConfig = await putJSON(`${baseUrl}/api/station-manager/task-duration`, "Bearer station_manager:station_manager_1", { daily_task_duration_minutes: 420, daily_fixed_order_count: 28 });
  const riderPerformance = await getJSON(`${baseUrl}/api/station-manager/rider-performance`, "Bearer station_manager:station_manager_1");
  const prepay = await postJSON(`${baseUrl}/api/payments/wechat/prepay`, "Bearer user:user_1", { order_id: "ord_1" });

  server.close();
  upstream.close();

  assert.equal(login.data.access_token, "signed.token");
  assert.equal(phoneCode.data.dev_code, "135790");
  assert.equal(phoneLogin.data.access_token, "phone.login.token");
  assert.equal(phoneRegister.data.access_token, "phone.register.token");
  assert.equal(logout.data.revoked, true);
  assert.equal(logout.data.authorization, "Bearer signed.token");
  assert.equal(adminRiderInvite.data.authorization, "Bearer admin:admin_1");
  assert.equal(adminRiderInvite.data.type, "station_manager");
  assert.equal(refundSettings.data.default_refund_strategy, "balance_first");
  assert.equal(refundSettings.data.authorization, "Bearer admin:admin_1");
  assert.equal(savedRefundSettings.data.default_refund_strategy, "balance_first");
  assert.equal(compensatedOrder.data.authorization, "Bearer admin:admin_1");
  assert.equal(compensatedOrder.data.changed, true);
  assert.equal(compensatedOrder.data.previous_status, "pending_payment");
  assert.equal(compensatedOrder.data.expected_status, "dispatching");
  assert.equal(compensatedOrder.data.request.now, "2026-05-22T12:00:00Z");
  assert.equal(outboxEvents.data[0].authorization, "Bearer admin:admin_1");
  assert.equal(outboxEvents.data[0].topic, "order.paid");
  assert.equal(outboxEventDetail.data.authorization, "Bearer admin:admin_1");
  assert.equal(outboxEventDetail.data.incident_code, "outbox.retry_backoff");
  assert.equal(outboxEventDetail.data.recommended_operation.key, "outbox-replay-event");
  assert.equal(claimedOutbox.data.authorization, "Bearer admin:admin_1");
  assert.equal(claimedOutbox.data.claimed, 1);
	  assert.equal(claimedOutbox.data.events[0].lease_owner, "relay-bff");
	  assert.equal(claimedOutbox.data.request.lease_seconds, 60);
	  assert.equal(claimedOutbox.data.request.now, "2026-05-22T12:00:00Z");
	  assert.equal(renewedOutboxLease.data.authorization, "Bearer admin:admin_1");
	  assert.equal(renewedOutboxLease.data.lease_owner, "relay-bff");
	  assert.equal(renewedOutboxLease.data.lease_expires_at, "2026-05-22T12:01:30Z");
	  assert.equal(renewedOutboxLease.data.request.now, "2026-05-22T12:00:30Z");
	  assert.equal(failedOutbox.data.authorization, "Bearer admin:admin_1");
  assert.equal(failedOutbox.data.status, "failed");
  assert.equal(failedOutbox.data.request.retry_after_seconds, 120);
  assert.equal(failedOutbox.data.request.max_attempts, 10);
  assert.equal(outboxStats.data.authorization, "Bearer admin:admin_1");
  assert.equal(outboxStats.data.topic, "order.paid");
  assert.equal(outboxStats.data.dead_letter, 0);
  assert.equal(outboxStats.data.lease_expiring_within_seconds, 60);
  assert.equal(outboxStats.data.lease_expiring_soon, 0);
  assert.equal(outboxStats.data.next_lease_expires_in_seconds, 0);
  assert.equal(outboxStats.data.lease_owners.length, 0);
  assert.equal(outboxStats.data.ready, 1);
  assert.equal(outboxStats.data.blocked, 0);
  assert.equal(outboxStats.data.oldest_ready_lag_seconds, 60);
  assert.equal(outboxStats.data.topics[0].failed, 1);
  assert.equal(outboxStats.data.topics[0].lease_expiring_soon, 0);
  assert.equal(replayedOutboxBatch.data.authorization, "Bearer admin:admin_1");
  assert.equal(replayedOutboxBatch.data.topic, "order.paid");
  assert.equal(replayedOutboxBatch.data.replayed, 1);
  assert.equal(replayedOutboxBatch.data.request.limit, 10);
  assert.equal(replayedOutboxBatch.data.request.now, "2026-05-22T12:01:20Z");
  assert.equal(replayedOutbox.data.authorization, "Bearer admin:admin_1");
  assert.equal(replayedOutbox.data.status, "pending");
  assert.equal(replayedOutbox.data.available_at, "2026-05-22T12:01:30Z");
  assert.equal(replayedOutbox.data.request.now, "2026-05-22T12:01:30Z");
  assert.equal(publishedOutbox.data.authorization, "Bearer admin:admin_1");
  assert.equal(publishedOutbox.data.status, "published");
  assert.equal(publishedOutbox.data.request.published_at, "2026-05-22T12:02:00Z");
  assert.equal(stationRiderInvite.data.authorization, "Bearer station_manager:station_manager_1");
  assert.equal(stationRiderInvite.data.type, "rider");
  assert.equal(riderInviteRegister.data.rider.id, "rider_3");
  assert.equal(riderLogin.data.access_token, "rider.login.token");
  assert.equal(riderLogin.data.rider.id, "rider_3");
  assert.equal(merchantInviteRegister.data.access_token, "merchant.invite.token");
  assert.equal(merchantInviteRegister.data.profile.account.display_name, "蓝海商户");
  assert.equal(merchantLogin.data.access_token, "merchant.login.token");
  assert.equal(merchantLogin.data.profile.account.id, "merchant_3");
  assert.equal(adminLogin.data.access_token, "admin.login.token");
  assert.equal(adminLogin.data.admin.role, "admin");
  assert.equal(shops.data[0].id, "shop_1");
  assert.equal(groupbuyDeals.data[0].id, "deal_two_person_set");
  assert.equal(cart.data.authorization, "Bearer user:user_1");
  assert.equal(cart.data.body.product_id, "prod_beef_rice");
  assert.equal(userNotificationPreferences.data[0].authorization, "Bearer user:user_1");
  assert.equal(userNotificationPreferences.data[0].preference_key, "user:user_1:after_sales.updated");
  assert.equal(savedUserNotificationPreference.data.authorization, "Bearer user:user_1");
  assert.match(savedUserNotificationPreference.data.body, /after_sales\.updated/);
  assert.equal(orders.data[0].id, "ord_1");
  assert.equal(refundedOrder.data.refund.status, "success");
  assert.equal(refundedOrder.data.order.status, "refunded");
  assert.equal(refundedOrder.data.wallet_account.balance_fen, 1200);
  assert.equal(userAfterSales.data[0].authorization, "Bearer user:user_1");
  assert.equal(createdAfterSales.data.status, "pending_merchant");
  assert.equal(createdAfterSales.data.request.reason, "餐品漏送");
  assert.equal(afterSalesEvents.data[0].authorization, "Bearer user:user_1");
  assert.equal(addedAfterSalesEvent.data.event.action, "customer_service_intervention");
  assert.equal(afterSalesUploadTicket.data.authorization, "Bearer user:user_1");
  assert.equal(afterSalesUploadTicket.data.method, "PUT");
  assert.equal(confirmedAfterSalesEvidence.data.authorization, "Bearer user:user_1");
  assert.equal(confirmedAfterSalesEvidence.data.event.action, "evidence_uploaded");
  assert.equal(afterSalesEvidence.data[0].authorization, "Bearer merchant:merchant_1");
  assert.equal(afterSalesEvidence.data[0].status, "uploaded");
  assert.equal(reviewedAfterSales.data.after_sales.status, "refunded");
  assert.equal(reviewedAfterSales.data.wallet_account.balance_fen, 1200);
  assert.equal(adminAfterSales.data[0].authorization, "Bearer admin:admin_1");
  assert.equal(adminOperationsSnapshot.data.authorization, "Bearer admin:admin_1");
  assert.equal(adminOperationsSnapshot.data.counts.total_orders, 3);
  assert.equal(adminOperationsSnapshot.data.refund_settings.default_refund_strategy, "balance_first");
  assert.equal(adminAuditLogs.data[0].authorization, "Bearer admin:admin_1");
  assert.equal(adminAuditLogs.data[0].action, "admin.order.refunded");
  assert.equal(adminAuditExport.data.authorization, "Bearer admin:admin_1");
  assert.equal(adminAuditExport.data.format, "csv");
  assert.equal(adminAuditExport.data.row_count, 1);
  assert.equal(adminAuditRetention.data.authorization, "Bearer admin:admin_1");
  assert.equal(adminAuditRetention.data.status, "warning");
  assert.equal(adminAuditRetention.data.export_events, 1);
  assert.equal(adminAuditRetentionAlert.data.emission.authorization, "Bearer admin:admin_1");
  assert.equal(adminAuditRetentionAlert.data.emission.topic, "audit.retention_alerts");
  assert.equal(adminAuditRetentionAlert.data.outbox_event.event_type, "audit.retention_alerts.emitted");
  assert.equal(adminAuditArchive.data.archive.authorization, "Bearer admin:admin_1");
  assert.equal(adminAuditArchive.data.archive.topic, "audit.archive_requested");
  assert.equal(adminAuditArchive.data.outbox_event.event_type, "audit.archive_requested");
  assert.equal(adminAuditArchiveRecords.data[0].authorization, "Bearer admin:admin_1");
  assert.equal(adminAuditArchiveRecords.data[0].status, "archived");
  assert.equal(adminAuditArchiveVerifications.data[0].authorization, "Bearer admin:admin_1");
  assert.equal(adminAuditArchiveVerifications.data[0].status, "verified");
  assert.equal(adminAuditArchiveVerifications.data[0].actual_content_hash, "content_hash_bff");
  assert.equal(adminAuditArchiveComplete.data.archive.authorization, "Bearer admin:admin_1");
  assert.equal(adminAuditArchiveComplete.data.audit_log.action, "admin.audit_archive.completed");
  assert.equal(adminAuditArchiveVerify.data.verification.authorization, "Bearer admin:admin_1");
  assert.equal(adminAuditArchiveVerify.data.verification.status, "verified");
  assert.equal(adminAuditArchiveVerify.data.audit_log.action, "admin.audit_archive.verified");
  assert.equal(adminRBACPolicy.data.authorization, "Bearer admin:admin_1");
  assert.equal(adminRBACPolicy.data.can_request_changes, true);
  assert.equal(adminRBACChanges.data.authorization, "Bearer admin:admin_1");
  assert.equal(adminRBACChanges.data.pending_count, 1);
  assert.equal(adminRBACChange.data.authorization, "Bearer admin:admin_1");
  assert.equal(adminRBACChange.data.status, "pending_approval");
  assert.equal(adminRBACChange.data.request.role, "support_admin");
  assert.equal(adminRBACReview.data.authorization, "Bearer admin:admin_2");
  assert.equal(adminRBACReview.data.change_request.status, "approved");
  assert.equal(adminRBACReview.data.auto_applied, false);
  assert.equal(adminRBACApply.data.authorization, "Bearer admin:admin_2");
  assert.equal(adminRBACApply.data.change_request.status, "applied");
  assert.equal(adminRBACApply.data.runtime_applied, true);
  assert.equal(adminRBACRollback.data.authorization, "Bearer admin:admin_3");
  assert.equal(adminRBACRollback.data.change_request.status, "rolled_back");
  assert.equal(adminRBACRollback.data.runtime_applied, true);
  assert.equal(adminRBACRollback.data.rolled_back, true);
  assert.equal(adminMerchantQualifications.data.qualifications[0].authorization, "Bearer admin:ops_1");
  assert.equal(adminMerchantQualifications.data.qualifications[0].recommended_operation.key, "merchant-qualification-review");
  assert.equal(adminMerchantQualificationDetail.data.authorization, "Bearer admin:ops_1");
  assert.equal(adminMerchantQualificationDetail.data.recommended_operation.key, "merchant-qualification-review");
  assert.equal(adminMerchantQualificationReview.data.authorization, "Bearer admin:ops_1");
  assert.equal(adminMerchantQualificationReview.data.qualification.status, "approved");
  assert.equal(adminMerchantQualificationReview.data.outbox_event.topic, "merchant.qualification_reviewed");
  assert.equal(adminNotifications.data[0].authorization, "Bearer admin:support_1");
  assert.equal(adminNotifications.data[0].source_topic, "merchant.qualification_reviewed");
  assert.equal(adminNotificationDeliveries.data[0].authorization, "Bearer admin:support_1");
  assert.equal(adminNotificationDeliveries.data[0].error_code, "invalid_openid");
  assert.equal(adminNotificationPreferences.data[0].authorization, "Bearer admin:support_1");
  assert.equal(adminNotificationPreferences.data[0].preference_key, "merchant:merchant_1:merchant.qualification_reviewed");
  assert.equal(savedAdminNotificationPreference.data.preference.authorization, "Bearer admin:ops_1");
  assert.match(savedAdminNotificationPreference.data.preference.body, /disabled_channels/);
  assert.equal(savedAdminNotificationPreferenceBatch.data.batch.authorization, "Bearer admin:ops_1");
  assert.equal(savedAdminNotificationPreferenceBatch.data.batch.saved, 2);
  assert.match(savedAdminNotificationPreferenceBatch.data.batch.body, /order\.status_changed/);
  assert.equal(adminNotificationPreferenceChanges.data.items[0].authorization, "Bearer admin:support_1");
  assert.equal(adminNotificationPreferenceChange.data.authorization, "Bearer admin:ops_1");
  assert.match(adminNotificationPreferenceChange.data.body, /approval needed/);
  assert.match(adminNotificationPreferenceChange.data.body, /target_ids/);
  assert.equal(adminNotificationPreferenceReview.data.change_request.status, "approved");
  assert.equal(adminNotificationPreferenceApply.data.change_request.status, "applied");
  assert.equal(adminNotificationPreferenceApply.data.batch.saved, 1);
  assert.equal(adminNotificationPreferenceApply.data.change_request.skipped_count, 1);
  assert.equal(adminNotificationFailureAlert.data.emission.authorization, "Bearer admin:ops_1");
  assert.equal(adminNotificationFailureAlert.data.emission.failed_count, 1);
  assert.match(adminNotificationFailureAlert.data.emission.body, /wechat_subscribe/);
  assert.equal(adminNotificationRetrySchedule.data.schedule.authorization, "Bearer admin:ops_1");
  assert.equal(adminNotificationRetrySchedule.data.schedule.scheduled_count, 1);
  assert.match(adminNotificationRetrySchedule.data.schedule.body, /retry_after_seconds/);
  assert.equal(adminNotificationQuietRetrySchedule.data.schedule.authorization, "Bearer admin:ops_1");
  assert.equal(adminNotificationQuietRetrySchedule.data.schedule.delivery_status, "queued");
  assert.match(adminNotificationQuietRetrySchedule.data.schedule.body, /push/);
  assert.equal(notificationProviderCallback.data.delivery.provider_message_id, "wx_msg_1");
  assert.equal(JSON.parse(notificationProviderCallback.data.body).signature, "signed");
  assert.equal(merchantNotifications.data[0].authorization, "Bearer merchant:merchant_1");
  assert.equal(merchantNotifications.data[0].status, "unread");
  assert.equal(merchantNotificationPreferences.data[0].authorization, "Bearer merchant:merchant_1");
  assert.equal(merchantNotificationPreferences.data[0].preference_key, "merchant:merchant_1:order.status_changed");
  assert.equal(savedMerchantNotificationPreference.data.authorization, "Bearer merchant:merchant_1");
  assert.match(savedMerchantNotificationPreference.data.body, /push/);
  assert.equal(readMerchantNotification.data.authorization, "Bearer merchant:merchant_1");
  assert.equal(readMerchantNotification.data.status, "read");
  assert.equal(objectCleanupCandidates.data[0].authorization, "Bearer admin:admin_1");
  assert.equal(objectCleanupCandidates.data[0].reason, "expired_unconfirmed");
  assert.equal(objectCleanupStats.data.authorization, "Bearer admin:admin_1");
  assert.equal(objectCleanupStats.data.failed, 1);
  assert.equal(failedObjectCleanup.data.authorization, "Bearer admin:admin_1");
  assert.equal(failedObjectCleanup.data.cleanup_attempts, 1);
  assert.equal(completedObjectCleanup.data.authorization, "Bearer admin:admin_1");
  assert.equal(completedObjectCleanup.data.status, "deleted");
  assert.equal(merchantProfile.data.authorization, "Bearer merchant:merchant_1");
  assert.equal(merchantProfile.data.staff[0].id, "staff_1");
  assert.equal(savedMerchantQualification.data.request.type, "health_certificate");
  assert.equal(savedMerchantQualification.data.authorization, "Bearer merchant:merchant_1");
  assert.equal(merchantStaff.data[0].authorization, "Bearer merchant:merchant_1");
  assert.equal(savedMerchantStaff.data.name, "李四");
  assert.equal(savedMerchantStaff.data.status, "active");
  assert.equal(merchantMaterials.data[0].type, "storefront_photo");
  assert.equal(savedMerchantMaterial.data.type, "kitchen_photo");
  assert.equal(savedMerchantMaterial.data.status, "submitted");
  assert.equal(merchantOrders.data[0].authorization, "Bearer merchant:merchant_1");
  assert.equal(merchantAfterSales.data[0].authorization, "Bearer merchant:merchant_1");
  assert.equal(merchantDeposit.data.status, "paid");
  assert.equal(paidMerchantDeposit.data.amount_fen, 5000);
  assert.equal(accepted.data.status, "preparing");
  assert.equal(ready.data.status, "dispatching");
  assert.equal(merchantProducts.data[0].authorization, "Bearer merchant:merchant_1");
  assert.equal(createdProduct.data.name, "轻食鸡胸饭");
  assert.equal(productStatus.data.status, "sold_out");
  assert.equal(groupbuyOrder.data.authorization, "Bearer user:user_1");
  assert.equal(groupbuyVouchers.data[0].authorization, "Bearer user:user_1");
  assert.equal(scannedVoucher.data.voucher.status, "redeemed");
  assert.equal(riderDeposit.data.authorization, "Bearer rider:rider_1");
  assert.equal(paidRiderDeposit.data.status, "paid");
  assert.equal(riderExempt.data.deposit.status, "wechat_exempt_approved");
  assert.equal(riderRefund.data.deposit.status, "refund_pending");
  assert.equal(riderOnline.data.authorization, "Bearer rider:rider_1");
  assert.equal(grabbedOrder.data.status, "rider_assigned");
  assert.equal(pickedUpOrder.data.status, "picked_up");
  assert.equal(deliveredOrder.data.status, "completed");
  assert.equal(autoAssign.data.decision.mode, "auto_assign");
  assert.equal(timeoutReassign.data.decision.candidate_rider_id, "rider_2");
  assert.equal(timeoutReassign.data.decision.reason, "assignment_timeout");
  assert.equal(timeoutReassign.data.decision.timeout_seconds, 60);
  assert.equal(timeoutReassign.data.authorization, "Bearer station_manager:station_manager_1");
  assert.equal(dispatchEvents.data[0].type, "dispatch.auto_assign");
  assert.equal(dispatchEvents.data[0].station_id, "station_1");
  assert.equal(dispatchEvents.data[0].authorization, "Bearer admin:admin_1");
  assert.equal(rejectAssign.data.decision.candidate_rider_id, "rider_2");
  assert.equal(rejectAssign.data.decision.can_decline_without_penalty, true);
  assert.equal(stationRiders.data[0].authorization, "Bearer station_manager:station_manager_1");
  assert.equal(stationOrders.data[0].status, "dispatching");
  assert.equal(manualAssign.data.decision.mode, "manual_assign");
  assert.equal(stationTaskConfig.data.daily_fixed_order_count, 30);
  assert.equal(savedStationTaskConfig.data.daily_task_duration_minutes, 420);
  assert.equal(riderPerformance.data[0].level, "A");
  assert.equal(prepay.data.prepay.out_trade_no, "wx_ord_1");
});

test("bff proxies meal match candidates and safety actions", async () => {
  const upstream = http.createServer((req, res) => {
    if (req.method === "GET" && req.url === "/api/meal-match/candidates") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({
	        success: true,
	        data: {
	          privacy_scope: "same_building",
	          device_risk_state: "passed",
	          candidates: [{ user_id: "user_buddy_lunch", match_score: 160, same_school: true, same_building: true, privacy_scope: "same_building", authorization: req.headers.authorization }]
	        }
	      }));
      return;
    }
    if (req.method === "POST" && req.url === "/api/meal-match/reports") {
      res.writeHead(201, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ success: true, data: { id: "mmod_1", action: "reported", authorization: req.headers.authorization } }));
      return;
    }
    if (req.method === "POST" && req.url === "/api/meal-match/blocks") {
      res.writeHead(201, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ success: true, data: { id: "mmod_block_1", action: "blocked", authorization: req.headers.authorization } }));
      return;
    }
    if (req.method === "GET" && req.url === "/api/admin/meal-match/moderation?status=pending_review&action=reported") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({
        success: true,
        data: {
          records: [{ id: "mmod_1", action: "reported", status: "pending_review", authorization: req.headers.authorization }],
          count: 1
        }
      }));
      return;
    }
    if (req.method === "POST" && req.url === "/api/admin/meal-match/moderation/mmod_1/review") {
      let body = "";
      req.on("data", (chunk) => {
        body += chunk;
      });
      req.on("end", () => {
        res.writeHead(200, { "Content-Type": "application/json" });
        res.end(JSON.stringify({
          success: true,
          data: {
            id: "mmod_1",
            action: "reported",
            status: JSON.parse(body).decision === "approve" ? "approved" : "rejected",
            authorization: req.headers.authorization
          }
        }));
      });
      return;
    }
    res.writeHead(404, { "Content-Type": "application/json" });
    res.end(JSON.stringify({ success: false }));
  });
  upstream.listen(0);
  await once(upstream, "listening");

  const server = createBffServer({ env: { API_BASE_URL: `http://127.0.0.1:${upstream.address().port}` } });
  server.listen(0);
  await once(server, "listening");
  const baseUrl = `http://127.0.0.1:${server.address().port}`;

  const candidates = await getJSON(`${baseUrl}/api/meal-match/candidates`, "Bearer user:user_1");
  const report = await postJSON(`${baseUrl}/api/meal-match/reports`, "Bearer user:user_1", { target_user_id: "user_buddy_lunch", reason: "unsafe_or_fake_profile" });
  const block = await postJSON(`${baseUrl}/api/meal-match/blocks`, "Bearer user:user_1", { target_user_id: "user_buddy_lunch" });
  const queue = await getJSON(`${baseUrl}/api/admin/meal-match/moderation?status=pending_review&action=reported`, "Bearer support_admin:support_1");
  const review = await postJSON(`${baseUrl}/api/admin/meal-match/moderation/mmod_1/review`, "Bearer support_admin:support_1", { decision: "approve", review_note: "举报成立" });

  server.close();
  upstream.close();

	  assert.equal(candidates.data.candidates[0].authorization, "Bearer user:user_1");
	  assert.equal(candidates.data.privacy_scope, "same_building");
	  assert.equal(candidates.data.device_risk_state, "passed");
	  assert.equal(candidates.data.candidates[0].same_school, true);
	  assert.equal(report.data.action, "reported");
  assert.equal(block.data.action, "blocked");
  assert.equal(queue.data.records[0].authorization, "Bearer support_admin:support_1");
  assert.equal(review.data.status, "approved");
});

test("bff proxies red packet expiry refund admin action", async () => {
  const upstream = http.createServer((req, res) => {
    if (req.method === "POST" && req.url === "/api/admin/red-packets/expire") {
      let body = "";
      req.on("data", (chunk) => {
        body += chunk;
      });
      req.on("end", () => {
        res.writeHead(200, { "Content-Type": "application/json" });
        res.end(JSON.stringify({
          success: true,
          data: {
            count: 1,
            now: JSON.parse(body).now,
            authorization: req.headers.authorization
          }
        }));
      });
      return;
    }
    res.writeHead(404, { "Content-Type": "application/json" });
    res.end(JSON.stringify({ success: false }));
  });
  upstream.listen(0);
  await once(upstream, "listening");

  const server = createBffServer({ env: { API_BASE_URL: `http://127.0.0.1:${upstream.address().port}` } });
  server.listen(0);
  await once(server, "listening");
  const baseUrl = `http://127.0.0.1:${server.address().port}`;

  const body = await postJSON(`${baseUrl}/api/admin/red-packets/expire`, "Bearer admin:admin_1", { now: "2026-05-28T10:00:01Z" });

  server.close();
  upstream.close();

  assert.equal(body.data.count, 1);
  assert.equal(body.data.authorization, "Bearer admin:admin_1");
});

test("bff proxies service ticket workbench and closure actions", async () => {
  const upstream = http.createServer((req, res) => {
    if (req.method === "GET" && req.url === "/api/admin/service-tickets?limit=5") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ success: true, data: [{ id: "st_1", authorization: req.headers.authorization }] }));
      return;
    }
    if (req.method === "POST" && req.url === "/api/admin/service-tickets/st_1/assign") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ success: true, data: { ticket: { id: "st_1", assigned_support_id: "support_1" }, authorization: req.headers.authorization } }));
      return;
    }
    if (req.method === "POST" && req.url === "/api/admin/service-tickets/st_1/resolve") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ success: true, data: { ticket: { id: "st_1", status: "waiting_confirm" }, authorization: req.headers.authorization } }));
      return;
    }
    if (req.method === "POST" && req.url === "/api/admin/service-tickets/st_1/escalate") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ success: true, data: { ticket: { id: "st_1", sla_status: "escalated" }, authorization: req.headers.authorization } }));
      return;
    }
    if (req.method === "GET" && req.url === "/api/admin/service-ticket-quality-reviews?support_id=support_1&limit=5") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ success: true, data: [{ id: "stq_1", support_id: "support_1", authorization: req.headers.authorization }] }));
      return;
    }
    if (req.method === "GET" && req.url === "/api/admin/service-ticket-performance?support_id=support_1&limit=5") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ success: true, data: [{ support_id: "support_1", quality_review_count: 1, authorization: req.headers.authorization }] }));
      return;
    }
    if (req.method === "POST" && req.url === "/api/admin/service-tickets/st_1/quality-review") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ success: true, data: { id: "stq_1", result: "needs_coaching", authorization: req.headers.authorization } }));
      return;
    }
    if (req.method === "POST" && req.url === "/api/service-tickets/st_1/close") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ success: true, data: { ticket: { id: "st_1", status: "closed" }, authorization: req.headers.authorization } }));
      return;
    }
    if (req.method === "POST" && req.url === "/api/service-tickets/st_1/follow-up") {
      res.writeHead(200, { "Content-Type": "application/json" });
      res.end(JSON.stringify({ success: true, data: { ticket: { id: "st_1", follow_up_rating: 5 }, authorization: req.headers.authorization } }));
      return;
    }
    res.writeHead(404, { "Content-Type": "application/json" });
    res.end(JSON.stringify({ success: false }));
  });
  upstream.listen(0);
  await once(upstream, "listening");

  const server = createBffServer({ env: { API_BASE_URL: `http://127.0.0.1:${upstream.address().port}` } });
  server.listen(0);
  await once(server, "listening");
  const baseUrl = `http://127.0.0.1:${server.address().port}`;

  const tickets = await getJSON(`${baseUrl}/api/admin/service-tickets?limit=5`, "Bearer support_admin:support_1");
  const assigned = await postJSON(`${baseUrl}/api/admin/service-tickets/st_1/assign`, "Bearer support_admin:support_1", { support_name: "客服小悦" });
  const escalated = await postJSON(`${baseUrl}/api/admin/service-tickets/st_1/escalate`, "Bearer support_admin:support_1", { reason: "超过 10 分钟未更新" });
  const qualityReview = await postJSON(`${baseUrl}/api/admin/service-tickets/st_1/quality-review`, "Bearer support_admin:support_1", { score: 74, notes: "需辅导" });
  const qualityReviews = await getJSON(`${baseUrl}/api/admin/service-ticket-quality-reviews?support_id=support_1&limit=5`, "Bearer support_admin:support_1");
  const performance = await getJSON(`${baseUrl}/api/admin/service-ticket-performance?support_id=support_1&limit=5`, "Bearer support_admin:support_1");
  const resolved = await postJSON(`${baseUrl}/api/admin/service-tickets/st_1/resolve`, "Bearer support_admin:support_1", { solution: "已发放补偿券" });
  const closed = await postJSON(`${baseUrl}/api/service-tickets/st_1/close`, "Bearer user:user_1", { reason: "接受方案" });
  const followUp = await postJSON(`${baseUrl}/api/service-tickets/st_1/follow-up`, "Bearer user:user_1", { rating: 5 });

  server.close();
  upstream.close();

  assert.equal(tickets.data[0].authorization, "Bearer support_admin:support_1");
  assert.equal(assigned.data.ticket.assigned_support_id, "support_1");
  assert.equal(escalated.data.ticket.sla_status, "escalated");
  assert.equal(qualityReview.data.result, "needs_coaching");
  assert.equal(qualityReviews.data[0].authorization, "Bearer support_admin:support_1");
  assert.equal(performance.data[0].quality_review_count, 1);
  assert.equal(resolved.data.ticket.status, "waiting_confirm");
  assert.equal(closed.data.ticket.status, "closed");
  assert.equal(followUp.data.ticket.follow_up_rating, 5);
});

function getJSON(url, authorization = "") {
  return new Promise((resolve, reject) => {
    const req = http.get(url, {
      headers: authorization ? { Authorization: authorization } : {}
    }, (res) => {
      let body = "";
      res.on("data", (chunk) => {
        body += chunk;
      });
      res.on("end", () => {
        resolve(JSON.parse(body));
      });
    });
    req.on("error", reject);
  });
}

function postJSON(url, authorization, payload) {
  return requestJSON("POST", url, authorization, payload);
}

function putJSON(url, authorization, payload) {
  return requestJSON("PUT", url, authorization, payload);
}

function requestRaw(method, url, headers = {}) {
  return new Promise((resolve, reject) => {
    const req = http.request(url, {
      method,
      headers
    }, (res) => {
      let body = "";
      res.on("data", (chunk) => {
        body += chunk;
      });
      res.on("end", () => {
        resolve({ statusCode: res.statusCode, headers: res.headers, body });
      });
    });
    req.on("error", reject);
    req.end();
  });
}

function requestJSON(method, url, authorization, payload) {
  return new Promise((resolve, reject) => {
    const req = http.request(url, {
      method,
      headers: {
        "Content-Type": "application/json",
        Authorization: authorization
      }
    }, (res) => {
      let body = "";
      res.on("data", (chunk) => {
        body += chunk;
      });
      res.on("end", () => {
        resolve(JSON.parse(body));
      });
    });
    req.on("error", reject);
    req.end(JSON.stringify(payload));
  });
}
