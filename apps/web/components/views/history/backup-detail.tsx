"use client";

import { useRef, useState } from "react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { ArrowLeft, Download } from "lucide-react";
import { LayerCard } from "@cloudflare/kumo/components/layer-card";
import { Button as KumoButton } from "@cloudflare/kumo/components/button";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Skeleton } from "@/components/ui/skeleton";
import { BackupStatus, RestoreCheckStatus } from "@/components/ui/backup-status";
import { getBackup } from "@/lib/api/backups";
import { ApiError } from "@/lib/api-client";
import { formatSize } from "@/lib/helper";
import { useBackup } from "@/hooks/use-backup";
import { RestorePanel } from "./restore-panel";
import { BranchPanel } from "./branch-panel";
import { SchemaDiffPanel } from "./schema-diff-panel";
import { PG18Rehearsal } from "./pg18-rehearsal";
import { cn } from "@/lib/utils";
import { recoveryError } from "./recovery-error";

const dateLabel = (date?: string | null) => date ? new Date(date).toLocaleString(undefined, {dateStyle: "medium", timeStyle: "short"}) : "—";

export function BackupDetail({id}: {id: string}) {
  const router = useRouter();
  const query = useSearchParams();
  const from = query.get("from") || "";
  const back = `/history${from ? `?${new URLSearchParams(from)}` : ""}`;
  const result = useQuery({queryKey: ["backup", id], queryFn: () => getBackup(id), refetchInterval: 10000, retry: (count, error) => !(error instanceof ApiError && error.status === 404) && count < 2});
  const {downloadBackupFile, isDownloading, deleteBackupById, isDeleting} = useBackup(undefined, false);
  const [deleting, setDeleting] = useState(false);
  const [deleteError, setDeleteError] = useState("");
  const deleteTrigger = useRef<HTMLButtonElement>(null);
  const backup = result.data;

  if (result.isPending) return <div className="space-y-6"><Skeleton className="h-5 w-36" /><Skeleton className="h-24 w-full" /><Skeleton className="h-64 w-full" /></div>;
  if (!backup) return <div className="space-y-4"><Link href={back} className="text-xs text-muted-foreground">← Backup history</Link><h1 className="text-2xl font-semibold">{result.error instanceof ApiError && result.error.status === 404 ? "Backup not found" : "Couldn’t load this backup"}</h1><p className="text-sm text-muted-foreground">It may have been removed or you may not have access.</p><Button variant="outline" onClick={() => result.refetch()}>Try again</Button></div>;

  const complete = backup.status === "completed";
  const canBranch = ["postgresql", "mysql", "mariadb", "mongodb"].includes(backup.database_type);
  const canCompare = ["postgresql", "mysql"].includes(backup.database_type);
  const sections = [{id: "overview", label: "Overview"}, ...(complete ? [{id: "restore", label: "Restore"}, ...(canBranch ? [{id: "branch", label: "Branch"}] : []), ...(canCompare ? [{id: "compare", label: "Compare schema"}] : []), ...(backup.database_type === "postgresql" ? [{id: "pg18", label: "Test on PG 18"}] : [])] : [])];
  const view = sections.some(section => section.id === query.get("view")) ? query.get("view")! : "overview";
  const check = backup.verification_status;
  const filename = backup.path.split(/[\\/]/).pop() || "Backup file";
  const objectKey = backup.s3_object_key;

  function changeView(nextView: string) {
    const next = new URLSearchParams(query.toString());
    if (nextView === "overview") next.delete("view"); else next.set("view", nextView);
    window.history.pushState(null, "", `/history/${id}${next.size ? `?${next}` : ""}`);
  }

  return <div>
    <Link href={back} className="mb-7 inline-flex items-center gap-1.5 rounded text-xs text-muted-foreground hover:text-foreground"><ArrowLeft className="h-3.5 w-3.5" />Backup history</Link>
    <div className="mb-7 flex flex-col justify-between gap-4 sm:flex-row sm:items-center">
      <div className="min-w-0"><p className="mb-2 text-xs text-muted-foreground">{backup.connection_name || "Database backup"} / {backup.database_type}</p><h1 className="break-words text-2xl font-semibold tracking-[-0.035em]">{backup.database_name || backup.connection_name || "Backup details"}</h1><p className="mt-2 text-xs text-muted-foreground">Created {dateLabel(backup.created_at)}</p></div>
      <KumoButton variant="secondary" disabled={!complete || isDownloading} onClick={() => downloadBackupFile(backup)}><Download className="h-4 w-4" />{isDownloading ? "Downloading…" : "Download backup"}</KumoButton>
    </div>
    <LayerCard className="mb-7 rounded-lg shadow-none">
      <dl className="grid grid-cols-2 gap-5 p-4 sm:grid-cols-4 sm:p-5">
        <div><dt className="mb-2 text-xs text-muted-foreground">Backup status</dt><dd><BackupStatus status={backup.status} /></dd></div>
        <div><dt className="mb-2 text-xs text-muted-foreground">Restore check</dt><dd><RestoreCheckStatus status={check || "skipped"} /></dd></div>
        <div><dt className="mb-2 text-xs text-muted-foreground">File size</dt><dd className="text-sm font-medium tabular-nums">{formatSize(backup.size)}</dd></div>
        <div><dt className="mb-2 text-xs text-muted-foreground">Storage</dt><dd className="text-sm font-medium">{objectKey ? "Object storage" : "Local"}</dd></div>
      </dl>
    </LayerCard>
    {result.isError && <p role="alert" className="mb-4 text-xs text-destructive">Couldn’t refresh the latest status. <button className="underline" onClick={() => result.refetch()}>Try again</button></p>}
    <nav aria-label="Backup sections" className="mb-7 flex flex-wrap gap-1 border-b pb-2">{sections.map(section => <button key={section.id} onClick={() => changeView(section.id)} aria-current={view === section.id ? "page" : undefined} className={cn("rounded-md px-3 py-2 text-xs font-medium focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary", view === section.id ? "bg-muted text-foreground" : "text-muted-foreground hover:bg-muted/60")}>{section.label}</button>)}</nav>
    <section hidden={view !== "overview"} aria-label="Backup overview" className="space-y-8">
      <div><div className="mb-3 flex flex-wrap items-center gap-3"><h2 className="text-sm font-semibold">Restore check</h2><RestoreCheckStatus status={check || "skipped"} /></div>
        <div className={cn("rounded-lg border p-5", check === "failed" ? "border-destructive/20 bg-destructive/5" : check === "passed" ? "border-success/20 bg-success/5" : "bg-muted/40")}>
          <p className="text-sm font-medium">{check === "passed" ? "This backup restored successfully in a temporary database." : check === "failed" ? "The file was saved, but its restore check failed." : check === "pending" ? "A restore check is in progress." : "This backup has not been restore-checked."}</p>
          <p className="mt-2 text-xs leading-relaxed text-muted-foreground">{check === "failed" ? "Review the result before relying on this backup. A saved backup and a successful restore are separate outcomes." : check === "passed" ? "The check does not change the source database or verify every application query." : check === "pending" ? "This result updates automatically. The source database is unchanged." : "A completed backup does not guarantee a successful restore. Checks need a supported engine and Docker on the Velld host."}</p>
          {backup.verification_message && <div className="mt-4"><h3 className="mb-2 text-[11px] font-medium">Check output</h3><pre className="max-h-80 overflow-auto whitespace-pre-wrap break-words rounded-md border bg-card p-3 font-mono text-[11px] leading-relaxed">{backup.verification_message}</pre></div>}
          {backup.verified_at && <p className="mt-3 text-[11px] text-muted-foreground">Checked {dateLabel(backup.verified_at)}</p>}
        </div>
      </div>
      {!complete && <p className="rounded-lg border bg-muted/40 p-4 text-sm text-muted-foreground">{backup.status === "failed" ? "The backup did not complete. Download and recovery actions are unavailable; run a new backup from Connections." : "Download and recovery actions become available when the backup completes."}</p>}
      <div><h2 className="mb-4 text-sm font-semibold">Backup details</h2><dl className="divide-y rounded-lg border px-4 text-sm">{[["Connection", backup.connection_name || backup.connection_id], ["Database", backup.database_name || "—"], ["Engine", backup.database_type], ["Started", dateLabel(backup.started_time)], ["Completed", dateLabel(backup.completed_time)], ["Filename", filename], ["Local path", backup.path], ...(objectKey ? [["Object key", objectKey]] : []), ["Backup ID", backup.id]].map(([label, value]) => <div key={label} className="grid gap-1.5 py-3 sm:grid-cols-[140px_minmax(0,1fr)] sm:gap-4"><dt className="text-muted-foreground">{label}</dt><dd className="min-w-0 [overflow-wrap:anywhere]">{value}</dd></div>)}</dl></div>
      <div className="flex flex-wrap items-center justify-between gap-4 border-t pt-5"><div><h2 className="text-sm font-medium">Delete backup</h2><p className="mt-1 text-xs text-muted-foreground">Remove this file from local storage{objectKey ? " and object storage" : ""}.</p></div><Button ref={deleteTrigger} variant="outline" size="sm" className="text-destructive" onClick={() => {setDeleteError(""); setDeleting(true);}}>Delete backup</Button></div>
    </section>
    {complete && <><section hidden={view !== "restore"} aria-label="Restore backup"><RestorePanel backup={backup} /></section>{canBranch && <section hidden={view !== "branch"} aria-label="Branch from backup"><BranchPanel backup={backup} /></section>}{canCompare && <section hidden={view !== "compare"} aria-label="Compare schema"><SchemaDiffPanel backup={backup} /></section>}</>}
    {complete && backup.database_type === "postgresql" && view === "pg18" && <PG18Rehearsal backupId={id} onClose={() => changeView("overview")} />}
    <Dialog open={deleting} onOpenChange={open => {if (!isDeleting) setDeleting(open);}}><DialogContent onCloseAutoFocus={event => {event.preventDefault(); deleteTrigger.current?.focus();}}><DialogHeader><DialogTitle>Delete this backup?</DialogTitle><DialogDescription>The file for {backup.database_name || backup.connection_name} will be removed from local storage{objectKey ? " and object storage" : ""}. This cannot be undone.</DialogDescription></DialogHeader>{deleteError && <p role="alert" className="text-sm text-destructive">{deleteError}</p>}<DialogFooter><Button variant="outline" disabled={isDeleting} onClick={() => setDeleting(false)}>Cancel</Button><Button variant="destructive" disabled={isDeleting} onClick={() => deleteBackupById(id, {onSuccess: () => router.push(back), onError: error => setDeleteError(recoveryError(error, "Couldn’t delete this backup. Try again."))})}>{isDeleting ? "Deleting…" : "Delete backup"}</Button></DialogFooter></DialogContent></Dialog>
  </div>;
}
