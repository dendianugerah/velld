'use client';

import { formatSize, getScheduleFrequency } from '@/lib/helper';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { 
  Table, 
  TableBody, 
  TableCell, 
  TableHead, 
  TableHeader, 
  TableRow 
} from '@/components/ui/table';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';
import { 
  Play, 
  CheckCircle2, 
  Pencil, 
  Trash2,
  Settings2,
  ArrowUpDown
} from 'lucide-react';
import { cn } from '@/lib/utils';
import type { Connection } from '@/types/connection';
import { typeLabels } from '@/types/base';
import { useBackup } from '@/hooks/use-backup';
import { useState } from 'react';

interface ConnectionsTableProps {
  connections: Connection[];
  onEdit: (connectionId: string) => void;
  onDelete: (connection: Connection) => void;
  onSchedule: (connectionId: string) => void;
}

type SortField = 'name' | 'type' | 'status' | 'size' | 'lastBackup';
type SortOrder = 'asc' | 'desc';

export function ConnectionsTable({
  connections,
  onEdit,
  onDelete,
  onSchedule,
}: ConnectionsTableProps) {
  const { createBackup, isCreating } = useBackup();
  const [sortField, setSortField] = useState<SortField>('name');
  const [sortOrder, setSortOrder] = useState<SortOrder>('asc');

  const handleSort = (field: SortField) => {
    if (sortField === field) {
      setSortOrder(sortOrder === 'asc' ? 'desc' : 'asc');
    } else {
      setSortField(field);
      setSortOrder('asc');
    }
  };

  const sortedConnections = [...connections].sort((a, b) => {
    let comparison = 0;
    
    switch (sortField) {
      case 'name':
        comparison = a.name.localeCompare(b.name);
        break;
      case 'type':
        comparison = a.type.localeCompare(b.type);
        break;
      case 'status':
        comparison = a.status.localeCompare(b.status);
        break;
      case 'size':
        comparison = (a.database_size || 0) - (b.database_size || 0);
        break;
      case 'lastBackup':
        const aTime = a.last_backup_time ? new Date(a.last_backup_time).getTime() : 0;
        const bTime = b.last_backup_time ? new Date(b.last_backup_time).getTime() : 0;
        comparison = aTime - bTime;
        break;
    }
    
    return sortOrder === 'asc' ? comparison : -comparison;
  });

  const SortButton = ({ field, children }: { field: SortField; children: React.ReactNode }) => (
    <button
      onClick={() => handleSort(field)}
      className="flex items-center gap-1 hover:text-foreground transition-colors"
    >
      {children}
      <ArrowUpDown className={cn(
        "h-3 w-3 transition-opacity",
        sortField === field ? "opacity-100" : "opacity-0 group-hover:opacity-50"
      )} />
    </button>
  );

  return (
    <div className="w-full">
      <TooltipProvider>
        <div className="overflow-hidden">
          <Table>
            <TableHeader className="border-b border-border/50">
              <TableRow className="hover:bg-transparent">
                <TableHead className="w-[280px] group font-semibold text-foreground/90 h-11">
                  <SortButton field="name">Name</SortButton>
                </TableHead>
                <TableHead className="group font-semibold text-foreground/90 h-11">
                  <SortButton field="type">Type</SortButton>
                </TableHead>
                <TableHead className="font-semibold text-foreground/90 h-11">Host</TableHead>
                <TableHead className="font-semibold text-foreground/90 h-11">Schedule</TableHead>
                <TableHead className="group font-semibold text-foreground/90 h-11">
                  <SortButton field="lastBackup">Last Backup</SortButton>
                </TableHead>
                <TableHead className="group text-right font-semibold text-foreground/90 h-11">
                  <SortButton field="size">Size</SortButton>
                </TableHead>
                <TableHead className="text-right font-semibold text-foreground/90 h-11">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {sortedConnections.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={9} className="h-32">
                    <div className="flex flex-col items-center justify-center gap-3 text-muted-foreground">
                      <div className="flex items-center justify-center w-12 h-12 rounded-full bg-muted">
                        <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4" />
                        </svg>
                      </div>
                      <div className="text-center">
                        <p className="text-sm font-medium">No connections found</p>
                        <p className="text-xs text-muted-foreground/60 mt-1">Try adjusting your search or add a new connection</p>
                      </div>
                    </div>
                  </TableCell>
                </TableRow>
              ) : (
                sortedConnections.map((connection) => {
                  const scheduleFrequency = getScheduleFrequency(connection.cron_schedule);
                  
                  return (
                    <TableRow 
                      key={connection.id} 
                      className="group border-b border-border/40 hover:bg-accent/20 transition-all duration-150"
                    >
                      <TableCell className="font-medium py-3.5">
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <span className="truncate block max-w-[260px] font-medium cursor-help text-foreground">
                              {connection.name}
                            </span>
                          </TooltipTrigger>
                          <TooltipContent side="top" align="start">
                            <p>{connection.name}</p>
                          </TooltipContent>
                        </Tooltip>
                      </TableCell>
                      <TableCell className="py-3.5">
                        <Badge variant="secondary" className="text-xs font-medium bg-accent/50 hover:bg-accent/70 border-0 px-2.5 py-0.5">
                          {typeLabels[connection.type]}
                        </Badge>
                      </TableCell>
                      <TableCell className="py-3.5">
                        <Tooltip>
                          <TooltipTrigger asChild>
                            <span className="text-sm text-muted-foreground/90 truncate block max-w-[180px] font-mono text-[13px] cursor-help">
                              {connection.host}:{connection.port}
                            </span>
                          </TooltipTrigger>
                          <TooltipContent side="top" align="start">
                            <p className="font-mono text-xs">{connection.host}:{connection.port}</p>
                          </TooltipContent>
                        </Tooltip>
                      </TableCell>
                      <TableCell className="py-3.5">
                        {connection.backup_enabled ? (
                          <div className="flex items-center gap-2">
                            <CheckCircle2 className="h-3.5 w-3.5 text-green-600 dark:text-green-500" />
                            <span className="text-sm font-medium text-foreground/80">
                              {scheduleFrequency || 'Enabled'}
                            </span>
                          </div>
                        ) : (
                          <span className="text-sm text-muted-foreground/60">—</span>
                        )}
                      </TableCell>
                      <TableCell className="py-3.5">
                        {connection.last_backup_time ? (
                          <span className="text-sm text-muted-foreground/90 font-medium">
                            {new Date(connection.last_backup_time).toLocaleDateString('en-US', {
                              month: 'short',
                              day: 'numeric',
                              hour: '2-digit',
                              minute: '2-digit'
                            })}
                          </span>
                        ) : (
                          <span className="text-sm text-muted-foreground/60">—</span>
                        )}
                      </TableCell>
                      <TableCell className="text-right py-3.5">
                        <span className="text-sm font-semibold text-foreground/90">
                          {formatSize(connection.database_size)}
                        </span>
                      </TableCell>
                      <TableCell className="text-right py-3.5">
                        <div className="flex items-center justify-end gap-0.5 opacity-0 group-hover:opacity-100 transition-opacity duration-200">
                          <Tooltip delayDuration={300}>
                            <TooltipTrigger asChild>
                              <Button
                                variant="ghost"
                                size="sm"
                                onClick={() => createBackup(connection.id)}
                                disabled={isCreating}
                                className="h-8 w-8 p-0 hover:bg-accent/80 transition-all duration-150 hover:scale-105"
                              >
                                <Play className="h-3.5 w-3.5" />
                              </Button>
                            </TooltipTrigger>
                            <TooltipContent side="top">
                              <p className="text-xs">Backup Now</p>
                            </TooltipContent>
                          </Tooltip>
                          
                          <Tooltip delayDuration={300}>
                            <TooltipTrigger asChild>
                              <Button
                                variant="ghost"
                                size="sm"
                                onClick={() => onSchedule(connection.id)}
                                className="h-8 w-8 p-0 hover:bg-accent/80 transition-all duration-150 hover:scale-105"
                              >
                                <Settings2 className="h-3.5 w-3.5" />
                              </Button>
                            </TooltipTrigger>
                            <TooltipContent side="top">
                              <p className="text-xs">Schedule Backup</p>
                            </TooltipContent>
                          </Tooltip>
                          
                          <Tooltip delayDuration={300}>
                            <TooltipTrigger asChild>
                              <Button
                                variant="ghost"
                                size="sm"
                                onClick={() => onEdit(connection.id)}
                                className="h-8 w-8 p-0 hover:bg-accent/80 transition-all duration-150 hover:scale-105"
                              >
                                <Pencil className="h-3.5 w-3.5" />
                              </Button>
                            </TooltipTrigger>
                            <TooltipContent side="top">
                              <p className="text-xs">Edit Connection</p>
                            </TooltipContent>
                          </Tooltip>
                          
                          <Tooltip delayDuration={300}>
                            <TooltipTrigger asChild>
                              <Button
                                variant="ghost"
                                size="sm"
                                onClick={() => onDelete(connection)}
                                className="h-8 w-8 p-0 text-destructive/80 hover:text-destructive hover:bg-destructive/10 transition-all duration-150 hover:scale-105"
                              >
                                <Trash2 className="h-3.5 w-3.5" />
                              </Button>
                            </TooltipTrigger>
                            <TooltipContent side="top">
                              <p className="text-xs">Delete Connection</p>
                            </TooltipContent>
                          </Tooltip>
                        </div>
                      </TableCell>
                    </TableRow>
                  );
                })
              )}
            </TableBody>
          </Table>
        </div>
      </TooltipProvider>
    </div>
  );
}
