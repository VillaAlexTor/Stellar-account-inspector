import * as React from "react";
import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "@/lib/utils";

const badgeVariants = cva(
  "inline-flex min-h-6 items-center gap-1.5 border px-2 py-0.5 font-label text-[0.7rem] font-semibold uppercase tracking-[0.12em]",
  {
    variants: {
      variant: {
        default: "border-ink/35 bg-panel-strong text-ink",
        active: "border-amber-dark bg-amber text-ink",
        safe: "border-safe/45 bg-safe/12 text-safe-dark",
        warning: "border-warning/50 bg-warning/14 text-warning-dark",
        bajo: "border-safe/45 bg-safe/12 text-safe-dark",
        medio: "border-amber-dark/55 bg-amber/18 text-amber-dark",
        alto: "border-warning/55 bg-warning/14 text-warning-dark",
        crítico: "border-warning-dark bg-warning-dark text-panel-strong",
        low: "border-safe/45 bg-safe/12 text-safe-dark",
        medium: "border-amber-dark/55 bg-amber/18 text-amber-dark",
        high: "border-warning/55 bg-warning/14 text-warning-dark",
        critical: "border-warning-dark bg-warning-dark text-panel-strong",
        muted: "border-ink/20 bg-transparent text-ink/60",
      },
    },
    defaultVariants: { variant: "default" },
  },
);

export interface BadgeProps
  extends React.HTMLAttributes<HTMLSpanElement>,
    VariantProps<typeof badgeVariants> {}

export function Badge({ className, variant, ...props }: BadgeProps) {
  return <span className={cn(badgeVariants({ variant }), className)} {...props} />;
}
