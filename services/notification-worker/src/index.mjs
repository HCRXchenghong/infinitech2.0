import { createHash, createHmac } from "node:crypto";
import { createIdempotentConsumer } from "../../../packages/domain-core/src/index.mjs";

export const workerName = "notification-worker";
export const subscribedTopics = ["order.status_changed", "message.created", "dispatch.assigned", "rtc.call_status_changed", "audit.retention_alerts", "merchant.qualification_reviewed", "notification.delivery_failed_alerts", "notification.delivery_retries", "notification.preferences_changed"];
export const providerDeliveryChannels = ["wechat_subscribe", "sms", "enterprise_wechat", "push"];

function compact(value, fallback = "") {
  const normalized = String(value ?? "").trim();
  return normalized || fallback;
}

function shortHash(value) {
  return createHash("sha256").update(String(value ?? "")).digest("hex").slice(0, 12);
}

function eventPayload(event = {}) {
  if (event.payload && typeof event.payload === "object") return event.payload;
  if (event.notification && typeof event.notification === "object") return event.notification;
  return {};
}

function eventType(event = {}) {
  const payload = eventPayload(event);
  return compact(event.type || payload.type || event.event_type || payload.event_type);
}

function parseChannelList(value) {
  if (Array.isArray(value)) {
    return value.flatMap((item) => parseChannelList(item));
  }
  return String(value ?? "")
    .split(",")
    .map((item) => item.trim())
    .filter(Boolean);
}

function normalizeProviderChannel(channel) {
  const normalized = compact(channel);
  return providerDeliveryChannels.includes(normalized) ? normalized : "";
}

function uniqueProviderChannels(channels = []) {
  const seen = new Set();
  const normalized = [];
  for (const channel of channels) {
    const value = normalizeProviderChannel(channel);
    if (!value || seen.has(value)) continue;
    seen.add(value);
    normalized.push(value);
  }
  return normalized;
}

function providerEnvKey(channel, suffix) {
  return `NOTIFICATION_PROVIDER_${String(channel || "").toUpperCase().replace(/[^A-Z0-9]+/g, "_")}_${suffix}`;
}

function providerEndpointFor(channel, env = process.env) {
  const fromJSON = (() => {
    try {
      const endpoints = JSON.parse(env.NOTIFICATION_PROVIDER_ENDPOINTS || "{}");
      return endpoints?.[channel] || "";
    } catch {
      return "";
    }
  })();
  return compact(fromJSON || env[providerEnvKey(channel, "ENDPOINT")]);
}

function providerTokenFor(channel, env = process.env) {
  return compact(env[providerEnvKey(channel, "TOKEN")] || env.NOTIFICATION_PROVIDER_TOKEN);
}

function parseProviderJSON(value, fallback = {}) {
  if (value && typeof value === "object" && !Array.isArray(value)) return value;
  try {
    const parsed = JSON.parse(String(value || ""));
    return parsed && typeof parsed === "object" && !Array.isArray(parsed) ? parsed : fallback;
  } catch {
    return fallback;
  }
}

export function normalizeProviderTemplates(value = {}) {
  const raw = parseProviderJSON(value, {});
  const templates = {};
  for (const [key, config] of Object.entries(raw)) {
    if (!key || !config || typeof config !== "object" || Array.isArray(config)) continue;
    templates[compact(key)] = config;
  }
  return templates;
}

function providerTemplatesForOptions(options = {}) {
  return normalizeProviderTemplates(options.templates || options.env?.NOTIFICATION_PROVIDER_TEMPLATES || process.env.NOTIFICATION_PROVIDER_TEMPLATES || {});
}

export function normalizeDeliveryPreferences(value = {}) {
  const raw = parseProviderJSON(value, {});
  const preferences = {};
  for (const [key, config] of Object.entries(raw)) {
    if (!key || !config || typeof config !== "object" || Array.isArray(config)) continue;
    preferences[compact(key)] = config;
  }
  return preferences;
}

function deliveryPreferencesForOptions(options = {}) {
  return normalizeDeliveryPreferences(options.preferences || options.env?.NOTIFICATION_DELIVERY_PREFERENCES || process.env.NOTIFICATION_DELIVERY_PREFERENCES || {});
}

function preferenceKeyForTarget(targetRole = "", targetID = "", type = "") {
  const role = compact(targetRole);
  const id = compact(targetID);
  const notificationType = compact(type);
  if (!role && !id && !notificationType) return "default";
  if (!role && notificationType) return `type:${notificationType}`;
  if (!id) return role;
  if (!notificationType) return `${role}:${id}`;
  return `${role}:${id}:${notificationType}`;
}

function preferenceKeyForRecord(record = {}) {
  return compact(
    record.preference_key || record.preferenceKey,
    preferenceKeyForTarget(record.target_role || record.targetRole, record.target_id || record.targetId, record.notification_type || record.notificationType || record.type)
  );
}

function preferenceKeysForDispatch(dispatch = {}) {
  const targetRole = compact(dispatch.target_role);
  const targetID = compact(dispatch.target_id);
  const type = compact(dispatch.type);
  return [
    "default",
    targetRole ? preferenceKeyForTarget(targetRole) : "",
    targetRole && targetID ? preferenceKeyForTarget(targetRole, targetID) : "",
    type ? preferenceKeyForTarget("", "", type) : "",
    targetRole && targetID && type ? preferenceKeyForTarget(targetRole, targetID, type) : ""
  ].filter(Boolean);
}

function uniquePreferenceKeysForDispatches(dispatches = []) {
  const seen = new Set();
  const keys = [];
  for (const dispatch of dispatches) {
    for (const key of preferenceKeysForDispatch(dispatch)) {
      if (seen.has(key)) continue;
      seen.add(key);
      keys.push(key);
    }
  }
  return keys;
}

function isNotificationPreferenceChangedEvent(event = {}) {
  return compact(event.topic) === "notification.preferences_changed" || eventType(event) === "notification.preferences.changed";
}

export function notificationPreferenceInvalidationKeys(event = {}) {
  const payload = eventPayload(event);
  const source = [
    payload.preference_keys,
    payload.invalidate_keys,
    payload.preference_key,
    event.preference_keys,
    event.invalidate_keys,
    event.preference_key
  ];
  const keys = source.flatMap((value) => Array.isArray(value) ? value : parseChannelList(value));
  const derived = preferenceKeyForTarget(payload.target_role || event.target_role, payload.target_id || event.target_id, payload.notification_type || event.notification_type);
  if (derived) keys.push(derived);
  return [...new Set(keys.map((key) => compact(key)).filter(Boolean))];
}

export function invalidateNotificationPreferenceCache(cache, eventOrKeys = {}) {
  const keys = Array.isArray(eventOrKeys) ? eventOrKeys.map((key) => compact(key)).filter(Boolean) : notificationPreferenceInvalidationKeys(eventOrKeys);
  let invalidated = 0;
  if (cache && typeof cache.delete === "function") {
    for (const key of keys) {
      if (cache.delete(key)) {
        invalidated += 1;
      }
    }
  }
  return {
    status: keys.length > 0 ? "invalidated" : "skipped",
    invalidated_count: invalidated,
    preference_keys: keys
  };
}

export function deliveryPreferencesFromRecords(records = []) {
  const source = Array.isArray(records) ? records : Array.isArray(records?.data) ? records.data : [];
  const preferences = {};
  for (const record of source) {
    if (!record || typeof record !== "object" || Array.isArray(record)) continue;
    const key = preferenceKeyForRecord(record);
    if (!key) continue;
    const rule = {};
    if (record.enabled_channels || record.enabledChannels) {
      rule.enabled_channels = uniqueProviderChannels(parseChannelList(record.enabled_channels || record.enabledChannels));
    }
    if (record.disabled_channels || record.disabledChannels) {
      rule.disabled_channels = uniqueProviderChannels(parseChannelList(record.disabled_channels || record.disabledChannels));
    }
    const quiet = record.quiet_hours || record.quietHours || {};
    if (quiet && typeof quiet === "object" && !Array.isArray(quiet)) {
      rule.quiet_hours = {
        enabled: quiet.enabled,
        start: compact(quiet.start || quiet.start_time || quiet.startTime),
        end: compact(quiet.end || quiet.end_time || quiet.endTime),
        timezone_offset: compact(quiet.timezone_offset || quiet.timezoneOffset || quiet.utc_offset || quiet.utcOffset),
        channels: uniqueProviderChannels(parseChannelList(quiet.channels || quiet.apply_channels || quiet.applyChannels)),
        exempt_types: parseChannelList(quiet.exempt_types || quiet.exemptTypes),
        status: compact(quiet.status)
      };
    }
    preferences[key] = rule;
  }
  return normalizeDeliveryPreferences(preferences);
}

function mergeDeliveryPreferences(...sources) {
  const merged = {};
  for (const source of sources) {
    const preferences = normalizeDeliveryPreferences(source);
    for (const [key, rule] of Object.entries(preferences)) {
      merged[key] = mergePreferenceRule(merged[key] || {}, rule || {});
    }
  }
  return merged;
}

function notificationPreferenceEndpoint(apiBaseUrl, key) {
  const url = new URL(`${apiBaseUrl}/api/admin/notification-preferences`);
  url.searchParams.set("preference_key", key);
  url.searchParams.set("limit", "1");
  return url.toString();
}

export function createNotificationPreferenceResolver(options = {}) {
  const apiBaseUrl = String(options.apiBaseUrl || options.env?.API_BASE_URL || process.env.API_BASE_URL || "").replace(/\/+$/, "");
  const fetchImpl = options.fetchImpl || globalThis.fetch;
  const token = compact(options.token || options.env?.NOTIFICATION_WORKER_TOKEN || process.env.NOTIFICATION_WORKER_TOKEN);
  const env = options.env || process.env;
  const clock = options.clock || (() => new Date());
  const rawCacheTtlMs = Number(options.cacheTtlMs ?? env.NOTIFICATION_PREFERENCE_CACHE_TTL_MS ?? 30000);
  const rawCacheStaleMs = Number(options.cacheStaleMs ?? env.NOTIFICATION_PREFERENCE_CACHE_STALE_MS ?? 300000);
  const rawCacheMaxKeys = Number(options.cacheMaxKeys ?? env.NOTIFICATION_PREFERENCE_CACHE_MAX_KEYS ?? 500);
  const cacheTtlMs = Number.isFinite(rawCacheTtlMs) && rawCacheTtlMs >= 0 ? rawCacheTtlMs : 30000;
  const cacheStaleMs = Number.isFinite(rawCacheStaleMs) && rawCacheStaleMs >= 0 ? rawCacheStaleMs : 300000;
  const cacheMaxKeys = Number.isFinite(rawCacheMaxKeys) && rawCacheMaxKeys > 0 ? rawCacheMaxKeys : 500;
  const cache = options.cache instanceof Map ? options.cache : new Map();
  if (!apiBaseUrl || typeof fetchImpl !== "function") {
    return async () => ({});
  }
  const nowMillis = () => {
    const value = clock();
    const date = value instanceof Date ? value : new Date(value || Date.now());
    return Number.isFinite(date.getTime()) ? date.getTime() : Date.now();
  };
  const cachedRecords = (key, now, allowStale = false) => {
    const entry = cache.get(key);
    if (!entry) return null;
    if (entry.expires_at > now || allowStale && entry.stale_until > now) {
      cache.delete(key);
      cache.set(key, entry);
      return entry.records;
    }
    return null;
  };
  const rememberRecords = (key, records, now) => {
    if (cacheTtlMs <= 0 && cacheStaleMs <= 0) return;
    while (cache.size >= cacheMaxKeys) {
      const oldest = cache.keys().next().value;
      if (!oldest) break;
      cache.delete(oldest);
    }
    cache.set(key, {
      records: Array.isArray(records) ? records : [],
      expires_at: now + cacheTtlMs,
      stale_until: now + cacheTtlMs + cacheStaleMs
    });
  };
  async function resolveNotificationPreferences(dispatches = [], context = {}) {
    const keys = uniquePreferenceKeysForDispatches(dispatches);
    if (keys.length === 0) return {};
    const headers = { ...(context.headers || {}), "Content-Type": "application/json" };
    if (token && !headers.Authorization) {
      headers.Authorization = token.startsWith("Bearer ") ? token : `Bearer ${token}`;
    }
    const records = [];
    for (const key of keys) {
      const now = nowMillis();
      const fresh = cachedRecords(key, now);
      if (fresh) {
        records.push(...fresh);
        continue;
      }
      try {
        const response = await fetchImpl(notificationPreferenceEndpoint(apiBaseUrl, key), {
          method: "GET",
          headers
        });
        if (!response.ok) {
          const text = await response.text();
          throw new Error(`notification preferences api failed: ${response.status} ${text}`);
        }
        const payload = await response.json();
        const data = Array.isArray(payload?.data) ? payload.data : Array.isArray(payload) ? payload : [];
        rememberRecords(key, data, nowMillis());
        records.push(...data);
      } catch (error) {
        const stale = cachedRecords(key, nowMillis(), true);
        if (stale) {
          records.push(...stale);
          continue;
        }
        throw error;
      }
    }
    return deliveryPreferencesFromRecords(records);
  }
  resolveNotificationPreferences.invalidate = (eventOrKeys) => invalidateNotificationPreferenceCache(cache, eventOrKeys);
  resolveNotificationPreferences.cache = cache;
  return resolveNotificationPreferences;
}

function providerChannelsForEvent(event = {}, options = {}) {
  const payload = eventPayload(event);
  const requested = payload.delivery_channels || payload.channels || event.delivery_channels || event.channels || options.channels || options.env?.NOTIFICATION_PROVIDER_CHANNELS || process.env.NOTIFICATION_PROVIDER_CHANNELS || "";
  return uniqueProviderChannels(parseChannelList(requested));
}

function notificationSnapshotsById(payload = {}) {
  const source = Array.isArray(payload.notifications) ? payload.notifications : [];
  return source.reduce((snapshots, notification) => {
    const id = compact(notification?.id || notification?.notification_id);
    if (id) snapshots.set(id, notification);
    return snapshots;
  }, new Map());
}

function providerForChannel(channel, event = {}) {
  const payload = eventPayload(event);
  const providers = payload.providers && typeof payload.providers === "object" ? payload.providers : {};
  return compact(providers[channel] || payload.provider || channel, channel);
}

function providerTemplateFor(dispatch = {}, templates = {}) {
  const channel = compact(dispatch.channel);
  const keys = [dispatch.template_key, dispatch.type, "notification.default"].map((key) => compact(key)).filter(Boolean);
  for (const key of keys) {
    const byType = templates[key];
    if (byType && typeof byType === "object" && !Array.isArray(byType)) {
      if (byType[channel] && typeof byType[channel] === "object") return byType[channel];
      if (byType[dispatch.provider] && typeof byType[dispatch.provider] === "object") return byType[dispatch.provider];
      if (byType.template_id || byType.template_code || byType.id) return byType;
    }
    const byChannel = templates[channel];
    if (byChannel && typeof byChannel === "object" && !Array.isArray(byChannel)) {
      if (byChannel[key] && typeof byChannel[key] === "object") return byChannel[key];
      if (byChannel.template_id || byChannel.template_code || byChannel.id) return byChannel;
    }
  }
  return {};
}

function valueAtPath(source = {}, path = "") {
  const parts = String(path || "").replace(/^\$\.?/, "").split(".").filter(Boolean);
  if (parts.length === 0) return "";
  let current = source;
  for (const part of parts) {
    if (!current || typeof current !== "object") return "";
    current = current[part];
  }
  return current ?? "";
}

function templateParamsFor(dispatch = {}, template = {}) {
  const defaults = {
    title: dispatch.title,
    body: dispatch.body,
    type: dispatch.type,
    target_id: dispatch.target_id
  };
  const mapping = template.params || template.variables || template.data || {};
  const staticParams = template.static_params || template.staticParams || {};
  const params = {};
  if (mapping && typeof mapping === "object" && !Array.isArray(mapping)) {
    for (const [name, source] of Object.entries(mapping)) {
      if (!name) continue;
      const value = typeof source === "string" ? valueAtPath(dispatch, source) : source;
      params[name] = compact(value);
    }
  }
  for (const [name, value] of Object.entries(staticParams && typeof staticParams === "object" ? staticParams : {})) {
    if (name) params[name] = compact(value);
  }
  return Object.keys(params).length > 0 ? params : defaults;
}

function wechatTemplateData(params = {}) {
  return Object.fromEntries(Object.entries(params).map(([key, value]) => [key, { value: compact(value) }]));
}

function providerPayloadFor(dispatch = {}, template = {}, params = {}) {
  const metadata = {
    notification_id: compact(dispatch.notification_id),
    idempotency_key: compact(dispatch.idempotency_key),
    attempt: compact(dispatch.attempt),
    type: compact(dispatch.type),
    template_key: compact(dispatch.template_key)
  };
  const templateID = compact(template.template_id || template.templateId || template.template_code || template.templateCode || template.id);
  if (dispatch.channel === "wechat_subscribe") {
    return {
      touser: compact(dispatch.target_id),
      template_id: templateID,
      page: compact(template.page || template.miniprogram_page || template.miniprogramPage),
      data: wechatTemplateData(params),
      miniprogram_state: compact(template.miniprogram_state || template.miniprogramState),
      lang: compact(template.lang, "zh_CN"),
      metadata
    };
  }
  if (dispatch.channel === "sms") {
    return {
      phone: compact(dispatch.target_id),
      template_code: templateID,
      sign_name: compact(template.sign_name || template.signName),
      params,
      metadata
    };
  }
  if (dispatch.channel === "enterprise_wechat") {
    return {
      touser: compact(dispatch.target_id),
      agentid: compact(template.agent_id || template.agentId),
      msgtype: compact(template.msgtype, templateID ? "template_card" : "text"),
      template_id: templateID,
      text: { content: compact(dispatch.body) },
      params,
      metadata
    };
  }
  if (dispatch.channel === "push") {
    return {
      audience: compact(dispatch.target_id),
      title: compact(dispatch.title),
      body: compact(dispatch.body),
      template_id: templateID,
      extras: { ...params, notification_id: metadata.notification_id, type: metadata.type },
      metadata
    };
  }
  return {
    recipient: { role: compact(dispatch.target_role), id: compact(dispatch.target_id) },
    title: compact(dispatch.title),
    body: compact(dispatch.body),
    template_id: templateID,
    params,
    metadata
  };
}

export function applyProviderTemplate(dispatch = {}, options = {}) {
  const templates = options.templates ? normalizeProviderTemplates(options.templates) : providerTemplatesForOptions(options);
  const template = providerTemplateFor(dispatch, templates);
  const templateID = compact(template.template_id || template.templateId || template.template_code || template.templateCode || template.id);
  const params = templateParamsFor(dispatch, template);
  return {
    ...dispatch,
    template_id: templateID,
    template_params: params,
    provider_payload: providerPayloadFor({ ...dispatch, template_id: templateID }, template, params)
  };
}

function applyProviderTemplates(dispatches = [], options = {}) {
  return dispatches.map((dispatch) => applyProviderTemplate(dispatch, options));
}

function normalizePreferenceChannelList(value) {
  return new Set(uniqueProviderChannels(parseChannelList(value)));
}

function mergePreferenceRule(base = {}, next = {}) {
  const merged = { ...base, ...next };
  const quietA = base.quiet_hours || base.quietHours || {};
  const quietB = next.quiet_hours || next.quietHours || {};
  if (quietA && typeof quietA === "object" || quietB && typeof quietB === "object") {
    merged.quiet_hours = { ...(quietA && typeof quietA === "object" ? quietA : {}), ...(quietB && typeof quietB === "object" ? quietB : {}) };
  }
  return merged;
}

function preferenceRuleForDispatch(dispatch = {}, preferences = {}) {
  const targetRole = compact(dispatch.target_role);
  const targetID = compact(dispatch.target_id);
  const type = compact(dispatch.type);
  const keys = [
    "default",
    targetRole,
    targetRole && targetID ? `${targetRole}:${targetID}` : "",
    type ? `type:${type}` : "",
    targetRole && targetID && type ? `${targetRole}:${targetID}:${type}` : ""
  ].filter(Boolean);
  return keys.reduce((rule, key) => mergePreferenceRule(rule, preferences[key] || {}), {});
}

function minutesFromClock(value) {
  const match = String(value || "").trim().match(/^(\d{1,2}):(\d{2})$/);
  if (!match) return null;
  const hour = Number(match[1]);
  const minute = Number(match[2]);
  if (!Number.isInteger(hour) || !Number.isInteger(minute) || hour < 0 || hour > 23 || minute < 0 || minute > 59) return null;
  return hour * 60 + minute;
}

function offsetMinutes(value) {
  const match = String(value || "+00:00").trim().match(/^([+-])(\d{2}):?(\d{2})$/);
  if (!match) return 0;
  const sign = match[1] === "-" ? -1 : 1;
  return sign * (Number(match[2]) * 60 + Number(match[3]));
}

function localMinutesAt(now, offset) {
  const date = now instanceof Date ? now : new Date(now || Date.now());
  const minutes = date.getUTCHours() * 60 + date.getUTCMinutes() + offsetMinutes(offset);
  return ((minutes % 1440) + 1440) % 1440;
}

function isWithinQuietHours(quiet = {}, now = new Date()) {
  if (!quiet || typeof quiet !== "object" || quiet.enabled === false) return false;
  const start = minutesFromClock(quiet.start || quiet.start_time || quiet.startTime);
  const end = minutesFromClock(quiet.end || quiet.end_time || quiet.endTime);
  if (start === null || end === null || start === end) return false;
  const current = localMinutesAt(now, quiet.timezone_offset || quiet.timezoneOffset || quiet.utc_offset || quiet.utcOffset || "+00:00");
  return start < end ? current >= start && current < end : current >= start || current < end;
}

function quietRetryAt(quiet = {}, now = new Date()) {
  const start = minutesFromClock(quiet.start || quiet.start_time || quiet.startTime);
  const end = minutesFromClock(quiet.end || quiet.end_time || quiet.endTime);
  if (start === null || end === null || start === end) return "";
  const offset = offsetMinutes(quiet.timezone_offset || quiet.timezoneOffset || quiet.utc_offset || quiet.utcOffset || "+00:00");
  const date = now instanceof Date ? now : new Date(now || Date.now());
  if (!Number.isFinite(date.getTime())) return "";
  const current = localMinutesAt(date, quiet.timezone_offset || quiet.timezoneOffset || quiet.utc_offset || quiet.utcOffset || "+00:00");
  const localDate = new Date(date.getTime() + offset * 60000);
  const localMidnightUTC = Date.UTC(localDate.getUTCFullYear(), localDate.getUTCMonth(), localDate.getUTCDate()) - offset * 60000;
  const nextDay = start > end && current >= start ? 1 : 0;
  return new Date(localMidnightUTC + nextDay * 86400000 + end * 60000).toISOString();
}

export function notificationDeliveryPreferenceDecision(dispatch = {}, options = {}) {
  const preferences = options.preferences ? normalizeDeliveryPreferences(options.preferences) : deliveryPreferencesForOptions(options);
  const rule = preferenceRuleForDispatch(dispatch, preferences);
  const channel = normalizeProviderChannel(dispatch.channel);
  const enabled = normalizePreferenceChannelList(rule.enabled_channels || rule.enabledChannels);
  const disabled = normalizePreferenceChannelList(rule.disabled_channels || rule.disabledChannels);
  if (disabled.has(channel) || (enabled.size > 0 && !enabled.has(channel))) {
    return {
      allowed: false,
      status: "queued",
      reason: "notification_preference_disabled",
      message: `notification preference disabled ${channel}`
    };
  }
  const quiet = rule.quiet_hours || rule.quietHours || {};
  const quietChannels = normalizePreferenceChannelList(quiet.channels || quiet.apply_channels || quiet.applyChannels || providerDeliveryChannels);
  const exemptTypes = new Set(parseChannelList(quiet.exempt_types || quiet.exemptTypes));
  if (quietChannels.has(channel) && !exemptTypes.has(compact(dispatch.type)) && isWithinQuietHours(quiet, options.now || new Date())) {
    const retryAt = quietRetryAt(quiet, options.now || new Date());
    return {
      allowed: false,
      status: compact(quiet.status, "queued"),
      reason: "notification_quiet_window",
      message: `notification quiet window suppressed ${channel}`,
      quiet_until: quiet.end || quiet.end_time || quiet.endTime || "",
      retry_at: retryAt
    };
  }
  return { allowed: true, status: "send", reason: "allowed", message: "allowed" };
}

export function buildProviderDispatches(event = {}, notification = {}, options = {}) {
  const payload = eventPayload(event);
  const type = eventType(event);
  if (event.topic === "notification.delivery_retries" || type === "notification.delivery_retries.scheduled") {
    const snapshots = notificationSnapshotsById(payload);
    const dispatches = (Array.isArray(payload.deliveries) ? payload.deliveries : [])
      .map((delivery) => {
        const channel = normalizeProviderChannel(delivery?.channel || payload.channel);
        if (!channel) return null;
        const notificationID = compact(delivery?.notification_id || delivery?.NotificationID);
        const source = snapshots.get(notificationID) || {};
        const provider = compact(delivery?.provider || payload.provider || channel, channel);
        const idempotencySeed = compact(event.id || payload.idempotency_key || delivery?.id || notificationID);
        return {
          attempt: "retry",
          notification_id: notificationID,
          source_delivery_id: compact(delivery?.id),
          target_role: compact(delivery?.target_role || source.target_role || payload.target_role || notification.target_role),
          target_id: compact(delivery?.target_id || source.target_id || payload.target_id || notification.target_id),
          type: compact(source.type || type || notification.type),
          channel,
          provider,
          title: compact(source.title || payload.title || notification.title, "Infinitech 通知"),
          body: compact(source.body || payload.body || notification.body, `通知 ${notificationID} 正在重试投递。`),
          template_key: compact(payload.template_key || source.type || type || notification.type, "notification.default"),
          retry_policy: compact(payload.retry_policy),
          retry_at: compact(payload.retry_at),
          idempotency_key: `provider:${channel}:${provider}:${notificationID}:${idempotencySeed}:retry`
        };
      })
      .filter((dispatch) => dispatch && dispatch.notification_id);
    return applyProviderTemplates(dispatches, options);
  }

  const notificationID = compact(notification.id || notification.notification_id || payload.notification_id);
  const dispatches = providerChannelsForEvent(event, options).map((channel) => {
    const provider = providerForChannel(channel, event);
    const idempotencySeed = compact(notification.idempotency_key || payload.idempotency_key || event.id || notificationID);
    return {
      attempt: "initial",
      notification_id: notificationID,
      source_delivery_id: "",
      target_role: compact(notification.target_role || payload.target_role),
      target_id: compact(notification.target_id || payload.target_id),
      type: compact(notification.type || type),
      channel,
      provider,
      title: compact(notification.title || payload.title, "Infinitech 通知"),
      body: compact(notification.body || payload.body),
      template_key: compact(payload.template_key || notification.type || type, "notification.default"),
      idempotency_key: `provider:${channel}:${provider}:${notificationID}:${idempotencySeed}:initial`
    };
  }).filter((dispatch) => dispatch.notification_id);
  return applyProviderTemplates(dispatches, options);
}

export function normalizeProviderDeliveryRecord(dispatch = {}, result = {}, options = {}) {
  const attemptedAt = options.attemptedAt || new Date().toISOString();
  const status = compact(result.status, result.ok === false ? "failed" : "delivered");
  const normalizedStatus = ["queued", "delivered", "failed"].includes(status) ? status : "failed";
  const errorCode = compact(result.error_code, normalizedStatus === "failed" ? "provider_failed" : "");
  const errorMessage = compact(result.error_message, normalizedStatus === "failed" ? "provider delivery failed" : "");
  return {
    notification_id: compact(dispatch.notification_id),
    channel: compact(dispatch.channel),
    provider: compact(dispatch.provider, dispatch.channel),
    status: normalizedStatus,
    provider_message_id: compact(result.provider_message_id, `${compact(dispatch.provider, dispatch.channel)}_${shortHash(dispatch.idempotency_key)}`),
    error_code: errorCode,
    error_message: errorMessage,
    idempotency_key: `delivery:${compact(dispatch.idempotency_key)}`,
    attempted_at: attemptedAt,
    delivered_at: normalizedStatus === "delivered" ? attemptedAt : undefined,
    retry_at: result.retry_at || result.retryAt || undefined
  };
}

function providerCallbackUnix(value) {
  const normalized = compact(value);
  if (!normalized) return "0";
  const millis = Date.parse(normalized);
  if (!Number.isFinite(millis)) return "0";
  return String(Math.floor(millis / 1000));
}

function envFlag(value) {
  return ["1", "true", "yes", "on"].includes(compact(value).toLowerCase());
}

function envInt(value, fallback) {
  const parsed = Number.parseInt(compact(value), 10);
  return Number.isFinite(parsed) ? parsed : fallback;
}

function providerCallbackCanonicalLines(payload = {}) {
  return [
    compact(payload.notification_id),
    compact(payload.channel),
    compact(payload.provider),
    compact(payload.status),
    compact(payload.provider_message_id),
    compact(payload.error_code),
    compact(payload.error_message),
    compact(payload.idempotency_key),
    providerCallbackUnix(payload.attempted_at),
    providerCallbackUnix(payload.delivered_at),
    providerCallbackUnix(payload.callback_at)
  ];
}

export function signProviderCallback(payload = {}, secret = "") {
  return createHmac("sha256", compact(secret)).update(providerCallbackCanonicalLines(payload).join("\n")).digest("hex");
}

export function normalizeProviderCallbackPayload(dispatch = {}, result = {}, options = {}) {
  const callbackAt = options.callbackAt || options.attemptedAt || new Date().toISOString();
  const delivery = normalizeProviderDeliveryRecord(dispatch, result, { attemptedAt: options.attemptedAt || callbackAt });
  const provider = compact(delivery.provider, delivery.channel);
  const providerMessageID = compact(delivery.provider_message_id);
  const payload = {
    notification_id: delivery.notification_id,
    channel: delivery.channel,
    provider,
    status: delivery.status,
    provider_message_id: providerMessageID,
    error_code: delivery.error_code,
    error_message: delivery.error_message,
    idempotency_key: compact(result.callback_idempotency_key || result.idempotency_key, `provider_callback:${provider}:${providerMessageID || shortHash(dispatch.idempotency_key)}`),
    attempted_at: delivery.attempted_at,
    delivered_at: delivery.delivered_at,
    callback_at: callbackAt
  };
  const callbackSecret = compact(options.callbackSecret || options.env?.NOTIFICATION_PROVIDER_CALLBACK_SECRET || process.env.NOTIFICATION_PROVIDER_CALLBACK_SECRET);
  if (callbackSecret) {
    payload.signature = signProviderCallback(payload, callbackSecret);
  }
  return payload;
}

export function createNotificationProviderDispatcher(options = {}) {
  const env = options.env || process.env;
  const fetchImpl = options.fetchImpl || globalThis.fetch;
  const adapters = options.adapters || {};
  const clock = options.clock || (() => new Date());
  const preferenceResolver = typeof options.preferenceResolver === "function" ? options.preferenceResolver : null;
  async function sendViaConfiguredHttp(dispatch) {
    const endpoint = providerEndpointFor(dispatch.channel, env);
    if (!endpoint) {
      return {
        ok: false,
        status: "failed",
        error_code: "provider_not_configured",
        error_message: `missing provider endpoint for ${dispatch.channel}`
      };
    }
    if (typeof fetchImpl !== "function") {
      return {
        ok: false,
        status: "failed",
        error_code: "provider_fetch_missing",
        error_message: "fetch implementation is required"
      };
    }
    const headers = { "Content-Type": "application/json" };
    const token = providerTokenFor(dispatch.channel, env);
    if (token) headers.Authorization = token.startsWith("Bearer ") ? token : `Bearer ${token}`;
    const response = await fetchImpl(endpoint, {
      method: "POST",
      headers,
      body: JSON.stringify(dispatch)
    });
    if (!response.ok) {
      const text = await response.text();
      return {
        ok: false,
        status: "failed",
        error_code: `http_${response.status}`,
        error_message: text || `provider http ${response.status}`
      };
    }
    const payload = await response.json().catch(() => ({}));
    return {
      ok: payload.ok !== false,
      status: compact(payload.status, payload.ok === false ? "failed" : "delivered"),
      provider_message_id: payload.provider_message_id || payload.message_id,
      error_code: payload.error_code,
      error_message: payload.error_message
    };
  }
  return {
    dispatchesFor(event = {}, notification = {}) {
      return buildProviderDispatches(event, notification, { channels: options.channels, env, templates: options.templates });
    },
    async deliver(event = {}, notification = {}) {
      const attemptedAt = clock().toISOString();
      const dispatches = buildProviderDispatches(event, notification, { channels: options.channels, env, templates: options.templates });
      const staticPreferences = deliveryPreferencesForOptions({ preferences: options.preferences, env });
      let preferences = staticPreferences;
      let preferenceLookupError = null;
      if (preferenceResolver && dispatches.length > 0) {
        try {
          preferences = mergeDeliveryPreferences(
            staticPreferences,
            await preferenceResolver(dispatches, { event, notification, attemptedAt })
          );
        } catch (error) {
          preferenceLookupError = error instanceof Error ? error.message : String(error);
        }
      }
      const results = [];
      for (const dispatch of dispatches) {
        const preferenceDecision = preferenceLookupError
          ? {
              allowed: false,
              status: "queued",
              reason: "notification_preference_lookup_failed",
              message: `notification preference lookup failed: ${preferenceLookupError}`
            }
          : notificationDeliveryPreferenceDecision(dispatch, { preferences, env, now: attemptedAt });
        if (!preferenceDecision.allowed) {
          const providerResult = {
            ok: true,
            status: compact(preferenceDecision.status, "queued"),
            provider_message_id: `preference_${shortHash(`${dispatch.idempotency_key}:${preferenceDecision.reason}`)}`,
            error_code: preferenceDecision.reason,
            error_message: preferenceDecision.message,
            retry_at: preferenceDecision.retry_at
          };
          results.push({
            dispatch: { ...dispatch, preference_decision: preferenceDecision },
            result: providerResult,
            delivery: normalizeProviderDeliveryRecord(dispatch, providerResult, { attemptedAt }),
            preference_decision: preferenceDecision
          });
          continue;
        }
        const adapter = adapters[dispatch.provider] || adapters[dispatch.channel] || sendViaConfiguredHttp;
        try {
          const providerResult = await adapter(dispatch);
          results.push({
            dispatch,
            result: providerResult,
            delivery: normalizeProviderDeliveryRecord(dispatch, providerResult, { attemptedAt })
          });
        } catch (error) {
          const providerResult = {
            ok: false,
            status: "failed",
            error_code: "provider_exception",
            error_message: error instanceof Error ? error.message : String(error)
          };
          results.push({
            dispatch,
            result: providerResult,
            delivery: normalizeProviderDeliveryRecord(dispatch, providerResult, { attemptedAt })
          });
        }
      }
      return results;
    }
  };
}

export function buildNotification(event = {}) {
  const payload = event.payload && typeof event.payload === "object" ? event.payload : event.notification || {};
  const type = eventType(event);
  if (event.topic === "audit.retention_alerts" || type === "audit.retention_alerts.emitted") {
    const alertCount = Number(payload.alert_count || event.alert_count || 0);
    const criticalCount = Number(payload.critical_count || event.critical_count || 0);
    return {
      type: "audit.retention_alerts.emitted",
      target_role: "security",
      target_id: "audit_retention",
      title: "审计留存告警",
      body: `critical=${criticalCount}; total=${alertCount}`,
      idempotency_key: `notify:audit.retention_alerts:${event.id || event.idempotency_key || payload.idempotency_key || ""}`
    };
  }
  if (event.topic === "notification.delivery_failed_alerts" || type === "notification.delivery_failed_alerts.emitted") {
    const failedCount = Number(payload.failed_count || event.failed_count || 0);
    const channel = String(payload.channel || event.channel || "all").trim() || "all";
    const provider = String(payload.provider || event.provider || channel).trim() || channel;
    return {
      type: "notification.delivery_failed_alerts.emitted",
      target_role: "security",
      target_id: "notification_delivery",
      title: "通知投递失败告警",
      body: `failed=${failedCount}; channel=${channel}; provider=${provider}`,
      idempotency_key: `notify:notification.delivery_failed_alerts:${event.id || event.idempotency_key || payload.idempotency_key || ""}`
    };
  }
  if (event.topic === "notification.delivery_retries" || type === "notification.delivery_retries.scheduled") {
    const scheduledCount = Number(payload.scheduled_count || event.scheduled_count || 0);
    const channel = String(payload.channel || event.channel || "all").trim() || "all";
    const provider = String(payload.provider || event.provider || channel).trim() || channel;
    const deliveryStatus = String(payload.delivery_status || event.delivery_status || "failed").trim() || "failed";
    const errorCode = String(payload.error_code || event.error_code || "all").trim() || "all";
    const retryAfterSeconds = Number(payload.retry_after_seconds || event.retry_after_seconds || 0);
    return {
      type: "notification.delivery_retries.scheduled",
      target_role: "security",
      target_id: "notification_delivery",
      title: "通知投递重试计划",
      body: `scheduled=${scheduledCount}; status=${deliveryStatus}; error=${errorCode}; channel=${channel}; provider=${provider}; retry_after=${retryAfterSeconds}s`,
      idempotency_key: `notify:notification.delivery_retries:${event.id || event.idempotency_key || payload.idempotency_key || ""}`
    };
  }
  if (event.topic === "merchant.qualification_reviewed" || type === "merchant.qualification_reviewed") {
    const merchantID = String(payload.merchant_id || event.merchant_id || "").trim();
    const status = String(payload.status || event.status || "").trim();
    const approved = status === "approved";
    return {
      type: "merchant.qualification_reviewed",
      target_role: "merchant",
      target_id: merchantID,
      title: payload.title || "商户资质审核结果",
      body: payload.body || (approved ? "资质审核已通过，系统已更新商户接单资格。" : "资质审核未通过，请补充有效文件后重新提交。"),
      idempotency_key: `notify:merchant.qualification_reviewed:${event.id || event.idempotency_key || payload.outbox_event_id || payload.qualification_id || ""}`
    };
  }
  const target = payload.target || event.target || {};
  return {
    type,
    target_role: String(target.role || "").trim(),
    target_id: String(target.id || "").trim(),
    title: payload.title || event.title || "Infinitech 通知",
    body: payload.body || event.body || "",
    idempotency_key: `notify:${type}:${event.id || event.order_id || payload.order_id || event.message_id || payload.message_id || ""}`
  };
}

export function normalizeNotificationRecord(event = {}, notification = buildNotification(event)) {
  const payload = event.payload && typeof event.payload === "object" ? event.payload : {};
  return {
    target_role: String(notification.target_role || "").trim(),
    target_id: String(notification.target_id || "").trim(),
    type: String(notification.type || "").trim(),
    channel: String(notification.channel || "in_app").trim() || "in_app",
    title: String(notification.title || "").trim(),
    body: String(notification.body || "").trim(),
    source_topic: String(event.topic || payload.source_topic || "").trim(),
    source_event_id: String(event.id || event.outbox_event_id || payload.outbox_event_id || "").trim(),
    idempotency_key: String(notification.idempotency_key || "").trim(),
    created_at: event.created_at || payload.reviewed_at || payload.created_at || undefined
  };
}

export function normalizeNotificationDeliveryRecord(event = {}, notification = {}) {
  const payload = event.payload && typeof event.payload === "object" ? event.payload : {};
  const notificationID = String(notification.id || notification.notification_id || payload.notification_id || "").trim();
  const channel = String(notification.channel || payload.channel || "in_app").trim() || "in_app";
  const idempotencyKey = String(notification.idempotency_key || payload.idempotency_key || event.idempotency_key || event.id || "").trim();
  return {
    notification_id: notificationID,
    channel,
    provider: String(payload.provider || channel).trim() || channel,
    status: String(payload.delivery_status || "delivered").trim() || "delivered",
    provider_message_id: String(payload.provider_message_id || notificationID).trim(),
    idempotency_key: `delivery:${idempotencyKey}:${channel}`,
    attempted_at: event.created_at || payload.reviewed_at || payload.created_at || undefined,
    delivered_at: event.created_at || payload.reviewed_at || payload.created_at || undefined
  };
}

export function createNotificationApiClient(options = {}) {
  const apiBaseUrl = String(options.apiBaseUrl || process.env.API_BASE_URL || "").replace(/\/+$/, "");
  const token = String(options.token || process.env.NOTIFICATION_WORKER_TOKEN || "").trim();
  const fetchImpl = options.fetchImpl || globalThis.fetch;
  if (!apiBaseUrl) {
    throw new Error("API_BASE_URL is required for notification api client");
  }
  if (typeof fetchImpl !== "function") {
    throw new Error("fetch implementation is required");
  }
  const recordDelivery = options.recordDelivery !== false;
  const preferenceResolver = options.preferenceResolver || (options.fetchDeliveryPreferences === false
    ? null
    : createNotificationPreferenceResolver({
        apiBaseUrl,
        token,
        env: options.env,
        fetchImpl,
        clock: options.clock,
        cache: options.preferenceCache,
        cacheTtlMs: options.preferenceCacheTtlMs,
        cacheStaleMs: options.preferenceCacheStaleMs,
        cacheMaxKeys: options.preferenceCacheMaxKeys
      }));
  const providerDispatcher = options.providerDispatcher || createNotificationProviderDispatcher({
    adapters: options.providerAdapters,
    channels: options.providerChannels,
    clock: options.clock,
    env: options.env,
    fetchImpl,
    preferences: options.deliveryPreferences || options.providerPreferences,
    preferenceResolver,
    templates: options.providerTemplates
  });
  async function recordDeliveryPayload(notificationID, deliveryPayload, headers) {
    if (!notificationID) {
      return null;
    }
    const deliveryResponse = await fetchImpl(`${apiBaseUrl}/api/notifications/${encodeURIComponent(notificationID)}/deliveries`, {
      method: "POST",
      headers,
      body: JSON.stringify(deliveryPayload)
    });
    if (!deliveryResponse.ok) {
      const text = await deliveryResponse.text();
      throw new Error(`notification delivery api failed: ${deliveryResponse.status} ${text}`);
    }
    return deliveryResponse.json();
  }
  return {
    async invalidateNotificationPreferences(event = {}) {
      if (typeof preferenceResolver?.invalidate === "function") {
        return { success: true, data: preferenceResolver.invalidate(event) };
      }
      return { success: true, data: invalidateNotificationPreferenceCache(options.preferenceCache, event) };
    },
    async recordNotification(event = {}) {
      const payload = event.target_role
        ? normalizeNotificationRecord({}, event)
        : normalizeNotificationRecord(event);
      const headers = { "Content-Type": "application/json" };
      if (token) {
        headers.Authorization = token.startsWith("Bearer ") ? token : `Bearer ${token}`;
      }
      const response = await fetchImpl(`${apiBaseUrl}/api/notifications`, {
        method: "POST",
        headers,
        body: JSON.stringify(payload)
      });
      if (!response.ok) {
        const text = await response.text();
        throw new Error(`notification api failed: ${response.status} ${text}`);
      }
      const created = await response.json();
      const notification = created?.data || created;
      if (!recordDelivery) {
        return created;
      }
      const deliveryPayload = normalizeNotificationDeliveryRecord(event, notification);
      const deliveries = [];
      if (!deliveryPayload.notification_id) {
        return created;
      }
      deliveries.push(await recordDeliveryPayload(deliveryPayload.notification_id, deliveryPayload, headers));
      const providerResults = await providerDispatcher.deliver(event, notification);
      for (const providerResult of providerResults) {
        if (providerResult.delivery?.notification_id) {
          deliveries.push(await recordDeliveryPayload(providerResult.delivery.notification_id, providerResult.delivery, headers));
        }
      }
      return { ...created, delivery: deliveries[0], provider_deliveries: providerResults, deliveries };
    }
  };
}

export function normalizeQuietWindowRetrySchedulePayload(options = {}) {
  const now = options.now || options.clock?.().toISOString?.() || new Date().toISOString();
  const payload = {
    target_role: compact(options.target_role || options.targetRole),
    target_id: compact(options.target_id || options.targetId),
    channel: compact(options.channel),
    provider: compact(options.provider),
    limit: Number(options.limit || 50),
    retry_after_seconds: Number(options.retry_after_seconds || options.retryAfterSeconds || 0),
    now
  };
  if (!Number.isFinite(payload.limit) || payload.limit <= 0) payload.limit = 50;
  if (payload.limit > 100) payload.limit = 100;
  if (!Number.isFinite(payload.retry_after_seconds) || payload.retry_after_seconds < 0) payload.retry_after_seconds = 0;
  if (payload.retry_after_seconds > 86400) payload.retry_after_seconds = 86400;
  return payload;
}

export function createQuietWindowRetryScheduler(options = {}) {
  const env = options.env || process.env;
  const apiBaseUrl = String(options.apiBaseUrl || env.API_BASE_URL || "").replace(/\/+$/, "");
  const token = compact(options.token || env.NOTIFICATION_WORKER_TOKEN);
  const fetchImpl = options.fetchImpl || globalThis.fetch;
  const clock = options.clock || (() => new Date());
  if (!apiBaseUrl) {
    throw new Error("API_BASE_URL is required for quiet-window retry scheduler");
  }
  if (typeof fetchImpl !== "function") {
    throw new Error("fetch implementation is required");
  }
  return {
    async scheduleQuietWindowRetries(overrides = {}) {
      const payload = normalizeQuietWindowRetrySchedulePayload({
        target_role: options.target_role || env.NOTIFICATION_QUIET_RETRY_TARGET_ROLE,
        target_id: options.target_id || env.NOTIFICATION_QUIET_RETRY_TARGET_ID,
        channel: options.channel || env.NOTIFICATION_QUIET_RETRY_CHANNEL,
        provider: options.provider || env.NOTIFICATION_QUIET_RETRY_PROVIDER,
        limit: options.limit ?? env.NOTIFICATION_QUIET_RETRY_LIMIT,
        retry_after_seconds: options.retry_after_seconds ?? options.retryAfterSeconds ?? env.NOTIFICATION_QUIET_RETRY_AFTER_SECONDS,
        clock,
        now: clock().toISOString(),
        ...overrides
      });
      const headers = { "Content-Type": "application/json" };
      if (token) headers.Authorization = token.startsWith("Bearer ") ? token : `Bearer ${token}`;
      const response = await fetchImpl(`${apiBaseUrl}/api/admin/notification-deliveries/quiet-window-retries/schedule`, {
        method: "POST",
        headers,
        body: JSON.stringify(payload)
      });
      if (!response.ok) {
        const text = await response.text();
        throw new Error(`quiet-window retry schedule api failed: ${response.status} ${text}`);
      }
      return response.json();
    }
  };
}

export function createQuietWindowRetryLoop(options = {}) {
  const env = options.env || process.env;
  const intervalMs = Math.max(1000, Number(options.intervalMs || env.NOTIFICATION_QUIET_RETRY_INTERVAL_MS || 60000));
  const setIntervalImpl = options.setIntervalImpl || setInterval;
  const clearIntervalImpl = options.clearIntervalImpl || clearInterval;
  const scheduler = options.scheduler || createQuietWindowRetryScheduler(options);
  const onResult = typeof options.onResult === "function" ? options.onResult : () => {};
  const onError = typeof options.onError === "function" ? options.onError : () => {};
  const tick = async () => {
    try {
      const result = await scheduler.scheduleQuietWindowRetries();
      onResult(result);
      return result;
    } catch (error) {
      onError(error);
      throw error;
    }
  };
  const timer = setIntervalImpl(() => {
    tick().catch(() => {});
  }, intervalMs);
  return {
    tick,
    stop() {
      clearIntervalImpl(timer);
    }
  };
}

export function createNotificationConsumer(options = {}) {
  const apiClient = options.apiClient;
  return createIdempotentConsumer({
    consumerName: options.consumerName || workerName,
    ledger: options.ledger,
    clock: options.clock,
    handler: options.handler || ((event = {}) => {
      if (isNotificationPreferenceChangedEvent(event)) {
        return apiClient?.invalidateNotificationPreferences
          ? apiClient.invalidateNotificationPreferences(event)
          : { success: true, data: invalidateNotificationPreferenceCache(options.preferenceCache, event) };
      }
      return apiClient?.recordNotification ? apiClient.recordNotification(event) : buildNotification(event);
    })
  });
}

if (import.meta.url === `file://${process.argv[1]}`) {
  console.log(`${workerName} ready; topics=${subscribedTopics.join(",")}`);
  if (envFlag(process.env.NOTIFICATION_QUIET_RETRY_AUTO_SCHEDULE)) {
    const loop = createQuietWindowRetryLoop({
      onResult(result) {
        const schedule = result?.data?.schedule || result?.schedule || {};
        console.log(`${workerName} quiet retry tick; scheduled=${schedule.scheduled_count || 0}`);
      },
      onError(error) {
        console.error(`${workerName} quiet retry tick failed: ${error.message}`);
      }
    });
    loop.tick().catch(() => {});
    process.once("SIGINT", () => loop.stop());
    process.once("SIGTERM", () => loop.stop());
  }
}
