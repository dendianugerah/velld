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
import {
  getConnectionUpgradeInfo,
  upgradeConnectionContainer,
  type UpgradeInfo,
  type UpgradeResult,
} from "@/lib/api/connections";
import type { Connection } from "@/types/connection";

const placeholderByType: Record<string, string> = {
  postgresql: "postgres:16-alpine",
  mysql: "mysql:8.4",
  mariadb: "mariadb:11.4",
  mongodb: "mongo:7.0",
};

function parseApiError(err: unknown, fallback: string) {
  if (err instanceof Error) {
    try {
      return (
        (JSON.parse(err.message) as { message?: string }).message || fallback
      );
    } catch {
      return err.message || fallback;
    }
  }
  return fallback;
}

export function UpgradeDialog({
  connection,
  onClose = () => {},
  inline = false,
}: {
  connection: Connection;
  onClose?: () => void;
  inline?: boolean;
}) {
  const [info, setInfo] = useState<UpgradeInfo | null>(null);
  const [loadingInfo, setLoadingInfo] = useState(true);
  const [targetImage, setTargetImage] = useState("");
  const [pending, setPending] = useState(false);
  const [error, setError] = useState("");
  const [result, setResult] = useState<UpgradeResult | null>(null);
  const request = useRef<AbortController | null>(null);

  useEffect(() => {
    const controller = new AbortController();
    getConnectionUpgradeInfo(connection.id, controller.signal)
      .then(setInfo)
      .catch((err) => {
        if (!controller.signal.aborted)
          setError(
            parseApiError(err, "Couldn’t inspect the Docker container."),
          );
      })
      .finally(() => {
        if (!controller.signal.aborted) setLoadingInfo(false);
      });
    return () => {
      controller.abort();
      request.current?.abort();
    };
  }, [connection.id]);

  async function runUpgrade() {
    const nextImage = targetImage.trim();
    if (!nextImage) return;
    const controller = new AbortController();
    request.current = controller;
    setError("");
    setResult(null);
    setPending(true);
    try {
      setResult(
        await upgradeConnectionContainer(
          connection.id,
          nextImage,
          controller.signal,
        ),
      );
    } catch (err) {
      if (!controller.signal.aborted)
        setError(
          parseApiError(err, "Couldn’t upgrade the database container."),
        );
    } finally {
      if (!controller.signal.aborted) setPending(false);
    }
  }

  const content = (
    <>
      {inline ? (
        <div>
          <div className="flex items-center gap-1">
            <h2 className="text-sm font-semibold">Safe upgrades</h2>
            <FieldHelp label="About safe upgrades">
              Velld takes a snapshot, replaces the Docker container, checks its
              data, and attempts a rollback if the upgrade fails.
            </FieldHelp>
          </div>
          <p className="mt-2 text-[13px] leading-relaxed text-muted-foreground">
            Upgrading interrupts the live database.
          </p>
        </div>
      ) : (
        <DialogHeader>
          <DialogTitle>Upgrade {connection.name}</DialogTitle>
          <DialogDescription>
            Takes a backup snapshot, replaces the container with the target
            image, verifies data health, and rolls back automatically on
            failure.
          </DialogDescription>
        </DialogHeader>
      )}

      {loadingInfo ? (
        <p className="text-sm text-muted-foreground">
          Checking Docker container…
        </p>
      ) : info?.supported ? (
        <div className="space-y-3">
          <div className="rounded-md border bg-muted/40 px-3 py-2 text-xs">
            <span className="text-muted-foreground">Current:</span>{" "}
            <span className="font-mono">{info.current_image}</span>{" "}
            <span className="text-muted-foreground">
              ({info.container_name})
            </span>
          </div>
          {!result && (
            <div className="space-y-1.5">
              <label
                htmlFor="upgrade-target-image"
                className="text-sm font-medium"
              >
                Target image
              </label>
              <Input
                id="upgrade-target-image"
                className="font-mono text-xs"
                placeholder={
                  placeholderByType[connection.type] ?? "postgres:16-alpine"
                }
                value={targetImage}
                onChange={(e) => setTargetImage(e.target.value)}
                disabled={pending}
              />
            </div>
          )}
        </div>
      ) : (
        <p className="text-sm text-muted-foreground">
          {info?.hint ||
            "No matching Docker container found for this connection."}
        </p>
      )}

      {error && (
        <p role="alert" className="text-sm text-destructive">
          {error}
        </p>
      )}

      {result && (
        <div role="status" className="space-y-1 rounded-md border p-3 text-xs">
          <p className="font-medium">
            {result.rolled_back
              ? `Rolled back to ${result.previous_image}${result.rollback_verified ? " (verified healthy)" : ""}`
              : `Upgraded to ${result.target_image}`}
          </p>
          {result.health_check && (
            <p className="text-muted-foreground">{result.health_check}</p>
          )}
          {result.error && <p className="text-destructive">{result.error}</p>}
        </div>
      )}

      <DialogFooter>
        {!inline && (
          <Button variant="outline" onClick={onClose} disabled={pending}>
            {result ? "Close" : "Cancel"}
          </Button>
        )}
        {info?.supported && !result && (
          <Button
            onClick={runUpgrade}
            disabled={pending || !targetImage.trim()}
          >
            {pending ? "Upgrading…" : "Upgrade safely"}
          </Button>
        )}
      </DialogFooter>
    </>
  );
  if (inline) return <div className="space-y-5">{content}</div>;
  return (
    <Dialog
      open
      onOpenChange={(open) => {
        if (!open && !pending) onClose();
      }}
    >
      <DialogContent>{content}</DialogContent>
    </Dialog>
  );
}
