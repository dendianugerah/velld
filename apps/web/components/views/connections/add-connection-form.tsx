"use client";

import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Button } from "@cloudflare/kumo/components/button";
import { LayerCard } from "@cloudflare/kumo/components/layer-card";
import { Select } from "@cloudflare/kumo/components/select";
import { FormField, FormInput } from "@/components/ui/form-field";
import { SegmentedControl } from "@/components/ui/segmented-control";
import {
  discoverDockerDatabases,
  importDockerDatabase,
  saveConnection,
  testConnection,
  testDockerDatabase,
} from "@/lib/api/connections";
import { connectionError } from "@/lib/connection-form";
import { typeLabels, type DatabaseType } from "@/types/base";
import { ConnectionForm } from "./connection-form";

const placeholders = {
  postgresql: "postgresql://user:password@host/database",
  mysql: "mysql://user:password@host/database",
  mongodb: "mongodb://user:password@host/database",
};

export function AddConnectionForm({
  onSuccess,
  onCancel,
}: {
  onSuccess: (id: string) => void;
  onCancel: () => void;
}) {
  const [engine, setEngine] = useState<DatabaseType>("postgresql");
  const [method, setMethod] = useState<"url" | "docker" | "manual">("url");
  const [name, setName] = useState("");
  const [url, setUrl] = useState("");
  const [containerId, setContainerId] = useState("");
  const [database, setDatabase] = useState("");
  const [tested, setTested] = useState(false);
  const [error, setError] = useState("");
  const queryClient = useQueryClient();
  const discovery = useQuery({
    queryKey: ["docker-databases"],
    queryFn: discoverDockerDatabases,
    enabled: method === "docker" && engine !== "d1",
    retry: false,
  });
  const containers =
    discovery.data?.containers.filter((item) => item.type === engine) ?? [];
  const container = containers.find((item) => item.id === containerId);
  const ready =
    !!name.trim() &&
    (method === "url" ? !!url.trim() : !!container && !!database.trim());
  const action = useMutation({
    mutationFn: async (save: boolean) => {
      if (method === "docker") {
        const input = { name: name.trim(), database: database.trim() };
        if (save) return (await importDockerDatabase(containerId, input)).id;
        await testDockerDatabase(containerId, input);
      } else {
        const input = { name: name.trim(), type: engine, connection_url: url };
        if (save) return (await saveConnection(input)).id;
        await testConnection(input);
      }
    },
    onSuccess: (id) => {
      if (!id) {
        setTested(true);
        return;
      }
      void queryClient.invalidateQueries({ queryKey: ["connections"] });
      void queryClient.invalidateQueries({ queryKey: ["docker-databases"] });
      onSuccess(id);
    },
    onError: (err) => {
      setTested(false);
      setError(connectionError(err));
    },
  });

  function changed() {
    setTested(false);
    setError("");
  }

  function run(save: boolean) {
    if (!ready || action.isPending || (save && !tested)) return;
    changed();
    action.mutate(save);
  }

  return (
    <div className="space-y-5 pb-8">
      <FormField id="add-engine" label="Database engine">
        <Select
          id="add-engine"
          aria-label="Database engine"
          className="w-full"
          value={engine}
          disabled={action.isPending}
          items={{ ...typeLabels, mysql: "MySQL / MariaDB" }}
          onValueChange={(value) => {
            if (!value) return;
            changed();
            setEngine(value);
            setUrl("");
            setContainerId("");
            setDatabase("");
          }}
        >
          {Object.entries({ ...typeLabels, mysql: "MySQL / MariaDB" }).map(
            ([value, label]) => (
              <Select.Option key={value} value={value}>
                {label}
              </Select.Option>
            ),
          )}
        </Select>
      </FormField>
      {engine !== "d1" && (
        <fieldset disabled={action.isPending}>
          <SegmentedControl
            label="Connection method"
            value={method}
            options={[
              { value: "url", label: "Connection URL" },
              { value: "docker", label: "Docker" },
              { value: "manual", label: "Details" },
            ]}
            onChange={(value) => {
              changed();
              setMethod(value);
            }}
          />
        </fieldset>
      )}
      {engine === "d1" || method === "manual" ? (
        <ConnectionForm
          key={engine}
          initialType={engine}
          onSuccess={onSuccess}
          onCancel={onCancel}
        />
      ) : (
        <form
          className="space-y-5"
          onSubmit={(event) => {
            event.preventDefault();
            run(tested);
          }}
        >
          <LayerCard className="rounded-lg p-5 shadow-none sm:p-6">
            <fieldset disabled={action.isPending} className="min-w-0 space-y-5">
              {method === "docker" && (
                <FormField id="add-container" label="Docker container">
                  <div className="flex gap-2">
                    <Select
                      id="add-container"
                      aria-label="Docker container"
                      className="min-w-0 flex-1"
                      placeholder={
                        discovery.isFetching
                          ? "Scanning…"
                          : "Select a container"
                      }
                      items={Object.fromEntries(
                        containers.map((item) => [item.id, item.name]),
                      )}
                      value={container?.id ?? null}
                      disabled={!containers.length || action.isPending}
                      onValueChange={(id) => {
                        const selected = containers.find(
                          (item) => item.id === id,
                        );
                        if (!selected) return;
                        changed();
                        setContainerId(selected.id);
                        setName(selected.name);
                        setDatabase(selected.database);
                      }}
                    >
                      {containers.map((item) => (
                        <Select.Option key={item.id} value={item.id}>
                          {item.name}
                        </Select.Option>
                      ))}
                    </Select>
                    <Button
                      type="button"
                      variant="secondary"
                      disabled={discovery.isFetching}
                      onClick={() => void discovery.refetch()}
                    >
                      Scan again
                    </Button>
                  </div>
                  {!discovery.isFetching && !containers.length && (
                    <p role="status" className="text-sm text-muted-foreground">
                      {discovery.isError || discovery.data?.hint
                        ? "Docker discovery unavailable."
                        : "No matching containers."}
                    </p>
                  )}
                </FormField>
              )}
              <FormField id="add-name" label="Name">
                <FormInput
                  id="add-name"
                  aria-label="Name"
                  required
                  autoComplete="off"
                  value={name}
                  onChange={(event) => {
                    changed();
                    setName(event.target.value);
                  }}
                />
              </FormField>
              {method === "url" ? (
                <FormField id="add-url" label="Connection URL">
                  <FormInput
                    id="add-url"
                    aria-label="Connection URL"
                    type="password"
                    required
                    autoComplete="off"
                    spellCheck={false}
                    placeholder={placeholders[engine]}
                    value={url}
                    onChange={(event) => {
                      changed();
                      setUrl(event.target.value);
                    }}
                  />
                </FormField>
              ) : (
                container && (
                  <FormField id="add-database" label="Database">
                    <FormInput
                      id="add-database"
                      aria-label="Database"
                      value={database}
                      required
                      onChange={(event) => {
                        changed();
                        setDatabase(event.target.value);
                      }}
                    />
                  </FormField>
                )
              )}
            </fieldset>
          </LayerCard>
          {error && (
            <p role="alert" className="text-sm text-destructive">
              {error}
            </p>
          )}
          {tested && (
            <p role="status" className="text-sm text-foreground">
              Connection successful
            </p>
          )}
          <div className="flex flex-wrap justify-end gap-3">
            <Button
              type="button"
              variant="secondary"
              disabled={action.isPending}
              onClick={onCancel}
            >
              Cancel
            </Button>
            <Button
              type="button"
              variant="secondary"
              disabled={!ready || action.isPending}
              onClick={() => run(false)}
            >
              {action.isPending && !action.variables
                ? "Testing…"
                : "Test connection"}
            </Button>
            <Button
              type="submit"
              variant="primary"
              disabled={!ready || !tested || action.isPending}
            >
              {action.isPending && action.variables
                ? "Saving…"
                : "Save connection"}
            </Button>
          </div>
        </form>
      )}
    </div>
  );
}
