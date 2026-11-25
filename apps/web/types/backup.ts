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
}

export type BackupList = Backup;

export type BackupListResponse = Base<BackupList[]>;