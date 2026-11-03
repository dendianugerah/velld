'use client';

import { ShieldCheck, ShieldAlert, Cloud, Copy, Loader2 } from 'lucide-react';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import { Separator } from '@/components/ui/separator';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { useToast } from '@/hooks/use-toast';
import {
  createBaseBackup,
  getWALStatus,
  getWALCommandConfig,
  preparePITR,
  type WALStatus,
  type WALCommandConfig,
} from '@/lib/api/pitr';
import type { Connection } from '@/types/connection';
import { useState } from 'react';

interface PITRDialogProps {
  connection: Connection | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

// apiBaseUrl and archivePath are operator inputs; we expose sensible defaults so
// the generated commands are copy-pasteable.
const DEFAULT_API_BASE = typeof window !== 'undefined' ? window.location.origin : 'https://velld.example.com';
const DEFAULT_ARCHIVE_PATH = '/var/lib/velld/wal';

export function PITRDialog({ connection, open, onOpenChange }: PITRDialogProps) {
  const { toast } = useToast();
  const [isCreating, setIsCreating] = useState(false);
  const [isLoadingStatus, setIsLoadingStatus] = useState(false);
  const [isPreparing, setIsPreparing] = useState(false);
  const [status, setStatus] = useState<WALStatus | null>(null);
  const [config, setConfig] = useState<WALCommandConfig | null>(null);
  const [targetTime, setTargetTime] = useState('');
  const [dataDir, setDataDir] = useState('/var/lib/postgresql/restore');
  const [restoreCommand, setRestoreCommand] = useState('');

  const reset = () => {
    setStatus(null);
    setConfig(null);
    setTargetTime('');
    setRestoreCommand('');
  };

  const handleBackup = async () => {
    if (!connection) return;
    setIsCreating(true);
    try {
      const result = await createBaseBackup(connection.id);
      toast({ title: 'Base backup created', description: result.object_key });
    } catch (err) {
      toast({ title: 'Base backup failed', description: (err as Error).message, variant: 'destructive' });
    } finally {
      setIsCreating(false);
    }
  };

  const handleStatus = async () => {
    if (!connection) return;
    setIsLoadingStatus(true);
    try {
      const [walStatus, walConfig] = await Promise.all([
        getWALStatus(connection.id),
        getWALCommandConfig(connection.id, DEFAULT_API_BASE, DEFAULT_ARCHIVE_PATH),
      ]);
      setStatus(walStatus);
      setConfig(walConfig);
      setRestoreCommand(walConfig.restore_command);
    } catch (err) {
      toast({ title: 'Failed to load WAL status', description: (err as Error).message, variant: 'destructive' });
    } finally {
      setIsLoadingStatus(false);
    }
  };

  const handlePrepare = async () => {
    if (!connection) return;
    if (!targetTime || !dataDir || !restoreCommand) {
      toast({ title: 'Missing fields', description: 'Target time, data directory, and restore command are required', variant: 'destructive' });
      return;
    }
    setIsPreparing(true);
    try {
      const result = await preparePITR({
        connection_id: connection.id,
        target_time: new Date(targetTime).toISOString(),
        data_dir: dataDir,
        restore_command: restoreCommand,
      });
      toast({ title: 'Recovery configuration prepared', description: result.recovery_file });
    } catch (err) {
      toast({ title: 'Could not prepare recovery', description: (err as Error).message, variant: 'destructive' });
    } finally {
      setIsPreparing(false);
    }
  };

  const copy = (text: string) => {
    navigator.clipboard?.writeText(text);
    toast({ title: 'Copied to clipboard' });
  };

  return (
    <Dialog open={open} onOpenChange={(o) => { if (!o) reset(); onOpenChange(o); }}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>Point-in-time recovery</DialogTitle>
          <DialogDescription>
            {connection?.name} — physical base backups plus archived WAL let you recover to any past moment.
          </DialogDescription>
        </DialogHeader>

        <Tabs defaultValue="status">
          <TabsList className="grid w-full grid-cols-3">
            <TabsTrigger value="status">Status</TabsTrigger>
            <TabsTrigger value="setup">Setup</TabsTrigger>
            <TabsTrigger value="recover">Recover</TabsTrigger>
          </TabsList>

          <TabsContent value="status" className="space-y-4 pt-4">
            <div className="flex items-center gap-2">
              <Button onClick={handleStatus} disabled={isLoadingStatus} variant="outline">
                {isLoadingStatus ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Cloud className="mr-2 h-4 w-4" />}
                Refresh WAL status
              </Button>
              <Button onClick={handleBackup} disabled={isCreating} variant="outline">
                {isCreating ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <ShieldCheck className="mr-2 h-4 w-4" />}
                Create base backup
              </Button>
            </div>

            {status && (
              <div className="rounded-lg border p-4 space-y-2">
                <div className="flex items-center justify-between">
                  <span className="text-sm text-muted-foreground">Archived segments</span>
                  <span className="font-medium">{status.segment_count}</span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-sm text-muted-foreground">Recovery window</span>
                  {status.has_gaps ? (
                    <Badge variant="destructive"><ShieldAlert className="h-3 w-3 mr-1" />Gaps detected</Badge>
                  ) : (
                    <Badge variant="secondary" className="bg-emerald-100 text-emerald-700 dark:bg-emerald-900 dark:text-emerald-300">
                      <ShieldCheck className="h-3 w-3 mr-1" />Intact
                    </Badge>
                  )}
                </div>
                {status.has_gaps && (
                  <div className="text-xs text-muted-foreground font-mono break-all">
                    Missing: {status.gaps.slice(0, 5).join(', ')}{status.gaps.length > 5 ? ` (+${status.gaps.length - 5} more)` : ''}
                  </div>
                )}
              </div>
            )}
          </TabsContent>

          <TabsContent value="setup" className="space-y-4 pt-4">
            <p className="text-sm text-muted-foreground">
              Paste these into the database server&apos;s <code>postgresql.conf</code>. Set <code>archive_mode = on</code> and export <code>VELD_TOKEN</code> on that host.
            </p>
            <CommandRow label="archive_command" value={config?.archive_command ?? ''} onCopy={copy} />
            <CommandRow label="restore_command" value={config?.restore_command ?? ''} onCopy={copy} />
          </TabsContent>

          <TabsContent value="recover" className="space-y-4 pt-4">
            <div className="space-y-2">
              <Label htmlFor="target-time">Target time</Label>
              <Input id="target-time" type="datetime-local" value={targetTime} onChange={(e) => setTargetTime(e.target.value)} />
            </div>
            <div className="space-y-2">
              <Label htmlFor="data-dir">Restored data directory</Label>
              <Input id="data-dir" value={dataDir} onChange={(e) => setDataDir(e.target.value)} />
            </div>
            <div className="space-y-2">
              <Label htmlFor="restore-cmd">restore_command</Label>
              <Input id="restore-cmd" value={restoreCommand} onChange={(e) => setRestoreCommand(e.target.value)} />
            </div>
            <Separator />
            <Button onClick={handlePrepare} disabled={isPreparing} className="w-full">
              {isPreparing ? 'Preparing…' : 'Prepare recovery configuration'}
            </Button>
            <p className="text-xs text-muted-foreground">
              This writes the recovery files into the data directory. Starting and promoting the server is a deliberate operator step.
            </p>
          </TabsContent>
        </Tabs>
      </DialogContent>
    </Dialog>
  );
}

function CommandRow({ label, value, onCopy }: { label: string; value: string; onCopy: (v: string) => void }) {
  return (
    <div className="space-y-1">
      <Label>{label}</Label>
      <div className="flex items-center gap-2">
        <code className="flex-1 truncate rounded bg-muted px-2 py-1 text-xs">{value || '—'}</code>
        <Button size="icon" variant="ghost" disabled={!value} onClick={() => onCopy(value)}>
          <Copy className="h-4 w-4" />
        </Button>
      </div>
    </div>
  );
}
