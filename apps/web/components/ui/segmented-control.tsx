import { cn } from "@/lib/utils";
import type { ComponentType, ReactNode } from "react";

interface SegmentedControlOption<T extends string> {
  value: T;
  label: string;
  icon?: ComponentType<{ className?: string; "aria-hidden"?: boolean }>;
}

interface SegmentedControlProps<T extends string> {
  label: string;
  value: T;
  options: SegmentedControlOption<T>[];
  onChange: (value: T) => void;
  className?: string;
  children?: ReactNode;
}

export function SegmentedControl<T extends string>({ label, value, options, onChange, className, children }: SegmentedControlProps<T>) {
  return <div role="group" aria-label={label} className={cn("inline-flex flex-wrap items-center gap-1 rounded-lg bg-muted p-1", className)}>
    {options.map(option => {
      const Icon = option.icon;
      const active = value === option.value;
      return <button key={option.value} type="button" aria-pressed={active} onClick={() => onChange(option.value)} className={cn("flex min-h-8 flex-1 items-center justify-center gap-2 whitespace-nowrap rounded-md px-3 text-[13px] font-medium transition-colors focus-visible:outline focus-visible:outline-2 focus-visible:outline-ring sm:flex-none", active ? "bg-card text-foreground shadow-sm" : "text-muted-foreground hover:text-foreground")}>
        {Icon && <Icon className="h-3.5 w-3.5" aria-hidden={true} />}
        {option.label}
      </button>;
    })}
    {children}
  </div>;
}
