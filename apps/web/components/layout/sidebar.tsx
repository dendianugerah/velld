"use client";

import { useState } from "react";
import { usePathname } from "next/navigation";
import Link from "next/link";
import { BookOpen, ChevronUp, Database, Github, History, LayoutDashboard, LogOut, Menu, Settings } from "lucide-react";
import { cn } from "@/lib/utils";
import { logout } from "@/lib/helper";
import { useProfile } from "@/hooks/use-profile";
import { Button } from "@/components/ui/button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuLabel, DropdownMenuSeparator, DropdownMenuTrigger } from "@/components/ui/dropdown-menu";
import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle, SheetTrigger } from "@/components/ui/sheet";
import Logo from "@/components/ui/logo";

const navigation = [
  { name: "Overview", href: "/", icon: LayoutDashboard },
  { name: "Connections", href: "/connections", icon: Database },
  { name: "Backup history", href: "/history", icon: History },
];

export function Sidebar() {
  const pathname = usePathname();
  const { data: profile } = useProfile();
  const username = profile?.data.username || "Your account";
  const [open, setOpen] = useState(false);

  const itemClass = (active: boolean) => cn("flex items-center gap-2.5 rounded-md px-2.5 py-2 text-[13px] font-medium focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary",
    active ? "bg-[#e8edf5] text-[#354c77]" : "text-[#60636b] hover:bg-black/[0.035] hover:text-foreground");

  const contents = <>
    <Link href="/" aria-label="Velld overview" onClick={() => setOpen(false)} className="mb-9 flex w-fit items-center gap-2.5 px-2 pt-1">
      <span aria-hidden="true" className="flex h-7 w-7 -rotate-[4deg] items-center justify-center rounded-lg border border-[#ced9ed] bg-[#e4ecfa] font-semibold text-[#354c77] shadow-[0_2px_0_#354c7710]">v</span>
      <Logo />
    </Link>
    <p className="mb-2 px-2.5 text-[10px] font-medium uppercase tracking-[0.08em] text-muted-foreground">Workspace</p>
    <nav aria-label="Main navigation" className="space-y-1">
      {navigation.map(item => <Link key={item.href} href={item.href} aria-current={pathname === item.href ? "page" : undefined} onClick={() => setOpen(false)} className={itemClass(pathname === item.href)}><item.icon className="h-4 w-4 shrink-0 stroke-[1.6]" />{item.name}</Link>)}
    </nav>
    <div className="mt-7">
      <Link href="/settings" aria-current={pathname === "/settings" ? "page" : undefined} onClick={() => setOpen(false)} className={itemClass(pathname === "/settings")}><Settings className="h-4 w-4 stroke-[1.6]" />Settings</Link>
    </div>
    <div className="mt-auto space-y-1 pt-10">
      <a href="https://velld.vercel.app/docs" target="_blank" rel="noreferrer" className={itemClass(false)}><BookOpen className="h-4 w-4 stroke-[1.6]" />Documentation<span className="ms-auto text-muted-foreground">↗</span></a>
      <a href="https://github.com/dendianugerah/velld" target="_blank" rel="noreferrer" className={itemClass(false)}><Github className="h-4 w-4 stroke-[1.6]" />Source code<span className="ms-auto text-muted-foreground">↗</span></a>
      <div className="!mt-4 border-t pt-3">
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <button className="flex w-full items-center gap-2.5 rounded-lg px-2 py-2 text-start hover:bg-black/[0.035] focus-visible:outline focus-visible:outline-2 focus-visible:outline-primary" aria-label={`Account menu for ${username}`}>
              <span className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full border border-[#d7dce5] bg-white text-xs font-medium uppercase">{username[0]}</span>
              <span className="min-w-0 flex-1"><span className="block break-words text-xs font-medium">{username}</span><span className="mt-0.5 block text-[10px] text-muted-foreground">Self-hosted instance</span></span>
              <ChevronUp className="h-3.5 w-3.5 shrink-0 text-muted-foreground" />
            </button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="start" side="top" className="min-w-52">
            <DropdownMenuLabel className="max-w-64 break-words text-xs font-normal text-muted-foreground">Signed in as <span className="font-medium text-foreground">{username}</span></DropdownMenuLabel>
            <DropdownMenuSeparator />
            <DropdownMenuItem asChild><Link href="/settings" onClick={() => setOpen(false)}>Account & settings</Link></DropdownMenuItem>
            <DropdownMenuItem onClick={logout}><LogOut className="me-2 h-3.5 w-3.5" />Sign out</DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </div>
  </>;

  return <>
    <aside className="fixed inset-y-0 start-0 z-30 hidden w-[224px] flex-col px-3 pb-3 pt-5 lg:flex">{contents}</aside>
    <header className="sticky top-0 z-30 flex h-14 items-center gap-3 border-b bg-[#f7f7f8] px-4 lg:hidden">
      <Sheet open={open} onOpenChange={setOpen}>
        <SheetTrigger asChild><Button variant="ghost" size="icon" className="-ms-2 h-9 w-9" aria-label="Open navigation"><Menu className="h-4 w-4" /></Button></SheetTrigger>
        <SheetContent side="left" className="flex w-[264px] flex-col bg-[#f7f7f8] px-3 pb-3 pt-5">
          <SheetHeader className="sr-only"><SheetTitle>Workspace navigation</SheetTitle><SheetDescription>Navigate your Velld workspace and account.</SheetDescription></SheetHeader>
          {contents}
        </SheetContent>
      </Sheet>
      <Link href="/" aria-label="Velld overview"><Logo /></Link>
      <span className="ms-auto text-[11px] text-muted-foreground">Self-hosted</span>
    </header>
  </>;
}