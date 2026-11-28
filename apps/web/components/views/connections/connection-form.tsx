"use client";

import { useState } from "react";
import { Button } from "@cloudflare/kumo/components/button";
import { Select } from "@cloudflare/kumo/components/select";
import { LayerCard } from "@cloudflare/kumo/components/layer-card";
import { ChevronDown, Link2 } from "lucide-react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import {
  FormInput as Input,
  FormField as Field,
} from "@/components/ui/form-field";
import { Label } from "@/components/ui/label";
import { SegmentedControl } from "@/components/ui/segmented-control";
import { Switch } from "@/components/ui/switch";
import { useToast } from "@/hooks/use-toast";
import {
  testConnection,
  saveConnection,
  updateConnection,
} from "@/lib/api/connections";
import {
  connectionError,
  connectionFormValues,
  connectionPorts,
  d1ConnectionPayload,
  parseConnectionUrl,
} from "@/lib/connection-form";
import { typeLabels, type DatabaseType } from "@/types/base";
import type { Connection, ConnectionForm as Values } from "@/types/connection";

export function ConnectionForm({
  connection,
  initialType,
  onSuccess,
  onCancel,
}: {
  connection?: Connection;
  initialType?: DatabaseType;
  onSuccess?: (id: string) => void;
  onCancel?: () => void;
}) {
  const [values, setValues] = useState(() => ({
    ...connectionFormValues(connection),
    ...(initialType && {
      type: initialType,
      port: connectionPorts[initialType],
    }),
  }));
  const [editMethod, setEditMethod] = useState<"manual" | "url">("manual");
  const isD1 = values.type === "d1";
  const method = isD1 ? "manual" : editMethod;
  const engines = Object.entries(typeLabels).filter(([type]) =>
    initialType
      ? type === initialType
      : !connection || (type === "d1") === (connection.type === "d1"),
  );
  const [url, setUrl] = useState("");
  const [error, setError] = useState("");
  const [auth, setAuth] = useState<"password" | "key">(
    connection?.ssh_private_key ? "key" : "password",
  );
  const queryClient = useQueryClient();
  const { toast } = useToast();
  const set = <K extends keyof Values>(key: K, value: Values[K]) =>
    setValues((previous) => ({ ...previous, [key]: value }));
  const save = useMutation({
    mutationFn: async (data: Values) => {
      if (data.type !== "d1") await testConnection(data);
      return connection
        ? updateConnection({ ...data, id: connection.id })
        : saveConnection(data);
    },
    onSuccess: (result) => {
      queryClient.invalidateQueries({ queryKey: ["connections"] });
      if (connection)
        queryClient.invalidateQueries({
          queryKey: ["connection", connection.id],
        });
      toast({
        title: connection ? "Connection updated" : "Connection added",
      });
      onSuccess?.(connection?.id ?? result.id);
    },
    onError: (err) => setError(connectionError(err)),
  });

  function submit(event: React.FormEvent) {
    event.preventDefault();
    setError("");
    let data = { ...values };
    if (method === "url") {
      try {
        data = { ...data, ...parseConnectionUrl(url) };
      } catch (err) {
        setError(connectionError(err));
        return;
      }
    }
    data.name = data.name.trim();
    if (data.type === "d1") {
      if (!data.name) {
        setError("Enter a connection name.");
        return;
      }
      save.mutate(d1ConnectionPayload(data));
      return;
    }
    delete data.d1_account_id;
    delete data.d1_database_id;
    delete data.d1_api_token;
    data.host = data.host.trim();
    if (
      !data.name ||
      !data.host ||
      !Number.isInteger(data.port) ||
      data.port < 1 ||
      data.port > 65535
    ) {
      setError("Enter a name, host, and a port between 1 and 65535.");
      return;
    }
    if (
      data.ssh_enabled &&
      (!data.ssh_host?.trim() ||
        !data.ssh_username?.trim() ||
        !data.ssh_port ||
        !(auth === "key" ? data.ssh_private_key?.trim() : data.ssh_password))
    ) {
      setError(
        "Complete the SSH host, port, username, and authentication details.",
      );
      return;
    }
    save.mutate(data);
  }

  return (
    <form onSubmit={submit} className="max-w-[720px] space-y-6">
      <fieldset
        disabled={save.isPending}
        className="min-w-0 space-y-5 disabled:opacity-70"
      >
        <div className="max-w-lg">
          <Field id="connection-name" label="Name">
            <Input
              id="connection-name"
              aria-label="Name"
              required
              autoComplete="off"
              placeholder="e.g. Production database"
              value={values.name}
              onChange={(event) => set("name", event.target.value)}
            />
          </Field>
        </div>
        <LayerCard className="rounded-lg p-5 shadow-none sm:p-6">
          <section aria-label="Connection details" className="space-y-5">
            {connection && !isD1 && (
              <SegmentedControl
                label="Connection input method"
                value={method}
                onChange={(value) => {
                  setEditMethod(value);
                  setError("");
                }}
                options={[
                  { value: "manual", label: "Fields" },
                  { value: "url", label: "Connection URL", icon: Link2 },
                ]}
              />
            )}
            {method === "url" ? (
              <Field id="connection-url" label="Connection URL">
                <Input
                  id="connection-url"
                  aria-label="Connection URL"
                  type="password"
                  required
                  autoComplete="off"
                  spellCheck={false}
                  placeholder="postgres://user:password@host:5432/database"
                  value={url}
                  onChange={(event) => setUrl(event.target.value)}
                />
              </Field>
            ) : (
              <>
                {!initialType && (
                  <Field id="connection-engine" label="Database engine">
                    <Select
                      id="connection-engine"
                      aria-label="Database engine"
                      value={values.type}
                      onValueChange={(value) => {
                        if (!value) return;
                        const type = value as DatabaseType;
                        setValues((previous) => ({
                          ...previous,
                          type,
                          port: connectionPorts[type],
                        }));
                      }}
                      items={Object.fromEntries(engines)}
                      className="w-full"
                    >
                      {engines.map(([type, label]) => (
                        <Select.Option key={type} value={type}>
                          {label}
                        </Select.Option>
                      ))}
                    </Select>
                  </Field>
                )}
                {isD1 ? (
                  <>
                    <Field id="d1-account-id" label="Cloudflare account ID">
                      <Input
                        id="d1-account-id"
                        aria-label="Cloudflare account ID"
                        required
                        autoComplete="off"
                        spellCheck={false}
                        value={values.d1_account_id ?? ""}
                        onChange={(event) =>
                          set("d1_account_id", event.target.value)
                        }
                      />
                    </Field>
                    <Field id="d1-database-id" label="D1 database ID">
                      <Input
                        id="d1-database-id"
                        aria-label="D1 database ID"
                        required
                        autoComplete="off"
                        spellCheck={false}
                        value={values.d1_database_id ?? ""}
                        onChange={(event) =>
                          set("d1_database_id", event.target.value)
                        }
                      />
                    </Field>
                    <Field
                      id="d1-api-token"
                      label={
                        connection ? "New API token (optional)" : "API token"
                      }
                    >
                      <Input
                        id="d1-api-token"
                        aria-label={
                          connection ? "New API token (optional)" : "API token"
                        }
                        type="password"
                        required={!connection}
                        autoComplete="new-password"
                        value={values.d1_api_token ?? ""}
                        onChange={(event) =>
                          set("d1_api_token", event.target.value)
                        }
                      />
                    </Field>
                  </>
                ) : (
                  <>
                    <div className="grid gap-5 sm:grid-cols-[minmax(0,1fr)_140px]">
                      <Field id="connection-host" label="Host">
                        <Input
                          id="connection-host"
                          aria-label="Host"
                          required
                          autoComplete="off"
                          placeholder="db.example.com"
                          value={values.host}
                          onChange={(event) => set("host", event.target.value)}
                        />
                      </Field>
                      <Field id="connection-port" label="Port">
                        <Input
                          id="connection-port"
                          aria-label="Port"
                          type="number"
                          min={1}
                          max={65535}
                          required
                          value={values.port || ""}
                          onChange={(event) =>
                            set("port", event.target.valueAsNumber)
                          }
                        />
                      </Field>
                    </div>
                    <div className="grid gap-5 sm:grid-cols-2">
                      <Field
                        id="connection-username"
                        label={`Username${values.type === "mongodb" ? " (optional)" : ""}`}
                      >
                        <Input
                          id="connection-username"
                          aria-label={`Username${values.type === "mongodb" ? " (optional)" : ""}`}
                          required={values.type !== "mongodb"}
                          autoComplete="off"
                          value={values.username}
                          onChange={(event) =>
                            set("username", event.target.value)
                          }
                        />
                      </Field>
                      <Field
                        id="connection-password"
                        label={`Password${values.type === "mongodb" ? " (optional)" : ""}`}
                      >
                        <Input
                          id="connection-password"
                          aria-label={`Password${values.type === "mongodb" ? " (optional)" : ""}`}
                          type="password"
                          autoComplete="new-password"
                          value={values.password}
                          onChange={(event) =>
                            set("password", event.target.value)
                          }
                        />
                      </Field>
                    </div>
                    <Field id="connection-database" label="Database (optional)">
                      <Input
                        id="connection-database"
                        aria-label="Database (optional)"
                        placeholder={
                          values.type === "postgresql"
                            ? "postgres"
                            : values.type === "mongodb"
                              ? "admin"
                              : "Leave empty to discover databases"
                        }
                        value={values.database}
                        onChange={(event) =>
                          set("database", event.target.value)
                        }
                      />
                    </Field>
                    {values.type === "mongodb" && (
                      <Field
                        id="mongo-auth-source"
                        label="Authentication database (optional)"
                      >
                        <Input
                          id="mongo-auth-source"
                          aria-label="Authentication database (optional)"
                          placeholder="admin"
                          value={values.mongo_auth_source}
                          onChange={(event) =>
                            set("mongo_auth_source", event.target.value)
                          }
                        />
                      </Field>
                    )}
                    <div className="flex items-center justify-between gap-5 border-t pt-5">
                      <Label htmlFor="connection-ssl">Use TLS</Label>
                      <Switch
                        id="connection-ssl"
                        aria-label="Use TLS"
                        checked={values.ssl}
                        onCheckedChange={(value) => set("ssl", value)}
                      />
                    </div>
                  </>
                )}
              </>
            )}
          </section>
          {(!isD1 || method === "url") && (
            <details
              className="group mt-6 rounded-lg border"
              open={values.ssh_enabled || undefined}
            >
              <summary className="flex cursor-pointer list-none items-center gap-3 rounded-lg px-4 py-3 text-sm font-medium focus-visible:outline focus-visible:outline-2 focus-visible:outline-ring">
                SSH tunnel{" "}
                {values.ssh_enabled && (
                  <span className="text-xs font-normal text-muted-foreground">
                    On
                  </span>
                )}
                <ChevronDown
                  className="ms-auto h-4 w-4 text-muted-foreground group-open:rotate-180"
                  aria-hidden="true"
                />
              </summary>
              <div className="space-y-5 border-t p-4 sm:p-5">
                <div className="flex items-center justify-between gap-5">
                  <Label htmlFor="ssh-enabled">Connect through SSH</Label>
                  <Switch
                    id="ssh-enabled"
                    aria-label="Connect through SSH"
                    checked={values.ssh_enabled}
                    onCheckedChange={(value) => set("ssh_enabled", value)}
                  />
                </div>
                {values.ssh_enabled && (
                  <>
                    <div className="grid gap-5 sm:grid-cols-[minmax(0,1fr)_140px]">
                      <Field id="ssh-host" label="SSH host">
                        <Input
                          id="ssh-host"
                          aria-label="SSH host"
                          required
                          value={values.ssh_host}
                          onChange={(event) =>
                            set("ssh_host", event.target.value)
                          }
                        />
                      </Field>
                      <Field id="ssh-port" label="SSH port">
                        <Input
                          id="ssh-port"
                          aria-label="SSH port"
                          type="number"
                          min={1}
                          max={65535}
                          required
                          value={values.ssh_port || ""}
                          onChange={(event) =>
                            set("ssh_port", event.target.valueAsNumber)
                          }
                        />
                      </Field>
                    </div>
                    <Field id="ssh-user" label="SSH username">
                      <Input
                        id="ssh-user"
                        aria-label="SSH username"
                        required
                        autoComplete="off"
                        value={values.ssh_username}
                        onChange={(event) =>
                          set("ssh_username", event.target.value)
                        }
                      />
                    </Field>
                    <Field id="ssh-auth" label="Authentication">
                      <Select
                        id="ssh-auth"
                        aria-label="Authentication"
                        value={auth}
                        onValueChange={(value) => {
                          const next = value as "password" | "key";
                          setAuth(next);
                          setValues((previous) => ({
                            ...previous,
                            ssh_password:
                              next === "key" ? "" : previous.ssh_password,
                            ssh_private_key:
                              next === "password"
                                ? ""
                                : previous.ssh_private_key,
                          }));
                        }}
                        items={{ password: "Password", key: "Private key" }}
                        className="w-full"
                      >
                        <Select.Option value="password">Password</Select.Option>
                        <Select.Option value="key">Private key</Select.Option>
                      </Select>
                    </Field>
                    {auth === "password" ? (
                      <Field id="ssh-password" label="SSH password">
                        <Input
                          id="ssh-password"
                          aria-label="SSH password"
                          required
                          type="password"
                          autoComplete="new-password"
                          value={values.ssh_password}
                          onChange={(event) =>
                            set("ssh_password", event.target.value)
                          }
                        />
                      </Field>
                    ) : (
                      <Field id="ssh-key" label="SSH private key">
                        <textarea
                          id="ssh-key"
                          required
                          rows={6}
                          spellCheck={false}
                          value={values.ssh_private_key}
                          onChange={(event) =>
                            set("ssh_private_key", event.target.value)
                          }
                          className="w-full rounded-lg border border-input bg-card p-3 font-mono text-xs focus-visible:outline focus-visible:outline-2 focus-visible:outline-ring"
                        />
                      </Field>
                    )}
                  </>
                )}
              </div>
            </details>
          )}
        </LayerCard>
      </fieldset>
      {error && (
        <p
          role="alert"
          className="rounded-lg border border-destructive/20 bg-destructive/5 p-4 text-sm text-destructive"
        >
          {error}
        </p>
      )}
      <div className="flex flex-wrap justify-end gap-3">
        <Button
          type="button"
          variant="secondary"
          onClick={onCancel}
          disabled={save.isPending}
        >
          Cancel
        </Button>
        <Button type="submit" variant="primary" disabled={save.isPending}>
          {save.isPending
            ? "Testing and saving…"
            : connection
              ? "Save changes"
              : "Add connection"}
        </Button>
      </div>
    </form>
  );
}
