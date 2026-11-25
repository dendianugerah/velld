"use client";

import Link from "next/link";
import { useState } from "react";
import { ChevronDown, Download } from "lucide-react";
import { formatDistanceToNow, isValid, parseISO } from "date-fns";
import { useBackup } from "@/hooks/use-backup";
import { useConnections } from "@/hooks/use-connections";
import { formatSize, getScheduleFrequency } from "@/lib/helper";
import { typeLabels } from "@/types/base";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { DatabaseMark } from "@/components/ui/database-mark";
import { BackupStatus } from "@/components/ui/backup-status";
import { PageHeader } from "@/components/layout/page-header";
import { BackupScheduleDialog } from "@/components/views/connections/connection-list/backup-schedule-dialog";
import { AddConnectionDialog } from "@/components/views/connections/add-connection-dialog";

function relativeTime(value?: string) {
  if (!value) return "No backups yet";
  const date = parseISO(value);
  return isValid(date) ? formatDistanceToNow(date, { addSuffix: true }) : "Time unavailable";
}

function frequencyLabel(cron?: string) {
  const frequency = getScheduleFrequency(cron);
  return frequency === "test" ? "Every minute" : frequency === "custom" ? "Custom" : frequency || "Scheduled";
}

export function ActivityList() {
  const { backups, isLoading: backupsLoading, error: backupsError, createBackup, isCreating, downloadBackupFile, isDownloading } = useBackup();
  const { connections, isLoading: connectionsLoading, error: connectionsError, refetch } = useConnections();
  const [scheduleId, setScheduleId] = useState<string | null>(null);
  const [scheduleTrigger, setScheduleTrigger] = useState<HTMLElement | null>(null);
  const scheduled = connections?.filter(c => c.backup_enabled).length ?? 0;
  const recent = [...(backups ?? [])].sort((a, b) => Date.parse(b.created_at) - Date.parse(a.created_at)).slice(0, 5);

  return <>
    <PageHeader title="Overview" description="Your databases and their latest backups." action={connections?.length ? <AddConnectionDialog /> : undefined} />
    <div className="space-y-9">
      <section aria-labelledby="databases-heading">
        <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
          <div className="flex flex-wrap items-center gap-x-3 gap-y-1">
            <h2 id="databases-heading" className="text-[13px] font-semibold">Connections</h2>
            {!!connections?.length && <span className="text-[11px] text-muted-foreground">{connections.length} total · {scheduled} scheduled</span>}
          </div>
          <Link href="/connections" className="text-[11px] font-medium text-muted-foreground hover:text-foreground">Manage connections <span aria-hidden="true">→</span></Link>
        </div>

        {connectionsLoading ? <div className="space-y-3 rounded-lg border p-4">{[1, 2, 3].map(key => <Skeleton key={key} className="h-12 w-full" />)}</div> : connectionsError ? (
          <div role="alert" className="flex flex-wrap items-center justify-between gap-4 rounded-lg border p-5 text-sm">
            <div><p className="font-medium">Your connections couldn’t be loaded.</p><p className="mt-1.5 text-xs text-muted-foreground">Check that the Velld API is running, then try again.</p></div>
            <Button variant="outline" size="sm" onClick={() => refetch()}>Try again</Button>
          </div>
        ) : connections?.length ? (
          <div className="overflow-hidden rounded-lg border">
            <div className="hidden grid-cols-[minmax(0,1fr)_145px_130px_112px] items-center gap-4 border-b bg-[#fafafa] px-4 py-2.5 text-[11px] text-muted-foreground md:grid"><span>Connection</span><span>Schedule</span><span>Last backup</span><span className="sr-only">Actions</span></div>
            {connections.slice(0, 6).map(connection => <div key={connection.id} className="grid grid-cols-[minmax(0,1fr)_auto] items-center gap-x-4 gap-y-3 border-b px-4 py-3.5 last:border-b-0 md:grid-cols-[minmax(0,1fr)_145px_130px_112px]">
              <div className="col-span-2 flex min-w-0 items-center gap-3 md:col-span-1">
                <DatabaseMark type={connection.type} small />
                <div className="min-w-0">
                  <p className="break-words text-[13px] font-medium">{connection.name}</p>
                  <p className="mt-1 break-all text-[11px] text-muted-foreground">{typeLabels[connection.type] ?? connection.type} · {connection.database || connection.host}</p>
                  <p className="mt-1 text-[10px] text-muted-foreground md:hidden">{connection.last_backup_time ? `Last backup ${relativeTime(connection.last_backup_time)}` : "No backups yet"}</p>
                </div>
              </div>
              <div>
                <button onClick={event => { setScheduleTrigger(event.currentTarget); setScheduleId(connection.id); }} className="inline-flex items-center gap-2 rounded-md border bg-white px-2 py-1.5 text-[11px] font-medium capitalize hover:bg-muted focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary" aria-label={`Edit schedule for ${connection.name}`}>
                  {!connection.backup_enabled && <span className="h-1.5 w-1.5 rounded-full bg-[#b78c45]" aria-hidden="true" />}
                  {connection.backup_enabled ? frequencyLabel(connection.cron_schedule) : "Manual only"}<ChevronDown className="h-3 w-3 text-muted-foreground" />
                </button>
                {connection.backup_enabled && !!connection.retention_days && <p className="mt-1.5 text-[10px] text-muted-foreground">{connection.retention_days}-day retention</p>}
                {connection.backup_enabled && !!connection.retention_count && <p className="mt-1 text-[10px] text-muted-foreground">Latest {connection.retention_count} backups</p>}
              </div>
              <p className="hidden text-[11px] text-muted-foreground md:block">{relativeTime(connection.last_backup_time)}</p>
              <Button variant="outline" size="sm" className="h-8 justify-self-end text-[11px]" disabled={isCreating} onClick={() => createBackup(connection.id)} aria-label={`Back up ${connection.name} now`}>{isCreating ? "Starting…" : "Back up now"}</Button>
            </div>)}
          </div>
        ) : (
          <div className="rounded-lg border px-6 py-9">
            <div className="mb-4 flex gap-2"><DatabaseMark type="postgresql" small /><DatabaseMark type="mysql" small /><DatabaseMark type="mongodb" small /></div>
            <h3 className="text-base font-semibold tracking-tight">A place for your first database.</h3>
            <p className="mb-5 mt-2 max-w-md text-xs leading-relaxed text-muted-foreground">Add a PostgreSQL, MySQL, or MongoDB connection, then run a backup or set a schedule.</p>
            <AddConnectionDialog />
          </div>
        )}
        {connections && connections.length > 6 && <p className="mt-3 text-[11px] text-muted-foreground">Showing 6 of {connections.length}. <Link href="/connections" className="text-primary underline underline-offset-4">View all connections</Link></p>}
      </section>

      <section aria-labelledby="recent-heading">
        <div className="mb-4 flex items-center justify-between gap-4">
          <div className="flex flex-wrap items-center gap-x-3 gap-y-1"><h2 id="recent-heading" className="text-[13px] font-semibold">Recent backups</h2><span className="text-[11px] text-muted-foreground">Updated every 30 seconds</span></div>
          <Link href="/history" className="shrink-0 text-[11px] font-medium text-muted-foreground hover:text-foreground">View history <span aria-hidden="true">→</span></Link>
        </div>
        <div className="overflow-hidden rounded-lg border">
          <div className="hidden grid-cols-[minmax(0,1fr)_100px_72px_115px_32px] items-center gap-4 border-b bg-[#fafafa] px-4 py-2.5 text-[11px] text-muted-foreground md:grid"><span>Database / file</span><span>Status</span><span className="text-end">Size</span><span>Created</span><span className="sr-only">Download</span></div>
          {backupsLoading ? <div className="space-y-4 p-4"><Skeleton className="h-10 w-full" /><Skeleton className="h-10 w-full" /></div> : backupsError ? (
            <p role="alert" className="p-5 text-xs text-destructive">Couldn’t load recent backups. Try refreshing the page.</p>
          ) : recent.length ? recent.map(backup => {
            const connection = connections?.find(c => c.id === backup.connection_id);
            const filename = backup.path.split(/[\\/]/).pop() || "Backup file";
            return <div key={backup.id} className="grid grid-cols-[minmax(0,1fr)_auto] items-center gap-x-4 gap-y-2 border-b px-4 py-3.5 last:border-b-0 md:grid-cols-[minmax(0,1fr)_100px_72px_115px_32px] md:gap-4">
              <div className="min-w-0"><p className="break-words text-[12px] font-medium">{backup.database_name || connection?.name || "Database backup"}</p><p className="mt-1 break-all font-mono text-[10px] leading-relaxed text-muted-foreground">{filename}</p></div>
              <BackupStatus status={backup.status} />
              <span className="text-[11px] tabular-nums text-muted-foreground md:text-end">{formatSize(backup.size)}<span className="ms-2 md:hidden">· {relativeTime(backup.created_at)}</span></span>
              <span className="hidden text-[11px] text-muted-foreground md:block">{relativeTime(backup.created_at)}</span>
              <div className="justify-self-end">{backup.status === "completed" && <Button variant="ghost" size="icon" className="h-8 w-8 text-muted-foreground" aria-label={`Download ${filename}`} disabled={isDownloading} onClick={() => downloadBackupFile(backup)}><Download className="!h-3.5 !w-3.5" /></Button>}</div>
            </div>;
          }) : <div className="px-5 py-8"><p className="text-sm font-medium">Your first backup goes here.</p><p className="mt-2 text-xs text-muted-foreground">Choose a connection above and select Back up now.</p></div>}
        </div>
      </section>
      <BackupScheduleDialog connectionId={scheduleId} connection={connections?.find(c => c.id === scheduleId)} returnFocusTo={scheduleTrigger} onClose={() => setScheduleId(null)} />
    </div>
  </>;
}