import { Suspense } from "react";
import { AppShell } from "@/components/layout/app-shell";
import { BackupDetail } from "@/components/views/history/backup-detail";

export default async function BackupPage({params}: {params: Promise<{id: string}>}) {
  const {id} = await params;
  return <AppShell><Suspense fallback={<p className="text-sm text-muted-foreground">Loading backup…</p>}><BackupDetail id={id} /></Suspense></AppShell>;
}
