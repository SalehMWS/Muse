"use client";

import { useQuery } from "@tanstack/react-query";

import { analyticsService } from "../services/analytics.service";

export const analyticsKeys = {
  all: ["analytics"] as const,
  summary: () => [...analyticsKeys.all, "summary"] as const,
};

/**
 * Shared by every analytics component. React Query dedupes on the key, so the
 * three underlying endpoints are hit once per page load no matter how many
 * components read from this hook.
 */
export function useAnalyticsSummaryQuery() {
  return useQuery({
    queryKey: analyticsKeys.summary(),
    queryFn: ({ signal }) => analyticsService.summary(signal),
  });
}
