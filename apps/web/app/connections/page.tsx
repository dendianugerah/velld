import { AppShell } from "@/components/layout/app-shell";
import { ConnectionsList } from "@/components/views/connections/connection-list";

export default function ConnectionsPage() {
  return <AppShell>
    <ConnectionsList />
  </AppShell>;
}