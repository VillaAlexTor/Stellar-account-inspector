import * as React from "react";
import { cn } from "@/lib/utils";

export function Input({ className, type, ...props }: React.ComponentProps<"input">) {
  return (
    <input
      type={type}
      className={cn(
        "flex min-h-12 w-full border border-ink/45 bg-display px-4 py-3 font-mono text-sm text-display-foreground caret-amber outline-none transition-[border-color,box-shadow] placeholder:text-display-muted focus:border-amber focus:shadow-[0_0_0_3px_rgba(228,153,35,.22)] disabled:cursor-not-allowed disabled:opacity-50",
        className,
      )}
      {...props}
    />
  );
}
