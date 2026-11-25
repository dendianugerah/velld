"use client";

import { useEffect } from "react";
import { useRouter, usePathname } from "next/navigation";
import { isAuthenticated } from "@/lib/helper";

export function AuthMiddleware({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const pathname = usePathname();

  useEffect(() => {
    if (!isAuthenticated() && pathname !== "/login") {
      router.push("/login");
    } else if (isAuthenticated() && pathname === "/login") {
      router.push("/");
    }
  }, [pathname, router]);

  return <>{children}</>;
}
