"use client";

import { Switch as KumoSwitch, type SwitchProps } from "@cloudflare/kumo/components/switch";
import { cn } from "@/lib/utils";

export function Switch({ className, ...props }: SwitchProps) {
  return <KumoSwitch {...props} className={cn("shrink-0 !rounded-full [corner-shape:round] supports-[corner-shape:squircle]:!rounded-full [&>div]:!rounded-full [&>div]:[corner-shape:round] [&>div]:supports-[corner-shape:squircle]:!rounded-full", className)} />;
}
