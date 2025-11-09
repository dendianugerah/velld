import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { saveConnection, testConnection, getConnection, getConnections, updateConnection, deleteConnection, updateSelectedDatabases } from '@/lib/api/connections';
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

  const { mutate: editConnection, isPending: isEditing } = useMutation({
    mutationFn: async (connection: ConnectionForm & { id: string }) => {
      await testConnection(connection);
      await updateConnection(connection);
    },
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({ queryKey: ['connections'] });
      queryClient.invalidateQueries({ queryKey: ['connection', variables.id] });
      toast({
        title: "Success",
        description: "Connection updated successfully",
      });
    },
    onError: (error) => {
      toast({
        title: "Error",
        description: error.message || "Failed to update connection",
        variant: "destructive",
      });
    },
  });

  const { mutate: removeConnection, isPending: isDeleting } = useMutation({
    mutationFn: deleteConnection,
    onSuccess: (_data, id) => {
      queryClient.invalidateQueries({ queryKey: ['connections'] });
      queryClient.invalidateQueries({ queryKey: ['connection', id] });
      toast({
        title: "Success",
        description: "Connection deleted successfully",
      });
    },
    onError: (error) => {
      toast({
        title: "Error",
        description: error.message || "Failed to delete connection",
        variant: "destructive",
      });
    },
  });

  const { mutate: updateDatabases, isPending: isUpdatingDatabases } = useMutation({
    mutationFn: ({ id, databases }: { id: string; databases: string[] }) => 
      updateSelectedDatabases(id, databases),
    onSuccess: (_data, { id, databases }) => {
      queryClient.invalidateQueries({ queryKey: ['connections'] });
      queryClient.invalidateQueries({ queryKey: ['connection', id] });
      toast({
        title: "Success",
        description: `${databases.length} database(s) selected for backup`,
      });
    },
    onError: (error) => {
      toast({
        title: "Error",
        description: error.message || "Failed to update selected databases",
        variant: "destructive",
      });
    },
  });

  return {
    connections,
    isLoading,
    addConnection,
    isAdding,
    editConnection,
    isEditing,
    removeConnection,
    isDeleting,
    updateDatabases,
    isUpdatingDatabases,
  };
}

export function useConnection(id: string | null) {
  return useQuery({
    queryKey: ['connection', id],
    queryFn: () => {
      return getConnection(id!);
    },
    enabled: !!id,
  });
}