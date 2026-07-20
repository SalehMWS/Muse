"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

import { ApiError } from "@/lib/api/errors";
import { contentKeys } from "@/features/content/queries/use-content-query";

import { publishingService } from "../services/publishing.service";
import { publishingKeys } from "../queries/use-publications-query";
import type { PublishInput } from "../types/publication";

export function usePublishContent(contentId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (input: PublishInput) => publishingService.publish(contentId, input),
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: publishingKeys.publications(contentId) }),
        queryClient.invalidateQueries({ queryKey: contentKeys.detail(contentId) }),
        queryClient.invalidateQueries({ queryKey: contentKeys.lists() }),
      ]);
      toast.success("Published to Instagram.");
    },
    onError: async (error) => {
      // A failed attempt is still recorded as a publication row, so refresh the
      // history even though the request itself errored.
      await queryClient.invalidateQueries({ queryKey: publishingKeys.publications(contentId) });
      toast.error(error instanceof ApiError ? error.message : "Could not publish this content.");
    },
  });
}
