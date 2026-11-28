"use client";

import { useState } from "react";
import { CardContent, CardHeader } from "@/components/ui/card";
import { LayerCard as Card } from "@cloudflare/kumo/components/layer-card";
import { Button } from "@cloudflare/kumo/components/button";
import { FormInput as Input, FormField as SettingsField, FormFieldGroup as SettingsFieldGroup, FormFieldRow as SettingsFieldRow } from "@/components/ui/form-field";
import { Switch } from "@/components/ui/switch";
import { Mail, Webhook } from "lucide-react";
import { useSettings } from "@/hooks/use-settings";
import { Skeleton } from "@/components/ui/skeleton";
import { UpdateSettingsRequest } from "@/types/settings";

type SettingsValue = string | number | boolean;

export function NotificationSettings() {
  const { settings, isLoading, error, updateSettings, isUpdating } = useSettings();
  const [formValues, setFormValues] = useState<UpdateSettingsRequest>({});

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <Skeleton className="h-6 w-48" />
          <Skeleton className="h-4 w-64 mt-2" />
        </CardHeader>
        <CardContent className="space-y-6">
          {[1, 2].map((i) => (
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

  if (error) return <section id="notifications" className="scroll-mt-20"><h2 className="text-[15px] font-semibold">Notifications</h2><p role="alert" className="mt-3 text-sm text-destructive">Couldn’t load notification settings. Please try again.</p></section>;

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (Object.keys(formValues).length > 0) {
      void updateSettings(formValues).then(() => setFormValues({})).catch(() => {});
    }
  };

  const handleSwitchChange = async (field: keyof UpdateSettingsRequest, checked: boolean) => {
    try {
      await updateSettings({ [field]: checked });
    } catch (error) {
      console.error(`Failed to update ${field}:`, error);
    }
  };

  const handleChange = (field: keyof UpdateSettingsRequest, value: SettingsValue) => {
    setFormValues(prev => ({ ...prev, [field]: value }));
  };

  return (
    <section id="notifications" className="scroll-mt-20" aria-labelledby="notifications-heading">
      <h2 id="notifications-heading" className="text-xl font-semibold tracking-[-0.025em]">Notifications</h2>
      <p className="mt-1.5 text-sm text-muted-foreground">Choose how to hear about failed backups. Switches save immediately; changes to fields must be saved.</p>
      <Card className="mt-7 rounded-lg shadow-none">
      <CardContent className="p-5">
        <form onSubmit={handleSubmit} className="space-y-6">
          {/* Email Notifications */}
          <div className="space-y-3">
            <div className="flex items-center justify-between gap-4">
              <div className="flex items-center gap-3">
                <div className="rounded-md bg-muted p-2">
                  <Mail className="h-4 w-4 text-muted-foreground" />
                </div>
                <div>
                  <p className="text-sm font-medium">Email Notifications</p>
                  <p className="text-xs text-muted-foreground">
                    Receive notifications via email
                  </p>
                </div>
              </div>
              <Switch
                aria-label="Email notifications"
                checked={settings?.notify_email ?? false}
                disabled={isUpdating}
                onCheckedChange={(checked) => handleSwitchChange('notify_email', checked)}
              />
            </div>

            {settings?.notify_email && (
              <div className="space-y-4 border-l pl-4 sm:ml-4 sm:pl-6">
                <SettingsFieldGroup>
                <SettingsField id="notification-email" label={<>
                    Email Address
                    {settings?.env_configured?.email && (
                      <span className="ml-2 text-xs text-muted-foreground">(set via environment variable)</span>
                    )}
                  </>}>
                  <Input
                    id="notification-email"
                    aria-label="Email Address"
                    type="email"
                    placeholder="your@email.com"
                    defaultValue={settings?.email || ''}
                    onChange={(e) => handleChange('email', e.target.value)}
                    disabled={settings?.env_configured?.email}
                    className={settings?.env_configured?.email ? 'bg-muted cursor-not-allowed' : ''}
                  />
                </SettingsField>
                
                <SettingsFieldRow>
                  <SettingsField id="smtp-host" label={<>
                      SMTP Host
                      {settings?.env_configured?.smtp_host && (
                        <span className="ml-2 text-xs text-muted-foreground">(env var)</span>
                      )}
                    </>}>
                    <Input
                      id="smtp-host"
                      aria-label="SMTP Host"
                      placeholder="smtp.gmail.com"
                      defaultValue={settings.smtp_host}
                      onChange={(e) => handleChange('smtp_host', e.target.value)}
                      disabled={settings?.env_configured?.smtp_host}
                      className={settings?.env_configured?.smtp_host ? 'bg-muted cursor-not-allowed' : ''}
                    />
                  </SettingsField>
                  <SettingsField id="smtp-port" label={<>
                      SMTP Port
                      {settings?.env_configured?.smtp_port && (
                        <span className="ml-2 text-xs text-muted-foreground">(env var)</span>
                      )}
                    </>}>
                    <Input
                      id="smtp-port"
                      aria-label="SMTP Port"
                      type="number"
                      placeholder="587"
                      defaultValue={settings?.smtp_port?.toString() || ''}
                      onChange={(e) => handleChange('smtp_port', parseInt(e.target.value))}
                      disabled={settings?.env_configured?.smtp_port}
                      className={settings?.env_configured?.smtp_port ? 'bg-muted cursor-not-allowed' : ''}
                    />
                  </SettingsField>
                </SettingsFieldRow>

                <SettingsField id="smtp-username" label={<>
                    SMTP Username
                    {settings?.env_configured?.smtp_username && (
                      <span className="ml-2 text-xs text-muted-foreground">(env var)</span>
                    )}
                  </>}>
                  <Input
                    id="smtp-username"
                    aria-label="SMTP Username"
                    placeholder="username"
                    defaultValue={settings.smtp_username}
                    onChange={(e) => handleChange('smtp_username', e.target.value)}
                    disabled={settings?.env_configured?.smtp_username}
                    className={settings?.env_configured?.smtp_username ? 'bg-muted cursor-not-allowed' : ''}
                  />
                </SettingsField>
                
                <SettingsField id="smtp-password" hint="Leave empty to keep current password" label={<>
                    SMTP Password
                    {settings?.env_configured?.smtp_password && (
                      <span className="ml-2 text-xs text-muted-foreground">(env var)</span>
                    )}
                  </>}>
                  <Input
                    id="smtp-password"
                    aria-describedby="smtp-password-hint"
                    aria-label="SMTP Password"
                    type="password"
                    placeholder="••••••••"
                    onChange={(e) => handleChange('smtp_password', e.target.value)}
                    disabled={settings?.env_configured?.smtp_password}
                    className={settings?.env_configured?.smtp_password ? 'bg-muted cursor-not-allowed' : ''}
                  />
                </SettingsField>
                </SettingsFieldGroup>
              </div>
            )}
          </div>

          {/* Webhook Notifications */}
          <div className="space-y-3 border-t pt-6">
            <div className="flex items-center justify-between gap-4">
              <div className="flex items-center gap-3">
                <div className="rounded-md bg-muted p-2">
                  <Webhook className="h-4 w-4 text-muted-foreground" />
                </div>
                <div>
                  <p className="text-sm font-medium">Webhook Notifications</p>
                  <p className="text-xs text-muted-foreground">
                    Send notifications to a webhook URL
                  </p>
                </div>
              </div>
              <Switch
                aria-label="Webhook notifications"
                checked={settings?.notify_webhook ?? false}
                disabled={isUpdating}
                onCheckedChange={(checked) => handleSwitchChange('notify_webhook', checked)}
              />
            </div>

            {settings?.notify_webhook && (
              <div className="border-l pl-4 sm:ml-4 sm:pl-6">
                <SettingsField id="notification-webhook" label="Webhook URL" hint="Supports Slack, Discord, and custom webhooks">
                  <Input
                    id="notification-webhook"
                    aria-describedby="notification-webhook-hint"
                    aria-label="Webhook URL"
                    placeholder="https://hooks.slack.com/services/..."
                    defaultValue={settings.webhook_url}
                    onChange={(e) => handleChange('webhook_url', e.target.value)}
                  />
                </SettingsField>
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
    </section>
  );
}
