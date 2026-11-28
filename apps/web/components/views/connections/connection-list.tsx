"use client";

import Link from "next/link";
import { useMemo, useState } from "react";
import { Search, Plus } from "lucide-react";
import { Button } from "@cloudflare/kumo/components/button";
import { Select } from "@cloudflare/kumo/components/select";
import { PageHeader } from "@/components/layout/page-header";
import { Input } from "@cloudflare/kumo/components/input";
import { Skeleton } from "@/components/ui/skeleton";
import { buttonVariants } from "@/components/ui/button";
import { useConnections } from "@/hooks/use-connections";
import { typeLabels } from "@/types/base";
import { ConnectionsTable } from "./connection-list/connections-table";

export function ConnectionsList() {
  const { connections, isLoading, error, refetch } = useConnections();
  const [search, setSearch] = useState("");
  const [engine, setEngine] = useState("all");
  const filtered = useMemo(
    () =>
      (connections ?? []).filter(
        (connection) =>
          (engine === "all" || connection.type === engine) &&
          [
            connection.name,
            connection.host,
            connection.database,
            connection.database_name,
            typeLabels[connection.type],
          ].some((value) =>
            value?.toLowerCase().includes(search.trim().toLowerCase()),
          ),
      ),
    [connections, engine, search],
  );

  return (
    <>
      <PageHeader
        title="Connections"
        action={
          !!connections?.length || error ? (
            <Link href="/connections/new" className={buttonVariants()}>
              <Plus aria-hidden="true" />
              Add connection
            </Link>
          ) : undefined
        }
      />
      {(isLoading || !!connections?.length) && (
        <div className="mb-4 flex flex-col gap-3 sm:flex-row">
          <div className="relative min-w-0 flex-1 sm:max-w-sm">
            <Search
              className="pointer-events-none absolute start-3 top-1/2 z-10 h-4 w-4 -translate-y-1/2 text-muted-foreground"
              aria-hidden="true"
            />
            <Input
              aria-label="Search connections"
              type="search"
              placeholder="Search connections…"
              value={search}
              onChange={(event) => setSearch(event.target.value)}
              className="block w-full ps-9"
            />
          </div>
          <Select
            aria-label="Filter by database engine"
            value={engine}
            onValueChange={(value) => {
              if (value) setEngine(value);
            }}
            items={{ all: "All engines", ...typeLabels }}
            className="w-full sm:w-40"
          >
            <Select.Option value="all">All engines</Select.Option>
            {Object.entries(typeLabels).map(([value, label]) => (
              <Select.Option key={value} value={value}>
                {label}
              </Select.Option>
            ))}
          </Select>
        </div>
      )}
      {error ? (
        <div role="alert" className="rounded-lg border p-6 text-sm">
          Couldn’t load connections.
          <div className="mt-4">
            <Button variant="secondary" onClick={() => refetch()}>
              Try again
            </Button>
          </div>
        </div>
      ) : isLoading ? (
        <div className="rounded-lg border p-5">
          {[1, 2, 3].map((key) => (
            <Skeleton key={key} className="mb-3 h-12 w-full last:mb-0" />
          ))}
        </div>
      ) : filtered.length ? (
        <ConnectionsTable connections={filtered} />
      ) : (
        <div className="rounded-lg border border-dashed px-5 py-14 text-center">
          <h2 className="text-sm font-semibold">
            {connections?.length ? "No matches" : "No connections yet"}
          </h2>
          {!connections?.length && (
            <p className="mt-2 text-[13px] text-muted-foreground">
              Add a PostgreSQL, MySQL, MongoDB, or Cloudflare D1 database.
            </p>
          )}
          <div className="mt-5">
            {connections?.length ? (
              <Button
                variant="secondary"
                onClick={() => {
                  setSearch("");
                  setEngine("all");
                }}
              >
                Clear filters
              </Button>
            ) : (
              <Link href="/connections/new" className={buttonVariants()}>
                Add connection
              </Link>
            )}
          </div>
        </div>
      )}
    </>
  );
}
