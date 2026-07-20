"use client";

import { useQuery } from "@tanstack/react-query";

import { ApiError } from "@/lib/api/errors";

import { settingsService } from "../services/settings.service";

export const settingsKeys = {
  all: ["settings"] as const,
  profile: () => [...settingsKeys.all, "profile"] as const,
};

export function useProfileQuery() {
  return useQuery({
    queryKey: settingsKeys.profile(),
    queryFn: ({ signal }) => settingsService.profile(signal),
    // Retrying an expired session just burns requests — it will not start working.
    retry: (failureCount, error) =>
      error instanceof ApiError && error.kind === "authentication" ? false : failureCount < 2,
  });
}
