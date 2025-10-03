'use client';

import { Info, Link2, Loader2 } from "lucide-react";
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
  });
  
  const [connectionString, setConnectionString] = useState("");
  const [inputMethod, setInputMethod] = useState<"manual" | "url">("url");

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
      };
      
      const mappedType = typeMapping[type];
      
      if (!mappedType) {
        toast({
          variant: "destructive",
          title: "Unsupported Database Type",
          description: `The database type "${type}" is not supported. Supported types: PostgreSQL, MySQL, MongoDB`,
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
          <Label htmlFor="username">Username</Label>
          <Input
            id="username"
            required
            value={formData.username || ''}
            onChange={(e) => setFormData({ ...formData, username: e.target.value })}
          />
        </div>
        <div className="space-y-2">
          <Label htmlFor="password">Password</Label>
          <Input
            id="password"
            type="password"
            required
            value={formData.password || ''}
            onChange={(e) => setFormData({ ...formData, password: e.target.value })}
          />
        </div>
      </div>

      <div className="space-y-2">
        <Label htmlFor="database">Database Name</Label>
        <Input
          id="database"
          required
          value={formData.database || ''}
          onChange={(e) => setFormData({ ...formData, database: e.target.value })}
        />
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
          disabled={isAdding || !formData.type || !formData.host || !formData.database}
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
