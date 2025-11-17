'use client';

import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Switch } from "@/components/ui/switch";
import { Label } from "@/components/ui/label";
import { AlertCircle, Cloud } from "lucide-react";
import { useConnections } from "@/hooks/use-connections";
import { useSettings } from "@/hooks/use-settings";
import type { Connection } from "@/types/connection";
import { useState } from "react";

interface DeleteConnectionDialogProps {
  connection: Connection | null;
  onClose: () => void;
}

export function DeleteConnectionDialog({
  connection,
  onClose,
}: DeleteConnectionDialogProps) {
  const { removeConnection, isDeleting } = useConnections();
  const { settings } = useSettings();
  const [cleanupS3, setCleanupS3] = useState(false);

  const handleDelete = () => {
    if (connection) {
      removeConnection({ id: connection.id, cleanupS3 }, {
        onSuccess: () => {
          onClose();
        },
      });
    }
  };

  return (
    <Dialog open={!!connection} onOpenChange={() => onClose()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <AlertCircle className="h-5 w-5 text-destructive" />
            Delete Connection
          </DialogTitle>
          <DialogDescription>
            Are you sure you want to delete this connection?
          </DialogDescription>
        </DialogHeader>
        
        {connection && (
          <div className="py-4 space-y-4">
            <div className="">
              <div className="flex">
                <span className="text-sm font-medium mr-2">Name:</span>
                <span className="text-sm text-muted-foreground">{connection.name}</span>
              </div>
              <div className="flex">
                <span className="text-sm font-medium mr-2">Type:</span>
                <span className="text-sm text-muted-foreground">{connection.type}</span>
              </div>
            </div>

            {settings?.s3_enabled && (
              <div className="flex items-center justify-between rounded-lg border p-4 bg-muted/50">
                <div className="space-y-0.5">
                  <Label htmlFor="cleanup-s3" className="text-sm font-medium flex items-center gap-2">
                    <Cloud className="h-4 w-4" />
                    Also delete S3 backups
                  </Label>
                  <p className="text-xs text-muted-foreground">
                    Permanently delete all backups stored in S3 for this connection
                  </p>
                </div>
                <Switch
                  id="cleanup-s3"
                  checked={cleanupS3}
                  onCheckedChange={(checked) => setCleanupS3(checked)}
                />
              </div>
            )}

            <p className="text-sm text-muted-foreground">
              This action cannot be undone. This will permanently delete the connection and all associated backup schedules.
            </p>
          </div>
        )}


        <DialogFooter>
          <Button
            variant="outline"
            onClick={onClose}
            disabled={isDeleting}
          >
            Cancel
          </Button>
          <Button
            variant="destructive"
            onClick={handleDelete}
            disabled={isDeleting}
          >
            {isDeleting ? "Deleting..." : "Delete Connection"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
