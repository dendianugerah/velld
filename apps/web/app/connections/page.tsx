"use client";

import { Sidebar } from "@/components/layout/sidebar";
import { PageHeader } from "@/components/layout/page-header";
import { ConnectionsList } from "@/components/views/connections/connection-list";
import { AddConnectionDialog } from "@/components/views/connections/add-connection-dialog";

export default function ConnectionsPage() {
  return (
    <div className="flex h-screen bg-background">
      <Sidebar />
      <div className="flex-1 overflow-auto lg:ml-64">
        <div className="h-full px-8 py-6 space-y-6">
          <PageHeader
            title="Connections"
            description="Manage your database connections and sources"
            action={<AddConnectionDialog />}
          />

          <ConnectionsList />
        </div>
      </div>
    </div>
  );
}