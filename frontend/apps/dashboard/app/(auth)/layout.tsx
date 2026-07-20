import Link from "next/link";
import type { ReactNode } from "react";

export default function AuthLayout({ children }: { children: ReactNode }) {
  return (
    <div className="flex min-h-dvh flex-col items-center justify-center gap-8 px-6 py-12">
      <Link href="/" className="text-lg font-semibold tracking-tight">
        NovaFlow
      </Link>
      <div className="w-full max-w-sm">{children}</div>
    </div>
  );
}
