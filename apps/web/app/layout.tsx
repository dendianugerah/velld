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

// const geistMono = Geist_Mono({
//   variable: "--font-geist-mono",
//   subsets: ["latin"],
// });

export const metadata: Metadata = {
  title: "Velld",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" className="dark">
      <body
        className={`${geistSans.variable}  antialiased bg-background text-foreground`}
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
