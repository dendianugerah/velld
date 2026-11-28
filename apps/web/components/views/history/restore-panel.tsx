"use client";

import { useState } from "react";
import Link from "next/link";
import { useMutation } from "@tanstack/react-query";
import { Button } from "@/components/ui/button";
import { Select } from "@cloudflare/kumo/components/select";
import { useConnections } from "@/hooks/use-connections";
import { restoreBackup } from "@/lib/api/backups";
import type { Backup } from "@/types/backup";
import { recoveryError } from "./recovery-error";

export function RestorePanel({ backup }: { backup: Backup }) {
  const [target, setTarget] = useState("");
  const [confirmed, setConfirmed] = useState(false);
  const { connections, isLoading, error, refetch } = useConnections();
  const restore = useMutation({
    mutationFn: () =>
      restoreBackup({ backup_id: backup.id, connection_id: target }),
  });
  const compatible =
    connections?.filter(
      (connection) => connection.type === backup.database_type,
    ) ?? [];
  const selected = compatible.find((connection) => connection.id === target);
  const targets = Object.fromEntries(
    compatible.map((connection) => [
      connection.id,
      `${connection.name} · ${connection.database || connection.host}`,
    ]),
  );

  return (
    <div className="space-y-5">
      <h2 className="text-lg font-semibold tracking-tight">
        Restore to a database
      </h2>
      {error ? (
        <div role="alert" className="space-y-2 text-sm">
          <p>Couldn’t load target connections.</p>
          <Button variant="outline" size="sm" onClick={() => refetch()}>
            Try again
          </Button>
        </div>
      ) : (
        <div className="max-w-lg space-y-2 text-sm">
          <span className="block">Target connection</span>
          <Select
            aria-label="Target connection"
            value={target}
            disabled={isLoading || restore.isPending}
            onValueChange={(value) => {
              setTarget(value ?? "");
              setConfirmed(false);
              restore.reset();
            }}
            items={targets}
            placeholder="Choose a connection"
            className="w-full"
          >
            {compatible.map((connection) => (
              <Select.Option key={connection.id} value={connection.id}>
                {connection.name} · {connection.database || connection.host}
              </Select.Option>
            ))}
          </Select>
        </div>
      )}
      {!isLoading && !error && !compatible.length && (
        <p className="text-sm text-muted-foreground">
          No compatible connections.{" "}
          <Link className="text-primary underline" href="/connections">
            Add a connection
          </Link>
          .
        </p>
      )}
      {selected && (
        <label className="flex max-w-lg items-start gap-2.5 rounded-md border border-warning/30 bg-warning/5 p-3 text-sm leading-relaxed">
          <input
            type="checkbox"
            className="mt-1 h-4 w-4 shrink-0 accent-primary"
            checked={confirmed}
            disabled={restore.isPending}
            onChange={(event) => setConfirmed(event.target.checked)}
          />
          <span>
            I understand this will write to <strong>{selected.name}</strong>
            {selected.id === backup.connection_id
              ? ", the original connection"
              : ""}
            . Existing objects or data may be replaced.
          </span>
        </label>
      )}
      {restore.isError && (
        <p role="alert" className="break-words text-sm text-destructive">
          {recoveryError(
            restore.error,
            "Couldn’t restore this backup. Try again.",
          )}
        </p>
      )}
      {restore.isSuccess && (
        <p role="status" className="text-sm text-success">
          Backup restored to {selected?.name}.
        </p>
      )}
      <Button
        disabled={
          !selected || !confirmed || restore.isPending || restore.isSuccess
        }
        onClick={() => restore.mutate()}
      >
        {restore.isPending ? "Restoring…" : "Restore backup"}
      </Button>
    </div>
  );
}
