"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { ArrowLeft } from "lucide-react";
import { Button } from "@cloudflare/kumo/components/button";
import { Tabs } from "@cloudflare/kumo/components/tabs";
import { PageHeader } from "@/components/layout/page-header";
import { Skeleton } from "@/components/ui/skeleton";
import { useConnection } from "@/hooks/use-connections";
import { ConnectionForm } from "./connection-form";
import { DockerDiscovery } from "./docker-discovery";

export function ConnectionEditor({ id }: { id?: string }) {
  const router = useRouter();
  const [method, setMethod] = useState<"manual" | "url" | "docker">("manual");
  const [visitedDocker, setVisitedDocker] = useState(false);
  const { data, isLoading, error, refetch } = useConnection(id ?? null);
  const back = id ? `/connections/${id}` : "/connections";
  return (
    <>
      <Link
        href={back}
        className="mb-7 inline-flex min-h-8 items-center gap-2 rounded-sm text-xs text-muted-foreground hover:text-foreground focus-visible:outline focus-visible:outline-2 focus-visible:outline-ring"
      >
        <ArrowLeft className="h-3.5 w-3.5" aria-hidden="true" />
        {id ? "Back to connection" : "All connections"}
      </Link>
      <div className="mx-auto max-w-[720px]">
        <PageHeader title={id ? "Edit connection" : "Add connection"} />
        {id && isLoading ? (
          <div className="max-w-[720px] space-y-6">
            <Skeleton className="h-16 w-full" />
            <Skeleton className="h-48 w-full" />
          </div>
        ) : id && (error || !data) ? (
          <div role="alert" className="space-y-4 text-sm">
            <p>Couldn’t load this connection.</p>
            <Button variant="secondary" onClick={() => refetch()}>
              Try again
            </Button>
          </div>
        ) : id ? (
          <ConnectionForm
            key={id}
            connection={data}
            onSuccess={(connectionId) =>
              router.push(`/connections/${connectionId}`)
            }
            onCancel={() => router.push(back)}
          />
        ) : (
          <div className="max-w-[720px] pb-8">
            <div className="mb-6" role="group" aria-label="Connection method">
              <Tabs
                variant="underline"
                listClassName="!h-auto !pb-0"
                value={method}
                tabs={[
                  {
                    value: "manual",
                    label: "Details",
                    className: "min-h-10 !py-2.5",
                  },
                  {
                    value: "url",
                    label: "Connection URL",
                    className: "min-h-10 !py-2.5",
                  },
                  {
                    value: "docker",
                    label: "Docker",
                    className: "min-h-10 !py-2.5",
                  },
                ]}
                onValueChange={(value) => {
                  if (
                    value === "manual" ||
                    value === "url" ||
                    value === "docker"
                  ) {
                    if (value === "docker") setVisitedDocker(true);
                    setMethod(value);
                  }
                }}
              />
            </div>
            <div className="min-w-0">
              <div hidden={method === "docker"}>
                <ConnectionForm
                  inputMethod={method === "url" ? "url" : "manual"}
                  onSuccess={(connectionId) =>
                    router.push(`/connections/${connectionId}`)
                  }
                  onCancel={() => router.push(back)}
                />
              </div>
              {visitedDocker && (
                <div hidden={method !== "docker"}>
                  <DockerDiscovery onManual={() => setMethod("manual")} />
                </div>
              )}
            </div>
          </div>
        )}
      </div>
    </>
  );
}
