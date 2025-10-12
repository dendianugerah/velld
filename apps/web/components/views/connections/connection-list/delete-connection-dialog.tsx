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
import { AlertCircle } from "lucide-react";
import { useConnections } from "@/hooks/use-connections";
import type { Connection } from "@/types/connection";

interface DeleteConnectionDialogProps {
  connection: Connection | null;
  onClose: () => void;
}

export function DeleteConnectionDialog({
  connection,
  onClose,
}: DeleteConnectionDialogProps) {
  const { removeConnection, isDeleting } = useConnections();

  const handleDelete = () => {
    if (connection) {
      removeConnection(connection.id, {
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
          <div className="py-4">
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
            <p className="text-sm text-muted-foreground mt-4">
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
