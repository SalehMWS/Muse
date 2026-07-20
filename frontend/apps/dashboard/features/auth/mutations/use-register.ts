"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";

import type { RegisterInput } from "../schemas/auth.schema";
import { authService } from "../services/auth.service";
import { authKeys } from "../queries/use-current-user-query";

export function useRegister() {
  const router = useRouter();
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (input: RegisterInput) => authService.register(input),
    onSuccess: async (user) => {
      queryClient.setQueryData(authKeys.currentUser(), user);
      await queryClient.invalidateQueries({ queryKey: authKeys.all });
      router.replace("/dashboard");
      router.refresh();
    },
  });
}
