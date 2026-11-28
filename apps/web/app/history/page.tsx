import { AppShell } from "@/components/layout/app-shell";
import { PageHeader } from "@/components/layout/page-header";
import { HistoryList } from "@/components/views/history/history-list";
import { Suspense } from "react";

export default function HistoryPage() {
  return <AppShell>
    <PageHeader title="Backup history" />
    <Suspense fallback={<p className="text-sm text-muted-foreground">Loading backup history…</p>}><HistoryList /></Suspense>
  </AppShell>;
}