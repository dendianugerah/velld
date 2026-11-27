import { cn } from "@/lib/utils";

export function BackupStatus({ status }: { status: string }) {
  return <span className={cn("inline-flex w-fit items-center gap-1.5 rounded-md px-2 py-1 text-[10px] font-medium capitalize",
    status === "completed" ? "bg-success/10 text-success" :
    status === "failed" ? "bg-destructive/10 text-destructive" :
    status === "running" ? "bg-info/10 text-info" : "bg-warning/10 text-warning")}>
    <span className="h-1.5 w-1.5 shrink-0 rounded-full bg-current" aria-hidden="true" />{status}
  </span>;
}

const restoreCheckLabels = {
  pending: "Restore check running",
  passed: "Restore verified",
  failed: "Restore check failed",
  skipped: "Restore not checked",
} as const;

export function RestoreCheckStatus({ status, message }: { status?: keyof typeof restoreCheckLabels | null; message?: string | null }) {
  if (!status || !restoreCheckLabels[status]) return null;
  return <span title={message || undefined} className={cn(
    "inline-flex w-fit items-center gap-1.5 rounded-md px-2 py-1 text-[10px] font-medium",
    status === "passed" ? "bg-success/10 text-success" :
    status === "failed" ? "bg-destructive/10 text-destructive" :
    status === "pending" ? "bg-info/10 text-info" :
    "bg-muted text-muted-foreground"
  )}>
    <span className="h-1.5 w-1.5 shrink-0 rounded-full bg-current" aria-hidden="true" />{restoreCheckLabels[status]}
  </span>;
}
