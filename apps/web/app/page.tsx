"use client";

import { Sidebar } from "@/components/layout/sidebar";
import { PageHeader } from "@/components/layout/page-header";
import { StatsCards } from "@/components/views/dashboard/stats-card";
import { ActivityList } from "@/components/views/dashboard/activity-list";

export default function Home() {
  return (
    <div className="flex h-screen bg-background">
      <Sidebar />
      <div className="flex-1 overflow-auto lg:ml-64">
        <div className="h-full px-8 py-6 space-y-6">
          <PageHeader
            title="Dashboard"
            description="Monitor and manage your database backups"
          />

          <StatsCards />

          <ActivityList />
        </div>
      </div>
    </div>
  );
}