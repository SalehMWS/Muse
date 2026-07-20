import type { ComponentProps } from "react";

import { cn } from "@/lib/utils";

export function Label({ className, children, ...props }: ComponentProps<"label"> & { required?: boolean }) {
  const { required, ...rest } = props as ComponentProps<"label"> & { required?: boolean };

  return (
    <label className={cn("text-sm font-medium leading-none", className)} {...rest}>
      {children}
      {required ? (
        <span className="ml-0.5 text-danger" aria-hidden>
          *
        </span>
      ) : null}
    </label>
  );
}
