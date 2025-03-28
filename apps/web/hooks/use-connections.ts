import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { saveConnection, testConnection, getConnections } from '@/lib/api/connections';
import { ConnectionForm } from '@/types/connection';
import { useToast } from '@/hooks/use-toast';

export function useConnections() {
  const queryClient = useQueryClient();
  const { toast } = useToast();

  const { data: connections, isLoading } = useQuery({
    queryKey: ['connections'],
    queryFn: getConnections,
  });

  const { mutate: addConnection, isPending: isAdding } = useMutation({
    mutationFn: async (connection: ConnectionForm) => {
      await testConnection(connection);
      await saveConnection(connection);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ 
        queryKey: ['connections']
      });
      toast({
        title: "Success",
        description: "Connection added successfully",
      });
    },
    onError: (error) => {
      toast({
        title: "Error",
        description: error.message || "Failed to add connection",
        variant: "destructive",
      });
    },
  });

  return {
    connections,
    isLoading,
    addConnection,
    isAdding
  };
}
