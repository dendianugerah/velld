import { Suspense } from "react";
import { AppShell } from "@/components/layout/app-shell";
import { ConnectionDetail } from "@/components/views/connections/connection-detail";

export default async function ConnectionPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  return <AppShell><Suspense fallback={<p className="text-sm text-muted-foreground">Loading connection…</p>}><ConnectionDetail id={id} /></Suspense></AppShell>;
}
