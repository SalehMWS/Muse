"use client";

import { FileText } from "lucide-react";
import { Bar, BarChart, Cell, LabelList, ResponsiveContainer, Tooltip, XAxis, YAxis } from "recharts";

import { EmptyState } from "@/components/shared/empty-state";
import { ErrorCard } from "@/components/shared/error-card";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { ApiError } from "@/lib/api/errors";

import { useAnalyticsSummaryQuery } from "../queries/use-analytics-query";
import type { ContentStatusKey } from "../types/summary";
import { AXIS_LINE, AXIS_TICK, ChartTooltipCard, toChartNumber } from "./chart-tooltip";

/**
 * Status colors come from the theme tokens so both light and dark resolve at
 * paint time. `--danger` is deliberately not used here: it stays reserved for
 * genuine failure states in worker health, per the status-color rule.
 *
 * A CVD check on this set flags Published(green) vs Scheduled(amber) as a weak
 * pair for protanopia. That is safe here only because color is never the sole
 * identity channel — every bar is named on the x-axis and carries its own value
 * label, so the chart reads correctly in greyscale.
 */
const STATUS_COLORS: Record<ContentStatusKey, string> = {
  draft: "var(--info)",
  scheduled: "var(--warning)",
  published: "var(--success)",
  archived: "var(--primary)",
};

export function ContentStatusChart() {
  const query = useAnalyticsSummaryQuery();

  if (query.isPending) {
    return <ChartShell><Skeleton className="h-[260px] w-full" /></ChartShell>;
  }

  if (query.isError) {
    return (
      <ChartShell>
        <ErrorCard
          message={
            query.error instanceof ApiError ? query.error.message : "Could not load your content."
          }
          action={
            <Button variant="outline" size="sm" onClick={() => query.refetch()}>
              Try again
            </Button>
          }
        />
      </ChartShell>
    );
  }

  const { byStatus, sampledContent } = query.data;

  if (sampledContent === 0) {
    return (
      <ChartShell>
        <EmptyState
          icon={FileText}
          title="No content yet"
          description="Once you create your first draft, its lifecycle will be charted here."
        />
      </ChartShell>
    );
  }

  const description = byStatus.map((slice) => `${slice.label}: ${slice.count}`).join(", ");

  return (
    <ChartShell>
      <div
        role="img"
        aria-label={`Bar chart of content grouped by status. ${description}.`}
        className="h-[260px] w-full"
      >
        <ResponsiveContainer width="100%" height="100%">
          <BarChart data={byStatus} margin={{ top: 24, right: 8, bottom: 4, left: 0 }}>
            <XAxis
              dataKey="label"
              tick={AXIS_TICK}
              axisLine={AXIS_LINE}
              tickLine={false}
              interval={0}
            />
            <YAxis
              allowDecimals={false}
              tick={AXIS_TICK}
              axisLine={false}
              tickLine={false}
              width={32}
            />
            <Tooltip
              cursor={{ fill: "var(--muted)", opacity: 0.5 }}
              content={({ active, payload }) => {
                if (!active || !payload?.length) return null;
                const slice = payload[0]?.payload as { label: string; status: ContentStatusKey };
                return (
                  <ChartTooltipCard
                    title={slice.label}
                    value={toChartNumber(payload[0]?.value)}
                    valueLabel="items"
                    swatch={STATUS_COLORS[slice.status]}
                  />
                );
              }}
            />
            <Bar dataKey="count" radius={[4, 4, 0, 0]} maxBarSize={64} isAnimationActive={false}>
              {byStatus.map((slice) => (
                <Cell key={slice.status} fill={STATUS_COLORS[slice.status]} />
              ))}
              {/* Direct value labels: the secondary encoding that keeps identity
                  off color alone and covers the low-contrast amber on white. */}
              <LabelList
                dataKey="count"
                position="top"
                offset={8}
                fill="var(--foreground)"
                fontSize={12}
              />
            </Bar>
          </BarChart>
        </ResponsiveContainer>
      </div>
    </ChartShell>
  );
}

function ChartShell({ children }: { children: React.ReactNode }) {
  return (
    <Card>
      <CardHeader className="pb-2">
        <CardTitle className="text-base">Content by status</CardTitle>
        <CardDescription>Where your content currently sits in its lifecycle.</CardDescription>
      </CardHeader>
      <CardContent>{children}</CardContent>
    </Card>
  );
}
