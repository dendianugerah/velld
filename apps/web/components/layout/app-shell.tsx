import { Sidebar } from "@/components/layout/sidebar";

export function AppShell({ children }: { children: React.ReactNode }) {
  return <div className="min-h-dvh bg-[#f7f7f8]">
    <a href="#main-content" className="sr-only focus:not-sr-only focus:fixed focus:start-4 focus:top-4 focus:z-[100] focus:rounded-lg focus:bg-white focus:p-3">Skip to content</a>
    <Sidebar />
    <div className="p-0 lg:ms-[224px] lg:p-2 lg:ps-0">
      <main id="main-content" tabIndex={-1} className="min-h-[calc(100dvh-56px)] bg-white outline-none lg:min-h-[calc(100dvh-16px)] lg:rounded-xl lg:border lg:shadow-[0_2px_5px_#1d253008]">
        <div className="mx-auto max-w-[1400px] px-4 py-6 sm:px-7 lg:px-9 lg:py-7 xl:px-10">
          {children}
        </div>
      </main>
    </div>
  </div>;
}
