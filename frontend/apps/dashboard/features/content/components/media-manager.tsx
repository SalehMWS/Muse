"use client";

import { ImageOff, Images } from "lucide-react";
import { useId, useState } from "react";

import { EmptyState } from "@/components/shared/empty-state";
import { ErrorCard } from "@/components/shared/error-card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select } from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import { ApiError } from "@/lib/api/errors";

import { useAttachMedia } from "../mutations/use-attach-media";
import { useDeleteMedia } from "../mutations/use-delete-media";
import { useMediaQuery } from "../queries/use-media-query";
import type { Media, MediaType } from "../types/media";

function fileNameOf(url: string): string {
  try {
    const path = new URL(url).pathname;
    return decodeURIComponent(path.split("/").filter(Boolean).pop() ?? url);
  } catch {
    return url;
  }
}

function isValidUrl(value: string): boolean {
  try {
    const parsed = new URL(value);
    return parsed.protocol === "http:" || parsed.protocol === "https:";
  } catch {
    return false;
  }
}

/**
 * Remote hosts are arbitrary and next/image is not configured for them, so the
 * preview is a plain img that degrades to a filename + link when it cannot load.
 */
function MediaPreview({ media }: { media: Media }) {
  const [broken, setBroken] = useState(false);
  const name = fileNameOf(media.url);

  if (media.mediaType === "video" || broken) {
    return (
      <div className="flex size-16 shrink-0 items-center justify-center rounded-md border border-border bg-secondary">
        <ImageOff className="size-5 text-muted-foreground" aria-hidden />
      </div>
    );
  }

  return (
    // eslint-disable-next-line @next/next/no-img-element
    <img
      src={media.url}
      alt={`Attached media: ${name}`}
      className="size-16 shrink-0 rounded-md border border-border object-cover"
      onError={() => setBroken(true)}
    />
  );
}

export function MediaManager({ contentId }: { contentId: string }) {
  const urlId = useId();
  const typeId = useId();

  const query = useMediaQuery(contentId);
  const remove = useDeleteMedia(contentId);

  const [url, setUrl] = useState("");
  const [mediaType, setMediaType] = useState<MediaType>("image");
  const [touched, setTouched] = useState(false);

  const attach = useAttachMedia(contentId, () => {
    setUrl("");
    setTouched(false);
  });

  const urlError = touched && !isValidUrl(url) ? "Enter a full http(s) URL." : undefined;

  function submit() {
    setTouched(true);
    if (!isValidUrl(url)) return;
    attach.mutate({ url: url.trim(), mediaType, position: query.data?.length ?? 0 });
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Media</CardTitle>
      </CardHeader>

      <CardContent className="flex flex-col gap-4">
        <form
          className="flex flex-col gap-3 sm:flex-row sm:items-end"
          onSubmit={(event) => {
            event.preventDefault();
            submit();
          }}
          noValidate
        >
          <div className="flex flex-1 flex-col gap-1.5">
            <Label htmlFor={urlId} required>
              Media URL
            </Label>
            <Input
              id={urlId}
              value={url}
              placeholder="https://cdn.example.com/photo.jpg"
              aria-invalid={Boolean(urlError) || undefined}
              aria-describedby={urlError ? `${urlId}-error` : undefined}
              onChange={(event) => setUrl(event.target.value)}
              onBlur={() => setTouched(true)}
            />
            {urlError ? (
              <p id={`${urlId}-error`} role="alert" className="text-xs text-danger">
                {urlError}
              </p>
            ) : null}
          </div>

          <div className="flex flex-col gap-1.5 sm:w-40">
            <Label htmlFor={typeId}>Type</Label>
            <Select
              id={typeId}
              value={mediaType}
              onChange={(event) => setMediaType(event.target.value as MediaType)}
            >
              <option value="image">Image</option>
              <option value="video">Video</option>
            </Select>
          </div>

          <Button type="submit" loading={attach.isPending}>
            Add media
          </Button>
        </form>

        {query.isPending ? (
          <div className="flex flex-col gap-2">
            {Array.from({ length: 2 }).map((_, index) => (
              <Skeleton key={index} className="h-20 w-full" />
            ))}
          </div>
        ) : query.isError ? (
          <ErrorCard
            message={query.error instanceof ApiError ? query.error.message : "Could not load media."}
            action={
              <Button variant="outline" size="sm" onClick={() => query.refetch()}>
                Try again
              </Button>
            }
          />
        ) : query.data.length === 0 ? (
          <EmptyState
            icon={Images}
            title="No media attached"
            description="Add at least one image or video URL — publishing to Instagram requires media."
          />
        ) : (
          <ul className="flex flex-col gap-3">
            {query.data.map((media) => (
              <li
                key={media.id}
                className="flex items-center gap-3 rounded-lg border border-border p-3"
              >
                <MediaPreview media={media} />

                <div className="flex min-w-0 flex-1 flex-col gap-1">
                  <a
                    href={media.url}
                    target="_blank"
                    rel="noreferrer noopener"
                    className="truncate text-sm text-primary underline-offset-4 hover:underline"
                  >
                    {fileNameOf(media.url)}
                  </a>
                  <div className="flex items-center gap-2">
                    <Badge variant="outline">{media.mediaType}</Badge>
                    <span className="text-xs text-muted-foreground">position {media.position}</span>
                  </div>
                </div>

                <Button
                  variant="ghost"
                  size="sm"
                  loading={remove.isPending && remove.variables === media.id}
                  onClick={() => remove.mutate(media.id)}
                >
                  Remove
                </Button>
              </li>
            ))}
          </ul>
        )}
      </CardContent>
    </Card>
  );
}
