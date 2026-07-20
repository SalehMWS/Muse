"use client";

import { FileText } from "lucide-react";
import { useState } from "react";

import { EmptyState } from "@/components/shared/empty-state";
import { ErrorCard } from "@/components/shared/error-card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { ApiError } from "@/lib/api/errors";

import { useArchiveContent } from "../mutations/use-archive-content";
import { useContentListQuery } from "../queries/use-content-query";
import type { ContentStatus } from "../types/content";

const statusVariant: Record<ContentStatus, "default" | "primary" | "success" | "warning"> = {
  draft: "default",
  scheduled: "warning",
  published: "success",
  archived: "default",
};

const FILTERS: Array<{ label: string; value: ContentStatus | undefined }> = [
  { label: "All", value: undefined },
  { label: "Drafts", value: "draft" },
  { label: "Scheduled", value: "scheduled" },
  { label: "Published", value: "published" },
];

export function ContentList() {
  const [status, setStatus] = useState<ContentStatus | undefined>(undefined);
  const query = useContentListQuery({ status, limit: 20 });
  const archive = useArchiveContent();

  if (query.isPending) {
    return (
      <div className="flex flex-col gap-3">
        {Array.from({ length: 3 }).map((_, index) => (
          <Skeleton key={index} className="h-28 w-full" />
        ))}
      </div>
    );
  }

  if (query.isError) {
    return (
      <ErrorCard
        message={query.error instanceof ApiError ? query.error.message : "Could not load content."}
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
    <div className="flex flex-col gap-4">
      <div className="flex flex-wrap gap-2">
        {FILTERS.map((filter) => (
          <Button
            key={filter.label}
            size="sm"
            variant={status === filter.value ? "primary" : "outline"}
            onClick={() => setStatus(filter.value)}
          >
            {filter.label}
          </Button>
        ))}
      </div>

      {items.length === 0 ? (
        <EmptyState
          icon={FileText}
          title="No content yet"
          description="Create your first draft and NovaFlow will help you caption and schedule it."
        />
      ) : (
        <ul className="flex flex-col gap-3">
          {items.map((content) => (
            <li key={content.id}>
              <Card interactive>
                <CardHeader className="flex-row items-start justify-between gap-4 pb-3">
                  <div className="flex min-w-0 flex-col gap-1">
                    <CardTitle className="truncate">{content.title}</CardTitle>
                    <div className="flex flex-wrap items-center gap-2">
                      <Badge variant={statusVariant[content.status] ?? "default"}>{content.status}</Badge>
                      <span className="text-xs text-muted-foreground">{content.contentType}</span>
                      {content.tags.slice(0, 3).map((tag) => (
                        <Badge key={tag} variant="outline">
                          #{tag}
                        </Badge>
                      ))}
                    </div>
                  </div>
                  {content.status !== "archived" ? (
                    <Button
                      variant="ghost"
                      size="sm"
                      loading={archive.isPending && archive.variables === content.id}
                      onClick={() => archive.mutate(content.id)}
                    >
                      Archive
                    </Button>
                  ) : null}
                </CardHeader>
                {content.caption ? (
                  <CardContent className="pt-0">
                    <p className="line-clamp-2 text-sm text-muted-foreground">{content.caption}</p>
                  </CardContent>
                ) : null}
              </Card>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
