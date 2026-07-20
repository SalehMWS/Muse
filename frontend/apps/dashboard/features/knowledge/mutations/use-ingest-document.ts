"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

import { ApiError } from "@/lib/api/errors";

import { knowledgeKeys } from "../queries/use-documents-query";
import type { IngestDocumentInput } from "../schemas/ingest-document.schema";
import { knowledgeService } from "../services/knowledge.service";

export function useIngestDocument(onIngested?: () => void) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (input: IngestDocumentInput) => knowledgeService.ingest(input),
    onSuccess: async (document) => {
      await queryClient.invalidateQueries({ queryKey: knowledgeKeys.lists() });
      toast.success(`"${document.title}" queued for indexing.`);
      onIngested?.();
    },
    onError: (error) => {
      toast.error(error instanceof ApiError ? error.message : "Could not add the document.");
    },
  });
}
