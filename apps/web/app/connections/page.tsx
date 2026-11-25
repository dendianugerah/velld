import { AppShell } from "@/components/layout/app-shell";
import { PageHeader } from "@/components/layout/page-header";
import { ConnectionsList } from "@/components/views/connections/connection-list";

export default function ConnectionsPage() {
  return <AppShell>
    <PageHeader title="Connections" description="Your databases, their schedules, and the next backup." />
    <ConnectionsList />
  </AppShell>;
}