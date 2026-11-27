"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { ArrowLeft } from "lucide-react";
import { Button } from "@cloudflare/kumo/components/button";
import { PageHeader } from "@/components/layout/page-header";
import { Skeleton } from "@/components/ui/skeleton";
import { useConnection } from "@/hooks/use-connections";
import { ConnectionForm } from "./connection-form";

export function ConnectionEditor({ id }: { id?: string }) {
  const router = useRouter();
  const { data, isLoading, error, refetch } = useConnection(id ?? null);
  const back = id ? `/connections/${id}` : "/connections";
  return <>
    <Link href={back} className="mb-7 inline-flex min-h-8 items-center gap-2 rounded-sm text-xs text-muted-foreground hover:text-foreground focus-visible:outline focus-visible:outline-2 focus-visible:outline-ring"><ArrowLeft className="h-3.5 w-3.5" aria-hidden="true" />{id ? "Back to connection" : "All connections"}</Link>
    <PageHeader title={id ? "Edit connection" : "Add connection"} description={id ? "Update how Velld connects to your database." : "Connect a database to start backing it up."} />
    {id && isLoading ? <div className="max-w-[720px] space-y-6"><Skeleton className="h-16 w-full" /><Skeleton className="h-48 w-full" /></div> : id && (error || !data) ? <div role="alert" className="space-y-4 text-sm"><p>Couldn’t load this connection. Try again or return to Connections.</p><Button variant="secondary" onClick={() => refetch()}>Try again</Button></div> : <ConnectionForm key={id ?? "new"} connection={data} onSuccess={connectionId => router.push(`/connections/${connectionId}`)} onCancel={() => router.push(back)} />}
  </>;
}
