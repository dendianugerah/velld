package backup

import (
	"context"
	"testing"
	"time"

	"github.com/robfig/cron/v3"
)

func TestLeaderElector(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	first := NewLeaderElector(db, "scheduler", time.Minute)
	second := NewLeaderElector(db, "scheduler", time.Minute)

	got, err := first.TryAcquire(ctx)
	if err != nil || !got {
		t.Fatalf("first acquire: got=%v err=%v", got, err)
	}

	// A second instance should not get the lease while the first holds it.
	if got, err := second.TryAcquire(ctx); err != nil || got {
		t.Fatalf("second acquire should fail while held: got=%v err=%v", got, err)
	}

	// The holder can renew.
	if got, err := first.TryAcquire(ctx); err != nil || !got {
		t.Fatalf("holder renew should succeed: got=%v err=%v", got, err)
	}

	// Release frees it for the standby.
	if err := first.Release(ctx); err != nil {
		t.Fatalf("release: %v", err)
	}
	if got, err := second.TryAcquire(ctx); err != nil || !got {
		t.Fatalf("second acquire after release: got=%v err=%v", got, err)
	}
}

func TestLeaderElectorTakesExpiredLease(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Seed an already-expired lease held by someone else.
	seeder := &LeaderElector{db: db, name: "scheduler", holder: "other", ttl: -time.Minute}
	if err := seeder.initExpired(ctx); err != nil {
		t.Fatalf("seed expired lease: %v", err)
	}

	elector := NewLeaderElector(db, "scheduler", time.Minute)
	got, err := elector.TryAcquire(ctx)
	if err != nil {
		t.Fatalf("acquire: %v", err)
	}
	if !got {
		t.Fatal("expected to take over an expired lease")
	}
}

// initExpired inserts a lease that is already past its expiry.
func (e *LeaderElector) initExpired(ctx context.Context) error {
	_, err := e.db.ExecContext(ctx, `
		INSERT INTO scheduler_leases (name, holder, expires_at)
		VALUES ($1, $2, $3)`,
		e.name, e.holder, time.Now().Add(-time.Hour).Format(time.RFC3339))
	return err
}

func TestStartSchedulerWithElectorStartsAndStops(t *testing.T) {
	db := newTestDB(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	svc := &BackupService{
		backupRepo:  NewBackupRepository(db),
		cronManager: cron.New(cron.WithSeconds()), // not started; Start() is called by StartScheduler
		cronEntries: map[string]cron.EntryID{},
		sem:         make(chan struct{}, 1),
		running:     map[string]struct{}{},
		elector:     NewLeaderElector(db, "scheduler", time.Minute),
	}

	svc.StartScheduler(ctx)

	// Stop must not panic and must release the lease.
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer stopCancel()
	if err := svc.Stop(stopCtx); err != nil {
		t.Fatalf("stop: %v", err)
	}
}
