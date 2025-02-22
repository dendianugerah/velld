"use client";

import { useEffect } from "react";
import { useRouter, usePathname } from "next/navigation";
import { checkIsAllowRegister, isAuthenticated } from "@/lib/helper";

const publicPaths = ["/login", "/register"];

export function AuthMiddleware({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const pathname = usePathname();

  useEffect(() => {
    if (!isAuthenticated() && !publicPaths.includes(pathname)) {
      router.push("/login");
    } else if (isAuthenticated() && publicPaths.includes(pathname)) {
      router.push("/");
    } else if (!checkIsAllowRegister() && pathname === "/register") {
      router.push("/login");
    }
  }, [pathname, router]);

  return <>{children}</>;
}
