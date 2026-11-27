"use client";

import { Button } from "@cloudflare/kumo/components/button";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { getPG18Report } from "@/lib/api/backups";

const issueGroups = [
  ["extension", "Extensions"],
  ["feature", "SQL and feature compatibility"],
  ["role", "Roles and permissions"],
  ["restore", "Other restore diagnostics"],
];

export function PG18Rehearsal({ backupId, onClose }: { backupId: string; onClose: () => void }) {
  const queryClient = useQueryClient();
  const queryKey = ["pg18-rehearsal", backupId];
  const reportQuery = useQuery({
    queryKey,
    queryFn: ({ signal }) => getPG18Report(backupId, "GET", signal),
    refetchInterval: query => query.state.data?.status === "pending" ? 1500 : false,
  });
  const start = useMutation({
    mutationFn: () => getPG18Report(backupId, "POST"),
    onSuccess: async report => {
      await queryClient.cancelQueries({ queryKey });
      queryClient.setQueryData(queryKey, report);
    },
  });

  const report = reportQuery.data;
  const running = start.isPending || report?.status === "pending";
  const finished = !running && (report?.status === "passed" || report?.status === "failed");
  return (
    <section aria-label="PostgreSQL 18 compatibility" className="col-span-full space-y-3 rounded-md border bg-muted/30 p-4">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="space-y-1">
          <h3 className="text-sm font-medium">PostgreSQL 18 compatibility</h3>
          <p className="text-xs text-muted-foreground">Restore this backup in a temporary PG 18 container. The source database is not changed.</p>
        </div>
        <div className="flex gap-2">
          <Button type="button" size="sm" variant="secondary" onClick={onClose}>Close</Button>
          <Button type="button" size="sm" variant="primary" onClick={() => start.mutate()} disabled={running || reportQuery.isPending}>
            {running ? "Testing…" : "Test on PG 18"}
          </Button>
        </div>
      </div>
      {start.isError && <p role="alert" className="text-xs text-destructive">Couldn’t start the rehearsal. Try again.</p>}
      {reportQuery.isError && <p role="alert" className="text-xs text-destructive">Couldn’t load the report. <button type="button" className="underline" onClick={() => reportQuery.refetch()}>Retry</button></p>}
      <div role="status" className="space-y-1">
        {running && <p className="text-xs text-muted-foreground">Restoring and checking the backup. You can close this panel while it runs.</p>}
        {finished && (
          <>
            <p className={report.status === "failed" ? "text-sm font-medium text-destructive" : "text-sm font-medium"}>
              {report.status === "passed" ? "Restore passed" : "Rehearsal failed"}
              <span className="ms-2 text-xs font-normal text-muted-foreground">
                {report.source_version ? `PG ${report.source_version} → ` : ""}PG {report.target_version}
                {report.duration_ms > 0 && ` · ${(report.duration_ms / 1000).toFixed(1)}s`}
              </span>
            </p>
            <p className="break-words text-xs text-muted-foreground">{report.message}</p>
          </>
        )}
      </div>
      {finished && issueGroups.map(([category, label]) => {
        const issues = report.issues.filter(issue => issue.category === category);
        return issues.length > 0 && (
          <div key={category} className="space-y-1">
            <h4 className="text-xs font-medium">{label}</h4>
            <ul className="max-h-48 space-y-1 overflow-y-auto text-xs text-muted-foreground">
              {issues.map((issue, index) => (
                <li key={index} className="break-words"><span className="capitalize font-medium">{issue.severity}:</span> {issue.message}</li>
              ))}
            </ul>
          </div>
        );
      })}
      {finished && !!report.omitted_issues && <p className="text-xs text-muted-foreground">{report.omitted_issues} additional diagnostics omitted.</p>}
    </section>
  );
}
