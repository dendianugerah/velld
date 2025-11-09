"use client";

import { Card } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { EmptyState } from "@/components/ui/empty-state";
import { Database, Download, History, GitCompare, RefreshCw, RotateCcw } from "lucide-react";
import { formatDistanceToNow, parseISO, subDays, isAfter } from "date-fns";
import { useBackup } from "@/hooks/use-backup";
import { BackupList } from "@/types/backup";
import { statusColors } from "@/types/base";
import { HistoryListSkeleton } from "@/components/ui/skeleton/history-list";
import { calculateDuration, formatSize } from "@/lib/helper";
import { HistoryFilters } from "./history-filters";
import { CustomPagination } from "@/components/ui/custom-pagination";
import { useNotifications } from '@/hooks/use-notifications';
import { NotificationSidebar } from "./notification-sidebar";
import { BackupCompareDialog } from "./backup-compare-dialog";
import { RestoreDialog } from "./restore-dialog";
import { BackupDetailsSheet } from "./backup-details-sheet";
import { Tooltip, TooltipTrigger, TooltipContent, TooltipProvider } from "@/components/ui/tooltip";
import { useState, useMemo } from "react";
import { useIsFetching } from "@tanstack/react-query";

export function HistoryList() {
  const { backups, isLoading, pagination, page, setPage, downloadBackupFile, isDownloading, search, setSearch } = useBackup();
  const { notifications, isLoading: isLoadingNotifications, markNotificationsAsRead } = useNotifications();
  const [compareDialogOpen, setCompareDialogOpen] = useState(false);
  const [selectedBackupForCompare, setSelectedBackupForCompare] = useState<BackupList | undefined>();
  const [restoreDialogOpen, setRestoreDialogOpen] = useState(false);
  const [selectedBackupForRestore, setSelectedBackupForRestore] = useState<BackupList | null>(null);
  const [detailsSheetOpen, setDetailsSheetOpen] = useState(false);
  const [selectedBackupForDetails, setSelectedBackupForDetails] = useState<BackupList | null>(null);
  const isFetchingBackups = useIsFetching({ queryKey: ['backups'] });
  
  const [dateRange, setDateRange] = useState("all");
  const [status, setStatus] = useState("all");
  const [databaseType, setDatabaseType] = useState("all");

  const totalPages = pagination?.total_pages || 1;

  const handleCompare = (backup?: BackupList) => {
    if (backup) {
      setSelectedBackupForCompare(backup);
    } else if (filteredBackups && filteredBackups.length > 0) {
      setSelectedBackupForCompare(filteredBackups[0]);
    }
    setCompareDialogOpen(true);
  };

  const handleRestore = (backup: BackupList) => {
    setSelectedBackupForRestore(backup);
    setRestoreDialogOpen(true);
  };

  const handleResetFilters = () => {
    setSearch("");
    setDateRange("all");
    setStatus("all");
    setDatabaseType("all");
  };

  const filteredBackups = useMemo(() => {
    if (!backups) return [];

    return backups.filter((backup) => {
      const searchLower = search.toLowerCase();
      const matchesSearch = !search || 
        backup.path.toLowerCase().includes(searchLower) ||
        backup.database_type.toLowerCase().includes(searchLower);

      let matchesDateRange = true;
      if (dateRange !== "all") {
        const backupDate = parseISO(backup.created_at);
        const now = new Date();
        
        switch (dateRange) {
          case "24hours":
            matchesDateRange = isAfter(backupDate, subDays(now, 1));
            break;
          case "7days":
            matchesDateRange = isAfter(backupDate, subDays(now, 7));
            break;
          case "30days":
            matchesDateRange = isAfter(backupDate, subDays(now, 30));
            break;
        }
      }

      const matchesStatus = status === "all" || backup.status.toLowerCase() === status.toLowerCase();
      const matchesDatabaseType = databaseType === "all" || backup.database_type.toLowerCase() === databaseType.toLowerCase();

      return matchesSearch && matchesDateRange && matchesStatus && matchesDatabaseType;
    });
  }, [backups, search, dateRange, status, databaseType]);

  return (
    <Card className="col-span-3 bg-card border">
      <div className="flex flex-col h-full">
        <div className="p-4 sm:p-6 border-b">
          <div className="space-y-4">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
              <h3 className="text-lg font-semibold">Recent Backups</h3>
              {isFetchingBackups > 0 && !isLoading && (
                <div className="flex items-center gap-2 text-xs text-muted-foreground">
                  <RefreshCw className="h-3 w-3 animate-spin" />
                  <span>Updating...</span>
                </div>
              )}
            </div>
            <HistoryFilters 
              onCompare={filteredBackups && filteredBackups.length > 1 ? () => handleCompare() : undefined}
              search={search}
              onSearchChange={setSearch}
              dateRange={dateRange}
              onDateRangeChange={setDateRange}
              status={status}
              onStatusChange={setStatus}
              databaseType={databaseType}
              onDatabaseTypeChange={setDatabaseType}
              onReset={handleResetFilters}
              isLoading={isLoading}
            />
          </div>
        </div>
        <div className="flex flex-col lg:flex-row p-4 sm:p-6 gap-4 flex-1">
          <div className="flex-1 flex flex-col min-w-0">
            <div className="flex-1 space-y-4">
              {isLoading ? (
                <HistoryListSkeleton />
              ) : filteredBackups && filteredBackups.length > 0 ? (
                <>
                  {filteredBackups.map((item: BackupList) => (
                    <div
                      key={item.id}
                      className="p-4 rounded-lg bg-background/50 hover:bg-background/60 transition-colors border"
                    >
                      {/* Mobile Layout */}
                      <div className="flex md:hidden flex-col gap-3">
                        <div className="flex items-start gap-3">
                          <div className="p-2 rounded-md bg-primary/10 shrink-0">
                            <Database className="h-5 w-5 text-primary" />
                          </div>
                          <div className="min-w-0 flex-1">
                            <h4 className="font-medium truncate">{item.path.split('\\').pop()}</h4>
                            <div className="flex items-center gap-2 mt-1 flex-wrap">
                              <Badge variant="secondary" className="text-xs">
                                {item.database_type}
                              </Badge>
                              <Badge
                                variant="secondary"
                                className={statusColors[item.status as keyof typeof statusColors]}
                              >
                                {item.status}
                              </Badge>
                            </div>
                            <p className="text-xs text-muted-foreground mt-2">
                              {formatDistanceToNow(parseISO(item.created_at), { addSuffix: true })}
                            </p>
                            <p className="text-xs text-muted-foreground">
                              {formatSize(item.size)} • {calculateDuration(item.started_time, item.completed_time)}
                            </p>
                          </div>
                        </div>
                        <div className="flex flex-col gap-2">
                          <div className="flex gap-2">
                            <Button 
                              variant="outline" 
                              size="sm"
                              onClick={() => handleRestore(item)}
                              className="flex-1"
                            >
                              <RotateCcw className="h-4 w-4 mr-1" />
                              Restore
                            </Button>
                            <Button 
                              variant="outline" 
                              size="sm" 
                              onClick={() => downloadBackupFile({ id: item.id, path: item.path })}
                              disabled={isDownloading}
                              className="flex-1"
                            >
                              <Download className="h-4 w-4 mr-1" />
                              Download
                            </Button>
                          </div>
                          <Button 
                            variant="outline" 
                            size="sm"
                            onClick={() => handleCompare(item)}
                            className="w-full"
                          >
                            <GitCompare className="h-4 w-4 mr-1" />
                            Compare with Another
                          </Button>
                        </div>
                      </div>

                      {/* Desktop Layout */}
                      <div className="hidden md:flex items-center justify-between">
                        <div className="flex items-center space-x-4 min-w-0 flex-1">
                          <div className="p-2 rounded-md bg-primary/10 shrink-0">
                            <Database className="h-5 w-5 text-primary" />
                          </div>
                          <div className="min-w-0 flex-1">
                            <div className="flex items-center space-x-2">
                              <h4 className="font-medium truncate">{item.path.split('\\').pop()}</h4>
                              <Badge variant="secondary" className="text-xs shrink-0">
                                {item.database_type}
                              </Badge>
                            </div>
                            <p className="text-sm text-muted-foreground mt-1">
                              {formatDistanceToNow(parseISO(item.created_at), { addSuffix: true })} | {formatSize(item.size)}
                            </p>
                          </div>
                        </div>
                        <div className="flex items-center space-x-4 shrink-0">
                          <div className="text-right">
                            <Badge
                              variant="secondary"
                              className={statusColors[item.status as keyof typeof statusColors]}
                            >
                              {item.status}
                            </Badge>
                            <p className="text-sm text-muted-foreground mt-1">
                              {calculateDuration(item.started_time, item.completed_time)}
                            </p>
                          </div>
                          <div className="flex space-x-2">
                            <TooltipProvider>
                              <Tooltip>
                                <TooltipTrigger asChild>
                                  <Button
                                    variant="ghost"
                                    size="icon"
                                    onClick={() => handleRestore(item)}
                                    >
                                    <RotateCcw className="h-4 w-4" />
                                  </Button>
                                </TooltipTrigger>
                                <TooltipContent side="top">
                                  <p className="text-xs">Restore this backup</p>
                                </TooltipContent>
                              </Tooltip>

                              <Tooltip>
                                <TooltipTrigger asChild>
                                  <Button
                                    variant="ghost"
                                    size="icon"
                                    onClick={() => handleCompare(item)}
                                    >
                                    <GitCompare className="h-4 w-4" />
                                  </Button>
                                </TooltipTrigger>
                                <TooltipContent side="top">
                                  <p className="text-xs">Compare with another backup</p>
                                </TooltipContent>
                              </Tooltip>

                              <Tooltip>
                                <TooltipTrigger asChild>
                                  <Button
                                    variant="ghost"
                                    size="icon"
                                    onClick={() => downloadBackupFile({ id: item.id, path: item.path })}
                                    disabled={isDownloading}
                                    >
                                    <Download className="h-4 w-4" />
                                  </Button>
                                </TooltipTrigger>
                                <TooltipContent side="top">
                                  <p className="text-xs">Download backup</p>
                                </TooltipContent>
                              </Tooltip>
                            </TooltipProvider>
                          </div>
                        </div>
                      </div>
                    </div>
                  ))}
                </>
              ) : (
                <EmptyState
                  icon={History}
                  title="No backup history"
                  description="Your backup history will appear here once you start creating backups."
                  variant="minimal"
                />
              )}
            </div>
            
            {pagination && filteredBackups && filteredBackups.length > 0 && (
              <div className="pt-4 sm:pt-6 flex justify-center sm:justify-end border-t mt-4">
                <CustomPagination
                  currentPage={page}
                  totalPages={totalPages}
                  onPageChange={setPage}
                />
              </div>
            )}
          </div>

          <NotificationSidebar 
            notifications={notifications}
            isLoading={isLoadingNotifications}
            onMarkAsRead={markNotificationsAsRead}
            onNotificationClick={(notification) => {
              // For failed backups, create a backup object from notification metadata
              // since failed backups don't exist in the database
              if (notification.metadata) {
                const failedBackup: BackupList = {
                  id: notification.id,
                  connection_id: notification.metadata.connection_id || '',
                  database_name: notification.metadata.database_name || '',
                  database_type: notification.metadata.database_type || '',
                  path: '',
                  size: 0,
                  status: 'failed',
                  started_time: notification.created_at,
                  completed_time: notification.created_at,
                  created_at: notification.created_at,
                  scheduled_time: '',
                  updated_at: notification.created_at,
                };
                
                setSelectedBackupForDetails(failedBackup);
                setDetailsSheetOpen(true);
              }
            }}
          />
        </div>
      </div>

      <BackupCompareDialog
        open={compareDialogOpen}
        onClose={() => {
          setCompareDialogOpen(false);
          setSelectedBackupForCompare(undefined);
        }}
        backups={filteredBackups || []}
        selectedBackup={selectedBackupForCompare}
      />

      <RestoreDialog
        backup={selectedBackupForRestore}
        open={restoreDialogOpen}
        onOpenChange={setRestoreDialogOpen}
      />

      <BackupDetailsSheet
        backup={selectedBackupForDetails}
        open={detailsSheetOpen}
        onClose={() => {
          setDetailsSheetOpen(false);
          setSelectedBackupForDetails(null);
        }}
      />
    </Card>
  );
}