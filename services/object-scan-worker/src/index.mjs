import crypto from "node:crypto";
import net from "node:net";
import { createIdempotentConsumer } from "../../../packages/domain-core/src/index.mjs";

export const workerName = "object-scan-worker";
export const subscribedTopics = ["object.uploaded", "after_sales.evidence.object_uploaded"];

function clean(value) {
  return String(value || "").trim();
}

function normalizeContentType(value) {
  const contentType = clean(value).toLowerCase().split(";")[0].trim();
  if (contentType === "image/jpg") return "image/jpeg";
  return ["image/jpeg", "image/png", "image/webp", "image/heic", "application/pdf"].includes(contentType) ? contentType : "";
}

function normalizeScanStatus(value) {
  const status = clean(value).toLowerCase();
  if (status === "clean" || status === "pass" || status === "passed") return "passed";
  if (status === "infected" || status === "blocked" || status === "reject" || status === "rejected") return "rejected";
  return "pending";
}

function normalizeInteger(value) {
  const numeric = Number(value);
  if (!Number.isFinite(numeric)) return 0;
  return Math.max(0, Math.trunc(numeric));
}

function positiveInteger(value, fallback) {
  const normalized = normalizeInteger(value);
  return normalized > 0 ? normalized : fallback;
}

function isoSecond(value) {
  const date = value instanceof Date ? value : new Date(value || Date.now());
  if (Number.isNaN(date.getTime())) return new Date(0).toISOString();
  date.setMilliseconds(0);
  return date.toISOString();
}

function unixSeconds(value) {
  const date = new Date(isoSecond(value));
  return Math.trunc(date.getTime() / 1000);
}

export function normalizeObjectUploadEvent(event = {}) {
  const payload = event.payload || event.object || event;
  const ticketId = clean(payload.ticket_id || payload.ticketId);
  const objectKey = clean(payload.object_key || payload.objectKey);
  const bucket = clean(payload.bucket || payload.object_bucket || payload.objectBucket);
  const objectUrl = clean(payload.object_url || payload.objectUrl || payload.download_url || payload.downloadUrl || payload.public_url || payload.publicUrl);
  const contentType = normalizeContentType(payload.content_type || payload.contentType);
  const sizeBytes = normalizeInteger(payload.size_bytes || payload.sizeBytes);
  const contentSha = clean(payload.content_sha || payload.contentSha);
  const uploadedAt = isoSecond(payload.uploaded_at || payload.uploadedAt || event.created_at || event.createdAt || new Date());
  const idempotencyKey = clean(event.idempotency_key || event.idempotencyKey || payload.idempotency_key || payload.idempotencyKey || event.id)
    || `object-scan:${ticketId || objectKey}:${contentSha || sizeBytes}`;
  return {
    type: "object_uploaded",
    ticket_id: ticketId,
    object_key: objectKey,
    bucket,
    object_url: objectUrl,
    content_type: contentType,
    size_bytes: sizeBytes,
    content_sha: contentSha,
    uploaded_at: uploadedAt,
    idempotency_key: idempotencyKey
  };
}

export function normalizeObjectScanResult(upload = {}, result = {}) {
  const checkedAt = isoSecond(result.scan_checked_at || result.scanCheckedAt || result.checked_at || result.checkedAt || new Date());
  return {
    ticket_id: clean(result.ticket_id || result.ticketId || upload.ticket_id),
    object_key: clean(result.object_key || result.objectKey || upload.object_key),
    scan_status: normalizeScanStatus(result.scan_status || result.scanStatus || result.status),
    scan_result: clean(result.scan_result || result.scanResult || result.reason || result.message || ""),
    scanner: clean(result.scanner || "object-scan-worker"),
    scan_checked_at: checkedAt
  };
}

function encodeObjectPath(value) {
  return clean(value)
    .split("/")
    .map((part) => clean(part))
    .filter(Boolean)
    .map((part) => encodeURIComponent(part))
    .join("/");
}

export function buildObjectDownloadURL(upload = {}, options = {}) {
  const directURL = clean(upload.object_url || upload.objectUrl || upload.download_url || upload.downloadUrl);
  if (directURL) return directURL;
  const baseUrl = clean(options.downloadBaseUrl || options.objectStorageDownloadBaseUrl || process.env.OBJECT_STORAGE_DOWNLOAD_BASE_URL).replace(/\/+$/, "");
  if (!baseUrl) return "";
  const bucket = encodeObjectPath(upload.bucket || upload.object_bucket || upload.objectBucket);
  const objectKey = encodeObjectPath(upload.object_key || upload.objectKey);
  const path = [bucket, objectKey].filter(Boolean).join("/");
  return path ? `${baseUrl}/${path}` : baseUrl;
}

function responseHeader(response = {}, name = "") {
  if (typeof response.headers?.get === "function") return response.headers.get(name);
  return response.headers?.[name] || response.headers?.[name.toLowerCase()] || "";
}

export async function downloadObjectForScan(upload = {}, options = {}) {
  const fetchImpl = options.fetchImpl || globalThis.fetch;
  if (typeof fetchImpl !== "function") {
    throw new Error("fetch implementation is required");
  }
  const url = buildObjectDownloadURL(upload, options);
  if (!url) {
    throw new Error("object download URL is required for scan");
  }
  const maxBytes = positiveInteger(options.maxBytes || process.env.OBJECT_SCAN_MAX_BYTES, 10 * 1024 * 1024);
  const timeoutMs = positiveInteger(options.downloadTimeoutMs || process.env.OBJECT_SCAN_DOWNLOAD_TIMEOUT_MS, 15000);
  const controller = typeof AbortController !== "undefined" ? new AbortController() : null;
  const timeout = controller ? setTimeout(() => controller.abort(), timeoutMs) : null;
  try {
    const response = await fetchImpl(url, {
      method: "GET",
      headers: options.headers || undefined,
      signal: controller?.signal
    });
    if (!response || response.ok === false) {
      throw new Error(`object download failed: ${response?.status || 0}`);
    }
    const contentLength = normalizeInteger(responseHeader(response, "content-length"));
    if (contentLength > maxBytes) {
      throw new Error(`object exceeds scan limit: ${contentLength}`);
    }
    const body = typeof response.arrayBuffer === "function"
      ? await response.arrayBuffer()
      : typeof response.buffer === "function"
        ? await response.buffer()
        : null;
    if (!body) {
      throw new Error("object download response body is required");
    }
    const bytes = Buffer.from(body);
    if (bytes.byteLength > maxBytes) {
      throw new Error(`object exceeds scan limit: ${bytes.byteLength}`);
    }
    return bytes;
  } finally {
    if (timeout) clearTimeout(timeout);
  }
}

function parseClamAVResponse(rawResponse = "", options = {}) {
  const response = clean(String(rawResponse).replace(/\0/g, ""));
  const checkedAt = isoSecond(typeof options.clock === "function" ? options.clock() : new Date());
  if (/\bOK\b/i.test(response)) {
    return {
      scan_status: "passed",
      scan_result: "clean",
      scanner: "clamav",
      scan_checked_at: checkedAt
    };
  }
  if (/FOUND/i.test(response)) {
    return {
      scan_status: "rejected",
      scan_result: response,
      scanner: "clamav",
      scan_checked_at: checkedAt
    };
  }
  throw new Error(`unexpected ClamAV response: ${response || "empty"}`);
}

export function scanBufferWithClamAV(input, options = {}) {
  const buffer = Buffer.isBuffer(input) ? input : Buffer.from(input || "");
  const maxBytes = positiveInteger(options.maxBytes || process.env.OBJECT_SCAN_MAX_BYTES, 10 * 1024 * 1024);
  if (buffer.byteLength > maxBytes) {
    return Promise.reject(new Error(`object exceeds scan limit: ${buffer.byteLength}`));
  }
  const host = clean(options.host || options.clamavHost || process.env.CLAMAV_HOST || "localhost");
  const port = positiveInteger(options.port || options.clamavPort || process.env.CLAMAV_PORT, 3310);
  const timeoutMs = positiveInteger(options.timeoutMs || options.clamavTimeoutMs || process.env.OBJECT_SCAN_CLAMAV_TIMEOUT_MS, 15000);
  const chunkBytes = positiveInteger(options.chunkBytes || process.env.OBJECT_SCAN_CLAMAV_CHUNK_BYTES, 64 * 1024);
  const connect = options.connect || ((connectionOptions) => net.createConnection(connectionOptions));
  return new Promise((resolve, reject) => {
    let socket;
    let settled = false;
    const responseChunks = [];
    const timeout = setTimeout(() => fail(new Error("ClamAV scan timed out")), timeoutMs);
    function cleanup() {
      clearTimeout(timeout);
      if (socket && typeof socket.destroy === "function") {
        socket.destroy();
      }
    }
    function finish(result) {
      if (settled) return;
      settled = true;
      cleanup();
      resolve(result);
    }
    function fail(error) {
      if (settled) return;
      settled = true;
      cleanup();
      reject(error);
    }
    function parseBufferedResponse() {
      try {
        const response = Buffer.concat(responseChunks).toString("utf8");
        finish(parseClamAVResponse(response, options));
      } catch (error) {
        fail(error);
      }
    }
    function writeFrame(chunk) {
      const length = Buffer.alloc(4);
      length.writeUInt32BE(chunk.byteLength, 0);
      socket.write(length);
      if (chunk.byteLength > 0) socket.write(chunk);
    }
    function sendScan() {
      try {
        socket.write(Buffer.from("zINSTREAM\0"));
        for (let offset = 0; offset < buffer.byteLength; offset += chunkBytes) {
          writeFrame(buffer.subarray(offset, offset + chunkBytes));
        }
        socket.write(Buffer.alloc(4));
      } catch (error) {
        fail(error);
      }
    }
    try {
      socket = connect({ host, port });
      if (!socket || typeof socket.write !== "function") {
        throw new Error("ClamAV socket is required");
      }
      if (typeof socket.setTimeout === "function") {
        socket.setTimeout(timeoutMs);
      }
      if (typeof socket.on === "function") {
        socket.on("data", (chunk) => {
          responseChunks.push(Buffer.from(chunk));
          const response = Buffer.concat(responseChunks).toString("utf8");
          if (/\bOK\b/i.test(response) || /FOUND/i.test(response)) {
            parseBufferedResponse();
          }
        });
        socket.on("error", fail);
        socket.on("timeout", () => fail(new Error("ClamAV scan timed out")));
        socket.on("end", () => {
          if (!settled && responseChunks.length > 0) parseBufferedResponse();
        });
        socket.on("close", () => {
          if (!settled && responseChunks.length > 0) parseBufferedResponse();
        });
      }
      if (socket.connecting === false || typeof socket.once !== "function") {
        queueMicrotask(sendScan);
      } else {
        socket.once("connect", sendScan);
      }
    } catch (error) {
      fail(error);
    }
  });
}

export function createClamAVScanner(options = {}) {
  return async (upload = {}) => {
    const bytes = await downloadObjectForScan(upload, options);
    return scanBufferWithClamAV(bytes, options);
  };
}

export function signUploadCallback(payload = {}, secret = "") {
  if (!clean(secret)) return "";
  const lines = [
    clean(payload.ticket_id || payload.ticketId),
    clean(payload.object_key || payload.objectKey),
    normalizeContentType(payload.content_type || payload.contentType),
    String(normalizeInteger(payload.size_bytes || payload.sizeBytes)),
    clean(payload.content_sha || payload.contentSha),
    String(unixSeconds(payload.uploaded_at || payload.uploadedAt))
  ];
  return crypto.createHmac("sha256", clean(secret)).update(lines.join("\n")).digest("hex");
}

export function signScanResult(payload = {}, secret = "") {
  if (!clean(secret)) return "";
  const lines = [
    clean(payload.ticket_id || payload.ticketId),
    clean(payload.object_key || payload.objectKey),
    normalizeScanStatus(payload.scan_status || payload.scanStatus),
    clean(payload.scan_result || payload.scanResult),
    clean(payload.scanner),
    String(unixSeconds(payload.scan_checked_at || payload.scanCheckedAt))
  ];
  return crypto.createHmac("sha256", clean(secret)).update(lines.join("\n")).digest("hex");
}

async function postJSON(fetchImpl, url, token, body) {
  const headers = { "Content-Type": "application/json" };
  if (token) headers.Authorization = `Bearer ${token}`;
  const response = await fetchImpl(url, { method: "POST", headers, body: JSON.stringify(body) });
  if (!response || response.ok === false) {
    const status = response?.status || 0;
    throw new Error(`object scan API callback failed: ${status}`);
  }
  return typeof response.json === "function" ? response.json() : {};
}

export function createObjectStorageApiClient(options = {}) {
  const baseUrl = clean(options.baseUrl || options.apiBaseUrl || "http://localhost:1029").replace(/\/+$/, "");
  const token = clean(options.token || options.apiToken || "");
  const callbackSecret = clean(options.callbackSecret || "");
  const fetchImpl = options.fetchImpl || globalThis.fetch;
  if (typeof fetchImpl !== "function") {
    throw new Error("fetch implementation is required");
  }
  return {
    async confirmUpload(upload = {}) {
      const payload = {
        ticket_id: upload.ticket_id,
        object_key: upload.object_key,
        content_type: upload.content_type,
        size_bytes: upload.size_bytes,
        content_sha: upload.content_sha,
        uploaded_at: upload.uploaded_at
      };
      payload.signature = signUploadCallback(payload, callbackSecret);
      return postJSON(fetchImpl, `${baseUrl}/api/object-storage/upload-callback`, token, payload);
    },
    async reportScanResult(scan = {}) {
      const payload = {
        ticket_id: scan.ticket_id,
        object_key: scan.object_key,
        scan_status: scan.scan_status,
        scan_result: scan.scan_result,
        scanner: scan.scanner,
        scan_checked_at: scan.scan_checked_at
      };
      payload.signature = signScanResult(payload, callbackSecret);
      return postJSON(fetchImpl, `${baseUrl}/api/object-storage/scan-result`, token, payload);
    }
  };
}

export function createObjectScanConsumer(options = {}) {
  const apiClient = options.apiClient || createObjectStorageApiClient({
    baseUrl: options.apiBaseUrl || process.env.API_BASE_URL,
    token: options.token || process.env.OBJECT_SCAN_WORKER_TOKEN,
    callbackSecret: options.callbackSecret || process.env.OBJECT_STORAGE_CALLBACK_SIGNING_SECRET,
    fetchImpl: options.fetchImpl
  });
  const scannerMode = clean(options.scannerMode || process.env.OBJECT_SCAN_SCANNER).toLowerCase();
  const scanner = typeof options.scanner === "function"
    ? options.scanner
    : scannerMode === "clamav"
      ? createClamAVScanner(options)
      : async () => ({ scan_status: "pending", scan_result: "scanner_not_configured", scanner: workerName });
  return createIdempotentConsumer({
    consumerName: options.consumerName || workerName,
    ledger: options.ledger,
    clock: options.clock,
    handler: async (event = {}) => {
      const upload = normalizeObjectUploadEvent(event);
      if (!upload.ticket_id || !upload.object_key || !upload.content_type || upload.size_bytes <= 0) {
        throw new Error("invalid object uploaded event");
      }
      await apiClient.confirmUpload(upload);
      const scanResult = normalizeObjectScanResult(upload, await scanner(upload));
      if (scanResult.scan_status !== "passed" && scanResult.scan_status !== "rejected") {
        return { upload, scan_result: scanResult, reported_scan: false };
      }
      await apiClient.reportScanResult(scanResult);
      return { upload, scan_result: scanResult, reported_scan: true };
    }
  });
}

if (import.meta.url === `file://${process.argv[1]}`) {
  console.log(`${workerName} ready; topics=${subscribedTopics.join(",")}`);
}
