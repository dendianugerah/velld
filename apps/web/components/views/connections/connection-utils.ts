import { formatDistanceToNow, isValid, parseISO } from "date-fns";
import { getScheduleFrequency } from "@/lib/helper";
import type { Connection } from "@/types/connection";

export function connectionTime(value?: string) {
  if (!value) return "Never";
  const date = parseISO(value);
  return isValid(date) ? formatDistanceToNow(date, { addSuffix: true }) : "Unavailable";
}

export function scheduleLabel(connection: Connection) {
  if (!connection.backup_enabled) return "Manual";
  const labels: Record<string, string> = { hourly: "Hourly", daily: "Daily", weekly: "Weekly", monthly: "Monthly", test: "Every minute", custom: "Custom schedule" };
  return labels[getScheduleFrequency(connection.cron_schedule)] ?? "Scheduled";
}
