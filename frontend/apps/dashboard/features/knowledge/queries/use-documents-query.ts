"use client";

import { useQuery } from "@tanstack/react-query";

import { knowledgeService } from "../services/knowledge.service";

export const knowledgeKeys = {
  all: ["knowledge"] as const,
  lists: () => [...knowledgeKeys.all, "documents"] as const,
  list: () => [...knowledgeKeys.lists()] as const,
  detail: (id: string) => [...knowledgeKeys.all, id] as const,
};

export function useDocumentsQuery() {
  return useQuery({
    queryKey: knowledgeKeys.list(),
    queryFn: ({ signal }) => knowledgeService.list(signal),
  });
}
