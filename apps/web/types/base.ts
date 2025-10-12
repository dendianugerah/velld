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

export const statusColors: Record<StatusColor, string> = {
  completed: "bg-emerald-500/15 text-emerald-500 border-emerald-500/20",
  connected: "bg-emerald-500/15 text-emerald-500 border-emerald-500/20",
  pending: "bg-amber-500/15 text-amber-500 border-amber-500/20",
  disconnected: "bg-amber-500/15 text-amber-500 border-amber-500/20",
  failed: "bg-red-500/15 text-red-500 border-red-500/20",
  error: "bg-red-500/15 text-red-500 border-red-500/20",
  running: "bg-blue-500/15 text-blue-500 border-blue-500/20",
};

export type DatabaseType = 'mysql' | 'postgresql' | 'mongodb' | 'redis';

export const typeLabels: Record<DatabaseType, string> = {
  mysql: 'MySQL',
  postgresql: 'PostgreSQL',
  mongodb: 'MongoDB',
  redis: 'Redis',
} as const;