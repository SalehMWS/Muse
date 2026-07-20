"use client";

import { useQuery } from "@tanstack/react-query";

import { schedulerService } from "../services/scheduler.service";

export const schedulerKeys = {
  all: ["scheduler"] as const,
  schedules: (contentId: string) => [...schedulerKeys.all, contentId, "schedules"] as const,
  contentOptions: () => [...schedulerKeys.all, "content-options"] as const,
  accountOptions: () => [...schedulerKeys.all, "account-options"] as const,
};

export function useSchedulesQuery(contentId: string) {
  return useQuery({
    queryKey: schedulerKeys.schedules(contentId),
    queryFn: ({ signal }) => schedulerService.list(contentId, signal),
    enabled: Boolean(contentId),
  });
}

export function useContentOptionsQuery() {
  return useQuery({
    queryKey: schedulerKeys.contentOptions(),
    queryFn: ({ signal }) => schedulerService.listContentOptions(signal),
  });
}

export function useAccountOptionsQuery() {
  return useQuery({
    queryKey: schedulerKeys.accountOptions(),
    queryFn: ({ signal }) => schedulerService.listAccountOptions(signal),
  });
}
