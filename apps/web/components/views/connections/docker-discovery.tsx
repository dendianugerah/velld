"use client";

import Link from "next/link";
import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { ArrowLeft } from "lucide-react";
import { Button } from "@cloudflare/kumo/components/button";
import { PageHeader } from "@/components/layout/page-header";
import { DatabaseMark } from "@/components/ui/database-mark";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import { discoverDockerDatabases, importDockerDatabase, type DockerDatabase } from "@/lib/api/connections";
import { connectionError } from "@/lib/connection-form";
import { typeLabels } from "@/types/base";
import { useToast } from "@/hooks/use-toast";

export function DockerDiscovery() {
  const [databaseNames, setDatabaseNames] = useState<Record<string, string>>({});
  const queryClient = useQueryClient();
  const { toast } = useToast();
  const discovery = useQuery({ queryKey: ["docker-databases"], queryFn: discoverDockerDatabases, retry: false });
  const add = useMutation({
    mutationFn: ({ container, database }: { container: DockerDatabase; database: string }) => importDockerDatabase(container.id, database),
    onSuccess: async () => {
      await Promise.all([queryClient.invalidateQueries({ queryKey: ["connections"] }), queryClient.invalidateQueries({ queryKey: ["docker-databases"] })]);
      toast({ title: "Docker database added", description: "Daily backups are scheduled with 30 days of retention." });
    },
  });

  return <>
    <Link href="/connections" className="mb-7 inline-flex min-h-8 items-center gap-2 rounded-sm text-xs text-muted-foreground hover:text-foreground focus-visible:outline focus-visible:outline-2 focus-visible:outline-ring"><ArrowLeft className="h-3.5 w-3.5" aria-hidden="true" />All connections</Link>
    <PageHeader title="Discover Docker databases" description="Find running databases reachable from this Velld instance." action={<Button variant="secondary" disabled={discovery.isFetching || add.isPending} onClick={() => discovery.refetch()}>{discovery.isFetching ? "Scanning…" : "Scan again"}</Button>} />
    <div className="max-w-[760px] space-y-6">
      <p className="text-[13px] leading-relaxed text-muted-foreground">Adding a database also schedules a daily backup at midnight, in the server’s time zone. Backups are kept for 30 days. You can change this on the connection’s schedule page.</p>
      {discovery.isFetching && <div aria-label="Scanning Docker" className="space-y-4"><Skeleton className="h-36 w-full" /><Skeleton className="h-36 w-full" /></div>}
      {discovery.error && <div role="alert" className="rounded-lg border p-5 text-sm text-destructive">{connectionError(discovery.error)}<p className="mt-3 text-muted-foreground">Check Docker access on your Velld server, then scan again.</p></div>}
      {!discovery.isFetching && discovery.data?.hint && <p className="rounded-lg border bg-muted/30 p-5 text-sm leading-relaxed">{discovery.data.hint}</p>}
      {!discovery.isFetching && discovery.data?.containers.length === 0 && <div className="rounded-lg border border-dashed p-8 text-center"><h2 className="text-sm font-semibold">No new database containers</h2><p className="mt-2 text-[13px] text-muted-foreground">Containers may already be connected, or none are reachable.</p><Link href="/connections/new" className="mt-5 inline-block text-sm font-medium underline underline-offset-4">Add a connection manually</Link></div>}
      {!discovery.isFetching && discovery.data?.containers.map(container => {
        const database = databaseNames[container.id] ?? container.database;
        return <form key={container.id} className="space-y-5 rounded-lg border p-5" onSubmit={event => { event.preventDefault(); add.mutate({ container, database: database.trim() }); }}>
          <div className="flex min-w-0 items-center gap-3"><DatabaseMark type={container.type} small /><div className="min-w-0"><h2 className="break-all text-sm font-medium">{container.name}</h2><p className="mt-1 break-all text-xs text-muted-foreground">{typeLabels[container.type] ?? container.type} · {container.host}:{container.port}</p></div></div>
          <div className="flex flex-col gap-3 sm:flex-row sm:items-end"><div className="min-w-0 flex-1 space-y-2"><Label htmlFor={`docker-db-${container.id}`}>Database name</Label><Input id={`docker-db-${container.id}`} required value={database} disabled={add.isPending} onChange={event => setDatabaseNames(previous => ({ ...previous, [container.id]: event.target.value }))} /></div><Button type="submit" disabled={add.isPending || !database.trim()}>{add.isPending && add.variables?.container.id === container.id ? "Adding…" : "Add and schedule"}</Button></div>
          {add.error && add.variables?.container.id === container.id && <p role="alert" className="text-sm text-destructive">{connectionError(add.error)}</p>}
        </form>;
      })}
    </div>
  </>;
}
