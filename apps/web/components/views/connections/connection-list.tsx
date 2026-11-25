'use client';

import { useConnections } from "@/hooks/use-connections";
import { ConnectionsTableSkeleton } from "@/components/ui/skeleton/connections-table";
import { useState } from 'react';
import { ConnectionListHeader } from './connection-list/header';
import { ConnectionsTable } from './connection-list/connections-table';
import { BackupScheduleDialog } from './connection-list/backup-schedule-dialog';
import { EditConnectionDialog } from './connection-list/edit-connection-dialog';
import { DeleteConnectionDialog } from './connection-list/delete-connection-dialog';
import { DiscoverDatabasesModal } from './discover-databases-modal';
import { AddConnectionDialog } from './add-connection-dialog';
import type { Connection } from '@/types/connection';

export function ConnectionsList() {
  const { connections, isLoading, error } = useConnections();
  const [searchQuery, setSearchQuery] = useState('');
  const [scheduleDialogConnection, setScheduleDialogConnection] = useState<string | null>(null);
  const [editDialogConnection, setEditDialogConnection] = useState<string | null>(null);
  const [deleteDialogConnection, setDeleteDialogConnection] = useState<Connection | null>(null);
  const [discoverConnectionId, setDiscoverConnectionId] = useState<string | null>(null);

  const filteredConnections = connections?.filter(connection => 
    connection.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    connection.host.toLowerCase().includes(searchQuery.toLowerCase()) ||
    (connection.database || '').toLowerCase().includes(searchQuery.toLowerCase())
  );

  return (
    <div className="space-y-4">
      <ConnectionListHeader 
        searchQuery={searchQuery}
        onSearchChange={setSearchQuery}
        actionButton={<AddConnectionDialog />}
        resultCount={searchQuery ? filteredConnections?.length : undefined}
        totalCount={connections?.length}
      />
      
      {isLoading ? (
        <ConnectionsTableSkeleton />
      ) : filteredConnections && filteredConnections.length > 0 ? (
        <ConnectionsTable
          connections={filteredConnections}
          onEdit={setEditDialogConnection}
          onDelete={setDeleteDialogConnection}
          onSchedule={setScheduleDialogConnection}
          onDiscover={(connection) => setDiscoverConnectionId(connection.id)}
        />
      ) : (
        <div className="rounded-xl border px-5 py-7">
          <p className="text-sm font-medium">{error ? "Couldn’t load connections" : searchQuery ? "No matching connections" : "No database connections yet"}</p>
          <p className="mt-1 text-sm text-muted-foreground">{error ? "Refresh the page to try again." : searchQuery ? "Try a different name or host." : "Add your first connection to run a backup."}</p>
        </div>
      )}

      <BackupScheduleDialog
        connectionId={scheduleDialogConnection}
        connection={connections?.find(c => c.id === scheduleDialogConnection)}
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