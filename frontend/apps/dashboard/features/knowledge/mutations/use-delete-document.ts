"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

import { ApiError } from "@/lib/api/errors";

import { knowledgeKeys } from "../queries/use-documents-query";
import { knowledgeService } from "../services/knowledge.service";

export function useDeleteDocument() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => knowledgeService.remove(id),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: knowledgeKeys.lists() });
      toast.success("Document removed.");
    },
    onError: (error) => {
      toast.error(error instanceof ApiError ? error.message : "Could not remove the document.");
    },
  });
}
