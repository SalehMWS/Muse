import { AlertTriangle } from "lucide-react";
import type { ReactNode } from "react";

export function ErrorCard({ message, action }: { message: string; action?: ReactNode }) {
  return (
    <div
      role="alert"
      className="flex flex-col items-start gap-3 rounded-lg border border-danger/40 bg-danger/5 p-6"
    >
      <div className="flex items-center gap-2 text-danger">
        <AlertTriangle className="size-5" aria-hidden />
        <h2 className="text-sm font-semibold">Something went wrong</h2>
      </div>
      <p className="text-sm text-muted-foreground">{message}</p>
      {action}
    </div>
  );
}
