import assert from "node:assert/strict";
import { EventEmitter } from "node:events";
import test from "node:test";
import {
  buildObjectDownloadURL,
  createClamAVScanner,
  createObjectScanConsumer,
  createObjectStorageApiClient,
  downloadObjectForScan,
  normalizeObjectScanResult,
  normalizeObjectUploadEvent,
  scanBufferWithClamAV,
  signScanResult,
  signUploadCallback,
  subscribedTopics
} from "./index.mjs";

function createFakeClamAVSocket(responseText = "stream: OK\0") {
  const socket = new EventEmitter();
  socket.connecting = true;
  socket.writes = [];
  socket.setTimeout = () => socket;
  socket.destroy = () => {
    socket.destroyed = true;
  };
  socket.write = (chunk) => {
    const bytes = Buffer.from(chunk);
    socket.writes.push(bytes);
    if (bytes.byteLength === 4 && bytes.readUInt32BE(0) === 0) {
      queueMicrotask(() => socket.emit("data", Buffer.from(responseText)));
    }
    return true;
  };
  queueMicrotask(() => {
    socket.connecting = false;
    socket.emit("connect");
  });
  return socket;
}

test("object scan worker watches uploaded object topics", () => {
  assert.ok(subscribedTopics.includes("object.uploaded"));
  assert.ok(subscribedTopics.includes("after_sales.evidence.object_uploaded"));
});

test("object upload events normalize to API callback payload", () => {
  assert.deepEqual(normalizeObjectUploadEvent({
    id: "obe_object_1",
    topic: "object.uploaded",
    created_at: "2026-05-22T12:00:00.999Z",
    payload: {
      ticket_id: " aset_1 ",
      object_key: " after-sales/asr_1/sig/evidence.jpg ",
      bucket: " after-sales-evidence ",
      object_url: " https://object.test/private/evidence.jpg ",
      content_type: " image/jpg; charset=binary ",
      size_bytes: "2048",
      content_sha: " sha256:evidence "
    }
  }), {
    type: "object_uploaded",
    ticket_id: "aset_1",
    object_key: "after-sales/asr_1/sig/evidence.jpg",
    bucket: "after-sales-evidence",
    object_url: "https://object.test/private/evidence.jpg",
    content_type: "image/jpeg",
    size_bytes: 2048,
    content_sha: "sha256:evidence",
    uploaded_at: "2026-05-22T12:00:00.000Z",
    idempotency_key: "obe_object_1"
  });
});

test("object download URLs prefer signed URL and build encoded storage paths", () => {
  assert.equal(buildObjectDownloadURL({
    object_url: " https://signed.test/evidence.jpg?token=abc "
  }, {
    downloadBaseUrl: "https://objects.test"
  }), "https://signed.test/evidence.jpg?token=abc");
  assert.equal(buildObjectDownloadURL({
    bucket: "after sales",
    object_key: "asr_1/食品 证据.jpg"
  }, {
    downloadBaseUrl: "https://objects.test/"
  }), "https://objects.test/after%20sales/asr_1/%E9%A3%9F%E5%93%81%20%E8%AF%81%E6%8D%AE.jpg");
});

test("object download enforces scan size limits", async () => {
  const bytes = Buffer.from("clean file");
  const downloaded = await downloadObjectForScan({
    bucket: "after-sales",
    object_key: "asr_1/evidence.jpg"
  }, {
    downloadBaseUrl: "https://objects.test",
    maxBytes: 100,
    fetchImpl: async (url) => {
      assert.equal(url, "https://objects.test/after-sales/asr_1/evidence.jpg");
      return {
        ok: true,
        headers: { "content-length": String(bytes.byteLength) },
        arrayBuffer: async () => bytes
      };
    }
  });
  assert.equal(downloaded.toString(), "clean file");
  await assert.rejects(() => downloadObjectForScan({
    object_url: "https://objects.test/too-large.pdf"
  }, {
    maxBytes: 10,
    fetchImpl: async () => ({
      ok: true,
      headers: { "content-length": "11" },
      arrayBuffer: async () => Buffer.alloc(11)
    })
  }), /exceeds scan limit/);
});

test("clamav scanner streams INSTREAM frames and maps clean or infected results", async () => {
  let cleanSocket;
  const cleanResult = await scanBufferWithClamAV(Buffer.from("safe"), {
    clock: () => new Date("2026-05-22T12:02:03.999Z"),
    chunkBytes: 2,
    timeoutMs: 1000,
    connect: (options) => {
      assert.deepEqual(options, { host: "localhost", port: 3310 });
      cleanSocket = createFakeClamAVSocket("stream: OK\0");
      return cleanSocket;
    }
  });
  assert.deepEqual(cleanResult, {
    scan_status: "passed",
    scan_result: "clean",
    scanner: "clamav",
    scan_checked_at: "2026-05-22T12:02:03.000Z"
  });
  assert.equal(cleanSocket.writes[0].toString(), "zINSTREAM\0");
  assert.equal(cleanSocket.writes.at(-1).readUInt32BE(0), 0);

  const infectedResult = await scanBufferWithClamAV(Buffer.from("bad"), {
    timeoutMs: 1000,
    connect: () => createFakeClamAVSocket("stream: Eicar-Test-Signature FOUND\0")
  });
  assert.equal(infectedResult.scan_status, "rejected");
  assert.match(infectedResult.scan_result, /FOUND/);
});

test("clamav scanner downloads object bytes before reporting scan status", async () => {
  const scanner = createClamAVScanner({
    downloadBaseUrl: "https://objects.test",
    maxBytes: 1024,
    fetchImpl: async (url) => {
      assert.equal(url, "https://objects.test/after-sales/asr_1/evidence.pdf");
      return {
        ok: true,
        headers: { "content-length": "7" },
        arrayBuffer: async () => Buffer.from("payload")
      };
    },
    timeoutMs: 1000,
    connect: () => createFakeClamAVSocket("stream: OK\0")
  });
  const result = await scanner({
    bucket: "after-sales",
    object_key: "asr_1/evidence.pdf"
  });
  assert.equal(result.scan_status, "passed");
  assert.equal(result.scanner, "clamav");
});

test("scan results normalize scanner vocabulary", () => {
  assert.deepEqual(normalizeObjectScanResult({ ticket_id: "aset_1", object_key: "k" }, {
    status: "clean",
    reason: "ok",
    scanner: "clamav",
    checked_at: "2026-05-22T12:01:01.999Z"
  }), {
    ticket_id: "aset_1",
    object_key: "k",
    scan_status: "passed",
    scan_result: "ok",
    scanner: "clamav",
    scan_checked_at: "2026-05-22T12:01:01.000Z"
  });
  assert.equal(normalizeObjectScanResult({}, { status: "infected" }).scan_status, "rejected");
});

test("callback signatures match API canonical fields", () => {
  const uploadPayload = {
    ticket_id: "aset_1",
    object_key: "after-sales/asr_1/sig/evidence.jpg",
    content_type: "image/jpeg",
    size_bytes: 2048,
    content_sha: "sha256:evidence",
    uploaded_at: "2026-05-22T12:00:00.000Z"
  };
  const scanPayload = {
    ticket_id: "aset_1",
    object_key: "after-sales/asr_1/sig/evidence.jpg",
    scan_status: "passed",
    scan_result: "clean",
    scanner: "clamav",
    scan_checked_at: "2026-05-22T12:01:00.000Z"
  };
  assert.equal(signUploadCallback(uploadPayload, "callback-secret").length, 64);
  assert.equal(signUploadCallback(uploadPayload, "callback-secret"), signUploadCallback({ ...uploadPayload, content_type: "image/jpg" }, "callback-secret"));
  assert.equal(signScanResult(scanPayload, "callback-secret").length, 64);
});

test("api client posts upload callback and scan result with signatures", async () => {
  const calls = [];
  const client = createObjectStorageApiClient({
    baseUrl: "https://api.test/",
    token: "worker-token",
    callbackSecret: "callback-secret",
    fetchImpl: async (url, init) => {
      calls.push({ url, init, body: JSON.parse(init.body) });
      return { ok: true, json: async () => ({ success: true }) };
    }
  });
  await client.confirmUpload({
    ticket_id: "aset_1",
    object_key: "after-sales/asr_1/sig/evidence.jpg",
    content_type: "image/jpeg",
    size_bytes: 2048,
    content_sha: "sha256:evidence",
    uploaded_at: "2026-05-22T12:00:00.000Z"
  });
  await client.reportScanResult({
    ticket_id: "aset_1",
    object_key: "after-sales/asr_1/sig/evidence.jpg",
    scan_status: "passed",
    scan_result: "clean",
    scanner: "clamav",
    scan_checked_at: "2026-05-22T12:01:00.000Z"
  });
  assert.equal(calls[0].url, "https://api.test/api/object-storage/upload-callback");
  assert.equal(calls[1].url, "https://api.test/api/object-storage/scan-result");
  assert.equal(calls[0].init.headers.Authorization, "Bearer worker-token");
  assert.equal(calls[0].body.signature.length, 64);
  assert.equal(calls[1].body.signature.length, 64);
});

test("object scan consumer reports scan once and ignores duplicate outbox deliveries", async () => {
  const calls = [];
  let scanned = 0;
  const consumer = createObjectScanConsumer({
    clock: () => new Date("2026-05-22T12:02:00.000Z"),
    apiClient: {
      async confirmUpload(upload) {
        calls.push(["upload", upload.ticket_id]);
      },
      async reportScanResult(scan) {
        calls.push(["scan", scan.scan_status]);
      }
    },
    scanner: async () => {
      scanned += 1;
      return { scan_status: "passed", scan_result: "clean", scanner: "clamav", scan_checked_at: "2026-05-22T12:01:00.000Z" };
    }
  });
  const event = {
    id: "obe_object_1",
    topic: "object.uploaded",
    idempotency_key: "object:after-sales/asr_1/sig/evidence.jpg",
    payload: {
      ticket_id: "aset_1",
      object_key: "after-sales/asr_1/sig/evidence.jpg",
      content_type: "image/jpeg",
      size_bytes: 2048,
      content_sha: "sha256:evidence",
      uploaded_at: "2026-05-22T12:00:00.000Z"
    }
  };
  assert.equal((await consumer(event)).status, "processed");
  assert.equal((await consumer({ ...event, id: "obe_object_replay" })).status, "duplicate");
  assert.equal(scanned, 1);
  assert.deepEqual(calls, [["upload", "aset_1"], ["scan", "passed"]]);
  assert.equal(consumer.ledger.snapshot()[0].consumer_name, "object-scan-worker");
});
