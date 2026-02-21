'use client';

import { Clock, Calendar, Cloud, AlertTriangle, CheckCircle } from 'lucide-react';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Switch } from "@/components/ui/switch";
import { Label } from '@/components/ui/label';
import { Input } from '@/components/ui/input';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import type { Connection } from '@/types/connection';
import { useBackup } from '@/hooks/use-backup';
import { useSettings } from '@/hooks/use-settings';
import { updateConnectionSettings } from '@/lib/api/connections';
import { useToast } from '@/hooks/use-toast';
import { useState, useEffect, useMemo } from 'react';
import { getScheduleFrequency } from '@/lib/helper';

// Simple cron validation and next-run calculation
function parseCronExpression(expression: string): { valid: boolean; nextRuns: Date[]; tooFrequent: boolean; error?: string } {
  const parts = expression.trim().split(/\s+/);
  if (parts.length !== 6) {
    return { valid: false, nextRuns: [], tooFrequent: false, error: 'Cron expression must have 6 fields' };
  }

  const [secPart, minPart, hourPart, domPart, monthPart, dowPart] = parts;
  let runsPerMinute = 1;
  if (secPart === '*') runsPerMinute = 60;
  else if (secPart.includes(',')) runsPerMinute = secPart.split(',').length;
  else if (secPart.includes('/')) {
    const step = parseInt(secPart.split('/')[1] || '1');
    runsPerMinute = Math.ceil(60 / step);
  }

  const tooFrequent = runsPerMinute > 1;

  try {
    // Generate next few run times for preview
    const nextRuns: Date[] = [];
    let current = new Date();
    current.setMilliseconds(0);

    // Simple cron parser for preview (handles basic patterns)
    for (let i = 0; i < 3 && nextRuns.length < 3; i++) {
      current = new Date(current.getTime() + 1000);

      // Check if current time matches cron pattern
      const matches = (
        matchesField(current.getSeconds(), secPart) &&
        matchesField(current.getMinutes(), minPart) &&
        matchesField(current.getHours(), hourPart) &&
        matchesField(current.getDate(), domPart) &&
        matchesField(current.getMonth() + 1, monthPart) &&
        matchesField(current.getDay(), dowPart)
      );

      if (matches) {
        nextRuns.push(new Date(current));
      }

      // Safety: don't iterate more than 2 years ahead
      if (current.getTime() - Date.now() > 2 * 365 * 24 * 60 * 60 * 1000) break;
    }

    return { valid: true, nextRuns, tooFrequent };
  } catch {
    return { valid: false, nextRuns: [], tooFrequent: false, error: 'Invalid cron expression' };
  }
}

function matchesField(value: number, pattern: string): boolean {
  if (pattern === '*') return true;
  if (pattern === '?') return true;

  // Handle ranges like 1-5
  if (pattern.includes('-')) {
    const [start, end] = pattern.split('-').map(Number);
    return value >= start && value <= end;
  }

  // Handle steps like */5 or 0/5
  if (pattern.includes('/')) {
    const [base, step] = pattern.split('/');
    const stepNum = parseInt(step);
    const baseNum = base === '*' ? 0 : parseInt(base);
    return (value - baseNum) % stepNum === 0 && value >= baseNum;
  }

  // Handle lists like 1,3,5
  if (pattern.includes(',')) {
    const values = pattern.split(',').map(Number);
    return values.includes(value);
  }

  // Handle single value
  const numVal = parseInt(pattern);
  return value === numVal;
}

const CRON_SCHEDULES = {
  'test': '0 */1 * * * *',  // Run every 1 minutes
  'hourly': '0 0 * * * *',
  'daily': '0 0 0 * * *',
  'weekly': '0 0 0 * * 0',
  'monthly': '0 0 0 1 * *',
  'custom': 'custom'
};

const RETENTION_DAYS = {
  '7': 7,
  '30': 30,
  '90': 90,
  '365': 365
};

interface BackupScheduleDialogProps {
  connectionId: string | null;
  connection?: Connection;
  onClose: () => void;
}

export function BackupScheduleDialog({
  connectionId,
  connection,
  onClose,
}: BackupScheduleDialogProps) {
  const { createSchedule, updateExistingSchedule, disableSchedule, isScheduling, isDisabling, isUpdating } = useBackup();
  const { settings } = useSettings();
  const { toast } = useToast();
  const [enabled, setEnabled] = useState(false);
  const [s3Cleanup, setS3Cleanup] = useState(true);

  const getRetentionFromDays = (days?: number | null) => {
    if (!days) return '30';
    return days.toString();
  };

  const [schedule, setSchedule] = useState(getScheduleFrequency(connection?.cron_schedule) || 'daily');
  const [retention, setRetention] = useState(getRetentionFromDays(connection?.retention_days));
  const [customCron, setCustomCron] = useState(connection?.cron_schedule || '0 0 * * * *');

  // Validate cron expression
  const cronValidation = useMemo(() => {
    if (schedule !== 'custom') return { valid: true, nextRuns: [], tooFrequent: false };
    return parseCronExpression(customCron);
  }, [customCron, schedule]);

  useEffect(() => {
    if (connection) {
      setEnabled(connection.backup_enabled);
      const freq = getScheduleFrequency(connection.cron_schedule);
      setSchedule(freq || 'daily');
      if (freq === 'custom' && connection.cron_schedule) {
        setCustomCron(connection.cron_schedule);
      }
      setRetention(getRetentionFromDays(connection.retention_days));
      setS3Cleanup(connection.s3_cleanup_on_retention ?? true);
    }
  }, [connection]);

  if (!connectionId || !connection) return null;

  const handleScheduleChange = async (checked: boolean) => {
    try {
      if (checked) {
        const cronSchedule = schedule === 'custom' ? customCron : CRON_SCHEDULES[schedule as keyof typeof CRON_SCHEDULES];
        await createSchedule({
          connection_id: connectionId,
          cron_schedule: cronSchedule,
          retention_days: RETENTION_DAYS[retention as keyof typeof RETENTION_DAYS]
        });
        setEnabled(true);
      } else {
        await disableSchedule(connectionId);
        setEnabled(false);
      }
    } catch (error) {
      // If there's an error, revert the switch state
      setEnabled(!checked);
      console.error('Failed to update schedule:', error);
    }
  };

  const handleScheduleSubmit = async (newSchedule: string, newRetention: string, newCustomCron?: string) => {
    try {
      const cronValue = newCustomCron || customCron;
      const cronSchedule = newSchedule === 'custom' ? cronValue : CRON_SCHEDULES[newSchedule as keyof typeof CRON_SCHEDULES];

      if (enabled) {
        // Update existing schedule
        await updateExistingSchedule({
          connectionId,
          params: {
            cron_schedule: cronSchedule,
            retention_days: RETENTION_DAYS[newRetention as keyof typeof RETENTION_DAYS]
          }
        });
      } else {
        // Create new schedule
        await createSchedule({
          connection_id: connectionId,
          cron_schedule: cronSchedule,
          retention_days: RETENTION_DAYS[newRetention as keyof typeof RETENTION_DAYS]
        });
        setEnabled(true);
      }
      setSchedule(newSchedule);
      setRetention(newRetention);
    } catch (error) {
      console.error('Failed to update schedule:', error);
    }
  };

  return (
    <Dialog open={!!connectionId} onOpenChange={() => onClose()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Backup Schedule - {connection.name}</DialogTitle>
        </DialogHeader>
        <div className="space-y-6 py-4">
          <div className="flex items-center justify-between space-x-4">
            <div className="space-y-0.5">
              <Label>Automatic Backups</Label>
              <div className="text-sm text-muted-foreground">
                Enable scheduled backups for this database
              </div>
            </div>
            <Switch
              checked={enabled}
              disabled={isScheduling || isDisabling || isUpdating}
              onCheckedChange={handleScheduleChange}
            />
          </div>

          {enabled && (
            <>
              <div className="space-y-2">
                <Label className="text-sm flex items-center">
                  <Clock className="h-4 w-4 mr-1" />
                  Backup Frequency
                </Label>
                <Select
                  value={schedule}
                  onValueChange={(value) => handleScheduleSubmit(value, retention)}
                  disabled={isScheduling}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="test">Every Minute (Test)</SelectItem>
                    <SelectItem value="hourly">Every Hour</SelectItem>
                    <SelectItem value="daily">Daily</SelectItem>
                    <SelectItem value="weekly">Weekly</SelectItem>
                    <SelectItem value="monthly">Monthly</SelectItem>
                    <SelectItem value="custom">Custom (Cron)</SelectItem>
                  </SelectContent>
                </Select>

                {schedule === 'custom' && (
                  <div className="space-y-3 mt-3">
                    <div className="space-y-2">
                      <Label className="text-sm text-muted-foreground">Custom Cron Expression</Label>
                      <Input
                        placeholder="0 0 * * * *"
                        value={customCron}
                        onChange={(e) => {
                          setCustomCron(e.target.value);
                        }}
                        onBlur={() => {
                          if (cronValidation.valid && !cronValidation.tooFrequent) {
                            handleScheduleSubmit(schedule, retention, customCron);
                          }
                        }}
                        disabled={isScheduling}
                        className={!cronValidation.valid ? 'border-red-500' : cronValidation.tooFrequent ? 'border-yellow-500' : ''}
                      />
                      {!cronValidation.valid && (
                        <div className="flex items-center gap-2 text-xs text-red-600">
                          <AlertTriangle className="h-3 w-3" />
                          {cronValidation.error || 'Invalid cron expression'}
                        </div>
                      )}
                      {cronValidation.valid && cronValidation.tooFrequent && (
                        <div className="flex items-center gap-2 text-xs text-yellow-600">
                          <AlertTriangle className="h-3 w-3" />
                          Warning: This will run multiple times per minute. Use &quot;0&quot; for seconds to run once per minute.
                        </div>
                      )}
                      {cronValidation.valid && !cronValidation.tooFrequent && cronValidation.nextRuns.length > 0 && (
                        <div className="flex items-start gap-2 text-xs text-green-600">
                          <CheckCircle className="h-3 w-3 mt-0.5" />
                          <div>
                            <div>Next run: {cronValidation.nextRuns[0]?.toLocaleString()}</div>
                            <div className="text-muted-foreground mt-1">
                              Format: seconds minutes hours day month weekday
                            </div>
                          </div>
                        </div>
                      )}
                    </div>

                    {/* Quick presets */}
                    <div className="pt-2 border-t">
                      <p className="text-xs text-muted-foreground mb-2">Quick presets:</p>
                      <div className="flex flex-wrap gap-2">
                        {[
                          { label: 'Every 5 min', value: '0 */5 * * * *' },
                          { label: 'Every 15 min', value: '0 */15 * * * *' },
                          { label: 'Every 30 min', value: '0 */30 * * * *' },
                          { label: 'At 2 AM', value: '0 0 2 * * *' },
                          { label: 'At 6 PM', value: '0 0 18 * * *' },
                        ].map((preset) => (
                          <button
                            key={preset.value}
                            type="button"
                            onClick={() => {
                              setCustomCron(preset.value);
                              handleScheduleSubmit(schedule, retention, preset.value);
                            }}
                            className="text-xs px-2 py-1 rounded bg-secondary hover:bg-secondary/80 transition-colors"
                          >
                            {preset.label}
                          </button>
                        ))}
                      </div>
                    </div>
                  </div>
                )}
              </div>

              <div className="space-y-2">
                <Label className="text-sm flex items-center">
                  <Calendar className="h-4 w-4 mr-1" />
                  Retention Period
                </Label>
                <Select
                  value={retention}
                  onValueChange={(value) => handleScheduleSubmit(schedule, value)}
                  disabled={isScheduling}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="7">7 Days</SelectItem>
                    <SelectItem value="30">30 Days</SelectItem>
                    <SelectItem value="90">90 Days</SelectItem>
                    <SelectItem value="365">1 Year</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              {settings?.s3_enabled && (
                <div className="flex items-center justify-between space-x-4 rounded-lg border p-4 bg-background/50">
                  <div className="space-y-0.5">
                    <Label className="text-sm font-medium flex items-center gap-2">
                      <Cloud className="h-4 w-4" />
                      Cleanup S3 Backups on Retention
                    </Label>
                    <div className="text-sm text-muted-foreground">
                      Automatically delete S3 backups older than retention period
                    </div>
                  </div>
                  <Switch
                    checked={s3Cleanup}
                    onCheckedChange={async (checked) => {
                      const previousValue = s3Cleanup;
                      setS3Cleanup(checked);
                      if (connectionId) {
                        try {
                          await updateConnectionSettings(connectionId, {
                            s3_cleanup_on_retention: checked,
                          });
                          toast({
                            title: "Success",
                            description: `S3 cleanup on retention ${checked ? 'enabled' : 'disabled'}`,
                          });
                        } catch (error) {
                          setS3Cleanup(previousValue);
                          const errorMessage = error instanceof Error ? error.message : 'Failed to update S3 cleanup setting';
                          console.error('S3 cleanup setting error:', error);
                          toast({
                            title: "Error",
                            description: errorMessage,
                            variant: "destructive",
                          });
                        }
                      }
                    }}
                  />
                </div>
              )}
            </>
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}