"use client";

import { Popover } from "@cloudflare/kumo/components/popover";
import { CircleHelp } from "lucide-react";
import type { ReactNode } from "react";

export function FieldHelp({
  label,
  children,
}: {
  label: string;
  children: ReactNode;
}) {
  return (
    <Popover.Root>
      <Popover.Trigger
        aria-label={label}
        render={
          <button
            type="button"
            className="inline-flex size-6 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground focus-visible:outline focus-visible:outline-2 focus-visible:outline-ring"
          />
        }
      >
        <CircleHelp className="size-4" aria-hidden="true" />
      </Popover.Trigger>
      <Popover.Portal>
        <Popover.Positioner side="top" align="start" sideOffset={6}>
          <Popover.Popup className="max-w-64 text-sm leading-relaxed">
            <Popover.Description>{children}</Popover.Description>
          </Popover.Popup>
        </Popover.Positioner>
      </Popover.Portal>
    </Popover.Root>
  );
}
