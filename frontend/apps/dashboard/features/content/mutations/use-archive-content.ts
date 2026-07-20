"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

import { ApiError } from "@/lib/api/errors";

import { contentService } from "../services/content.service";
import { contentKeys } from "../queries/use-content-query";

export function useArchiveContent() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => contentService.archive(id),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: contentKeys.lists() });
      toast.success("Content archived.");
    },
    onError: (error) => {
      toast.error(error instanceof ApiError ? error.message : "Could not archive content.");
    },
  });
}
