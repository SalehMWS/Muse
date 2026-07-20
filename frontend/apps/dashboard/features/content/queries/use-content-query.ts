"use client";

import { useQuery } from "@tanstack/react-query";

import { contentService } from "../services/content.service";
import type { ContentFilters } from "../types/content";

export const contentKeys = {
  all: ["content"] as const,
  lists: () => [...contentKeys.all, "list"] as const,
  list: (filters: ContentFilters) => [...contentKeys.lists(), filters] as const,
  detail: (id: string) => [...contentKeys.all, id] as const,
};

export function useContentListQuery(filters: ContentFilters) {
  return useQuery({
    queryKey: contentKeys.list(filters),
    queryFn: ({ signal }) => contentService.list(filters, signal),
  });
}

export function useContentQuery(id: string) {
  return useQuery({
    queryKey: contentKeys.detail(id),
    queryFn: ({ signal }) => contentService.get(id, signal),
    enabled: Boolean(id),
  });
}
