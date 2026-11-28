"use client";

import { useState, useEffect } from "react";
import { CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { LayerCard as Card } from "@cloudflare/kumo/components/layer-card";
import { Button } from "@cloudflare/kumo/components/button";
import {
  FormInput as Input,
  FormField as SettingsField,
  FormFieldGroup as SettingsFieldGroup,
  FormFieldRow as SettingsFieldRow,
} from "@/components/ui/form-field";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Loader2, Cloud } from "lucide-react";
import { useSettings } from "@/hooks/use-settings";
import { useToast } from "@/hooks/use-toast";
import { testStorageConnection } from "@/lib/api/settings";
import type { StorageProvider } from "@/types/settings";
import { Select } from "@cloudflare/kumo/components/select";

const S3_PROVIDERS = [
  {
    value: "aws",
    label: "AWS S3",
    endpoint: "s3.amazonaws.com",
    region: "us-east-1",
    ssl: true,
  },
  {
    value: "minio",
    label: "MinIO",
    endpoint: "localhost:9000",
    region: "us-east-1",
    ssl: false,
  },
  {
    value: "backblaze",
    label: "Backblaze B2",
    endpoint: "s3.us-west-002.backblazeb2.com",
    region: "us-west-002",
    ssl: true,
  },
  {
    value: "scaleway",
    label: "Scaleway",
    endpoint: "s3.fr-par.scw.cloud",
    region: "fr-par",
    ssl: true,
  },
  {
    value: "storj",
    label: "Storj DCS",
    endpoint: "gateway.storjshare.io",
    region: "global",
    ssl: true,
  },
  {
    value: "digitalocean",
    label: "DigitalOcean Spaces",
    endpoint: "nyc3.digitaloceanspaces.com",
    region: "nyc3",
    ssl: true,
  },
  {
    value: "wasabi",
    label: "Wasabi",
    endpoint: "s3.wasabisys.com",
    region: "us-east-1",
    ssl: true,
  },
  {
    value: "custom",
    label: "Custom / Other",
    endpoint: "",
    region: "",
    ssl: true,
  },
];
const providerItems = Object.fromEntries(
  S3_PROVIDERS.map((provider) => [provider.value, provider.label]),
);

export function S3StorageSettings() {
  const { settings, isLoading, error, updateSettings, isUpdating } =
    useSettings();
  const { toast } = useToast();

  const [formData, setFormData] = useState({
    s3_enabled: false,
    s3_provider: "s3" as StorageProvider,
    r2_account_id: "",
    s3_endpoint: "",
    s3_region: "",
    s3_bucket: "",
    s3_access_key: "",
    s3_secret_key: "",
    s3_use_ssl: true,
    s3_path_prefix: "",
    s3_purge_local: false,
  });
  const [isTesting, setIsTesting] = useState(false);
  const [testResult, setTestResult] = useState<{
    form: typeof formData;
    error?: string;
  }>();
  const isR2 = formData.s3_provider === "r2";
  const canSubmit =
    formData.s3_bucket.trim() &&
    (isR2 ? formData.r2_account_id.trim() : formData.s3_endpoint.trim());

  // Sync form data when settings load
  useEffect(() => {
    if (settings) {
      setFormData({
        s3_enabled: settings.s3_enabled || false,
        s3_provider: settings.s3_provider ?? "s3",
        r2_account_id: settings.r2_account_id || "",
        s3_endpoint: settings.s3_endpoint || "",
        s3_region: settings.s3_region || "",
        s3_bucket: settings.s3_bucket || "",
        s3_access_key: settings.s3_access_key || "",
        s3_secret_key: "", // Don't show the secret key for security
        s3_use_ssl: settings.s3_use_ssl ?? true,
        s3_path_prefix: settings.s3_path_prefix || "",
        s3_purge_local: settings.s3_purge_local || false,
      });
    }
  }, [settings]);

  const handlePresetChange = (provider: string) => {
    const providerConfig = S3_PROVIDERS.find((p) => p.value === provider);
    if (providerConfig && provider !== "custom") {
      setFormData((prev) => ({
        ...prev,
        s3_endpoint: providerConfig.endpoint,
        s3_region: providerConfig.region,
        s3_use_ssl: providerConfig.ssl,
      }));
    }
  };

  const storageRequest = () => ({
    ...formData,
    s3_secret_key: formData.s3_secret_key || undefined,
  });

  const handleTest = async () => {
    setIsTesting(true);
    setTestResult(undefined);
    try {
      await testStorageConnection(storageRequest());
      setTestResult({ form: formData });
    } catch (error) {
      setTestResult({
        form: formData,
        error:
          error instanceof Error
            ? error.message
            : "Unable to test the storage connection.",
      });
    } finally {
      setIsTesting(false);
    }
  };

  const handleSave = async () => {
    try {
      await updateSettings(storageRequest());

      // Clear the secret key field after successful save for security
      setFormData((prev) => ({ ...prev, s3_secret_key: "" }));
    } catch (error) {
      toast({
        variant: "destructive",
        title: "Save Failed",
        description:
          error instanceof Error ? error.message : "Failed to save settings",
      });
    }
  };

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="text-lg flex items-center gap-2">
            <Cloud className="w-5 h-5" />
            Backup storage
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div className="h-10 bg-muted animate-pulse rounded" />
            <div className="h-10 bg-muted animate-pulse rounded" />
            <div className="h-10 bg-muted animate-pulse rounded" />
          </div>
        </CardContent>
      </Card>
    );
  }

  if (error)
    return (
      <section id="storage" className="scroll-mt-20">
        <h2 className="text-[15px] font-semibold">Storage</h2>
        <p role="alert" className="mt-3 text-sm text-destructive">
          Couldn’t load storage settings. Please try again.
        </p>
      </section>
    );

  return (
    <section
      id="storage"
      className="scroll-mt-20"
      aria-labelledby="storage-heading"
    >
      <h2
        id="storage-heading"
        className="text-xl font-semibold tracking-[-0.025em]"
      >
        Storage
      </h2>
      <Card className="mt-7 rounded-lg shadow-none">
        <CardContent className="space-y-6 p-5">
          {/* Enable S3 Storage */}
          <div className="flex items-center justify-between gap-4">
            <div className="space-y-1">
              <Label htmlFor="s3-enabled" className="text-sm font-medium">
                Object storage
              </Label>
              <p className="text-sm text-muted-foreground">
                Saves immediately.
              </p>
            </div>
            <div className="flex items-center gap-2">
              <Switch
                id="s3-enabled"
                aria-label="Object storage"
                disabled={isUpdating || isTesting}
                checked={formData.s3_enabled}
                onCheckedChange={async (checked) => {
                  const previousState = formData.s3_enabled;
                  setFormData({ ...formData, s3_enabled: checked });
                  // Auto-save when toggling on/off
                  try {
                    await updateSettings({ s3_enabled: checked });
                    toast({
                      title: "Success",
                      description: `Object storage ${checked ? "enabled" : "disabled"}`,
                    });
                  } catch {
                    // Revert the toggle on error
                    setFormData({ ...formData, s3_enabled: previousState });
                    toast({
                      variant: "destructive",
                      title: "Error",
                      description: "Failed to update storage status",
                    });
                  }
                }}
              />
            </div>
          </div>

          {formData.s3_enabled && (
            <>
              <SettingsFieldGroup>
                <SettingsField id="storage-provider" label="Storage provider">
                  <Select
                    id="storage-provider"
                    aria-label="Storage provider"
                    value={formData.s3_provider}
                    onValueChange={(value) => {
                      if (value)
                        setFormData({
                          ...formData,
                          s3_provider: value as StorageProvider,
                        });
                    }}
                    items={{ s3: "S3-compatible", r2: "Cloudflare R2" }}
                    className="w-full"
                  >
                    <Select.Option value="s3">S3-compatible</Select.Option>
                    <Select.Option value="r2">Cloudflare R2</Select.Option>
                  </Select>
                </SettingsField>
                {isR2 ? (
                  <SettingsField
                    id="r2-account-id"
                    label="Cloudflare Account ID"
                  >
                    <Input
                      id="r2-account-id"
                      aria-label="Cloudflare Account ID"
                      placeholder="32-character account ID"
                      value={formData.r2_account_id}
                      onChange={(e) =>
                        setFormData({
                          ...formData,
                          r2_account_id: e.target.value,
                        })
                      }
                    />
                  </SettingsField>
                ) : (
                  <>
                    <SettingsField
                      id="provider"
                      label="Provider preset"
                    >
                      <Select
                        id="provider"
                        aria-label="Provider preset"
                        onValueChange={(value) => {
                          if (typeof value === "string")
                            handlePresetChange(value);
                        }}
                        items={providerItems}
                        placeholder="Select a provider or use custom"
                        className="w-full"
                      >
                        {S3_PROVIDERS.map((provider) => (
                          <Select.Option
                            key={provider.value}
                            value={provider.value}
                          >
                            {provider.label}
                          </Select.Option>
                        ))}
                      </Select>
                    </SettingsField>

                    {/* Grid Layout for Endpoint and Region */}
                    <SettingsFieldRow>
                      <SettingsField
                        id="s3-endpoint"
                        label="Endpoint"
                        hint="Enter the host without https://."
                      >
                        <Input
                          id="s3-endpoint"
                          aria-label="Endpoint"
                          aria-describedby="s3-endpoint-hint"
                          placeholder="s3.amazonaws.com"
                          value={formData.s3_endpoint}
                          onChange={(e) =>
                            setFormData({
                              ...formData,
                              s3_endpoint: e.target.value,
                            })
                          }
                        />
                      </SettingsField>

                      <SettingsField id="s3-region" label="Region">
                        <Input
                          id="s3-region"
                          aria-label="Region"
                          placeholder="us-east-1"
                          value={formData.s3_region}
                          onChange={(e) =>
                            setFormData({
                              ...formData,
                              s3_region: e.target.value,
                            })
                          }
                        />
                      </SettingsField>
                    </SettingsFieldRow>
                  </>
                )}

                {/* Bucket Name */}
                <SettingsField id="s3-bucket" label="Bucket name">
                  <Input
                    id="s3-bucket"
                    aria-label="Bucket name"
                    placeholder="velld-backups"
                    value={formData.s3_bucket}
                    onChange={(e) =>
                      setFormData({ ...formData, s3_bucket: e.target.value })
                    }
                  />
                </SettingsField>

                {/* Access Key and Secret Key */}
                <SettingsFieldRow>
                  <SettingsField id="s3-access-key" label="Access Key ID">
                    <Input
                      id="s3-access-key"
                      aria-label="Access Key ID"
                      placeholder={isR2 ? undefined : "AKIAIOSFODNN7EXAMPLE"}
                      value={formData.s3_access_key}
                      onChange={(e) =>
                        setFormData({
                          ...formData,
                          s3_access_key: e.target.value,
                        })
                      }
                    />
                  </SettingsField>

                  <SettingsField id="s3-secret-key" label="Secret Access Key">
                    <Input
                      id="s3-secret-key"
                      aria-label="Secret Access Key"
                      type="password"
                      placeholder="Leave blank to use the saved key"
                      value={formData.s3_secret_key}
                      onChange={(e) =>
                        setFormData({
                          ...formData,
                          s3_secret_key: e.target.value,
                        })
                      }
                    />
                  </SettingsField>
                </SettingsFieldRow>

                {/* Path Prefix */}
                <SettingsField
                  id="s3-path-prefix"
                  label="Path prefix (optional)"
                >
                  <Input
                    id="s3-path-prefix"
                    aria-label="Path prefix (optional)"
                    placeholder="backups/production"
                    value={formData.s3_path_prefix}
                    onChange={(e) =>
                      setFormData({
                        ...formData,
                        s3_path_prefix: e.target.value,
                      })
                    }
                  />
                </SettingsField>
              </SettingsFieldGroup>

              {!isR2 && (
                <div className="flex items-center justify-between gap-4 border-t pt-5">
                  <div className="space-y-0.5">
                    <Label htmlFor="s3-use-ssl" className="text-sm font-medium">
                      Use HTTPS
                    </Label>
                  </div>
                  <Switch
                    id="s3-use-ssl"
                    aria-label="Use HTTPS"
                    checked={formData.s3_use_ssl}
                    onCheckedChange={(checked) =>
                      setFormData({ ...formData, s3_use_ssl: checked })
                    }
                  />
                </div>
              )}

              {/* Purge Local Backup */}
              <div className="flex items-center justify-between gap-4 border-t pt-5">
                <div className="space-y-0.5">
                  <Label
                    htmlFor="s3-purge-local"
                    className="text-sm font-medium"
                  >
                    Delete local files after upload
                  </Label>
                  <p className="text-sm text-muted-foreground">
                    Restores use the object-storage copy.
                  </p>
                </div>
                <Switch
                  id="s3-purge-local"
                  aria-label="Delete local files after upload"
                  checked={formData.s3_purge_local}
                  onCheckedChange={(checked) =>
                    setFormData({ ...formData, s3_purge_local: checked })
                  }
                />
              </div>
            </>
          )}
        </CardContent>
      </Card>
      {formData.s3_enabled && (
        <div className="mt-6 space-y-4">
          {testResult?.form === formData && (
            <p
              role={testResult.error ? "alert" : "status"}
              className={
                testResult.error
                  ? "text-sm text-destructive"
                  : "text-sm text-muted-foreground"
              }
            >
              {testResult.error
                ? `Connection failed: ${testResult.error}`
                : "Connection successful."}
            </p>
          )}
          <div className="flex flex-wrap justify-end gap-3">
            <Button
              variant="secondary"
              onClick={handleTest}
              disabled={isUpdating || isTesting || !canSubmit}
            >
              {isTesting ? "Testing…" : "Test connection"}
            </Button>
            <Button
              variant="primary"
              onClick={handleSave}
              disabled={isUpdating || isTesting || !canSubmit}
            >
              {isUpdating ? (
                <>
                  <Loader2 className="h-4 w-4 animate-spin" />
                  Saving…
                </>
              ) : (
                "Save settings"
              )}
            </Button>
          </div>
        </div>
      )}
    </section>
  );
}
