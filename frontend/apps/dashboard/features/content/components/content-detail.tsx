"use client";

import { useId, useState } from "react";

import { ErrorCard } from "@/components/shared/error-card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import { Textarea } from "@/components/ui/textarea";
import { ApiError } from "@/lib/api/errors";

import { useUpdateContent } from "../mutations/use-update-content";
import { useContentQuery } from "../queries/use-content-query";
import type { ContentStatus } from "../types/content";

const statusVariant: Record<ContentStatus, "default" | "primary" | "success" | "warning"> = {
  draft: "default",
  scheduled: "warning",
  published: "success",
  archived: "default",
};

const MAX_CAPTION = 2200;

export function ContentDetail({ contentId }: { contentId: string }) {
  const query = useContentQuery(contentId);
  const update = useUpdateContent(contentId);
  const captionId = useId();

  const serverCaption = query.data?.caption ?? "";
  const [caption, setCaption] = useState(serverCaption);
  const [syncedCaption, setSyncedCaption] = useState(serverCaption);

  // Adjust state during render when the server value changes (initial load, AI
  // caption generation) so local edits are not stranded on a stale caption.
  if (serverCaption !== syncedCaption) {
    setSyncedCaption(serverCaption);
    setCaption(serverCaption);
  }

  if (query.isPending) {
    return (
      <Card>
        <CardHeader className="gap-3">
          <Skeleton className="h-6 w-64" />
          <Skeleton className="h-4 w-40" />
        </CardHeader>
        <CardContent>
          <Skeleton className="h-32 w-full" />
        </CardContent>
      </Card>
    );
  }

  if (query.isError) {
    return (
      <ErrorCard
        message={query.error instanceof ApiError ? query.error.message : "Could not load this content."}
        action={
          <Button variant="outline" size="sm" onClick={() => query.refetch()}>
            Try again
          </Button>
        }
      />
    );
  }

  const content = query.data;
  const dirty = caption !== serverCaption;
  const tooLong = caption.length > MAX_CAPTION;

  return (
    <Card>
      <CardHeader className="gap-3">
        <CardTitle>{content.title}</CardTitle>
        <div className="flex flex-wrap items-center gap-2">
          <Badge variant={statusVariant[content.status] ?? "default"}>{content.status}</Badge>
          <span className="text-xs text-muted-foreground">{content.contentType}</span>
          <span className="text-xs text-muted-foreground">{content.language}</span>
          {content.tags.length > 0 ? (
            content.tags.map((tag) => (
              <Badge key={tag} variant="outline">
                #{tag}
              </Badge>
            ))
          ) : (
            <span className="text-xs text-muted-foreground">No tags yet</span>
          )}
        </div>
      </CardHeader>

      <CardContent className="flex flex-col gap-3">
        <div className="flex flex-col gap-1.5">
          <Label htmlFor={captionId}>Caption</Label>
          <Textarea
            id={captionId}
            value={caption}
            rows={8}
            placeholder="Write a caption, or let NovaFlow generate one for you."
            aria-invalid={tooLong || undefined}
            aria-describedby={tooLong ? `${captionId}-error` : undefined}
            onChange={(event) => setCaption(event.target.value)}
          />
          {tooLong ? (
            <p id={`${captionId}-error`} role="alert" className="text-xs text-danger">
              Instagram captions cap out at {MAX_CAPTION} characters.
            </p>
          ) : (
            <p className="text-xs text-muted-foreground">
              {caption.length} / {MAX_CAPTION} characters
            </p>
          )}
        </div>

        <div className="flex flex-wrap gap-2">
          <Button
            loading={update.isPending}
            disabled={!dirty || tooLong}
            onClick={() => update.mutate({ caption })}
          >
            Save caption
          </Button>
          {dirty ? (
            <Button variant="ghost" onClick={() => setCaption(serverCaption)}>
              Discard changes
            </Button>
          ) : null}
        </div>
      </CardContent>
    </Card>
  );
}
