"use client";

import { useState } from "react";
import { ArrowUpDown, MoreHorizontal } from "lucide-react";
import { formatDistanceToNow } from "date-fns";
import { getScheduleFrequency } from "@/lib/helper";
import { Button } from "@/components/ui/button";
import { DatabaseMark } from "@/components/ui/database-mark";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuSeparator, DropdownMenuTrigger } from "@/components/ui/dropdown-menu";
import { typeLabels } from "@/types/base";
import { useBackup } from "@/hooks/use-backup";
import type { Connection } from "@/types/connection";
import { PointInTimeRecoveryDialog } from "./point-in-time-recovery-dialog";

interface ConnectionsTableProps {
  connections: Connection[];
  onEdit: (connectionId: string) => void;
  onDelete: (connection: Connection) => void;
  onSchedule: (connectionId: string, trigger?: HTMLElement | null) => void;
  onDiscover?: (connection: Connection) => void;
  onUpgrade?: (connection: Connection) => void;
}

const upgradeableTypes = new Set(["postgresql", "mysql", "mariadb", "mongodb"]);

export function ConnectionsTable({ connections, onEdit, onDelete, onSchedule, onDiscover, onUpgrade }: ConnectionsTableProps) {
  const { createBackup, isCreating } = useBackup();
  const [sort, setSort] = useState<"name" | "lastBackup">("name");
  const [ascending, setAscending] = useState(true);
  const [recovery, setRecovery] = useState<Connection | null>(null);
  const sorted = [...connections].sort((a, b) => {
    const result = sort === "name" ? a.name.localeCompare(b.name) : (Date.parse(a.last_backup_time || "") || 0) - (Date.parse(b.last_backup_time || "") || 0);
    return ascending ? result : -result;
  });
  const sortBy = (field: typeof sort) => { setSort(field); setAscending(sort === field ? !ascending : true); };

  return <div className="overflow-hidden rounded-lg border bg-white">
    <div className="hidden grid-cols-[minmax(0,1fr)_130px_120px_165px] items-center gap-4 border-b bg-[#fafafa] px-4 py-2.5 text-[11px] text-muted-foreground xl:grid">
      <button onClick={() => sortBy("name")} className="flex w-fit items-center gap-2" aria-label="Sort connections by name">Connection<ArrowUpDown className="h-3 w-3" /></button>
      <span>Schedule</span>
      <button onClick={() => sortBy("lastBackup")} className="flex w-fit items-center gap-2" aria-label="Sort connections by last backup">Last backup<ArrowUpDown className="h-3 w-3" /></button>
      <span className="text-right">Actions</span>
    </div>
    {sorted.map(connection => {
      const frequency = getScheduleFrequency(connection.cron_schedule);
      return <div key={connection.id} className="grid grid-cols-[minmax(0,1fr)_auto] items-center gap-x-4 gap-y-3 border-b px-4 py-3.5 last:border-b-0 xl:grid-cols-[minmax(0,1fr)_130px_120px_165px]">
        <div className="col-span-2 flex min-w-0 items-start gap-3.5 xl:col-span-1">
          <DatabaseMark type={connection.type} small />
          <div className="min-w-0">
          <p className="break-words text-[13px] font-medium">{connection.name}</p>
          <p className="mt-1 break-all text-[11px] text-muted-foreground">{typeLabels[connection.type] ?? connection.type} · {connection.host}:{connection.port}</p>
          </div>
        </div>
        <button className="w-fit rounded-md border bg-white px-2 py-1.5 text-start text-[11px] capitalize hover:bg-muted" onClick={event => onSchedule(connection.id, event.currentTarget)} aria-label={`Edit schedule for ${connection.name}`}>
          {connection.backup_enabled ? (frequency === "test" ? "Every minute" : frequency === "custom" ? "Custom schedule" : frequency || "Scheduled") : "Add schedule"}
        </button>
        <p className="text-end text-[11px] text-muted-foreground xl:text-start">{connection.last_backup_time ? formatDistanceToNow(new Date(connection.last_backup_time), { addSuffix: true }) : "No backups yet"}</p>
        <div className="col-span-2 flex items-center justify-end gap-2 xl:col-span-1">
          <Button variant="outline" size="sm" className="text-[11px]" onClick={() => createBackup(connection.id)} disabled={isCreating} aria-label={`Back up ${connection.name} now`}>{isCreating ? "Starting…" : "Back up now"}</Button>
          <DropdownMenu modal={false}>
            <DropdownMenuTrigger asChild><Button id={`connection-actions-${connection.id}`} variant="ghost" size="icon" className="h-8 w-8" aria-label={`Actions for ${connection.name}`}><MoreHorizontal /></Button></DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem onClick={() => onSchedule(connection.id, document.getElementById(`connection-actions-${connection.id}`))}>Backup schedule</DropdownMenuItem>
              {onDiscover && <DropdownMenuItem onClick={() => onDiscover(connection)}>Select databases</DropdownMenuItem>}
              {onUpgrade && upgradeableTypes.has(connection.type) && <DropdownMenuItem onClick={() => onUpgrade(connection)}>Upgrade container</DropdownMenuItem>}
              {["postgresql", "mysql"].includes(connection.type) && <DropdownMenuItem onClick={() => setRecovery(connection)}>Point-in-time recovery…</DropdownMenuItem>}
              <DropdownMenuItem onClick={() => onEdit(connection.id)}>Edit connection</DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem onClick={() => onDelete(connection)} className="text-destructive focus:text-destructive">Delete connection</DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>;
    })}
    {recovery && <PointInTimeRecoveryDialog connection={recovery} onClose={() => setRecovery(null)} />}
  </div>;
}
