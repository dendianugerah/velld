package backup

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// LeaderElector decides which process runs the backup scheduler. With more than
// one replicas, letting every replica run cron means every scheduled backup
// fires N times. A database lease is enough: the holder renews it, and only the
// holder schedules. This trades the generality of Raft for something that works
// on SQLite and Postgres alike and is easy to reason about.
type LeaderElector struct {
	db      *sql.DB
	name    string
	holder  string
	ttl     time.Duration
	leaseID string
}

// NewLeaderElector creates an elector for the named lease.
func NewLeaderElector(db *sql.DB, name string, ttl time.Duration) *LeaderElector {
	return &LeaderElector{
		db:     db,
		name:   name,
		holder: uuid.New().String(),
		ttl:    ttl,
	}
}

// TryAcquire attempts to take or renew the lease. It succeeds when the lease is
// free, expired, or already held by this instance.
func (e *LeaderElector) TryAcquire(ctx context.Context) (bool, error) {
	now := time.Now()
	expires := now.Add(e.ttl)

	// Take over a free or expired lease.
	res, err := e.db.ExecContext(ctx, `
		UPDATE scheduler_leases
		SET holder = $1, expires_at = $2, updated_at = $2
		WHERE name = $3 AND (holder = $1 OR expires_at < $2)`,
		e.holder, expires.Format(time.RFC3339), e.name)
	if err != nil {
		return false, fmt.Errorf("renew lease: %w", err)
	}

	if n, _ := res.RowsAffected(); n == 1 {
		return true, nil
	}

	// Row may not exist yet; insert it. A concurrent insert fails the
	// primary-key constraint and falls through to a re-check.
	_, err = e.db.ExecContext(ctx, `
		INSERT INTO scheduler_leases (name, holder, expires_at, updated_at)
		VALUES ($1, $2, $3, $3)`,
		e.name, e.holder, expires.Format(time.RFC3339))
	if err != nil {
		// Lost the race to insert; re-read to see if we nonetheless hold it.
		return e.stillHolder(ctx)
	}

	return true, nil
}

func (e *LeaderElector) stillHolder(ctx context.Context) (bool, error) {
	var holder string
	err := e.db.QueryRowContext(ctx,
		`SELECT holder FROM scheduler_leases WHERE name = $1`, e.name).Scan(&holder)
	if err != nil {
		return false, err
	}
	return holder == e.holder, nil
}

// Release frees the lease if this instance holds it, so a clean shutdown lets a
// standby take over immediately.
func (e *LeaderElector) Release(ctx context.Context) error {
	_, err := e.db.ExecContext(ctx,
		`DELETE FROM scheduler_leases WHERE name = $1 AND holder = $2`, e.name, e.holder)
	return err
}
