"use client";

import { useState } from "react";
import { Button } from "@cloudflare/kumo/components/button";
import { LayerCard } from "@cloudflare/kumo/components/layer-card";
import { Bell, Cloud, LogOut, UserRound } from "lucide-react";
import { useProfile } from "@/hooks/use-profile";
import { logout } from "@/lib/helper";
import { cn } from "@/lib/utils";
import { navigationIconClass, navigationItemClass } from "@/components/layout/navigation-styles";
import { Skeleton } from "@/components/ui/skeleton";
import { NotificationSettings } from "./notification-settings";
import { S3StorageSettings } from "./s3-storage-settings";

const sections = [
  { id: "account", label: "Account", icon: UserRound },
  { id: "storage", label: "Storage", icon: Cloud },
  { id: "notifications", label: "Notifications", icon: Bell },
] as const;

type Section = (typeof sections)[number]["id"];

export function ProfileSettings() {
  const { data: profile, isLoading } = useProfile();
  const [section, setSection] = useState<Section>("account");

  return <div className="mx-auto max-w-[1000px]">
    <div className="mb-8 border-b border-border pb-6">
      <h1 className="text-2xl font-semibold tracking-[-0.035em]">Settings</h1>
    </div>

    <div className="flex flex-col gap-8 md:flex-row md:gap-12">
      <nav aria-label="Settings sections" className="flex shrink-0 flex-wrap gap-1 border-b border-border pb-3 md:w-44 md:flex-col md:border-0 md:pb-0 lg:sticky lg:top-6 lg:max-h-[calc(100dvh-3rem)] lg:self-start lg:overflow-y-auto">
        {sections.map(({ id, label, icon: Icon }) => <button
          key={id}
          type="button"
          onClick={() => setSection(id)}
          aria-pressed={section === id}
          className={cn(navigationItemClass(section === id), "shrink-0")}
        >
          <Icon className={navigationIconClass} strokeWidth={1.7} aria-hidden="true" />
          {label}
        </button>)}
      </nav>

      <div className="min-w-0 flex-1 pb-8">
        <section hidden={section !== "account"} aria-labelledby="account-heading">
          <h2 id="account-heading" className="text-xl font-semibold tracking-[-0.025em]">Account</h2>

          <LayerCard className="mt-7 rounded-lg shadow-none">
            <div className="flex items-center gap-4 p-5">
              <div className="grid h-10 w-10 shrink-0 place-items-center rounded-full bg-muted text-muted-foreground" aria-hidden="true"><UserRound className="h-5 w-5" strokeWidth={1.7} /></div>
              <div className="min-w-0"><p className="text-xs text-muted-foreground">Username</p>{isLoading ? <Skeleton className="mt-2 h-4 w-32" /> : <p className="mt-1 break-words text-sm font-medium">{profile?.data?.username || "Unavailable"}</p>}</div>
            </div>
          </LayerCard>

          <div className="mt-8 flex flex-wrap items-center justify-between gap-4 border-t border-border pt-6">
            <Button variant="secondary" onClick={logout}><LogOut className="h-4 w-4" aria-hidden="true" />Sign out</Button>
          </div>
        </section>

        {/* Keep forms mounted so switching sections doesn't discard unsaved changes. */}
        <div hidden={section !== "storage"}><S3StorageSettings /></div>
        <div hidden={section !== "notifications"}><NotificationSettings /></div>
      </div>
    </div>
  </div>;
}
