"use client";

import { Camera } from "lucide-react";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";

import { useAnalyticsSummaryQuery } from "../queries/use-analytics-query";

/**
 * Third of the three real sources: GET /instagram/accounts. Only the counts are
 * derived here — total connected, and how many are still in `active` status.
 */
export function ConnectedAccounts() {
  const query = useAnalyticsSummaryQuery();

  if (query.isPending) {
    return <Skeleton className="h-24" />;
  }

  // A failed summary is already reported by the charts; keep this tile quiet
  // rather than stacking a third error card on the same screen.
  if (query.isError || !query.data.accounts) {
    return (
      <Card>
        <CardHeader className="pb-2">
          <CardTitle className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
            <Camera className="size-4" aria-hidden />
            Connected accounts
          </CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground">Currently unavailable.</p>
        </CardContent>
      </Card>
    );
  }

  const { total, active } = query.data.accounts;

  return (
    <Card>
      <CardHeader className="pb-2">
        <CardTitle className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
          <Camera className="size-4" aria-hidden />
          Connected accounts
        </CardTitle>
      </CardHeader>
      <CardContent>
        <p className="text-3xl font-semibold tabular-nums">{total}</p>
        <p className="mt-1 text-xs text-muted-foreground">
          {total === 0
            ? "No Instagram accounts connected yet."
            : `${active} of ${total} currently active.`}
        </p>
      </CardContent>
    </Card>
  );
}
