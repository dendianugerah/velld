'use client';

import { useConnections } from "@/hooks/use-connections";
import { ConnectionsTableSkeleton } from "@/components/ui/skeleton/connections-table";
import { EmptyState } from '@/components/ui/empty-state';
import { useState } from 'react';
import { ConnectionListHeader } from './connection-list/header';
import { ConnectionsTable } from './connection-list/connections-table';
import { BackupScheduleDialog } from './connection-list/backup-schedule-dialog';
import { EditConnectionDialog } from './connection-list/edit-connection-dialog';
import { DeleteConnectionDialog } from './connection-list/delete-connection-dialog';
import { UpgradeDialog } from './connection-list/upgrade-dialog';
import { DiscoverDatabasesModal } from './discover-databases-modal';
import { AddConnectionDialog } from './add-connection-dialog';
import { DockerDiscoveryDialog } from './docker-discovery-dialog';
import { Database } from 'lucide-react';
import type { Connection } from '@/types/connection';

export function ConnectionsList() {
  const { connections, isLoading, error, refetch } = useConnections();
  const [searchQuery, setSearchQuery] = useState('');
  const [scheduleDialogConnection, setScheduleDialogConnection] = useState<string | null>(null);
  const [scheduleTrigger, setScheduleTrigger] = useState<HTMLElement | null>(null);
  const [editDialogConnection, setEditDialogConnection] = useState<string | null>(null);
  const [deleteDialogConnection, setDeleteDialogConnection] = useState<Connection | null>(null);
  const [upgradeDialogConnection, setUpgradeDialogConnection] = useState<Connection | null>(null);
  const [discoverConnectionId, setDiscoverConnectionId] = useState<string | null>(null);

  const filteredConnections = connections?.filter(connection => 
    connection.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    connection.host.toLowerCase().includes(searchQuery.toLowerCase()) ||
    (connection.database || '').toLowerCase().includes(searchQuery.toLowerCase())
  );

  return (
    <div className="col-span-3 space-y-3">
      <ConnectionListHeader 
        searchQuery={searchQuery}
        onSearchChange={setSearchQuery}
        actionButton={<div className="grid gap-2 sm:flex sm:flex-wrap sm:justify-end"><DockerDiscoveryDialog /><AddConnectionDialog /></div>}
        resultCount={searchQuery ? filteredConnections?.length : undefined}
        totalCount={connections?.length}
      />
      
      {error ? (
        <div role="alert" className="rounded-xl border p-6 text-sm">Couldn’t load connections. <button className="ml-2 underline underline-offset-4" onClick={() => refetch()}>Try again</button></div>
      ) : isLoading ? (
        <ConnectionsTableSkeleton />
      ) : filteredConnections && filteredConnections.length > 0 ? (
        <ConnectionsTable
          connections={filteredConnections}
          onEdit={setEditDialogConnection}
          onDelete={setDeleteDialogConnection}
          onSchedule={(id, trigger) => { setScheduleTrigger(trigger ?? null); setScheduleDialogConnection(id); }}
          onDiscover={(connection) => setDiscoverConnectionId(connection.id)}
          onUpgrade={setUpgradeDialogConnection}
        />
      ) : (
        <EmptyState
          icon={Database}
          title={searchQuery ? "No matching connections" : "No database connections"}
          description={searchQuery ? "Try a different name, host, or database." : "Add a connection above, then run a backup or set a schedule."}
          variant="minimal"
        />
      )}

      <BackupScheduleDialog
        connectionId={scheduleDialogConnection}
        connection={connections?.find(c => c.id === scheduleDialogConnection)}
        returnFocusTo={scheduleTrigger}
        onClose={() => setScheduleDialogConnection(null)}
      />

      <EditConnectionDialog
        connectionId={editDialogConnection}
        onClose={() => setEditDialogConnection(null)}
      />

      <DeleteConnectionDialog
        connection={deleteDialogConnection}
        onClose={() => setDeleteDialogConnection(null)}
      />

      {upgradeDialogConnection && (
        <UpgradeDialog
          connection={upgradeDialogConnection}
          onClose={() => setUpgradeDialogConnection(null)}
        />
      )}

      {discoverConnectionId && (
        <DiscoverDatabasesModal
          connectionId={discoverConnectionId}
          open={!!discoverConnectionId}
          onOpenChange={(open) => !open && setDiscoverConnectionId(null)}
        />
      )}
    </div>
  );
}