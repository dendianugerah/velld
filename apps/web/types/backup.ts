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
  s3_object_key?: string;
  scheduled_time: string;
  started_time: string;
  completed_time: string;
  created_at: string;
  updated_at: string;
  verification_status?: string;
}

export type VerificationStatus = "running" | "passed" | "failed" | "";

export interface BackupVerification {
  id: string;
  backup_id: string;
  status: VerificationStatus;
  method: string;
  checksum_matches: boolean;
  details?: string;
  started_at: string;
  completed_at?: string;
  created_at: string;
}

export interface BackupVerificationResponse extends Base<BackupVerification> {
  data: BackupVerification;
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