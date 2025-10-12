"use client";

import { useState } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Bell, Mail, Info } from "lucide-react";
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
          <div className="flex items-start gap-3 p-4 rounded-lg border bg-blue-50 dark:bg-blue-950/20 border-blue-200 dark:border-blue-900">
            <Info className="w-5 h-5 text-blue-600 dark:text-blue-400 mt-0.5 flex-shrink-0" />
            <div className="space-y-1">
              <p className="text-sm font-medium text-blue-900 dark:text-blue-100">
                Notifications for Failed Backups Only
              </p>
              <p className="text-xs text-blue-700 dark:text-blue-300">
                You will only receive notifications when a backup fails. Successful backups run silently in the background.
              </p>
            </div>
          </div>

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
                  <Label className="text-sm">
                    Email Address
                    {settings?.env_configured?.email && (
                      <span className="ml-2 text-xs text-muted-foreground">(set via environment variable)</span>
                    )}
                  </Label>
                  <Input
                    type="email"
                    placeholder="your@email.com"
                    defaultValue={settings?.email || ''}
                    onChange={(e) => handleChange('email', e.target.value)}
                    disabled={settings?.env_configured?.email}
                    className={settings?.env_configured?.email ? 'bg-muted cursor-not-allowed' : ''}
                  />
                </div>
                
                <div className="grid grid-cols-2 gap-4">
                  <div className="grid gap-2">
                    <Label className="text-sm">
                      SMTP Host
                      {settings?.env_configured?.smtp_host && (
                        <span className="ml-2 text-xs text-muted-foreground">(env var)</span>
                      )}
                    </Label>
                    <Input
                      placeholder="smtp.gmail.com"
                      defaultValue={settings.smtp_host}
                      onChange={(e) => handleChange('smtp_host', e.target.value)}
                      disabled={settings?.env_configured?.smtp_host}
                      className={settings?.env_configured?.smtp_host ? 'bg-muted cursor-not-allowed' : ''}
                    />
                  </div>
                  <div className="grid gap-2">
                    <Label className="text-sm">
                      SMTP Port
                      {settings?.env_configured?.smtp_port && (
                        <span className="ml-2 text-xs text-muted-foreground">(env var)</span>
                      )}
                    </Label>
                    <Input
                      type="number"
                      placeholder="587"
                      defaultValue={settings?.smtp_port?.toString() || ''}
                      onChange={(e) => handleChange('smtp_port', parseInt(e.target.value))}
                      disabled={settings?.env_configured?.smtp_port}
                      className={settings?.env_configured?.smtp_port ? 'bg-muted cursor-not-allowed' : ''}
                    />
                  </div>
                </div>

                <div className="grid gap-2">
                  <Label className="text-sm">
                    SMTP Username
                    {settings?.env_configured?.smtp_username && (
                      <span className="ml-2 text-xs text-muted-foreground">(env var)</span>
                    )}
                  </Label>
                  <Input
                    placeholder="username"
                    defaultValue={settings.smtp_username}
                    onChange={(e) => handleChange('smtp_username', e.target.value)}
                    disabled={settings?.env_configured?.smtp_username}
                    className={settings?.env_configured?.smtp_username ? 'bg-muted cursor-not-allowed' : ''}
                  />
                </div>
                
                <div className="grid gap-2">
                  <Label className="text-sm">
                    SMTP Password
                    {settings?.env_configured?.smtp_password && (
                      <span className="ml-2 text-xs text-muted-foreground">(env var)</span>
                    )}
                  </Label>
                  <Input
                    type="password"
                    placeholder="••••••••"
                    onChange={(e) => handleChange('smtp_password', e.target.value)}
                    disabled={settings?.env_configured?.smtp_password}
                    className={settings?.env_configured?.smtp_password ? 'bg-muted cursor-not-allowed' : ''}
                  />
                  <p className="text-xs text-muted-foreground">
                    Leave empty to keep current password
                  </p>
                </div>
              </div>
            )}
          </div>

          {/* Webhook Notifications - Commented out until proper implementation */}
          {/* TODO: Implement proper webhook formatting for Slack/Discord/etc */}
          {/* <div className="space-y-3">
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
          </div> */}

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
