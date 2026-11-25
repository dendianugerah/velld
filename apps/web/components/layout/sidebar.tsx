"use client";

import { useState } from "react";
import { Database, History, LayoutGrid, LogOut, Menu, Settings } from "lucide-react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Sheet, SheetContent, SheetTitle, SheetTrigger } from "@/components/ui/sheet";
import { useProfile } from "@/hooks/use-profile";
import { logout } from "@/lib/helper";
import { cn } from "@/lib/utils";

const sections = [
  { label: "Workspace", items: [
    { label: "Overview", href: "/", icon: LayoutGrid },
    { label: "Connections", href: "/connections", icon: Database },
    { label: "Backup History", href: "/history", icon: History },
  ] },
  { label: "Account", items: [
    { label: "Settings", href: "/settings", icon: Settings },
  ] },
];

export function Sidebar() {
  const pathname = usePathname();
  const { data: profile } = useProfile();
  const [open, setOpen] = useState(false);
  const username = profile?.data.username;

  const navigation = (
    <div className="flex h-full flex-col">
      <Link href="/" className="mx-4 mb-10 w-fit text-2xl font-bold tracking-[-0.07em]" onClick={() => setOpen(false)}>
        velld<span className="text-[#3e5543]">.</span>
      </Link>
      <nav aria-label="Workspace navigation" className="flex-1 space-y-7">
        {sections.map(({ label, items }) => (
          <div key={label}>
            <p className="mb-2 px-4 text-[11px] font-semibold uppercase tracking-[0.16em] text-muted-foreground">{label}</p>
            <ul className="space-y-1">
              {items.map(({ label, href, icon: Icon }) => (
                <li key={href}>
                  <Link
                    href={href}
                    aria-current={pathname === href ? "page" : undefined}
                    className={cn(
                      "flex items-center gap-3 rounded-lg border border-transparent px-3 py-2.5 text-sm font-medium text-muted-foreground hover:bg-accent hover:text-foreground",
                      pathname === href && "border-border bg-card text-foreground shadow-sm hover:bg-card"
                    )}
                    onClick={() => setOpen(false)}
                  >
                    <Icon aria-hidden="true" className="size-4" />
                    {label}
                  </Link>
                </li>
              ))}
            </ul>
          </div>
        ))}
      </nav>
      <div className="mt-8 flex items-center gap-3 border-t px-2 pt-4">
        <span aria-hidden="true" className="grid size-9 shrink-0 place-items-center rounded-full bg-[#e5ebe2] text-sm font-semibold text-[#354d3b]">
          {username?.slice(0, 1).toUpperCase() || "V"}
        </span>
        <span className="min-w-0 flex-1 truncate text-sm font-medium">{username || "Velld account"}</span>
        <Button variant="ghost" size="icon" onClick={logout} aria-label="Sign out"><LogOut /></Button>
      </div>
    </div>
  );

  return (
    <>
      <aside className="sticky top-0 hidden h-dvh w-56 shrink-0 border-r px-3 pb-5 pt-7 lg:block">
        {navigation}
      </aside>
      <div className="flex items-center justify-between border-b px-5 py-3 lg:hidden">
        <Link href="/" className="text-xl font-bold tracking-[-0.07em]">velld.</Link>
        <Sheet open={open} onOpenChange={setOpen}>
          <SheetTrigger asChild>
            <Button variant="ghost" size="icon" aria-label="Open navigation"><Menu /></Button>
          </SheetTrigger>
          <SheetContent side="left" className="w-64 px-3 pt-7" aria-describedby={undefined}>
            <SheetTitle className="sr-only">Velld navigation</SheetTitle>
            {navigation}
          </SheetContent>
        </Sheet>
      </div>
    </>
  );
}