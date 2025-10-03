"use client";

import { Sidebar } from "@/components/layout/sidebar";
import { HistoryStats } from "@/components/views/history/history-stats";
import { HistoryList } from "@/components/views/history/history-list";

export default function HistoryPage() {
  return (
    <div className="flex h-screen bg-background">
      <Sidebar />
      <div className="flex-1 bg-gradient-to-br from-blue-50/50 via-background to-blue-50/30 dark:from-blue-950/10 dark:via-background dark:to-blue-950/5">
        <div className="h-full px-8 py-6">
          <div className="flex flex-col gap-6">
            <HistoryStats />
            <HistoryList />
          </div>
        </div>
      </div>
    </div>
  );
}