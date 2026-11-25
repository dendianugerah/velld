import { Workspace } from "@/components/layout/workspace";
import { HistoryList } from "@/components/views/history/history-list";

export default function HistoryPage() {
  return (
    <Workspace title="Backup History" description="Find, download, or restore your backups.">
      <HistoryList />
    </Workspace>
  );
}