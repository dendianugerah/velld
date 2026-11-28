"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useLogin } from "@/hooks/use-auth";

export function LoginForm() {
  const [formData, setFormData] = useState({
    username: "",
    password: "",
  });

  const { mutate: login, isPending } = useLogin();

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setFormData({
      ...formData,
      [e.target.id]: e.target.value,
    });
  };

  async function onSubmit(event: React.FormEvent) {
    event.preventDefault();
    login(formData);
  }

  return (
    <form onSubmit={onSubmit} className="space-y-6">
      <div className="space-y-4">
        <div className="space-y-2">
          <Label htmlFor="username">Username</Label>
          <Input
            id="username"
            type="text"
            disabled={isPending}
            required
            value={formData.username}
            onChange={handleChange}
          />
        </div>
        <div className="space-y-2">
          <Label htmlFor="password">Password</Label>
          <Input
            id="password"
            type="password"
            disabled={isPending}
            required
            value={formData.password}
            onChange={handleChange}
          />
        </div>
      </div>

      <Button className="w-full" type="submit" disabled={isPending}>
        {isPending ? "Signing in..." : "Sign in"}
      </Button>

    </form>
  );
}
