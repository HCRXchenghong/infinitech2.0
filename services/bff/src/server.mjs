import { createServer } from "node:http";
import { createRuntimeConfig, defaultHomeCards, defaultHomeModules } from "./runtime.mjs";

const DEFAULT_ALLOWED_ORIGINS = [
  "http://127.0.0.1:4173",
  "http://localhost:4173",
  "http://127.0.0.1:5173",
  "http://localhost:5173",
  "http://127.0.0.1:8080",
  "http://localhost:8080"
];

function writeJson(res, statusCode, payload, extraHeaders = {}) {
  res.writeHead(statusCode, {
    "Content-Type": "application/json; charset=utf-8",
    "X-Content-Type-Options": "nosniff",
    ...extraHeaders
  });
  res.end(JSON.stringify(payload));
}

function success(data) {
  return { success: true, message: "ok", data };
}

function notFound() {
  return { success: false, code: "NOT_FOUND", message: "route not found" };
}

function badGateway(message = "upstream service unavailable") {
  return { success: false, code: "BAD_GATEWAY", message };
}

function forbiddenOrigin() {
  return { success: false, code: "FORBIDDEN_ORIGIN", message: "origin is not allowed" };
}

function parseAllowedOrigins(value) {
  if (!value) return DEFAULT_ALLOWED_ORIGINS;
  return value
    .split(",")
    .map((item) => item.trim())
    .filter((item) => item !== "*")
    .filter(Boolean);
}

function corsHeaders(req, allowedOrigins) {
  const origin = req.headers.origin;
  if (!origin) return {};
  if (!allowedOrigins.includes(origin)) return null;
  return {
    "Access-Control-Allow-Origin": origin,
    "Access-Control-Allow-Methods": "GET,POST,PUT,OPTIONS",
    "Access-Control-Allow-Headers": "Authorization,Content-Type,X-Client-Kind",
    "Access-Control-Max-Age": "600",
    "Vary": "Origin"
  };
}

function isApiProxyRoute(method, pathname) {
  if (method === "POST" && pathname === "/api/auth/wechat-mini/login") return true;
  if (method === "POST" && pathname === "/api/auth/logout") return true;
  if (method === "POST" && pathname === "/api/admin/merchant-invites") return true;
  if (method === "POST" && pathname === "/api/admin/rider-invites") return true;
  if (method === "GET" && pathname === "/api/admin/refund-settings") return true;
  if (method === "PUT" && pathname === "/api/admin/refund-settings") return true;
  if (method === "GET" && pathname === "/api/admin/after-sales") return true;
  if (method === "GET" && pathname === "/api/admin/operations/snapshot") return true;
  if (method === "GET" && pathname === "/api/admin/audit-logs") return true;
  if (method === "GET" && pathname === "/api/admin/audit-logs/export") return true;
  if (method === "GET" && pathname === "/api/admin/audit-logs/retention-report") return true;
  if (method === "POST" && pathname === "/api/admin/audit-logs/retention-alerts/emit") return true;
  if (method === "POST" && pathname === "/api/admin/audit-logs/archive/request") return true;
  if (method === "GET" && pathname === "/api/admin/rbac/policy") return true;
  if (method === "GET" && pathname === "/api/admin/rbac/change-requests") return true;
  if (method === "POST" && pathname === "/api/admin/rbac/change-requests") return true;
  if (method === "POST" && /^\/api\/admin\/rbac\/change-requests\/[^/]+\/review$/.test(pathname)) return true;
  if (method === "POST" && /^\/api\/admin\/rbac\/change-requests\/[^/]+\/apply$/.test(pathname)) return true;
  if (method === "POST" && /^\/api\/admin\/rbac\/change-requests\/[^/]+\/rollback$/.test(pathname)) return true;
  if (method === "GET" && pathname === "/api/admin/object-storage/cleanup-candidates") return true;
  if (method === "GET" && pathname === "/api/admin/object-storage/cleanup-stats") return true;
  if (method === "POST" && pathname === "/api/admin/object-storage/cleanup-complete") return true;
  if (method === "POST" && pathname === "/api/admin/object-storage/cleanup-failed") return true;
  if (method === "POST" && /^\/api\/admin\/orders\/[^/]+\/state\/compensate$/.test(pathname)) return true;
  if (method === "GET" && pathname === "/api/admin/outbox/events") return true;
  if (method === "GET" && pathname === "/api/admin/outbox/stats") return true;
  if (method === "POST" && pathname === "/api/admin/outbox/events/claim") return true;
  if (method === "POST" && /^\/api\/admin\/outbox\/events\/[^/]+\/lease\/renew$/.test(pathname)) return true;
  if (method === "POST" && /^\/api\/admin\/outbox\/events\/[^/]+\/published$/.test(pathname)) return true;
  if (method === "POST" && /^\/api\/admin\/outbox\/events\/[^/]+\/failed$/.test(pathname)) return true;
  if (method === "POST" && pathname === "/api/admin/outbox/events/replay") return true;
  if (method === "POST" && /^\/api\/admin\/outbox\/events\/[^/]+\/replay$/.test(pathname)) return true;
  if (method === "POST" && pathname === "/api/station-manager/rider-invites") return true;
  if (method === "POST" && pathname === "/api/auth/merchant/invite-register") return true;
  if (method === "POST" && pathname === "/api/auth/merchant/login") return true;
  if (method === "POST" && pathname === "/api/auth/admin/login") return true;
  if (method === "POST" && pathname === "/api/auth/rider/invite-register") return true;
  if (method === "POST" && pathname === "/api/auth/rider/login") return true;
  if (method === "GET" && pathname === "/api/merchant/me") return true;
  if (method === "POST" && pathname === "/api/merchant/qualifications") return true;
  if (method === "GET" && pathname === "/api/merchant/staff") return true;
  if (method === "POST" && pathname === "/api/merchant/staff") return true;
  if (method === "GET" && pathname === "/api/merchant/materials") return true;
  if (method === "POST" && pathname === "/api/merchant/materials") return true;
  if (method === "GET" && pathname === "/api/merchant/orders") return true;
  if (method === "GET" && pathname === "/api/merchant/after-sales") return true;
  if (method === "GET" && pathname === "/api/merchant/deposit") return true;
  if (method === "POST" && pathname === "/api/merchant/deposit/pay") return true;
  if (method === "POST" && /^\/api\/merchant\/orders\/[^/]+\/accept$/.test(pathname)) return true;
  if (method === "POST" && /^\/api\/merchant\/orders\/[^/]+\/ready$/.test(pathname)) return true;
  if (method === "GET" && pathname === "/api/merchant/products") return true;
  if (method === "POST" && pathname === "/api/merchant/products") return true;
  if (method === "POST" && /^\/api\/merchant\/products\/[^/]+\/status$/.test(pathname)) return true;
  if (method === "GET" && pathname === "/api/shops") return true;
  if (method === "GET" && /^\/api\/shops\/[^/]+\/products$/.test(pathname)) return true;
  if (method === "GET" && /^\/api\/shops\/[^/]+\/groupbuy-deals$/.test(pathname)) return true;
  if (method === "GET" && pathname === "/api/user/addresses") return true;
  if (method === "POST" && pathname === "/api/user/addresses") return true;
  if (method === "GET" && pathname === "/api/cart") return true;
  if (method === "POST" && pathname === "/api/cart/items") return true;
  if (method === "POST" && pathname === "/api/groupbuy/orders") return true;
  if (method === "GET" && pathname === "/api/groupbuy/vouchers") return true;
  if (method === "POST" && pathname === "/api/merchant/groupbuy/vouchers/scan") return true;
  if (method === "GET" && pathname === "/api/orders") return true;
  if (method === "GET" && /^\/api\/orders\/[^/]+$/.test(pathname)) return true;
  if (method === "POST" && /^\/api\/orders\/[^/]+\/refund$/.test(pathname)) return true;
  if (method === "GET" && pathname === "/api/after-sales") return true;
  if (method === "POST" && pathname === "/api/after-sales") return true;
  if (method === "GET" && /^\/api\/after-sales\/[^/]+\/events$/.test(pathname)) return true;
  if (method === "POST" && /^\/api\/after-sales\/[^/]+\/events$/.test(pathname)) return true;
  if (method === "GET" && /^\/api\/after-sales\/[^/]+\/evidence$/.test(pathname)) return true;
  if (method === "POST" && /^\/api\/after-sales\/[^/]+\/evidence\/upload-ticket$/.test(pathname)) return true;
  if (method === "POST" && /^\/api\/after-sales\/[^/]+\/evidence\/confirm$/.test(pathname)) return true;
  if (method === "POST" && /^\/api\/after-sales\/[^/]+\/review$/.test(pathname)) return true;
  if (method === "POST" && pathname === "/api/orders/checkout") return true;
  if (method === "POST" && pathname === "/api/wallet/credit") return true;
  if (method === "POST" && pathname === "/api/wallet/payment-password") return true;
  if (method === "POST" && pathname === "/api/wallet/pay") return true;
  if (method === "POST" && pathname === "/api/payments/wechat/prepay") return true;
  if (method === "GET" && pathname === "/api/rider/deposit") return true;
  if (method === "POST" && pathname === "/api/rider/deposit/pay") return true;
  if (method === "POST" && pathname === "/api/rider/deposit/wechat-exempt") return true;
  if (method === "POST" && pathname === "/api/rider/deposit/refund-request") return true;
  if (method === "POST" && pathname === "/api/rider/online") return true;
  if (method === "POST" && /^\/api\/rider\/orders\/[^/]+\/grab$/.test(pathname)) return true;
  if (method === "POST" && /^\/api\/rider\/orders\/[^/]+\/pickup$/.test(pathname)) return true;
  if (method === "POST" && /^\/api\/rider\/orders\/[^/]+\/delivered$/.test(pathname)) return true;
  if (method === "POST" && /^\/api\/dispatch\/orders\/[^/]+\/auto-assign$/.test(pathname)) return true;
  if (method === "POST" && /^\/api\/dispatch\/orders\/[^/]+\/timeout-reassign$/.test(pathname)) return true;
  if (method === "GET" && /^\/api\/dispatch\/orders\/[^/]+\/events$/.test(pathname)) return true;
  if (method === "POST" && /^\/api\/rider\/orders\/[^/]+\/reject-assignment$/.test(pathname)) return true;
  if (method === "GET" && pathname === "/api/station-manager/riders") return true;
  if (method === "GET" && pathname === "/api/station-manager/orders") return true;
  if (method === "POST" && /^\/api\/station-manager\/dispatch\/[^/]+\/manual-assign$/.test(pathname)) return true;
  if (method === "GET" && pathname === "/api/station-manager/task-duration") return true;
  if (method === "PUT" && pathname === "/api/station-manager/task-duration") return true;
  if (method === "GET" && pathname === "/api/station-manager/rider-performance") return true;
  return false;
}

function readRequestBody(req) {
  return new Promise((resolve, reject) => {
    let body = "";
    req.on("data", (chunk) => {
      body += chunk;
    });
    req.on("end", () => resolve(body));
    req.on("error", reject);
  });
}

async function proxyToApi(req, res, runtime, url, responseCorsHeaders = {}) {
  const body = req.method === "GET" || req.method === "HEAD" ? undefined : await readRequestBody(req);
  const headers = {
    "Content-Type": req.headers["content-type"] || "application/json",
    "X-Client-Kind": req.headers["x-client-kind"] || "unknown"
  };
  if (req.headers.authorization) {
    headers.Authorization = req.headers.authorization;
  }
  const upstreamUrl = `${runtime.apiBaseUrl}${url.pathname}${url.search}`;
  try {
    const upstream = await fetch(upstreamUrl, {
      method: req.method,
      headers,
      body
    });
    const text = await upstream.text();
    res.writeHead(upstream.status, {
      "Content-Type": upstream.headers.get("content-type") || "application/json; charset=utf-8",
      "X-Content-Type-Options": "nosniff",
      ...responseCorsHeaders
    });
    res.end(text);
  } catch (_error) {
    writeJson(res, 502, badGateway(), responseCorsHeaders);
  }
}

export function createBffServer(options = {}) {
  const env = options.env || process.env;
  const runtime = createRuntimeConfig(env);
  const allowedOrigins = parseAllowedOrigins(env.BFF_ALLOWED_ORIGINS);
  return createServer((req, res) => {
    const responseCorsHeaders = corsHeaders(req, allowedOrigins);
    if (responseCorsHeaders === null) {
      writeJson(res, 403, forbiddenOrigin());
      return;
    }
    if (req.method === "OPTIONS") {
      res.writeHead(204, responseCorsHeaders || {});
      res.end();
      return;
    }
    const url = new URL(req.url || "/", `http://${req.headers.host || "127.0.0.1"}`);
    if (req.method === "GET" && url.pathname === "/healthz") {
      writeJson(res, 200, success({ status: "ok", service: "bff" }), responseCorsHeaders);
      return;
    }
    if (req.method === "GET" && url.pathname === "/readyz") {
      writeJson(res, 200, success({ status: "ready", service: "bff", apiBaseUrl: runtime.apiBaseUrl }), responseCorsHeaders);
      return;
    }
    if (req.method === "GET" && url.pathname === "/api/runtime") {
      writeJson(res, 200, success(runtime), responseCorsHeaders);
      return;
    }
    if (req.method === "GET" && url.pathname === "/api/home/modules") {
      writeJson(res, 200, success(defaultHomeModules()), responseCorsHeaders);
      return;
    }
    if (req.method === "GET" && url.pathname === "/api/home/cards") {
      writeJson(res, 200, success(defaultHomeCards()), responseCorsHeaders);
      return;
    }
    if (isApiProxyRoute(req.method || "GET", url.pathname)) {
      void proxyToApi(req, res, runtime, url, responseCorsHeaders);
      return;
    }
    writeJson(res, 404, notFound(), responseCorsHeaders);
  });
}
