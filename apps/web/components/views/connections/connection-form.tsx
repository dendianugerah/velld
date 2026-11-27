"use client";

import { useState } from "react";
import { Button } from "@cloudflare/kumo/components/button";
import { ChevronDown, Link2 } from "lucide-react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { useToast } from "@/hooks/use-toast";
import { testConnection, saveConnection, updateConnection } from "@/lib/api/connections";
import { connectionError, connectionFormValues, connectionPorts, parseConnectionUrl } from "@/lib/connection-form";
import { typeLabels, type DatabaseType } from "@/types/base";
import type { Connection, ConnectionForm as Values } from "@/types/connection";

export function ConnectionForm({ connection, onSuccess, onCancel }: { connection?: Connection; onSuccess?: (id: string) => void; onCancel?: () => void }) {
  const [values, setValues] = useState(() => connectionFormValues(connection));
  const [method, setMethod] = useState<"manual" | "url">("manual");
  const [url, setUrl] = useState("");
  const [error, setError] = useState("");
  const [auth, setAuth] = useState<"password" | "key">(connection?.ssh_private_key ? "key" : "password");
  const queryClient = useQueryClient();
  const { toast } = useToast();
  const set = <K extends keyof Values>(key: K, value: Values[K]) => setValues(previous => ({ ...previous, [key]: value }));
  const save = useMutation({
    mutationFn: async (data: Values) => {
      await testConnection(data);
      return connection ? updateConnection({ ...data, id: connection.id }) : saveConnection(data);
    },
    onSuccess: result => {
      queryClient.invalidateQueries({ queryKey: ["connections"] });
      if (connection) queryClient.invalidateQueries({ queryKey: ["connection", connection.id] });
      toast({ title: connection ? "Connection updated" : "Connection added", description: "The database connection was tested successfully." });
      onSuccess?.(connection?.id ?? result.id);
    },
    onError: err => setError(connectionError(err)),
  });

  function submit(event: React.FormEvent) {
    event.preventDefault();
    setError("");
    let data = { ...values };
    if (method === "url") {
      try { data = { ...data, ...parseConnectionUrl(url) }; }
      catch (err) { setError(connectionError(err)); return; }
    }
    data.name = data.name.trim(); data.host = data.host.trim();
    if (!data.name || !data.host || !Number.isInteger(data.port) || data.port < 1 || data.port > 65535) {
      setError("Enter a name, host, and a port between 1 and 65535."); return;
    }
    if (data.ssh_enabled && (!data.ssh_host?.trim() || !data.ssh_username?.trim() || !data.ssh_port || !(auth === "key" ? data.ssh_private_key?.trim() : data.ssh_password))) {
      setError("Complete the SSH host, port, username, and authentication details."); return;
    }
    save.mutate(data);
  }

  return <form onSubmit={submit} className="max-w-[720px] space-y-8">
    <fieldset disabled={save.isPending} className="min-w-0 space-y-8 disabled:opacity-70">
      <section aria-labelledby="identity-heading">
        <h2 id="identity-heading" className="text-sm font-semibold">Connection details</h2>
        <p className="mt-1.5 text-[13px] text-muted-foreground">Velld tests the connection before saving it.</p>
        <div className="mt-6 space-y-5">
          <Field id="connection-name" label="Name"><Input id="connection-name" required autoComplete="off" placeholder="e.g. Production database" value={values.name} onChange={event => set("name", event.target.value)} /></Field>
          <div className="flex flex-wrap items-center gap-1 rounded-lg bg-muted p-1" role="group" aria-label="Connection input method">
            {(["manual", "url"] as const).map(value => <button key={value} type="button" aria-pressed={method === value} onClick={() => { setMethod(value); setError(""); }} className={`flex min-h-9 flex-1 items-center justify-center gap-2 rounded-md px-3 text-[13px] font-medium focus-visible:outline focus-visible:outline-2 focus-visible:outline-ring ${method === value ? "bg-card shadow-sm" : "text-muted-foreground hover:text-foreground"}`}>{value === "url" && <Link2 className="h-3.5 w-3.5" aria-hidden="true" />}{value === "manual" ? "Enter details" : "Connection URL"}</button>)}
          </div>
          {method === "url" ? <Field id="connection-url" label="Connection URL" hint="PostgreSQL, MySQL, or MongoDB. TLS and MongoDB authSource options are read when you save."><Input id="connection-url" type="password" required autoComplete="off" spellCheck={false} placeholder="postgres://user:password@host:5432/database" value={url} onChange={event => setUrl(event.target.value)} /></Field> : <>
            <Field id="connection-engine" label="Database engine"><select id="connection-engine" value={values.type} onChange={event => { const type = event.target.value as DatabaseType; setValues(previous => ({ ...previous, type, port: connectionPorts[type] })); }} className="h-10 w-full rounded-lg border border-input bg-card px-3 text-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-ring">{Object.entries(typeLabels).map(([type, label]) => <option key={type} value={type}>{label}</option>)}</select></Field>
            <div className="grid gap-5 sm:grid-cols-[minmax(0,1fr)_140px]">
              <Field id="connection-host" label="Host"><Input id="connection-host" required autoComplete="off" placeholder="db.example.com" value={values.host} onChange={event => set("host", event.target.value)} /></Field>
              <Field id="connection-port" label="Port"><Input id="connection-port" type="number" min={1} max={65535} required value={values.port || ""} onChange={event => set("port", event.target.valueAsNumber)} /></Field>
            </div>
            <div className="grid gap-5 sm:grid-cols-2">
              <Field id="connection-username" label={`Username${values.type === "mongodb" ? " (optional)" : ""}`}><Input id="connection-username" required={values.type !== "mongodb"} autoComplete="off" value={values.username} onChange={event => set("username", event.target.value)} /></Field>
              <Field id="connection-password" label={`Password${values.type === "mongodb" ? " (optional)" : ""}`}><Input id="connection-password" type="password" autoComplete="new-password" value={values.password} onChange={event => set("password", event.target.value)} /></Field>
            </div>
            <Field id="connection-database" label="Database (optional)" hint="You can discover and select databases after saving."><Input id="connection-database" placeholder={values.type === "postgresql" ? "postgres" : values.type === "mongodb" ? "admin" : "Leave empty to discover databases"} value={values.database} onChange={event => set("database", event.target.value)} /></Field>
            {values.type === "mongodb" && <Field id="mongo-auth-source" label="Authentication database (optional)" hint="The database where your MongoDB user was created."><Input id="mongo-auth-source" placeholder="admin" value={values.mongo_auth_source} onChange={event => set("mongo_auth_source", event.target.value)} /></Field>}
            <div className="flex items-start justify-between gap-5 rounded-lg border p-4"><div><Label htmlFor="connection-ssl">Use TLS</Label><p className="mt-1.5 text-xs leading-relaxed text-muted-foreground">Encrypt traffic between Velld and your database.</p></div><Switch id="connection-ssl" checked={values.ssl} onCheckedChange={value => set("ssl", value)} /></div>
          </>}
        </div>
      </section>
      <details className="group rounded-lg border" open={values.ssh_enabled || undefined}>
        <summary className="flex cursor-pointer list-none items-center justify-between gap-4 rounded-lg p-4 text-sm font-medium focus-visible:outline focus-visible:outline-2 focus-visible:outline-ring">SSH tunnel <span className="ms-auto text-xs font-normal text-muted-foreground">{values.ssh_enabled ? "Enabled" : "Optional"}</span><ChevronDown className="h-4 w-4 text-muted-foreground group-open:rotate-180" aria-hidden="true" /></summary>
        <div className="space-y-5 border-t p-4 sm:p-5">
          <div className="flex items-start justify-between gap-5"><div><Label htmlFor="ssh-enabled">Connect through SSH</Label><p className="mt-1.5 text-xs leading-relaxed text-muted-foreground">Use a jump host to reach a private database.</p></div><Switch id="ssh-enabled" checked={values.ssh_enabled} onCheckedChange={value => set("ssh_enabled", value)} /></div>
          {values.ssh_enabled && <>
            <div className="grid gap-5 sm:grid-cols-[minmax(0,1fr)_140px]"><Field id="ssh-host" label="SSH host"><Input id="ssh-host" required value={values.ssh_host} onChange={event => set("ssh_host", event.target.value)} /></Field><Field id="ssh-port" label="SSH port"><Input id="ssh-port" type="number" min={1} max={65535} required value={values.ssh_port || ""} onChange={event => set("ssh_port", event.target.valueAsNumber)} /></Field></div>
            <Field id="ssh-user" label="SSH username"><Input id="ssh-user" required autoComplete="off" value={values.ssh_username} onChange={event => set("ssh_username", event.target.value)} /></Field>
            <Field id="ssh-auth" label="Authentication"><select id="ssh-auth" value={auth} onChange={event => { const value = event.target.value as "password" | "key"; setAuth(value); setValues(previous => ({ ...previous, ssh_password: value === "key" ? "" : previous.ssh_password, ssh_private_key: value === "password" ? "" : previous.ssh_private_key })); }} className="h-10 w-full rounded-lg border border-input bg-card px-3 text-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-ring"><option value="password">Password</option><option value="key">Private key</option></select></Field>
            {auth === "password" ? <Field id="ssh-password" label="SSH password"><Input id="ssh-password" required type="password" autoComplete="new-password" value={values.ssh_password} onChange={event => set("ssh_password", event.target.value)} /></Field> : <Field id="ssh-key" label="SSH private key"><textarea id="ssh-key" required rows={6} spellCheck={false} value={values.ssh_private_key} onChange={event => set("ssh_private_key", event.target.value)} className="w-full rounded-lg border border-input bg-card p-3 font-mono text-xs focus-visible:outline focus-visible:outline-2 focus-visible:outline-ring" /></Field>}
          </>}
        </div>
      </details>
    </fieldset>
    {error && <p role="alert" className="rounded-lg border border-destructive/20 bg-destructive/5 p-4 text-sm text-destructive">{error}</p>}
    <div className="flex flex-wrap gap-3 border-t pt-6"><Button type="submit" disabled={save.isPending}>{save.isPending ? "Testing and saving…" : connection ? "Save changes" : "Add connection"}</Button><Button type="button" variant="secondary" onClick={onCancel} disabled={save.isPending}>Cancel</Button></div>
  </form>;
}

function Field({ id, label, hint, children }: { id: string; label: string; hint?: string; children: React.ReactNode }) {
  return <div className="min-w-0 space-y-2"><Label htmlFor={id}>{label}</Label>{children}{hint && <p className="text-xs leading-relaxed text-muted-foreground">{hint}</p>}</div>;
}
