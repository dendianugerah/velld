import { Base, DatabaseType, StatusColor } from "./base";

export interface Connection {
  id: string;
  name: string;
  type: DatabaseType;
  host: string;
  port: number;
  username: string;
  password: string;
  database: string;
  database_name: string;
  mongo_auth_source?: string;
  d1_account_id?: string;
  d1_database_id?: string;
  database_size: number;
  selected_databases?: string[];
  ssl: boolean;
  ssh_enabled: boolean;
  ssh_host?: string;
  ssh_port?: number;
  ssh_username?: string;
  ssh_password?: string;
  ssh_private_key?: string;
  status: StatusColor;
  last_backup_time?: string;
  backup_enabled: boolean;
  cron_schedule?: string;
  retention_days?: number;
  retention_count?: number | null;
  s3_cleanup_on_retention: boolean;
}

export type ConnectionForm = Pick<Connection, 
  | "name" 
  | "type" 
  | "host" 
  | "port" 
  | "username" 
  | "password" 
  | "database" 
  | "mongo_auth_source"
  | "d1_account_id"
  | "d1_database_id"
  | "ssl"
  | "ssh_enabled"
  | "ssh_host"
  | "ssh_port"
  | "ssh_username"
  | "ssh_password"
  | "ssh_private_key"
> & {
  d1_api_token?: string;
  s3_cleanup_on_retention?: boolean;
};

export type ConnectionListResponse = Base<Connection[]>;

export type SortBy = 'name' | 'status' | 'type' | 'lastBackup';

export interface BackupConfig {
  enabled: boolean;
  schedule: string;
  retention: string;
  lastBackup?: string;
}