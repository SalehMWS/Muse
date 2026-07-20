"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";

import { settingsService } from "../services/settings.service";

export function useSignOut() {
  const router = useRouter();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: () => settingsService.signOut(),
    onSuccess: () => {
      // Drop every cached response so the next session cannot read this one's data.
      queryClient.clear();
      router.replace("/login");
      router.refresh();
    },
  });
}
