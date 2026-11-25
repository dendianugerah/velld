"use client";

import { Skeleton } from "@/components/ui/skeleton";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { useProfile } from "@/hooks/use-profile";
import { NotificationSettings } from "./notification-settings";
import { S3StorageSettings } from "./s3-storage-settings";

export function ProfileSettings() {
  const { data: profile, isLoading } = useProfile();

  return (
    <Tabs defaultValue="account" className="space-y-5">
      <TabsList aria-label="Settings categories" className="grid h-auto w-full max-w-md grid-cols-3 p-1">
        <TabsTrigger value="account">Account</TabsTrigger>
        <TabsTrigger value="alerts">Alerts</TabsTrigger>
        <TabsTrigger value="storage">Storage</TabsTrigger>
      </TabsList>
      <TabsContent value="account">
        <section aria-labelledby="account-heading" className="rounded-xl border px-5 py-6 sm:px-6">
          <h2 id="account-heading" className="text-base font-semibold">Profile</h2>
          <dl className="mt-5 flex items-center justify-between gap-4 border-t pt-4 text-sm">
            <dt className="text-muted-foreground">Username</dt>
            <dd className="min-w-0 truncate font-medium">{isLoading ? <Skeleton className="h-4 w-24" /> : (profile?.data.username ?? "Unavailable")}</dd>
          </dl>
        </section>
      </TabsContent>
      <TabsContent value="alerts" forceMount className="data-[state=inactive]:hidden">
        <NotificationSettings />
      </TabsContent>
      <TabsContent value="storage" forceMount className="data-[state=inactive]:hidden">
        <S3StorageSettings />
      </TabsContent>
    </Tabs>
  );
}