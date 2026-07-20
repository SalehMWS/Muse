"use client";

import { useId, type ComponentProps } from "react";

import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { cn } from "@/lib/utils";

interface TextFieldProps extends ComponentProps<"input"> {
  label: string;
  error?: string;
  hint?: string;
}

export function TextField({ label, error, hint, required, className, ...props }: TextFieldProps) {
  const id = useId();
  const errorId = `${id}-error`;
  const hintId = `${id}-hint`;

  return (
    <div className="flex flex-col gap-1.5">
      <Label htmlFor={id} required={required}>
        {label}
      </Label>
      <Input
        id={id}
        aria-invalid={Boolean(error) || undefined}
        aria-describedby={cn(error && errorId, hint && !error && hintId) || undefined}
        required={required}
        className={className}
        {...props}
      />
      {error ? (
        <p id={errorId} role="alert" className="text-xs text-danger">
          {error}
        </p>
      ) : hint ? (
        <p id={hintId} className="text-xs text-muted-foreground">
          {hint}
        </p>
      ) : null}
    </div>
  );
}
