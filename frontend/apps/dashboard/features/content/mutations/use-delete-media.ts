"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

import { ApiError } from "@/lib/api/errors";

import { contentService } from "../services/content.service";
import { mediaKeys } from "../queries/use-media-query";

export function useDeleteMedia(contentId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (mediaId: string) => contentService.deleteMedia(contentId, mediaId),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: mediaKeys.list(contentId) });
      toast.success("Media removed.");
    },
    onError: (error) => {
      toast.error(error instanceof ApiError ? error.message : "Could not remove media.");
    },
  });
}
