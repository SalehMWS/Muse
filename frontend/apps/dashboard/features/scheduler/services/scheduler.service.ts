import { request } from "@/services/api-client";

import {
  parseMaxRetries,
  toRfc3339,
  type CreateSchedulePayload,
} from "../schemas/create-schedule.schema";
import {
  toAccountOption,
  toContentOption,
  toSchedule,
  type AccountOptionDto,
  type ContentOptionListDto,
  type Schedule,
  type ScheduleDto,
  type ScheduleTarget,
} from "../types/schedule";

const OPTION_LIMIT = 100;

export const schedulerService = {
  async list(contentId: string, signal?: AbortSignal): Promise<Schedule[]> {
    const dtos = await request<ScheduleDto[]>({
      url: `/contents/${contentId}/schedules`,
      method: "GET",
      signal,
    });
    return (dtos ?? []).map(toSchedule);
  },

  async create(contentId: string, input: CreateSchedulePayload): Promise<Schedule> {
    const scheduledFor = toRfc3339(input.scheduledFor);
    const cronExpression = input.cronExpression?.trim() || undefined;
    const maxRetries = parseMaxRetries(input.maxRetries);

    const dto = await request<ScheduleDto>({
      url: `/contents/${contentId}/schedule`,
      method: "POST",
      data: {
        instagram_account_id: input.instagramAccountId,
        scheduled_for: scheduledFor,
        cron_expression: cronExpression,
        timezone: input.timezone,
        max_retries: maxRetries === "invalid" ? undefined : maxRetries,
      },
    });
    return toSchedule(dto);
  },

  async cancel(contentId: string, scheduleId: string): Promise<void> {
    await request<null>({
      url: `/contents/${contentId}/schedules/${scheduleId}`,
      method: "DELETE",
    });
  },

  /**
   * Pickers for the create form. These duplicate a minimal slice of the content
   * and instagram APIs on purpose — cross-feature imports are not allowed.
   */
  async listContentOptions(signal?: AbortSignal): Promise<ScheduleTarget[]> {
    const data = await request<ContentOptionListDto>({
      url: "/contents",
      method: "GET",
      signal,
      params: { limit: OPTION_LIMIT },
    });
    return (data.items ?? []).map(toContentOption);
  },

  async listAccountOptions(signal?: AbortSignal): Promise<ScheduleTarget[]> {
    const dtos = await request<AccountOptionDto[]>({
      url: "/instagram/accounts",
      method: "GET",
      signal,
    });
    return (dtos ?? []).map(toAccountOption);
  },
};
