'use client';

import { useState, useEffect } from 'react';
import { Database, Loader2, Search, CheckCircle2, Check } from 'lucide-react';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { ScrollArea } from '@/components/ui/scroll-area';
import { cn } from '@/lib/utils';
import { discoverDatabases } from '@/lib/api/connections';
import { useToast } from '@/hooks/use-toast';
import { useConnection, useConnections } from '@/hooks/use-connections';

interface DiscoverDatabasesModalProps {
  connectionId: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess?: () => void;
}

export function DiscoverDatabasesModal({
  connectionId,
  open,
  onOpenChange,
  onSuccess,
}: DiscoverDatabasesModalProps) {
  const [isDiscovering, setIsDiscovering] = useState(false);
  const [databases, setDatabases] = useState<string[]>([]);
  const [selectedDatabases, setSelectedDatabases] = useState<string[]>([]);
  const [searchQuery, setSearchQuery] = useState('');
  const { toast } = useToast();
  const { updateDatabases, isUpdatingDatabases } = useConnections();
  
  const { data: connection } = useConnection(connectionId);

  useEffect(() => {
    if (open && connection) {
      setSearchQuery('');
      
      if (connection.selected_databases && connection.selected_databases.length > 0) {
        setSelectedDatabases(connection.selected_databases);
      } else {
        setSelectedDatabases([]);
      }
      
      handleDiscover();
    }
  }, [open, connection?.selected_databases]);

  const handleDiscover = async () => {
    setIsDiscovering(true);
    try {
      const response = await discoverDatabases(connectionId);
      setDatabases(response.databases);
      
      if (response.databases.length === 0) {
        toast({
          title: 'No databases found',
          description: 'No accessible databases were found on this server.',
          variant: 'destructive',
        });
      }
    } catch (error) {
      toast({
        title: 'Discovery failed',
        description: error instanceof Error ? error.message : 'Failed to discover databases',
        variant: 'destructive',
      });
    } finally {
      setIsDiscovering(false);
    }
  };

  const handleToggleDatabase = (dbName: string) => {
    setSelectedDatabases((prev) =>
      prev.includes(dbName)
        ? prev.filter((db) => db !== dbName)
        : [...prev, dbName]
    );
  };

  const handleSelectAll = () => {
    const filtered = databases.filter((db) =>
      db.toLowerCase().includes(searchQuery.toLowerCase())
    );
    setSelectedDatabases(filtered);
  };

  const handleDeselectAll = () => {
    setSelectedDatabases([]);
  };

  const handleSave = () => {
    updateDatabases(
      { id: connectionId, databases: selectedDatabases },
      {
        onSuccess: () => {
          onSuccess?.();
          onOpenChange(false);
        },
      }
    );
  };

  const filteredDatabases = databases.filter((db) =>
    db.toLowerCase().includes(searchQuery.toLowerCase())
  );

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[540px]">
        <DialogHeader>
          <div className="flex items-center gap-3">
            <div className="flex items-center justify-center w-10 h-10 rounded-lg bg-primary/10">
              <Database className="h-5 w-5 text-primary" />
            </div>
            <div>
              <DialogTitle>Discover Databases</DialogTitle>
              <DialogDescription>
                Select databases to backup from <span className="font-medium text-foreground">{connection?.name || 'this connection'}</span>
              </DialogDescription>
            </div>
          </div>
        </DialogHeader>

        <div className="space-y-4 py-4">
          {/* Search & Actions Bar */}
          <div className="flex items-center gap-2">
            <div className="relative flex-1">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
              <Input
                placeholder="Search databases..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="pl-9"
              />
            </div>
            <Button
              variant="outline"
              size="sm"
              onClick={handleSelectAll}
              disabled={isDiscovering || filteredDatabases.length === 0}
            >
              Select All
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={handleDeselectAll}
              disabled={isDiscovering || selectedDatabases.length === 0}
            >
              Clear
            </Button>
          </div>

          {/* Selected Count Badge */}
          {selectedDatabases.length > 0 && (
            <div className="flex items-center gap-2 px-3 py-2 rounded-lg bg-primary/5 border border-primary/20">
              <CheckCircle2 className="h-4 w-4 text-primary" />
              <span className="text-sm font-medium">
                {selectedDatabases.length} database{selectedDatabases.length !== 1 ? 's' : ''} selected
              </span>
            </div>
          )}

          {/* Database List */}
          <ScrollArea className="h-[320px] rounded-lg border">
            {isDiscovering ? (
              <div className="flex flex-col items-center justify-center h-full py-12">
                <Loader2 className="h-8 w-8 animate-spin text-muted-foreground mb-3" />
                <p className="text-sm text-muted-foreground">Discovering databases...</p>
              </div>
            ) : filteredDatabases.length === 0 ? (
              <div className="flex flex-col items-center justify-center h-full py-12">
                <Database className="h-8 w-8 text-muted-foreground/50 mb-3" />
                <p className="text-sm text-muted-foreground">
                  {searchQuery ? 'No databases match your search' : 'No databases found'}
                </p>
              </div>
            ) : (
              <div className="p-2 space-y-0.5">
                {filteredDatabases.map((dbName) => {
                  const isSelected = selectedDatabases.includes(dbName);
                  return (
                    <button
                      key={dbName}
                      onClick={() => handleToggleDatabase(dbName)}
                      className={cn(
                        'w-full flex items-center gap-3 px-3 py-2 rounded-md transition-colors',
                        'hover:bg-accent/50',
                        isSelected && 'bg-accent/30'
                      )}
                    >
                      <div
                        className={cn(
                          'flex items-center justify-center w-4 h-4 rounded border transition-colors flex-shrink-0',
                          isSelected
                            ? 'bg-primary border-primary'
                            : 'border-border bg-background'
                        )}
                      >
                        {isSelected && (
                          <Check className="w-3 h-3 text-primary-foreground" />
                        )}
                      </div>
                      <span className="text-sm text-foreground flex-1 text-left">
                        {dbName}
                      </span>
                    </button>
                  );
                })}
              </div>
            )}
          </ScrollArea>

          {/* Info Box */}
          <div className="flex items-start gap-3 p-3 rounded-lg bg-muted/50 text-sm">
            <div className="flex-1 space-y-1">
              <p className="font-medium text-foreground/90">How it works</p>
              <p className="text-muted-foreground text-xs leading-relaxed">
                Selected databases will be backed up individually when you create a backup. 
                Each database gets its own backup file for granular restore capabilities.
              </p>
            </div>
          </div>
        </div>

        <DialogFooter className="gap-2">
          <Button
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={isUpdatingDatabases}
          >
            Cancel
          </Button>
          <Button
            onClick={handleSave}
            disabled={isUpdatingDatabases || selectedDatabases.length === 0}
          >
            {isUpdatingDatabases ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Saving...
              </>
            ) : (
              `Save Selection${selectedDatabases.length > 0 ? ` (${selectedDatabases.length})` : ''}`
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
