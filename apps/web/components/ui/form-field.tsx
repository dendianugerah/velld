import { Input, type InputProps } from "@cloudflare/kumo/components/input";
import { Label } from "@/components/ui/label";
import { cn } from "@/lib/utils";
import type { ReactNode } from "react";

export function FormInput({ className, ...props }: InputProps) {
  return <Input {...props} className={cn("block w-full min-w-0", className)} />;
}

export function FormField({ id, label, hint, children }: { id: string; label: ReactNode; hint?: ReactNode; children: ReactNode }) {
  return <div className="grid min-w-0 content-start gap-2">
    <Label htmlFor={id} className="text-sm">{label}</Label>
    {children}
    {hint && <p id={`${id}-hint`} className="text-xs leading-relaxed text-muted-foreground">{hint}</p>}
  </div>;
}

export function FormFieldGroup({ children }: { children: ReactNode }) {
  return <div className="grid gap-5">{children}</div>;
}

export function FormFieldRow({ children }: { children: ReactNode }) {
  return <div className="grid min-w-0 gap-5 sm:grid-cols-2">{children}</div>;
}
