import './global.css';
import { RootProvider } from 'fumadocs-ui/provider';
import { Inter } from 'next/font/google';
import type { ReactNode } from 'react';
import type { Metadata } from 'next';

const inter = Inter({
  subsets: ['latin'],
});

export const metadata: Metadata = {
  title: {
    default: 'Velld - Database Backup Management Tool',
    template: '%s | Velld',
  },
  description: 'Automated scheduling, monitoring, and recovery for PostgreSQL, MySQL, and MongoDB. Open-source database backup management made easy.',
  keywords: ['database backup', 'postgresql', 'mysql', 'mongodb', 'backup automation', 'database management', 'open source'],
  authors: [{ name: 'Velld' }],
  openGraph: {
    type: 'website',
    locale: 'en_US',
    url: 'https://velld.io',
    title: 'Velld - Database Backup Management Tool',
    description: 'Automated scheduling, monitoring, and recovery for PostgreSQL, MySQL, and MongoDB.',
    siteName: 'Velld',
  },
  twitter: {
    card: 'summary_large_image',
    title: 'Velld - Database Backup Management Tool',
    description: 'Automated scheduling, monitoring, and recovery for PostgreSQL, MySQL, and MongoDB.',
  },
  icons: {
    icon: '/images/logo.ico',
    shortcut: '/images/logo.ico',
    apple: '/images/logo.png',
  },
};

export default function Layout({ children }: { children: ReactNode }) {
  return (
    <html lang="en" className={inter.className} suppressHydrationWarning>
      <body className="flex flex-col min-h-screen">
        <RootProvider>{children}</RootProvider>
      </body>
    </html>
  );
}
