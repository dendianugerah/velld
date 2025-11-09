'use client';

import { Database, Calendar, Clock, AlertCircle } from 'lucide-react';
import { Badge } from '@/components/ui/badge';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Separator } from '@/components/ui/separator';
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from '@/components/ui/sheet';
import type { Backup } from '@/types/backup';
import { useNotifications } from '@/hooks/use-notifications';
import { useConnections } from '@/hooks/use-connections';
import { useEffect, useState } from 'react';

interface BackupDetailsSheetProps {
  backup: Backup | null;
  open: boolean;
  onClose: () => void;
}

export function BackupDetailsSheet({ backup, open, onClose }: BackupDetailsSheetProps) {
  const { notifications } = useNotifications();
  const { connections } = useConnections();
  const [relatedNotification, setRelatedNotification] = useState<any>(null);
  const [connectionName, setConnectionName] = useState<string>('Unknown Connection');

  useEffect(() => {
    if (backup && connections) {
      const connection = connections.find((c: any) => c.id === backup.connection_id);
      setConnectionName(connection?.name || 'Unknown Connection');
    }
  }, [backup, connections]);

  useEffect(() => {
    if (backup && backup.status === 'failed' && notifications) {
      // Find notification related to this backup
      const notification = notifications.find(
        (n: any) => 
          n.type === 'backup_failed' && 
          n.metadata?.connection_id === backup.connection_id &&
          Math.abs(new Date(n.created_at).getTime() - new Date(backup.created_at).getTime()) < 5000
      );
      setRelatedNotification(notification);
    }
  }, [backup, notifications]);

  if (!backup) return null;

  return (
    <Sheet open={open} onOpenChange={(isOpen) => {
      if (!isOpen) {
        onClose();
      }
    }}>
      <SheetContent className="w-full sm:max-w-[500px]">
        <SheetHeader>
          <div className="flex items-center gap-3">
            <div className={`flex items-center justify-center w-10 h-10 rounded-lg ${
              backup.status === 'failed' ? 'bg-destructive/10' : 'bg-primary/10'
            }`}>
              <Database className={`h-5 w-5 ${
                backup.status === 'failed' ? 'text-destructive' : 'text-primary'
              }`} />
            </div>
            <div>
              <SheetTitle>Backup Error Details</SheetTitle>
              <SheetDescription>{connectionName}</SheetDescription>
            </div>
          </div>
        </SheetHeader>

        <ScrollArea className="h-[calc(100vh-120px)] mt-6">
          <div className="space-y-6 pr-4">
            {/* Status */}
            <div>
              <label className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
                Status
              </label>
              <div className="mt-2">
                <Badge variant="destructive" className="text-sm">
                  <AlertCircle className="mr-1.5 h-3.5 w-3.5" />
                  Failed
                </Badge>
              </div>
            </div>

            <Separator />

            {/* Error Message */}
            {relatedNotification && (
              <>
                <div className="space-y-4">
                  <label className="text-xs font-medium text-muted-foreground uppercase tracking-wider flex items-center gap-2">
                    <AlertCircle className="h-4 w-4 text-destructive" />
                    Error Message
                  </label>
                  <div className="rounded-lg bg-destructive/5 border border-destructive/20 p-4">
                    <p className="text-sm text-foreground leading-relaxed">
                      {relatedNotification.message}
                    </p>
                    {relatedNotification.metadata?.error && (
                      <pre className="mt-3 text-xs font-mono bg-background/50 p-3 rounded border border-destructive/10 overflow-x-auto whitespace-pre-wrap break-words">
                        {relatedNotification.metadata.error}
                      </pre>
                    )}
                  </div>
                </div>

                <Separator />
              </>
            )}

            {/* Database Info */}
            <div className="space-y-4">
              <label className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
                Database Information
              </label>
              <div className="space-y-3">
                <div className="flex items-start justify-between">
                  <span className="text-sm text-muted-foreground">Connection</span>
                  <span className="text-sm font-medium text-right">
                    {connectionName}
                  </span>
                </div>
                {backup.database_name && (
                  <div className="flex items-start justify-between">
                    <span className="text-sm text-muted-foreground">Database</span>
                    <span className="text-sm font-medium font-mono text-right">
                      {backup.database_name}
                    </span>
                  </div>
                )}
              </div>
            </div>

            <Separator />

            {/* Timing Info */}
            <div className="space-y-4">
              <label className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
                Timing
              </label>
              <div className="space-y-3">
                <div className="flex items-start justify-between">
                  <span className="text-sm text-muted-foreground flex items-center gap-2">
                    <Calendar className="h-4 w-4" />
                    Started
                  </span>
                  <span className="text-sm font-medium text-right">
                    {new Date(backup.started_time).toLocaleString()}
                  </span>
                </div>
                {backup.completed_time && (
                  <div className="flex items-start justify-between">
                    <span className="text-sm text-muted-foreground flex items-center gap-2">
                      <Clock className="h-4 w-4" />
                      Failed At
                    </span>
                    <span className="text-sm font-medium text-right">
                      {new Date(backup.completed_time).toLocaleString()}
                    </span>
                  </div>
                )}
              </div>
            </div>

            {/* Path */}
            {backup.path && (
              <>
                <Separator />
                <div className="space-y-4">
                  <label className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
                    Attempted Path
                  </label>
                  <div className="rounded-lg bg-muted/50 p-3">
                    <code className="text-xs font-mono text-foreground/80 break-all">
                      {backup.path}
                    </code>
                  </div>
                </div>
              </>
            )}
          </div>
        </ScrollArea>
      </SheetContent>
    </Sheet>
  );
}
