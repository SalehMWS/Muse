"use client";

import { CalendarClock } from "lucide-react";
import { useId, useState } from "react";

import { EmptyState } from "@/components/shared/empty-state";
import { ErrorCard } from "@/components/shared/error-card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Select } from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import { ApiError } from "@/lib/api/errors";

import { useCancelSchedule } from "../mutations/use-cancel-schedule";
import { useContentOptionsQuery, useSchedulesQuery } from "../queries/use-schedules-query";
import type { ScheduleStatus } from "../types/schedule";

const statusVariant: Record<
  ScheduleStatus,
  "default" | "primary" | "success" | "warning" | "danger" | "info"
> = {
  pending: "warning",
  scheduled: "warning",
  publishing: "info",
  queued: "info",
  published: "success",
  failed: "danger",
  cancelled: "default",
};

const CANCELLABLE: ScheduleStatus[] = ["pending", "scheduled", "queued"];

function formatDateTime(iso: string | null): string {
  if (!iso) return "No fixed time";
  const parsed = Date.parse(iso);
  if (Number.isNaN(parsed)) return "Unknown time";
  return new Date(parsed).toLocaleString(undefined, {
    dateStyle: "medium",
    timeStyle: "short",
  });
}

export function ScheduleList() {
  const selectId = useId();
  const [contentId, setContentId] = useState("");

  const contents = useContentOptionsQuery();
  const schedules = useSchedulesQuery(contentId);
  const cancel = useCancelSchedule();

  const picker = (
    <div className="flex flex-col gap-1.5">
      <Label htmlFor={selectId}>Show schedules for</Label>
      <Select
        id={selectId}
        value={contentId}
        disabled={contents.isPending}
        onChange={(event) => setContentId(event.target.value)}
        className="max-w-sm"
      >
        <option value="">
          {contents.isPending ? "Loading content…" : "Select content"}
        </option>
        {(contents.data ?? []).map((option) => (
          <option key={option.id} value={option.id}>
            {option.label}
          </option>
        ))}
      </Select>
    </div>
  );

  function body() {
    if (contents.isError) {
      return (
        <ErrorCard
          message={
            contents.error instanceof ApiError
              ? contents.error.message
              : "Could not load your content."
          }
          action={
            <Button variant="outline" size="sm" onClick={() => contents.refetch()}>
              Try again
            </Button>
          }
        />
      );
    }

    if (!contentId) {
      return (
        <EmptyState
          icon={CalendarClock}
          title="Pick a piece of content"
          description="Schedules are attached to a single piece of content. Choose one above to see when it goes out."
        />
      );
    }

    if (schedules.isPending) {
      return (
        <div className="flex flex-col gap-3">
          {Array.from({ length: 2 }).map((_, index) => (
            <Skeleton key={index} className="h-24 w-full" />
          ))}
        </div>
      );
    }

    if (schedules.isError) {
      return (
        <ErrorCard
          message={
            schedules.error instanceof ApiError
              ? schedules.error.message
              : "Could not load the schedules for this content."
          }
          action={
            <Button variant="outline" size="sm" onClick={() => schedules.refetch()}>
              Try again
            </Button>
          }
        />
      );
    }

    if (schedules.data.length === 0) {
      return (
        <EmptyState
          icon={CalendarClock}
          title="Nothing scheduled yet"
          description="This content has no schedules. Use the form above to queue it up for publishing."
        />
      );
    }

    return (
      <ul className="flex flex-col gap-3">
        {schedules.data.map((schedule) => (
          <li key={schedule.id}>
            <Card>
              <CardHeader className="flex-row items-start justify-between gap-4 pb-3">
                <div className="flex min-w-0 flex-col gap-1">
                  <CardTitle className="truncate">
                    {schedule.cronExpression
                      ? `Repeats: ${schedule.cronExpression}`
                      : formatDateTime(schedule.scheduledFor)}
                  </CardTitle>
                  <div className="flex flex-wrap items-center gap-2">
                    <Badge variant={statusVariant[schedule.status] ?? "default"}>
                      {schedule.status}
                    </Badge>
                    <span className="text-xs text-muted-foreground">{schedule.timezone}</span>
                    {schedule.attempts > 0 ? (
                      <span className="text-xs text-muted-foreground">
                        {schedule.attempts}/{schedule.maxRetries} attempts
                      </span>
                    ) : null}
                  </div>
                </div>

                {CANCELLABLE.includes(schedule.status) ? (
                  <Button
                    variant="ghost"
                    size="sm"
                    loading={cancel.isPending && cancel.variables?.scheduleId === schedule.id}
                    onClick={() => cancel.mutate({ contentId, scheduleId: schedule.id })}
                  >
                    Cancel
                  </Button>
                ) : null}
              </CardHeader>

              {schedule.cronExpression && schedule.scheduledFor ? (
                <CardContent className="pt-0">
                  <p className="text-sm text-muted-foreground">
                    Next run {formatDateTime(schedule.scheduledFor)}
                  </p>
                </CardContent>
              ) : null}

              {schedule.lastError ? (
                <CardContent className="pt-0">
                  <p role="alert" className="text-sm text-danger">
                    {schedule.lastError}
                  </p>
                </CardContent>
              ) : null}
            </Card>
          </li>
        ))}
      </ul>
    );
  }

  return (
    <div className="flex flex-col gap-4">
      {picker}
      {body()}
    </div>
  );
}
