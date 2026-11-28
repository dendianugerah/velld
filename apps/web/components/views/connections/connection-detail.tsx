"use client";

import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { useState } from "react";
import { ArrowLeft } from "lucide-react";
import { Button } from "@cloudflare/kumo/components/button";
import { Button as IconButton, buttonVariants } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { DatabaseMark } from "@/components/ui/database-mark";
import { useConnection, useConnections } from "@/hooks/use-connections";
import { useBackup } from "@/hooks/use-backup";
import { typeLabels } from "@/types/base";
import { formatSize } from "@/lib/helper";
import { connectionTime, scheduleLabel } from "./connection-utils";
import { cn } from "@/lib/utils";
import { DatabaseSelection } from "./database-selection";
import { ScheduleForm } from "./connection-list/schedule-form";
import { DeleteConnectionDialog } from "./connection-list/delete-connection-dialog";
import { PointInTimeRecoveryDialog } from "./connection-list/point-in-time-recovery-dialog";
import { UpgradeDialog } from "./connection-list/upgrade-dialog";
import {
  connectionSectionClass,
  connectionSectionsClass,
} from "./connection-section-styles";
import { ConnectionBackups } from "./connection-backups";

const sections = [
  { id: "overview", label: "Overview" },
  { id: "schedule", label: "Backup schedule" },
  { id: "recovery", label: "Recovery" },
  { id: "configuration", label: "Configuration" },
];

export function ConnectionDetail({ id }: { id: string }) {
  const router = useRouter();
  const searchParams = useSearchParams();
  const requested = searchParams.get("tab");
  const detail = useConnection(id);
  const list = useConnections();
  const { createBackup, isCreating } = useBackup();
  const [deleting, setDeleting] = useState(false);
  const connection = detail.data && {
    ...detail.data,
    ...list.connections?.find((value) => value.id === id),
  };
  const isD1 = connection?.type === "d1";
  const visibleSections = sections.filter(
    (item) => !isD1 || item.id !== "recovery",
  );
  const section = visibleSections.some((item) => item.id === requested)
    ? requested
    : "overview";
  const href = `/connections/${id}`;
  const historyHref = `/history?connection_id=${encodeURIComponent(id)}`;

  return (
    <>
      <Link
        href="/connections"
        className="mb-7 inline-flex min-h-8 items-center gap-2 rounded-sm text-xs text-muted-foreground hover:text-foreground focus-visible:outline focus-visible:outline-2 focus-visible:outline-ring"
      >
        <ArrowLeft className="h-3.5 w-3.5" aria-hidden="true" />
        All connections
      </Link>
      {detail.isLoading || list.isLoading ? (
        <div className="mx-auto max-w-[720px] space-y-8">
          <Skeleton className="h-16 w-2/3" />
          <Skeleton className="h-11 w-full" />
          <Skeleton className="h-52 w-full" />
        </div>
      ) : !connection || detail.error || list.error ? (
        <div role="alert" className="mx-auto max-w-[720px] space-y-4 text-sm">
          <h1 className="text-xl font-semibold">Connection unavailable</h1>
          <Button
            variant="secondary"
            onClick={() => {
              detail.refetch();
              list.refetch();
            }}
          >
            Try again
          </Button>
        </div>
      ) : (
        <div className="mx-auto max-w-[720px]">
          <header className="flex flex-col justify-between gap-6 sm:flex-row sm:items-center">
            <div className="flex min-w-0 items-center gap-4">
              <DatabaseMark type={connection.type} />
              <div className="min-w-0">
                <h1 className="break-words text-2xl font-semibold tracking-[-0.035em]">
                  {connection.name}
                </h1>
                <p className="mt-1.5 break-all text-[13px] text-muted-foreground">
                  {typeLabels[connection.type] ?? connection.type}
                  {!isD1 && ` · ${connection.host}:${connection.port}`}
                </p>
              </div>
            </div>
            <Button onClick={() => createBackup(id)} disabled={isCreating}>
              {isCreating ? "Starting backup…" : "Back up now"}
            </Button>
          </header>
          <nav
            aria-label="Connection sections"
            className={cn("mb-8 mt-8", connectionSectionsClass)}
          >
            {visibleSections.map((item) => (
              <Link
                key={item.id}
                href={item.id === "overview" ? href : `${href}?tab=${item.id}`}
                scroll={false}
                aria-current={section === item.id ? "page" : undefined}
                className={connectionSectionClass(section === item.id)}
              >
                {item.label}
              </Link>
            ))}
          </nav>
          <div hidden={section !== "overview"} className="space-y-10">
            <dl className="grid grid-cols-2 gap-x-6 gap-y-6 rounded-lg border px-5 py-6 sm:grid-cols-3">
              <Stat
                label="Last backup"
                value={connectionTime(connection.last_backup_time)}
              />
              <Stat label="Backup schedule" value={scheduleLabel(connection)} />
              <Stat
                label="Database size"
                value={
                  Number.isFinite(connection.database_size)
                    ? formatSize(connection.database_size)
                    : "Unavailable"
                }
              />
            </dl>
            {!isD1 && (
              <DatabaseSelection
                key={`${id}:${JSON.stringify(connection.selected_databases)}`}
                connection={connection}
              />
            )}
            {section === "overview" && <ConnectionBackups connectionId={id} />}
          </div>
          <section
            hidden={section !== "schedule"}
            aria-labelledby="schedule-heading"
          >
            <h2 id="schedule-heading" className="text-sm font-semibold">
              Backup schedule
            </h2>
            <div className="mt-6">
              <ScheduleForm
                key={`${id}:${connection.backup_enabled}:${connection.cron_schedule}:${connection.retention_days}:${connection.s3_cleanup_on_retention}`}
                connection={connection}
              />
            </div>
          </section>
          {section === "recovery" && (
            <div className="space-y-9">
              {["postgresql", "mysql"].includes(connection.type) ? (
                <PointInTimeRecoveryDialog
                  key={id}
                  connection={connection}
                  inline
                />
              ) : (
                <div>
                  <h2 className="text-sm font-semibold">
                    Point-in-time recovery
                  </h2>
                  <p className="mt-2 text-[13px] leading-relaxed text-muted-foreground">
                    Available for PostgreSQL and MySQL. For this connection,
                    restore a completed backup from Backup history.
                  </p>
                  <Link
                    href={historyHref}
                    className="mt-4 inline-block text-sm underline underline-offset-4"
                  >
                    View backup history
                  </Link>
                </div>
              )}
              <div className="border-t pt-8">
                <UpgradeDialog key={id} connection={connection} inline />
              </div>
            </div>
          )}
          <section
            hidden={section !== "configuration"}
            className="space-y-9"
            aria-labelledby="configuration-heading"
          >
            <div>
              <div className="flex flex-wrap items-center justify-between gap-4">
                <h2
                  id="configuration-heading"
                  className="text-sm font-semibold"
                >
                  Connection configuration
                </h2>
                <Link
                  href={`${href}/edit`}
                  className={buttonVariants({ variant: "outline", size: "sm" })}
                >
                  Edit connection
                </Link>
              </div>
              <dl className="mt-5 divide-y rounded-lg border px-5">
                <Config
                  label="Engine"
                  value={typeLabels[connection.type] ?? connection.type}
                />
                {isD1 ? (
                  <>
                    <Config
                      label="Cloudflare account ID"
                      value={connection.d1_account_id ?? ""}
                    />
                    <Config
                      label="D1 database ID"
                      value={connection.d1_database_id ?? ""}
                    />
                  </>
                ) : (
                  <>
                    <Config label="Host" value={connection.host} />
                    <Config label="Port" value={String(connection.port)} />
                    <Config
                      label="Username"
                      value={connection.username || "Not set"}
                    />
                    <Config
                      label="Default database"
                      value={connection.database_name || "Not set"}
                    />
                    <Config
                      label="TLS"
                      value={connection.ssl ? "Enabled" : "Disabled"}
                    />
                    <Config
                      label="SSH tunnel"
                      value={
                        connection.ssh_enabled
                          ? `${connection.ssh_username}@${connection.ssh_host}:${connection.ssh_port}`
                          : "Disabled"
                      }
                    />
                  </>
                )}
              </dl>
            </div>
            <div className="flex flex-col justify-between gap-4 border-t pt-7 sm:flex-row sm:items-center">
              <div>
                <h3 className="text-sm font-semibold">Delete connection</h3>
              </div>
              <IconButton
                variant="outline"
                onClick={() => setDeleting(true)}
                className="w-fit text-destructive"
              >
                Delete connection…
              </IconButton>
            </div>
          </section>
          {deleting && (
            <DeleteConnectionDialog
              connection={connection}
              onClose={() => setDeleting(false)}
              onDeleted={() => router.push("/connections")}
            />
          )}
        </div>
      )}
    </>
  );
}

function Stat({ label, value }: { label: string; value: string }) {
  return (
    <div className="min-w-0">
      <dt className="text-xs text-muted-foreground">{label}</dt>
      <dd className="mt-2 break-words text-sm font-medium">{value}</dd>
    </div>
  );
}

function Config({ label, value }: { label: string; value: string }) {
  return (
    <div className="grid gap-1.5 py-4 text-[13px] sm:grid-cols-[160px_minmax(0,1fr)] sm:gap-5">
      <dt className="text-muted-foreground">{label}</dt>
      <dd className="break-all">{value}</dd>
    </div>
  );
}
