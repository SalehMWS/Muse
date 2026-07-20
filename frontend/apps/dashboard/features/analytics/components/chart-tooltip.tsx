"use client";

/**
 * Recharts' default tooltip paints a hardcoded white surface, which breaks in
 * dark mode. This one wears the popover tokens so it follows the theme, and
 * keeps all text on ink tokens rather than the series color.
 */
export function ChartTooltipCard({
  title,
  valueLabel,
  value,
  swatch,
}: {
  title: string;
  valueLabel: string;
  value: number;
  swatch?: string;
}) {
  return (
    <div className="rounded-md border border-border bg-popover px-3 py-2 text-popover-foreground shadow-md">
      <p className="text-xs font-medium">{title}</p>
      <p className="mt-1 flex items-center gap-1.5 text-xs text-muted-foreground">
        {swatch ? (
          <span
            className="inline-block size-2 shrink-0 rounded-full"
            style={{ backgroundColor: swatch }}
            aria-hidden
          />
        ) : null}
        <span className="font-semibold tabular-nums text-foreground">{value}</span>
        {valueLabel}
      </p>
    </div>
  );
}

/** Recharts hands back `string | number | (string | number)[]`; we only chart counts. */
export function toChartNumber(value: unknown): number {
  const parsed = typeof value === "number" ? value : Number(value);
  return Number.isFinite(parsed) ? parsed : 0;
}

/** Shared recessive axis/grid styling so both charts read as one system. */
export const AXIS_TICK = { fill: "var(--muted-foreground)", fontSize: 12 } as const;
export const AXIS_LINE = { stroke: "var(--border)" } as const;
