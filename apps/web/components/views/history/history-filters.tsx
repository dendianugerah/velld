"use client";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Select } from "@cloudflare/kumo/components/select";
import { Search } from "lucide-react";

interface HistoryFiltersProps {
  search: string;
  onSearchChange: (value: string) => void;
  dateRange: string;
  onDateRangeChange: (value: string) => void;
  status: string;
  onStatusChange: (value: string) => void;
  databaseType: string;
  onDatabaseTypeChange: (value: string) => void;
  onReset: () => void;
  verification: string;
  onVerificationChange: (value: string) => void;
  active: boolean;
}

export function HistoryFilters({
  search,
  onSearchChange,
  dateRange,
  onDateRangeChange,
  status,
  onStatusChange,
  databaseType,
  onDatabaseTypeChange,
  onReset,
  verification,
  onVerificationChange,
  active,
}: HistoryFiltersProps) {

  return (
    <div className="flex flex-col gap-3 2xl:flex-row 2xl:items-center 2xl:justify-between">
      <div className="relative w-full xl:max-w-xs">
        <Search aria-hidden="true" className="pointer-events-none absolute start-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
        <Input
          type="search"
          placeholder="Search backups…"
          aria-label="Search backup history"
          value={search}
          onChange={(e) => onSearchChange(e.target.value)}
          className="ps-9"
        />
      </div>

      <div className="grid grid-cols-2 items-center gap-2 sm:flex sm:flex-wrap">
        <Select aria-label="Filter by date" value={dateRange} onValueChange={value => { if (value) onDateRangeChange(value); }} items={{ all: "All Time", "24hours": "Last 24 Hours", "7days": "Last 7 Days", "30days": "Last 30 Days" }} className="w-full sm:w-[140px]">
          <Select.Option value="all">All Time</Select.Option>
          <Select.Option value="24hours">Last 24 Hours</Select.Option>
          <Select.Option value="7days">Last 7 Days</Select.Option>
          <Select.Option value="30days">Last 30 Days</Select.Option>
        </Select>

        <Select aria-label="Filter by status" value={status} onValueChange={value => { if (value) onStatusChange(value); }} items={{ all: "All Status", completed: "Completed", failed: "Failed", running: "Running", pending: "Pending" }} className="w-full sm:w-[130px]">
          <Select.Option value="all">All Status</Select.Option>
          <Select.Option value="completed">Completed</Select.Option>
          <Select.Option value="failed">Failed</Select.Option>
          <Select.Option value="running">Running</Select.Option>
          <Select.Option value="pending">Pending</Select.Option>
        </Select>

        <Select aria-label="Filter by database engine" value={databaseType} onValueChange={value => { if (value) onDatabaseTypeChange(value); }} items={{ all: "All engines", mysql: "MySQL", postgresql: "PostgreSQL", mongodb: "MongoDB", d1: "Cloudflare D1", mariadb: "MariaDB", redis: "Redis" }} className="w-full sm:w-[140px]">
          <Select.Option value="all">All engines</Select.Option>
          <Select.Option value="mysql">MySQL</Select.Option>
          <Select.Option value="postgresql">PostgreSQL</Select.Option>
          <Select.Option value="mongodb">MongoDB</Select.Option>
          <Select.Option value="d1">Cloudflare D1</Select.Option>
          <Select.Option value="mariadb">MariaDB</Select.Option>
          <Select.Option value="redis">Redis</Select.Option>
        </Select>

        <Select aria-label="Filter by restore check" value={verification} onValueChange={value => { if (value) onVerificationChange(value); }} items={{ all: "Restore checks", passed: "Verified", failed: "Failed", pending: "Running", skipped: "Not checked" }} className="w-full sm:w-[145px]">
          <Select.Option value="all">Restore checks</Select.Option>
          <Select.Option value="passed">Verified</Select.Option>
          <Select.Option value="failed">Failed</Select.Option>
          <Select.Option value="pending">Running</Select.Option>
          <Select.Option value="skipped">Not checked</Select.Option>
        </Select>
        {active && <Button variant="ghost" size="sm" onClick={onReset} className="w-full sm:w-auto">
          Clear filters
        </Button>}
      </div>
    </div>
  );
}
