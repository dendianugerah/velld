import { AppShell } from "@/components/layout/app-shell";
import { DockerDiscovery } from "@/components/views/connections/docker-discovery";

export default function DiscoverConnectionsPage() {
  return <AppShell><DockerDiscovery /></AppShell>;
}
