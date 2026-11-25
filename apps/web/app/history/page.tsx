import { AppShell } from "@/components/layout/app-shell";
import { PageHeader } from "@/components/layout/page-header";
import { HistoryList } from "@/components/views/history/history-list";

export default function HistoryPage() {
  return <AppShell>
    <PageHeader title="Backup history" description="Find a backup, download a copy, or restore a database." />
    <HistoryList />
  </AppShell>;
}