'use client';

import { Input } from '@/components/ui/input';
import { Search } from 'lucide-react';
import { useEffect, useState } from 'react';

interface ConnectionListHeaderProps {
  searchQuery: string;
  onSearchChange: (value: string) => void;
  actionButton?: React.ReactNode;
  resultCount?: number;
  totalCount?: number;
}

export function ConnectionListHeader({
  searchQuery,
  onSearchChange,
  actionButton,
  resultCount,
  totalCount,
}: ConnectionListHeaderProps) {
  const [isMac, setIsMac] = useState(false);

  useEffect(() => {
    setIsMac(navigator.platform.toUpperCase().indexOf('MAC') >= 0);
  }, []);

  useEffect(() => {
    const down = (e: KeyboardEvent) => {
      if (e.key === 'k' && (e.metaKey || e.ctrlKey)) {
        e.preventDefault();
        document.getElementById('connection-search')?.focus();
      }
    };

    document.addEventListener('keydown', down);
    return () => document.removeEventListener('keydown', down);
  }, []);

  return (
    <div className="flex items-center justify-between gap-4 pb-3">
      <div className="flex items-center gap-3 flex-1">
        <div className="relative max-w-md group flex-1">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground transition-colors group-focus-within:text-foreground" />
          <Input
            id="connection-search"
            placeholder="Search connections..."
            value={searchQuery}
            onChange={(e) => onSearchChange(e.target.value)}
            className="pl-9 pr-14 h-10 bg-background/50 border-border/50 transition-all duration-200 focus:bg-background focus:border-border group-hover:border-border/80 rounded-lg shadow-sm"
          />
          <kbd className="absolute right-3 top-1/2 -translate-y-1/2 pointer-events-none hidden sm:inline-flex h-5 select-none items-center gap-1 rounded border border-border/50 bg-muted/50 px-1.5 font-mono text-[10px] font-medium text-muted-foreground opacity-100 group-focus-within:opacity-0 transition-opacity">
            <span className="text-xs">{isMac ? '⌘' : 'Ctrl'}</span>K
          </kbd>
        </div>

        {resultCount !== undefined && totalCount !== undefined && (
          <div className="hidden sm:flex items-center gap-2 px-3 py-2 rounded-lg bg-accent/50 border border-border/40 shrink-0">
            <span className="text-xs font-semibold text-foreground">
              {resultCount}
            </span>
            <span className="text-xs text-muted-foreground">
              of {totalCount}
            </span>
          </div>
        )}
      </div>

      <div className="flex items-center gap-3">
        {actionButton && (
          <div className="shrink-0">
            {actionButton}
          </div>
        )}
      </div>
    </div>
  );
}
