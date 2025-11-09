"use client";

import { useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Button } from "@/components/ui/button";
import { AlertCircle, Database } from "lucide-react";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { useConnections } from "@/hooks/use-connections";
import { useBackup } from "@/hooks/use-backup";
import type { Backup } from "@/types/backup";

interface RestoreDialogProps {
  backup: Backup | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function RestoreDialog({ backup, open, onOpenChange }: RestoreDialogProps) {
  const [selectedConnectionId, setSelectedConnectionId] = useState<string>("");
  const [confirmed, setConfirmed] = useState(false);
  
  const { connections, isLoading: isLoadingConnections } = useConnections();
  const { restoreBackupToDatabase, isRestoring } = useBackup();

  const handleRestore = () => {
    if (!backup || !selectedConnectionId) return;

    restoreBackupToDatabase(
      {
        backupId: backup.id,
        connectionId: selectedConnectionId,
      },
      {
        onSuccess: () => {
          onOpenChange(false);
          setSelectedConnectionId("");
          setConfirmed(false);
        },
      }
    );
  };

  const handleCancel = () => {
    onOpenChange(false);
    setSelectedConnectionId("");
    setConfirmed(false);
  };

  const selectedConnection = connections?.find(
    (conn) => conn.id === selectedConnectionId
  );

  const compatibleConnections = connections?.filter(
    (conn) => conn.type === backup?.database_type
  );

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Database className="h-5 w-5" />
            Restore Database Backup
          </DialogTitle>
          <DialogDescription>
            Select a target connection to restore this backup.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-4">
          <div className="rounded-lg border p-3 space-y-1">
            <p className="text-sm font-medium">Backup Details</p>
            <p className="text-sm text-muted-foreground">
              Database: {backup?.database_name}
            </p>
            <p className="text-sm text-muted-foreground">
              Type: {backup?.database_type}
            </p>
            <p className="text-sm text-muted-foreground">
              Created: {backup?.created_at ? new Date(backup.created_at).toLocaleString() : 'N/A'}
            </p>
          </div>

          <div className="space-y-2">
            <label className="text-sm font-medium">Target Database</label>
            <Select
              value={selectedConnectionId}
              onValueChange={setSelectedConnectionId}
              disabled={isLoadingConnections || isRestoring}
            >
              <SelectTrigger>
                <SelectValue placeholder="Select a database connection..." />
              </SelectTrigger>
              <SelectContent>
                {compatibleConnections && compatibleConnections.length > 0 ? (
                  compatibleConnections.map((conn) => (
                    <SelectItem key={conn.id} value={conn.id}>
                      <div className="flex items-center gap-2 min-w-0">
                        <span className="truncate max-w-[200px]" title={conn.name}>
                          {conn.name}
                        </span>
                        <span className="text-xs text-muted-foreground flex-shrink-0">
                          ({conn.type})
                        </span>
                      </div>
                    </SelectItem>
                  ))
                ) : (
                  <SelectItem value="none" disabled>
                    No compatible connections found
                  </SelectItem>
                )}
              </SelectContent>
            </Select>
          </div>

          <Alert>
            <AlertCircle className="h-4 w-4" />
            <AlertDescription className="text-sm">
              <strong>Tip:</strong> Create a new empty database first (e.g. <strong>{selectedConnection?.name}_new</strong>), restore there, test, then switch.
            </AlertDescription>
          </Alert>

          <div className="flex items-start space-x-2">
            <input
              type="checkbox"
              id="confirm-restore"
              checked={confirmed}
              onChange={(e) => setConfirmed(e.target.checked)}
              className="h-4 w-4 rounded border-gray-300"
              disabled={!selectedConnectionId || isRestoring}
            />
            <label
              htmlFor="confirm-restore"
              className="text-sm leading-tight cursor-pointer"
            >
              Confirm restore to <strong>{selectedConnection?.name || 'selected database'}</strong>
            </label>
          </div>
        </div>

        <DialogFooter>
          <Button
            variant="outline"
            onClick={handleCancel}
            disabled={isRestoring}
          >
            Cancel
          </Button>
          <Button
            onClick={handleRestore}
            disabled={!selectedConnectionId || !confirmed || isRestoring}
          >
            {isRestoring ? "Restoring..." : "Restore Backup"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
