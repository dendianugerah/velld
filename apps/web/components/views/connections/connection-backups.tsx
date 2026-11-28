"use client";

import Link from "next/link";
import { Button } from "@cloudflare/kumo/components/button";
import { LayerCard } from "@cloudflare/kumo/components/layer-card";
import { BackupStatus } from "@/components/ui/backup-status";
import { Skeleton } from "@/components/ui/skeleton";
import { useBackup } from "@/hooks/use-backup";
import { formatSize } from "@/lib/helper";
import { connectionTime } from "./connection-utils";

export function ConnectionBackups({ connectionId }: { connectionId: string }) {
  const { backups, isLoading, error, refetch } = useBackup({
    connection_id: connectionId,
    limit: 5,
  });
  const filter = new URLSearchParams({
    connection_id: connectionId,
  }).toString();
  const historyHref = `/history?${filter}`;

  return (
    <section aria-labelledby="connection-backups-heading" className="space-y-4">
      <div className="flex items-center justify-between gap-4">
        <h2 id="connection-backups-heading" className="text-sm font-semibold">
          Recent backups
        </h2>
        <Link
          href={historyHref}
          className="rounded-sm text-xs font-medium text-muted-foreground hover:text-foreground focus-visible:outline focus-visible:outline-2 focus-visible:outline-ring"
        >
          View backup history
        </Link>
      </div>
      <LayerCard className="overflow-hidden rounded-lg shadow-none">
        {isLoading ? (
          <div className="space-y-4 p-5" aria-label="Loading backups">
            <Skeleton className="h-10 w-full" />
            <Skeleton className="h-10 w-full" />
          </div>
        ) : error ? (
          <div
            role="alert"
            className="flex flex-wrap items-center justify-between gap-4 p-5 text-sm"
          >
            <p>Couldn’t load backups.</p>
            <Button variant="secondary" onClick={() => refetch()}>
              Try again
            </Button>
          </div>
        ) : backups?.length ? (
          backups.map((backup) => {
            const filename = backup.path.split(/[\\/]/).pop() || "Backup file";
            const href = `/history/${encodeURIComponent(backup.id)}?${new URLSearchParams({ from: filter })}`;
            return (
              <div
                key={backup.id}
                className="flex flex-wrap items-center justify-between gap-x-5 gap-y-3 border-b px-5 py-4 last:border-0"
              >
                <div className="min-w-0 flex-1">
                  <Link
                    href={href}
                    className="break-all text-sm font-medium hover:underline focus-visible:outline focus-visible:outline-2 focus-visible:outline-ring"
                  >
                    {filename}
                  </Link>
                  <p className="mt-1 text-xs text-muted-foreground">
                    {connectionTime(backup.created_at)} ·{" "}
                    {formatSize(backup.size)}
                  </p>
                </div>
                <BackupStatus status={backup.status} />
              </div>
            );
          })
        ) : (
          <p className="px-5 py-8 text-center text-sm text-muted-foreground">
            No backups for this connection yet.
          </p>
        )}
      </LayerCard>
    </section>
  );
}
