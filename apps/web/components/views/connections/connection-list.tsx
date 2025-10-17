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
import { AddConnectionDialog } from './add-connection-dialog';
import { Database } from 'lucide-react';
import type { Connection } from '@/types/connection';

export function ConnectionsList() {
  const { connections, isLoading } = useConnections();
  const [searchQuery, setSearchQuery] = useState('');
  const [scheduleDialogConnection, setScheduleDialogConnection] = useState<string | null>(null);
  const [editDialogConnection, setEditDialogConnection] = useState<string | null>(null);
  const [deleteDialogConnection, setDeleteDialogConnection] = useState<Connection | null>(null);

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
        />
      ) : (
        <EmptyState
          icon={Database}
          title="No database connections"
          description="Get started by adding your first database connection."
          variant="minimal"
        />
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
    </div>
  );
}