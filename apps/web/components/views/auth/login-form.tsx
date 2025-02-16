"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useLogin } from "@/hooks/use-auth";
import Link from "next/link";
import { checkIsAllowRegister } from "@/lib/helper";

export function LoginForm() {
  const [formData, setFormData] = useState({
    username: "",
    password: "",
  });

  const isSignupActive = checkIsAllowRegister();
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
            placeholder="Enter your username"
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
            placeholder="Enter your password"
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

      {isSignupActive && (
        <div className="text-center text-sm">
          <span className="text-muted-foreground">Dont have an account? </span>
          <Link
            href="/register"
            className="font-medium hover:text-primary underline-offset-4 hover:underline"
          >
            Sign up
          </Link>
        </div>
      )}
    </form>
  );
}
