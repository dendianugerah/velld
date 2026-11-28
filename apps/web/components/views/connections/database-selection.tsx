"use client";

import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Button } from "@cloudflare/kumo/components/button";
import { Input } from "@/components/ui/input";
import { discoverDatabases, updateSelectedDatabases } from "@/lib/api/connections";
import { connectionError } from "@/lib/connection-form";
import { useToast } from "@/hooks/use-toast";
import type { Connection } from "@/types/connection";

export function DatabaseSelection({ connection }: { connection: Connection }) {
  const [editing, setEditing] = useState(false);
  const [selected, setSelected] = useState(connection.selected_databases ?? []);
  const [search, setSearch] = useState("");
  const queryClient = useQueryClient();
  const { toast } = useToast();
  const discovery = useQuery({ queryKey: ["connection-databases", connection.id], queryFn: () => discoverDatabases(connection.id), enabled: editing, retry: false });
  const save = useMutation({
    mutationFn: () => updateSelectedDatabases(connection.id, selected),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["connection", connection.id] });
      queryClient.invalidateQueries({ queryKey: ["connections"] });
      toast({ title: "Database selection saved" }); setEditing(false);
    },
  });
  const databases = discovery.data?.databases ?? [];
  return <section aria-labelledby="databases-heading" className="space-y-5">
    <div className="flex flex-wrap items-start justify-between gap-4"><h2 id="databases-heading" className="text-sm font-semibold">Databases to back up</h2>{!editing && <Button variant="secondary" onClick={() => { setSelected(connection.selected_databases ?? []); setEditing(true); }}>Choose databases</Button>}</div>
    {!editing ? <div className="rounded-lg border px-5 py-4">{connection.selected_databases?.length ? <ul className="flex flex-wrap gap-2">{connection.selected_databases.map(name => <li key={name} className="max-w-full break-all rounded-md bg-muted px-2.5 py-1.5 font-mono text-xs">{name}</li>)}</ul> : <p className="break-all text-[13px] text-muted-foreground">Default database{connection.database_name && <>: <span className="font-mono text-foreground">{connection.database_name}</span></>}</p>}</div> : <div className="space-y-4 rounded-lg border p-4 sm:p-5">
      {discovery.isFetching && <p role="status" className="text-sm text-muted-foreground">Discovering databases…</p>}
      {discovery.error && <div role="alert" className="space-y-3 text-sm"><p className="text-destructive">{connectionError(discovery.error)}</p><Button variant="secondary" onClick={() => discovery.refetch()}>Try again</Button></div>}
      {!discovery.error && discovery.data && <>
        <Input aria-label="Search databases" type="search" placeholder="Find a database…" value={search} onChange={event => setSearch(event.target.value)} />
        <div className="max-h-72 overflow-y-auto rounded-md border">{databases.filter(name => name.toLowerCase().includes(search.toLowerCase())).map(name => <label key={name} className="flex min-h-11 cursor-pointer items-center gap-3 border-b px-3 py-3 last:border-0 hover:bg-muted/50"><input type="checkbox" className="h-4 w-4 shrink-0 accent-primary" checked={selected.includes(name)} disabled={save.isPending} onChange={event => setSelected(previous => event.target.checked ? [...previous, name] : previous.filter(value => value !== name))} /><span className="break-all font-mono text-xs">{name}</span></label>)}{databases.length === 0 && <p className="p-4 text-sm text-muted-foreground">No accessible databases found. Check the database user’s permissions and try again.</p>}{databases.length > 0 && !databases.some(name => name.toLowerCase().includes(search.toLowerCase())) && <p className="p-4 text-sm text-muted-foreground">No matching databases. Try another name.</p>}</div>
        <p className="text-xs text-muted-foreground">{selected.length} selected. With none selected, Velld uses the connection’s default database.</p>
      </>}
      {save.error && <p role="alert" className="text-sm text-destructive">{connectionError(save.error)}</p>}
      <div className="flex flex-wrap gap-3"><Button disabled={save.isPending || discovery.isFetching || !discovery.data || !!discovery.error} onClick={() => save.mutate()}>{save.isPending ? "Saving…" : "Save selection"}</Button><Button variant="secondary" disabled={save.isPending} onClick={() => setEditing(false)}>Cancel</Button></div>
    </div>}
  </section>;
}
