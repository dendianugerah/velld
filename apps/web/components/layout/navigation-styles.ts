import { cn } from "@/lib/utils";

export function navigationItemClass(active: boolean) {
  return cn(
    "group flex min-h-9 items-center gap-2.5 rounded-md px-2.5 text-left text-[13px] transition-colors duration-150 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary",
    active ? "bg-sidebar-active font-medium text-foreground" : "font-normal text-muted-foreground hover:bg-sidebar-hover hover:text-foreground"
  );
}

export const navigationIconClass = "h-4 w-4 shrink-0 text-muted-foreground transition-colors group-hover:text-foreground group-aria-[current=page]:text-foreground group-aria-pressed:text-foreground";
