"use client";

import { useQuery } from "@tanstack/react-query";

import { instagramService } from "../services/instagram.service";

export const instagramKeys = {
  all: ["instagram"] as const,
  accounts: () => [...instagramKeys.all, "accounts"] as const,
  account: (id: string) => [...instagramKeys.accounts(), id] as const,
};

export function useAccountsQuery() {
  return useQuery({
    queryKey: instagramKeys.accounts(),
    queryFn: ({ signal }) => instagramService.list(signal),
  });
}
