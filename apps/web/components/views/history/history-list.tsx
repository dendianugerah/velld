"use client";

import { useMemo, useState } from "react";
import { formatDistanceToNow, isAfter, parseISO, subDays } from "date-fns";
import { MoreHorizontal } from "lucide-react";
import { useBackup } from "@/hooks/use-backup";
import { formatSize } from "@/lib/helper";
import { typeLabels, DatabaseType } from "@/types/base";
import type { BackupList } from "@/types/backup";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { BackupStatus, RestoreCheckStatus } from "@/components/ui/backup-status";
import { CustomPagination } from "@/components/ui/custom-pagination";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuSeparator, DropdownMenuTrigger } from "@/components/ui/dropdown-menu";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { RestoreDialog } from "./restore-dialog";
import { BranchDialog } from "./branch-dialog";
import { SchemaDiffDialog } from "./schema-diff-dialog";
import { HistoryFilters } from "./history-filters";

export function HistoryList() {
  const { backups, isLoading, error, pagination, page, setPage, downloadBackupFile, isDownloading, search, setSearch, deleteBackupById, isDeleting } = useBackup();
  const [restore, setRestore] = useState<BackupList | null>(null);
  const [branch, setBranch] = useState<BackupList | null>(null);
  const [schemaDiff, setSchemaDiff] = useState<BackupList | null>(null);
  const [deleting, setDeleting] = useState<BackupList | null>(null);
  const [actionTrigger, setActionTrigger] = useState<HTMLElement | null>(null);
  const [dateRange, setDateRange] = useState("all");
  const [status, setStatus] = useState("all");
  const [databaseType, setDatabaseType] = useState("all");
  const filtered = useMemo(() => (backups ?? []).filter(backup => {
    const days = { "24hours": 1, "7days": 7, "30days": 30 }[dateRange];
    return (!days || isAfter(parseISO(backup.created_at), subDays(new Date(), days)))
      && (status === "all" || backup.status === status)
      && (databaseType === "all" || backup.database_type === databaseType);
  }), [backups, dateRange, status, databaseType]);
  const reset = () => { setSearch(""); setDateRange("all"); setStatus("all"); setDatabaseType("all"); setPage(1); };

  return <div className="space-y-5">
    <HistoryFilters search={search} onSearchChange={value => { setSearch(value); setPage(1); }}
      dateRange={dateRange} onDateRangeChange={setDateRange} status={status} onStatusChange={setStatus}
      databaseType={databaseType} onDatabaseTypeChange={setDatabaseType} onReset={reset} isLoading={isLoading} />
    <div className="overflow-hidden rounded-lg border bg-white">
      <div className="hidden grid-cols-[minmax(0,1fr)_100px_72px_115px_32px] gap-4 border-b bg-[#fafafa] px-4 py-2.5 text-[11px] text-muted-foreground md:grid">
        <span>Database / file</span><span>Status</span><span>Size</span><span>Created</span><span className="sr-only">Actions</span>
      </div>
      {isLoading ? <div className="space-y-5 p-5"><Skeleton className="h-12 w-full" /><Skeleton className="h-12 w-full" /></div> : error ? <p role="alert" className="p-6 text-sm text-destructive">Couldn’t load backups. Refresh the page to try again.</p> : filtered.length ? filtered.map(backup => {
        const filename = backup.path.split(/[\\/]/).pop() || "Backup file";
        return <div key={backup.id} className="grid grid-cols-[minmax(0,1fr)_auto] items-center gap-3 border-b px-4 py-3.5 last:border-b-0 md:grid-cols-[minmax(0,1fr)_100px_72px_115px_32px] md:gap-4">
          <div className="min-w-0">
            <p className="break-words text-[12px] font-medium">{backup.database_name || typeLabels[backup.database_type as DatabaseType] || backup.database_type}<span className="ms-2 text-[10px] font-normal text-muted-foreground">{backup.s3_object_key ? "S3" : "Local"}</span></p>
            <p className="mt-1 break-all font-mono text-[10px] leading-relaxed text-muted-foreground">{filename}</p>
          </div>
          <div className="flex flex-wrap gap-1"><BackupStatus status={backup.status} /><RestoreCheckStatus status={backup.verification_status} message={backup.verification_message} /></div>
          <p className="text-xs text-muted-foreground">{formatSize(backup.size)}<span className="ml-2 md:hidden">· {formatDistanceToNow(parseISO(backup.created_at), { addSuffix: true })}</span></p>
          <p className="hidden text-xs text-muted-foreground md:block">{formatDistanceToNow(parseISO(backup.created_at), { addSuffix: true })}</p>
          <div className="justify-self-end">
            <DropdownMenu modal={false}>
              <DropdownMenuTrigger asChild><Button id={`backup-actions-${backup.id}`} variant="ghost" size="icon" className="h-8 w-8" aria-label={`Actions for ${filename}`}><MoreHorizontal /></Button></DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuItem disabled={backup.status !== "completed" || isDownloading} onClick={() => downloadBackupFile(backup)}>Download backup</DropdownMenuItem>
                <DropdownMenuItem disabled={backup.status !== "completed"} onClick={() => { setActionTrigger(document.getElementById(`backup-actions-${backup.id}`)); setRestore(backup); }}>Restore database…</DropdownMenuItem>
                {["postgresql", "mysql", "mariadb", "mongodb"].includes(backup.database_type) && <DropdownMenuItem disabled={backup.status !== "completed"} onClick={() => { setActionTrigger(document.getElementById(`backup-actions-${backup.id}`)); setBranch(backup); }}>Branch from backup…</DropdownMenuItem>}
                {["postgresql", "mysql"].includes(backup.database_type) && <DropdownMenuItem disabled={backup.status !== "completed"} onClick={() => { setActionTrigger(document.getElementById(`backup-actions-${backup.id}`)); setSchemaDiff(backup); }}>Compare schema…</DropdownMenuItem>}
                <DropdownMenuSeparator />
                <DropdownMenuItem className="text-destructive focus:text-destructive" onClick={() => { setActionTrigger(document.getElementById(`backup-actions-${backup.id}`)); setDeleting(backup); }}>Delete backup…</DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </div>;
      }) : <div className="px-6 py-12"><h2 className="text-sm font-medium">{search || dateRange !== "all" || status !== "all" || databaseType !== "all" ? "No matching backups on this page." : "No backups yet."}</h2><p className="mt-2 text-xs text-muted-foreground">{search || dateRange !== "all" || status !== "all" || databaseType !== "all" ? "Try another page or clear the filters." : "Run a backup from Connections to get started."}</p></div>}
    </div>
    <div className="flex flex-wrap items-center justify-between gap-4">
      <p className="text-[11px] text-muted-foreground">{dateRange !== "all" || status !== "all" || databaseType !== "all" ? "Filters apply to the current page. " : ""}{pagination ? `${pagination.total} backups · Page ${page} of ${Math.max(1, pagination.total_pages)}` : "Backups refresh every 30 seconds."}</p>
      {pagination && pagination.total_pages > 1 && <CustomPagination currentPage={page} totalPages={pagination.total_pages} onPageChange={setPage} />}
    </div>
    <RestoreDialog backup={restore} open={!!restore} returnFocusTo={actionTrigger} onOpenChange={open => { if (!open) setRestore(null); }} />
    {branch && <BranchDialog backup={branch} returnFocusTo={actionTrigger} onClose={() => setBranch(null)} />}
    {schemaDiff && <SchemaDiffDialog backup={schemaDiff} returnFocusTo={actionTrigger} onClose={() => setSchemaDiff(null)} />}
    <Dialog open={!!deleting} onOpenChange={open => { if (!open && !isDeleting) setDeleting(null); }}>
      <DialogContent onCloseAutoFocus={event => { if (actionTrigger?.isConnected) { event.preventDefault(); actionTrigger.focus(); } }}><DialogHeader><DialogTitle>Delete this backup?</DialogTitle><DialogDescription>The file will be removed from local storage{deleting?.s3_object_key ? " and S3" : ""}. This cannot be undone.</DialogDescription></DialogHeader>
        <DialogFooter><Button variant="outline" onClick={() => setDeleting(null)} disabled={isDeleting}>Cancel</Button><Button variant="destructive" disabled={isDeleting} onClick={() => { if (deleting) deleteBackupById(deleting.id, { onSuccess: () => setDeleting(null) }); }}>{isDeleting ? "Deleting…" : "Delete backup"}</Button></DialogFooter>
      </DialogContent>
    </Dialog>
  </div>;
}