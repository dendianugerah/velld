'use client';

import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Container } from 'lucide-react';
import { Button } from '@/components/ui/button';
import {
  Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger,
} from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { useToast } from '@/hooks/use-toast';
import { discoverDockerDatabases, importDockerDatabase, type DockerDatabase } from '@/lib/api/connections';

function errorMessage(error: Error) {
  try {
    return (JSON.parse(error.message) as { message?: string }).message || error.message;
  } catch {
    return error.message;
  }
}

export function DockerDiscoveryDialog() {
  const [open, setOpen] = useState(false);
  const [databaseNames, setDatabaseNames] = useState<Record<string, string>>({});
  const queryClient = useQueryClient();
  const { toast } = useToast();
  const discovery = useQuery({
    queryKey: ['docker-databases'],
    queryFn: discoverDockerDatabases,
    enabled: open,
    retry: false,
  });
  const add = useMutation({
    mutationFn: ({ container, database }: { container: DockerDatabase; database: string }) =>
      importDockerDatabase(container.id, database),
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ['connections'] }),
        queryClient.invalidateQueries({ queryKey: ['docker-databases'] }),
      ]);
      toast({ title: 'Docker database added', description: 'Daily backups are scheduled.' });
    },
    onError: (error) => {
      toast({ title: 'Could not add Docker database', description: errorMessage(error), variant: 'destructive' });
    },
  });

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button variant="outline"><Container aria-hidden="true" /> Discover Docker</Button>
      </DialogTrigger>
      <DialogContent className="max-w-xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Discover Docker databases</DialogTitle>
          <DialogDescription>
            Add a running database and schedule daily backups at midnight (server time) with 30 days of retention.
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-3">
          {discovery.isFetching && <p role="status" className="text-sm text-muted-foreground">Scanning Docker…</p>}
          {!discovery.isFetching && discovery.error && <p role="alert" className="text-sm text-destructive">{errorMessage(discovery.error)}</p>}
          {discovery.data?.hint && <p className="text-sm text-muted-foreground">{discovery.data.hint}</p>}
          {!discovery.isFetching && discovery.data?.containers.length === 0 && !discovery.data.hint && (
            <p className="text-sm text-muted-foreground">No new reachable database containers found.</p>
          )}
          {discovery.data?.containers.map((container) => {
            const database = databaseNames[container.id] ?? container.database;
            return (
              <form key={container.id} className="space-y-3 rounded-lg border p-3" onSubmit={(event) => {
                event.preventDefault();
                add.mutate({ container, database });
              }}>
                <div>
                  <p className="break-all font-medium">{container.name}</p>
                  <p className="break-all text-sm text-muted-foreground">{container.type} · {container.host}:{container.port}</p>
                </div>
                <div className="flex flex-wrap items-end gap-2">
                  <div className="min-w-36 flex-1 space-y-1">
                    <Label htmlFor={`docker-db-${container.id}`}>Database name</Label>
                    <Input
                      id={`docker-db-${container.id}`}
                      value={database}
                      onChange={(event) => setDatabaseNames({ ...databaseNames, [container.id]: event.target.value })}
                      required
                    />
                  </div>
                  <Button type="submit" disabled={add.isPending || discovery.isFetching || !database.trim()}>
                    {add.isPending && add.variables?.container.id === container.id ? 'Adding…' : 'Add & schedule'}
                  </Button>
                </div>
              </form>
            );
          })}
          <Button type="button" variant="outline" disabled={discovery.isFetching || add.isPending} onClick={() => discovery.refetch()}>
            Rescan
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
