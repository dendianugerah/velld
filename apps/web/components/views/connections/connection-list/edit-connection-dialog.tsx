'use client';

import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Info, Loader2, ChevronDown, ChevronUp, KeyRound } from "lucide-react";
import { useState, useEffect } from "react";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Switch } from "@/components/ui/switch";
import { Select, SelectContent, SelectItem, SelectSeparator, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import { useConnections, useConnection } from "@/hooks/use-connections";
import { type ConnectionForm as ConnectionFormType } from "@/types/connection";
import { type DatabaseType } from "@/types/base";

interface EditConnectionDialogProps {
  connectionId: string | null;
  onClose: () => void;
}

export function EditConnectionDialog({
  connectionId,
  onClose,
}: EditConnectionDialogProps) {
  const [formData, setFormData] = useState<ConnectionFormType & { id?: string }>({
    name: "",
    type: "" as DatabaseType,
    host: "",
    port: 0,
    username: "",
    password: "",
    database: "",
    ssl: true,
    ssh_enabled: false,
    ssh_host: "",
    ssh_port: 0,
    ssh_username: "",
    ssh_password: "",
    ssh_private_key: "",
  });
  const [sshExpanded, setSSHExpanded] = useState(false);
  const [sshAuthMethod, setSSHAuthMethod] = useState<"password" | "key">("password");

  const { editConnection, isEditing } = useConnections();
  const { data: connectionDetail, isLoading: isLoadingConnection } = useConnection(connectionId);

  useEffect(() => {
    if (!connectionId) return;

    if (isLoadingConnection) {
      setFormData({
        name: "",
        type: "" as DatabaseType,
        host: "",
        port: 0,
        username: "",
        password: "",
        database: "",
        ssl: true,
        ssh_enabled: false,
        ssh_host: "",
        ssh_port: 0,
        ssh_username: "",
        ssh_password: "",
        ssh_private_key: "",
      });
      setSSHExpanded(false);
      setSSHAuthMethod("password");
    } else if (connectionDetail) {
      setFormData({
        id: connectionDetail.id,
        name: connectionDetail.name,
        type: connectionDetail.type,
        host: connectionDetail.host,
        port: connectionDetail.port,
        username: connectionDetail.username,
        password: connectionDetail.password,
        database: connectionDetail.database_name,
        ssl: connectionDetail.ssl,
        ssh_enabled: connectionDetail.ssh_enabled,
        ssh_host: connectionDetail.ssh_host || "",
        ssh_port: connectionDetail.ssh_port || 0,
        ssh_username: connectionDetail.ssh_username || "",
        ssh_password: connectionDetail.ssh_password || "",
        ssh_private_key: connectionDetail.ssh_private_key || "",
      });
      setSSHExpanded(connectionDetail.ssh_enabled);
      setSSHAuthMethod(connectionDetail.ssh_private_key ? "key" : "password");
    }
  }, [connectionId, isLoadingConnection, connectionDetail]);

  const getDefaultPort = (type: string): number => {
    const ports: Record<string, number> = {
      'postgresql': 5432,
      'mysql': 3306,
      'mongodb': 27017,
      'redis': 6379,
    };
    return ports[type] || 5432;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!formData.id) return;
    editConnection(formData as ConnectionFormType & { id: string }, {
      onSuccess: () => {
        onClose();
      },
    });
  };

  if (!connectionId) return null;

  return (
    <Dialog open={!!connectionId} onOpenChange={onClose}>
      <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Edit Connection</DialogTitle>
          <DialogDescription>
            Update your database connection settings
          </DialogDescription>
        </DialogHeader>

        {isLoadingConnection || !formData.id ? (
          <div className="flex items-center justify-center p-8">
            <Loader2 className="h-8 w-8 animate-spin text-primary" />
          </div>
        ) : (
          <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="edit-name">Connection Name</Label>
            <Input
              id="edit-name"
              required
              value={formData.name || ''}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="edit-type">Database Type</Label>
            <Select
              value={formData.type}
              onValueChange={(value) => {
                const port = getDefaultPort(value);
                setFormData({ ...formData, type: value as DatabaseType, port });
              }}
            >
              <SelectTrigger>
                <SelectValue placeholder="Select a database" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="postgresql">PostgreSQL</SelectItem>
                <SelectItem value="mysql">MySQL</SelectItem>
                <SelectSeparator />
                <SelectItem value="mongodb">MongoDB</SelectItem>
                <SelectItem value="redis">Redis</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="edit-host">Host</Label>
              <Input
                id="edit-host"
                required
                value={formData.host || ''}
                onChange={(e) => setFormData({ ...formData, host: e.target.value })}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="edit-port">Port</Label>
              <Input
                id="edit-port"
                type="number"
                required
                value={formData.port || ''}
                onChange={(e) => setFormData({ ...formData, port: parseInt(e.target.value) })}
              />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="edit-username">
                Username {formData.type === 'redis' && <span className="text-xs text-muted-foreground">(optional)</span>}
              </Label>
              <Input
                id="edit-username"
                required={formData.type !== 'redis'}
                value={formData.username || ''}
                onChange={(e) => setFormData({ ...formData, username: e.target.value })}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="edit-password">
                Password {formData.type === 'redis' && <span className="text-xs text-muted-foreground">(optional)</span>}
              </Label>
              <Input
                id="edit-password"
                type="password"
                required={formData.type !== 'redis'}
                value={formData.password || ''}
                onChange={(e) => setFormData({ ...formData, password: e.target.value })}
              />
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="edit-database">
              Database Name
              {formData.type === 'redis' && <span className="text-xs text-muted-foreground ml-1">(optional)</span>}
              {formData.type === 'mongodb' && <span className="text-xs text-muted-foreground ml-1">(optional)</span>}
            </Label>
            <Input
              id="edit-database"
              required={formData.type !== 'redis' && formData.type !== 'mongodb'}
              placeholder={formData.type === 'redis' ? 'Leave empty for default (0)' : formData.type === 'mongodb' ? 'Leave empty for admin' : ''}
              value={formData.database || ''}
              onChange={(e) => setFormData({ ...formData, database: e.target.value })}
            />
            {formData.type === 'redis' && (
              <p className="text-xs text-muted-foreground">
                Redis databases are numbered 0-15. Most users just use the default (0).
              </p>
            )}
          </div>

          <div className="flex items-center justify-between space-x-2 rounded-lg border p-3">
            <div className="flex items-center space-x-2">
              <Label htmlFor="edit-ssl" className="cursor-pointer">Enable SSL</Label>
              <TooltipProvider>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Info className="h-4 w-4 text-muted-foreground" />
                  </TooltipTrigger>
                  <TooltipContent>
                    <p>Secure Socket Layer for encrypted connections</p>
                  </TooltipContent>
                </Tooltip>
              </TooltipProvider>
            </div>
            <Switch
              id="edit-ssl"
              checked={formData.ssl}
              onCheckedChange={(checked) => setFormData({ ...formData, ssl: checked })}
            />
          </div>

          <div className="border rounded-lg">
            <button
              type="button"
              onClick={() => {
                setSSHExpanded(!sshExpanded);
                if (!sshExpanded) {
                  setFormData({ ...formData, ssh_enabled: true });
                }
              }}
              className="w-full flex items-center justify-between p-4 hover:bg-accent/50 transition-colors rounded-lg"
            >
              <div className="flex items-center gap-3">
                <div className="p-2 rounded-md bg-primary/10">
                  <KeyRound className="h-4 w-4 text-primary" />
                </div>
                <div className="text-left">
                  <div className="font-medium">SSH Tunnel (Optional)</div>
                  <div className="text-sm text-muted-foreground">
                    {formData.ssh_enabled 
                      ? `Connect via ${formData.ssh_host || 'SSH server'}` 
                      : "Connect to databases behind a firewall"}
                  </div>
                </div>
              </div>
              <div className="flex items-center gap-2">
                {formData.ssh_enabled && (
                  <div className="text-xs px-2 py-1 rounded-full bg-primary/10 text-primary">
                    Enabled
                  </div>
                )}
                {sshExpanded ? (
                  <ChevronUp className="h-4 w-4 text-muted-foreground" />
                ) : (
                  <ChevronDown className="h-4 w-4 text-muted-foreground" />
                )}
              </div>
            </button>

            {sshExpanded && (
              <div className="px-4 pb-4 space-y-4 border-t pt-4">
                <div className="flex items-center justify-between">
                  <div className="space-y-0.5">
                    <Label htmlFor="ssh-enabled">Enable SSH Tunnel</Label>
                    <div className="text-sm text-muted-foreground">
                      Use an SSH server as a jump host to reach your database
                    </div>
                  </div>
                  <Switch
                    id="ssh-enabled"
                    checked={formData.ssh_enabled}
                    onCheckedChange={(checked) => {
                      setFormData({ ...formData, ssh_enabled: checked });
                      if (!checked) {
                        setSSHExpanded(false);
                      }
                    }}
                  />
                </div>

                {formData.ssh_enabled && (
                  <>
                    <div className="grid grid-cols-2 gap-4">
                      <div className="space-y-2">
                        <Label htmlFor="ssh-host">SSH Host</Label>
                        <Input
                          id="ssh-host"
                          placeholder="bastion.example.com"
                          value={formData.ssh_host || ''}
                          onChange={(e) => setFormData({ ...formData, ssh_host: e.target.value })}
                        />
                      </div>
                      <div className="space-y-2">
                        <Label htmlFor="ssh-port">SSH Port</Label>
                        <Input
                          id="ssh-port"
                          type="number"
                          placeholder="22"
                          value={formData.ssh_port}
                          onChange={(e) => setFormData({ ...formData, ssh_port: parseInt(e.target.value) || 22 })}
                        />
                      </div>
                    </div>

                    <div className="space-y-2">
                      <Label htmlFor="ssh-username">SSH Username</Label>
                      <Input
                        id="ssh-username"
                        placeholder="ubuntu"
                        value={formData.ssh_username || ''}
                        onChange={(e) => setFormData({ ...formData, ssh_username: e.target.value })}
                      />
                    </div>

                    <div className="space-y-2">
                      <Label>Authentication Method</Label>
                      <div className="grid grid-cols-2 gap-2">
                        <Button
                          type="button"
                          variant={sshAuthMethod === "password" ? "default" : "outline"}
                          size="sm"
                          onClick={() => {
                            setSSHAuthMethod("password");
                            setFormData({ ...formData, ssh_private_key: "" });
                          }}
                        >
                          Password
                        </Button>
                        <Button
                          type="button"
                          variant={sshAuthMethod === "key" ? "default" : "outline"}
                          size="sm"
                          onClick={() => {
                            setSSHAuthMethod("key");
                            setFormData({ ...formData, ssh_password: "" });
                          }}
                        >
                          Private Key
                        </Button>
                      </div>
                    </div>

                    {sshAuthMethod === "password" ? (
                      <div className="space-y-2">
                        <Label htmlFor="ssh-password">SSH Password</Label>
                        <Input
                          id="ssh-password"
                          type="password"
                          placeholder="Your SSH password"
                          value={formData.ssh_password || ''}
                          onChange={(e) => setFormData({ ...formData, ssh_password: e.target.value })}
                        />
                      </div>
                    ) : (
                      <div className="space-y-2">
                        <Label htmlFor="ssh-key">SSH Private Key</Label>
                        <textarea
                          id="ssh-key"
                          placeholder="-----BEGIN OPENSSH PRIVATE KEY-----&#10;...&#10;-----END OPENSSH PRIVATE KEY-----"
                          value={formData.ssh_private_key || ''}
                          onChange={(e) => setFormData({ ...formData, ssh_private_key: e.target.value })}
                          rows={6}
                          className="w-full px-3 py-2 text-sm rounded-md border border-input bg-background font-mono"
                        />
                        <p className="text-xs text-muted-foreground">
                          Paste your private SSH key. It will be encrypted and stored securely.
                        </p>
                      </div>
                    )}

                    <div className="bg-muted/50 p-3 rounded-md text-sm">
                      <p className="font-medium mb-1">How it works:</p>
                      <p className="text-muted-foreground text-xs">
                        Velld will connect to <span className="font-mono">{formData.ssh_host || 'your SSH server'}</span>, 
                        then tunnel through to <span className="font-mono">{formData.host || 'your database'}:{formData.port || 'port'}</span>.
                        This allows you to reach databases behind firewalls or on private networks.
                      </p>
                    </div>
                  </>
                )}
              </div>
            )}
          </div>

          <div className="flex gap-2 pt-4">
            <Button type="submit" disabled={isEditing} className="flex-1">
              {isEditing ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Updating...
                </>
              ) : (
                'Update Connection'
              )}
            </Button>
            <Button type="button" variant="outline" onClick={onClose} disabled={isEditing}>
              Cancel
            </Button>
          </div>
        </form>
        )}
      </DialogContent>
    </Dialog>
  );
}
