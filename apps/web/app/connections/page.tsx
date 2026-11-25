import { Workspace } from "@/components/layout/workspace";
import { ConnectionsList } from "@/components/views/connections/connection-list";

export default function ConnectionsPage() {
  return (
    <Workspace title="Connections" description="Manage your databases and set up automatic backups.">
      <ConnectionsList />
    </Workspace>
  );
}