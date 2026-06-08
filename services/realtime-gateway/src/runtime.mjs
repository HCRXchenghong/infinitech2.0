export function createRealtimeConfig(env = process.env) {
  const redisUrl = String(env.REALTIME_REDIS_URL || "").trim();
  return {
    service: "realtime-gateway",
    transportMode: "websocket-only",
    instanceID: env.REALTIME_INSTANCE_ID || "",
    websocketAuthRequired: env.REALTIME_WS_AUTH_REQUIRED === "true",
    websocketDevTokensAllowed: env.REALTIME_ALLOW_DEV_TOKENS === "true",
    membershipAuthUrl: env.REALTIME_MEMBERSHIP_AUTH_URL || "",
    membershipAuthConfigured: Boolean(String(env.REALTIME_MEMBERSHIP_AUTH_URL || "").trim()),
    redisAdapter: redisUrl ? "redis-pubsub" : env.REALTIME_REDIS_ADAPTER || "disabled",
    redisConfigured: Boolean(redisUrl),
    redisChannel: env.REALTIME_REDIS_CHANNEL || "infinitech:realtime:events",
    kafkaEvents: env.KAFKA_BROKERS || "required-in-production",
    internalPublishPath: "/internal/realtime/publish",
    namespaces: ["/im", "/orders", "/rider", "/dispatch", "/rtc", "/support"],
    events: [
      "message.sent",
      "order.status_changed",
      "rider.location_updated",
      "dispatch.assigned",
      "rtc.signal"
    ],
    capacityGate: {
      requiredStages: ["10k", "30k", "60k", "100k"],
      claimAllowedWithoutReports: false
    }
  };
}
