'use client';

import { Info, Link2, Loader2, ChevronDown, ChevronUp, KeyRound } from "lucide-react";
import { useState } from "react";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Switch } from "@/components/ui/switch";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Select, SelectContent, SelectItem, SelectSeparator, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import { useConnections } from "@/hooks/use-connections";
import { useToast } from "@/hooks/use-toast";
import { type ConnectionForm as ConnectionFormType } from "@/types/connection";
import { type DatabaseType } from "@/types/base";

interface ConnectionFormProps {
  onSuccess?: () => void;
  onCancel?: () => void;
}

export function ConnectionForm({ onSuccess, onCancel }: ConnectionFormProps) {
  const [formData, setFormData] = useState<ConnectionFormType>({
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
  
  const [connectionString, setConnectionString] = useState("");
  const [inputMethod, setInputMethod] = useState<"manual" | "url">("url");
  const [sshExpanded, setSSHExpanded] = useState(false);
  const [sshAuthMethod, setSSHAuthMethod] = useState<"password" | "key">("password");

  const { addConnection, isAdding } = useConnections();
  const { toast } = useToast();

  const parseConnectionString = (urlString: string) => {
    if (!urlString.trim()) {
      toast({
        variant: "destructive",
        title: "Invalid Connection String",
        description: "Connection string cannot be empty",
      });
      return false;
    }
    
    try {
      const url = new URL(urlString);
      const type = url.protocol.replace(':', '') as DatabaseType;
      const typeMapping: Record<string, DatabaseType> = {
        'postgres': 'postgresql',
        'postgresql': 'postgresql',
        'mysql': 'mysql',
        'mongodb': 'mongodb',
        'mongo': 'mongodb',
        'redis': 'redis',
        'rediss': 'redis',
      };
      
      const mappedType = typeMapping[type];
      
      if (!mappedType) {
        toast({
          variant: "destructive",
          title: "Unsupported Database Type",
          description: `The database type "${type}" is not supported. Supported types: PostgreSQL, MySQL, MongoDB, Redis`,
        });
        return false;
      }
      
      if (!url.hostname) {
        toast({
          variant: "destructive",
          title: "Invalid Connection String",
          description: "Hostname is required in the connection string",
        });
        return false;
      }
      
      const parsedData: Partial<ConnectionFormType> = {
        type: mappedType,
        host: url.hostname,
        port: parseInt(url.port) || getDefaultPort(mappedType),
        username: decodeURIComponent(url.username || ''),
        password: decodeURIComponent(url.password || ''),
        database: url.pathname.substring(1) || '',
        ssl: url.searchParams.get('ssl') !== 'false',
        name: formData.name || `${mappedType} - ${url.hostname}`,
      };
      
      setFormData(prev => ({ ...prev, ...parsedData }));
      return true;
    } catch (error) {
      toast({
        variant: "destructive",
        title: "Invalid Connection String Format",
        description: "Expected format: protocol://user:password@host:port/database",
      });
      console.error('Invalid connection string:', error);
      return false;
    }
  };
  
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
  
    if (inputMethod === 'url' && connectionString) {
      const parsed = parseConnectionString(connectionString);
      if (!parsed) {
        return;
      }
    }
    
    addConnection(formData, {
      onSuccess: () => {
        onSuccess?.();
      },
    });
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div className="space-y-2">
        <Label htmlFor="name">Connection Name</Label>
        <Input
          id="name"
          placeholder="My Database"
          value={formData.name || ''}
          onChange={(e) => setFormData({ ...formData, name: e.target.value })}
        />
        <p className="text-xs text-muted-foreground">
          {inputMethod === 'url' && !formData.name 
            ? "A name will be auto-generated from the connection string" 
            : "Give your connection a memorable name"}
        </p>
      </div>

      <Tabs value={inputMethod} onValueChange={(v) => setInputMethod(v as "manual" | "url")} className="w-full">
        <TabsList className="grid w-full grid-cols-2">
          <TabsTrigger value="url" className="gap-2">
            <Link2 className="h-4 w-4" />
            Connection String
          </TabsTrigger>
          <TabsTrigger value="manual">Manual Input</TabsTrigger>
        </TabsList>

        <TabsContent value="url" className="space-y-4 mt-4">
          <div className="space-y-2">
            <Label htmlFor="connection-string">Connection String</Label>
            <Input
              id="connection-string"
              placeholder="postgresql://user:password@localhost:5432/database"
              value={connectionString}
              onChange={(e) => {
                setConnectionString(e.target.value);
              }}
              onBlur={() => {
                if (connectionString) {
                  parseConnectionString(connectionString);
                }
              }}
            />
          </div>
          
          <div className="flex items-center justify-between space-x-4 border p-3 rounded-lg">
            <div className="space-y-0.5">
              <div className="flex items-center gap-2">
                <Label htmlFor="ssl-url">SSL Connection</Label>
                <TooltipProvider>
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <Info className="h-4 w-4 text-muted-foreground cursor-help" />
                    </TooltipTrigger>
                    <TooltipContent>
                      <p className="w-[240px]">
                        {formData.ssl 
                          ? "Using SSL encryption for secure connection to your database." 
                          : "Warning: Disabling SSL means your connection to the database will not be encrypted. Only use on trusted networks."}
                      </p>
                    </TooltipContent>
                  </Tooltip>
                </TooltipProvider>
              </div>
              <div className="text-sm text-muted-foreground">
                {formData.ssl 
                  ? "Connection will be encrypted (recommended)" 
                  : "Connection will not be encrypted"}
              </div>
            </div>
            <Switch
              id="ssl-url"
              checked={formData.ssl}
              onCheckedChange={(checked) => setFormData({ ...formData, ssl: checked })}
            />
          </div>

          <div className="flex space-x-2 pt-2">
            <Button 
              type="submit" 
              disabled={isAdding || !connectionString || !formData.type}
            >
              {isAdding ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Testing Connection...
                </>
              ) : (
                'Save Connection'
              )}
            </Button>
            <Button type="button" variant="outline" onClick={onCancel} disabled={isAdding}>
              Cancel
            </Button>
          </div>
        </TabsContent>

        <TabsContent value="manual" className="space-y-4 mt-4">
      <div className="space-y-2">
        <Label htmlFor="type">Database Type</Label>
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
          <Label htmlFor="host">Host</Label>
          <Input
            id="host"
            required
            value={formData.host || ''}
            onChange={(e) => setFormData({ ...formData, host: e.target.value })}
          />
        </div>
        <div className="space-y-2">
          <Label htmlFor="port">Port</Label>
          <Input
            id="port"
            type="number"
            required
            value={formData.port || ''}
            onChange={(e) => setFormData({ ...formData, port: parseInt(e.target.value) })}
          />
        </div>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label htmlFor="username">
            Username {formData.type === 'redis' && <span className="text-xs text-muted-foreground">(optional)</span>}
          </Label>
          <Input
            id="username"
            required={formData.type !== 'redis'}
            value={formData.username || ''}
            onChange={(e) => setFormData({ ...formData, username: e.target.value })}
          />
        </div>
        <div className="space-y-2">
          <Label htmlFor="password">
            Password {formData.type === 'redis' && <span className="text-xs text-muted-foreground">(optional)</span>}
          </Label>
          <Input
            id="password"
            type="password"
            required={formData.type !== 'redis'}
            value={formData.password || ''}
            onChange={(e) => setFormData({ ...formData, password: e.target.value })}
          />
        </div>
      </div>

      <div className="space-y-2">
        <Label htmlFor="database">
          Database Name <span className="text-xs text-muted-foreground ml-1">(optional - use discover to select databases)</span>
        </Label>
        <Input
          id="database"
          required={false}
          placeholder={
            formData.type === 'redis' ? 'Default: 0' : 
            formData.type === 'mongodb' ? 'Default: admin' : 
            formData.type === 'postgresql' ? 'Default: postgres' :
            formData.type === 'mysql' ? 'Leave empty to discover databases' :
            'Leave empty to discover databases'
          }
          value={formData.database || ''}
          onChange={(e) => setFormData({ ...formData, database: e.target.value })}
        />
      </div>

      {/* SSH Tunnel Configuration */}
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
      
      <div className="flex items-center justify-between space-x-4 border p-3 rounded-lg">
        <div className="space-y-0.5">
          <div className="flex items-center gap-2">
            <Label htmlFor="ssl">SSL Connection</Label>
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger asChild>
                  <Info className="h-4 w-4 text-muted-foreground cursor-help" />
                </TooltipTrigger>
                <TooltipContent>
                  <p className="w-[240px]">
                    {formData.ssl 
                      ? "Using SSL encryption for secure connection to your database." 
                      : "Warning: Disabling SSL means your connection to the database will not be encrypted. Only use on trusted networks."}
                  </p>
                </TooltipContent>
              </Tooltip>
            </TooltipProvider>
          </div>
          <div className="text-sm text-muted-foreground">
            {formData.ssl 
              ? "Connection will be encrypted (recommended)" 
              : "Connection will not be encrypted (useful for local development)"}
          </div>
        </div>
        <Switch
          id="ssl"
          checked={formData.ssl}
          onCheckedChange={(checked) => setFormData({ ...formData, ssl: checked })}
        />
      </div>

      <div className="flex space-x-2 pt-2">
        <Button 
          type="submit" 
          disabled={isAdding || !formData.type || !formData.host}
        >
          {isAdding ? (
            <>
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              Testing Connection...
            </>
          ) : (
            'Save Connection'
          )}
        </Button>
        <Button type="button" variant="outline" onClick={onCancel} disabled={isAdding}>
          Cancel
        </Button>
      </div>
        </TabsContent>
      </Tabs>
    </form>
  );
}
