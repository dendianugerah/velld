"use client";

import { useEffect, useRef, useState } from "react";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { createBackupBranch, type BackupBranch } from "@/lib/api/backups";
import type { BackupList } from "@/types/backup";

export function BranchDialog({ backup, returnFocusTo, onClose }: { backup: BackupList; returnFocusTo: HTMLElement | null; onClose: () => void }) {
  const [branch, setBranch] = useState<BackupBranch | null>(null);
  const [error, setError] = useState("");
  const [pending, setPending] = useState(false);
  const [copyStatus, setCopyStatus] = useState("");
  const request = useRef<AbortController | null>(null);
  const connection = useRef<HTMLInputElement>(null);

  useEffect(() => () => request.current?.abort(), []);

  async function create() {
    const controller = new AbortController();
    request.current = controller;
    setError("");
    setPending(true);
    try {
      setBranch(await createBackupBranch(backup.id, controller.signal));
    } catch (err) {
      if (controller.signal.aborted) return;
      let message = "Couldn’t create the backup branch. Try again.";
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
    if (!branch) return;
    try {
      await navigator.clipboard.writeText(branch.connection_string);
      setCopyStatus("Copied");
    } catch {
      connection.current?.select();
      setCopyStatus("Clipboard unavailable. Copy the selected text manually.");
    }
  }

  return <Dialog open onOpenChange={open => { if (!open) onClose(); }}>
    <DialogContent onCloseAutoFocus={event => { if (returnFocusTo?.isConnected) { event.preventDefault(); returnFocusTo.focus(); } }}>
      <DialogHeader>
        <DialogTitle>Branch from backup</DialogTitle>
        <DialogDescription>Create a temporary database copy. Common email, name, and phone fields are masked based on their field names. Other data is unchanged.</DialogDescription>
      </DialogHeader>
      {error && <p role="alert" className="text-sm text-destructive">{error}</p>}
      {branch && <div className="min-w-0 space-y-2">
        <label htmlFor="branch-connection" className="text-sm font-medium">Connection string</label>
        <Input id="branch-connection" ref={connection} className="font-mono text-xs" readOnly value={branch.connection_string} />
        <p className="text-sm text-muted-foreground">Expires {new Date(branch.expires_at).toLocaleString()}. Connect on the Docker host or forward the port over SSH.</p>
        <p role="status" className="text-sm text-muted-foreground">{copyStatus}</p>
      </div>}
      <DialogFooter>
        <Button variant="outline" onClick={onClose}>{branch ? "Close" : "Cancel"}</Button>
        {branch ? <Button onClick={copy}>Copy connection string</Button> : <Button onClick={create} disabled={pending}>{pending ? "Creating branch…" : "Create masked branch"}</Button>}
      </DialogFooter>
    </DialogContent>
  </Dialog>;
}
