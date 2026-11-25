import type { Metadata } from "next";
import { Plus_Jakarta_Sans } from "next/font/google";
import "./globals.css";
import { QueryProvider } from "@/components/query-provider"
import { AuthMiddleware } from "@/components/auth-middleware";
import { Toaster } from "@/components/ui/toaster";

const geistSans = Plus_Jakarta_Sans({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: {
    default: "Velld",
    template: "%s | Velld",
  },
  description: "Automated scheduling, monitoring, and recovery for PostgreSQL, MySQL, and MongoDB.",
  icons: {
    icon: "/logo.ico",
    shortcut: "/logo.ico",
    apple: "/logo.png",
  },
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body
        className={`${geistSans.variable} font-sans antialiased bg-background text-foreground`}
      >
          <QueryProvider>
            <AuthMiddleware>
              {children}
            </AuthMiddleware>
          </QueryProvider>
          <Toaster />
      </body>
    </html>
  );
}
