/**
 * Analytics is derived, not fetched: the backend has no analytics module.
 * Everything in this file is computed from three real endpoints:
 *   GET /contents?limit=100   -> content rows
 *   GET /worker/stats         -> job queue counters
 *   GET /instagram/accounts   -> connected accounts
 */

export const CONTENT_STATUSES = ["draft", "scheduled", "published", "archived"] as const;

export type ContentStatusKey = (typeof CONTENT_STATUSES)[number];

/** How many content rows the summary samples. The API caps a page at this size. */
export const CONTENT_SAMPLE_LIMIT = 100;

/** Width of the "content created per day" window. */
export const TREND_DAYS = 14;

/* ------------------------------------------------------------------ */
/* Wire shapes                                                         */
/* ------------------------------------------------------------------ */

export interface AnalyticsContentDto {
  id: string;
  title: string;
  status: string;
  content_type: string;
  language: string;
  tags?: string[] | null;
  created_at: string;
  updated_at: string;
  published_at?: string | null;
}

export interface AnalyticsContentListDto {
  items?: AnalyticsContentDto[] | null;
  next_cursor?: string;
}

export interface WorkerStatsDto {
  processed: number;
  succeeded: number;
  retried: number;
  dead_lettered: number;
  failed: number;
}

export interface InstagramAccountDto {
  id: string;
  status: string;
}

/* ------------------------------------------------------------------ */
/* Derived shapes                                                      */
/* ------------------------------------------------------------------ */

export interface StatusSlice {
  status: ContentStatusKey;
  label: string;
  count: number;
}

export interface DailyPoint {
  /** Local calendar day, `YYYY-MM-DD`. */
  date: string;
  /** Short axis label, e.g. `Mar 4`. */
  label: string;
  count: number;
}

export interface WorkerHealthSummary {
  processed: number;
  succeeded: number;
  retried: number;
  deadLettered: number;
  failed: number;
  /** succeeded / processed, 0-100. `null` when nothing has been processed yet. */
  successRate: number | null;
}

export interface AccountsSummary {
  total: number;
  active: number;
}

export interface AnalyticsSummary {
  /** Number of content rows in the sample, not necessarily the account total. */
  sampledContent: number;
  /** True when the API reported another page, i.e. the sample is partial. */
  sampleTruncated: boolean;
  byStatus: StatusSlice[];
  perDay: DailyPoint[];
  /** Content rows that fall inside the trend window. */
  contentInTrendWindow: number;
  /** `null` when /worker/stats could not be reached. */
  worker: WorkerHealthSummary | null;
  /** `null` when /instagram/accounts could not be reached. */
  accounts: AccountsSummary | null;
}

export const STATUS_LABELS: Record<ContentStatusKey, string> = {
  draft: "Draft",
  scheduled: "Scheduled",
  published: "Published",
  archived: "Archived",
};
