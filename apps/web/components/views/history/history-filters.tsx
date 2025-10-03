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
import { useBackup } from "@/hooks/use-backup";
import { Search } from "lucide-react";

export function HistoryFilters() {
  const { search, setSearch, isLoading } = useBackup();
  
  return (
    <div className="flex flex-col sm:flex-row gap-3 items-start sm:items-center">
      <div className="relative flex-1 w-full">
        <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
        <Input
          type="text"
          placeholder="Search history..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="pl-10"
          disabled={isLoading}
        />
      </div>
      
      <div className="flex items-center gap-2 w-full sm:w-auto">
        <Select defaultValue="7days">
          <SelectTrigger className="w-full sm:w-[140px]">
            <SelectValue placeholder="Last 7 Days" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="24hours">Last 24 Hours</SelectItem>
            <SelectItem value="7days">Last 7 Days</SelectItem>
            <SelectItem value="30days">Last 30 Days</SelectItem>
            <SelectItem value="custom">Custom Range</SelectItem>
          </SelectContent>
        </Select>

        <Select defaultValue="all">
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

        <Select defaultValue="all">
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

        <Button variant="ghost" size="sm">
          Reset Filters
        </Button>
      </div>
    </div>
  );
}