"use client";

import { ErrorCard } from "@/components/shared/error-card";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { useContentListQuery } from "@/features/content/queries/use-content-query";
import { ApiError } from "@/lib/api/errors";
import type { ContentStatus } from "@/features/content/types/content";

const TILES: Array<{ label: string; status?: ContentStatus }> = [
  { label: "Total" },
  { label: "Drafts", status: "draft" },
  { label: "Scheduled", status: "scheduled" },
  { label: "Published", status: "published" },
];

export function OverviewStats() {
  const query = useContentListQuery({ limit: 100 });

  if (query.isPending) {
    return (
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {TILES.map((tile) => (
          <Skeleton key={tile.label} className="h-24" />
        ))}
      </div>
    );
  }

  if (query.isError) {
    return (
      <ErrorCard
        message={query.error instanceof ApiError ? query.error.message : "Could not load your content."}
        action={
          <Button variant="outline" size="sm" onClick={() => query.refetch()}>
            Try again
          </Button>
        }
      />
    );
  }

  const items = query.data.items;

  return (
    <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
      {TILES.map((tile) => {
        const count = tile.status ? items.filter((item) => item.status === tile.status).length : items.length;

        return (
          <Card key={tile.label}>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">{tile.label}</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-3xl font-semibold tabular-nums">{count}</p>
            </CardContent>
          </Card>
        );
      })}
    </div>
  );
}
