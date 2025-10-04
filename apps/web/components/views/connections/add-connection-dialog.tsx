"use client";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { useState } from "react";
import { ConnectionForm } from "./connection-form";
import { ScrollArea } from "@/components/ui/scroll-area";

export function AddConnectionDialog() {
  const [open, setOpen] = useState(false);

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button className="font-medium">
          New Connection
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[600px] max-h-[85vh]">
        <DialogHeader>
          <DialogTitle>Add New Connection</DialogTitle>
          <DialogDescription>
            Add a new database connection. The connection will be tested before saving.
          </DialogDescription>
        </DialogHeader>
        <ScrollArea className="max-h-[calc(85vh-120px)] pr-4">
          <ConnectionForm onSuccess={() => setOpen(false)} onCancel={() => setOpen(false)} />
        </ScrollArea>
      </DialogContent>
    </Dialog>
  );
}
