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
import { useConnections } from "@/hooks/use-connections";
import { useToast } from "@/hooks/use-toast";
import { type ConnectionForm as ConnectionFormType } from "@/types/connection";
import { type DatabaseType } from "@/types/base";
import type { Connection } from "@/types/connection";

interface EditConnectionDialogProps {
  connection: Connection | null;
  onClose: () => void;
}

export function EditConnectionDialog({
  connection,
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

  useEffect(() => {
    if (connection) {
      setFormData({
        id: connection.id,
        name: connection.name,
        type: connection.type,
        host: connection.host,
        port: connection.port,
        username: connection.username,
        password: connection.password,
        database: connection.database,
        ssl: connection.ssl,
        ssh_enabled: connection.ssh_enabled,
        ssh_host: connection.ssh_host || "",
        ssh_port: connection.ssh_port || 0,
        ssh_username: connection.ssh_username || "",
        ssh_password: connection.ssh_password || "",
        ssh_private_key: connection.ssh_private_key || "",
      });
      setSSHExpanded(connection.ssh_enabled);
      setSSHAuthMethod(connection.ssh_private_key ? "key" : "password");
    }
  }, [connection]);

  const getDefaultPort = (type: string): number => {
    const ports: Record<string, number> = {
      'postgresql': 5432,
      'mysql': 3306,
      'mongodb': 27017,
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

  return (
    <Dialog open={!!connection} onOpenChange={() => onClose()}>
      <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Edit Connection</DialogTitle>
          <DialogDescription>
            Update your database connection settings
          </DialogDescription>
        </DialogHeader>

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
              <Label htmlFor="edit-username">Username</Label>
              <Input
                id="edit-username"
                required
                value={formData.username || ''}
                onChange={(e) => setFormData({ ...formData, username: e.target.value })}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="edit-password">Password</Label>
              <Input
                id="edit-password"
                type="password"
                required
                value={formData.password || ''}
                onChange={(e) => setFormData({ ...formData, password: e.target.value })}
              />
            </div>
          </div>

          <div className="space-y-2">
            <Label htmlFor="edit-database">Database Name</Label>
            <Input
              id="edit-database"
              required
              value={formData.database || ''}
              onChange={(e) => setFormData({ ...formData, database: e.target.value })}
            />
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

          <div className="space-y-2">
            <button
              type="button"
              onClick={() => setSSHExpanded(!sshExpanded)}
              className="flex items-center justify-between w-full p-3 rounded-lg border hover:bg-accent transition-colors"
            >
              <div className="flex items-center space-x-2">
                <span className="font-medium">SSH Tunnel (Optional)</span>
                <TooltipProvider>
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <Info className="h-4 w-4 text-muted-foreground" />
                    </TooltipTrigger>
                    <TooltipContent>
                      <p>Connect through an SSH tunnel for additional security</p>
                    </TooltipContent>
                  </Tooltip>
                </TooltipProvider>
              </div>
              {sshExpanded ? <ChevronUp className="h-4 w-4" /> : <ChevronDown className="h-4 w-4" />}
            </button>

            {sshExpanded && (
              <div className="space-y-4 p-4 rounded-lg border bg-muted/50">
                <div className="flex items-center justify-between">
                  <Label htmlFor="edit-ssh-enabled">Enable SSH Tunnel</Label>
                  <Switch
                    id="edit-ssh-enabled"
                    checked={formData.ssh_enabled}
                    onCheckedChange={(checked) => setFormData({ ...formData, ssh_enabled: checked })}
                  />
                </div>

                {formData.ssh_enabled && (
                  <>
                    <div className="grid grid-cols-2 gap-4">
                      <div className="space-y-2">
                        <Label htmlFor="edit-ssh-host">SSH Host</Label>
                        <Input
                          id="edit-ssh-host"
                          value={formData.ssh_host || ''}
                          onChange={(e) => setFormData({ ...formData, ssh_host: e.target.value })}
                        />
                      </div>
                      <div className="space-y-2">
                        <Label htmlFor="edit-ssh-port">SSH Port</Label>
                        <Input
                          id="edit-ssh-port"
                          type="number"
                          value={formData.ssh_port || 22}
                          onChange={(e) => setFormData({ ...formData, ssh_port: parseInt(e.target.value) })}
                        />
                      </div>
                    </div>

                    <div className="space-y-2">
                      <Label htmlFor="edit-ssh-username">SSH Username</Label>
                      <Input
                        id="edit-ssh-username"
                        value={formData.ssh_username || ''}
                        onChange={(e) => setFormData({ ...formData, ssh_username: e.target.value })}
                      />
                    </div>

                    <div className="flex items-center space-x-2 rounded-lg border p-2">
                      <Button
                        type="button"
                        variant={sshAuthMethod === "password" ? "default" : "ghost"}
                        size="sm"
                        onClick={() => setSSHAuthMethod("password")}
                        className="flex-1"
                      >
                        <KeyRound className="h-3 w-3 mr-1" />
                        Password
                      </Button>
                      <Button
                        type="button"
                        variant={sshAuthMethod === "key" ? "default" : "ghost"}
                        size="sm"
                        onClick={() => setSSHAuthMethod("key")}
                        className="flex-1"
                      >
                        <KeyRound className="h-3 w-3 mr-1" />
                        Private Key
                      </Button>
                    </div>

                    {sshAuthMethod === "password" ? (
                      <div className="space-y-2">
                        <Label htmlFor="edit-ssh-password">SSH Password</Label>
                        <Input
                          id="edit-ssh-password"
                          type="password"
                          value={formData.ssh_password || ''}
                          onChange={(e) => setFormData({ ...formData, ssh_password: e.target.value, ssh_private_key: "" })}
                        />
                      </div>
                    ) : (
                      <div className="space-y-2">
                        <Label htmlFor="edit-ssh-private-key">SSH Private Key</Label>
                        <textarea
                          id="edit-ssh-private-key"
                          className="w-full min-h-[100px] rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2"
                          placeholder="-----BEGIN OPENSSH PRIVATE KEY-----"
                          value={formData.ssh_private_key || ''}
                          onChange={(e) => setFormData({ ...formData, ssh_private_key: e.target.value, ssh_password: "" })}
                        />
                      </div>
                    )}
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
      </DialogContent>
    </Dialog>
  );
}
