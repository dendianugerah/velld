"use client";

import { Sidebar } from "@/components/layout/sidebar";
import { ConnectionsList } from "@/components/views/connections/connection-list";
import { AddConnectionDialog } from "@/components/views/connections/add-connection-dialog";

export default function ConnectionsPage() {
  return (
    <div className="flex h-screen bg-background">
      <Sidebar />
      <div className="flex-1 bg-gradient-to-br from-blue-50/50 via-background to-blue-50/30 dark:from-blue-950/10 dark:via-background dark:to-blue-950/5">
        <div className="h-full px-8 py-6">
          <div className="flex justify-between items-center mb-6">
            <div>
              <h2 className="text-3xl font-bold tracking-tight">Connections</h2>
              <p className="text-base text-muted-foreground">
                Manage your database connections and sources
              </p>
            </div>
            <AddConnectionDialog />
          </div>

          <ConnectionsList />
        </div>
      </div>
    </div>
  );
}