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
  onReset: () => void;
  isLoading?: boolean;
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
  isLoading
}: HistoryFiltersProps) {
  
  return (
    <div className="flex flex-col gap-3 2xl:flex-row 2xl:items-center 2xl:justify-between">
      <div className="relative w-full xl:max-w-xs">
        <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
        <Input
          type="text"
          placeholder="Search database, connection or file…"
          aria-label="Search backup history"
          value={search}
          onChange={(e) => onSearchChange(e.target.value)}
          className="pl-10"
          disabled={isLoading}
        />
      </div>
      
      <div className="grid grid-cols-2 items-center gap-2 sm:flex sm:flex-wrap">
        <Select value={dateRange} onValueChange={onDateRangeChange}>
          <SelectTrigger aria-label="Filter by date" className="w-full sm:w-[140px]">
            <SelectValue placeholder="Last 7 Days" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Time</SelectItem>
            <SelectItem value="24hours">Last 24 Hours</SelectItem>
            <SelectItem value="7days">Last 7 Days</SelectItem>
            <SelectItem value="30days">Last 30 Days</SelectItem>
          </SelectContent>
        </Select>

        <Select value={status} onValueChange={onStatusChange}>
          <SelectTrigger aria-label="Filter by status" className="w-full sm:w-[130px]">
            <SelectValue placeholder="All Status" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Status</SelectItem>
            <SelectItem value="completed">Completed</SelectItem>
            <SelectItem value="failed">Failed</SelectItem>
            <SelectItem value="running">Running</SelectItem>
            <SelectItem value="pending">Pending</SelectItem>
          </SelectContent>
        </Select>

        <Select value={databaseType} onValueChange={onDatabaseTypeChange}>
          <SelectTrigger aria-label="Filter by database engine" className="w-full sm:w-[140px]">
            <SelectValue placeholder="All Databases" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Databases</SelectItem>
            <SelectItem value="mysql">MySQL</SelectItem>
            <SelectItem value="postgresql">PostgreSQL</SelectItem>
            <SelectItem value="mongodb">MongoDB</SelectItem>
            <SelectItem value="mariadb">MariaDB</SelectItem>
            <SelectItem value="redis">Redis</SelectItem>
          </SelectContent>
        </Select>

        <Button variant="ghost" size="sm" onClick={onReset} className="w-full sm:w-auto">
          Reset
        </Button>
      </div>
    </div>
  );
}