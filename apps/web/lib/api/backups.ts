import { BackupListResponse, BackupStatsResponse, BackupDiffResponse } from '@/types/backup';
import { apiRequest } from '../api-client';

export interface GetBackupsParams {
  page?: number;
  limit?: number;
  search?: string;
}

export async function saveBackup(connectionId: string): Promise<void> {
  return apiRequest('/api/backups', {
    method: 'POST',
    body: JSON.stringify({ connection_id: connectionId }),
  });
}

export async function getBackups(params?: GetBackupsParams): Promise<BackupListResponse> {
  let url = `/api/backups`;

  if (params?.page && params?.limit) {
    url += `?page=${params.page}&limit=${params.limit}`;
  }

  if (params?.search) {
    url += `&search=${params.search}`;
  }

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

export async function getBackupStats(): Promise<BackupStatsResponse> {
  return apiRequest<BackupStatsResponse>('/api/backups/stats', {
    method: 'GET',
  });
}

export async function compareBackups(sourceId: string, targetId: string): Promise<BackupDiffResponse> {
  return apiRequest<BackupDiffResponse>(`/api/backups/compare/${sourceId}/${targetId}`, {
    method: 'GET',
  });
}