'use client';

import { Card } from '@/components/ui/card';
import { useConnections } from "@/hooks/use-connections";
import { ConnectionListSkeleton } from "@/components/ui/skeleton/connection-list";
import { EmptyState } from '@/components/ui/empty-state';
import { useState } from 'react';
import { ConnectionListHeader } from './connection-list/header';
import { ConnectionCard } from './connection-list/connection-card';
import { BackupScheduleDialog } from './connection-list/backup-schedule-dialog';
import { SortBy } from '@/types/connection';
import { Database } from 'lucide-react';

export function ConnectionsList() {
  const { connections, isLoading } = useConnections();
  const [searchQuery, setSearchQuery] = useState('');
  const [sortBy, setSortBy] = useState<SortBy>('name');
  const [scheduleDialogConnection, setScheduleDialogConnection] = useState<string | null>(null);

  const filteredConnections = connections?.filter(connection => 
    connection.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    connection.host.toLowerCase().includes(searchQuery.toLowerCase())
  ).sort((a, b) => {
    switch (sortBy) {
      case 'name':
        return a.name.localeCompare(b.name);
      case 'status':
        return a.status.localeCompare(b.status);
      case 'type':
        return a.type.localeCompare(b.type);
      case 'lastBackup':
        return 0;
      default:
        return 0;
    }
  });

  return (
    <Card className="col-span-3 backdrop-blur-xl bg-card/50">
      <div className="p-6">
        <ConnectionListHeader 
          searchQuery={searchQuery}
          onSearchChange={setSearchQuery}
          sortBy={sortBy}
          onSortChange={setSortBy}
        />
        
        {isLoading ? (
          <ConnectionListSkeleton />
        ) : filteredConnections && filteredConnections.length > 0 ? (
          <div className="space-y-3">
            {filteredConnections.map((connection) => (
              <ConnectionCard
                key={connection.id}
                connection={connection}
                onSchedule={() => setScheduleDialogConnection(connection.id)}
              />
            ))}
          </div>
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
      </div>
    </Card>
  );
}