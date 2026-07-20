"use client";

import { useQuery } from "@tanstack/react-query";

import { contentService } from "../services/content.service";
import { contentKeys } from "./use-content-query";

export const mediaKeys = {
  list: (contentId: string) => [...contentKeys.detail(contentId), "media"] as const,
};

export function useMediaQuery(contentId: string) {
  return useQuery({
    queryKey: mediaKeys.list(contentId),
    queryFn: ({ signal }) => contentService.listMedia(contentId, signal),
    enabled: Boolean(contentId),
  });
}
