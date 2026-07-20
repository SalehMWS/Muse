import { z } from "zod";

export const DEFAULT_TIMEZONE = "UTC";

export const createScheduleSchema = z
  .object({
    contentId: z.string().min(1, "Pick the content you want to schedule."),
    instagramAccountId: z.string().min(1, "Pick the Instagram account to publish to."),
    scheduledFor: z.string().optional(),
    cronExpression: z.string().optional(),
    timezone: z.string().min(1, "Timezone is required.").default(DEFAULT_TIMEZONE),
    maxRetries: z.string().optional(),
  })
  .refine(
    (values) => Boolean(values.scheduledFor?.trim()) || Boolean(values.cronExpression?.trim()),
    {
      message: "Choose a date and time, or enter a cron expression.",
      path: ["scheduledFor"],
    },
  )
  .refine((values) => parseMaxRetries(values.maxRetries) !== "invalid", {
    message: "Retries must be a whole number between 0 and 10.",
    path: ["maxRetries"],
  });

/** What `register` sees — `timezone` is optional here because the schema defaults it. */
export type CreateScheduleInput = z.input<typeof createScheduleSchema>;
/** What `handleSubmit` hands to the service — defaults already applied. */
export type CreateSchedulePayload = z.output<typeof createScheduleSchema>;

/**
 * `datetime-local` gives a local wall-clock string with no zone ("2026-07-20T15:30").
 * The API wants RFC3339, so resolve it against the browser's zone.
 */
export function toRfc3339(local?: string): string | undefined {
  const trimmed = local?.trim();
  if (!trimmed) return undefined;
  const parsed = new Date(trimmed);
  if (Number.isNaN(parsed.getTime())) return undefined;
  return parsed.toISOString();
}

export function parseMaxRetries(raw?: string): number | undefined | "invalid" {
  const trimmed = raw?.trim();
  if (!trimmed) return undefined;
  const parsed = Number(trimmed);
  if (!Number.isInteger(parsed) || parsed < 0 || parsed > 10) return "invalid";
  return parsed;
}
