"use client";

import { useState } from "react";
import { usePathname } from "next/navigation";
import Link from "next/link";
import { BookOpen, Database, Github, History, LayoutDashboard, Menu, Settings } from "lucide-react";
import { navigationIconClass, navigationItemClass } from "@/components/layout/navigation-styles";
import { Button } from "@/components/ui/button";
import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle, SheetTrigger } from "@/components/ui/sheet";
import Logo from "@/components/ui/logo";

const navigation = [
  { name: "Overview", href: "/", icon: LayoutDashboard },
  { name: "Connections", href: "/connections", icon: Database },
  { name: "Backup history", href: "/history", icon: History },
  { name: "Settings", href: "/settings", icon: Settings },
];

export function Sidebar() {
  const pathname = usePathname();
  const [open, setOpen] = useState(false);

  const contents = <>
    <Link href="/" aria-label="Velld overview" onClick={() => setOpen(false)} className="mb-6 flex h-9 w-fit items-center px-2.5">
      <Logo />
    </Link>
    <p className="mb-2 px-2.5 text-[11px] font-medium text-muted-foreground">Workspace</p>
    <nav aria-label="Main navigation" className="space-y-0.5">
      {navigation.map(item => {
        const active = item.href === "/" ? pathname === "/" : pathname === item.href || pathname.startsWith(`${item.href}/`);
        return <Link key={item.href} href={item.href} aria-current={active ? "page" : undefined} onClick={() => setOpen(false)} className={navigationItemClass(active)}><item.icon className={navigationIconClass} strokeWidth={1.7} aria-hidden="true" />{item.name}</Link>;
      })}
    </nav>
    <div className="mt-8 space-y-0.5">
      <p className="mb-2 px-2.5 text-[11px] font-medium text-muted-foreground">Resources</p>
      <a href="https://velld.vercel.app/docs" target="_blank" rel="noreferrer" className={navigationItemClass(false)}><BookOpen className={navigationIconClass} strokeWidth={1.7} aria-hidden="true" />Documentation<span className="sr-only"> (opens in a new tab)</span></a>
      <a href="https://github.com/dendianugerah/velld" target="_blank" rel="noreferrer" className={navigationItemClass(false)}><Github className={navigationIconClass} strokeWidth={1.7} aria-hidden="true" />Source code<span className="sr-only"> (opens in a new tab)</span></a>
    </div>
  </>;

  return <>
    <aside className="fixed inset-y-0 start-0 z-30 hidden w-[224px] flex-col overflow-y-auto px-3 pb-5 pt-5 lg:flex">{contents}</aside>
    <header className="sticky top-0 z-30 flex h-14 items-center gap-3 border-b border-border bg-sidebar px-4 lg:hidden">
      <Sheet open={open} onOpenChange={setOpen}>
        <SheetTrigger asChild><Button variant="ghost" size="icon" className="-ms-2 h-9 w-9" aria-label="Open navigation"><Menu className="h-[18px] w-[18px]" strokeWidth={1.7} aria-hidden="true" /></Button></SheetTrigger>
        <SheetContent side="left" className="flex w-[264px] flex-col gap-0 overflow-y-auto bg-sidebar px-3 pb-5 pt-5">
          <SheetHeader className="sr-only"><SheetTitle>Workspace navigation</SheetTitle><SheetDescription>Navigate your Velld workspace and account.</SheetDescription></SheetHeader>
          {contents}
        </SheetContent>
      </Sheet>
      <Link href="/" aria-label="Velld overview"><Logo /></Link>
    </header>
  </>;
}
