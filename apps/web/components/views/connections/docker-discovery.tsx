"use client";

import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Button } from "@cloudflare/kumo/components/button";
import { DatabaseMark } from "@/components/ui/database-mark";
import { FormInput as Input } from "@/components/ui/form-field";
import { LayerCard } from "@cloudflare/kumo/components/layer-card";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import {
  discoverDockerDatabases,
  importDockerDatabase,
  type DockerDatabase,
} from "@/lib/api/connections";
import { connectionError } from "@/lib/connection-form";
import { typeLabels } from "@/types/base";
import { useToast } from "@/hooks/use-toast";

export function DockerDiscovery({ onManual }: { onManual: () => void }) {
  const [databaseNames, setDatabaseNames] = useState<Record<string, string>>(
    {},
  );
  const queryClient = useQueryClient();
  const { toast } = useToast();
  const discovery = useQuery({
    queryKey: ["docker-databases"],
    queryFn: discoverDockerDatabases,
    retry: false,
  });
  const add = useMutation({
    mutationFn: ({
      container,
      database,
    }: {
      container: DockerDatabase;
      database: string;
    }) => importDockerDatabase(container.id, database),
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ["connections"] }),
        queryClient.invalidateQueries({ queryKey: ["docker-databases"] }),
      ]);
      toast({
        title: "Docker database added",
        description: "Daily backups are scheduled with 30 days of retention.",
      });
    },
  });

  async function scanAgain() {
    if (discovery.isFetching || add.isPending) return;

    const previousIds = new Set(
      discovery.data?.containers.map((container) => container.id) ?? [],
    );
    const result = await discovery.refetch();
    if (result.isError) {
      toast({
        title: "Docker scan failed",
        description: "Check Docker access and try again.",
        variant: "destructive",
      });
      return;
    }

    const newCount =
      result.data?.containers.filter(
        (container) => !previousIds.has(container.id),
      ).length ?? 0;
    toast({
      title: newCount
        ? `Found ${newCount} ${newCount === 1 ? "database" : "databases"}`
        : "No new databases found",
    });
  }

  return (
    <>
      <div className="mb-6 flex flex-wrap items-center justify-between gap-4">
        <h2 className="text-sm font-semibold">Docker databases</h2>
        <Button
          variant="secondary"
          className="min-w-24 justify-center"
          aria-disabled={discovery.isFetching || add.isPending}
          onClick={() => void scanAgain()}
        >
          {discovery.isFetching ? "Scanning…" : "Scan again"}
        </Button>
      </div>
      <div className="max-w-[760px] space-y-6">
        <span className="sr-only" role="status">
          {discovery.isFetching ? "Scanning Docker databases" : ""}
        </span>
        {discovery.isLoading && (
          <div aria-label="Scanning Docker" className="space-y-4">
            <Skeleton className="h-36 w-full" />
            <Skeleton className="h-36 w-full" />
          </div>
        )}
        {discovery.error && (
          <div
            role="alert"
            className="rounded-lg border p-5 text-sm text-destructive"
          >
            {discovery.data
              ? "Couldn’t refresh Docker databases. Showing the previous results. "
              : "Couldn’t scan Docker databases. "}
            {connectionError(discovery.error)} Try scanning again.
          </div>
        )}
        {discovery.data?.hint && (
          <p className="rounded-lg border bg-muted/30 p-5 text-sm leading-relaxed">
            {discovery.data.hint}
          </p>
        )}
        {discovery.data?.containers.length === 0 && (
          <div className="rounded-lg border border-dashed p-8 text-center">
            <h2 className="text-sm font-semibold">No databases found</h2>
            <button
              type="button"
              onClick={onManual}
              className="mt-4 rounded-sm text-sm font-medium underline underline-offset-4 focus-visible:outline focus-visible:outline-2 focus-visible:outline-ring"
            >
              Enter connection details
            </button>
          </div>
        )}
        {!!discovery.data?.containers.length && (
          <p className="text-xs text-muted-foreground">
            Imported databases get daily backups at midnight (server time), kept
            for 30 days. You can change this later.
          </p>
        )}
        {discovery.data?.containers.map((container) => {
          const database = databaseNames[container.id] ?? container.database;
          return (
            <LayerCard key={container.id} className="rounded-lg shadow-none">
              <form
                className="space-y-5 p-5"
                onSubmit={(event) => {
                  event.preventDefault();
                  add.mutate({ container, database: database.trim() });
                }}
              >
                <div className="flex min-w-0 items-center gap-3">
                  <DatabaseMark type={container.type} small />
                  <div className="min-w-0">
                    <h2 className="break-all text-sm font-medium">
                      {container.name}
                    </h2>
                    <p className="mt-1 break-all text-xs text-muted-foreground">
                      {typeLabels[container.type] ?? container.type} ·{" "}
                      {container.host}:{container.port}
                    </p>
                  </div>
                </div>
                <div className="flex flex-col gap-3 sm:flex-row sm:items-end">
                  <div className="min-w-0 flex-1 space-y-2">
                    <Label htmlFor={`docker-db-${container.id}`}>
                      Database name
                    </Label>
                    <Input
                      id={`docker-db-${container.id}`}
                      aria-label="Database name"
                      required
                      value={database}
                      disabled={add.isPending}
                      onChange={(event) =>
                        setDatabaseNames((previous) => ({
                          ...previous,
                          [container.id]: event.target.value,
                        }))
                      }
                    />
                  </div>
                  <Button
                    type="submit"
                    variant="primary"
                    disabled={
                      discovery.isFetching ||
                      !!discovery.error ||
                      add.isPending ||
                      !database.trim()
                    }
                  >
                    {add.isPending &&
                    add.variables?.container.id === container.id
                      ? "Adding…"
                      : "Add connection"}
                  </Button>
                </div>
                {add.error && add.variables?.container.id === container.id && (
                  <p role="alert" className="text-sm text-destructive">
                    {connectionError(add.error)}
                  </p>
                )}
              </form>
            </LayerCard>
          );
        })}
      </div>
    </>
  );
}
