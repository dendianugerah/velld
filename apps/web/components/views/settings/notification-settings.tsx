"use client";

import { useState } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Bell, Mail, Webhook } from "lucide-react";
import { useSettings } from "@/hooks/use-settings";
import { Skeleton } from "@/components/ui/skeleton";
import { UpdateSettingsRequest } from "@/types/settings";

type SettingsValue = string | number | boolean;

export function NotificationSettings() {
  const { settings, isLoading, updateSettings, isUpdating } = useSettings();
  const [formValues, setFormValues] = useState<UpdateSettingsRequest>({});

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <Skeleton className="h-6 w-48" />
          <Skeleton className="h-4 w-64 mt-2" />
        </CardHeader>
        <CardContent className="space-y-6">
          {[1, 2, 3].map((i) => (
            <div key={i} className="flex items-center justify-between p-4 rounded-lg border">
              <div className="flex items-center gap-3">
                <Skeleton className="h-10 w-10 rounded-md" />
                <div>
                  <Skeleton className="h-5 w-32" />
                  <Skeleton className="h-4 w-48 mt-1" />
                </div>
              </div>
              <Skeleton className="h-6 w-12" />
            </div>
          ))}
        </CardContent>
      </Card>
    );
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (Object.keys(formValues).length > 0) {
      updateSettings(formValues);
    }
  };

  const handleSwitchChange = async (field: keyof UpdateSettingsRequest, checked: boolean) => {
    try {
      setFormValues(prev => ({ ...prev, [field]: checked }));
      await updateSettings({ [field]: checked });
    } catch (error) {
      setFormValues(prev => ({ ...prev, [field]: !checked }));
      console.error(`Failed to update ${field}:`, error);
    }
  };

  const handleChange = (field: keyof UpdateSettingsRequest, value: SettingsValue) => {
    setFormValues(prev => ({ ...prev, [field]: value }));
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-lg">Notifications</CardTitle>
        <CardDescription>Configure how you receive backup notifications</CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="space-y-6">
          {/* Dashboard Notifications */}
          <div className="flex items-center justify-between p-4 rounded-lg border bg-background/50">
            <div className="flex items-center gap-3">
              <div className="p-2 rounded-md bg-primary/10">
                <Bell className="w-4 h-4 text-primary" />
              </div>
              <div>
                <p className="text-sm font-medium">Dashboard Notifications</p>
                <p className="text-xs text-muted-foreground">
                  Show notifications in the dashboard
                </p>
              </div>
            </div>
            <Switch
              checked={settings?.notify_dashboard ?? false}
              disabled={isUpdating}
              onCheckedChange={(checked) => handleSwitchChange('notify_dashboard', checked)}
            />
          </div>

          {/* Email Notifications */}
          <div className="space-y-3">
            <div className="flex items-center justify-between p-4 rounded-lg border bg-background/50">
              <div className="flex items-center gap-3">
                <div className="p-2 rounded-md bg-primary/10">
                  <Mail className="w-4 h-4 text-primary" />
                </div>
                <div>
                  <p className="text-sm font-medium">Email Notifications</p>
                  <p className="text-xs text-muted-foreground">
                    Receive notifications via email
                  </p>
                </div>
              </div>
              <Switch
                checked={settings?.notify_email ?? false}
                disabled={isUpdating}
                onCheckedChange={(checked) => handleSwitchChange('notify_email', checked)}
              />
            </div>

            {settings?.notify_email && (
              <div className="ml-4 pl-6 border-l-2 space-y-4">
                <div className="grid gap-2">
                  <Label className="text-sm">Email Address</Label>
                  <Input
                    type="email"
                    placeholder="your@email.com"
                    defaultValue={settings?.email || ''}
                    onChange={(e) => handleChange('email', e.target.value)}
                  />
                </div>
                
                <div className="grid grid-cols-2 gap-4">
                  <div className="grid gap-2">
                    <Label className="text-sm">SMTP Host</Label>
                    <Input
                      placeholder="smtp.gmail.com"
                      defaultValue={settings.smtp_host}
                      onChange={(e) => handleChange('smtp_host', e.target.value)}
                    />
                  </div>
                  <div className="grid gap-2">
                    <Label className="text-sm">SMTP Port</Label>
                    <Input
                      type="number"
                      placeholder="587"
                      defaultValue={settings?.smtp_port?.toString() || ''}
                      onChange={(e) => handleChange('smtp_port', parseInt(e.target.value))}
                    />
                  </div>
                </div>

                <div className="grid gap-2">
                  <Label className="text-sm">SMTP Username</Label>
                  <Input
                    placeholder="username"
                    defaultValue={settings.smtp_username}
                    onChange={(e) => handleChange('smtp_username', e.target.value)}
                  />
                </div>
                
                <div className="grid gap-2">
                  <Label className="text-sm">SMTP Password</Label>
                  <Input
                    type="password"
                    placeholder="••••••••"
                    onChange={(e) => handleChange('smtp_password', e.target.value)}
                  />
                  <p className="text-xs text-muted-foreground">
                    Leave empty to keep current password
                  </p>
                </div>
              </div>
            )}
          </div>

          {/* Webhook Notifications */}
          <div className="space-y-3">
            <div className="flex items-center justify-between p-4 rounded-lg border bg-background/50">
              <div className="flex items-center gap-3">
                <div className="p-2 rounded-md bg-primary/10">
                  <Webhook className="w-4 h-4 text-primary" />
                </div>
                <div>
                  <p className="text-sm font-medium">Webhook Notifications</p>
                  <p className="text-xs text-muted-foreground">
                    Send notifications to a webhook URL
                  </p>
                </div>
              </div>
              <Switch
                checked={settings?.notify_webhook ?? false}
                disabled={isUpdating}
                onCheckedChange={(checked) => handleSwitchChange('notify_webhook', checked)}
              />
            </div>

            {settings?.notify_webhook && (
              <div className="ml-4 pl-6 border-l-2">
                <div className="grid gap-2">
                  <Label className="text-sm">Webhook URL</Label>
                  <Input
                    placeholder="https://hooks.slack.com/services/..."
                    defaultValue={settings.webhook_url}
                    onChange={(e) => handleChange('webhook_url', e.target.value)}
                  />
                  <p className="text-xs text-muted-foreground">
                    Supports Slack, Discord, and custom webhooks
                  </p>
                </div>
              </div>
            )}
          </div>

          {Object.keys(formValues).length > 0 && (
            <div className="pt-4 border-t flex justify-end">
              <Button type="submit" disabled={isUpdating}>
                {isUpdating ? "Saving..." : "Save Changes"}
              </Button>
            </div>
          )}
        </form>
      </CardContent>
    </Card>
  );
}
