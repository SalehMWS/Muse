"use client";

import { useQuery } from "@tanstack/react-query";

import { publishingService } from "../services/publishing.service";

export const publishingKeys = {
  all: ["publishing"] as const,
  publications: (contentId: string) => [...publishingKeys.all, "publications", contentId] as const,
  accounts: () => [...publishingKeys.all, "accounts"] as const,
};

export function usePublicationsQuery(contentId: string) {
  return useQuery({
    queryKey: publishingKeys.publications(contentId),
    queryFn: ({ signal }) => publishingService.listPublications(contentId, signal),
    enabled: Boolean(contentId),
  });
}

export function useInstagramAccountsQuery() {
  return useQuery({
    queryKey: publishingKeys.accounts(),
    queryFn: ({ signal }) => publishingService.listAccounts(signal),
  });
}
