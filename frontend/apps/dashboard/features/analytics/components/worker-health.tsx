"use client";

import { AlertOctagon, CheckCircle2, Layers, RotateCcw, XCircle, type LucideIcon } from "lucide-react";

import { ErrorCard } from "@/components/shared/error-card";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { ApiError } from "@/lib/api/errors";

import { useAnalyticsSummaryQuery } from "../queries/use-analytics-query";
import type { WorkerHealthSummary } from "../types/summary";

/**
 * Status colors here are real status colors — success for succeeded, danger for
 * failed and dead-lettered, warning for retried. Each tile ships an icon and a
 * text label, so state is never carried by color alone.
 */
const TILES: Array<{
  key: keyof Omit<WorkerHealthSummary, "successRate">;
  label: string;
  icon: LucideIcon;
  tone: string;
  hint: string;
}> = [
  { key: "processed", label: "Processed", icon: Layers, tone: "text-muted-foreground", hint: "Jobs the pool has taken off the queue." },
  { key: "succeeded", label: "Succeeded", icon: CheckCircle2, tone: "text-success", hint: "Completed without error." },
  { key: "failed", label: "Failed", icon: XCircle, tone: "text-danger", hint: "Ended in an error." },
  { key: "retried", label: "Retried", icon: RotateCcw, tone: "text-warning", hint: "Re-queued after a failure." },
  { key: "deadLettered", label: "Dead-lettered", icon: AlertOctagon, tone: "text-danger", hint: "Gave up after exhausting retries." },
];

export function WorkerHealth() {
  const query = useAnalyticsSummaryQuery();

  if (query.isPending) {
    return (
      <Shell>
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-5">
          {TILES.map((tile) => (
            <Skeleton key={tile.key} className="h-24" />
          ))}
        </div>
      </Shell>
    );
  }

  if (query.isError) {
    return (
      <Shell>
        <ErrorCard
          message={
            query.error instanceof ApiError ? query.error.message : "Could not load worker stats."
          }
          action={
            <Button variant="outline" size="sm" onClick={() => query.refetch()}>
              Try again
            </Button>
          }
        />
      </Shell>
    );
  }

  const worker = query.data.worker;

  // The summary loaded but /worker/stats did not. Say so rather than showing zeros.
  if (!worker) {
    return (
      <Shell>
        <ErrorCard
          message="Worker stats are unavailable right now, so queue health cannot be shown."
          action={
            <Button variant="outline" size="sm" onClick={() => query.refetch()}>
              Try again
            </Button>
          }
        />
      </Shell>
    );
  }

  return (
    <Shell>
      <div className="flex flex-col gap-4">
        <SuccessRate worker={worker} />

        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-5">
          {TILES.map((tile) => {
            const Icon = tile.icon;
            return (
              <Card key={tile.key}>
                <CardHeader className="pb-2">
                  <CardTitle className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
                    <Icon className={`size-4 ${tile.tone}`} aria-hidden />
                    {tile.label}
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-3xl font-semibold tabular-nums">{worker[tile.key]}</p>
                  <p className="mt-1 text-xs text-muted-foreground">{tile.hint}</p>
                </CardContent>
              </Card>
            );
          })}
        </div>
      </div>
    </Shell>
  );
}

function SuccessRate({ worker }: { worker: WorkerHealthSummary }) {
  // No jobs processed means there is no rate — not 0%, not 100%.
  if (worker.successRate === null) {
    return (
      <Card>
        <CardContent className="p-6">
          <p className="text-sm font-medium">Success rate</p>
          <p className="mt-1 text-sm text-muted-foreground">
            No jobs have been processed yet, so there is no rate to report.
          </p>
        </CardContent>
      </Card>
    );
  }

  const rate = worker.successRate;
  const rounded = Math.round(rate * 10) / 10;

  return (
    <Card>
      <CardContent className="flex flex-col gap-3 p-6">
        <div className="flex flex-wrap items-baseline justify-between gap-2">
          <p className="text-sm font-medium">Success rate</p>
          <p className="text-xs text-muted-foreground tabular-nums">
            {worker.succeeded} succeeded of {worker.processed} processed
          </p>
        </div>

        <p className="text-4xl font-semibold tabular-nums">{rounded}%</p>

        <div
          className="h-2 w-full overflow-hidden rounded-full bg-muted"
          role="meter"
          aria-valuenow={rounded}
          aria-valuemin={0}
          aria-valuemax={100}
          aria-label={`Job success rate: ${rounded} percent`}
        >
          <div
            className="h-full rounded-full"
            style={{
              width: `${Math.max(0, Math.min(100, rate))}%`,
              backgroundColor: "var(--success)",
            }}
          />
        </div>
      </CardContent>
    </Card>
  );
}

function Shell({ children }: { children: React.ReactNode }) {
  return (
    <section className="flex flex-col gap-4" aria-labelledby="worker-health-heading">
      <div className="flex flex-col gap-1">
        <h2 id="worker-health-heading" className="text-lg font-semibold tracking-tight">
          Worker queue health
        </h2>
        <CardDescription>Counters reported by the background job pool.</CardDescription>
      </div>
      {children}
    </section>
  );
}
