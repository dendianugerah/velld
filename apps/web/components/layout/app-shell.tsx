import { Sidebar } from "@/components/layout/sidebar";

export function AppShell({ children }: { children: React.ReactNode }) {
  return <div className="isolate min-h-dvh bg-sidebar">
    <a href="#main-content" className="sr-only focus:not-sr-only focus:fixed focus:start-4 focus:top-4 focus:z-[100] focus:rounded-lg focus:bg-card focus:p-3">Skip to content</a>
    <Sidebar />
    <div className="p-0 lg:ms-[224px] lg:py-2 lg:pe-2">
      <main id="main-content" tabIndex={-1} className="min-h-[calc(100dvh-56px)] bg-card outline-none lg:min-h-[calc(100dvh-16px)] lg:overflow-clip lg:rounded-xl lg:border lg:border-border lg:shadow-sm">
        <div className="mx-auto max-w-[1120px] px-4 py-8 sm:px-8 lg:px-12 lg:py-12">
          {children}
        </div>
      </main>
    </div>
  </div>;
}
