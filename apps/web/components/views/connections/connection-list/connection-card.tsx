'use client';

import { formatSize, getScheduleFrequency } from '@/lib/helper';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Database, Clock, Settings2, Play, CheckCircle2 } from 'lucide-react';
import { cn } from '@/lib/utils';
import type { Connection } from '@/types/connection';
import { statusColors, typeLabels } from '@/types/base';
import { useBackup } from '@/hooks/use-backup';

interface ConnectionCardProps {
  connection: Connection;
  onSchedule: () => void;
}

export function ConnectionCard({
  connection,
  onSchedule,
}: ConnectionCardProps) {
  const { createBackup, isCreating } = useBackup();
  const scheduleFrequency = getScheduleFrequency(connection.cron_schedule);

  return (
    <div className="p-4 rounded-lg border bg-card">
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-4">
          <div className="p-2 rounded-md bg-primary/10">
            <Database className="h-5 w-5 text-primary" />
          </div>
          <div>
            <div className="flex items-center space-x-2">
              <h4 className="font-medium">{connection.name}</h4>
              <Badge variant="outline" className="text-xs">
                {typeLabels[connection.type]}
              </Badge>
              <Badge
                variant="outline"
                className={cn("border", statusColors[connection.status])}
              >
                {connection.status}
              </Badge>
              {connection.backup_enabled && (
                <Badge variant="default" className="text-xs">
                  <CheckCircle2 className="h-3 w-3 mr-1" />
                  Scheduled
                </Badge>
              )}
            </div>
            <div className="flex items-center space-x-4 text-sm text-muted-foreground mt-1.5">
              <span>{connection.host}</span>
              <span>•</span>
              <span>{formatSize(connection.database_size)}</span>
              {connection.backup_enabled && scheduleFrequency && (
                <>
                  <span>•</span>
                  <span className="flex items-center">
                    <Clock className="h-3 w-3 mr-1" />
                    {scheduleFrequency}
                  </span>
                </>
              )}
              {connection.last_backup_time && (
                <>
                  <span>•</span>
                  <span className="flex items-center">
                    Last backup: {new Date(connection.last_backup_time).toLocaleString()}
                  </span>
                </>
              )}
            </div>
          </div>
        </div>
        
        <div className="flex items-center space-x-2">
          <Button 
            variant="outline"
            size="sm"
            onClick={() => createBackup(connection.id)}
            disabled={isCreating}
            className="space-x-1"
          >
            <Play className="h-3 w-3" />
            <span>Backup Now</span>
          </Button>
          
          <Button 
            variant="ghost" 
            size="sm" 
            className="space-x-1"
            onClick={onSchedule}
          >
            <Settings2 className="h-3 w-3" />
            <span>Schedule</span>
          </Button>
        </div>
      </div>
    </div>
  );
}