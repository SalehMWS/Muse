"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

import { ApiError } from "@/lib/api/errors";

import { schedulerKeys } from "../queries/use-schedules-query";
import type { CreateSchedulePayload } from "../schemas/create-schedule.schema";
import { schedulerService } from "../services/scheduler.service";

export function useCreateSchedule(onCreated?: () => void) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (input: CreateSchedulePayload) => schedulerService.create(input.contentId, input),
    onSuccess: async (schedule) => {
      await queryClient.invalidateQueries({
        queryKey: schedulerKeys.schedules(schedule.contentId),
      });
      toast.success("Post scheduled.");
      onCreated?.();
    },
    onError: (error) => {
      toast.error(error instanceof ApiError ? error.message : "Could not schedule the post.");
    },
  });
}
