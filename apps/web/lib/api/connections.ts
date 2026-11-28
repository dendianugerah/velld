import { Connection, ConnectionForm } from "@/types/connection";
import { apiRequest } from "@/lib/api-client";

export interface DockerDatabase {
  id: string;
  name: string;
  type: Connection['type'];
  host: string;
  port: number;
  database: string;
}

export function discoverDockerDatabases() {
  return apiRequest<{ containers: DockerDatabase[]; hint?: string }>('/api/connections/docker');
}

type DockerConnectionInput = { database: string; name: string };

export function importDockerDatabase(id: string, input: DockerConnectionInput) {
  return apiRequest<{ id: string }>(`/api/connections/docker/${encodeURIComponent(id)}`, {
    method: 'POST',
    body: JSON.stringify(input),
  });
}

export function testDockerDatabase(id: string, input: DockerConnectionInput) {
  return apiRequest(`/api/connections/docker/${encodeURIComponent(id)}/test`, {
    method: 'POST',
    body: JSON.stringify(input),
  });
}

export async function getConnections(): Promise<Connection[]> {
  return apiRequest<Connection[]>("/api/connections");
}

export async function getConnection(id: string): Promise<Connection> {
  return apiRequest<Connection>(`/api/connections/${id}`);
}

type ConnectionInput = ConnectionForm | { name: string; type: Connection['type']; connection_url: string };

export async function testConnection(connection: ConnectionInput) {
  const result = await apiRequest<{ isConnected: boolean; error?: string }>("/api/connections/test", {
    method: "POST",
    body: JSON.stringify(connection),
  });
  if (!result.isConnected) throw new Error(result.error || "Couldn’t connect.");
  return result;
}

export async function saveConnection(connection: ConnectionInput) {
  return apiRequest<Connection>("/api/connections", {
    method: "POST",
    body: JSON.stringify(connection),
  });
}

export async function updateConnection(connection: ConnectionForm & { id: string }) {
  return apiRequest<Connection>("/api/connections", {
    method: "PUT",
    body: JSON.stringify(connection),
  });
}

export async function deleteConnection(params: { id: string; cleanupS3?: boolean }) {
  const url = `/api/connections/${params.id}${params.cleanupS3 ? '?cleanup_s3=true' : ''}`;
  return apiRequest(url, {
    method: "DELETE",
  });
}

export async function discoverDatabases(id: string): Promise<{ databases: string[] }> {
  return apiRequest<{ databases: string[] }>(`/api/connections/${id}/discover`);
}

export async function updateSelectedDatabases(id: string, databases: string[]) {
  return apiRequest(`/api/connections/${id}/databases`, {
    method: "PUT",
    body: JSON.stringify({ databases }),
  });
}

export async function updateConnectionSettings(id: string, settings: { s3_cleanup_on_retention?: boolean }) {
  return apiRequest(`/api/connections/${id}/settings`, {
    method: "POST",
    body: JSON.stringify(settings),
  });
}

export interface UpgradeInfo {
  supported: boolean;
  container_id?: string;
  container_name?: string;
  current_image?: string;
  database_type: string;
  hint?: string;
}

export interface UpgradeResult {
  status: "upgraded" | "rolled_back";
  container_id: string;
  container_name: string;
  previous_image: string;
  target_image: string;
  previous_version?: string;
  current_version?: string;
  snapshot_backup_id: string;
  health_check?: string;
  rolled_back: boolean;
  rollback_verified?: boolean;
  error?: string;
}

export async function getConnectionUpgradeInfo(id: string, signal?: AbortSignal): Promise<UpgradeInfo> {
  const { data } = await apiRequest<{ data: UpgradeInfo }>(`/api/connections/${encodeURIComponent(id)}/upgrade`, {
    signal,
  });
  return data;
}

export async function upgradeConnectionContainer(id: string, targetImage: string, signal?: AbortSignal): Promise<UpgradeResult> {
  const { data } = await apiRequest<{ data: UpgradeResult }>(`/api/connections/${encodeURIComponent(id)}/upgrade`, {
    method: "POST",
    body: JSON.stringify({ target_image: targetImage }),
    signal,
  });
  return data;
}