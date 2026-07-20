"use client";

import { useQuery } from "@tanstack/react-query";

import { ApiError } from "@/lib/api/errors";

import { authService } from "../services/auth.service";

export const authKeys = {
  all: ["auth"] as const,
  currentUser: () => [...authKeys.all, "me"] as const,
};

export function useCurrentUserQuery() {
  return useQuery({
    queryKey: authKeys.currentUser(),
    queryFn: ({ signal }) => authService.currentUser(signal),
    retry: (failureCount, error) =>
      error instanceof ApiError && error.kind === "authentication" ? false : failureCount < 2,
  });
}
