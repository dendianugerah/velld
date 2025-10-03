"use client";

import { Layout, Server, Settings, History, Menu, X } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { usePathname } from "next/navigation";
import Link from "next/link";
import Logo from "../ui/logo";
import { useState } from "react";
import { ThemeToggle } from "@/components/theme-toggle";

const navigation = [
  { name: "Dashboard", href: "/", icon: Layout },
  { name: "Connections", href: "/connections", icon: Server },
  { name: "History", href: "/history", icon: History },
  // { name: "Security", href: "/security", icon: Shield },
  { name: "Settings", href: "/settings", icon: Settings },
];

export function Sidebar() {
  const pathname = usePathname();
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);

  const sidebarContent = (
    <div className="flex flex-col h-full">
      <div className="flex justify-between items-start mb-8">
        <Logo />
        <div className="flex items-center gap-2">
          <ThemeToggle />
          <Button
            variant="ghost"
            size="icon"
            className="lg:hidden"
            onClick={() => setIsMobileMenuOpen(false)}
          >
            <X className="h-6 w-6" />
          </Button>
        </div>
      </div>
      <nav className="space-y-1 flex-1">
        {navigation.map((item) => {
          const Icon = item.icon;
          return (
            <Link key={item.name} href={item.href} onClick={() => setIsMobileMenuOpen(false)}>
              <Button
                variant={pathname === item.href ? "secondary" : "ghost"}
                className={cn(
                  "w-full justify-start my-0.5",
                  pathname === item.href && "bg-secondary"
                )}
              >
                <Icon className="mr-2 h-4 w-4" />
                {item.name}
              </Button>
            </Link>
          );
        })}
      </nav>
    </div>
  );

  return (
    <>
      <Button
        variant="ghost"
        size="icon"
        className="fixed top-4 left-4 lg:hidden z-50"
        onClick={() => setIsMobileMenuOpen(true)}
      >
        <Menu className="h-6 w-6" />
      </Button>

      <div
        className={cn(
          "fixed inset-0 bg-background/80 backdrop-blur-sm z-40 lg:hidden",
          isMobileMenuOpen ? "block" : "hidden"
        )}
        onClick={() => setIsMobileMenuOpen(false)}
      />

      <div
        className={cn(
          "fixed inset-y-0 left-0 w-64 border-r bg-card backdrop-blur-xl p-6 z-50 transition-transform duration-300",
          isMobileMenuOpen ? "translate-x-0" : "-translate-x-full lg:translate-x-0"
        )}
      >
        {sidebarContent}
      </div>
    </>
  );
}