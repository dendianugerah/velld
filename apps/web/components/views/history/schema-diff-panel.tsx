"use client";

import { useEffect, useRef, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Button } from "@/components/ui/button";
import { compareBackupSchema, getBackups } from "@/lib/api/backups";
import type { BackupList } from "@/types/backup";

export function SchemaDiffPanel({ backup }: { backup: BackupList }) {
  const [mode, setMode] = useState("live");
  const [page, setPage] = useState(1);
  const [target, setTarget] = useState("");
  const [sql, setSQL] = useState<string | null>(null);
  const [error, setError] = useState("");
  const [pending, setPending] = useState(false);
  const [copyStatus, setCopyStatus] = useState("");
  const request = useRef<AbortController | null>(null);
  const output = useRef<HTMLTextAreaElement>(null);
  const { data, isPending, isError, refetch } = useQuery({
    queryKey: ["backups", { page, limit: 25, status: "completed", engine: backup.database_type }],
    queryFn: () => getBackups({ page, limit: 25, status: "completed", engine: backup.database_type }),
    enabled: mode === "backup",
  });
  const candidates = (data?.data ?? []).filter(item => item.id !== backup.id &&
    item.status === "completed" && item.database_type === backup.database_type);
  const pages = data?.pagination?.total_pages ?? 1;

  useEffect(() => () => request.current?.abort(), []);

  function clearResult() {
    setSQL(null);
    setError("");
    setCopyStatus("");
  }

  async function compare() {
    const controller = new AbortController();
    request.current = controller;
    clearResult();
    setPending(true);
    try {
      const result = await compareBackupSchema(backup.id, mode === "live" ? { target_live: true } : { target_backup_id: target }, controller.signal);
      if (!controller.signal.aborted) setSQL(result);
    } catch (err) {
      if (controller.signal.aborted) return;
      let message = "Couldn’t compare the schemas. Try again.";
      if (err instanceof Error) {
        try {
          message = (JSON.parse(err.message) as { message?: string }).message || message;
        } catch {}
      }
      setError(message);
    } finally {
      if (!controller.signal.aborted) setPending(false);
    }
  }

  async function copy() {
    if (!sql) return;
    try {
      await navigator.clipboard.writeText(sql);
      setCopyStatus("Copied.");
    } catch {
      output.current?.select();
      setCopyStatus("Clipboard unavailable. Copy the selected text manually.");
    }
  }

  const selectClass = "w-full rounded-md border bg-background px-2 py-1.5 text-sm";
  function download() {
    if (!sql) return;
    const url = URL.createObjectURL(new Blob([sql], {type: "application/sql"}));
    const link = document.createElement("a");
    link.href = url;
    link.download = `schema-diff-${backup.id}.sql`;
    link.click();
    URL.revokeObjectURL(url);
  }
  return <div className="space-y-5">
      <div><h2 className="text-lg font-semibold tracking-tight">Compare schema</h2>
        <p className="mt-2 max-w-2xl text-sm leading-relaxed text-muted-foreground">Generate SQL to change this backup’s schema to match the target. Nothing is applied. Review the SQL before running it; it can drop tables and columns.</p>
      </div>
      <p className="text-xs text-muted-foreground">Compares tables, columns, indexes, and constraints. Does not compare data, views, routines, or all engine-specific options.</p>
      <label className="block max-w-lg space-y-2 text-sm">
        <span>Target</span>
        <select className={selectClass} value={mode} disabled={pending} onChange={event => { setMode(event.target.value); clearResult(); }}>
          <option value="live">Live database</option>
          <option value="backup">Another backup</option>
        </select>
      </label>
      {mode === "backup" && <div className="min-w-0 space-y-2">
        <label className="block space-y-1 text-sm">
          <span>Target backup</span>
          <select className={selectClass} value={target} disabled={pending || isPending || !candidates.length} onChange={event => { setTarget(event.target.value); clearResult(); }}>
            <option value="">{isPending ? "Loading backups…" : candidates.length ? "Choose a backup" : "No matching backups on this page"}</option>
            {candidates.map(item => <option key={item.id} value={item.id}>{item.database_name || item.connection_name} · {new Date(item.created_at).toLocaleString()}</option>)}
          </select>
        </label>
        {isError && <div role="alert" className="space-y-2 text-sm text-destructive"><p>Couldn’t load backups.</p><Button variant="outline" size="sm" onClick={() => refetch()}>Try again</Button></div>}
        {(pages > 1 || page > 1) && <div className="flex items-center gap-2 text-xs">
          <Button variant="outline" size="sm" disabled={pending || page <= 1} onClick={() => { setPage(page - 1); setTarget(""); clearResult(); }}>Previous</Button>
          <span>Page {page}</span>
          <Button variant="outline" size="sm" disabled={pending || page >= pages} onClick={() => { setPage(page + 1); setTarget(""); clearResult(); }}>Next</Button>
        </div>}
      </div>}
      {error && <p role="alert" className="text-sm text-destructive">{error}</p>}
      {sql === "" && <p role="status" className="text-sm">No differences in the supported schema objects.</p>}
      {sql && <label className="block min-w-0 space-y-2 text-sm">
        <span>Migration SQL</span>
        <textarea ref={output} readOnly value={sql} rows={16} spellCheck={false} className="w-full rounded-md border bg-muted p-4 font-mono text-xs leading-relaxed" />
      </label>}
      {copyStatus && <p role="status" className="text-xs text-muted-foreground">{copyStatus}</p>}
      <div className="flex flex-wrap gap-2">
        {sql && <Button variant="outline" onClick={copy}>Copy SQL</Button>}
        {sql && <Button variant="outline" onClick={download}>Download SQL</Button>}
        <Button onClick={compare} disabled={pending || (mode === "backup" && !candidates.some(item => item.id === target))}>
          {pending ? "Comparing…" : "Compare schemas"}
        </Button>
      </div>
    </div>;
}
