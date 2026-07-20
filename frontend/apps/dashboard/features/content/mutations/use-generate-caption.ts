"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

import { ApiError } from "@/lib/api/errors";

import { contentService } from "../services/content.service";
import { contentKeys } from "../queries/use-content-query";

export function useGenerateCaption(id: string, onGenerated?: () => void) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (prompt?: string) => contentService.generateCaption(id, prompt),
    onSuccess: async (content) => {
      queryClient.setQueryData(contentKeys.detail(id), content);
      await queryClient.invalidateQueries({ queryKey: contentKeys.lists() });
      toast.success("Caption generated.");
      onGenerated?.();
    },
    onError: (error) => {
      // Caption generation needs AI_API_KEY on the backend. When it is unset the
      // backend explains why, so surface that message verbatim instead of a
      // generic fallback that would hide the real cause.
      toast.error(
        error instanceof ApiError ? error.message : "Could not generate a caption.",
      );
    },
  });
}
