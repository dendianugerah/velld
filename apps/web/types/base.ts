export interface Pagination {
  total: number;
  page: number;
  limit: number;
  total_pages: number;
}

export interface Base<T> {
  data: T;
  message: string;
  pagination?: Pagination;
}

export type StatusColor = 'completed' | 'pending' | 'failed' | 'running' | 'connected' | 'disconnected' | 'error';

export type DatabaseType = 'mysql' | 'postgresql' | 'mongodb' | 'd1';

export const typeLabels: Record<DatabaseType, string> = {
  mysql: 'MySQL',
  postgresql: 'PostgreSQL',
  mongodb: 'MongoDB',
  d1: 'Cloudflare D1',
} as const;