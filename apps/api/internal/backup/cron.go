package backup

import (
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
)

// cronScheduleParser accepts six-field schedules (seconds included) because
// Velld lets users pick sub-minute intervals.
var cronScheduleParser = cron.NewParser(
	cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow,
)

func parseCronSchedule(expression string) (cron.Schedule, error) {
	schedule, err := cronScheduleParser.Parse(expression)
	if err != nil {
		return nil, fmt.Errorf("invalid cron schedule: %w", err)
	}
	return schedule, nil
}

func nextRunTime(expression string, from time.Time) (time.Time, error) {
	schedule, err := parseCronSchedule(expression)
	if err != nil {
		return time.Time{}, err
	}
	return schedule.Next(from), nil
}
