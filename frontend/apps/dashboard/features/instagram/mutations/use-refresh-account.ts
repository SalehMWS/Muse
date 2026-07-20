"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

import { ApiError } from "@/lib/api/errors";

import { instagramKeys } from "../queries/use-accounts-query";
import { instagramService } from "../services/instagram.service";

export function useRefreshAccount() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => instagramService.refresh(id),
    onSuccess: async (account) => {
      await queryClient.invalidateQueries({ queryKey: instagramKeys.accounts() });
      toast.success(`@${account.username} token refreshed.`);
    },
    onError: (error) => {
      toast.error(error instanceof ApiError ? error.message : "Could not refresh the account.");
    },
  });
}
