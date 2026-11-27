import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { getUserSettings, updateUserSettings } from '@/lib/api/settings';
import { useToast } from '@/hooks/use-toast';
import { UpdateSettingsRequest } from '@/types/settings';

export function useSettings() {
  const { toast } = useToast();
  const queryClient = useQueryClient();

  const { data, isLoading, error } = useQuery({
    queryKey: ['settings'],
    queryFn: getUserSettings,
  });

  const mutation = useMutation({
    mutationFn: updateUserSettings,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['settings'] });
      toast({
        title: "Success",
        description: "Settings updated successfully",
      });
    },
    onError: () => {
      toast({
        title: "Error",
        description: "Failed to update settings",
        variant: "destructive",
      });
    },
  });

  return {
    settings: data?.data,
    isLoading,
    error,
    updateSettings: (data: UpdateSettingsRequest) => mutation.mutateAsync(data),
    isUpdating: mutation.isPending,
  };
}
