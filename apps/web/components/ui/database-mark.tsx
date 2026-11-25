import Image from "next/image";
import { Database } from "lucide-react";
import { cn } from "@/lib/utils";

// The database marks match the public homepage (Simple Icons 15, CC0).
const engines: Record<string, { file: string; className: string }> = {
  postgresql: { file: "/postgresql.svg", className: "border-[#dce9f3] bg-[#eef6fc]" },
  mysql: { file: "/mysql.svg", className: "border-[#efdfcd] bg-[#fcf4eb]" },
  mongodb: { file: "/mongodb.svg", className: "border-[#dce9d5] bg-[#f0f7ea]" },
};

export function DatabaseMark({ type, small = false }: { type: string; small?: boolean }) {
  const engine = engines[type];
  return <span aria-hidden="true" className={cn("flex shrink-0 items-center justify-center rounded-lg border", small ? "h-8 w-8" : "h-10 w-10", engine?.className ?? "bg-muted")}>
    {engine ? <Image src={engine.file} width={small ? 17 : 22} height={small ? 17 : 22} alt="" /> : <Database className="h-4 w-4 text-muted-foreground" />}
  </span>;
}
