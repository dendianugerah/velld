import type { Connection, ConnectionForm } from "@/types/connection";
import type { DatabaseType } from "@/types/base";

export const connectionPorts: Record<DatabaseType, number> = { postgresql: 5432, mysql: 3306, mongodb: 27017, d1: 443 };

export function connectionFormValues(connection?: Connection): ConnectionForm {
  return {
    name: connection?.name ?? "", type: connection?.type ?? "postgresql", host: connection?.host ?? "",
    port: connection?.port ?? 5432, username: connection?.username ?? "", password: connection?.password ?? "",
    database: connection?.database_name ?? "", mongo_auth_source: connection?.mongo_auth_source ?? "",
    d1_account_id: connection?.d1_account_id ?? "", d1_database_id: connection?.d1_database_id ?? "",
    ssl: connection?.ssl ?? true, ssh_enabled: connection?.ssh_enabled ?? false,
    ssh_host: connection?.ssh_host ?? "", ssh_port: connection?.ssh_port || 22,
    ssh_username: connection?.ssh_username ?? "", ssh_password: connection?.ssh_password ?? "",
    ssh_private_key: connection?.ssh_private_key ?? "",
  };
}

export function d1ConnectionPayload(values: ConnectionForm): ConnectionForm {
  return {
    name: values.name.trim(), type: "d1", host: "", port: connectionPorts.d1, username: "", password: "", database: "", ssl: true, ssh_enabled: false,
    d1_account_id: values.d1_account_id?.trim(), d1_database_id: values.d1_database_id?.trim(),
    d1_api_token: values.d1_api_token?.trim(),
  };
}

// Submit these values directly: parsing must not depend on a React state update.
export function parseConnectionUrl(value: string): Partial<ConnectionForm> {
  let url: URL;
  try { url = new URL(value.trim()); } catch { throw new Error("Enter a URL like postgres://user:password@host:5432/database."); }
  const types: Record<string, DatabaseType> = { "postgres:": "postgresql", "postgresql:": "postgresql", "mysql:": "mysql", "mongodb:": "mongodb", "mongo:": "mongodb" };
  const type = types[url.protocol];
  if (!type) throw new Error("Use a PostgreSQL, MySQL, or MongoDB URL with an individual host and port.");
  if (!url.hostname) throw new Error("Add a database hostname to the URL.");
  const port = url.port ? Number(url.port) : connectionPorts[type];
  if (!Number.isInteger(port) || port < 1 || port > 65535) throw new Error("Use a port between 1 and 65535.");
  try {
    return {
      type, host: url.hostname.replace(/^\[|\]$/g, ""), port,
      username: decodeURIComponent(url.username), password: decodeURIComponent(url.password),
      database: decodeURIComponent(url.pathname.slice(1)), mongo_auth_source: url.searchParams.get("authSource") ?? "",
      ssl: url.searchParams.get("sslmode") !== "disable" && url.searchParams.get("ssl") !== "false" && url.searchParams.get("tls") !== "false",
    };
  } catch { throw new Error("Check the URL’s percent-encoded username, password, and database name."); }
}

export function connectionError(error: unknown): string {
  if (!(error instanceof Error)) return "Something went wrong. Please try again.";
  try { return (JSON.parse(error.message) as { message?: string }).message || "Check the connection details and try again."; }
  catch { return error.message; }
}
