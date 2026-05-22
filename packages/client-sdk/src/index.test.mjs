import assert from "node:assert/strict";
import test from "node:test";
import { buildAuthHeaders, buildJsonRequest } from "./index.mjs";

test("auth headers normalize bearer token", () => {
  assert.deepEqual(buildAuthHeaders("abc"), { Authorization: "Bearer abc" });
  assert.deepEqual(buildAuthHeaders("Bearer abc"), { Authorization: "Bearer abc" });
  assert.deepEqual(buildAuthHeaders(""), {});
});

test("json requests include auth and serialized body", () => {
  assert.deepEqual(buildJsonRequest({ ok: true }, "token"), {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: "Bearer token"
    },
    body: "{\"ok\":true}"
  });
});

