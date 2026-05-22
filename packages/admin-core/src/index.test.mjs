import assert from "node:assert/strict";
import test from "node:test";
import { ADMIN_MODULES, getAdminModule } from "./index.mjs";

test("admin core exposes critical operations modules", () => {
  assert.ok(getAdminModule("dispatch"));
  assert.ok(getAdminModule("rider-performance"));
  assert.ok(getAdminModule("after-sales"));
  assert.ok(getAdminModule("refund-settings"));
  assert.ok(getAdminModule("home-cards"));
  assert.ok(getAdminModule("featured-products"));
  assert.ok(getAdminModule("home-campaigns"));
  assert.ok(getAdminModule("coupons"));
  assert.ok(getAdminModule("circle"));
  assert.ok(getAdminModule("reviews"));
  assert.ok(getAdminModule("groups"));
  assert.ok(getAdminModule("pricing"));
  assert.ok(getAdminModule("payment"));
  assert.ok(getAdminModule("points-membership"));
  assert.ok(getAdminModule("notifications"));
  assert.ok(getAdminModule("rtc"));
  assert.ok(getAdminModule("contact-audits"));
  assert.ok(getAdminModule("risk-control"));
  assert.ok(getAdminModule("data-management"));
  assert.ok(getAdminModule("content-settings"));
  assert.ok(ADMIN_MODULES.length >= 20);
});
