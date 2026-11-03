import { apiRequest } from '../api-client';

export interface WALStatus {
  segment_count: number;
  gaps: string[];
  has_gaps: boolean;
}

export interface WALCommandConfig {
  archive_command: string;
  restore_command: string;
  wal_prefix: string;
}

export interface PITRPrepareRequest {
  connection_id: string;
  target_time: string;
  data_dir: string;
  restore_command: string;
}

export interface PITRPrepareResult {
  data_dir: string;
  target_time: string;
  recovery_file: string;
}

interface Wrapped<T> {
  data: T;
}

export async function createBaseBackup(connectionId: string): Promise<{ object_key: string }> {
  const res = await apiRequest<Wrapped<{ object_key: string }>>('/api/backups/base', {
    method: 'POST',
    body: JSON.stringify({ connection_id: connectionId }),
  });
  return res.data;
}

export async function getWALStatus(connectionId: string): Promise<WALStatus> {
  const res = await apiRequest<Wrapped<WALStatus>>(`/api/wal/${connectionId}/status`, {
    method: 'GET',
  });
  return res.data;
}

export async function getWALCommandConfig(connectionId: string, apiBaseUrl: string, archivePath: string): Promise<WALCommandConfig> {
  const params = new URLSearchParams({ api_base_url: apiBaseUrl, archive_path: archivePath });
  const res = await apiRequest<Wrapped<WALCommandConfig>>(`/api/wal/${connectionId}/config?${params.toString()}`, {
    method: 'GET',
  });
  return res.data;
}

export async function preparePITR(req: PITRPrepareRequest): Promise<PITRPrepareResult> {
  const res = await apiRequest<Wrapped<PITRPrepareResult>>('/api/pitr/prepare', {
    method: 'POST',
    body: JSON.stringify(req),
  });
  return res.data;
}
