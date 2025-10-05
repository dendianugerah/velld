import { Card } from "@/components/ui/card";
import { ArrowUpIcon, ArrowDownIcon } from "lucide-react";
import { useBackup } from "@/hooks/use-backup";
import { BackupStatsSkeleton } from "@/components/ui/skeleton/backup-stats";
import { formatSize } from "@/lib/helper";

export function HistoryStats() {
  const { stats, isLoadingStats } = useBackup();

  if (isLoadingStats) {
    return <BackupStatsSkeleton />;
  }

  if (!stats?.data) {
    return null;
  }

  const statsData = [
    {
      name: "Success Rate",
      value: stats.data.success_rate ? `${Math.round(stats.data.success_rate)}%` : '0%',
      status: "Overall success rate",
      trend: "up",
    },
    {
      name: "Failed Backups",
      value: stats.data.failed_backups !== undefined ? stats.data.failed_backups : '0',
      status: `${stats.data.total_backups || 0} total backups`,
      trend: stats.data.failed_backups === 0 ? "up" : "down",
    },
    {
      name: "Average Duration",
      value: stats.data.average_duration ? 
        `${(stats.data.average_duration * 60).toFixed(0)}s` : '0s',
      status: "Average backup time",
      showProgress: true,
    },
    {
      name: "Storage Used",
      value: formatSize(stats.data.total_size || 0),
      status: "Total backup size",
      showProgress: true,
    },
  ];

  return (
    <div className="grid gap-4 grid-cols-2 lg:grid-cols-4">
      {statsData.map((stat) => (
        <Card
          key={stat.name}
          className="p-4 sm:p-6 bg-background border border-border/50 hover:border-border/80 transition-colors"
        >
          <div className="flex justify-between items-start">
            <div className="space-y-2">
              <p className="text-xs sm:text-sm font-medium text-muted-foreground">
                {stat.name}
              </p>
              <p className="text-xl sm:text-2xl font-semibold">{stat.value}</p>
            </div>
            {stat.trend && (
              <div className={`flex items-center ${
                stat.trend === 'up' ? 'text-emerald-500' : 'text-amber-500'
              } text-xs font-medium`}>
                {stat.trend === 'up' ? (
                  <ArrowUpIcon className="h-3 w-3" />
                ) : (
                  <ArrowDownIcon className="h-3 w-3" />
                )}
              </div>
            )}
          </div>
          <div className="mt-4 flex items-center text-[13px] text-muted-foreground/80">
            <div className={`h-1.5 w-1.5 rounded-full ${
              stat.trend === 'up' ? 'bg-emerald-500' : 
              stat.trend === 'down' ? 'bg-amber-500' : 
              'bg-blue-500'
            } mr-2`} />
            {stat.status}
          </div>
        </Card>
      ))}
    </div>
  );
}

export default HistoryStats;