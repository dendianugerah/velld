import type { BaseLayoutProps } from 'fumadocs-ui/layouts/shared';

/**
 * Shared layout configurations
 *
 * you can customise layouts individually from:
 * Home Layout: app/(home)/layout.tsx
 * Docs Layout: app/docs/layout.tsx
 */
export const baseOptions: BaseLayoutProps = {
  nav: {
    title: (
      <>
        <img
          src="/images/logo.png"
          alt="Velld Logo"
          width={32}
          height={32}
          className="rounded-lg"
        />
        <span className="text-base font-semibold bg-clip-text text-transparent bg-gradient-to-r from-blue-500 to-blue-600">
          Velld
        </span>
      </>
    ),
  },
  // see https://fumadocs.dev/docs/ui/navigation/links
  links: [],
};
