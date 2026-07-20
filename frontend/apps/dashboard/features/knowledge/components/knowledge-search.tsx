"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { Copy, Search, SearchX } from "lucide-react";
import { useId } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";

import { EmptyState } from "@/components/shared/empty-state";
import { ErrorCard } from "@/components/shared/error-card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import { ApiError } from "@/lib/api/errors";

import { useQueryKnowledge } from "../mutations/use-query-knowledge";
import {
  knowledgeQuerySchema,
  type KnowledgeQueryInput,
} from "../schemas/ingest-document.schema";
import type { QueryHit } from "../types/document";

const TOP_K_OPTIONS = [3, 5, 10, 20];

function formatScore(score: number): string {
  return Number.isFinite(score) ? score.toFixed(2) : "—";
}

function scoreWidth(score: number): string {
  if (!Number.isFinite(score)) return "0%";
  return `${Math.min(Math.max(score, 0), 1) * 100}%`;
}

export function KnowledgeSearch() {
  const queryId = useId();
  const queryErrorId = `${queryId}-error`;
  const topKId = useId();

  const search = useQueryKnowledge();

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<KnowledgeQueryInput>({
    resolver: zodResolver(knowledgeQuerySchema),
    defaultValues: { query: "", topK: 5 },
  });

  const submit = handleSubmit((values) => search.mutate(values));

  return (
    <Card>
      <CardHeader>
        <CardTitle>Search your knowledge base</CardTitle>
        <p className="text-sm text-muted-foreground">
          Retrieval runs over every indexed chunk. The context below is exactly what gets handed to
          the model when it writes a caption.
        </p>
      </CardHeader>

      <CardContent className="flex flex-col gap-6">
        <form className="flex flex-col gap-4" onSubmit={submit} noValidate>
          <div className="flex flex-col gap-3 sm:flex-row sm:items-end">
            <div className="flex min-w-0 flex-1 flex-col gap-1.5">
              <Label htmlFor={queryId} required>
                Question
              </Label>
              <Input
                id={queryId}
                type="search"
                placeholder="What tone do we use for product launches?"
                aria-invalid={Boolean(errors.query) || undefined}
                aria-describedby={errors.query ? queryErrorId : undefined}
                required
                {...register("query")}
              />
            </div>

            <div className="flex flex-col gap-1.5">
              <Label htmlFor={topKId}>Results</Label>
              <select
                id={topKId}
                className="h-10 rounded-md border border-border bg-background px-3 text-sm"
                {...register("topK", { valueAsNumber: true })}
              >
                {TOP_K_OPTIONS.map((option) => (
                  <option key={option} value={option}>
                    {option}
                  </option>
                ))}
              </select>
            </div>

            <Button type="submit" loading={search.isPending}>
              <Search aria-hidden /> Search
            </Button>
          </div>

          {errors.query ? (
            <p id={queryErrorId} role="alert" className="text-xs text-danger">
              {errors.query.message}
            </p>
          ) : null}
        </form>

        <SearchResults
          status={search.status}
          hits={search.data?.hits ?? []}
          context={search.data?.context ?? ""}
          error={search.error}
          onRetry={submit}
        />
      </CardContent>
    </Card>
  );
}

function SearchResults({
  status,
  hits,
  context,
  error,
  onRetry,
}: {
  status: "idle" | "pending" | "error" | "success";
  hits: QueryHit[];
  context: string;
  error: Error | null;
  onRetry: () => void;
}) {
  if (status === "idle") {
    return (
      <EmptyState
        icon={Search}
        title="Nothing searched yet"
        description="Ask a question above to see which indexed chunks NovaFlow would retrieve, and how strongly each one matches."
      />
    );
  }

  if (status === "pending") {
    return (
      <div className="flex flex-col gap-3">
        {Array.from({ length: 3 }).map((_, index) => (
          <Skeleton key={index} className="h-24 w-full" />
        ))}
      </div>
    );
  }

  if (status === "error") {
    return (
      <ErrorCard
        message={error instanceof ApiError ? error.message : "Could not run that search."}
        action={
          <Button variant="outline" size="sm" onClick={onRetry}>
            Try again
          </Button>
        }
      />
    );
  }

  if (hits.length === 0) {
    return (
      <EmptyState
        icon={SearchX}
        title="No matches"
        description="Nothing in the knowledge base is close enough to that question. Try different wording, or add a document that covers it."
      />
    );
  }

  return (
    <div className="flex flex-col gap-6">
      <section className="flex flex-col gap-3" aria-label="Retrieved chunks">
        <div className="flex items-baseline justify-between gap-2">
          <h3 className="text-sm font-semibold">
            {hits.length} {hits.length === 1 ? "match" : "matches"}
          </h3>
          <span className="text-xs text-muted-foreground">ranked by similarity</span>
        </div>

        <ol className="flex flex-col gap-3">
          {hits.map((hit, index) => (
            <li key={`${hit.documentId}-${hit.chunkIndex}`}>
              <article className="flex flex-col gap-3 rounded-lg border border-border bg-card p-4">
                <div className="flex flex-wrap items-center gap-2">
                  <Badge variant="outline">#{index + 1}</Badge>
                  <Badge variant="primary">{formatScore(hit.score)}</Badge>
                  <span className="text-xs text-muted-foreground">chunk {hit.chunkIndex}</span>
                  <span
                    className="truncate font-mono text-xs text-muted-foreground"
                    title={hit.documentId}
                  >
                    {hit.documentId}
                  </span>
                </div>

                <div
                  className="h-1 w-full overflow-hidden rounded-full bg-muted"
                  role="img"
                  aria-label={`Similarity score ${formatScore(hit.score)}`}
                >
                  <div className="h-full bg-primary" style={{ width: scoreWidth(hit.score) }} />
                </div>

                <p className="whitespace-pre-wrap text-sm leading-relaxed text-foreground">
                  {hit.content}
                </p>
              </article>
            </li>
          ))}
        </ol>
      </section>

      {context ? <ContextPanel context={context} /> : null}
    </div>
  );
}

function ContextPanel({ context }: { context: string }) {
  async function copy() {
    try {
      await navigator.clipboard.writeText(context);
      toast.success("Context copied.");
    } catch {
      toast.error("Could not copy the context.");
    }
  }

  return (
    <section className="flex flex-col gap-2" aria-label="Assembled context">
      <div className="flex items-center justify-between gap-2">
        <div className="flex flex-col">
          <h3 className="text-sm font-semibold">Assembled context</h3>
          <p className="text-xs text-muted-foreground">
            {context.length.toLocaleString()} characters sent to the model.
          </p>
        </div>
        <Button type="button" variant="outline" size="sm" onClick={copy}>
          <Copy aria-hidden /> Copy
        </Button>
      </div>
      <pre className="max-h-80 overflow-auto rounded-lg border border-border bg-muted/40 p-4 text-xs leading-relaxed whitespace-pre-wrap text-muted-foreground">
        {context}
      </pre>
    </section>
  );
}
