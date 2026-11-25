"use client";

import { useState } from "react";
import { ArrowDown, ArrowUp, MoreHorizontal } from "lucide-react";
import { Button } from "@/components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuSeparator, DropdownMenuTrigger } from "@/components/ui/dropdown-menu";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { useBackup } from "@/hooks/use-backup";
import { formatSize, getScheduleFrequency } from "@/lib/helper";
import { typeLabels } from "@/types/base";
import type { Connection } from "@/types/connection";

interface ConnectionsTableProps {
  connections: Connection[];
  onEdit: (connectionId: string) => void;
  onDelete: (connection: Connection) => void;
  onSchedule: (connectionId: string) => void;
  onDiscover?: (connection: Connection) => void;
}

const sortColumns = {
  name: { label: "Connection", value: (connection: Connection) => connection.name },
  type: { label: "Database", value: (connection: Connection) => connection.type },
  lastBackup: { label: "Last backup", value: (connection: Connection) => connection.last_backup_time ? new Date(connection.last_backup_time).getTime() : 0 },
  size: { label: "Size", value: (connection: Connection) => connection.database_size || 0 },
};

type SortField = keyof typeof sortColumns;

export function ConnectionsTable({ connections, onEdit, onDelete, onSchedule, onDiscover }: ConnectionsTableProps) {
  const { createBackup, isCreating } = useBackup();
  const [sort, setSort] = useState<{ field: SortField; descending: boolean }>({ field: "name", descending: false });

  const sortedConnections = [...connections].sort((a, b) => {
    const left = sortColumns[sort.field].value(a);
    const right = sortColumns[sort.field].value(b);
    const order = typeof left === "number" && typeof right === "number" ? left - right : String(left).localeCompare(String(right));
    return sort.descending ? -order : order;
  });

  const renderSortHeader = (field: SortField) => {
    const active = sort.field === field;
    const Direction = sort.descending ? ArrowDown : ArrowUp;
    return (
      <TableHead className="h-10" aria-sort={active ? (sort.descending ? "descending" : "ascending") : "none"}>
        <button
          className="inline-flex items-center gap-1.5 whitespace-nowrap text-xs font-semibold hover:text-foreground"
          onClick={() => setSort({ field, descending: active && !sort.descending })}
        >
          {sortColumns[field].label}
          {active && <Direction aria-hidden="true" className="size-3" />}
        </button>
      </TableHead>
    );
  };

  return (
    <div className="overflow-hidden rounded-xl border">
      <Table>
        <TableHeader className="bg-muted/70">
          <TableRow className="hover:bg-transparent">
            {renderSortHeader("name")}
            {renderSortHeader("type")}
            <TableHead className="h-10 whitespace-nowrap text-xs font-semibold">Schedule</TableHead>
            {renderSortHeader("lastBackup")}
            {renderSortHeader("size")}
            <TableHead><span className="sr-only">Actions</span></TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {sortedConnections.map((connection) => (
            <TableRow key={connection.id} className="hover:bg-muted/40">
              <TableCell className="min-w-[9rem]">
                <p className="max-w-[16rem] truncate font-medium" title={connection.name}>{connection.name}</p>
                <p className="mt-1 max-w-[16rem] truncate text-xs text-muted-foreground" title={connection.host}>{connection.host}</p>
              </TableCell>
              <TableCell className="whitespace-nowrap text-muted-foreground">{typeLabels[connection.type] ?? connection.type}</TableCell>
              <TableCell className="whitespace-nowrap text-muted-foreground">
                {connection.backup_enabled ? getScheduleFrequency(connection.cron_schedule) || "Enabled" : "Not scheduled"}
              </TableCell>
              <TableCell className="whitespace-nowrap text-muted-foreground">
                {connection.last_backup_time ? new Date(connection.last_backup_time).toLocaleString("en", { month: "short", day: "numeric", hour: "numeric", minute: "2-digit" }) : "—"}
              </TableCell>
              <TableCell className="whitespace-nowrap text-muted-foreground">{formatSize(connection.database_size)}</TableCell>
              <TableCell className="text-right">
                <div className="flex justify-end gap-1">
                  <Button variant="outline" size="sm" onClick={() => createBackup(connection.id)} disabled={isCreating}>Back up</Button>
                  <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                      <Button variant="ghost" size="sm" className="w-8 p-0" aria-label={`Actions for ${connection.name}`}><MoreHorizontal /></Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end">
                      <DropdownMenuItem onSelect={() => onSchedule(connection.id)}>Manage schedule</DropdownMenuItem>
                      {onDiscover && <DropdownMenuItem onSelect={() => onDiscover(connection)}>Discover databases</DropdownMenuItem>}
                      <DropdownMenuItem onSelect={() => onEdit(connection.id)}>Edit connection</DropdownMenuItem>
                      <DropdownMenuSeparator />
                      <DropdownMenuItem className="text-destructive focus:text-destructive" onSelect={() => onDelete(connection)}>Delete connection</DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </div>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}
