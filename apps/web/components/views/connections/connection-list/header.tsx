'use client';

import { Input } from '@/components/ui/input';
import { Search, ArrowUpDown } from 'lucide-react';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
} from "@/components/ui/select";
import { SortBy } from '@/types/connection';

interface ConnectionListHeaderProps {
  searchQuery: string;
  onSearchChange: (value: string) => void;
  sortBy: SortBy;
  onSortChange: (value: SortBy) => void;
}

export function ConnectionListHeader({
  searchQuery,
  onSearchChange,
  sortBy,
  onSortChange,
}: ConnectionListHeaderProps) {
  return (
    <div className="flex flex-col sm:flex-row items-stretch sm:items-center justify-between gap-3 mb-6">
      <div className="relative flex-1 max-w-full sm:max-w-xs">
        <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
        <Input
          placeholder="Search connections..."
          value={searchQuery}
          onChange={(e) => onSearchChange(e.target.value)}
          className="pl-8"
        />
      </div>
      <Select value={sortBy} onValueChange={onSortChange}>
        <SelectTrigger className="w-full sm:w-[180px]">
          <div className="flex items-center">
            <ArrowUpDown className="mr-2 h-4 w-4" />
            <span>Sort by</span>
          </div>
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="name">Name</SelectItem>
          <SelectItem value="status">Status</SelectItem>
          <SelectItem value="type">Type</SelectItem>
          <SelectItem value="lastBackup">Last Backup</SelectItem>
        </SelectContent>
      </Select>
    </div>
  );
} 