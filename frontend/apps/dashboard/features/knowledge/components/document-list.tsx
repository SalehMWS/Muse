"use client";

import { Library, Trash2 } from "lucide-react";

import { EmptyState } from "@/components/shared/empty-state";
import { ErrorCard } from "@/components/shared/error-card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { ApiError } from "@/lib/api/errors";

import { useDeleteDocument } from "../mutations/use-delete-document";
import { useDocumentsQuery } from "../queries/use-documents-query";
import type { DocumentStatus } from "../types/document";

const statusVariant: Record<DocumentStatus, "default" | "success" | "warning" | "danger"> = {
  pending: "warning",
  indexed: "success",
  failed: "danger",
};

export function DocumentList() {
  const query = useDocumentsQuery();
  const remove = useDeleteDocument();

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
        message={query.error instanceof ApiError ? query.error.message : "Could not load documents."}
        action={
          <Button variant="outline" size="sm" onClick={() => query.refetch()}>
            Try again
          </Button>
        }
      />
    );
  }

  const documents = query.data;

  if (documents.length === 0) {
    return (
      <EmptyState
        icon={Library}
        title="No documents yet"
        description="Add a document above and NovaFlow will chunk, embed and index it so your captions can cite it."
      />
    );
  }

  return (
    <ul className="flex flex-col gap-3">
      {documents.map((document) => (
        <li key={document.id}>
          <Card interactive>
            <CardHeader className="flex-row items-start justify-between gap-4 pb-3">
              <div className="flex min-w-0 flex-col gap-1">
                <CardTitle className="truncate">{document.title}</CardTitle>
                <div className="flex flex-wrap items-center gap-2">
                  <Badge variant={statusVariant[document.status] ?? "default"}>
                    {document.status}
                  </Badge>
                  <span className="text-xs text-muted-foreground">
                    {document.chunkCount} {document.chunkCount === 1 ? "chunk" : "chunks"}
                  </span>
                  {document.source ? (
                    <span className="truncate text-xs text-muted-foreground">
                      {document.source}
                    </span>
                  ) : null}
                </div>
              </div>
              <Button
                variant="ghost"
                size="sm"
                aria-label={`Delete ${document.title}`}
                loading={remove.isPending && remove.variables === document.id}
                onClick={() => remove.mutate(document.id)}
              >
                <Trash2 aria-hidden /> Delete
              </Button>
            </CardHeader>
            {document.status === "failed" && document.lastError ? (
              <CardContent className="pt-0">
                <p
                  role="alert"
                  className="rounded-md border border-danger/40 bg-danger/5 px-3 py-2 text-xs text-danger"
                >
                  {document.lastError}
                </p>
              </CardContent>
            ) : null}
          </Card>
        </li>
      ))}
    </ul>
  );
}
