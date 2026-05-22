import { createHash } from "node:crypto";
import { createServer } from "node:http";
import { createRealtimeConfig } from "./runtime.mjs";

function writeJson(res, statusCode, payload) {
  res.writeHead(statusCode, {
    "Content-Type": "application/json; charset=utf-8",
    "X-Content-Type-Options": "nosniff"
  });
  res.end(JSON.stringify(payload));
}

export function createRealtimeServer(options = {}) {
  const config = createRealtimeConfig(options.env || process.env);
  const server = createServer((req, res) => {
    const url = new URL(req.url || "/", `http://${req.headers.host || "127.0.0.1"}`);
    if (req.method === "GET" && url.pathname === "/healthz") {
      writeJson(res, 200, { success: true, message: "ok", data: { status: "ok", service: "realtime-gateway" } });
      return;
    }
    if (req.method === "GET" && url.pathname === "/readyz") {
      writeJson(res, 200, { success: true, message: "ok", data: config });
      return;
    }
    writeJson(res, 426, { success: false, code: "WEBSOCKET_REQUIRED", message: "realtime gateway requires websocket upgrade" });
  });

  server.on("upgrade", (req, socket) => {
    if (req.url !== "/ws" || String(req.headers.upgrade || "").toLowerCase() !== "websocket") {
      socket.end("HTTP/1.1 404 Not Found\r\n\r\n");
      return;
    }
    const key = req.headers["sec-websocket-key"];
    if (!key) {
      socket.end("HTTP/1.1 400 Bad Request\r\n\r\n");
      return;
    }
    const accept = createHash("sha1")
      .update(`${key}258EAFA5-E914-47DA-95CA-C5AB0DC85B11`)
      .digest("base64");
    socket.write([
      "HTTP/1.1 101 Switching Protocols",
      "Upgrade: websocket",
      "Connection: Upgrade",
      `Sec-WebSocket-Accept: ${accept}`,
      "\r\n"
    ].join("\r\n"));
    socket.end();
  });

  return server;
}

