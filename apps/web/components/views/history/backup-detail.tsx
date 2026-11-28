"use client";

import { useRef, useState } from "react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { useQuery } from "@tanstack/react-query";
import { ArrowLeft, ChevronRight, Download } from "lucide-react";
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
  if (!backup) {
    const notFound = result.error instanceof ApiError && result.error.status === 404;
    return <div className="space-y-4"><Link href={back} className="text-xs text-muted-foreground">← Backup history</Link><h1 className="text-2xl font-semibold">{notFound ? "Backup not found" : "Couldn’t load this backup"}</h1>{!notFound && <Button variant="outline" onClick={() => result.refetch()}>Try again</Button>}</div>;
  }

  const complete = backup.status === "completed";
  const canBranch = ["postgresql", "mysql", "mariadb", "mongodb"].includes(backup.database_type);
  const canCompare = ["postgresql", "mysql"].includes(backup.database_type);
  const sections = [{id: "overview", label: "Overview"}, ...(complete ? [{id: "restore", label: "Restore"}, ...(canBranch ? [{id: "branch", label: "Branch"}] : []), ...(canCompare ? [{id: "compare", label: "Compare schema"}] : []), ...(backup.database_type === "postgresql" ? [{id: "pg18", label: "Test on PG 18"}] : [])] : [])];
  const view = sections.some(section => section.id === query.get("view")) ? query.get("view")! : "overview";
  const check = backup.verification_status;
  const objectKey = backup.s3_object_key;
  const fileLocations = [
    ...(backup.path ? [["Local path", backup.path]] : []),
    ...(objectKey ? [["Object key", objectKey]] : []),
  ];

  function changeView(nextView: string) {
    const next = new URLSearchParams(query.toString());
    if (nextView === "overview") next.delete("view"); else next.set("view", nextView);
    window.history.pushState(null, "", `/history/${id}${next.size ? `?${next}` : ""}`);
  }

  return <div>
    <Link href={back} className="mb-5 inline-flex items-center gap-1.5 rounded text-xs text-muted-foreground hover:text-foreground"><ArrowLeft aria-hidden="true" className="h-3.5 w-3.5" />Backup history</Link>
    <div className="mb-5 flex flex-col justify-between gap-3 sm:flex-row sm:items-center">
      <div className="min-w-0"><h1 className="break-words text-2xl font-semibold tracking-[-0.035em]">{backup.database_name || backup.connection_name || "Backup"}</h1><p className="mt-1.5 break-words text-xs text-muted-foreground">{[backup.database_name && backup.connection_name !== backup.database_name ? backup.connection_name : "", backup.database_type, dateLabel(backup.created_at)].filter(Boolean).join(" · ")}</p></div>
      {complete && <KumoButton variant="secondary" disabled={isDownloading} onClick={() => downloadBackupFile(backup)}><Download aria-hidden="true" className="h-4 w-4" />{isDownloading ? "Downloading…" : "Download backup"}</KumoButton>}
    </div>
    {result.isError && <p role="alert" className="mb-4 text-xs text-destructive">Couldn’t refresh. <button className="underline" onClick={() => result.refetch()}>Try again</button></p>}
    {sections.length > 1 && <nav aria-label="Backup sections" className="mb-5 flex flex-wrap gap-1 border-b pb-2">{sections.map(section => <button key={section.id} onClick={() => changeView(section.id)} aria-current={view === section.id ? "page" : undefined} className={cn("rounded-md px-3 py-2 text-xs font-medium focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary", view === section.id ? "bg-muted text-foreground" : "text-muted-foreground hover:bg-muted/60")}>{section.label}</button>)}</nav>}
    <section hidden={view !== "overview"} aria-label="Backup overview" className="space-y-4">
      <LayerCard className="overflow-hidden rounded-lg shadow-none">
        <dl className="grid grid-cols-2 gap-4 p-4 sm:grid-cols-4">
          <div><dt className="mb-1.5 text-xs text-muted-foreground">Backup status</dt><dd><BackupStatus status={backup.status} /></dd></div>
          <div><dt className="mb-1.5 text-xs text-muted-foreground">Restore check</dt><dd><RestoreCheckStatus status={check || "skipped"} /></dd></div>
          <div><dt className="mb-1.5 text-xs text-muted-foreground">File size</dt><dd className="text-sm font-medium tabular-nums">{formatSize(backup.size)}</dd></div>
          <div><dt className="mb-1.5 text-xs text-muted-foreground">Storage</dt><dd className="text-sm font-medium">{objectKey ? "Object storage" : "Local"}</dd></div>
        </dl>
        {fileLocations.length > 0 && <dl aria-label="Backup files" className="divide-y border-t px-4 text-sm">{fileLocations.map(([label, value]) => <div key={label} className="grid gap-1.5 py-3 sm:grid-cols-[120px_minmax(0,1fr)] sm:gap-4"><dt className="text-muted-foreground">{label}</dt><dd className="min-w-0 break-all font-mono text-xs">{value}</dd></div>)}</dl>}
        {(backup.verification_message || backup.verified_at) && <details className="group border-t">
          <summary className="flex cursor-pointer list-none items-center gap-1.5 px-4 py-3 text-sm font-medium hover:bg-muted/50 focus-visible:outline focus-visible:outline-2 focus-visible:-outline-offset-2 focus-visible:outline-primary"><ChevronRight aria-hidden="true" className="size-4 shrink-0 text-muted-foreground group-open:rotate-90" />Restore check output</summary>
          <div className="space-y-2 px-4 pb-4">
            {backup.verified_at && <p className="text-xs text-muted-foreground">Checked {dateLabel(backup.verified_at)}</p>}
            {backup.verification_message && <pre className="max-h-80 overflow-auto whitespace-pre-wrap break-words rounded-md border bg-muted/30 p-3 font-mono text-xs">{backup.verification_message}</pre>}
          </div>
        </details>}
      </LayerCard>
      <div className="flex justify-end"><Button ref={deleteTrigger} variant="outline" size="sm" className="text-destructive" onClick={() => {setDeleteError(""); setDeleting(true);}}>Delete backup</Button></div>
    </section>
    {complete && <><section hidden={view !== "restore"} aria-label="Restore backup"><RestorePanel backup={backup} /></section>{canBranch && <section hidden={view !== "branch"} aria-label="Branch from backup"><BranchPanel backup={backup} /></section>}{canCompare && <section hidden={view !== "compare"} aria-label="Compare schema"><SchemaDiffPanel backup={backup} /></section>}</>}
    {complete && backup.database_type === "postgresql" && view === "pg18" && <PG18Rehearsal backupId={id} />}
    <Dialog open={deleting} onOpenChange={open => {if (!isDeleting) setDeleting(open);}}><DialogContent onCloseAutoFocus={event => {event.preventDefault(); deleteTrigger.current?.focus();}}><DialogHeader><DialogTitle>Delete this backup?</DialogTitle><DialogDescription>{objectKey ? "This also deletes the object storage copy." : "This can’t be undone."}</DialogDescription></DialogHeader>{deleteError && <p role="alert" className="text-sm text-destructive">{deleteError}</p>}<DialogFooter><Button variant="outline" disabled={isDeleting} onClick={() => setDeleting(false)}>Cancel</Button><Button variant="destructive" disabled={isDeleting} onClick={() => deleteBackupById(id, {onSuccess: () => router.push(back), onError: error => setDeleteError(recoveryError(error, "Couldn’t delete this backup. Try again."))})}>{isDeleting ? "Deleting…" : "Delete backup"}</Button></DialogFooter></DialogContent></Dialog>
  </div>;
}
