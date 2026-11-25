"use client";

import { useEffect, useState } from "react";
import { Workspace } from "@/components/layout/workspace";
import { ActivityList } from "@/components/views/dashboard/activity-list";

export default function Home() {
  const [today, setToday] = useState("");
  useEffect(() => {
    setToday(new Intl.DateTimeFormat("en", { dateStyle: "full" }).format(new Date()));
  }, []);

  return (
    <Workspace title="Overview" description={<time>{today || "\u00A0"}</time>}>
      <ActivityList />
    </Workspace>
  );
}