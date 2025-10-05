import { Card } from "@/components/ui/card";
import { ArrowUpIcon, ArrowDownIcon } from "lucide-react";
import { useBackup } from "@/hooks/use-backup";
import { BackupStatsSkeleton } from "@/components/ui/skeleton/backup-stats";
import { formatSize } from "@/lib/helper";

export function StatsCards() {
  const { stats, isLoadingStats } = useBackup();

  if (isLoadingStats) {
    return <BackupStatsSkeleton />;
  }

  if (!stats?.data) {
    return null;
  }

  const statsData = [
    {
      name: "Total Backups",
      value: stats.data.total_backups.toString(),
      change: stats.data.total_backups > 0 ? "Active" : "No backups",
      status: "All systems operational",
      trend: "up",
    },
    {
      name: "Success Rate",
      value: `${Math.round(stats.data.success_rate)}%`,
      change: stats.data.failed_backups === 0 ? "No failures" : `${stats.data.failed_backups} failed`,
      status: stats.data.success_rate > 95 ? "Highly reliable" : "Needs attention",
      trend: stats.data.success_rate > 95 ? "up" : "down",
    },
    {
      name: "Storage Used",
      value: formatSize(stats.data.total_size),
      change: `${Math.min(((stats.data.total_size / (1024 * 1024 * 1024 * 100)) * 100), 100).toFixed(0)}%`,
      status: "Of allocated space",
      showProgress: true,
    },
  ];

  return (
    <div className="grid gap-4 grid-cols-1 sm:grid-cols-2 md:grid-cols-3">
      {statsData.map((stat) => (
        <Card
          key={stat.name}
          className="p-4 sm:p-6 bg-card border"
        >
          <div className="flex justify-between items-start">
            <div className="space-y-2">
              <p className="text-sm font-medium text-muted-foreground">
                {stat.name}
              </p>
              <p className="text-xl sm:text-2xl font-semibold">{stat.value}</p>
            </div>
            {!stat.showProgress && (
              <div className={`flex items-center ${
                stat.trend === 'up' ? 'text-emerald-500' : 'text-amber-500'
              } text-xs font-medium`}>
                {stat.trend === 'up' ? (
                  <ArrowUpIcon className="h-3 w-3 mr-1" />
                ) : (
                  <ArrowDownIcon className="h-3 w-3 mr-1" />
                )}
                {stat.change}
              </div>
            )}
          </div>
          {stat.showProgress ? (
            <div className="mt-4">
              <div className="h-1 bg-muted/30 rounded-full overflow-hidden">
                <div 
                  className="h-full bg-blue-500/50 rounded-full" 
                  style={{ width: stat.change }}
                />
              </div>
              <p className="text-xs text-muted-foreground mt-2">{stat.status}</p>
            </div>
          ) : (
            <div className="mt-4 flex items-center text-[13px] text-muted-foreground/80">
              <div className={`h-1.5 w-1.5 rounded-full ${
                stat.trend === 'up' ? 'bg-emerald-500' : 'bg-amber-500'
              } mr-2`} />
              {stat.status}
            </div>
          )}
        </Card>
      ))}
    </div>
  );
}