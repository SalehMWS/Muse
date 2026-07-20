"use client";

import { useEffect } from "react";

import { ErrorCard } from "@/components/shared/error-card";
import { Button } from "@/components/ui/button";

export default function ContentDetailError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    if (process.env.NODE_ENV === "development") {
      console.error(error);
    }
  }, [error]);

  return (
    <ErrorCard
      message="This content could not be loaded. It may have been archived or removed."
      action={
        <Button variant="outline" size="sm" onClick={reset}>
          Try again
        </Button>
      }
    />
  );
}
