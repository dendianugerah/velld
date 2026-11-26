import { cn } from "@/lib/utils";

export function BackupStatus({ status }: { status: string }) {
  return <span className={cn("inline-flex w-fit items-center gap-1.5 rounded-md px-2 py-1 text-[10px] font-medium capitalize",
    status === "completed" ? "bg-[#edf5eb] text-[#3e7045]" :
    status === "failed" ? "bg-red-50 text-red-700" :
    status === "running" ? "bg-blue-50 text-blue-700" : "bg-amber-50 text-amber-800")}>
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
    status === "passed" ? "bg-[#edf5eb] text-[#3e7045]" :
    status === "failed" ? "bg-red-50 text-red-700" :
    status === "pending" ? "bg-blue-50 text-blue-700" :
    "bg-muted text-muted-foreground"
  )}>
    <span className="h-1.5 w-1.5 shrink-0 rounded-full bg-current" aria-hidden="true" />{restoreCheckLabels[status]}
  </span>;
}
