import Link from "next/link";
import { Plus } from "lucide-react";
import { buttonVariants } from "@/components/ui/button";

// Keep the Overview entry point while using the same page-based flow everywhere.
export function AddConnectionDialog() {
  return <Link href="/connections/new" className={buttonVariants()}><Plus aria-hidden="true" />Add connection</Link>;
}
