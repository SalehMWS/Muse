"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

import { ApiError } from "@/lib/api/errors";

import { contentService } from "../services/content.service";
import { contentKeys } from "../queries/use-content-query";
import type { UpdateContentInput } from "../types/content";

export function useUpdateContent(id: string, onUpdated?: () => void) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (patch: UpdateContentInput) => contentService.update(id, patch),
    onSuccess: async (content) => {
      queryClient.setQueryData(contentKeys.detail(id), content);
      await queryClient.invalidateQueries({ queryKey: contentKeys.lists() });
      toast.success("Content saved.");
      onUpdated?.();
    },
    onError: (error) => {
      toast.error(error instanceof ApiError ? error.message : "Could not save content.");
    },
  });
}
