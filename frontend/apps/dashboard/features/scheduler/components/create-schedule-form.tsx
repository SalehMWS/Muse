"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useId } from "react";
import { useForm } from "react-hook-form";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select } from "@/components/ui/select";
import { ApiError } from "@/lib/api/errors";

import { useCreateSchedule } from "../mutations/use-create-schedule";
import { useAccountOptionsQuery, useContentOptionsQuery } from "../queries/use-schedules-query";
import {
  createScheduleSchema,
  DEFAULT_TIMEZONE,
  type CreateScheduleInput,
  type CreateSchedulePayload,
} from "../schemas/create-schedule.schema";

function browserTimezone(): string | null {
  try {
    return Intl.DateTimeFormat().resolvedOptions().timeZone || null;
  } catch {
    return null;
  }
}

function timezoneOptions(): string[] {
  const detected = browserTimezone();
  const base = [
    DEFAULT_TIMEZONE,
    "Europe/London",
    "Europe/Berlin",
    "America/New_York",
    "America/Los_Angeles",
    "Asia/Dubai",
    "Asia/Tokyo",
  ];
  return detected && !base.includes(detected) ? [DEFAULT_TIMEZONE, detected, ...base.slice(1)] : base;
}

function FieldError({ id, message }: { id: string; message?: string }) {
  if (!message) return null;
  return (
    <p id={id} role="alert" className="text-xs text-danger">
      {message}
    </p>
  );
}

export function CreateScheduleForm() {
  const fieldId = useId();
  const contents = useContentOptionsQuery();
  const accounts = useAccountOptionsQuery();

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<CreateScheduleInput, unknown, CreateSchedulePayload>({
    resolver: zodResolver(createScheduleSchema),
    defaultValues: {
      contentId: "",
      instagramAccountId: "",
      scheduledFor: "",
      cronExpression: "",
      timezone: DEFAULT_TIMEZONE,
      maxRetries: "",
    },
  });

  const create = useCreateSchedule(() => reset());

  const contentIdField = `${fieldId}-content`;
  const accountIdField = `${fieldId}-account`;
  const scheduledForField = `${fieldId}-scheduled-for`;
  const cronField = `${fieldId}-cron`;
  const timezoneField = `${fieldId}-timezone`;
  const retriesField = `${fieldId}-retries`;

  const noAccounts = accounts.isSuccess && accounts.data.length === 0;
  const noContent = contents.isSuccess && contents.data.length === 0;
  const blocked = noAccounts || noContent;

  return (
    <Card className="w-full">
      <CardHeader>
        <CardTitle>Schedule a post</CardTitle>
      </CardHeader>
      <CardContent>
        {accounts.isError ? (
          <p role="alert" className="mb-4 text-sm text-danger">
            {accounts.error instanceof ApiError
              ? accounts.error.message
              : "Could not load your Instagram accounts."}
          </p>
        ) : null}

        {noAccounts ? (
          <p
            role="alert"
            className="mb-4 rounded-md border border-warning/40 bg-warning/10 p-3 text-sm text-foreground"
          >
            You have no connected Instagram account. Scheduling publishes to Instagram, so connect
            an account on the Instagram page before you schedule a post.
          </p>
        ) : null}

        {noContent ? (
          <p
            role="alert"
            className="mb-4 rounded-md border border-warning/40 bg-warning/10 p-3 text-sm text-foreground"
          >
            You have no content yet. Create a draft on the Content page first, then come back to
            schedule it.
          </p>
        ) : null}

        <form
          className="flex flex-col gap-4"
          onSubmit={handleSubmit((values) => create.mutate(values))}
          noValidate
        >
          <div className="flex flex-col gap-1.5">
            <Label htmlFor={contentIdField} required>
              Content
            </Label>
            <Select
              id={contentIdField}
              disabled={contents.isPending || noContent}
              aria-invalid={Boolean(errors.contentId) || undefined}
              aria-describedby={errors.contentId ? `${contentIdField}-error` : undefined}
              {...register("contentId")}
            >
              <option value="">
                {contents.isPending ? "Loading content…" : "Select content to schedule"}
              </option>
              {(contents.data ?? []).map((option) => (
                <option key={option.id} value={option.id}>
                  {option.label}
                </option>
              ))}
            </Select>
            <FieldError id={`${contentIdField}-error`} message={errors.contentId?.message} />
          </div>

          <div className="flex flex-col gap-1.5">
            <Label htmlFor={accountIdField} required>
              Instagram account
            </Label>
            <Select
              id={accountIdField}
              disabled={accounts.isPending || noAccounts}
              aria-invalid={Boolean(errors.instagramAccountId) || undefined}
              aria-describedby={
                errors.instagramAccountId ? `${accountIdField}-error` : undefined
              }
              {...register("instagramAccountId")}
            >
              <option value="">
                {accounts.isPending ? "Loading accounts…" : "Select an account"}
              </option>
              {(accounts.data ?? []).map((option) => (
                <option key={option.id} value={option.id}>
                  {option.label}
                </option>
              ))}
            </Select>
            <FieldError
              id={`${accountIdField}-error`}
              message={errors.instagramAccountId?.message}
            />
          </div>

          <div className="flex flex-col gap-1.5">
            <Label htmlFor={scheduledForField}>Publish at</Label>
            <Input
              id={scheduledForField}
              type="datetime-local"
              aria-invalid={Boolean(errors.scheduledFor) || undefined}
              aria-describedby={
                errors.scheduledFor ? `${scheduledForField}-error` : `${scheduledForField}-hint`
              }
              {...register("scheduledFor")}
            />
            {errors.scheduledFor ? (
              <FieldError
                id={`${scheduledForField}-error`}
                message={errors.scheduledFor.message}
              />
            ) : (
              <p id={`${scheduledForField}-hint`} className="text-xs text-muted-foreground">
                Set a one-off time, or leave empty and use a cron expression for a repeating post.
              </p>
            )}
          </div>

          <div className="flex flex-col gap-1.5">
            <Label htmlFor={cronField}>Cron expression</Label>
            <Input
              id={cronField}
              placeholder="0 9 * * 1"
              aria-invalid={Boolean(errors.cronExpression) || undefined}
              aria-describedby={
                errors.cronExpression ? `${cronField}-error` : `${cronField}-hint`
              }
              {...register("cronExpression")}
            />
            {errors.cronExpression ? (
              <FieldError id={`${cronField}-error`} message={errors.cronExpression.message} />
            ) : (
              <p id={`${cronField}-hint`} className="text-xs text-muted-foreground">
                Optional. For example <code>0 9 * * 1</code> publishes every Monday at 09:00.
              </p>
            )}
          </div>

          <div className="flex flex-col gap-1.5">
            <Label htmlFor={timezoneField} required>
              Timezone
            </Label>
            <Select
              id={timezoneField}
              aria-invalid={Boolean(errors.timezone) || undefined}
              aria-describedby={errors.timezone ? `${timezoneField}-error` : undefined}
              {...register("timezone")}
            >
              {timezoneOptions().map((zone) => (
                <option key={zone} value={zone}>
                  {zone}
                </option>
              ))}
            </Select>
            <FieldError id={`${timezoneField}-error`} message={errors.timezone?.message} />
          </div>

          <div className="flex flex-col gap-1.5">
            <Label htmlFor={retriesField}>Max retries</Label>
            <Input
              id={retriesField}
              type="number"
              min={0}
              max={10}
              placeholder="3"
              aria-invalid={Boolean(errors.maxRetries) || undefined}
              aria-describedby={
                errors.maxRetries ? `${retriesField}-error` : `${retriesField}-hint`
              }
              {...register("maxRetries")}
            />
            {errors.maxRetries ? (
              <FieldError id={`${retriesField}-error`} message={errors.maxRetries.message} />
            ) : (
              <p id={`${retriesField}-hint`} className="text-xs text-muted-foreground">
                Optional. How many times NovaFlow retries a failed publish.
              </p>
            )}
          </div>

          <div>
            <Button type="submit" loading={create.isPending} disabled={blocked}>
              Schedule post
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  );
}
