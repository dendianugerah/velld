import { AppShell } from "@/components/layout/app-shell";
import { ActivityList } from "@/components/views/dashboard/activity-list";

export default function Home() {
  return <AppShell><ActivityList /></AppShell>;
}