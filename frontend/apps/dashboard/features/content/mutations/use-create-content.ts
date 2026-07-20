"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

import { ApiError } from "@/lib/api/errors";

import type { CreateContentInput } from "../schemas/create-content.schema";
import { contentService } from "../services/content.service";
import { contentKeys } from "../queries/use-content-query";

export function useCreateContent(onCreated?: () => void) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (input: CreateContentInput) => contentService.create(input),
    onSuccess: async (content) => {
      await queryClient.invalidateQueries({ queryKey: contentKeys.lists() });
      toast.success(`"${content.title}" created.`);
      onCreated?.();
    },
    onError: (error) => {
      toast.error(error instanceof ApiError ? error.message : "Could not create content.");
    },
  });
}
