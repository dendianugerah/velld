"use client";

import { useEffect } from "react";
import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { format, formatDistanceToNow, parseISO } from "date-fns";
import { ArrowUpRight, Download } from "lucide-react";
import { LayerCard } from "@cloudflare/kumo/components/layer-card";
import { Select } from "@cloudflare/kumo/components/select";
import { useBackup } from "@/hooks/use-backup";
import { useConnections } from "@/hooks/use-connections";
import { formatSize } from "@/lib/helper";
import { Button } from "@/components/ui/button";
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
      onReset={() => window.history.pushState(null, "", "/history")} />
    <div className="flex flex-wrap items-center justify-between gap-3">
      <div className="flex items-center gap-2 text-xs text-muted-foreground"><span>Restore check</span>
        <Select aria-label="Filter by restore check" value={verification} onValueChange={value => { if (value) change("verification", value); }} items={{ all: "All results", passed: "Verified", failed: "Failed", pending: "Running", skipped: "Not checked" }} className="w-32">
          <Select.Option value="all">All results</Select.Option><Select.Option value="passed">Verified</Select.Option><Select.Option value="failed">Failed</Select.Option><Select.Option value="pending">Running</Select.Option><Select.Option value="skipped">Not checked</Select.Option>
        </Select>
      </div>
      {connectionId && <div className="flex items-center gap-2 rounded-md bg-muted px-2.5 py-1.5 text-xs"><span className="max-w-52 truncate">{connection?.name || "Selected connection"}</span><button aria-label="Remove connection filter" className="rounded px-1 hover:bg-background" onClick={() => change("connection_id", "")}>×</button></div>}
    </div>
    <LayerCard className="overflow-hidden rounded-lg shadow-none">
      <div className="hidden grid-cols-[minmax(0,1fr)_94px_130px_90px] gap-4 border-b bg-muted/50 px-4 py-3 text-[11px] text-muted-foreground xl:grid">
        <span>Database / created</span><span>Backup</span><span>Restore check</span><span className="text-end">Size / storage</span>
      </div>
      <div aria-busy={isFetching} className={isPlaceholderData ? "opacity-50" : ""}>
        {isLoading ? <div className="space-y-5 p-5"><Skeleton className="h-14 w-full" /><Skeleton className="h-14 w-full" /><Skeleton className="h-14 w-full" /></div> : error ? <div role="alert" className="space-y-3 p-6"><p className="text-sm">Couldn’t load backup history.</p><Button variant="outline" size="sm" onClick={() => refetch()}>Try again</Button></div> : backups?.length ? backups.map(backup => {
          const created = parseISO(backup.created_at);
          const filename = backup.path.split(/[\\/]/).pop() || "Backup file";
          const href = `/history/${backup.id}?${new URLSearchParams({from: query.toString()})}`;
          return <div key={backup.id} className="grid grid-cols-[minmax(0,1fr)_auto] items-start gap-x-4 gap-y-3 border-b px-4 py-4 last:border-b-0 xl:grid-cols-[minmax(0,1fr)_94px_130px_90px] xl:items-center">
            <div className="min-w-0"><Link href={href} className="group inline-flex max-w-full items-center gap-1.5 text-[13px] font-medium hover:text-primary"><span className="truncate">{backup.database_name || backup.connection_name || "Database backup"}</span><ArrowUpRight className="h-3.5 w-3.5 shrink-0 text-muted-foreground group-hover:text-primary" /></Link>
              <p className="mt-1 break-words text-[11px] text-muted-foreground">{backup.connection_name || backup.database_type} · <time dateTime={backup.created_at} title={format(created, "PPpp")}>{formatDistanceToNow(created, {addSuffix: true})}</time></p>
              <p className="mt-1.5 truncate font-mono text-[10px] text-muted-foreground" title={filename}>{filename}</p>
            </div>
            <div className="flex flex-col items-end gap-2 xl:items-start"><span className="sr-only">Backup status: </span><BackupStatus status={backup.status} /><div className="xl:hidden"><RestoreCheckStatus status={backup.verification_status || "skipped"} /></div></div>
            <div className="hidden xl:block"><RestoreCheckStatus status={backup.verification_status || "skipped"} /></div>
            <div className="col-span-2 flex items-center justify-between border-t border-dashed pt-2.5 xl:col-span-1 xl:block xl:border-0 xl:pt-0 xl:text-end"><p className="text-[11px] tabular-nums text-muted-foreground">{formatSize(backup.size)}<span className="ms-2 xl:ms-0 xl:mt-1 xl:block">{backup.s3_object_key ? "S3" : "Local"}</span></p><div className="flex items-center gap-2 xl:hidden"><Link href={href} className="text-xs font-medium">View backup →</Link>{backup.status === "completed" && <Button variant="ghost" size="icon" className="h-7 w-7" aria-label={`Download ${filename}`} disabled={isDownloading} onClick={() => downloadBackupFile(backup)}><Download /></Button>}</div></div>
          </div>;
        }) : <div className="px-6 py-12 text-center"><h2 className="text-sm font-medium">{active ? "No matching backups" : "No backups yet"}</h2><p className="mt-2 text-xs text-muted-foreground">{active ? "Try a different search or clear your filters." : "Run a backup from Connections to get started."}</p>{active && <Button variant="outline" size="sm" className="mt-4" onClick={() => window.history.pushState(null, "", "/history")}>Clear filters</Button>}</div>}
      </div>
    </LayerCard>
    <div className="flex flex-wrap items-center justify-between gap-4"><p role="status" className="text-[11px] text-muted-foreground">{isFetching ? "Updating…" : pagination ? `${pagination.total} ${active ? "matching " : ""}backup${pagination.total === 1 ? "" : "s"} · Page ${page} of ${Math.max(1, pagination.total_pages)}` : ""}</p>
      {pagination && pagination.total_pages > 1 && <CustomPagination currentPage={page} totalPages={pagination.total_pages} onPageChange={value => change("page", String(value))} />}
    </div>
  </div>;
}
