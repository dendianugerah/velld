"use client";

import { Card } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { EmptyState } from "@/components/ui/empty-state";
import { Database, Download, MoreVertical, History, GitCompare } from "lucide-react";
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
import { useState, useMemo } from "react";

export function HistoryList() {
  const { backups, isLoading, pagination, page, setPage, downloadBackupFile, isDownloading, search, setSearch } = useBackup();
  const { notifications, isLoading: isLoadingNotifications, markNotificationsAsRead } = useNotifications();
  const [compareDialogOpen, setCompareDialogOpen] = useState(false);
  const [selectedBackupForCompare, setSelectedBackupForCompare] = useState<BackupList | undefined>();
  
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
        <div className="p-6 border-b">
          <div className="space-y-4">
            <h3 className="text-lg font-semibold">Recent Backups</h3>
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
        <div className="flex flex-row p-6 space-x-4 flex-1">
          <div className="flex-1 flex flex-col">
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
                      <div className="flex items-center justify-between">
                        <div className="flex items-center space-x-4">
                          <div className="p-2 rounded-md bg-primary/10">
                            <Database className="h-5 w-5 text-primary" />
                          </div>
                          <div>
                            <div className="flex items-center space-x-2">
                              <h4 className="font-medium">{item.path.split('\\').pop()}</h4>
                              <Badge variant="secondary" className="text-xs">
                                {item.database_type}
                              </Badge>
                            </div>
                            <p className="text-sm text-muted-foreground mt-1">
                              {formatDistanceToNow(parseISO(item.created_at), { addSuffix: true })} | {formatSize(item.size)}
                            </p>
                          </div>
                        </div>
                        <div className="flex items-center space-x-4">
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
                            <Button 
                              variant="ghost" 
                              size="icon"
                              onClick={() => handleCompare(item)}
                              title="Compare with another backup"
                            >
                              <GitCompare className="h-4 w-4" />
                            </Button>
                            <Button 
                              variant="ghost" 
                              size="icon" 
                              onClick={() => downloadBackupFile({ id: item.id, path: item.path })}
                              disabled={isDownloading}
                              title="Download backup"
                            >
                              <Download className="h-4 w-4" />
                            </Button>
                            {/* <Button variant="ghost" size="icon" title="More options">
                              <MoreVertical className="h-4 w-4" />
                            </Button> */}
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
              <div className="pt-6 flex justify-end border-t mt-4">
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
    </Card>
  );
}