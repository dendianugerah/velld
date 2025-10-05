"use client";

import { Sidebar } from "@/components/layout/sidebar";
import { ProfileSettings } from "@/components/views/settings/profile-settings";


export default function SettingsPage() {
  return (
    <div className="flex h-screen bg-background">
      <Sidebar />
      <div className="flex-1 overflow-auto lg:ml-64">
        <div className="px-4 sm:px-6 lg:px-8 py-4 sm:py-6 space-y-4 sm:space-y-6 pt-16 lg:pt-6">
          <ProfileSettings />
        </div>
      </div>
    </div>
  );
}