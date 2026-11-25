"use client";

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { useProfile } from "@/hooks/use-profile";
import { logout } from "@/lib/helper";
import { Skeleton } from "@/components/ui/skeleton";
import { NotificationSettings } from "./notification-settings";
import { S3StorageSettings } from "./s3-storage-settings";
import { PageHeader } from "@/components/layout/page-header";

export function ProfileSettings() {
  const { data: profile, isLoading } = useProfile();
  return <div>
    <PageHeader title="Settings" description="Your account, backup storage, and failure notifications." />
    <div className="max-w-3xl space-y-6">
      <Card>
        <CardHeader><CardTitle className="text-base">Account</CardTitle><CardDescription>You’re signed in to this Velld instance.</CardDescription></CardHeader>
        <CardContent className="flex items-center justify-between gap-4">
          <div className="min-w-0"><p className="text-xs text-muted-foreground">Username</p>{isLoading ? <Skeleton className="mt-2 h-5 w-32" /> : <p className="mt-2 break-words text-sm font-medium">{profile?.data.username || "Unavailable"}</p>}</div>
          <Button variant="outline" onClick={logout}>Sign out</Button>
        </CardContent>
      </Card>
      <S3StorageSettings />
      <NotificationSettings />
    </div>
  </div>;
}