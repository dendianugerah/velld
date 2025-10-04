'use client';

import { formatSize, getScheduleFrequency } from '@/lib/helper';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';
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
    <TooltipProvider>
      <div className="p-4 rounded-lg border bg-card">
        {/* Mobile Layout */}
        <div className="flex md:hidden flex-col gap-4">
          <div className="flex items-start gap-3">
            <div className="p-2 rounded-md bg-primary/10 shrink-0">
              <Database className="h-5 w-5 text-primary" />
            </div>
            <div className="min-w-0 flex-1">
              <div className="flex items-center gap-2 flex-wrap mb-2">
                <Tooltip>
                  <TooltipTrigger asChild>
                    <h4 className="font-medium truncate cursor-help max-w-[200px]">{connection.name}</h4>
                  </TooltipTrigger>
                  <TooltipContent>
                    <p>{connection.name}</p>
                  </TooltipContent>
                </Tooltip>
                <Badge variant="outline" className="text-xs shrink-0">
                  {typeLabels[connection.type]}
                </Badge>
              </div>
              <div className="flex items-center gap-2 flex-wrap mb-2">
                <Badge
                  variant="outline"
                  className={cn("border shrink-0 text-xs", statusColors[connection.status])}
                >
                  {connection.status}
                </Badge>
                {connection.backup_enabled && (
                  <Badge variant="default" className="text-xs shrink-0">
                    <CheckCircle2 className="h-3 w-3 mr-1" />
                    Scheduled
                  </Badge>
                )}
              </div>
              <div className="flex flex-col gap-1 text-sm text-muted-foreground">
                <Tooltip>
                  <TooltipTrigger asChild>
                    <span className="truncate cursor-help">{connection.host}</span>
                  </TooltipTrigger>
                  <TooltipContent>
                    <p>{connection.host}</p>
                  </TooltipContent>
                </Tooltip>
                <div className="flex items-center gap-2 flex-wrap">
                  <span className="shrink-0">{formatSize(connection.database_size)}</span>
                  {connection.backup_enabled && scheduleFrequency && (
                    <>
                      <span className="shrink-0">•</span>
                      <span className="flex items-center shrink-0">
                        <Clock className="h-3 w-3 mr-1" />
                        {scheduleFrequency}
                      </span>
                    </>
                  )}
                </div>
                {connection.last_backup_time && (
                  <span className="text-xs">
                    Last backup: {new Date(connection.last_backup_time).toLocaleString()}
                  </span>
                )}
              </div>
            </div>
          </div>
          <div className="flex gap-2 w-full">
            <Button 
              variant="outline"
              size="sm"
              onClick={() => createBackup(connection.id)}
              disabled={isCreating}
              className="flex-1 text-xs"
            >
              <Play className="h-3 w-3 mr-1" />
              Backup
            </Button>
            <Button 
              variant="ghost" 
              size="sm" 
              className="flex-1 text-xs"
              onClick={onSchedule}
            >
              <Settings2 className="h-3 w-3 mr-1" />
              Schedule
            </Button>
          </div>
        </div>

        {/* Desktop Layout */}
        <div className="hidden md:flex items-center justify-between gap-4">
          <div className="flex items-center space-x-4 min-w-0 flex-1">
            <div className="p-2 rounded-md bg-primary/10 shrink-0">
              <Database className="h-5 w-5 text-primary" />
            </div>
            <div className="min-w-0 flex-1">
              <div className="flex items-center space-x-2 flex-wrap">
                <Tooltip>
                  <TooltipTrigger asChild>
                    <h4 className="font-medium truncate cursor-help max-w-[300px]">{connection.name}</h4>
                  </TooltipTrigger>
                  <TooltipContent>
                    <p>{connection.name}</p>
                  </TooltipContent>
                </Tooltip>
                <Badge variant="outline" className="text-xs shrink-0">
                  {typeLabels[connection.type]}
                </Badge>
                <Badge
                  variant="outline"
                  className={cn("border shrink-0", statusColors[connection.status])}
                >
                  {connection.status}
                </Badge>
                {connection.backup_enabled && (
                  <Badge variant="default" className="text-xs shrink-0">
                    <CheckCircle2 className="h-3 w-3 mr-1" />
                    Scheduled
                  </Badge>
                )}
              </div>
              <div className="flex items-center space-x-4 text-sm text-muted-foreground mt-1.5 flex-wrap">
                <Tooltip>
                  <TooltipTrigger asChild>
                    <span className="truncate max-w-xs cursor-help">{connection.host}</span>
                  </TooltipTrigger>
                  <TooltipContent>
                    <p>{connection.host}</p>
                  </TooltipContent>
                </Tooltip>
              <span className="shrink-0">•</span>
              <span className="shrink-0">{formatSize(connection.database_size)}</span>
              {connection.backup_enabled && scheduleFrequency && (
                <>
                  <span className="shrink-0">•</span>
                  <span className="flex items-center shrink-0">
                    <Clock className="h-3 w-3 mr-1" />
                    {scheduleFrequency}
                  </span>
                </>
              )}
              {connection.last_backup_time && (
                <>
                  <span className="shrink-0">•</span>
                  <span className="flex items-center shrink-0">
                    Last backup: {new Date(connection.last_backup_time).toLocaleString()}
                  </span>
                </>
              )}
            </div>
          </div>
        </div>
        
        <div className="flex items-center space-x-2 shrink-0">
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
    </TooltipProvider>
  );
}