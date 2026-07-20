"use client";

import { useMutation } from "@tanstack/react-query";
import { toast } from "sonner";

import { ApiError } from "@/lib/api/errors";

import { instagramService } from "../services/instagram.service";

/**
 * Starts the Meta OAuth round trip.
 *
 * The backend needs Meta app credentials to build an authorization URL. When
 * they are unset it answers with a real error — we surface that message in a
 * toast instead of silently doing nothing, so the button never looks broken.
 */
export function useConnectInstagram() {
  return useMutation({
    mutationFn: () => instagramService.connect(),
    onSuccess: (connect) => {
      if (!connect.authorizationUrl) {
        toast.error("Instagram is not configured yet. Add your Meta app credentials to connect.");
        return;
      }
      window.location.href = connect.authorizationUrl;
    },
    onError: (error) => {
      toast.error(
        error instanceof ApiError
          ? error.message
          : "Could not start the Instagram connection. Please try again.",
      );
    },
  });
}
