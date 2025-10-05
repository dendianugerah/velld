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
import { Search, GitCompare } from "lucide-react";

interface HistoryFiltersProps {
  onCompare?: () => void;
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
  onCompare, 
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
    <div className="flex flex-col gap-3">
      <div className="relative w-full">
        <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
        <Input
          type="text"
          placeholder="Search history..."
          value={search}
          onChange={(e) => onSearchChange(e.target.value)}
          className="pl-10"
          disabled={isLoading}
        />
      </div>
      
      <div className="flex flex-wrap items-center gap-2">
        <Select value={dateRange} onValueChange={onDateRangeChange}>
          <SelectTrigger className="w-full sm:w-[140px]">
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
          <SelectTrigger className="w-full sm:w-[130px]">
            <SelectValue placeholder="All Status" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Status</SelectItem>
            <SelectItem value="completed">Completed</SelectItem>
            <SelectItem value="failed">Failed</SelectItem>
            <SelectItem value="running">Running</SelectItem>
          </SelectContent>
        </Select>

        <Select value={databaseType} onValueChange={onDatabaseTypeChange}>
          <SelectTrigger className="w-full sm:w-[140px]">
            <SelectValue placeholder="All Databases" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Databases</SelectItem>
            <SelectItem value="mysql">MySQL</SelectItem>
            <SelectItem value="postgresql">PostgreSQL</SelectItem>
            <SelectItem value="mongodb">MongoDB</SelectItem>
          </SelectContent>
        </Select>

        {onCompare && (
          <Button 
            variant="ghost" 
            size="sm"
            onClick={onCompare}
            className="gap-2 w-full sm:w-auto"
          >
            <GitCompare className="h-4 w-4" />
            <span className="sm:inline">Compare</span>
          </Button>
        )}

        <Button variant="ghost" size="sm" onClick={onReset} className="w-full sm:w-auto">
          Reset
        </Button>
      </div>
    </div>
  );
}