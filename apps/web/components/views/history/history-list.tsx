"use client";

import { Button } from "@/components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuSeparator, DropdownMenuTrigger } from "@/components/ui/dropdown-menu";
import { typeLabels, type DatabaseType } from "@/types/base";
import { MoreHorizontal } from "lucide-react";
import { format, isValid, parseISO, subDays, isAfter } from "date-fns";
import { useBackup } from "@/hooks/use-backup";
import { BackupList } from "@/types/backup";
import { HistoryListSkeleton } from "@/components/ui/skeleton/history-list";
import { formatSize } from "@/lib/helper";
import { HistoryFilters } from "./history-filters";
import { CustomPagination } from "@/components/ui/custom-pagination";
import { RestoreDialog } from "./restore-dialog";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { useState, useMemo } from "react";
import { useIsFetching } from "@tanstack/react-query";

export function HistoryList() {
  const { backups, isLoading, error, pagination, page, setPage, downloadBackupFile, isDownloading, search, setSearch, deleteBackupById, isDeleting } = useBackup();
  const [restoreDialogOpen, setRestoreDialogOpen] = useState(false);
  const [selectedBackupForRestore, setSelectedBackupForRestore] = useState<BackupList | null>(null);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [selectedBackupForDelete, setSelectedBackupForDelete] = useState<BackupList | null>(null);
  const isFetchingBackups = useIsFetching({ queryKey: ['backups'] });
  
  const [dateRange, setDateRange] = useState("all");
  const [status, setStatus] = useState("all");
  const [databaseType, setDatabaseType] = useState("all");

  const totalPages = pagination?.total_pages || 1;

  const handleRestore = (backup: BackupList) => {
    setSelectedBackupForRestore(backup);
    setRestoreDialogOpen(true);
  };

  const handleDelete = (backup: BackupList) => {
    setSelectedBackupForDelete(backup);
    setDeleteDialogOpen(true);
  };

  const confirmDelete = () => {
    if (!selectedBackupForDelete) return;
    deleteBackupById(selectedBackupForDelete.id, {
      onSuccess: () => {
        setDeleteDialogOpen(false);
        setSelectedBackupForDelete(null);
      },
    });
  };

  const handleResetFilters = () => {
    setPage(1);
    setSearch("");
    setDateRange("all");
    setStatus("all");
    setDatabaseType("all");
  };

  const filteredBackups = useMemo(() => {
    if (!backups) return [];

    return backups.filter((backup) => {
      let matchesDateRange = true;
      if (dateRange !== "all") {
        const backupDate = parseISO(backup.created_at);
        const now = new Date();
        
        switch (dateRange) {
          case "24hours":
            matchesDateRange = isAfter(backupDate, subDays(now, 1));
            break;
          case "7days":
            matchesDateRange = isAfter(backupDate, subDays(now, 7));
            break;
          case "30days":
            matchesDateRange = isAfter(backupDate, subDays(now, 30));
            break;
        }
      }

      const matchesStatus = status === "all" || backup.status.toLowerCase() === status.toLowerCase();
      const matchesDatabaseType = databaseType === "all" || backup.database_type.toLowerCase() === databaseType.toLowerCase();

      return matchesDateRange && matchesStatus && matchesDatabaseType;
    });
  }, [backups, dateRange, status, databaseType]);

  const hasFilters = Boolean(search) || dateRange !== "all" || status !== "all" || databaseType !== "all";

  return (
    <div className="space-y-4">
      <HistoryFilters
        search={search}
        onSearchChange={(value) => { setPage(1); setSearch(value); }}
        dateRange={dateRange}
        onDateRangeChange={setDateRange}
        status={status}
        onStatusChange={setStatus}
        databaseType={databaseType}
        onDatabaseTypeChange={setDatabaseType}
        onReset={handleResetFilters}
        isLoading={isLoading}
      />
      {isFetchingBackups > 0 && !isLoading && <p role="status" className="sr-only">Updating backups</p>}

      <div className="overflow-hidden rounded-xl border">
        {isLoading ? (
          <div className="p-5"><HistoryListSkeleton /></div>
        ) : filteredBackups.length ? (
          <ul className="divide-y">
            {filteredBackups.map((backup) => {
              const date = parseISO(backup.created_at);
              const filename = backup.path.replaceAll("\\", "/").split("/").pop() || backup.database_name || "Database backup";
              const database = typeLabels[backup.database_type as DatabaseType] ?? backup.database_type;
              return (
                <li key={backup.id} className="grid grid-cols-[minmax(0,1fr)_auto] items-center gap-3 px-4 py-4 sm:grid-cols-[minmax(0,1fr)_auto_auto] sm:gap-5 sm:px-5">
                  <div className="min-w-0">
                    <p title={filename} className="truncate text-sm font-medium">{filename}</p>
                    <p className="mt-1 flex flex-wrap gap-x-2 text-xs text-muted-foreground">
                      <span>{database}</span>
                      {isValid(date) && <time dateTime={backup.created_at}>{format(date, "MMM d, yyyy · p")}</time>}
                      {backup.size > 0 && <span>{formatSize(backup.size)}</span>}
                      {backup.s3_object_key && <span>Stored in S3</span>}
                    </p>
                  </div>
                  <div className="flex items-center gap-1.5">
                    <span aria-hidden="true" className={`size-1.5 rounded-full ${backup.status === "failed" ? "bg-destructive" : backup.status === "completed" ? "bg-[#49634d]" : "bg-muted-foreground"}`} />
                    <span className="text-xs capitalize text-muted-foreground">{backup.status.replaceAll("_", " ")}</span>
                  </div>
                  <div className="col-span-2 flex items-center justify-end gap-1 sm:col-span-1">
                    <Button variant="outline" size="sm" onClick={() => downloadBackupFile({ id: backup.id, path: backup.path })} disabled={isDownloading}>Download</Button>
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button variant="ghost" size="sm" className="w-8 p-0" aria-label={`Actions for ${filename}`}><MoreHorizontal /></Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        <DropdownMenuItem onSelect={() => handleRestore(backup)}>Restore backup</DropdownMenuItem>
                        <DropdownMenuSeparator />
                        <DropdownMenuItem className="text-destructive focus:text-destructive" onSelect={() => handleDelete(backup)}>Delete backup</DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </div>
                </li>
              );
            })}
          </ul>
        ) : (
          <div className="px-5 py-7">
            <p className="text-sm font-medium">{error ? "Couldn’t load backups" : hasFilters ? "No matching backups on this page" : "No backups yet"}</p>
            <p className="mt-1 text-sm text-muted-foreground">{error ? "Refresh the page to try again." : hasFilters ? "Try changing your search or filters." : "Backups will appear here after they run."}</p>
          </div>
        )}
      </div>
      {totalPages > 1 && <CustomPagination currentPage={page} totalPages={totalPages} onPageChange={setPage} />}

      <RestoreDialog
        backup={selectedBackupForRestore}
        open={restoreDialogOpen}
        onOpenChange={setRestoreDialogOpen}
      />

      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete Backup</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete this backup? The file will be removed from local storage
              {selectedBackupForDelete?.s3_object_key ? " and from S3" : ""}. This action cannot be undone.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteDialogOpen(false)} disabled={isDeleting}>
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={confirmDelete}
              disabled={isDeleting}
            >
              {isDeleting ? "Deleting..." : "Delete"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}