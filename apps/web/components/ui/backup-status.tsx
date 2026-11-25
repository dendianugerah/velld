import { cn } from "@/lib/utils";

export function BackupStatus({ status }: { status: string }) {
  return <span className={cn("inline-flex w-fit items-center gap-1.5 rounded-md px-2 py-1 text-[10px] font-medium capitalize",
    status === "completed" ? "bg-[#edf5eb] text-[#3e7045]" :
    status === "failed" ? "bg-red-50 text-red-700" :
    status === "running" ? "bg-blue-50 text-blue-700" : "bg-amber-50 text-amber-800")}>
    <span className="h-1.5 w-1.5 shrink-0 rounded-full bg-current" aria-hidden="true" />{status}
  </span>;
}
