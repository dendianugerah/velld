import { PageHeader } from "@/components/layout/page-header";
import { Sidebar } from "@/components/layout/sidebar";

interface WorkspaceProps {
  title: string;
  description: React.ReactNode;
  section?: string;
  children: React.ReactNode;
}

export function Workspace({ title, description, section = "Workspace", children }: WorkspaceProps) {
  return (
    <div className="min-h-dvh lg:flex">
      <Sidebar />
      <main className="min-w-0 flex-1 px-4 pb-6 pt-5 sm:px-7 lg:pt-8">
        <p className="mx-auto mb-4 max-w-7xl text-[11px] font-semibold uppercase tracking-[0.18em] text-muted-foreground sm:mb-5">{section}</p>
        <div className="mx-auto max-w-7xl rounded-2xl border bg-card px-5 py-7 shadow-sm sm:px-9 sm:py-9 xl:px-14">
          <PageHeader title={title} description={description} />
          <div className="pt-7">{children}</div>
        </div>
      </main>
    </div>
  );
}
