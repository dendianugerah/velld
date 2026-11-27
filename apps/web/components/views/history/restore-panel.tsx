"use client";

import { useState } from "react";
import { useMutation } from "@tanstack/react-query";
import { Button } from "@/components/ui/button";
import { useConnections } from "@/hooks/use-connections";
import { restoreBackup } from "@/lib/api/backups";
import type { Backup } from "@/types/backup";
import { recoveryError } from "./recovery-error";

export function RestorePanel({ backup }: { backup: Backup }) {
  const [target, setTarget] = useState("");
  const [confirmed, setConfirmed] = useState(false);
  const { connections, isLoading, error, refetch } = useConnections();
  const restore = useMutation({mutationFn: () => restoreBackup({backup_id: backup.id, connection_id: target})});
  const compatible = connections?.filter(connection => connection.type === backup.database_type) ?? [];
  const selected = compatible.find(connection => connection.id === target);

  return <div className="space-y-5">
    <div><h2 className="text-lg font-semibold tracking-tight">Restore to a database</h2><p className="mt-2 max-w-2xl text-sm leading-relaxed text-muted-foreground">Restore this file into an existing {backup.database_type} connection. Use a separate, empty database to check the data before switching over.</p></div>
    {error ? <div role="alert" className="space-y-2 text-sm"><p>Couldn’t load target connections.</p><Button variant="outline" size="sm" onClick={() => refetch()}>Try again</Button></div> : <label className="block max-w-lg space-y-2 text-sm"><span>Target connection</span><select value={target} disabled={isLoading || restore.isPending} onChange={event => {setTarget(event.target.value); setConfirmed(false); restore.reset();}} className="w-full rounded-md border bg-card px-3 py-2.5"><option value="">{isLoading ? "Loading connections…" : "Choose a connection"}</option>{compatible.map(connection => <option key={connection.id} value={connection.id}>{connection.name} · {connection.database || connection.host}</option>)}</select></label>}
    {!isLoading && !error && !compatible.length && <p className="text-sm text-muted-foreground">No compatible connections. <a className="text-primary underline" href="/connections">Add a {backup.database_type} connection</a> first.</p>}
    <div className="rounded-lg border border-warning/30 bg-warning/5 p-4 text-sm leading-relaxed"><p className="font-medium">This writes to the target database.</p><p className="mt-1 text-muted-foreground">Existing objects or data may be replaced. Restoring to the original connection changes your source database.</p></div>
    {selected && <label className="flex items-start gap-2.5 text-sm leading-relaxed"><input type="checkbox" className="mt-1 h-4 w-4 shrink-0 accent-primary" checked={confirmed} disabled={restore.isPending} onChange={event => setConfirmed(event.target.checked)} /><span>I understand this will write to <strong>{selected.name}</strong>{selected.id === backup.connection_id ? ", the original connection" : ""}.</span></label>}
    {restore.isError && <p role="alert" className="break-words text-sm text-destructive">{recoveryError(restore.error, "Couldn’t restore this backup. Try again.")}</p>}
    {restore.isSuccess && <p role="status" className="text-sm text-success">Backup restored to {selected?.name}.</p>}
    <Button disabled={!selected || !confirmed || restore.isPending || restore.isSuccess} onClick={() => restore.mutate()}>{restore.isPending ? "Restoring…" : "Restore backup"}</Button>
  </div>;
}
