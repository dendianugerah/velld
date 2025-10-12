import { Connection, ConnectionForm } from "@/types/connection";
import { apiRequest } from "@/lib/api-client";

export async function getConnections(): Promise<Connection[]> {
  return apiRequest<Connection[]>("/api/connections");
}

export async function getConnection(id: string): Promise<Connection> {
  return apiRequest<Connection>(`/api/connections/${id}`);
}

export async function testConnection(connection: ConnectionForm) {
  return apiRequest("/api/connections/test", {
    method: "POST",
    body: JSON.stringify(connection),
  });
}

export async function saveConnection(connection: ConnectionForm) {
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

export async function deleteConnection(id: string) {
  return apiRequest(`/api/connections/${id}`, {
    method: "DELETE",
  });
}