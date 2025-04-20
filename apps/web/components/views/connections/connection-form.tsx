'use client';

import { Info } from "lucide-react";
import { useState } from "react";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Switch } from "@/components/ui/switch";
import { Select, SelectContent, SelectItem, SelectSeparator, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import { useConnections } from "@/hooks/use-connections";
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

  const { addConnection, isAdding } = useConnections();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    addConnection(formData, {
      onSuccess: () => onSuccess?.(),
    });
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div className="space-y-2">
        <Label htmlFor="name">Connection Name</Label>
        <Input
          id="name"
          required
          value={formData.name || ''}
          onChange={(e) => setFormData({ ...formData, name: e.target.value })}
        />
      </div>

      <div className="space-y-2">
        <Label htmlFor="type">Database Type</Label>
        <Select
          value={formData.type}
          onValueChange={(value) => setFormData({ ...formData, type: value as DatabaseType })}
        >
          <SelectTrigger className="w-[180px]">
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

      <div className="flex space-x-2">
        <Button type="submit" disabled={isAdding}>
          {isAdding ? 'Testing Connection...' : 'Save Connection'}
        </Button>
        <Button type="button" variant="outline" onClick={onCancel}>
          Cancel
        </Button>
      </div>
    </form>
  );
}
