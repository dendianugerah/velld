"use client";

import Link from "next/link";
import { Download } from "lucide-react";
import { Button as KumoButton } from "@cloudflare/kumo/components/button";
import { LayerCard } from "@cloudflare/kumo/components/layer-card";
import { formatDistanceToNow, isValid, parseISO } from "date-fns";
import { useBackup } from "@/hooks/use-backup";
import { useConnections } from "@/hooks/use-connections";
import { formatSize, getScheduleFrequency } from "@/lib/helper";
import { typeLabels } from "@/types/base";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { BackupStatus } from "@/components/ui/backup-status";
import { PageHeader } from "@/components/layout/page-header";
import { AddConnectionDialog } from "@/components/views/connections/add-connection-dialog";

function relativeTime(value?: string) {
  if (!value) return "No backups yet";
  const date = parseISO(value);
  return isValid(date)
    ? formatDistanceToNow(date, { addSuffix: true })
    : "Time unavailable";
}

function frequencyLabel(cron?: string) {
  const frequency = getScheduleFrequency(cron);
  if (!frequency) return "Scheduled";
  return frequency === "test"
    ? "Every minute"
    : frequency === "custom"
      ? "Custom"
      : frequency.charAt(0).toUpperCase() + frequency.slice(1);
}

export function ActivityList() {
  const {
    backups,
    isLoading: backupsLoading,
    error: backupsError,
    createBackup,
    isCreating,
    downloadBackupFile,
    isDownloading,
  } = useBackup();
  const {
    connections,
    isLoading: connectionsLoading,
    error: connectionsError,
    refetch,
  } = useConnections();
  const recent = [...(backups ?? [])]
    .sort((a, b) => Date.parse(b.created_at) - Date.parse(a.created_at))
    .slice(0, 5);

  return (
    <>
      <PageHeader
        title="Overview"
        action={connections?.length ? <AddConnectionDialog /> : undefined}
      />
      <div className="space-y-10">
        <section aria-labelledby="databases-heading">
          <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
            <h2 id="databases-heading" className="text-sm font-semibold">
              Connections
            </h2>
            <Link
              href="/connections"
              className="text-xs font-medium text-muted-foreground hover:text-foreground"
            >
              Manage <span aria-hidden="true">→</span>
            </Link>
          </div>

          {connectionsLoading ? (
            <div className="space-y-3 rounded-lg border p-4">
              {[1, 2, 3].map((key) => (
                <Skeleton key={key} className="h-12 w-full" />
              ))}
            </div>
          ) : connectionsError ? (
            <div
              role="alert"
              className="flex flex-wrap items-center justify-between gap-4 rounded-lg border p-5 text-sm"
            >
              <p className="font-medium">Couldn’t load connections.</p>
              <KumoButton variant="secondary" onClick={() => refetch()}>
                Try again
              </KumoButton>
            </div>
          ) : connections?.length ? (
            <LayerCard className="overflow-hidden rounded-lg shadow-none">
              <div className="hidden grid-cols-[minmax(0,1fr)_145px_130px_112px] items-center gap-4 border-b bg-muted/50 px-4 py-2.5 text-[11px] text-muted-foreground md:grid">
                <span>Connection</span>
                <span>Schedule</span>
                <span>Last backup</span>
                <span className="sr-only">Actions</span>
              </div>
              {connections.slice(0, 6).map((connection) => (
                <div
                  key={connection.id}
                  className="grid grid-cols-[minmax(0,1fr)_auto] items-center gap-x-4 gap-y-3 border-b px-4 py-3.5 last:border-b-0 md:grid-cols-[minmax(0,1fr)_145px_130px_112px]"
                >
                  <div className="col-span-2 flex min-w-0 items-center md:col-span-1">
                    <div className="min-w-0">
                      <Link
                        href={`/connections/${connection.id}`}
                        className="break-words rounded-sm text-sm font-medium hover:underline focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary"
                      >
                        {connection.name}
                      </Link>
                      <p className="mt-1 break-all text-xs text-muted-foreground">
                        {typeLabels[connection.type] ?? connection.type} ·{" "}
                        {connection.database || connection.host}
                      </p>
                      <p className="mt-1 text-[10px] text-muted-foreground md:hidden">
                        {connection.last_backup_time
                          ? `Last backup ${relativeTime(connection.last_backup_time)}`
                          : "No backups yet"}
                      </p>
                    </div>
                  </div>
                  <div>
                    <Link
                      href={`/connections/${connection.id}?tab=schedule`}
                      className="rounded-sm text-[11px] text-muted-foreground underline decoration-border underline-offset-4 hover:text-foreground focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary"
                      aria-label={`View schedule for ${connection.name}`}
                    >
                      {connection.backup_enabled
                        ? frequencyLabel(connection.cron_schedule)
                        : "Manual only"}
                    </Link>
                  </div>
                  <p className="hidden text-[11px] text-muted-foreground md:block">
                    {relativeTime(connection.last_backup_time)}
                  </p>
                  <KumoButton
                    variant="secondary"
                    className="justify-self-end"
                    disabled={isCreating}
                    onClick={() => createBackup(connection.id)}
                    aria-label={`Back up ${connection.name} now`}
                  >
                    {isCreating ? "Starting…" : "Back up now"}
                  </KumoButton>
                </div>
              ))}
            </LayerCard>
          ) : (
            <div className="rounded-xl border border-border bg-muted/40 px-6 py-12 text-center">
              <h3 className="mb-5 text-sm font-semibold">
                Add your first database
              </h3>
              <AddConnectionDialog />
            </div>
          )}
        </section>

        <section aria-labelledby="recent-heading">
          <div className="mb-4 flex items-center justify-between gap-4">
            <h2 id="recent-heading" className="text-sm font-semibold">
              Recent backups
            </h2>
            <Link
              href="/history"
              className="shrink-0 text-xs font-medium text-muted-foreground hover:text-foreground"
            >
              View history <span aria-hidden="true">→</span>
            </Link>
          </div>
          <LayerCard className="overflow-hidden rounded-lg shadow-none">
            <div className="hidden grid-cols-[minmax(0,1fr)_100px_72px_115px_32px] items-center gap-4 border-b bg-muted/50 px-4 py-2.5 text-[11px] text-muted-foreground md:grid">
              <span>Database / file</span>
              <span>Status</span>
              <span className="text-end">Size</span>
              <span>Created</span>
              <span className="sr-only">Download</span>
            </div>
            {backupsLoading ? (
              <div className="space-y-4 p-4">
                <Skeleton className="h-10 w-full" />
                <Skeleton className="h-10 w-full" />
              </div>
            ) : backupsError ? (
              <p role="alert" className="p-5 text-xs text-destructive">
                Couldn’t load backups.
              </p>
            ) : recent.length ? (
              recent.map((backup) => {
                const connection = connections?.find(
                  (c) => c.id === backup.connection_id,
                );
                const filename =
                  backup.path.split(/[\\/]/).pop() || "Backup file";
                return (
                  <div
                    key={backup.id}
                    className="grid grid-cols-[minmax(0,1fr)_auto] items-center gap-x-4 gap-y-2 border-b px-4 py-3.5 last:border-b-0 md:grid-cols-[minmax(0,1fr)_100px_72px_115px_32px] md:gap-4"
                  >
                    <div className="min-w-0">
                      <p className="break-words text-[12px] font-medium">
                        {backup.database_name ||
                          connection?.name ||
                          "Database backup"}
                      </p>
                      <p className="mt-1 break-all font-mono text-[10px] leading-relaxed text-muted-foreground">
                        {filename}
                      </p>
                    </div>
                    <BackupStatus status={backup.status} />
                    <span className="text-[11px] tabular-nums text-muted-foreground md:text-end">
                      {formatSize(backup.size)}
                      <span className="ms-2 md:hidden">
                        · {relativeTime(backup.created_at)}
                      </span>
                    </span>
                    <span className="hidden text-[11px] text-muted-foreground md:block">
                      {relativeTime(backup.created_at)}
                    </span>
                    <div className="justify-self-end">
                      {backup.status === "completed" && (
                        <Button
                          variant="ghost"
                          size="icon"
                          className="h-8 w-8 text-muted-foreground"
                          aria-label={`Download ${filename}`}
                          disabled={isDownloading}
                          onClick={() => downloadBackupFile(backup)}
                        >
                          <Download className="!h-3.5 !w-3.5" />
                        </Button>
                      )}
                    </div>
                  </div>
                );
              })
            ) : (
              <p className="px-5 py-9 text-center text-sm text-muted-foreground">
                No backups yet.
              </p>
            )}
          </LayerCard>
        </section>
      </div>
    </>
  );
}
