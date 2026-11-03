import { Badge } from "@/components/ui/badge";
import { ShieldCheck, ShieldAlert, Shield } from "lucide-react";
import { formatDistanceToNow, parseISO } from "date-fns";
import { cn } from "@/lib/utils";

interface VerificationBadgeProps {
  status?: string;
  verifiedAt?: string;
  className?: string;
}

function verifiedTitle(status?: string, verifiedAt?: string): string | undefined {
  if (status !== "passed" || !verifiedAt) {
    return undefined;
  }
  try {
    return `Verified ${formatDistanceToNow(parseISO(verifiedAt), { addSuffix: true })}`;
  } catch {
    return undefined;
  }
}

export function VerificationBadge({ status, verifiedAt, className }: VerificationBadgeProps) {
  const title = verifiedTitle(status, verifiedAt);

  if (status === "passed") {
    return (
      <Badge
        variant="secondary"
        title={title}
        className={cn("text-xs bg-emerald-100 text-emerald-700 dark:bg-emerald-900 dark:text-emerald-300", className)}
      >
        <ShieldCheck className="h-3 w-3 mr-1" />
        Verified
      </Badge>
    );
  }

  if (status === "failed") {
    return (
      <Badge variant="destructive" className={cn("text-xs", className)}>
        <ShieldAlert className="h-3 w-3 mr-1" />
        Verify failed
      </Badge>
    );
  }

  return (
    <Badge variant="outline" className={cn("text-xs text-muted-foreground", className)}>
      <Shield className="h-3 w-3 mr-1" />
      Unverified
    </Badge>
  );
}
