"use client";

import { Sidebar } from "@/components/layout/sidebar";
import { PageHeader } from "@/components/layout/page-header";
import { HistoryStats } from "@/components/views/history/history-stats";
import { HistoryList } from "@/components/views/history/history-list";

export default function HistoryPage() {
  return (
    <div className="flex h-screen bg-background">
      <Sidebar />
      <div className="flex-1 overflow-auto lg:ml-64">
        <div className="h-full px-8 py-6 space-y-6">
          <PageHeader
            title="Backup History"
            description="View and manage all your database backups"
          />
          
          <div className="flex flex-col gap-6">
            <HistoryStats />
            <HistoryList />
          </div>
        </div>
      </div>
    </div>
  );
}