export const VIEW_PAGE_SIZE_OPTIONS = Object.freeze([4, 10, 20]);

function toPositiveInteger(value, fallback) {
  const number = Number(value);
  if (!Number.isInteger(number) || number < 1) {
    return fallback;
  }
  return number;
}

export function normalizeViewFilter(filter = {}) {
  const pageSize = VIEW_PAGE_SIZE_OPTIONS.includes(Number(filter.pageSize)) ? Number(filter.pageSize) : VIEW_PAGE_SIZE_OPTIONS[0];
  return {
    query: String(filter.query || "").trim(),
    page: toPositiveInteger(filter.page, 1),
    pageSize
  };
}

export function buildFilteredRows(view, filter = {}) {
  const normalized = normalizeViewFilter(filter);
  const query = normalized.query.toLowerCase();
  const rows = Array.isArray(view?.rows) ? view.rows : [];
  return rows
    .map((row, rowIndex) => ({ row, rowIndex }))
    .filter(({ row }) => {
      if (!query) {
        return true;
      }
      return row.some((cell) => String(cell ?? "").toLowerCase().includes(query));
    });
}

export function buildViewPage(view, filter = {}) {
  const normalized = normalizeViewFilter(filter);
  const filteredRows = buildFilteredRows(view, normalized);
  const totalRows = filteredRows.length;
  const totalPages = Math.max(1, Math.ceil(totalRows / normalized.pageSize));
  const page = Math.min(normalized.page, totalPages);
  const start = (page - 1) * normalized.pageSize;
  return {
    ...normalized,
    page,
    totalRows,
    totalPages,
    rows: filteredRows.slice(start, start + normalized.pageSize)
  };
}
