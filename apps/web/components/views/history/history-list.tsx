"use client";

import { Card } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { EmptyState } from "@/components/ui/empty-state";
import { Database, Download, MoreVertical, History } from "lucide-react";
import { formatDistanceToNow, parseISO } from "date-fns";
import { useBackup } from "@/hooks/use-backup";
import { BackupList } from "@/types/backup";
import { statusColors } from "@/types/base";
import { HistoryListSkeleton } from "@/components/ui/skeleton/history-list";
import { calculateDuration, formatSize } from "@/lib/helper";
import { HistoryFilters } from "./history-filters";
import { CustomPagination } from "@/components/ui/custom-pagination";
import { useNotifications } from '@/hooks/use-notifications';
import { NotificationSidebar } from "./notification-sidebar";

export function HistoryList() {
  const { backups, isLoading, pagination, page, setPage, downloadBackupFile, isDownloading } = useBackup();
  const { notifications, isLoading: isLoadingNotifications, markNotificationsAsRead } = useNotifications();
  const totalPages = pagination?.total_pages || 1;

  return (
    <Card className="col-span-3 backdrop-blur-xl bg-card/50">
      <div className="flex flex-col h-full">
        <div className="p-6 pb-0">
          <div className="flex justify-between items-center mb-6">
            <h3 className="text-lg font-semibold">Recent Backups</h3>
          </div>
          <HistoryFilters />
        </div>
        <div className="flex flex-row p-6 pt-4 space-x-4">
          <div className="flex-1 space-y-4">
            {isLoading ? (
              <HistoryListSkeleton />
            ) : backups && backups.length > 0 ? (
              <>
                {backups.map((item: BackupList) => (
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
                            onClick={() => downloadBackupFile({ id: item.id, path: item.path })}
                            disabled={isDownloading}
                          >
                            <Download className="h-4 w-4" />
                          </Button>
                          <Button variant="ghost" size="icon">
                            <MoreVertical className="h-4 w-4" />
                          </Button>
                        </div>
                      </div>
                    </div>
                  </div>
                ))}
                
                {pagination && (
                  <div className="pt-4">
                    <CustomPagination
                      currentPage={page}
                      totalPages={totalPages}
                      onPageChange={setPage}
                    />
                  </div>
                )}
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

          <NotificationSidebar 
            notifications={notifications}
            isLoading={isLoadingNotifications}
            onMarkAsRead={markNotificationsAsRead}
          />
        </div>
      </div>
    </Card>
  );
}