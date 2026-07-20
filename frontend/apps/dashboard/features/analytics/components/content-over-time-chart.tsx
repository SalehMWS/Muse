"use client";

import { CalendarRange } from "lucide-react";
import { CartesianGrid, Line, LineChart, ResponsiveContainer, Tooltip, XAxis, YAxis } from "recharts";

import { EmptyState } from "@/components/shared/empty-state";
import { ErrorCard } from "@/components/shared/error-card";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { ApiError } from "@/lib/api/errors";

import { useAnalyticsSummaryQuery } from "../queries/use-analytics-query";
import { TREND_DAYS } from "../types/summary";
import { AXIS_LINE, AXIS_TICK, ChartTooltipCard, toChartNumber } from "./chart-tooltip";

/** Single series, so one color and no legend — the card title names the measure. */
const SERIES_COLOR = "var(--primary)";

export function ContentOverTimeChart() {
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

  const { perDay, sampledContent, contentInTrendWindow } = query.data;

  if (sampledContent === 0) {
    return (
      <ChartShell>
        <EmptyState
          icon={CalendarRange}
          title="No content yet"
          description="Create some content and your publishing cadence will show up here."
        />
      </ChartShell>
    );
  }

  // Content exists, but none of it lands inside the window — a flat zero line
  // would read as a rendering bug, so say what is actually true.
  if (contentInTrendWindow === 0) {
    return (
      <ChartShell>
        <EmptyState
          icon={CalendarRange}
          title={`Nothing created in the last ${TREND_DAYS} days`}
          description="Your existing content was all created before this window."
        />
      </ChartShell>
    );
  }

  const peak = perDay.reduce((max, point) => Math.max(max, point.count), 0);
  const description = `${contentInTrendWindow} item${
    contentInTrendWindow === 1 ? "" : "s"
  } created over the last ${TREND_DAYS} days, peaking at ${peak} in a single day.`;

  return (
    <ChartShell>
      <div
        role="img"
        aria-label={`Line chart of content created per day. ${description}`}
        className="h-[260px] w-full"
      >
        <ResponsiveContainer width="100%" height="100%">
          <LineChart data={perDay} margin={{ top: 16, right: 12, bottom: 4, left: 0 }}>
            <CartesianGrid stroke="var(--border)" strokeDasharray="3 3" vertical={false} />
            <XAxis
              dataKey="label"
              tick={AXIS_TICK}
              axisLine={AXIS_LINE}
              tickLine={false}
              minTickGap={16}
            />
            <YAxis
              allowDecimals={false}
              tick={AXIS_TICK}
              axisLine={false}
              tickLine={false}
              width={32}
            />
            <Tooltip
              cursor={{ stroke: "var(--border)", strokeWidth: 1 }}
              content={({ active, payload, label }) => {
                if (!active || !payload?.length) return null;
                return (
                  <ChartTooltipCard
                    title={String(label)}
                    value={toChartNumber(payload[0]?.value)}
                    valueLabel="created"
                    swatch={SERIES_COLOR}
                  />
                );
              }}
            />
            <Line
              type="monotone"
              dataKey="count"
              name="Content created"
              stroke={SERIES_COLOR}
              strokeWidth={2}
              dot={{ r: 3, fill: SERIES_COLOR, stroke: SERIES_COLOR }}
              // 2px surface ring keeps the hovered point legible over the line.
              activeDot={{ r: 5, fill: SERIES_COLOR, stroke: "var(--card)", strokeWidth: 2 }}
              isAnimationActive={false}
            />
          </LineChart>
        </ResponsiveContainer>
      </div>
    </ChartShell>
  );
}

function ChartShell({ children }: { children: React.ReactNode }) {
  return (
    <Card>
      <CardHeader className="pb-2">
        <CardTitle className="text-base">Content created per day</CardTitle>
        <CardDescription>New content over the last {TREND_DAYS} days.</CardDescription>
      </CardHeader>
      <CardContent>{children}</CardContent>
    </Card>
  );
}
