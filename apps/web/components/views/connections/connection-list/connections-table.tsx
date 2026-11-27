"use client";

import Link from "next/link";
import { useState } from "react";
import { ArrowDown, ArrowUp, MoreHorizontal } from "lucide-react";
import { Button } from "@/components/ui/button";
import { DatabaseMark } from "@/components/ui/database-mark";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuSeparator, DropdownMenuTrigger } from "@/components/ui/dropdown-menu";
import { useBackup } from "@/hooks/use-backup";
import { typeLabels } from "@/types/base";
import { connectionTime, scheduleLabel } from "../connection-utils";
import type { Connection } from "@/types/connection";

export function ConnectionsTable({ connections, onDelete }: { connections: Connection[]; onDelete: (connection: Connection) => void }) {
  const { createBackup, isCreating } = useBackup();
  const [sort, setSort] = useState<"name" | "lastBackup">("name");
  const [ascending, setAscending] = useState(true);
  const sorted = [...connections].sort((a, b) => (ascending ? 1 : -1) * (sort === "name" ? a.name.localeCompare(b.name) : (Date.parse(a.last_backup_time || "") || 0) - (Date.parse(b.last_backup_time || "") || 0)));
  function sortBy(field: typeof sort) { setSort(field); setAscending(sort === field ? !ascending : true); }
  const SortIcon = ascending ? ArrowUp : ArrowDown;

  return <div className="overflow-hidden rounded-lg border">
    <div className="hidden grid-cols-[minmax(0,1fr)_125px_125px_36px] items-center gap-4 border-b bg-muted/40 px-5 py-3 text-xs text-muted-foreground md:grid">
      <button type="button" onClick={() => sortBy("name")} className="flex w-fit items-center gap-2 rounded-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-ring" aria-label={`Sort by name${sort === "name" ? ascending ? ", descending" : ", ascending" : ""}`}>Connection{sort === "name" && <SortIcon className="h-3 w-3" aria-hidden="true" />}</button>
      <span>Schedule</span><button type="button" onClick={() => sortBy("lastBackup")} className="flex w-fit items-center gap-2 rounded-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-ring" aria-label="Sort by last backup">Last backup{sort === "lastBackup" && <SortIcon className="h-3 w-3" aria-hidden="true" />}</button><span className="sr-only">Actions</span>
    </div>
    {sorted.map(connection => {
      const href = `/connections/${connection.id}`;
      return <div key={connection.id} className="grid grid-cols-[minmax(0,1fr)_36px] items-center gap-x-4 gap-y-4 border-b px-4 py-5 last:border-b-0 sm:px-5 md:grid-cols-[minmax(0,1fr)_125px_125px_36px]">
        <div className="min-w-0"><Link href={href} className="group flex w-fit max-w-full items-center gap-3 rounded-md focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-ring"><DatabaseMark type={connection.type} small /><div className="min-w-0"><p className="break-words text-sm font-medium group-hover:underline group-hover:underline-offset-4">{connection.name}</p><p className="mt-1 break-all text-xs text-muted-foreground">{typeLabels[connection.type] ?? connection.type} · {connection.host}</p></div></Link></div>
        <div className="col-start-1 row-start-2 flex flex-wrap items-center gap-x-6 gap-y-2 ps-[46px] text-xs md:contents">
          <Link href={`${href}?tab=schedule`} className="rounded-sm text-muted-foreground underline decoration-border underline-offset-4 hover:text-foreground focus-visible:outline focus-visible:outline-2 focus-visible:outline-ring">{scheduleLabel(connection)}</Link>
          <span className="text-muted-foreground"><span className="md:hidden">Last backup </span>{connectionTime(connection.last_backup_time)}</span>
        </div>
        <DropdownMenu modal={false}><DropdownMenuTrigger asChild><Button variant="ghost" size="icon" className="col-start-2 row-start-1 md:col-start-4" aria-label={`Actions for ${connection.name}`}><MoreHorizontal /></Button></DropdownMenuTrigger><DropdownMenuContent align="end"><DropdownMenuItem asChild><Link href={href}>View connection</Link></DropdownMenuItem><DropdownMenuItem disabled={isCreating} onClick={() => createBackup(connection.id)}>{isCreating ? "Starting backup…" : "Back up now"}</DropdownMenuItem><DropdownMenuItem asChild><Link href={`${href}/edit`}>Edit connection</Link></DropdownMenuItem><DropdownMenuItem asChild><Link href={`${href}?tab=recovery`}>Recovery & upgrades</Link></DropdownMenuItem><DropdownMenuSeparator /><DropdownMenuItem className="text-destructive focus:text-destructive" onClick={() => onDelete(connection)}>Delete connection…</DropdownMenuItem></DropdownMenuContent></DropdownMenu>
      </div>;
    })}
  </div>;
}
