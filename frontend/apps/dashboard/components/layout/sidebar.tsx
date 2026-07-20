"use client";

import {
  BarChart3,
  CalendarClock,
  FileText,
  Camera,
  LayoutDashboard,
  Library,
  Settings,
} from "lucide-react";
import Link from "next/link";
import { usePathname } from "next/navigation";

import { cn } from "@/lib/utils";
import { useSidebarStore } from "@/stores/sidebar.store";

const navigation = [
  { href: "/dashboard", label: "Overview", icon: LayoutDashboard },
  { href: "/dashboard/content", label: "Content", icon: FileText },
  { href: "/dashboard/scheduler", label: "Scheduler", icon: CalendarClock },
  { href: "/dashboard/knowledge", label: "Knowledge", icon: Library },
  { href: "/dashboard/instagram", label: "Instagram", icon: Camera },
  { href: "/dashboard/analytics", label: "Analytics", icon: BarChart3 },
  { href: "/dashboard/settings", label: "Settings", icon: Settings },
];

export function Sidebar() {
  const pathname = usePathname();
  const collapsed = useSidebarStore((state) => state.collapsed);

  return (
    <aside
      className={cn(
        "hidden shrink-0 border-r border-border bg-card transition-[width] duration-250 md:block",
        collapsed ? "w-16" : "w-60",
      )}
    >
      <div className="flex h-14 items-center border-b border-border px-4">
        <Link href="/dashboard" className="truncate text-sm font-semibold tracking-tight">
          {collapsed ? "NF" : "NovaFlow"}
        </Link>
      </div>

      <nav className="flex flex-col gap-1 p-2" aria-label="Main">
        {navigation.map(({ href, label, icon: Icon }) => {
          const active = pathname === href || (href !== "/dashboard" && pathname.startsWith(href));

          return (
            <Link
              key={href}
              href={href}
              aria-current={active ? "page" : undefined}
              title={collapsed ? label : undefined}
              className={cn(
                "flex items-center gap-3 rounded-md px-3 py-2 text-sm transition-colors duration-150",
                active
                  ? "bg-accent text-accent-foreground font-medium"
                  : "text-muted-foreground hover:bg-accent hover:text-accent-foreground",
              )}
            >
              <Icon className="size-4 shrink-0" aria-hidden />
              {collapsed ? <span className="sr-only">{label}</span> : label}
            </Link>
          );
        })}
      </nav>
    </aside>
  );
}
