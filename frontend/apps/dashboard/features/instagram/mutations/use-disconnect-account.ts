"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

import { ApiError } from "@/lib/api/errors";

import { instagramKeys } from "../queries/use-accounts-query";
import { instagramService } from "../services/instagram.service";

export function useDisconnectAccount() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => instagramService.disconnect(id),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: instagramKeys.accounts() });
      toast.success("Account disconnected.");
    },
    onError: (error) => {
      toast.error(error instanceof ApiError ? error.message : "Could not disconnect the account.");
    },
  });
}
