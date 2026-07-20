"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

import { ApiError } from "@/lib/api/errors";

import { schedulerKeys } from "../queries/use-schedules-query";
import { schedulerService } from "../services/scheduler.service";

interface CancelScheduleVariables {
  contentId: string;
  scheduleId: string;
}

export function useCancelSchedule() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ contentId, scheduleId }: CancelScheduleVariables) =>
      schedulerService.cancel(contentId, scheduleId),
    onSuccess: async (_data, variables) => {
      await queryClient.invalidateQueries({
        queryKey: schedulerKeys.schedules(variables.contentId),
      });
      toast.success("Schedule cancelled.");
    },
    onError: (error) => {
      toast.error(error instanceof ApiError ? error.message : "Could not cancel the schedule.");
    },
  });
}
