"use client";

import { useState } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { LogOut, Shield, User } from "lucide-react";
import { useProfile } from "@/hooks/use-profile";
import { logout } from "@/lib/helper";
import { Skeleton } from "@/components/ui/skeleton";
import { NotificationSettings } from "./notification-settings";
import { PageHeader } from "@/components/layout/page-header";

export function ProfileSettings() {
  const [isLoading, setIsLoading] = useState(false);
  const { data: profile, isLoading: isLoadingProfile } = useProfile();

  const handleLogout = () => {
    setIsLoading(true);
    logout();
  };

  return (
    <div className="space-y-6">
      <PageHeader
        title="Settings"
        description="Manage your account and notification preferences"
      />

      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Profile</CardTitle>
          <CardDescription>Your account information</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center justify-between p-4 rounded-lg border bg-background/50">
            <div className="flex items-center gap-3">
              <div className="p-2 rounded-md bg-primary/10">
                <User className="w-4 h-4 text-primary" />
              </div>
              <div>
                <p className="text-sm font-medium">Username</p>
                {isLoadingProfile ? (
                  <Skeleton className="h-4 w-32 mt-1" />
                ) : (
                  <p className="text-sm text-muted-foreground">{profile?.data.username}</p>
                )}
              </div>
            </div>
          </div>

          <div className="flex items-center justify-between p-4 rounded-lg border bg-background/50">
            <div className="flex items-center gap-3">
              <div className="p-2 rounded-md bg-primary/10">
                <Shield className="w-4 h-4 text-primary" />
              </div>
              <div>
                <p className="text-sm font-medium">Password</p>
                <p className="text-sm text-muted-foreground">Change your account password</p>
              </div>
            </div>
            <Button variant="outline" size="sm" disabled>
              Change
            </Button>
          </div>
        </CardContent>
      </Card>

      <NotificationSettings />

      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Sign Out</CardTitle>
          <CardDescription>End your current session</CardDescription>
        </CardHeader>
        <CardContent>
          <Button 
            variant="destructive" 
            onClick={handleLogout}
            disabled={isLoading}
            className="w-full sm:w-auto"
          >
            <LogOut className="w-4 h-4 mr-2" />
            {isLoading ? "Signing out..." : "Sign out"}
          </Button>
        </CardContent>
      </Card>
    </div>
  );
}