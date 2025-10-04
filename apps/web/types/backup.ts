import { Base } from './base';

export interface Backup {
  id: string;
  connection_id: string;
  database_type: string;
  database_name: string;
  schedule_id?: string;
  size: number;
  status: string;
  path: string;
  scheduled_time: string;
  started_time: string;
  completed_time: string;
  created_at: string;
  updated_at: string;
}

export interface BackupStats {
  total_backups: number;
  failed_backups: number;
  total_size: number;
  average_duration: number;
  success_rate: number;
}

export interface BackupStatsResponse extends Base<BackupStats> {
  data: BackupStats;
}

export type BackupList = Backup;

export type BackupListResponse = Base<BackupList[]>;

export interface DiffChange {
  type: string;
  content: string;
  line_number: number;
  old_line?: number;
  new_line?: number;
}

export interface BackupDiff {
  added: number;
  removed: number;
  modified: number;
  unchanged: number;
  changes: DiffChange[];
}

export type BackupDiffResponse = Base<BackupDiff>;