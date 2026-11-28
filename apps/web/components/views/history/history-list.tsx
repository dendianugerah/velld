"use client";

import { useEffect } from "react";
import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { format, formatDistanceToNow, parseISO } from "date-fns";
import { ChevronRight, Download, X } from "lucide-react";
import { LayerCard } from "@cloudflare/kumo/components/layer-card";
import { useBackup } from "@/hooks/use-backup";
import { useConnections } from "@/hooks/use-connections";
import { formatSize } from "@/lib/helper";
import { Button, buttonVariants } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { BackupStatus, RestoreCheckStatus } from "@/components/ui/backup-status";
import { CustomPagination } from "@/components/ui/custom-pagination";
import { HistoryFilters } from "./history-filters";

export function HistoryList() {
  const query = useSearchParams();
  const search = query.get("search") || "";
  const date = query.get("date") || "all";
  const status = query.get("status") || "all";
  const engine = query.get("engine") || "all";
  const verification = query.get("verification") || "all";
  const connectionId = query.get("connection_id") || "";
  const page = Math.max(1, Number(query.get("page")) || 1);
  const { backups, pagination, isLoading, error, refetch, isFetching, isPlaceholderData, downloadBackupFile, isDownloading } = useBackup({page, search, date, status, engine, verification, connection_id: connectionId});
  const { connections } = useConnections();
  const connection = connections?.find(item => item.id === connectionId);
  const active = !!(search || date !== "all" || status !== "all" || engine !== "all" || verification !== "all" || connectionId);

  function change(key: string, value: string) {
    const next = new URLSearchParams(query.toString());
    if (value && value !== "all") next.set(key, value); else next.delete(key);
    if (key !== "page") next.delete("page");
    window.history[key === "search" ? "replaceState" : "pushState"](null, "", `/history${next.size ? `?${next}` : ""}`);
  }
  useEffect(() => {
    if (!isFetching && !error && pagination && page > Math.max(1, pagination.total_pages)) {
      const next = new URLSearchParams(query.toString());
      next.delete("page");
      window.history.replaceState(null, "", `/history${next.size ? `?${next}` : ""}`);
    }
  }, [page, pagination, isFetching, error, query]);

  return <div className="space-y-5">
    <HistoryFilters search={search} onSearchChange={value => change("search", value)} dateRange={date} onDateRangeChange={value => change("date", value)}
      status={status} onStatusChange={value => change("status", value)} databaseType={engine} onDatabaseTypeChange={value => change("engine", value)}
      verification={verification} onVerificationChange={value => change("verification", value)}
      active={active} onReset={() => window.history.pushState(null, "", "/history")} />
    {connectionId && <Button variant="outline" size="sm" aria-label={`Remove connection filter: ${connection?.name || "Selected connection"}`} onClick={() => change("connection_id", "")}>
      <span className="max-w-52 truncate">{connection?.name || "Selected connection"}</span><X aria-hidden="true" />
    </Button>}
    <LayerCard className="overflow-hidden rounded-lg shadow-none">
      <div className="hidden grid-cols-[minmax(0,1fr)_94px_130px_150px] gap-4 border-b bg-muted/50 px-4 py-3 text-xs text-muted-foreground xl:grid">
        <span>Database</span><span>Backup status</span><span>Restore check</span><span>Size / storage</span>
      </div>
      <div aria-busy={isFetching} className={isPlaceholderData ? "opacity-50" : ""}>
        {isLoading ? <div className="space-y-5 p-5"><Skeleton className="h-14 w-full" /><Skeleton className="h-14 w-full" /><Skeleton className="h-14 w-full" /></div> : error ? <div role="alert" className="space-y-3 p-6"><p className="text-sm">Couldn’t load backup history.</p><Button variant="outline" size="sm" onClick={() => refetch()}>Try again</Button></div> : backups?.length ? backups.map(backup => {
          const created = parseISO(backup.created_at);
          const filename = backup.path.split(/[\\/]/).pop() || "Backup file";
          const href = `/history/${backup.id}?${new URLSearchParams({from: query.toString()})}`;
          return <div key={backup.id} className="grid grid-cols-[minmax(0,1fr)_auto] items-start gap-x-4 gap-y-3 border-b px-4 py-4 last:border-b-0 xl:grid-cols-[minmax(0,1fr)_94px_130px_150px] xl:items-center">
            <div className="min-w-0"><Link href={href} className="group inline-flex max-w-full items-center gap-1.5 rounded-sm text-sm font-semibold hover:text-primary focus-visible:outline focus-visible:outline-2 focus-visible:outline-ring"><span className="truncate">{backup.database_name || backup.connection_name || "Database backup"}</span><ChevronRight aria-hidden="true" className="h-3.5 w-3.5 shrink-0 text-muted-foreground group-hover:text-primary" /></Link>
              <p className="mt-1 break-words text-xs text-muted-foreground">{backup.connection_name || backup.database_type} · <time dateTime={backup.created_at} title={format(created, "PPpp")}>{formatDistanceToNow(created, {addSuffix: true})}</time></p>
            </div>
            <div className="flex flex-col items-end gap-2 xl:items-start"><span className="sr-only">Backup status: </span><BackupStatus status={backup.status} /><div className="xl:hidden"><RestoreCheckStatus status={backup.verification_status || "skipped"} /></div></div>
            <div className="hidden xl:block"><RestoreCheckStatus status={backup.verification_status || "skipped"} /></div>
            <div className="col-span-2 flex min-h-9 items-center justify-between gap-2 border-t pt-2 xl:col-span-1 xl:border-0 xl:pt-0">
              <p className="text-xs tabular-nums text-muted-foreground">{formatSize(backup.size)}<span className="ms-2 xl:ms-0 xl:mt-1 xl:block">{backup.s3_object_key ? "Object storage" : "Local"}</span></p>
              {backup.status === "completed" && <Button variant="ghost" size="icon" className="size-9 shrink-0" aria-label={`Download ${filename}`} title="Download backup" disabled={isDownloading} onClick={() => downloadBackupFile(backup)}><Download aria-hidden="true" /></Button>}
            </div>
          </div>;
        }) : <div className="space-y-3 px-6 py-12 text-center"><h2 className="text-sm font-medium">{active ? "No matching backups" : "No backups yet"}</h2>{!active && <Link href="/connections" className={buttonVariants({variant: "outline", size: "sm"})}>Go to Connections</Link>}</div>}
      </div>
    </LayerCard>
    <div className="flex flex-wrap items-center justify-between gap-4"><p role="status" className="text-xs text-muted-foreground">{isFetching ? "Updating…" : pagination && pagination.total > 0 ? `${pagination.total} ${active ? "matching " : ""}backup${pagination.total === 1 ? "" : "s"}` : ""}</p>
      {pagination && pagination.total_pages > 1 && <CustomPagination currentPage={page} totalPages={pagination.total_pages} onPageChange={value => change("page", String(value))} />}
    </div>
  </div>;
}
