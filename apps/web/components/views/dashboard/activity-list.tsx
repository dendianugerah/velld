import { Card } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { HistoryListSkeleton } from "@/components/ui/skeleton/history-list";
import { ConnectionListSkeleton } from "@/components/ui/skeleton/connection-list";

import { Database, Clock, HardDrive, Calendar } from "lucide-react";

import { formatDistanceToNow, parseISO } from "date-fns";
import { formatSize, getScheduleFrequency } from "@/lib/helper";
import { cn } from "@/lib/utils";

import { BackupList } from "@/types/backup";
import { Connection } from "@/types/connection";
import { statusColors, StatusColor, typeLabels, DatabaseType } from "@/types/base";

import { useBackup } from "@/hooks/use-backup";
import { useConnections } from "@/hooks/use-connections";

export function ActivityList() {
  const { backups, isLoading: isLoadingBackups } = useBackup();
  const { connections, isLoading: isLoadingConnections } = useConnections();

  const renderBackupItem = (item: BackupList) => {
    const connection = connections?.find(c => c.id === item.connection_id);
    return (
      <div
        key={item.id}
        className="group p-4 rounded-lg border border-border/50 hover:border-border/80 transition-colors"
      >
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-4">
            <div className="p-2.5 rounded-md bg-primary/5">
              <Database className="h-5 w-5 text-primary" />
            </div>
            <div>
              <div className="flex items-center space-x-2">
                <p className="font-medium">{item.path.split('\\').pop()}</p>
                <Badge variant="outline" className="text-xs font-normal">
                  {typeLabels[item.database_type as DatabaseType]}
                </Badge>
              </div>
              <div className="flex items-center space-x-2 text-sm text-muted-foreground mt-1.5">
                <HardDrive className="h-3.5 w-3.5" />
                <span>{connection?.name}</span>
                <span className="text-muted-foreground/40">•</span>
                <span>{formatSize(item.size)}</span>
                <span className="text-muted-foreground/40">•</span>
                <span className="flex items-center">
                  <Clock className="h-3.5 w-3.5 mr-1" />
                  {formatDistanceToNow(parseISO(item.created_at), { addSuffix: true })}
                </span>
              </div>
            </div>
          </div>
          <Badge 
            variant="outline" 
            className={`text-xs ${statusColors[item.status as StatusColor]}`}
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
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-4">
          <div className={cn(
            "p-2.5 rounded-md",
            "bg-emerald-500/10"
          )}>
            <Database className="h-5 w-5 text-emerald-500" />
          </div>
          <div>
            <div className="flex items-center space-x-2">
              <p className="font-medium">{connection.name}</p>
              <Badge variant="outline" className="text-xs font-normal">
                {typeLabels[connection.type]}
              </Badge>
            </div>
            <div className="flex items-center space-x-2 text-sm text-muted-foreground mt-1.5">
              <span className="flex items-center">
                <Calendar className="h-3.5 w-3.5 mr-1" />
                {getScheduleFrequency(connection.cron_schedule)}
              </span>
              {connection.retention_days && (
                <>
                  <span className="text-muted-foreground/40">•</span>
                  <span>Kept for {connection.retention_days} days</span>
                </>
              )}
              {connection.last_backup_time && (
                <>
                  <span className="text-muted-foreground/40">•</span>
                  <span className="flex items-center">
                    <Clock className="h-3.5 w-3.5 mr-1" />
                    Last backup {formatDistanceToNow(parseISO(connection.last_backup_time), { addSuffix: true })}
                  </span>
                </>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );

  return (
    <Card className="col-span-3 backdrop-blur-xl bg-card/50">
      <div className="p-6">
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
                ) : (
                  backups?.map(renderBackupItem)
                )}
              </div>
            </ScrollArea>
          </TabsContent>

          <TabsContent value="scheduled" className="m-0">
            <ScrollArea className="h-[600px] lg:h-[700px]">
              <div className="space-y-2">
                {isLoadingConnections ? (
                  <ConnectionListSkeleton />
                ) : (
                  connections?.filter(c => c.backup_enabled).map(renderScheduledConnection)
                )}
              </div>
            </ScrollArea>
          </TabsContent>
        </Tabs>
      </div>
    </Card>
  );
} 