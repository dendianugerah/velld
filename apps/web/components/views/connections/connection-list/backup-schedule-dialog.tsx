"use client";

import { useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { useSettings } from "@/hooks/use-settings";
import { useToast } from "@/hooks/use-toast";
import { scheduleBackup, updateSchedule, disableBackupSchedule } from "@/lib/api/backups";
import { updateConnectionSettings } from "@/lib/api/connections";
import { getScheduleFrequency } from "@/lib/helper";
import type { Connection } from "@/types/connection";

const schedules: Record<string, string> = {
  test: "0 */1 * * * *",
  hourly: "0 0 * * * *",
  daily: "0 0 0 * * *",
  weekly: "0 0 0 * * 0",
  monthly: "0 0 0 1 * *",
};

interface BackupScheduleDialogProps {
  connectionId: string | null;
  connection?: Connection;
  returnFocusTo?: HTMLElement | null;
  onClose: () => void;
}

export function BackupScheduleDialog({ connectionId, connection, returnFocusTo, onClose }: BackupScheduleDialogProps) {
  if (!connectionId || !connection) return null;
  return <ScheduleForm key={connectionId} connection={connection} returnFocusTo={returnFocusTo} onClose={onClose} />;
}

export function ScheduleForm({ connection, returnFocusTo, onClose = () => {}, inline = false }: { connection: Connection; returnFocusTo?: HTMLElement | null; onClose?: () => void; inline?: boolean }) {
  const { settings } = useSettings();
  const { toast } = useToast();
  const queryClient = useQueryClient();
  const [enabled, setEnabled] = useState(connection.backup_enabled);
  const [frequency, setFrequency] = useState(getScheduleFrequency(connection.cron_schedule) || "daily");
  const [customCron, setCustomCron] = useState(connection.cron_schedule || "0 0 0 * * *");
  const [retention, setRetention] = useState(connection.retention_days ?? 30);
  const [s3Cleanup, setS3Cleanup] = useState(connection.s3_cleanup_on_retention ?? true);
  const [validationError, setValidationError] = useState("");

  const save = useMutation({
    mutationFn: async () => {
      if (enabled) {
        const params = { cron_schedule: frequency === "custom" ? customCron.trim() : schedules[frequency], retention_days: retention };
        if (connection.backup_enabled) await updateSchedule(connection.id, params);
        else await scheduleBackup({ connection_id: connection.id, ...params });
        if (settings?.s3_enabled && s3Cleanup !== connection.s3_cleanup_on_retention) {
          await updateConnectionSettings(connection.id, { s3_cleanup_on_retention: s3Cleanup });
        }
      } else if (connection.backup_enabled) {
        await disableBackupSchedule(connection.id);
      }
    },
    onSettled: () => {
      // A schedule may have saved even if the optional S3 preference failed.
      queryClient.invalidateQueries({ queryKey: ["connections"] });
      queryClient.invalidateQueries({ queryKey: ["backups"] });
      queryClient.invalidateQueries({ queryKey: ["connection", connection.id] });
    },
    onSuccess: () => {
      toast({ title: "Schedule saved", description: enabled ? "Automatic backups are enabled." : "This connection uses manual backups." });
      onClose();
    },
  });

  function submit(event: React.FormEvent) {
    event.preventDefault();
    if (enabled && frequency === "custom" && customCron.trim().split(/\s+/).length !== 6) {
      setValidationError("Enter all six fields: second, minute, hour, day, month, weekday.");
      return;
    }
    setValidationError("");
    save.mutate();
  }

  const form = <form onSubmit={submit} className="space-y-6">
        <div className="flex items-center justify-between gap-6 rounded-xl bg-muted/70 p-4">
          <div><Label htmlFor="schedule-enabled">Automatic backups</Label><p className="mt-1.5 text-xs text-muted-foreground">Off means you’ll run backups manually.</p></div>
          <Switch id="schedule-enabled" checked={enabled} onCheckedChange={setEnabled} disabled={save.isPending} />
        </div>

        {enabled && <div className="space-y-5">
          <div className="space-y-2">
            <Label htmlFor="schedule-frequency">How often</Label>
            <Select value={frequency} onValueChange={setFrequency} disabled={save.isPending}>
              <SelectTrigger id="schedule-frequency"><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectItem value="hourly">Hourly</SelectItem>
                <SelectItem value="daily">Daily, at midnight</SelectItem>
                <SelectItem value="weekly">Weekly, on Sunday</SelectItem>
                <SelectItem value="monthly">Monthly, on the 1st</SelectItem>
                <SelectItem value="custom">Custom cron expression</SelectItem>
                <SelectItem value="test">Every minute (testing)</SelectItem>
              </SelectContent>
            </Select>
            <p className="text-xs leading-relaxed text-muted-foreground">Schedules use your Velld server’s time zone.</p>
          </div>
          {frequency === "custom" && <div className="space-y-2">
            <Label htmlFor="schedule-cron">Cron expression</Label>
            <Input id="schedule-cron" value={customCron} onChange={event => setCustomCron(event.target.value)} required disabled={save.isPending} aria-describedby="cron-help" className="font-mono" />
            <p id="cron-help" className="text-xs leading-relaxed text-muted-foreground">Second · minute · hour · day · month · weekday. Use 0 for seconds to run no more than once per minute.</p>
          </div>}
          <div className="space-y-2">
            <Label htmlFor="schedule-retention">Retention period</Label>
            <div className="flex items-center gap-3"><Input id="schedule-retention" type="number" min={connection.retention_count ? 0 : 1} step={1} required value={Number.isNaN(retention) ? "" : retention} onChange={event => setRetention(event.target.valueAsNumber)} disabled={save.isPending} className="w-28" /><span className="text-sm text-muted-foreground">days</span></div>
            <p className="text-xs leading-relaxed text-muted-foreground">{connection.retention_count ? `A limit of ${connection.retention_count} backups also applies. Use 0 days for count-only retention.` : "Older local backups are deleted automatically."}</p>
          </div>
          {settings?.s3_enabled && <div className="flex items-start justify-between gap-5 border-t pt-5">
            <div><Label htmlFor="schedule-s3-cleanup">Apply retention to S3, too</Label><p className="mt-1.5 text-xs leading-relaxed text-muted-foreground">Delete S3 backups outside this retention policy.</p></div>
            <Switch id="schedule-s3-cleanup" checked={s3Cleanup} onCheckedChange={setS3Cleanup} disabled={save.isPending} />
          </div>}
        </div>}
        {(validationError || save.error) && <p role="alert" className="text-sm text-destructive">{validationError || save.error?.message || "Couldn’t save the schedule. Please try again."}</p>}
        <DialogFooter className={inline ? "sm:justify-start" : undefined}>
          {!inline && <Button type="button" variant="outline" disabled={save.isPending} onClick={onClose}>Cancel</Button>}
          <Button type="submit" disabled={save.isPending}>{save.isPending ? "Saving…" : "Save schedule"}</Button>
        </DialogFooter>
      </form>;
  if (inline) return <div className="max-w-[640px]">{form}</div>;
  return <Dialog open onOpenChange={open => { if (!open && !save.isPending) onClose(); }}>
    <DialogContent onCloseAutoFocus={event => { if (returnFocusTo?.isConnected) { event.preventDefault(); returnFocusTo.focus(); } }}>
      <DialogHeader>
        <DialogTitle>Backup schedule</DialogTitle>
        <DialogDescription className="break-words">{connection.name}</DialogDescription>
      </DialogHeader>
      {form}
    </DialogContent>
  </Dialog>;
}