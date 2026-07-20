"use client";

import { ErrorCard } from "@/components/shared/error-card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { ApiError } from "@/lib/api/errors";

import { useProfileQuery } from "../queries/use-profile-query";

export function ProfileCard() {
  const query = useProfileQuery();

  if (query.isPending) {
    return (
      <Shell>
        <div className="flex flex-col gap-4">
          {Array.from({ length: 4 }).map((_, index) => (
            <Skeleton key={index} className="h-10 w-full" />
          ))}
        </div>
      </Shell>
    );
  }

  if (query.isError) {
    return (
      <Shell>
        <ErrorCard
          message={
            query.error instanceof ApiError ? query.error.message : "Could not load your profile."
          }
          action={
            <Button variant="outline" size="sm" onClick={() => query.refetch()}>
              Try again
            </Button>
          }
        />
      </Shell>
    );
  }

  const profile = query.data;

  return (
    <Shell>
      <dl className="flex flex-col divide-y divide-border">
        <Row label="Display name" value={profile.displayName || "Not set"} />
        <Row label="Email" value={profile.email} />
        <Row
          label="Email verified"
          value={
            <Badge variant={profile.emailVerified ? "success" : "warning"}>
              {profile.emailVerified ? "Verified" : "Not verified"}
            </Badge>
          }
        />
        <Row label="Account status" value={<Badge variant="outline">{profile.status}</Badge>} />
        <Row label="Member since" value={formatDate(profile.createdAt)} />
        <Row
          label="User ID"
          value={<span className="font-mono text-xs break-all">{profile.id}</span>}
        />
      </dl>

      <p className="mt-4 text-xs text-muted-foreground">
        These details are read-only. NovaFlow does not yet expose an API for changing your
        profile, email or password, so there is nothing to edit here for now.
      </p>
    </Shell>
  );
}

function Row({ label, value }: { label: string; value: React.ReactNode }) {
  return (
    <div className="flex flex-wrap items-center justify-between gap-2 py-3 first:pt-0">
      <dt className="text-sm text-muted-foreground">{label}</dt>
      <dd className="text-sm font-medium">{value}</dd>
    </div>
  );
}

function formatDate(iso: string): string {
  const parsed = new Date(iso);
  if (Number.isNaN(parsed.getTime())) {
    return iso;
  }
  return parsed.toLocaleDateString(undefined, { year: "numeric", month: "long", day: "numeric" });
}

function Shell({ children }: { children: React.ReactNode }) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">Profile</CardTitle>
        <CardDescription>Your account details, as the API reports them.</CardDescription>
      </CardHeader>
      <CardContent>{children}</CardContent>
    </Card>
  );
}
