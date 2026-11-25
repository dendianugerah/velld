"use client";

import { format, isValid, parseISO } from "date-fns";
import { ArrowRight } from "lucide-react";
import Link from "next/link";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { useBackup } from "@/hooks/use-backup";
import { useConnections } from "@/hooks/use-connections";
import { formatSize } from "@/lib/helper";
import type { BackupList } from "@/types/backup";

export function ActivityList() {
  const { backups, isLoading, error } = useBackup();
  const { connections, isLoading: loadingConnections, error: connectionError } = useConnections();
  const scheduled = connections?.filter((connection) => connection.backup_enabled).length ?? 0;
  const nextAction = connectionError
    ? { title: "Couldn’t load connections", detail: "Open Connections to check your database sources.", label: "View connections" }
    : !connections?.length
      ? { title: "Add a database connection", detail: "Connect PostgreSQL, MySQL, or MongoDB to run your first backup.", label: "Set up a connection" }
      : scheduled === 0
        ? { title: "Schedule a backup", detail: "Choose a database and set how often it should be backed up.", label: "Set up a schedule" }
        : { title: scheduled === 1 ? "1 active backup schedule" : `${scheduled} active backup schedules`, detail: "Scheduled backups run automatically. You can review or change them from Connections.", label: "Manage schedules" };

  const groups = new Map<string, BackupList[]>();
  for (const backup of backups?.slice(0, 6) ?? []) {
    const date = parseISO(backup.created_at);
    const day = isValid(date) ? format(date, "dd MMM yyyy") : "Unknown date";
    const items = groups.get(day) ?? [];
    items.push(backup);
    groups.set(day, items);
  }

  return (
    <div className="space-y-9">
      <section aria-labelledby="next-up" className="overflow-hidden rounded-xl border">
        <h2 id="next-up" className="flex items-center gap-2.5 border-b bg-muted/70 px-5 py-3 text-[11px] font-semibold uppercase tracking-[0.15em] text-muted-foreground sm:px-6">
          <span className="size-2 rounded-full bg-[#49634d]" aria-hidden="true" /> Next up
        </h2>
        <div className="p-5 sm:p-6">
          {loadingConnections ? <Skeleton className="h-20 w-full max-w-sm" /> : (
            <>
              <p className="text-lg font-semibold tracking-tight">{nextAction.title}</p>
              <p className="mt-1.5 text-sm leading-6 text-muted-foreground">{nextAction.detail}</p>
              <Button asChild className="mt-5">
                <Link href="/connections">{nextAction.label}<ArrowRight aria-hidden="true" /></Link>
              </Button>
            </>
          )}
        </div>
      </section>

      <section aria-labelledby="recent-activity">
        <div className="mb-3 flex items-center justify-between gap-3">
          <h2 id="recent-activity" className="text-base font-semibold tracking-tight">Recent activity</h2>
          <Link href="/history" className="inline-flex items-center gap-2 text-sm font-medium text-muted-foreground hover:text-foreground hover:underline">
            View history <ArrowRight aria-hidden="true" className="size-3.5" />
          </Link>
        </div>
        <div className="overflow-hidden rounded-xl border">
          {isLoading ? (
            <div className="p-5"><Skeleton className="h-16 w-full" /></div>
          ) : groups.size ? (
            Array.from(groups, ([day, items]) => (
              <div key={day}>
                <h3 className="border-b bg-muted/70 px-5 py-2.5 text-[11px] font-semibold uppercase tracking-[0.13em] text-muted-foreground sm:px-6">{day}</h3>
                <ul className="divide-y">
                  {items.map((backup) => {
                    const date = parseISO(backup.created_at);
                    const connection = connections?.find((item) => item.id === backup.connection_id);
                    return (
                      <li key={backup.id} className="flex items-center justify-between gap-4 px-5 py-3.5 sm:px-6">
                        <div className="min-w-0">
                          <p className="truncate text-sm font-medium">{backup.database_name || connection?.name || "Database backup"}</p>
                          <p className="mt-1 truncate text-xs text-muted-foreground">
                            {isValid(date) && <time dateTime={backup.created_at}>{format(date, "p")}</time>}
                            {isValid(date) && connection && " · "}
                            {connection?.name}
                          </p>
                        </div>
                        <div className="shrink-0 text-right text-xs">
                          <p className="font-medium capitalize">{backup.status.replaceAll("_", " ")}</p>
                          {backup.size > 0 && <p className="mt-1 text-muted-foreground">{formatSize(backup.size)}</p>}
                        </div>
                      </li>
                    );
                  })}
                </ul>
              </div>
            ))
          ) : (
            <div className="px-5 py-6 sm:px-6">
              <p className="text-sm font-medium">{error ? "Couldn’t load backup activity" : "No backups yet"}</p>
              <p className="mt-1 text-sm text-muted-foreground">
                {error ? "Open Backup History to try again." : "Backups will appear here after they run."}
              </p>
            </div>
          )}
        </div>
      </section>
    </div>
  );
}