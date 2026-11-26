"use client";

import { useEffect, useRef, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { compareBackupSchema, getBackups } from "@/lib/api/backups";
import type { BackupList } from "@/types/backup";

export function SchemaDiffDialog({ backup, returnFocusTo, onClose }: { backup: BackupList; returnFocusTo: HTMLElement | null; onClose: () => void }) {
  const [mode, setMode] = useState("live");
  const [page, setPage] = useState(1);
  const [target, setTarget] = useState("");
  const [sql, setSQL] = useState<string | null>(null);
  const [error, setError] = useState("");
  const [pending, setPending] = useState(false);
  const [copyStatus, setCopyStatus] = useState("");
  const request = useRef<AbortController | null>(null);
  const output = useRef<HTMLTextAreaElement>(null);
  const { data, isPending, isError } = useQuery({
    queryKey: ["backups", { page, limit: 25 }],
    queryFn: () => getBackups({ page, limit: 25 }),
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
  return <Dialog open onOpenChange={open => { if (!open) onClose(); }}>
    <DialogContent onCloseAutoFocus={event => { if (returnFocusTo?.isConnected) { event.preventDefault(); returnFocusTo.focus(); } }}>
      <DialogHeader>
        <DialogTitle>Compare schema</DialogTitle>
        <DialogDescription>Generate SQL to change this backup’s schema to match the target. Nothing is applied. Review the SQL before running it; it can drop tables and columns.</DialogDescription>
      </DialogHeader>
      <p className="text-xs text-muted-foreground">Compares tables, columns, indexes, and constraints. Does not compare data, views, routines, or all engine-specific options.</p>
      <label className="space-y-1 text-sm">
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
            {candidates.map(item => <option key={item.id} value={item.id}>{item.path.split(/[\\/]/).pop() || item.id}</option>)}
          </select>
        </label>
        {isError && <p role="alert" className="text-sm text-destructive">Couldn’t load backups.</p>}
        {(pages > 1 || page > 1) && <div className="flex items-center gap-2 text-xs">
          <Button variant="outline" size="sm" disabled={pending || page <= 1} onClick={() => { setPage(page - 1); setTarget(""); clearResult(); }}>Previous</Button>
          <span>Page {page}</span>
          <Button variant="outline" size="sm" disabled={pending || page >= pages} onClick={() => { setPage(page + 1); setTarget(""); clearResult(); }}>Next</Button>
        </div>}
      </div>}
      {error && <p role="alert" className="text-sm text-destructive">{error}</p>}
      {sql === "" && <p role="status" className="text-sm">No differences in the supported schema objects.</p>}
      {sql && <label className="min-w-0 space-y-1 text-sm">
        <span>Migration SQL</span>
        <textarea ref={output} readOnly value={sql} rows={8} spellCheck={false} className="w-full rounded-md border bg-muted p-2 font-mono text-xs" />
      </label>}
      {copyStatus && <p role="status" className="text-xs text-muted-foreground">{copyStatus}</p>}
      <DialogFooter>
        {sql && <Button variant="outline" onClick={copy}>Copy SQL</Button>}
        <Button onClick={compare} disabled={pending || (mode === "backup" && !candidates.some(item => item.id === target))}>
          {pending ? "Comparing…" : "Compare schemas"}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>;
}
