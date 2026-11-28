"use client";

import { useEffect, useRef, useState } from "react";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { FieldHelp } from "@/components/ui/field-help";
import { apiRequest } from "@/lib/api-client";
import type { BackupBranch } from "@/lib/api/backups";
import type { Connection } from "@/types/connection";

type RecoveryStatus = { from?: string; through?: string };

export function PointInTimeRecoveryDialog({
  connection,
  onClose = () => {},
  inline = false,
}: {
  connection: Connection;
  onClose?: () => void;
  inline?: boolean;
}) {
  const endpoint = `/api/connections/${encodeURIComponent(connection.id)}/pitr`;
  const mysql = connection.type === "mysql";
  const [status, setStatus] = useState<RecoveryStatus | null>(null);
  const [target, setTarget] = useState("");
  const [recovered, setRecovered] = useState<BackupBranch | null>(null);
  const [pending, setPending] = useState(false);
  const [error, setError] = useState("");
  const request = useRef<AbortController | null>(null);

  function showError(err: unknown) {
    let message = "Point-in-time recovery failed.";
    if (err instanceof Error) {
      try {
        message =
          (JSON.parse(err.message) as { message?: string }).message || message;
      } catch {}
    }
    setError(message);
  }

  useEffect(() => {
    const controller = new AbortController();
    request.current = controller;
    apiRequest<{ data: RecoveryStatus }>(endpoint, {
      signal: controller.signal,
    })
      .then(({ data }) => setStatus(data))
      .catch((err) => {
        if (!controller.signal.aborted) showError(err);
      });
    return () => request.current?.abort();
  }, [endpoint]);

  async function run(action: "prepare" | "restore" | "remove") {
    if (
      action === "remove" &&
      !window.confirm(
        "Remove this recovery baseline?",
      )
    )
      return;
    const controller = new AbortController();
    request.current = controller;
    setPending(true);
    setError("");
    try {
      if (action === "remove") {
        await apiRequest(endpoint, {
          method: "DELETE",
          signal: controller.signal,
        });
        setStatus({});
      } else if (action === "restore") {
        const date = new Date(target);
        if (!Number.isFinite(date.getTime())) {
          setError("Choose a valid timestamp.");
          return;
        }
        const { data } = await apiRequest<{ data: BackupBranch }>(endpoint, {
          method: "POST",
          body: JSON.stringify({ target: date.toISOString() }),
          signal: controller.signal,
        });
        setRecovered(data);
      } else {
        const { data } = await apiRequest<{ data: RecoveryStatus }>(endpoint, {
          method: "POST",
          body: "{}",
          signal: controller.signal,
        });
        setStatus(data);
      }
    } catch (err) {
      if (!controller.signal.aborted) showError(err);
    } finally {
      if (!controller.signal.aborted) setPending(false);
    }
  }

  const content = (
    <>
      {inline ? (
        <div>
          <div className="flex items-center gap-1">
            <h2 className="text-sm font-semibold">Point-in-time recovery</h2>
            <FieldHelp label="About point-in-time recovery">
              Restores an unmasked copy to a temporary Docker database without
              overwriting the source.
            </FieldHelp>
          </div>
        </div>
      ) : (
        <DialogHeader>
          <DialogTitle>Point-in-time recovery</DialogTitle>
          <DialogDescription>
            Restore an unmasked copy of {connection.name} in a temporary Docker
            database. The source is not overwritten.
          </DialogDescription>
        </DialogHeader>
      )}
      {error && (
        <p role="alert" className="text-sm text-destructive">
          {error}
        </p>
      )}
      {recovered ? (
        <div className="space-y-2">
          <label htmlFor="pitr-connection" className="text-sm font-medium">
            Connection string
          </label>
          <Input
            id="pitr-connection"
            value={recovered.connection_string}
            readOnly
            className="font-mono text-xs"
            onFocus={(event) => event.currentTarget.select()}
          />
          <p className="text-sm text-muted-foreground">
            Unmasked copy · Expires {new Date(recovered.expires_at).toLocaleString()}
          </p>
        </div>
      ) : status?.from ? (
        <div className="space-y-2">
          <p className="text-sm text-muted-foreground">
            Baseline: {new Date(status.from).toLocaleString()}.
            {mysql && status.through && (
              <>
                {" "}
                Logs archived through{" "}
                {new Date(status.through).toLocaleString()}.
              </>
            )}
          </p>
          <label htmlFor="pitr-target" className="text-sm font-medium">
            {mysql ? "Restore to before (local time)" : "Restore timestamp (local time)"}
          </label>
          <Input
            id="pitr-target"
            type="datetime-local"
            step={1}
            value={target}
            disabled={pending}
            onChange={(event) => setTarget(event.target.value)}
          />
          <div className="flex flex-wrap gap-2">
            {mysql && (
              <Button
                variant="outline"
                size="sm"
                disabled={pending}
                onClick={() => run("prepare")}
              >
                Archive new logs
              </Button>
            )}
            <Button
              variant="ghost"
              size="sm"
              disabled={pending}
              onClick={() => run("remove")}
            >
              Remove baseline
            </Button>
          </div>
        </div>
      ) : (
        status && (
          <p className="text-sm text-muted-foreground">
            Create a baseline first. The source needs{" "}
            {mysql
              ? "row-based binary logging"
              : "WAL archiving and replication access"}
            .
          </p>
        )
      )}
      {pending && (
        <p role="status" className="text-sm text-muted-foreground">
          Working…
        </p>
      )}
      <DialogFooter>
        {!inline && (
          <Button variant="outline" onClick={onClose}>
            {recovered ? "Close" : "Cancel"}
          </Button>
        )}
        {!recovered &&
          status &&
          (status.from ? (
            <Button
              disabled={pending || !target}
              onClick={() => run("restore")}
            >
              Restore copy
            </Button>
          ) : (
            <Button disabled={pending} onClick={() => run("prepare")}>
              Create baseline
            </Button>
          ))}
      </DialogFooter>
    </>
  );
  if (inline) return <div className="space-y-5">{content}</div>;
  return (
    <Dialog
      open
      onOpenChange={(open) => {
        if (!open) onClose();
      }}
    >
      <DialogContent
        onCloseAutoFocus={(event) => {
          const trigger = document.getElementById(
            `connection-actions-${connection.id}`,
          );
          if (trigger) {
            event.preventDefault();
            trigger.focus();
          }
        }}
      >
        {content}
      </DialogContent>
    </Dialog>
  );
}
