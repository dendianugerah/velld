import { AppShell } from "@/components/layout/app-shell";
import { ConnectionEditor } from "@/components/views/connections/connection-editor";

export default async function EditConnectionPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  return <AppShell><ConnectionEditor id={id} /></AppShell>;
}
