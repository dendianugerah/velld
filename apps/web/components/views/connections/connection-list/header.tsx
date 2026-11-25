"use client";

import { Search } from "lucide-react";
import { useEffect, useRef } from "react";
import { Input } from "@/components/ui/input";

interface ConnectionListHeaderProps {
  searchQuery: string;
  onSearchChange: (value: string) => void;
  actionButton: React.ReactNode;
  resultCount?: number;
  totalCount?: number;
}

export function ConnectionListHeader({ searchQuery, onSearchChange, actionButton, resultCount, totalCount }: ConnectionListHeaderProps) {
  const search = useRef<HTMLInputElement>(null);
  useEffect(() => {
    const focusSearch = (event: KeyboardEvent) => {
      if (event.key === "k" && (event.metaKey || event.ctrlKey)) {
        event.preventDefault();
        search.current?.focus();
      }
    };
    document.addEventListener("keydown", focusSearch);
    return () => document.removeEventListener("keydown", focusSearch);
  }, []);

  return (
    <div className="flex flex-wrap items-center justify-between gap-3">
      <div className="flex min-w-0 flex-1 flex-wrap items-center gap-3">
        <div className="relative w-full max-w-sm">
          <Search aria-hidden="true" className="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
          <Input ref={search} type="search" aria-label="Search connections" placeholder="Search connections" value={searchQuery} onChange={(event) => onSearchChange(event.target.value)} className="bg-card pl-9" />
        </div>
        {resultCount !== undefined && totalCount !== undefined && (
          <p className="text-xs text-muted-foreground" role="status">{resultCount} of {totalCount} connections</p>
        )}
      </div>
      {actionButton}
    </div>
  );
}
