"use client";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
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
  verification: string;
  onVerificationChange: (value: string) => void;
  onReset: () => void;
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
  verification,
  onVerificationChange,
  onReset,
  active
}: HistoryFiltersProps) {
  
  return (
    <div className="flex flex-wrap items-center gap-2">
      <div className="relative min-w-0 basis-full xl:flex-1">
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
      
      <div className="grid w-full grid-cols-2 items-center gap-2 sm:grid-cols-4 xl:w-auto xl:grid-cols-[repeat(4,auto)]">
        <Select value={dateRange} onValueChange={onDateRangeChange}>
          <SelectTrigger aria-label="Filter by date" className="min-w-0 xl:min-w-32">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">Any date</SelectItem>
            <SelectItem value="24hours">Last 24 hours</SelectItem>
            <SelectItem value="7days">Last 7 days</SelectItem>
            <SelectItem value="30days">Last 30 days</SelectItem>
          </SelectContent>
        </Select>

        <Select value={status} onValueChange={onStatusChange}>
          <SelectTrigger aria-label="Filter by status" className="min-w-0 xl:min-w-32">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All statuses</SelectItem>
            <SelectItem value="completed">Completed</SelectItem>
            <SelectItem value="failed">Failed</SelectItem>
            <SelectItem value="running">Running</SelectItem>
            <SelectItem value="pending">Pending</SelectItem>
          </SelectContent>
        </Select>

        <Select value={databaseType} onValueChange={onDatabaseTypeChange}>
          <SelectTrigger aria-label="Filter by database engine" className="min-w-0 xl:min-w-32">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All engines</SelectItem>
            <SelectItem value="mysql">MySQL</SelectItem>
            <SelectItem value="postgresql">PostgreSQL</SelectItem>
            <SelectItem value="mongodb">MongoDB</SelectItem>
            <SelectItem value="d1">Cloudflare D1</SelectItem>
            <SelectItem value="mariadb">MariaDB</SelectItem>
            <SelectItem value="redis">Redis</SelectItem>
          </SelectContent>
        </Select>

        <Select value={verification} onValueChange={onVerificationChange}>
          <SelectTrigger aria-label="Filter by restore check" className="min-w-0 xl:min-w-36">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">Restore checks</SelectItem>
            <SelectItem value="passed">Verified</SelectItem>
            <SelectItem value="failed">Failed</SelectItem>
            <SelectItem value="pending">Running</SelectItem>
            <SelectItem value="skipped">Not checked</SelectItem>
          </SelectContent>
        </Select>
      </div>
      {active && <Button variant="ghost" size="sm" onClick={onReset}>Clear filters</Button>}
    </div>
  );
}