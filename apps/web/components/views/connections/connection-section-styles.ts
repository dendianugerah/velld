import { cn } from "@/lib/utils";

export const connectionSectionsClass =
  "flex flex-wrap gap-x-6 gap-y-1 border-b";

export function connectionSectionClass(active: boolean) {
  return cn(
    "-mb-px cursor-pointer border-b-2 py-3 text-[13px] font-medium focus-visible:outline focus-visible:outline-2 focus-visible:outline-ring",
    active
      ? "border-foreground text-foreground"
      : "border-transparent text-muted-foreground hover:border-border hover:text-foreground",
  );
}
