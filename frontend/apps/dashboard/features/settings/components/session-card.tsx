"use client";

import { LogOut } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { ApiError } from "@/lib/api/errors";

import { useSignOut } from "../mutations/use-sign-out";

export function SessionCard() {
  const signOut = useSignOut();

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">Session</CardTitle>
        <CardDescription>Sign out of NovaFlow on this device.</CardDescription>
      </CardHeader>
      <CardContent className="flex flex-col gap-3">
        <Button
          variant="danger"
          className="w-fit"
          loading={signOut.isPending}
          onClick={() => signOut.mutate()}
        >
          <LogOut aria-hidden />
          Sign out
        </Button>

        {signOut.isError ? (
          <p role="alert" className="text-sm text-danger">
            {signOut.error instanceof ApiError
              ? signOut.error.message
              : "Could not sign you out. Please try again."}
          </p>
        ) : null}

        <p className="text-xs text-muted-foreground">
          Signing out clears your session cookie and every cached response in this tab. There is
          no API for deleting your account yet, so that option is not offered here.
        </p>
      </CardContent>
    </Card>
  );
}
