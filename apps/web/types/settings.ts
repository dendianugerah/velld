import { Base } from "./base";

export interface UserSettings {
  id: string;
  user_id: string;
  notify_dashboard: boolean;
  notify_email: boolean;
  notify_webhook: boolean;
  webhook_url?: string;
  email?: string;
  smtp_host?: string;
  smtp_port?: number;
  smtp_username?: string;
  smtp_password?: string;
  // S3 Storage settings
  s3_enabled: boolean;
  s3_endpoint?: string;
  s3_region?: string;
  s3_bucket?: string;
  s3_access_key?: string;
  s3_secret_key?: string;
  s3_use_ssl: boolean;
  s3_path_prefix?: string;
  s3_purge_local: boolean;
  env_configured?: Record<string, boolean>;
}

export type UpdateSettingsRequest = Partial<Omit<UserSettings, 'id' | 'user_id'>>;

export type SettingsResponse = Base<UserSettings>;
export type GetSettingsResponse = SettingsResponse;
export type UpdateSettingsResponse = SettingsResponse;
