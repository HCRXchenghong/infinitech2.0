import { createIdempotentConsumer } from "../../../packages/domain-core/src/index.mjs";

export const workerName = "integration-worker";
export const subscribedTopics = ["oauth.callback", "provider.sync.requested", "map.geocode.requested", "sms.send.requested"];

export function normalizeProviderConfig(config = {}) {
  return {
    provider: String(config.provider || "").trim().toLowerCase(),
    enabled: config.enabled === true,
    scopes: Array.isArray(config.scopes) ? config.scopes.map((scope) => String(scope).trim()).filter(Boolean) : [],
    rate_limit_per_minute: Number.isFinite(Number(config.rate_limit_per_minute)) ? Number(config.rate_limit_per_minute) : 60
  };
}

export function createIntegrationConsumer(options = {}) {
  return createIdempotentConsumer({
    consumerName: options.consumerName || workerName,
    ledger: options.ledger,
    clock: options.clock,
    handler: options.handler || ((event = {}) => normalizeProviderConfig(event.payload || event.provider || event))
  });
}

if (import.meta.url === `file://${process.argv[1]}`) {
  console.log(`${workerName} ready; topics=${subscribedTopics.join(",")}`);
}
