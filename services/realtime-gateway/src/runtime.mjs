export function createRealtimeConfig(env = process.env) {
  return {
    service: "realtime-gateway",
    transportMode: "websocket-only",
    redisAdapter: env.REALTIME_REDIS_ADAPTER || "required-in-production",
    kafkaEvents: env.KAFKA_BROKERS || "required-in-production",
    namespaces: ["/im", "/orders", "/rider", "/dispatch", "/rtc", "/support"],
    events: [
      "message.created",
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

