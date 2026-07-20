"use client";

import { Camera } from "lucide-react";

import { EmptyState } from "@/components/shared/empty-state";
import { ErrorCard } from "@/components/shared/error-card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { ApiError } from "@/lib/api/errors";

import { useDisconnectAccount } from "../mutations/use-disconnect-account";
import { useRefreshAccount } from "../mutations/use-refresh-account";
import { useAccountsQuery } from "../queries/use-accounts-query";
import { EXPIRY_WARNING_DAYS, type Account, type AccountStatus } from "../types/account";

import { ConnectInstagramButton } from "./connect-instagram-button";

const statusVariant: Record<AccountStatus, "success" | "warning" | "danger"> = {
  active: "success",
  expired: "danger",
  revoked: "danger",
};

function formatDate(iso: string): string {
  const parsed = Date.parse(iso);
  if (Number.isNaN(parsed)) return "unknown";
  return new Date(parsed).toLocaleDateString(undefined, {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}

function expiryLabel(account: Account): string {
  if (account.expiresInDays < 0) return `Token expired on ${formatDate(account.tokenExpiresAt)}`;
  if (account.expiresInDays === 0) return "Token expires today";
  if (account.expiresInDays === 1) return "Token expires tomorrow";
  return `Token expires in ${account.expiresInDays} days`;
}

export function AccountList() {
  const query = useAccountsQuery();
  const refresh = useRefreshAccount();
  const disconnect = useDisconnectAccount();

  if (query.isPending) {
    return (
      <div className="flex flex-col gap-3">
        {Array.from({ length: 2 }).map((_, index) => (
          <Skeleton key={index} className="h-28 w-full" />
        ))}
      </div>
    );
  }

  if (query.isError) {
    return (
      <ErrorCard
        message={
          query.error instanceof ApiError
            ? query.error.message
            : "Could not load your connected accounts."
        }
        action={
          <Button variant="outline" size="sm" onClick={() => query.refetch()}>
            Try again
          </Button>
        }
      />
    );
  }

  const accounts = query.data;

  if (accounts.length === 0) {
    return (
      <EmptyState
        icon={Camera}
        title="No Instagram account connected"
        description="Connect a business or creator account to publish and schedule posts from NovaFlow."
        action={<ConnectInstagramButton />}
      />
    );
  }

  return (
    <ul className="flex flex-col gap-3">
      {accounts.map((account) => {
        const expiringSoon = account.expiresInDays <= EXPIRY_WARNING_DAYS;

        return (
          <li key={account.id}>
            <Card>
              <CardHeader className="flex-row items-start justify-between gap-4 pb-3">
                <div className="flex min-w-0 flex-col gap-1">
                  <CardTitle className="truncate">@{account.username}</CardTitle>
                  <div className="flex flex-wrap items-center gap-2">
                    <Badge variant={statusVariant[account.status] ?? "default"}>
                      {account.status}
                    </Badge>
                    {account.accountType ? (
                      <Badge variant="outline">{account.accountType}</Badge>
                    ) : null}
                    {expiringSoon ? (
                      <Badge variant={account.expiresInDays < 0 ? "danger" : "warning"}>
                        {account.expiresInDays < 0 ? "Token expired" : "Expires soon"}
                      </Badge>
                    ) : null}
                  </div>
                </div>

                <div className="flex shrink-0 gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    loading={refresh.isPending && refresh.variables === account.id}
                    onClick={() => refresh.mutate(account.id)}
                  >
                    Refresh
                  </Button>
                  <Button
                    variant="ghost"
                    size="sm"
                    loading={disconnect.isPending && disconnect.variables === account.id}
                    onClick={() => disconnect.mutate(account.id)}
                  >
                    Disconnect
                  </Button>
                </div>
              </CardHeader>

              <CardContent className="pt-0">
                <p
                  className={
                    expiringSoon ? "text-sm text-danger" : "text-sm text-muted-foreground"
                  }
                >
                  {expiryLabel(account)}
                </p>
                <p className="text-xs text-muted-foreground">
                  Connected {formatDate(account.connectedAt)}
                </p>
              </CardContent>
            </Card>
          </li>
        );
      })}
    </ul>
  );
}
