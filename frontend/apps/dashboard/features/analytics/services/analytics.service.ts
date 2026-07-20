import { request } from "@/services/api-client";

import {
  CONTENT_SAMPLE_LIMIT,
  CONTENT_STATUSES,
  STATUS_LABELS,
  TREND_DAYS,
  type AccountsSummary,
  type AnalyticsContentDto,
  type AnalyticsContentListDto,
  type AnalyticsSummary,
  type ContentStatusKey,
  type DailyPoint,
  type InstagramAccountDto,
  type StatusSlice,
  type WorkerHealthSummary,
  type WorkerStatsDto,
} from "../types/summary";

/*
 * The fetches below are deliberately duplicated rather than imported from
 * features/content or features/instagram: the architecture forbids
 * cross-feature imports, so this feature owns its own calls.
 */

function fetchContents(signal?: AbortSignal): Promise<AnalyticsContentListDto> {
  return request<AnalyticsContentListDto>({
    url: "/contents",
    method: "GET",
    signal,
    params: { limit: CONTENT_SAMPLE_LIMIT },
  });
}

function fetchWorkerStats(signal?: AbortSignal): Promise<WorkerStatsDto> {
  return request<WorkerStatsDto>({ url: "/worker/stats", method: "GET", signal });
}

function fetchAccounts(signal?: AbortSignal): Promise<InstagramAccountDto[]> {
  return request<InstagramAccountDto[]>({ url: "/instagram/accounts", method: "GET", signal });
}

export const analyticsService = {
  async summary(signal?: AbortSignal): Promise<AnalyticsSummary> {
    // Content is the primary source — if it fails the screen has nothing to show
    // and must surface a real error. Worker stats and connected accounts are
    // secondary: when they fail their section reports itself as unavailable
    // rather than rendering invented zeros.
    const [contentsResult, workerResult, accountsResult] = await Promise.allSettled([
      fetchContents(signal),
      fetchWorkerStats(signal),
      fetchAccounts(signal),
    ]);

    if (contentsResult.status === "rejected") {
      throw contentsResult.reason;
    }

    const items = contentsResult.value.items ?? [];
    const perDay = buildDailySeries(items, TREND_DAYS);

    return {
      sampledContent: items.length,
      sampleTruncated: Boolean(contentsResult.value.next_cursor),
      byStatus: buildStatusSlices(items),
      perDay,
      contentInTrendWindow: perDay.reduce((total, point) => total + point.count, 0),
      worker: workerResult.status === "fulfilled" ? toWorkerHealth(workerResult.value) : null,
      accounts: accountsResult.status === "fulfilled" ? toAccountsSummary(accountsResult.value) : null,
    };
  },
};

/* ------------------------------------------------------------------ */
/* Derivations                                                         */
/* ------------------------------------------------------------------ */

function buildStatusSlices(items: AnalyticsContentDto[]): StatusSlice[] {
  const counts = new Map<ContentStatusKey, number>(CONTENT_STATUSES.map((status) => [status, 0]));

  for (const item of items) {
    const status = item.status as ContentStatusKey;
    if (counts.has(status)) {
      counts.set(status, (counts.get(status) ?? 0) + 1);
    }
  }

  return CONTENT_STATUSES.map((status) => ({
    status,
    label: STATUS_LABELS[status],
    count: counts.get(status) ?? 0,
  }));
}

function buildDailySeries(items: AnalyticsContentDto[], days: number): DailyPoint[] {
  const counts = new Map<string, number>();

  for (const item of items) {
    const key = toDayKey(item.created_at);
    if (key) {
      counts.set(key, (counts.get(key) ?? 0) + 1);
    }
  }

  const today = startOfDay(new Date());

  return Array.from({ length: days }, (_, index) => {
    const day = new Date(today);
    day.setDate(day.getDate() - (days - 1 - index));
    const key = formatDayKey(day);

    return {
      date: key,
      label: day.toLocaleDateString(undefined, { month: "short", day: "numeric" }),
      count: counts.get(key) ?? 0,
    };
  });
}

function toWorkerHealth(dto: WorkerStatsDto): WorkerHealthSummary {
  const processed = safeCount(dto.processed);
  const succeeded = safeCount(dto.succeeded);

  return {
    processed,
    succeeded,
    retried: safeCount(dto.retried),
    deadLettered: safeCount(dto.dead_lettered),
    failed: safeCount(dto.failed),
    // No jobs processed means there is no rate to report — not a 0% and not a 100%.
    successRate: processed > 0 ? (succeeded / processed) * 100 : null,
  };
}

function toAccountsSummary(accounts: InstagramAccountDto[]): AccountsSummary {
  const list = Array.isArray(accounts) ? accounts : [];

  return {
    total: list.length,
    active: list.filter((account) => account.status === "active").length,
  };
}

/* ------------------------------------------------------------------ */
/* Helpers                                                             */
/* ------------------------------------------------------------------ */

function safeCount(value: number | undefined | null): number {
  return typeof value === "number" && Number.isFinite(value) ? value : 0;
}

function startOfDay(date: Date): Date {
  const copy = new Date(date);
  copy.setHours(0, 0, 0, 0);
  return copy;
}

/** Local-calendar day key, so bucketing matches what the user sees on a clock. */
function formatDayKey(date: Date): string {
  const month = String(date.getMonth() + 1).padStart(2, "0");
  const day = String(date.getDate()).padStart(2, "0");
  return `${date.getFullYear()}-${month}-${day}`;
}

function toDayKey(iso: string): string | null {
  const parsed = new Date(iso);
  return Number.isNaN(parsed.getTime()) ? null : formatDayKey(parsed);
}
