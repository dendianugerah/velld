import { Skeleton } from "@/components/ui/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

export function ConnectionsTableSkeleton() {
  return (
    <div className="w-full">
      <div className="overflow-hidden">
        <Table>
          <TableHeader className="border-b border-border/50">
            <TableRow className="hover:bg-transparent">
              <TableHead className="w-[280px] font-semibold text-foreground/90 h-11">Name</TableHead>
              <TableHead className="font-semibold text-foreground/90 h-11">Type</TableHead>
              <TableHead className="font-semibold text-foreground/90 h-11">Host</TableHead>
              <TableHead className="font-semibold text-foreground/90 h-11">Schedule</TableHead>
              <TableHead className="font-semibold text-foreground/90 h-11">Last Backup</TableHead>
              <TableHead className="text-right font-semibold text-foreground/90 h-11">Size</TableHead>
              <TableHead className="text-right font-semibold text-foreground/90 h-11">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {[1, 2, 3, 4, 5].map((i) => (
              <TableRow key={i} className="border-b border-border/40">
                <TableCell className="py-3.5">
                  <Skeleton className="h-4 w-[200px]" />
                </TableCell>
                <TableCell className="py-3.5">
                  <Skeleton className="h-5 w-[80px] rounded-full" />
                </TableCell>
                <TableCell className="py-3.5">
                  <Skeleton className="h-4 w-[140px]" />
                </TableCell>
                <TableCell className="py-3.5">
                  <Skeleton className="h-4 w-[110px]" />
                </TableCell>
                <TableCell className="py-3.5">
                  <Skeleton className="h-4 w-[110px]" />
                </TableCell>
                <TableCell className="text-right py-3.5">
                  <Skeleton className="h-4 w-[60px] ml-auto" />
                </TableCell>
                <TableCell className="text-right py-3.5">
                  <div className="flex items-center justify-end gap-0.5">
                    <Skeleton className="h-8 w-8 rounded" />
                    <Skeleton className="h-8 w-8 rounded" />
                    <Skeleton className="h-8 w-8 rounded" />
                    <Skeleton className="h-8 w-8 rounded" />
                  </div>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>
    </div>
  );
}
