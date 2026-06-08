import assert from "node:assert/strict";
import test from "node:test";
import {
  applyProviderTemplate,
  buildNotification,
  buildProviderDispatches,
  createNotificationPreferenceResolver,
  createNotificationApiClient,
  createNotificationConsumer,
  createNotificationProviderDispatcher,
  createQuietWindowRetryLoop,
  createQuietWindowRetryScheduler,
  deliveryPreferencesFromRecords,
  invalidateNotificationPreferenceCache,
  notificationPreferenceInvalidationKeys,
  normalizeNotificationDeliveryRecord,
  normalizeNotificationRecord,
  normalizeQuietWindowRetrySchedulePayload,
  notificationDeliveryPreferenceDecision,
  normalizeProviderCallbackPayload,
  normalizeProviderDeliveryRecord,
  signProviderCallback,
  subscribedTopics
} from "./index.mjs";

test("notification worker watches order and message events", () => {
  assert.ok(subscribedTopics.includes("order.status_changed"));
  assert.ok(subscribedTopics.includes("message.created"));
  assert.ok(subscribedTopics.includes("audit.retention_alerts"));
  assert.ok(subscribedTopics.includes("merchant.qualification_reviewed"));
  assert.ok(subscribedTopics.includes("notification.delivery_failed_alerts"));
  assert.ok(subscribedTopics.includes("notification.delivery_retries"));
  assert.ok(subscribedTopics.includes("notification.preferences_changed"));
});

test("notification payload is idempotent", () => {
  const notification = buildNotification({ type: "order.status_changed", order_id: "ord_1", target: { role: "user", id: "u1" }, body: "已接单" });
  assert.equal(notification.idempotency_key, "notify:order.status_changed:ord_1");
  assert.equal(notification.target_role, "user");

  const auditNotification = buildNotification({ id: "obe_audit_1", topic: "audit.retention_alerts", payload: { alert_count: 2, critical_count: 1 } });
  assert.equal(auditNotification.type, "audit.retention_alerts.emitted");
  assert.equal(auditNotification.target_role, "security");
  assert.equal(auditNotification.idempotency_key, "notify:audit.retention_alerts:obe_audit_1");

  const failureAlertNotification = buildNotification({
    id: "obe_notification_failure_1",
    topic: "notification.delivery_failed_alerts",
    payload: { failed_count: 2, channel: "wechat_subscribe", provider: "wechat_subscribe" }
  });
  assert.equal(failureAlertNotification.type, "notification.delivery_failed_alerts.emitted");
  assert.equal(failureAlertNotification.target_role, "security");
  assert.equal(failureAlertNotification.target_id, "notification_delivery");
  assert.match(failureAlertNotification.body, /failed=2/);
  assert.equal(failureAlertNotification.idempotency_key, "notify:notification.delivery_failed_alerts:obe_notification_failure_1");

  const retryNotification = buildNotification({
    id: "obe_notification_retry_1",
    topic: "notification.delivery_retries",
    payload: { scheduled_count: 1, channel: "wechat_subscribe", provider: "wechat_subscribe", retry_after_seconds: 300 }
  });
  assert.equal(retryNotification.type, "notification.delivery_retries.scheduled");
  assert.equal(retryNotification.target_role, "security");
  assert.equal(retryNotification.target_id, "notification_delivery");
  assert.match(retryNotification.body, /retry_after=300s/);
  assert.equal(retryNotification.idempotency_key, "notify:notification.delivery_retries:obe_notification_retry_1");
  const quietRetryNotification = buildNotification({
    id: "obe_notification_retry_quiet_1",
    topic: "notification.delivery_retries",
    payload: { scheduled_count: 1, delivery_status: "queued", error_code: "notification_quiet_window", channel: "push", provider: "push", retry_after_seconds: 540 }
  });
  assert.match(quietRetryNotification.body, /status=queued/);
  assert.match(quietRetryNotification.body, /error=notification_quiet_window/);

  const merchantNotification = buildNotification({
    id: "obe_mq_1",
    topic: "merchant.qualification_reviewed",
    payload: {
      merchant_id: "merchant_1",
      qualification_id: "mq_1",
      status: "approved"
    }
  });
  assert.equal(merchantNotification.type, "merchant.qualification_reviewed");
  assert.equal(merchantNotification.target_role, "merchant");
  assert.equal(merchantNotification.target_id, "merchant_1");
  assert.equal(merchantNotification.idempotency_key, "notify:merchant.qualification_reviewed:obe_mq_1");
});

test("notification records normalize to backend create payload", () => {
  const record = normalizeNotificationRecord({
    id: "obe_mq_1",
    topic: "merchant.qualification_reviewed",
    payload: {
      merchant_id: "merchant_1",
      qualification_id: "mq_1",
      status: "approved",
      reviewed_at: "2026-05-25T12:00:00Z"
    }
  });
  assert.equal(record.target_role, "merchant");
  assert.equal(record.target_id, "merchant_1");
  assert.equal(record.channel, "in_app");
  assert.equal(record.source_topic, "merchant.qualification_reviewed");
  assert.equal(record.source_event_id, "obe_mq_1");
  assert.equal(record.idempotency_key, "notify:merchant.qualification_reviewed:obe_mq_1");
  assert.equal(record.created_at, "2026-05-25T12:00:00Z");

  const delivery = normalizeNotificationDeliveryRecord({ id: "obe_mq_1", created_at: "2026-05-25T12:00:00Z" }, { id: "ntf_1", channel: "in_app", idempotency_key: record.idempotency_key });
  assert.equal(delivery.notification_id, "ntf_1");
  assert.equal(delivery.status, "delivered");
  assert.equal(delivery.idempotency_key, "delivery:notify:merchant.qualification_reviewed:obe_mq_1:in_app");
  assert.equal(delivery.delivered_at, "2026-05-25T12:00:00Z");
});

test("provider dispatches preserve notification payloads for retry events", () => {
  const dispatches = buildProviderDispatches({
    id: "obe_notification_retry_1",
    topic: "notification.delivery_retries",
    payload: {
      retry_policy: "wechat_subscribe:backoff:300s",
      retry_at: "2026-05-25T12:07:00Z",
      deliveries: [
        {
          id: "ntfd_failed_1",
          notification_id: "ntf_1",
          target_role: "merchant",
          target_id: "merchant_1",
          channel: "wechat_subscribe",
          provider: "wechat_subscribe"
        }
      ],
      notifications: [
        {
          id: "ntf_1",
          target_role: "merchant",
          target_id: "merchant_1",
          type: "merchant.qualification_reviewed",
          title: "商户资质审核结果",
          body: "资质审核已通过，系统已更新商户接单资格。"
        }
      ]
    }
  });
  assert.equal(dispatches.length, 1);
  assert.equal(dispatches[0].attempt, "retry");
  assert.equal(dispatches[0].notification_id, "ntf_1");
  assert.equal(dispatches[0].channel, "wechat_subscribe");
  assert.equal(dispatches[0].title, "商户资质审核结果");
  assert.match(dispatches[0].body, /资质审核已通过/);
  assert.match(dispatches[0].idempotency_key, /provider:wechat_subscribe:wechat_subscribe:ntf_1:/);

  const delivery = normalizeProviderDeliveryRecord(dispatches[0], { ok: true, status: "delivered", provider_message_id: "wx_msg_1" }, { attemptedAt: "2026-05-25T12:07:01.000Z" });
  assert.equal(delivery.notification_id, "ntf_1");
  assert.equal(delivery.status, "delivered");
  assert.equal(delivery.provider_message_id, "wx_msg_1");
  assert.equal(delivery.delivered_at, "2026-05-25T12:07:01.000Z");

  const quietRetryDispatches = buildProviderDispatches({
    id: "obe_notification_retry_quiet_1",
    topic: "notification.delivery_retries",
    payload: {
      delivery_status: "queued",
      error_code: "notification_quiet_window",
      retry_policy: "push_backoff_540s",
      retry_at: "2026-05-25T12:12:00Z",
      deliveries: [
        {
          id: "ntfd_quiet_1",
          notification_id: "ntf_2",
          target_role: "merchant",
          target_id: "merchant_1",
          channel: "push",
          provider: "push",
          status: "queued",
          error_code: "notification_quiet_window"
        }
      ],
      notifications: [
        {
          id: "ntf_2",
          target_role: "merchant",
          target_id: "merchant_1",
          type: "merchant.qualification_reviewed",
          title: "商户资质审核结果",
          body: "静默结束后补发。"
        }
      ]
    }
  });
  assert.equal(quietRetryDispatches.length, 1);
  assert.equal(quietRetryDispatches[0].attempt, "retry");
  assert.equal(quietRetryDispatches[0].channel, "push");
  assert.equal(quietRetryDispatches[0].source_delivery_id, "ntfd_quiet_1");
  assert.equal(quietRetryDispatches[0].retry_at, "2026-05-25T12:12:00Z");
});

test("provider dispatches apply channel templates and provider payloads", () => {
  const templates = {
    "merchant.qualification_reviewed": {
      wechat_subscribe: {
        template_id: "wx_tmpl_qualification",
        page: "pages/merchant/qualification/index",
        params: {
          thing1: "title",
          thing2: "body",
          phrase3: "target_id"
        }
      },
      sms: {
        template_code: "SMS_QUALIFICATION_RESULT",
        sign_name: "Infinitech",
        params: {
          merchant: "target_id",
          result: "body"
        }
      }
    }
  };
  const dispatches = buildProviderDispatches(
    {
      id: "obe_mq_1",
      topic: "merchant.qualification_reviewed",
      payload: { delivery_channels: ["wechat_subscribe", "sms"] }
    },
    {
      id: "ntf_1",
      target_role: "merchant",
      target_id: "merchant_1",
      type: "merchant.qualification_reviewed",
      title: "商户资质审核结果",
      body: "资质审核已通过",
      idempotency_key: "notify:merchant.qualification_reviewed:obe_mq_1"
    },
    { templates }
  );
  const wechat = dispatches.find((dispatch) => dispatch.channel === "wechat_subscribe");
  const sms = dispatches.find((dispatch) => dispatch.channel === "sms");
  assert.equal(wechat.template_id, "wx_tmpl_qualification");
  assert.equal(wechat.template_params.thing1, "商户资质审核结果");
  assert.equal(wechat.provider_payload.template_id, "wx_tmpl_qualification");
  assert.equal(wechat.provider_payload.data.thing2.value, "资质审核已通过");
  assert.equal(wechat.provider_payload.page, "pages/merchant/qualification/index");
  assert.equal(sms.template_id, "SMS_QUALIFICATION_RESULT");
  assert.equal(sms.provider_payload.template_code, "SMS_QUALIFICATION_RESULT");
  assert.equal(sms.provider_payload.sign_name, "Infinitech");
  assert.equal(sms.provider_payload.params.merchant, "merchant_1");

  const push = applyProviderTemplate({
    notification_id: "ntf_2",
    channel: "push",
    provider: "push",
    target_role: "merchant",
    target_id: "merchant_2",
    type: "notification.default",
    template_key: "notification.default",
    title: "默认通知",
    body: "默认内容",
    idempotency_key: "provider:push:push:ntf_2:obe_2:initial"
  });
  assert.equal(push.provider_payload.audience, "merchant_2");
  assert.equal(push.template_params.title, "默认通知");
});

test("provider callbacks are signed with the backend canonical contract", () => {
  const dispatch = {
    notification_id: "ntf_1",
    channel: "wechat_subscribe",
    provider: "wechat_subscribe",
    idempotency_key: "provider:wechat_subscribe:wechat_subscribe:ntf_1:obe_1:initial"
  };
  const callback = normalizeProviderCallbackPayload(
    dispatch,
    { ok: true, status: "delivered", provider_message_id: "wx_msg_1" },
    {
      attemptedAt: "2026-05-25T12:09:00.000Z",
      callbackAt: "2026-05-25T12:09:02.000Z",
      callbackSecret: "callback-secret"
    }
  );
  assert.equal(callback.notification_id, "ntf_1");
  assert.equal(callback.idempotency_key, "provider_callback:wechat_subscribe:wx_msg_1");
  assert.equal(callback.signature.length, 64);
  assert.equal(callback.signature, signProviderCallback({ ...callback, signature: "ignored" }, "callback-secret"));
});

test("provider preferences suppress disabled channels and quiet windows", async () => {
  const preferences = {
    default: {
      quiet_hours: {
        start: "22:00",
        end: "08:00",
        timezone_offset: "+08:00",
        channels: ["wechat_subscribe", "push"]
      }
    },
    "merchant:merchant_1": {
      disabled_channels: ["sms"]
    }
  };
  const quietDecision = notificationDeliveryPreferenceDecision(
    { channel: "wechat_subscribe", target_role: "merchant", target_id: "merchant_1", type: "merchant.qualification_reviewed" },
    { preferences, now: "2026-05-25T15:30:00Z" }
  );
  assert.equal(quietDecision.allowed, false);
  assert.equal(quietDecision.reason, "notification_quiet_window");
  assert.equal(quietDecision.retry_at, "2026-05-26T00:00:00.000Z");
  const disabledDecision = notificationDeliveryPreferenceDecision(
    { channel: "sms", target_role: "merchant", target_id: "merchant_1", type: "merchant.qualification_reviewed" },
    { preferences, now: "2026-05-25T03:00:00Z" }
  );
  assert.equal(disabledDecision.allowed, false);
  assert.equal(disabledDecision.reason, "notification_preference_disabled");

  const providerCalls = [];
  const dispatcher = createNotificationProviderDispatcher({
    channels: ["wechat_subscribe", "sms", "enterprise_wechat"],
    preferences,
    clock: () => new Date("2026-05-25T15:30:00Z"),
    adapters: {
      enterprise_wechat: async (dispatch) => {
        providerCalls.push(dispatch);
        return { ok: true, status: "delivered", provider_message_id: "ew_msg_1" };
      },
      wechat_subscribe: async () => {
        throw new Error("quiet channel must not call provider");
      },
      sms: async () => {
        throw new Error("disabled channel must not call provider");
      }
    }
  });
  const results = await dispatcher.deliver(
    { id: "obe_mq_1", topic: "merchant.qualification_reviewed" },
    {
      id: "ntf_1",
      target_role: "merchant",
      target_id: "merchant_1",
      type: "merchant.qualification_reviewed",
      title: "商户资质审核结果",
      body: "审核通过",
      idempotency_key: "notify:merchant.qualification_reviewed:obe_mq_1"
    }
  );
  assert.equal(providerCalls.length, 1);
  assert.equal(results.find((item) => item.dispatch.channel === "wechat_subscribe").delivery.status, "queued");
  assert.equal(results.find((item) => item.dispatch.channel === "wechat_subscribe").delivery.error_code, "notification_quiet_window");
  assert.equal(results.find((item) => item.dispatch.channel === "wechat_subscribe").delivery.retry_at, "2026-05-26T00:00:00.000Z");
  assert.equal(results.find((item) => item.dispatch.channel === "sms").delivery.status, "queued");
  assert.equal(results.find((item) => item.dispatch.channel === "sms").delivery.error_code, "notification_preference_disabled");
  assert.equal(results.find((item) => item.dispatch.channel === "enterprise_wechat").delivery.status, "delivered");
});

test("provider preferences can be loaded from backend preference records", async () => {
  const preferences = deliveryPreferencesFromRecords([
    {
      preference_key: "merchant:merchant_1:merchant.qualification_reviewed",
      enabled_channels: ["wechat_subscribe", "push"],
      disabled_channels: ["sms"],
      quiet_hours: {
        enabled: true,
        start: "22:00",
        end: "08:00",
        timezone_offset: "+08:00",
        channels: ["wechat_subscribe"]
      }
    }
  ]);
  assert.deepEqual(preferences["merchant:merchant_1:merchant.qualification_reviewed"].disabled_channels, ["sms"]);
  const decision = notificationDeliveryPreferenceDecision(
    { channel: "sms", target_role: "merchant", target_id: "merchant_1", type: "merchant.qualification_reviewed" },
    { preferences, now: "2026-05-25T03:00:00Z" }
  );
  assert.equal(decision.allowed, false);
  assert.equal(decision.reason, "notification_preference_disabled");
});

test("provider preference lookup failure queues external dispatches", async () => {
  const dispatcher = createNotificationProviderDispatcher({
    channels: ["sms"],
    preferenceResolver: async () => {
      throw new Error("preferences unavailable");
    },
    adapters: {
      sms: async () => {
        throw new Error("provider must not be called when preferences are unavailable");
      }
    }
  });
  const [result] = await dispatcher.deliver(
    { id: "obe_mq_1", topic: "merchant.qualification_reviewed" },
    {
      id: "ntf_1",
      target_role: "merchant",
      target_id: "merchant_1",
      type: "merchant.qualification_reviewed",
      title: "商户资质审核结果",
      body: "审核通过",
      idempotency_key: "notify:merchant.qualification_reviewed:obe_mq_1"
    }
  );
  assert.equal(result.delivery.status, "queued");
  assert.equal(result.delivery.error_code, "notification_preference_lookup_failed");
});

test("provider dispatcher records configured provider failure instead of pretending success", async () => {
  const dispatcher = createNotificationProviderDispatcher({
    channels: ["sms"],
    clock: () => new Date("2026-05-25T12:08:00Z"),
    env: {},
    fetchImpl: async () => {
      throw new Error("should not call unconfigured provider");
    }
  });
  const [result] = await dispatcher.deliver(
    { id: "obe_1", topic: "merchant.qualification_reviewed" },
    {
      id: "ntf_1",
      target_role: "merchant",
      target_id: "merchant_1",
      type: "merchant.qualification_reviewed",
      title: "商户资质审核结果",
      body: "审核通过",
      idempotency_key: "notify:merchant.qualification_reviewed:obe_1"
    }
  );
  assert.equal(result.delivery.channel, "sms");
  assert.equal(result.delivery.status, "failed");
  assert.equal(result.delivery.error_code, "provider_not_configured");
});

test("notification api client records in-app messages with worker authorization", async () => {
  const calls = [];
  const client = createNotificationApiClient({
    apiBaseUrl: "http://api.test",
    token: "worker-token",
    fetchImpl: async (url, options) => {
      calls.push({ url, options });
      if (url.endsWith("/deliveries")) {
        return {
          ok: true,
          async json() {
            return { success: true, data: { id: "ntfd_1", status: "delivered" } };
          }
        };
      }
      return {
        ok: true,
        async json() {
          return { success: true, data: { id: "ntf_1", status: "unread", channel: "in_app", idempotency_key: "notify:merchant.qualification_reviewed:obe_mq_1" } };
        }
      };
    }
  });
  const result = await client.recordNotification({
    id: "obe_mq_1",
    topic: "merchant.qualification_reviewed",
    payload: { merchant_id: "merchant_1", qualification_id: "mq_1", status: "approved" }
  });
  assert.equal(result.data.status, "unread");
  assert.equal(result.delivery.data.status, "delivered");
  assert.equal(calls[0].url, "http://api.test/api/notifications");
  assert.equal(calls[0].options.headers.Authorization, "Bearer worker-token");
  assert.equal(JSON.parse(calls[0].options.body).target_id, "merchant_1");
  assert.equal(calls[1].url, "http://api.test/api/notifications/ntf_1/deliveries");
  assert.equal(JSON.parse(calls[1].options.body).status, "delivered");
});

test("notification preference resolver reads exact backend preference keys", async () => {
  const calls = [];
  const resolver = createNotificationPreferenceResolver({
    apiBaseUrl: "http://api.test",
    token: "worker-token",
    fetchImpl: async (url, options) => {
      calls.push({ url, options });
      const key = new URL(url).searchParams.get("preference_key");
      return {
        ok: true,
        async json() {
          return {
            success: true,
            data: key === "merchant:merchant_1:merchant.qualification_reviewed"
              ? [{ preference_key: key, disabled_channels: ["sms"] }]
              : []
          };
        }
      };
    }
  });
  const preferences = await resolver([
    { target_role: "merchant", target_id: "merchant_1", type: "merchant.qualification_reviewed", channel: "sms" }
  ]);
  assert.equal(preferences["merchant:merchant_1:merchant.qualification_reviewed"].disabled_channels[0], "sms");
  assert.equal(calls.length, 5);
  assert.equal(new URL(calls[0].url).searchParams.get("preference_key"), "default");
  assert.equal(calls[0].options.headers.Authorization, "Bearer worker-token");
});

test("notification preference resolver caches preference key lookups within ttl", async () => {
  const calls = [];
  let now = new Date("2026-05-25T12:00:00Z");
  const dispatches = [
    { target_role: "merchant", target_id: "merchant_1", type: "merchant.qualification_reviewed", channel: "sms" }
  ];
  const resolver = createNotificationPreferenceResolver({
    apiBaseUrl: "http://api.test",
    token: "worker-token",
    cacheTtlMs: 60000,
    cacheStaleMs: 300000,
    clock: () => now,
    fetchImpl: async (url) => {
      calls.push(url);
      const key = new URL(url).searchParams.get("preference_key");
      return {
        ok: true,
        async json() {
          return {
            success: true,
            data: key === "merchant:merchant_1:merchant.qualification_reviewed"
              ? [{ preference_key: key, disabled_channels: ["sms"] }]
              : []
          };
        }
      };
    }
  });

  await resolver(dispatches);
  const cachedPreferences = await resolver(dispatches);
  assert.equal(cachedPreferences["merchant:merchant_1:merchant.qualification_reviewed"].disabled_channels[0], "sms");
  assert.equal(calls.length, 5);

  now = new Date("2026-05-25T12:01:01Z");
  await resolver(dispatches);
  assert.equal(calls.length, 10);
});

test("notification preference resolver uses stale cache on lookup failure", async () => {
  const calls = [];
  let now = new Date("2026-05-25T12:00:00Z");
  let fail = false;
  const dispatches = [
    { target_role: "merchant", target_id: "merchant_1", type: "merchant.qualification_reviewed", channel: "sms" }
  ];
  const resolver = createNotificationPreferenceResolver({
    apiBaseUrl: "http://api.test",
    token: "worker-token",
    cacheTtlMs: 10000,
    cacheStaleMs: 50000,
    clock: () => now,
    fetchImpl: async (url) => {
      calls.push(url);
      if (fail) {
        return {
          ok: false,
          status: 503,
          async text() {
            return "preferences down";
          }
        };
      }
      const key = new URL(url).searchParams.get("preference_key");
      return {
        ok: true,
        async json() {
          return {
            success: true,
            data: key === "merchant:merchant_1:merchant.qualification_reviewed"
              ? [{ preference_key: key, disabled_channels: ["sms"] }]
              : []
          };
        }
      };
    }
  });

  await resolver(dispatches);
  now = new Date("2026-05-25T12:00:20Z");
  fail = true;
  const stalePreferences = await resolver(dispatches);
  assert.equal(stalePreferences["merchant:merchant_1:merchant.qualification_reviewed"].disabled_channels[0], "sms");
  assert.equal(calls.length, 10);

  now = new Date("2026-05-25T12:01:01Z");
  await assert.rejects(() => resolver(dispatches), /notification preferences api failed: 503 preferences down/);
});

test("notification preference change events invalidate resolver cache", async () => {
  const calls = [];
  let disabledChannels = ["sms"];
  const dispatches = [
    { target_role: "merchant", target_id: "merchant_1", type: "merchant.qualification_reviewed", channel: "sms" }
  ];
  const resolver = createNotificationPreferenceResolver({
    apiBaseUrl: "http://api.test",
    token: "worker-token",
    cacheTtlMs: 60000,
    fetchImpl: async (url) => {
      calls.push(url);
      const key = new URL(url).searchParams.get("preference_key");
      return {
        ok: true,
        async json() {
          return {
            success: true,
            data: key === "merchant:merchant_1:merchant.qualification_reviewed"
              ? [{ preference_key: key, disabled_channels: disabledChannels }]
              : []
          };
        }
      };
    }
  });

  const first = await resolver(dispatches);
  assert.deepEqual(first["merchant:merchant_1:merchant.qualification_reviewed"].disabled_channels, ["sms"]);
  disabledChannels = ["push"];
  const cached = await resolver(dispatches);
  assert.deepEqual(cached["merchant:merchant_1:merchant.qualification_reviewed"].disabled_channels, ["sms"]);
  assert.equal(calls.length, 5);

  const event = {
    topic: "notification.preferences_changed",
    payload: {
      preference_key: "merchant:merchant_1:merchant.qualification_reviewed",
      target_role: "merchant",
      target_id: "merchant_1",
      notification_type: "merchant.qualification_reviewed"
    }
  };
  assert.deepEqual(notificationPreferenceInvalidationKeys(event), ["merchant:merchant_1:merchant.qualification_reviewed"]);
  assert.equal(resolver.invalidate(event).invalidated_count, 1);
  const refreshed = await resolver(dispatches);
  assert.deepEqual(refreshed["merchant:merchant_1:merchant.qualification_reviewed"].disabled_channels, ["push"]);
  assert.equal(calls.length, 6);
});

test("notification consumer handles preference change events as cache invalidations", async () => {
  const cache = new Map([
    ["merchant:merchant_1:merchant.qualification_reviewed", { records: [], expires_at: Date.now() + 60000, stale_until: Date.now() + 120000 }]
  ]);
  const invalidation = invalidateNotificationPreferenceCache(cache, {
    topic: "notification.preferences_changed",
    payload: { preference_key: "merchant:merchant_1:merchant.qualification_reviewed" }
  });
  assert.equal(invalidation.invalidated_count, 1);
  assert.equal(cache.has("merchant:merchant_1:merchant.qualification_reviewed"), false);

  const seen = [];
  const consumer = createNotificationConsumer({
    apiClient: {
      invalidateNotificationPreferences: async (event) => {
        seen.push(event.topic);
        return { success: true, data: { invalidated_count: 1 } };
      },
      recordNotification: async () => {
        throw new Error("preference invalidation must not create an in-app notification");
      }
    }
  });
  const result = await consumer({
    id: "obe_notification_preference_1",
    topic: "notification.preferences_changed",
    payload: { preference_key: "merchant:merchant_1:merchant.qualification_reviewed" }
  });
  assert.equal(result.result.data.invalidated_count, 1);
  assert.deepEqual(seen, ["notification.preferences_changed"]);
});

test("notification api client sends configured provider channels and records provider receipt", async () => {
  const calls = [];
  const providerCalls = [];
  const client = createNotificationApiClient({
    apiBaseUrl: "http://api.test",
    token: "worker-token",
    providerChannels: ["wechat_subscribe"],
    fetchDeliveryPreferences: false,
    clock: () => new Date("2026-05-25T12:09:00Z"),
    providerAdapters: {
      wechat_subscribe: async (dispatch) => {
        providerCalls.push(dispatch);
        return { ok: true, status: "delivered", provider_message_id: "wx_msg_1" };
      }
    },
    fetchImpl: async (url, options) => {
      calls.push({ url, options });
      if (url.endsWith("/deliveries")) {
        return {
          ok: true,
          async json() {
            return { success: true, data: { id: `ntfd_${calls.length}`, status: JSON.parse(options.body).status } };
          }
        };
      }
      return {
        ok: true,
        async json() {
          return { success: true, data: { id: "ntf_1", status: "unread", channel: "in_app", title: "商户资质审核结果", body: "审核通过", target_role: "merchant", target_id: "merchant_1", type: "merchant.qualification_reviewed", idempotency_key: "notify:merchant.qualification_reviewed:obe_mq_1" } };
        }
      };
    }
  });
  const result = await client.recordNotification({
    id: "obe_mq_1",
    topic: "merchant.qualification_reviewed",
    payload: { merchant_id: "merchant_1", qualification_id: "mq_1", status: "approved" }
  });
  assert.equal(providerCalls.length, 1);
  assert.equal(providerCalls[0].channel, "wechat_subscribe");
  assert.equal(providerCalls[0].notification_id, "ntf_1");
  assert.equal(calls.length, 3);
  assert.equal(JSON.parse(calls[1].options.body).channel, "in_app");
  assert.equal(JSON.parse(calls[2].options.body).channel, "wechat_subscribe");
  assert.equal(JSON.parse(calls[2].options.body).provider_message_id, "wx_msg_1");
  assert.equal(result.provider_deliveries[0].delivery.status, "delivered");
});

test("notification api client applies backend delivery preferences before provider calls", async () => {
  const calls = [];
  const providerCalls = [];
  const client = createNotificationApiClient({
    apiBaseUrl: "http://api.test",
    token: "worker-token",
    providerChannels: ["sms", "wechat_subscribe"],
    clock: () => new Date("2026-05-25T12:09:00Z"),
    providerAdapters: {
      sms: async () => {
        throw new Error("disabled backend preference must not call sms provider");
      },
      wechat_subscribe: async (dispatch) => {
        providerCalls.push(dispatch);
        return { ok: true, status: "delivered", provider_message_id: "wx_msg_1" };
      }
    },
    fetchImpl: async (url, options) => {
      calls.push({ url, options });
      if (url.includes("/api/admin/notification-preferences")) {
        const key = new URL(url).searchParams.get("preference_key");
        return {
          ok: true,
          async json() {
            return {
              success: true,
              data: key === "merchant:merchant_1:merchant.qualification_reviewed"
                ? [{ preference_key: key, disabled_channels: ["sms"] }]
                : []
            };
          }
        };
      }
      if (url.endsWith("/deliveries")) {
        return {
          ok: true,
          async json() {
            return { success: true, data: { id: `ntfd_${calls.length}`, status: JSON.parse(options.body).status } };
          }
        };
      }
      return {
        ok: true,
        async json() {
          return { success: true, data: { id: "ntf_1", status: "unread", channel: "in_app", title: "商户资质审核结果", body: "审核通过", target_role: "merchant", target_id: "merchant_1", type: "merchant.qualification_reviewed", idempotency_key: "notify:merchant.qualification_reviewed:obe_mq_1" } };
        }
      };
    }
  });
  const result = await client.recordNotification({
    id: "obe_mq_1",
    topic: "merchant.qualification_reviewed",
    payload: { merchant_id: "merchant_1", qualification_id: "mq_1", status: "approved" }
  });
  assert.equal(providerCalls.length, 1);
  assert.equal(providerCalls[0].channel, "wechat_subscribe");
  assert.equal(result.provider_deliveries.find((item) => item.dispatch.channel === "sms").delivery.status, "queued");
  assert.equal(result.provider_deliveries.find((item) => item.dispatch.channel === "sms").delivery.error_code, "notification_preference_disabled");
  assert.equal(calls.filter((call) => call.url.includes("/api/admin/notification-preferences")).length, 5);
});

test("notification api client executes provider retries against original notification ids", async () => {
  const calls = [];
  const providerCalls = [];
  const client = createNotificationApiClient({
    apiBaseUrl: "http://api.test",
    token: "worker-token",
    fetchDeliveryPreferences: false,
    clock: () => new Date("2026-05-25T12:10:00Z"),
    providerAdapters: {
      wechat_subscribe: async (dispatch) => {
        providerCalls.push(dispatch);
        return { ok: true, status: "delivered", provider_message_id: "wx_retry_msg_1" };
      }
    },
    fetchImpl: async (url, options) => {
      calls.push({ url, options });
      if (url.endsWith("/deliveries")) {
        return {
          ok: true,
          async json() {
            return { success: true, data: { id: `ntfd_${calls.length}`, status: JSON.parse(options.body).status } };
          }
        };
      }
      return {
        ok: true,
        async json() {
          return { success: true, data: { id: "ntf_retry_notice_1", status: "unread", channel: "in_app", target_role: "security", target_id: "notification_delivery", type: "notification.delivery_retries.scheduled", idempotency_key: "notify:notification.delivery_retries:obe_notification_retry_1" } };
        }
      };
    }
  });
  await client.recordNotification({
    id: "obe_notification_retry_1",
    topic: "notification.delivery_retries",
    payload: {
      scheduled_count: 1,
      channel: "wechat_subscribe",
      provider: "wechat_subscribe",
      retry_after_seconds: 300,
      deliveries: [
        { id: "ntfd_failed_1", notification_id: "ntf_1", target_role: "merchant", target_id: "merchant_1", channel: "wechat_subscribe", provider: "wechat_subscribe" }
      ],
      notifications: [
        { id: "ntf_1", target_role: "merchant", target_id: "merchant_1", type: "merchant.qualification_reviewed", title: "商户资质审核结果", body: "审核通过" }
      ]
    }
  });
  assert.equal(providerCalls.length, 1);
  assert.equal(providerCalls[0].attempt, "retry");
  assert.equal(providerCalls[0].notification_id, "ntf_1");
  assert.equal(providerCalls[0].body, "审核通过");
  assert.equal(calls.length, 3);
  assert.match(calls[1].url, /ntf_retry_notice_1\/deliveries$/);
  assert.match(calls[2].url, /ntf_1\/deliveries$/);
  assert.equal(JSON.parse(calls[2].options.body).provider_message_id, "wx_retry_msg_1");
});

test("quiet-window retry scheduler scans due queued deliveries", async () => {
  const payload = normalizeQuietWindowRetrySchedulePayload({
    channel: "push",
    provider: "push",
    limit: "500",
    retry_after_seconds: "-1",
    now: "2026-05-25T12:10:00.000Z"
  });
  assert.equal(payload.channel, "push");
  assert.equal(payload.limit, 100);
  assert.equal(payload.retry_after_seconds, 0);

  const calls = [];
  const scheduler = createQuietWindowRetryScheduler({
    apiBaseUrl: "http://api.test",
    token: "worker-token",
    clock: () => new Date("2026-05-25T12:10:00.000Z"),
    channel: "push",
    provider: "push",
    limit: 10,
    fetchImpl: async (url, options) => {
      calls.push({ url, options });
      return {
        ok: true,
        async json() {
          return { success: true, data: { schedule: { status: "scheduled", scheduled_count: 1, delivery_status: "queued" } } };
        }
      };
    }
  });
  const result = await scheduler.scheduleQuietWindowRetries();
  assert.equal(result.data.schedule.delivery_status, "queued");
  assert.equal(calls[0].url, "http://api.test/api/admin/notification-deliveries/quiet-window-retries/schedule");
  assert.equal(calls[0].options.headers.Authorization, "Bearer worker-token");
  assert.equal(JSON.parse(calls[0].options.body).now, "2026-05-25T12:10:00.000Z");
  assert.equal(JSON.parse(calls[0].options.body).channel, "push");

  const ticks = [];
  let storedTimer;
  const loop = createQuietWindowRetryLoop({
    scheduler,
    intervalMs: 2000,
    setIntervalImpl: (fn, ms) => {
      storedTimer = { fn, ms };
      return "timer_1";
    },
    clearIntervalImpl: (timer) => ticks.push(["clear", timer]),
    onResult: (item) => ticks.push(["result", item.data.schedule.scheduled_count])
  });
  assert.equal(storedTimer.ms, 2000);
  await loop.tick();
  loop.stop();
  assert.deepEqual(ticks, [["result", 1], ["clear", "timer_1"]]);
});

test("notification consumer ignores duplicate outbox deliveries", async () => {
  let handled = 0;
  const consumer = createNotificationConsumer({
    clock: () => new Date("2026-05-22T12:00:00.000Z"),
    handler: async () => {
      handled += 1;
      return { pushed: true };
    }
  });
  const event = { id: "obe_notify_1", topic: "order.status_changed", idempotency_key: "notify:order.status_changed:ord_1" };
  assert.equal((await consumer(event)).status, "processed");
  assert.equal((await consumer({ ...event, id: "obe_notify_replay" })).status, "duplicate");
  assert.equal(handled, 1);
  assert.equal(consumer.ledger.snapshot()[0].consumer_name, "notification-worker");
});
