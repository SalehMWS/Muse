"use client";

import { useMutation } from "@tanstack/react-query";
import { toast } from "sonner";

import { ApiError } from "@/lib/api/errors";

import type { KnowledgeQueryInput } from "../schemas/ingest-document.schema";
import { knowledgeService } from "../services/knowledge.service";

export function useQueryKnowledge() {
  return useMutation({
    mutationFn: (input: KnowledgeQueryInput) => knowledgeService.query(input),
    onError: (error) => {
      toast.error(
        error instanceof ApiError ? error.message : "Could not search the knowledge base.",
      );
    },
  });
}
