/**
 * The backend emits "scheduled" and "publishing"; "pending" is accepted too so
 * an older or newer API revision never breaks rendering.
 */
export type ScheduleStatus =
  | "pending"
  | "scheduled"
  | "publishing"
  | "queued"
  | "published"
  | "failed"
  | "cancelled";

export interface Schedule {
  id: string;
  contentId: string;
  instagramAccountId: string;
  status: ScheduleStatus;
  scheduledFor: string | null;
  cronExpression: string | null;
  timezone: string;
  mediaType: string | null;
  attempts: number;
  maxRetries: number;
  lastError: string | null;
  createdAt: string;
}

export interface ScheduleDto {
  id: string;
  content_id: string;
  instagram_account_id: string;
  status: string;
  scheduled_for?: string | null;
  cron_expression?: string | null;
  timezone: string;
  media_type?: string | null;
  /** Current API field. */
  retry_count?: number;
  /** Documented alias for retry_count. */
  attempts?: number;
  max_retries: number;
  last_error?: string | null;
  created_at: string;
}

export function toSchedule(dto: ScheduleDto): Schedule {
  return {
    id: dto.id,
    contentId: dto.content_id,
    instagramAccountId: dto.instagram_account_id,
    status: dto.status as ScheduleStatus,
    scheduledFor: dto.scheduled_for ?? null,
    cronExpression: dto.cron_expression ?? null,
    timezone: dto.timezone,
    mediaType: dto.media_type ?? null,
    attempts: dto.retry_count ?? dto.attempts ?? 0,
    maxRetries: dto.max_retries,
    lastError: dto.last_error ?? null,
    createdAt: dto.created_at,
  };
}

/** Minimal local shapes — the architecture forbids importing other features. */
export interface ScheduleTarget {
  id: string;
  label: string;
}

export interface ContentOptionDto {
  id: string;
  title: string;
}

export interface ContentOptionListDto {
  items: ContentOptionDto[];
  next_cursor?: string;
}

export interface AccountOptionDto {
  id: string;
  username: string;
  status?: string;
}

export function toContentOption(dto: ContentOptionDto): ScheduleTarget {
  return { id: dto.id, label: dto.title || "Untitled content" };
}

export function toAccountOption(dto: AccountOptionDto): ScheduleTarget {
  return { id: dto.id, label: `@${dto.username}` };
}
