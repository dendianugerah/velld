import { BackupListResponse } from '@/types/backup';
import { apiRequest } from '../api-client';

export interface GetBackupsParams {
  page?: number;
  limit?: number;
  search?: string;
}

export interface RestoreBackupParams {
  backup_id: string;
  connection_id: string;
}

export async function saveBackup(connectionId: string): Promise<void> {
  return apiRequest('/api/backups', {
    method: 'POST',
    body: JSON.stringify({ connection_id: connectionId }),
  });
}

export async function getBackups(params?: GetBackupsParams): Promise<BackupListResponse> {
  const query = new URLSearchParams();
  if (params?.page) query.set("page", String(params.page));
  if (params?.limit) query.set("limit", String(params.limit));
  if (params?.search) query.set("search", params.search);
  const url = `/api/backups${query.size ? `?${query}` : ""}`;

  return apiRequest<BackupListResponse>(url, {
    method: 'GET',
  });
}

export async function downloadBackup(backupId: string): Promise<Blob> {
  return apiRequest<Blob>(`/api/backups/${backupId}/download`, {
    method: 'GET',
    responseType: 'blob',
  });
}

export interface ScheduleBackupParams {
  connection_id: string;
  cron_schedule: string;
  retention_days: number;
}

export async function scheduleBackup(params: ScheduleBackupParams): Promise<void> {
  return apiRequest('/api/backups/schedule', {
    method: 'POST',
    body: JSON.stringify(params),
  });
}

export async function updateSchedule(connectionId: string, params: Omit<ScheduleBackupParams, 'connection_id'>): Promise<void> {
  return apiRequest(`/api/backups/${connectionId}/schedule`, {
    method: 'PUT',
    body: JSON.stringify(params),
  });
}

export async function disableBackupSchedule(connectionId: string): Promise<void> {
  return apiRequest(`/api/backups/${connectionId}/schedule/disable`, {
    method: 'POST',
  });
}

export async function restoreBackup(params: RestoreBackupParams): Promise<void> {
  return apiRequest('/api/backups/restore', {
    method: 'POST',
    body: JSON.stringify(params),
  });
}

export interface BackupBranch {
  connection_string: string;
  expires_at: string;
}

export async function createBackupBranch(backupId: string, signal: AbortSignal): Promise<BackupBranch> {
  const response = await apiRequest<{ data: BackupBranch }>(`/api/backups/${encodeURIComponent(backupId)}/branch`, {
    method: 'POST',
    signal,
  });
  return response.data;
}

export interface PG18Report {
  target_version: string;
  source_version?: string;
  status: "not_run" | "pending" | "passed" | "failed";
  message?: string;
  duration_ms: number;
  issues: { category: string; severity: "error" | "warning"; message: string }[];
  omitted_issues?: number;
}

export async function getPG18Report(backupId: string, method: "GET" | "POST", signal?: AbortSignal): Promise<PG18Report> {
  const response = await apiRequest<{ data: PG18Report }>(
    `/api/backups/${encodeURIComponent(backupId)}/pg18-rehearsal`, { method, signal },
  );
  return response.data;
}

export async function compareBackupSchema(
  backupId: string,
  target: { target_backup_id: string } | { target_live: true },
  signal: AbortSignal,
): Promise<string> {
  const response = await apiRequest<{ data: { migration_sql: string } }>(`/api/backups/${encodeURIComponent(backupId)}/schema-diff`, {
    method: 'POST',
    body: JSON.stringify(target),
    signal,
  });
  return response.data.migration_sql;
}

export async function deleteBackup(backupId: string): Promise<void> {
  return apiRequest(`/api/backups/${backupId}`, {
    method: 'DELETE',
  });
}