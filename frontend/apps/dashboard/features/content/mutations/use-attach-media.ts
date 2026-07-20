"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

import { ApiError } from "@/lib/api/errors";

import { contentService } from "../services/content.service";
import { mediaKeys } from "../queries/use-media-query";
import type { AttachMediaInput } from "../types/media";

export function useAttachMedia(contentId: string, onAttached?: () => void) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (input: AttachMediaInput) => contentService.attachMedia(contentId, input),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: mediaKeys.list(contentId) });
      toast.success("Media attached.");
      onAttached?.();
    },
    onError: (error) => {
      toast.error(error instanceof ApiError ? error.message : "Could not attach media.");
    },
  });
}
