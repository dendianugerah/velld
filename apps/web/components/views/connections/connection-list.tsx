"use client";

import Link from "next/link";
import { useMemo, useState } from "react";
import { Search, Container, Plus } from "lucide-react";
import { Button } from "@cloudflare/kumo/components/button";
import { PageHeader } from "@/components/layout/page-header";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import { buttonVariants } from "@/components/ui/button";
import { useConnections } from "@/hooks/use-connections";
import { typeLabels } from "@/types/base";
import { ConnectionsTable } from "./connection-list/connections-table";
import { DeleteConnectionDialog } from "./connection-list/delete-connection-dialog";
import type { Connection } from "@/types/connection";

export function ConnectionsList() {
  const { connections, isLoading, error, refetch } = useConnections();
  const [search, setSearch] = useState("");
  const [engine, setEngine] = useState("all");
  const [deleting, setDeleting] = useState<Connection | null>(null);
  const filtered = useMemo(() => (connections ?? []).filter(connection =>
    (engine === "all" || connection.type === engine) &&
    [connection.name, connection.host, connection.database, connection.database_name, typeLabels[connection.type]].some(value => value?.toLowerCase().includes(search.trim().toLowerCase()))
  ), [connections, engine, search]);
  const scheduled = connections?.filter(connection => connection.backup_enabled).length ?? 0;

  return <>
    <PageHeader title="Connections" description="The databases you back up with Velld." action={<Link href="/connections/new" className={buttonVariants()}><Plus aria-hidden="true" />Add connection</Link>} />
    <div className="mb-7 flex flex-wrap items-center justify-between gap-3 text-xs text-muted-foreground">
      <p>{connections ? `${connections.length} ${connections.length === 1 ? "connection" : "connections"} · ${scheduled} scheduled` : "Your database connections"}</p>
      <Link href="/connections/discover" className="inline-flex min-h-9 items-center gap-2 rounded-md px-2 font-medium hover:bg-muted hover:text-foreground focus-visible:outline focus-visible:outline-2 focus-visible:outline-ring"><Container className="h-4 w-4" aria-hidden="true" />Discover Docker databases<span aria-hidden="true">→</span></Link>
    </div>
    <div className="mb-4 flex flex-col gap-3 sm:flex-row">
      <div className="relative min-w-0 flex-1"><Search className="pointer-events-none absolute start-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" aria-hidden="true" /><Input aria-label="Search connections" type="search" placeholder="Search by name or host…" value={search} onChange={event => setSearch(event.target.value)} className="ps-9" /></div>
      <select aria-label="Filter by database engine" value={engine} onChange={event => setEngine(event.target.value)} className="h-10 rounded-lg border border-input bg-card px-3 text-[13px] focus-visible:outline focus-visible:outline-2 focus-visible:outline-ring"><option value="all">All engines</option>{Object.entries(typeLabels).map(([value, label]) => <option value={value} key={value}>{label}</option>)}</select>
    </div>
    {error ? <div role="alert" className="rounded-lg border p-6 text-sm">Couldn’t load your connections.<div className="mt-4"><Button variant="secondary" onClick={() => refetch()}>Try again</Button></div></div>
      : isLoading ? <div className="space-y-6 rounded-lg border p-5" aria-label="Loading connections">{[1, 2, 3].map(key => <Skeleton key={key} className="h-12 w-full" />)}</div>
      : filtered.length ? <ConnectionsTable connections={filtered} onDelete={setDeleting} />
      : <div className="rounded-lg border border-dashed px-5 py-16 text-center"><h2 className="text-sm font-semibold">{connections?.length ? "No matching connections" : "Connect your first database"}</h2><p className="mx-auto mt-2 max-w-sm text-[13px] leading-relaxed text-muted-foreground">{connections?.length ? "Try another name or host, or clear your filters." : "Add a PostgreSQL, MySQL, or MongoDB database. Then run a backup or set a schedule."}</p><div className="mt-6">{connections?.length ? <Button variant="secondary" onClick={() => { setSearch(""); setEngine("all"); }}>Clear filters</Button> : <Link href="/connections/new" className={buttonVariants()}>Add connection</Link>}</div></div>}
    {(search || engine !== "all") && !error && !isLoading && <p role="status" className="mt-4 text-xs text-muted-foreground">{filtered.length} of {connections?.length ?? 0} connections</p>}
    {deleting && <DeleteConnectionDialog key={deleting.id} connection={deleting} onClose={() => setDeleting(null)} />}
  </>;
}
