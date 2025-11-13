"use client";

import { useState, useEffect } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Loader2, Cloud, Check, AlertCircle, HelpCircle } from "lucide-react";
import { useSettings } from "@/hooks/use-settings";
import { useToast } from "@/hooks/use-toast";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

const S3_PROVIDERS = [
  { value: "aws", label: "AWS S3", endpoint: "s3.amazonaws.com", region: "us-east-1", ssl: true },
  { value: "minio", label: "MinIO", endpoint: "localhost:9000", region: "us-east-1", ssl: false },
  { value: "backblaze", label: "Backblaze B2", endpoint: "s3.us-west-002.backblazeb2.com", region: "us-west-002", ssl: true },
  { value: "scaleway", label: "Scaleway", endpoint: "s3.fr-par.scw.cloud", region: "fr-par", ssl: true },
  { value: "storj", label: "Storj DCS", endpoint: "gateway.storjshare.io", region: "global", ssl: true },
  { value: "digitalocean", label: "DigitalOcean Spaces", endpoint: "nyc3.digitaloceanspaces.com", region: "nyc3", ssl: true },
  { value: "wasabi", label: "Wasabi", endpoint: "s3.wasabisys.com", region: "us-east-1", ssl: true },
  { value: "custom", label: "Custom / Other", endpoint: "", region: "", ssl: true },
];

export function S3StorageSettings() {
  const { settings, isLoading, updateSettings, isUpdating } = useSettings();
  const { toast } = useToast();
  const [isTesting, setIsTesting] = useState(false);
  const [testResult, setTestResult] = useState<"success" | "error" | null>(null);

  const [formData, setFormData] = useState({
    s3_enabled: false,
    s3_endpoint: "",
    s3_region: "",
    s3_bucket: "",
    s3_access_key: "",
    s3_secret_key: "",
    s3_use_ssl: true,
    s3_path_prefix: "",
    s3_purge_local: false,
  });

  // Sync form data when settings load
  useEffect(() => {
    if (settings) {
      setFormData({
        s3_enabled: settings.s3_enabled || false,
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

  const handleProviderChange = (provider: string) => {
    const providerConfig = S3_PROVIDERS.find(p => p.value === provider);
    if (providerConfig && provider !== "custom") {
      setFormData(prev => ({
        ...prev,
        s3_endpoint: providerConfig.endpoint,
        s3_region: providerConfig.region,
        s3_use_ssl: providerConfig.ssl,
      }));
    }
  };

  const handleTestConnection = async () => {
    setIsTesting(true);
    setTestResult(null);

    try {
      // TODO: Implement test S3 connection endpoint
      // const response = await fetch('/api/settings/test-s3', {
      //   method: 'POST',
      //   headers: { 'Content-Type': 'application/json' },
      //   body: JSON.stringify(formData),
      // });
      
      // Simulate test for now
      await new Promise(resolve => setTimeout(resolve, 1500));
      
      setTestResult("success");
      toast({
        title: "Connection Successful",
        description: "Successfully connected to S3 storage",
      });
    } catch (error) {
      setTestResult("error");
      toast({
        variant: "destructive",
        title: "Connection Failed",
        description: error instanceof Error ? error.message : "Failed to connect to S3 storage",
      });
    } finally {
      setIsTesting(false);
    }
  };

  const handleSave = async () => {
    try {
      // Only send secret key if it was changed (not empty)
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const dataToSave: Record<string, any> = { ...formData };
      if (!dataToSave.s3_secret_key) {
        delete dataToSave.s3_secret_key;
      }
      
      await updateSettings(dataToSave);
      setTestResult(null);
      
      // Clear the secret key field after successful save for security
      setFormData(prev => ({ ...prev, s3_secret_key: "" }));
    } catch (error) {
      toast({
        variant: "destructive",
        title: "Save Failed",
        description: error instanceof Error ? error.message : "Failed to save settings",
      });
    }
  };

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="text-lg flex items-center gap-2">
            <Cloud className="w-5 h-5" />
            S3 Object Storage
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

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-lg flex items-center gap-2">
          <Cloud className="w-5 h-5" />
          S3 Object Storage
        </CardTitle>
        <CardDescription>
          Store backups in S3-compatible object storage (AWS S3, MinIO, Backblaze B2, etc.)
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-6">
        {/* Enable S3 Storage */}
        <div className="flex items-center justify-between p-4 rounded-lg border bg-background/50">
          <div className="space-y-0.5">
            <Label htmlFor="s3-enabled" className="text-sm font-medium">
              Enable S3 Storage
            </Label>
            <p className="text-sm text-muted-foreground">
              Automatically upload backups to S3-compatible storage
            </p>
          </div>
          <div className="flex items-center gap-2">
            <Switch
              id="s3-enabled"
              checked={formData.s3_enabled}
              onCheckedChange={async (checked) => {
                const previousState = formData.s3_enabled;
                setFormData({ ...formData, s3_enabled: checked });
                // Auto-save when toggling on/off
                try {
                  await updateSettings({ s3_enabled: checked });
                  toast({
                    title: "Success",
                    description: `S3 storage ${checked ? 'enabled' : 'disabled'}`,
                  });
                } catch {
                  // Revert the toggle on error
                  setFormData({ ...formData, s3_enabled: previousState });
                  toast({
                    variant: "destructive",
                    title: "Error",
                    description: "Failed to update S3 storage status",
                  });
                }
              }}
            />
          </div>
        </div>

        {formData.s3_enabled && (
          <>
            {/* Provider Preset */}
            <div className="space-y-2">
              <Label htmlFor="provider">
                Provider Preset
                <TooltipProvider>
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <HelpCircle className="inline w-3.5 h-3.5 ml-1.5 text-muted-foreground" />
                    </TooltipTrigger>
                    <TooltipContent>
                      <p>Select a provider to auto-fill endpoint and region settings</p>
                    </TooltipContent>
                  </Tooltip>
                </TooltipProvider>
              </Label>
              <Select onValueChange={handleProviderChange}>
                <SelectTrigger>
                  <SelectValue placeholder="Select a provider or use custom" />
                </SelectTrigger>
                <SelectContent>
                  {S3_PROVIDERS.map((provider) => (
                    <SelectItem key={provider.value} value={provider.value}>
                      {provider.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            {/* Grid Layout for Endpoint and Region */}
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="s3-endpoint">
                  Endpoint
                  <TooltipProvider>
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <HelpCircle className="inline w-3.5 h-3.5 ml-1.5 text-muted-foreground" />
                      </TooltipTrigger>
                      <TooltipContent>
                        <p>S3 API endpoint URL (without https://)</p>
                      </TooltipContent>
                    </Tooltip>
                  </TooltipProvider>
                </Label>
                <Input
                  id="s3-endpoint"
                  placeholder="s3.amazonaws.com"
                  value={formData.s3_endpoint}
                  onChange={(e) => setFormData({ ...formData, s3_endpoint: e.target.value })}
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="s3-region">
                  Region
                  <TooltipProvider>
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <HelpCircle className="inline w-3.5 h-3.5 ml-1.5 text-muted-foreground" />
                      </TooltipTrigger>
                      <TooltipContent>
                        <p>S3 region (e.g., us-east-1)</p>
                      </TooltipContent>
                    </Tooltip>
                  </TooltipProvider>
                </Label>
                <Input
                  id="s3-region"
                  placeholder="us-east-1"
                  value={formData.s3_region}
                  onChange={(e) => setFormData({ ...formData, s3_region: e.target.value })}
                />
              </div>
            </div>

            {/* Bucket Name */}
            <div className="space-y-2">
              <Label htmlFor="s3-bucket">
                Bucket Name
                <TooltipProvider>
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <HelpCircle className="inline w-3.5 h-3.5 ml-1.5 text-muted-foreground" />
                    </TooltipTrigger>
                    <TooltipContent>
                      <p>S3 bucket where backups will be stored</p>
                    </TooltipContent>
                  </Tooltip>
                </TooltipProvider>
              </Label>
              <Input
                id="s3-bucket"
                placeholder="velld-backups"
                value={formData.s3_bucket}
                onChange={(e) => setFormData({ ...formData, s3_bucket: e.target.value })}
              />
            </div>

            {/* Access Key and Secret Key */}
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="s3-access-key">Access Key ID</Label>
                <Input
                  id="s3-access-key"
                  placeholder="AKIAIOSFODNN7EXAMPLE"
                  value={formData.s3_access_key}
                  onChange={(e) => setFormData({ ...formData, s3_access_key: e.target.value })}
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="s3-secret-key">Secret Access Key</Label>
                <Input
                  id="s3-secret-key"
                  type="password"
                  placeholder="wJalrXUtnFEMI/K7MDENG..."
                  value={formData.s3_secret_key}
                  onChange={(e) => setFormData({ ...formData, s3_secret_key: e.target.value })}
                />
              </div>
            </div>

            {/* Path Prefix */}
            <div className="space-y-2">
              <Label htmlFor="s3-path-prefix">
                Path Prefix (Optional)
                <TooltipProvider>
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <HelpCircle className="inline w-3.5 h-3.5 ml-1.5 text-muted-foreground" />
                    </TooltipTrigger>
                    <TooltipContent>
                      <p>Folder path inside bucket (e.g., backups/production)</p>
                    </TooltipContent>
                  </Tooltip>
                </TooltipProvider>
              </Label>
              <Input
                id="s3-path-prefix"
                placeholder="backups/production"
                value={formData.s3_path_prefix}
                onChange={(e) => setFormData({ ...formData, s3_path_prefix: e.target.value })}
              />
            </div>

            {/* SSL/TLS */}
            <div className="flex items-center justify-between p-4 rounded-lg border bg-background/50">
              <div className="space-y-0.5">
                <Label htmlFor="s3-use-ssl" className="text-sm font-medium">
                  Use SSL/TLS
                </Label>
                <p className="text-sm text-muted-foreground">
                  Connect to S3 over HTTPS (recommended)
                </p>
              </div>
              <Switch
                id="s3-use-ssl"
                checked={formData.s3_use_ssl}
                onCheckedChange={(checked) => setFormData({ ...formData, s3_use_ssl: checked })}
              />
            </div>

            {/* Purge Local Backup */}
            <div className="flex items-center justify-between p-4 rounded-lg border bg-background/50">
              <div className="space-y-0.5">
                <Label htmlFor="s3-purge-local" className="text-sm font-medium">
                  Purge Local Backup After Upload
                </Label>
                <p className="text-sm text-muted-foreground">
                  Automatically delete local backup files after successful S3 upload to save disk space
                </p>
              </div>
              <Switch
                id="s3-purge-local"
                checked={formData.s3_purge_local}
                onCheckedChange={(checked) => setFormData({ ...formData, s3_purge_local: checked })}
              />
            </div>

            {/* Test Connection Result */}
            {testResult && (
              <div className={`flex items-center gap-2 p-4 rounded-lg border ${
                testResult === "success" 
                  ? "bg-green-500/10 border-green-500/20 text-green-600 dark:text-green-400" 
                  : "bg-red-500/10 border-red-500/20 text-red-600 dark:text-red-400"
              }`}>
                {testResult === "success" ? (
                  <>
                    <Check className="w-4 h-4" />
                    <span className="text-sm font-medium">Connection successful!</span>
                  </>
                ) : (
                  <>
                    <AlertCircle className="w-4 h-4" />
                    <span className="text-sm font-medium">Connection failed. Check your credentials.</span>
                  </>
                )}
              </div>
            )}

            {/* Action Buttons */}
            <div className="flex gap-2">
              <Button
                onClick={handleTestConnection}
                variant="outline"
                // disabled={isTesting || !formData.s3_endpoint || !formData.s3_bucket}
                disabled={true}
              >
                {isTesting ? (
                  <>
                    <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                    Testing...
                  </>
                ) : (
                  "Test Connection"
                )}
              </Button>

              <Button
                onClick={handleSave}
                disabled={isUpdating || !formData.s3_endpoint || !formData.s3_bucket}
              >
                {isUpdating ? (
                  <>
                    <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                    Saving...
                  </>
                ) : (
                  "Save Settings"
                )}
              </Button>
            </div>
          </>
        )}
      </CardContent>
    </Card>
  );
}
