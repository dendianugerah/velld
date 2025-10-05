import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { getBackups, saveBackup, scheduleBackup, disableBackupSchedule, updateSchedule, getBackupStats, downloadBackup } from "@/lib/api/backups";
import { useToast } from "@/hooks/use-toast";
import { useState } from 'react';

export interface ScheduleBackupParams {
  connection_id: string;
  cron_schedule: string;
  retention_days: number;
}

export function useBackup() {
  const queryClient = useQueryClient();
  const { toast } = useToast();

  const [page, setPage] = useState(1);
  const [limit] = useState(10);
  const [search, setSearch] = useState('');

  const { data, isLoading, error } = useQuery({
    queryKey: ['backups', { page, limit, search }],
    queryFn: () => getBackups({ page, limit, search }),
    placeholderData: (previousData) => previousData,
    refetchInterval: 30000,
    refetchIntervalInBackground: false,
  });

  const { data: stats, isLoading: isLoadingStats } = useQuery({
    queryKey: ['backup-stats'],
    queryFn: () => getBackupStats(),
    refetchInterval: 30000,
    refetchIntervalInBackground: false,
  });

  const { mutate: createBackup, isPending: isCreating } = useMutation({
    mutationFn: async (connectionId: string) => {
      await saveBackup(connectionId);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['backups']});
      queryClient.invalidateQueries({ queryKey: ['connections']});
      toast({
        title: "Success",
        description: "Backup started successfully",
      });
    },
    onError: (error) => {
      toast({
        title: "Error",
        description: error.message || "Failed to start backup",
        variant: "destructive",
      });
    },
  });

  const { mutate: createSchedule, isPending: isScheduling } = useMutation({
    mutationFn: async (params: ScheduleBackupParams) => {
      await scheduleBackup(params);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['backups']});
      queryClient.invalidateQueries({ queryKey: ['connections']});
      toast({
        title: "Success",
        description: "Backup schedule created successfully",
      });
    },
    onError: (error) => {
      toast({
        title: "Error",
        description: error.message || "Failed to create backup schedule",
        variant: "destructive",
      });
    },
  });

  const { mutate: updateExistingSchedule, isPending: isUpdating } = useMutation({
    mutationFn: async ({ connectionId, params }: { connectionId: string; params: Omit<ScheduleBackupParams, 'connection_id'> }) => {
      await updateSchedule(connectionId, params);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['backups']});
      queryClient.invalidateQueries({ queryKey: ['connections']});
      toast({
        title: "Success",
        description: "Backup schedule updated successfully",
      });
    },
    onError: (error) => {
      toast({
        title: "Error",
        description: error.message || "Failed to update backup schedule",
        variant: "destructive",
      });
    },
  });

  const { mutate: disableSchedule, isPending: isDisabling } = useMutation({
    mutationFn: async (connectionId: string) => {
      await disableBackupSchedule(connectionId);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ['backups'],
      });
      queryClient.invalidateQueries({
        queryKey: ['connections'],
      });
      toast({
        title: "Success",
        description: "Backup schedule disabled successfully",
      });
    },
    onError: (error) => {
      toast({
        title: "Error",
        description: error.message || "Failed to disable backup schedule",
        variant: "destructive",
      });
    },
  });

  const { mutate: downloadBackupFile, isPending: isDownloading } = useMutation({
    mutationFn: async (backup: { id: string, path: string }) => {
      const blob = await downloadBackup(backup.id);
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      
      const filename = backup.path.split('\\').pop() || `backup-${backup.id}.sql`;
      a.download = filename;
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      a.remove();
    },
    onError: (error) => {
      toast({
        title: "Error",
        description: error.message || "Failed to download backup",
        variant: "destructive",
      });
    },
  });

  return {
    createBackup,
    isCreating,
    createSchedule,
    isScheduling,
    updateExistingSchedule,
    isUpdating,
    disableSchedule,
    isDisabling,
    downloadBackupFile,
    isDownloading,
    backups: data?.data,
    pagination: data?.pagination,
    isLoading,
    error,
    page,
    setPage,
    search,
    setSearch,
    stats,
    isLoadingStats,
  };
}