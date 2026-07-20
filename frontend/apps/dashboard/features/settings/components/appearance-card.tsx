"use client";

import { useTheme } from "next-themes";
import { useSyncExternalStore } from "react";

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Select } from "@/components/ui/select";

const THEME_OPTIONS = [
  { value: "light", label: "Light" },
  { value: "dark", label: "Dark" },
  { value: "system", label: "System" },
] as const;

/** Never fires: the value only needs to differ between server and client. */
const subscribeToNothing = () => () => {};

export function AppearanceCard() {
  const { theme, setTheme, resolvedTheme } = useTheme();

  // The active theme is only known in the browser, so the first client render
  // must match the server's. This reads false while rendering on the server and
  // true once hydrated, holding the control disabled until then.
  const mounted = useSyncExternalStore(
    subscribeToNothing,
    () => true,
    () => false,
  );

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">Appearance</CardTitle>
        <CardDescription>Choose how NovaFlow looks on this device.</CardDescription>
      </CardHeader>
      <CardContent className="flex flex-col gap-2">
        <Label htmlFor="theme-select">Theme</Label>
        <Select
          id="theme-select"
          className="max-w-xs"
          value={mounted ? (theme ?? "system") : "system"}
          disabled={!mounted}
          onChange={(event) => setTheme(event.target.value)}
        >
          {THEME_OPTIONS.map((option) => (
            <option key={option.value} value={option.value}>
              {option.label}
            </option>
          ))}
        </Select>
        <p className="text-xs text-muted-foreground">
          {mounted && theme === "system"
            ? `Following your device setting, currently ${resolvedTheme ?? "unknown"}.`
            : "This preference is stored in your browser and applies to this device only."}
        </p>
      </CardContent>
    </Card>
  );
}
