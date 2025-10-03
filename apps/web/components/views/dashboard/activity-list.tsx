import { Card } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { HistoryListSkeleton } from "@/components/ui/skeleton/history-list";
import { ConnectionListSkeleton } from "@/components/ui/skeleton/connection-list";
import { EmptyState } from "@/components/ui/empty-state";

import { Database, Clock, HardDrive, Calendar, Activity, Plus, Timer } from "lucide-react";

import { formatDistanceToNow, parseISO } from "date-fns";
import { formatSize, getScheduleFrequency } from "@/lib/helper";
import { cn } from "@/lib/utils";
import { useRouter } from "next/navigation";

import { BackupList } from "@/types/backup";
import { Connection } from "@/types/connection";
import { statusColors, StatusColor, typeLabels, DatabaseType } from "@/types/base";

import { useBackup } from "@/hooks/use-backup";
import { useConnections } from "@/hooks/use-connections";

export function ActivityList() {
  const { backups, isLoading: isLoadingBackups } = useBackup();
  const { connections, isLoading: isLoadingConnections } = useConnections();
  const router = useRouter();

  const renderBackupItem = (item: BackupList) => {
    const connection = connections?.find(c => c.id === item.connection_id);
    return (
      <div
        key={item.id}
        className="group p-4 rounded-lg border border-border/50 hover:border-border/80 transition-colors"
      >
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 sm:gap-0">
          <div className="flex items-start sm:items-center space-x-4">
            <div className={cn(
              "p-2.5 rounded-md shrink-0",
              item.status === 'completed' ? "bg-emerald-500/10" : 
              item.status === 'failed' ? "bg-red-500/10" : "bg-primary/5"
            )}>
              <Database className={cn(
                "h-5 w-5",
                item.status === 'completed' ? "text-emerald-500" :
                item.status === 'failed' ? "text-red-500" : "text-primary"
              )} />
            </div>
            <div className="min-w-0 flex-1">
              <div className="flex flex-wrap items-center gap-2">
                <p className="font-medium truncate">{item.path.split('\\').pop()}</p>
                <div className="flex items-center gap-2">
                  <Badge variant="outline" className="text-xs font-normal">
                    {typeLabels[item.database_type as DatabaseType]}
                  </Badge>
                  <Badge 
                    variant="outline" 
                    className={cn(
                      "text-xs sm:hidden",
                      statusColors[item.status as StatusColor]
                    )}
                  >
                    {item.status}
                  </Badge>
                </div>
              </div>
              <div className="flex flex-wrap items-center gap-2 text-sm text-muted-foreground mt-1.5">
                <span className="items-center gap-1 hidden sm:flex">
                  <HardDrive className="h-3.5 w-3.5" />
                  <span className="truncate max-w-[120px] sm:max-w-none">{connection?.name}</span>
                </span>
                <span className="hidden sm:inline text-muted-foreground/40">•</span>
                <span className="hidden sm:block">{formatSize(item.size)}</span>
                <span className="hidden sm:inline text-muted-foreground/40">•</span>
                <span className="flex items-center">
                  <Clock className="h-3.5 w-3.5 mr-1" />
                  <span className="truncate max-w-[120px] sm:max-w-none">
                    {formatDistanceToNow(parseISO(item.created_at), { addSuffix: true })}
                  </span>
                </span>
              </div>
            </div>
          </div>
          <Badge 
            variant="outline" 
            className={cn(
              "hidden sm:inline-flex text-xs",
              statusColors[item.status as StatusColor]
            )}
          >
            {item.status}
          </Badge>
        </div>
      </div>
    );
  };

  const renderScheduledConnection = (connection: Connection) => (
    <div
      key={connection.id}
      className="group p-4 rounded-lg border border-border/50 hover:border-border/80 transition-colors"
    >
      <div className="flex flex-col sm:flex-row sm:items-center gap-4 sm:gap-4">
        <div className="flex items-start sm:items-center space-x-4">
          <div className={cn(
            "p-2.5 rounded-md shrink-0",
            "bg-emerald-500/10"
          )}>
            <Database className="h-5 w-5 text-emerald-500" />
          </div>
          <div className="min-w-0">
            <div className="flex flex-wrap items-center gap-2">
              <p className="font-medium truncate">{connection.name}</p>
              <Badge variant="outline" className="text-xs font-normal">
                {typeLabels[connection.type]}
              </Badge>
            </div>
            <div className="flex flex-wrap items-center gap-2 text-sm text-muted-foreground mt-1.5">
              <span className="flex items-center">
                <Calendar className="h-3.5 w-3.5 mr-1" />
                {getScheduleFrequency(connection.cron_schedule)}
              </span>
              {connection.retention_days && (
                <div className="flex items-center gap-2">
                  <span className="hidden sm:inline text-muted-foreground/40">•</span>
                  <span>{connection.retention_days} days retention</span>
                </div>
              )}
              {connection.last_backup_time && (
                <div className="flex items-center gap-2">
                  <span className="hidden sm:inline text-muted-foreground/40">•</span>
                  <span className="flex items-center">
                    <Clock className="h-3.5 w-3.5 mr-1" />
                    <span className="truncate">
                      Last backup {formatDistanceToNow(parseISO(connection.last_backup_time), { addSuffix: true })}
                    </span>
                  </span>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );

  return (
    <Card className="col-span-3 bg-card border">
      <div className="p-4 sm:p-6">
        <Tabs defaultValue="recent" className="w-full">
          <div className="flex justify-between items-center mb-4">
            <TabsList className="h-8">
              <TabsTrigger value="recent" className="text-xs">Recent Activity</TabsTrigger>
              <TabsTrigger value="scheduled" className="text-xs">Scheduled</TabsTrigger>
            </TabsList>
          </div>

          <TabsContent value="recent" className="m-0">
            <ScrollArea className="h-[600px] lg:h-[700px]">
              <div className="space-y-2">
                {isLoadingBackups ? (
                  <HistoryListSkeleton />
                ) : backups && backups.length > 0 ? (
                  backups.map(renderBackupItem)
                ) : (
                  <EmptyState
                    icon={Activity}
                    title="No recent activity"
                    description="Set up database connections and run your first backup to see activity."
                    variant="minimal"
                  />
                )}
              </div>
            </ScrollArea>
          </TabsContent>

          <TabsContent value="scheduled" className="m-0">
            <ScrollArea className="h-[600px] lg:h-[700px]">
              <div className="space-y-2">
                {isLoadingConnections ? (
                  <ConnectionListSkeleton />
                ) : connections && connections.filter(c => c.backup_enabled).length > 0 ? (
                  connections.filter(c => c.backup_enabled).map(renderScheduledConnection)
                ) : (
                  <EmptyState
                    icon={Timer}
                    title="No scheduled backups"
                    description="Automate your database backups by setting up schedules. This ensures your data is regularly backed up without manual intervention."
                    action={{
                      label: "Manage Connections",
                      onClick: () => router.push('/connections'),
                      variant: "outline"
                    }}
                  />
                )}
              </div>
            </ScrollArea>
          </TabsContent>
        </Tabs>
      </div>
    </Card>
  );
}