export function buildAuthHeaders(token = "") {
  const normalized = String(token || "").trim();
  return normalized ? { Authorization: normalized.startsWith("Bearer ") ? normalized : `Bearer ${normalized}` } : {};
}

export function buildJsonRequest(body, token = "") {
  return {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      ...buildAuthHeaders(token)
    },
    body: JSON.stringify(body ?? {})
  };
}

