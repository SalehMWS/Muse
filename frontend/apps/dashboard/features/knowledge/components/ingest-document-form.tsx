"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { Plus } from "lucide-react";
import { useId, useState } from "react";
import { useForm } from "react-hook-form";

import { TextField } from "@/components/forms/text-field";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";

import { useIngestDocument } from "../mutations/use-ingest-document";
import {
  ingestDocumentSchema,
  type IngestDocumentInput,
} from "../schemas/ingest-document.schema";

export function IngestDocumentForm() {
  const [open, setOpen] = useState(false);
  const contentId = useId();
  const contentErrorId = `${contentId}-error`;

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<IngestDocumentInput>({
    resolver: zodResolver(ingestDocumentSchema),
    defaultValues: { title: "", source: "", content: "" },
  });

  const ingest = useIngestDocument(() => {
    reset();
    setOpen(false);
  });

  if (!open) {
    return (
      <div>
        <Button onClick={() => setOpen(true)}>
          <Plus aria-hidden /> Add document
        </Button>
      </div>
    );
  }

  return (
    <Card className="w-full">
      <CardHeader>
        <CardTitle>Add document</CardTitle>
      </CardHeader>
      <CardContent>
        <form
          className="flex flex-col gap-4"
          onSubmit={handleSubmit((values) => ingest.mutate(values))}
          noValidate
        >
          <TextField label="Title" required error={errors.title?.message} {...register("title")} />

          <TextField
            label="Source"
            hint="Optional. Where this came from, for example a URL or a file name."
            error={errors.source?.message}
            {...register("source")}
          />

          <div className="flex flex-col gap-1.5">
            <Label htmlFor={contentId} required>
              Content
            </Label>
            <Textarea
              id={contentId}
              rows={8}
              required
              placeholder="Paste the text you want NovaFlow to retrieve from."
              aria-invalid={Boolean(errors.content) || undefined}
              aria-describedby={errors.content ? contentErrorId : undefined}
              {...register("content")}
            />
            {errors.content ? (
              <p id={contentErrorId} role="alert" className="text-xs text-danger">
                {errors.content.message}
              </p>
            ) : (
              <p className="text-xs text-muted-foreground">
                Longer documents are chunked automatically before they are embedded.
              </p>
            )}
          </div>

          <div className="flex gap-2">
            <Button type="submit" loading={ingest.isPending}>
              Add to knowledge base
            </Button>
            <Button
              type="button"
              variant="ghost"
              onClick={() => {
                reset();
                setOpen(false);
              }}
            >
              Cancel
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  );
}
