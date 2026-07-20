"use client";

import { LogOut, Moon, PanelLeft, Sun } from "lucide-react";
import { useTheme } from "next-themes";

import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { useCurrentUserQuery } from "@/features/auth/queries/use-current-user-query";
import { useLogout } from "@/features/auth/mutations/use-logout";
import { useSidebarStore } from "@/stores/sidebar.store";

export function Header() {
  const toggleSidebar = useSidebarStore((state) => state.toggle);
  const { data: user, isLoading } = useCurrentUserQuery();
  const logout = useLogout();
  const { setTheme, resolvedTheme } = useTheme();

  return (
    <header className="flex h-14 items-center justify-between gap-4 border-b border-border px-4">
      <Button variant="ghost" size="icon" onClick={toggleSidebar} aria-label="Toggle sidebar">
        <PanelLeft />
      </Button>

      <div className="flex items-center gap-2">
        <Button
          variant="ghost"
          size="icon"
          onClick={() => setTheme(resolvedTheme === "dark" ? "light" : "dark")}
          aria-label="Toggle theme"
        >
          <Sun className="hidden dark:block" />
          <Moon className="block dark:hidden" />
        </Button>

        {isLoading ? (
          <Skeleton className="h-5 w-32" />
        ) : user ? (
          <span className="hidden text-sm text-muted-foreground sm:inline">{user.email}</span>
        ) : null}

        <Button
          variant="ghost"
          size="icon"
          onClick={() => logout.mutate()}
          loading={logout.isPending}
          aria-label="Sign out"
        >
          <LogOut />
        </Button>
      </div>
    </header>
  );
}
