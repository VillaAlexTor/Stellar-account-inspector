import * as React from "react";
import { Slot } from "@radix-ui/react-slot";
import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "@/lib/utils";

const buttonVariants = cva(
  "inline-flex items-center justify-center gap-2 whitespace-nowrap font-label text-sm font-semibold uppercase tracking-[0.12em] transition-[background-color,color,box-shadow,transform] duration-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-amber focus-visible:ring-offset-2 focus-visible:ring-offset-panel disabled:pointer-events-none disabled:opacity-50 active:translate-y-px",
  {
    variants: {
      variant: {
        default:
          "border border-ink bg-ink px-5 py-3 text-panel shadow-[inset_0_1px_0_rgba(255,255,255,.18),0_8px_18px_rgba(23,24,22,.16)] hover:bg-amber hover:text-ink",
        outline:
          "border border-ink/50 bg-panel/50 px-4 py-2.5 text-ink shadow-[inset_0_1px_0_rgba(255,255,255,.8)] hover:border-ink hover:bg-panel-strong",
        ghost: "px-3 py-2 text-ink/70 hover:bg-ink/8 hover:text-ink",
      },
      size: {
        default: "min-h-11",
        sm: "min-h-9 text-xs",
        icon: "size-11 p-0",
      },
    },
    defaultVariants: { variant: "default", size: "default" },
  },
);

export interface ButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement>,
    VariantProps<typeof buttonVariants> {
  asChild?: boolean;
}

export function Button({ className, variant, size, asChild, ...props }: ButtonProps) {
  const Comp = asChild ? Slot : "button";
  return <Comp className={cn(buttonVariants({ variant, size }), className)} {...props} />;
}
