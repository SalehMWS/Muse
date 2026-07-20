"use client";

import { Send } from "lucide-react";
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

import { usePublishContent } from "../mutations/use-publish-content";
import { useInstagramAccountsQuery, usePublicationsQuery } from "../queries/use-publications-query";
import type { PublicationStatus, PublishMediaType } from "../types/publication";

const statusVariant: Record<PublicationStatus, "success" | "danger" | "warning"> = {
  published: "success",
  failed: "danger",
  pending: "warning",
};

const MEDIA_TYPES: Array<{ label: string; value: PublishMediaType | "" }> = [
  { label: "Match the content type", value: "" },
  { label: "Image", value: "image" },
  { label: "Carousel", value: "carousel" },
  { label: "Reel", value: "reel" },
];

function formatDate(value: string): string {
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString();
}

export function PublishPanel({ contentId }: { contentId: string }) {
  const accountId = useId();
  const mediaTypeId = useId();

  const accounts = useInstagramAccountsQuery();
  const publications = usePublicationsQuery(contentId);
  const publish = usePublishContent(contentId);

  const [chosenAccount, setChosenAccount] = useState("");
  const [mediaType, setMediaType] = useState<PublishMediaType | "">("");

  // Default to the first account without syncing state in an effect.
  const selectedAccount = chosenAccount || (accounts.data?.[0]?.id ?? "");

  const publishError = publish.isError
    ? publish.error instanceof ApiError
      ? publish.error.message
      : "Could not publish this content."
    : null;

  return (
    <Card>
      <CardHeader>
        <CardTitle>Publish to Instagram</CardTitle>
      </CardHeader>

      <CardContent className="flex flex-col gap-5">
        {accounts.isPending ? (
          <Skeleton className="h-20 w-full" />
        ) : accounts.isError ? (
          <ErrorCard
            message={
              accounts.error instanceof ApiError
                ? accounts.error.message
                : "Could not load your Instagram accounts."
            }
            action={
              <Button variant="outline" size="sm" onClick={() => accounts.refetch()}>
                Try again
              </Button>
            }
          />
        ) : accounts.data.length === 0 ? (
          <p className="rounded-lg border border-dashed border-border p-4 text-sm text-muted-foreground">
            No Instagram account is connected yet. Connect one from the Instagram page before you
            can publish.
          </p>
        ) : (
          <div className="flex flex-col gap-3">
            <div className="flex flex-col gap-3 sm:flex-row sm:items-end">
              <div className="flex flex-1 flex-col gap-1.5">
                <Label htmlFor={accountId}>Instagram account</Label>
                <Select
                  id={accountId}
                  value={selectedAccount}
                  onChange={(event) => setChosenAccount(event.target.value)}
                >
                  {accounts.data.map((account) => (
                    <option key={account.id} value={account.id}>
                      @{account.username}
                      {account.status === "active" ? "" : ` (${account.status})`}
                    </option>
                  ))}
                </Select>
              </div>

              <div className="flex flex-col gap-1.5 sm:w-56">
                <Label htmlFor={mediaTypeId}>Media type</Label>
                <Select
                  id={mediaTypeId}
                  value={mediaType}
                  onChange={(event) => setMediaType(event.target.value as PublishMediaType | "")}
                >
                  {MEDIA_TYPES.map((option) => (
                    <option key={option.value || "auto"} value={option.value}>
                      {option.label}
                    </option>
                  ))}
                </Select>
              </div>

              <Button
                loading={publish.isPending}
                disabled={!selectedAccount}
                onClick={() =>
                  publish.mutate({
                    instagramAccountId: selectedAccount,
                    mediaType: mediaType || undefined,
                  })
                }
              >
                <Send aria-hidden /> Publish
              </Button>
            </div>

            {publishError ? (
              <p role="alert" className="text-sm text-danger">
                {publishError}
              </p>
            ) : null}
          </div>
        )}

        <div className="flex flex-col gap-3">
          <h3 className="text-sm font-semibold">Publication history</h3>

          {publications.isPending ? (
            <div className="flex flex-col gap-2">
              {Array.from({ length: 2 }).map((_, index) => (
                <Skeleton key={index} className="h-16 w-full" />
              ))}
            </div>
          ) : publications.isError ? (
            <ErrorCard
              message={
                publications.error instanceof ApiError
                  ? publications.error.message
                  : "Could not load the publication history."
              }
              action={
                <Button variant="outline" size="sm" onClick={() => publications.refetch()}>
                  Try again
                </Button>
              }
            />
          ) : publications.data.length === 0 ? (
            <EmptyState
              icon={Send}
              title="Not published yet"
              description="Once this content goes live, every attempt shows up here with its result."
            />
          ) : (
            <ul className="flex flex-col gap-2">
              {publications.data.map((publication) => (
                <li
                  key={publication.id}
                  className="flex flex-wrap items-center justify-between gap-3 rounded-lg border border-border p-3"
                >
                  <div className="flex flex-col gap-1">
                    <div className="flex items-center gap-2">
                      <Badge variant={statusVariant[publication.status] ?? "default"}>
                        {publication.status}
                      </Badge>
                      <span className="text-xs text-muted-foreground">
                        {formatDate(publication.publishedAt ?? publication.createdAt)}
                      </span>
                    </div>
                    {publication.status === "failed" ? (
                      <p role="alert" className="text-xs text-danger">
                        This attempt failed. Check the account connection and the attached media,
                        then try again.
                      </p>
                    ) : null}
                  </div>

                  {publication.permalink ? (
                    <a
                      href={publication.permalink}
                      target="_blank"
                      rel="noreferrer noopener"
                      className="text-sm text-primary underline-offset-4 hover:underline"
                    >
                      View post
                    </a>
                  ) : null}
                </li>
              ))}
            </ul>
          )}
        </div>
      </CardContent>
    </Card>
  );
}
