"use client";

import { useState, useEffect } from "react";
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { GitCompare, Loader2 } from "lucide-react";
import { BackupList, BackupDiff } from "@/types/backup";
import { formatSize } from "@/lib/helper";
import { compareBackups } from "@/lib/api/backups";
import { useToast } from "@/hooks/use-toast";

interface BackupCompareDialogProps {
  open: boolean;
  onClose: () => void;
  backups: BackupList[];
  selectedBackup?: BackupList;
}

export function BackupCompareDialog({ 
  open, 
  onClose, 
  backups,
  selectedBackup 
}: BackupCompareDialogProps) {
  const [sourceBackupId, setSourceBackupId] = useState<string>("");
  const [compareWith, setCompareWith] = useState<string>("");
  const [viewMode, setViewMode] = useState<"split" | "unified">("split");
  const [diffData, setDiffData] = useState<BackupDiff | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const { toast } = useToast();

  useState(() => {
    if (selectedBackup) {
      setSourceBackupId(selectedBackup.id);
    }
  });

  const selectedSourceBackup = backups.find(b => b.id === sourceBackupId);
  const selectedCompareBackup = backups.find(b => b.id === compareWith);

  useEffect(() => {
    async function fetchComparison() {
      if (selectedSourceBackup && selectedCompareBackup) {
        setIsLoading(true);
        try {
          const response = await compareBackups(selectedSourceBackup.id, selectedCompareBackup.id);
          setDiffData(response.data);
        } catch (error) {
          toast({
            title: "Error",
            description: "Failed to compare backups. Please try again.",
            variant: "destructive",
          });
          console.error("Comparison error:", error);
        } finally {
          setIsLoading(false);
        }
      }
    }

    fetchComparison();
  }, [selectedSourceBackup, selectedCompareBackup, toast]);

  useEffect(() => {
    if (selectedBackup && open) {
      setSourceBackupId(selectedBackup.id);
    }
  }, [selectedBackup, open]);

  return (
    <Dialog open={open} onOpenChange={onClose}>
      <DialogContent className="max-w-6xl max-h-[90vh] overflow-hidden flex flex-col">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <GitCompare className="h-5 w-5" />
            Compare Backups
          </DialogTitle>
          <DialogDescription>
            Select two backups to compare their database dumps and see the differences
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 flex-1 overflow-hidden flex flex-col">
          {/* Backup Selection */}
          <div className="grid md:grid-cols-2 gap-4">
            {/* Source Backup */}
            <div className="p-4 rounded-lg border bg-card space-y-3">
              <div className="flex items-center justify-between">
                <span className="text-sm font-medium text-muted-foreground">Source Backup</span>
                <Badge variant="secondary">Base</Badge>
              </div>
              <Select value={sourceBackupId} onValueChange={setSourceBackupId}>
                <SelectTrigger>
                  <SelectValue placeholder="Select source backup" />
                </SelectTrigger>
                <SelectContent>
                  {backups
                    .filter(b => b.id !== compareWith)
                    .map(backup => (
                      <SelectItem key={backup.id} value={backup.id}>
                        <div className="flex items-center gap-2">
                          <span className="truncate max-w-[300px]">{backup.path.split('\\').pop()}</span>
                          <span className="text-xs text-muted-foreground">
                            ({formatSize(backup.size)})
                          </span>
                        </div>
                      </SelectItem>
                    ))}
                </SelectContent>
              </Select>
            </div>

            <div className="p-4 rounded-lg border bg-card space-y-3">
              <div className="flex items-center justify-between">
                <span className="text-sm font-medium text-muted-foreground">Compare With</span>
                <Badge variant="secondary">Compare</Badge>
              </div>
              <Select value={compareWith} onValueChange={setCompareWith}>
                <SelectTrigger>
                  <SelectValue placeholder="Select backup to compare" />
                </SelectTrigger>
                <SelectContent>
                  {backups
                    .filter(b => b.id !== sourceBackupId)
                    .map(backup => (
                      <SelectItem key={backup.id} value={backup.id}>
                        <div className="flex items-center gap-2">
                          <span className="truncate max-w-[300px]">{backup.path.split('\\').pop()}</span>
                          <span className="text-xs text-muted-foreground">
                            ({formatSize(backup.size)})
                          </span>
                        </div>
                      </SelectItem>
                    ))}
                </SelectContent>
              </Select>
            </div>
          </div>

          {selectedSourceBackup && selectedCompareBackup && (
            <>
              {isLoading ? (
                <div className="flex items-center justify-center p-8">
                  <Loader2 className="h-8 w-8 animate-spin text-primary" />
                </div>
              ) : diffData ? (
                <div className="grid grid-cols-4 gap-4">
                  <div className="p-3 rounded-lg border bg-green-50 dark:bg-green-950/20 border-green-200 dark:border-green-900">
                    <div className="text-2xl font-bold text-green-600 dark:text-green-400">+{diffData.added}</div>
                    <div className="text-xs text-green-600 dark:text-green-400">Added</div>
                  </div>
                  <div className="p-3 rounded-lg border bg-red-50 dark:bg-red-950/20 border-red-200 dark:border-red-900">
                    <div className="text-2xl font-bold text-red-600 dark:text-red-400">-{diffData.removed}</div>
                    <div className="text-xs text-red-600 dark:text-red-400">Removed</div>
                  </div>
                  <div className="p-3 rounded-lg border bg-blue-50 dark:bg-blue-950/20 border-blue-200 dark:border-blue-900">
                    <div className="text-2xl font-bold text-blue-600 dark:text-blue-400">~{diffData.modified}</div>
                    <div className="text-xs text-blue-600 dark:text-blue-400">Modified</div>
                  </div>
                  <div className="p-3 rounded-lg border bg-gray-50 dark:bg-gray-950/20 border-gray-200 dark:border-gray-800">
                    <div className="text-2xl font-bold text-gray-600 dark:text-gray-400">{diffData.unchanged}</div>
                    <div className="text-xs text-gray-600 dark:text-gray-400">Unchanged</div>
                  </div>
                </div>
              ) : null}
            </>
          )}

          {selectedSourceBackup && selectedCompareBackup && diffData && !isLoading && (
            <Tabs value={viewMode} onValueChange={(v) => setViewMode(v as "split" | "unified")} className="flex-1 overflow-hidden flex flex-col">
              <TabsList className="grid w-full max-w-[400px] grid-cols-2">
                <TabsTrigger value="split">Split View</TabsTrigger>
                <TabsTrigger value="unified">Unified View</TabsTrigger>
              </TabsList>

              <TabsContent value="split" className="flex-1 overflow-auto mt-4">
                <div className="grid md:grid-cols-2 gap-4 h-full">
                  {/* Left: Source */}
                  <div className="rounded-lg border overflow-hidden">
                    <div className="bg-muted px-4 py-2 border-b">
                      <p className="text-sm font-medium">Source</p>
                    </div>
                    <div className="p-4 font-mono text-xs space-y-1 bg-card overflow-auto max-h-[400px]">
                      {diffData.changes.map((change, idx) => (
                        <div
                          key={idx}
                          className={`px-2 py-1 ${
                            change.type === 'removed' ? 'bg-red-50 dark:bg-red-950/20 text-red-700 dark:text-red-400' :
                            change.type === 'modified' ? 'bg-blue-50 dark:bg-blue-950/20 text-blue-700 dark:text-blue-400' :
                            ''
                          }`}
                        >
                          <span className="text-muted-foreground mr-4">{change.line_number}</span>
                          {change.content}
                        </div>
                      ))}
                    </div>
                  </div>

                  {/* Right: Compare */}
                  <div className="rounded-lg border overflow-hidden">
                    <div className="bg-muted px-4 py-2 border-b">
                      <p className="text-sm font-medium">Compare</p>
                    </div>
                    <div className="p-4 font-mono text-xs space-y-1 bg-card overflow-auto max-h-[400px]">
                      {diffData.changes.map((change, idx) => (
                        <div
                          key={idx}
                          className={`px-2 py-1 ${
                            change.type === 'added' ? 'bg-green-50 dark:bg-green-950/20 text-green-700 dark:text-green-400' :
                            change.type === 'modified' ? 'bg-blue-50 dark:bg-blue-950/20 text-blue-700 dark:text-blue-400' :
                            ''
                          }`}
                        >
                          <span className="text-muted-foreground mr-4">{change.line_number}</span>
                          {change.content}
                        </div>
                      ))}
                    </div>
                  </div>
                </div>
              </TabsContent>

              <TabsContent value="unified" className="flex-1 overflow-auto mt-4">
                <div className="rounded-lg border overflow-hidden">
                  <div className="bg-muted px-4 py-2 border-b">
                    <p className="text-sm font-medium">Unified Diff</p>
                  </div>
                  <div className="p-4 font-mono text-xs space-y-1 bg-card overflow-auto max-h-[400px]">
                    {diffData.changes.map((change, idx) => (
                      <div
                        key={idx}
                        className={`px-2 py-1 ${
                          change.type === 'added' ? 'bg-green-50 dark:bg-green-950/20 text-green-700 dark:text-green-400' :
                          change.type === 'removed' ? 'bg-red-50 dark:bg-red-950/20 text-red-700 dark:text-red-400' :
                          change.type === 'modified' ? 'bg-blue-50 dark:bg-blue-950/20 text-blue-700 dark:text-blue-400' :
                          ''
                        }`}
                      >
                        <span className="text-muted-foreground mr-4">{change.line_number}</span>
                        {change.content}
                      </div>
                    ))}
                  </div>
                </div>
              </TabsContent>
            </Tabs>
          )}

          {(!selectedSourceBackup || !selectedCompareBackup) && !isLoading && (
            <div className="flex-1 flex items-center justify-center border-2 border-dashed rounded-lg">
              <div className="text-center space-y-2 p-8">
                <GitCompare className="h-12 w-12 mx-auto text-muted-foreground" />
                <p className="text-sm font-medium">Select backups to compare</p>
                <p className="text-xs text-muted-foreground">
                  Choose two backups to see their differences
                </p>
              </div>
            </div>
          )}
        </div>

        <div className="flex justify-end gap-2 border-t pt-4">
          <Button variant="outline" onClick={onClose}>
            Close
          </Button>
          {selectedSourceBackup && selectedCompareBackup && (
            <Button disabled={true}>
              <GitCompare className="h-4 w-4 mr-2" />
              Export Diff Report
            </Button>
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}
