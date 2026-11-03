package backup

import (
	"context"
	"sync"
	"time"
)

// VerificationJob asks the worker to verify a backup.
type VerificationJob struct {
	BackupID string
}

// VerificationWorker runs verification jobs in the background so a slow
// restore-based check never blocks the API request that requested it. It
// retries transient failures and notifies the owner when a verification fails.
type VerificationWorker struct {
	service  *BackupService
	jobs     chan VerificationJob
	wg       sync.WaitGroup
	stopOnce sync.Once

	retries    int
	retryDelay time.Duration
}

const (
	// DefaultVerificationWorkers is the number of concurrent verification
	// workers started when the caller does not specify a count.
	DefaultVerificationWorkers = 2
	defaultVerificationQueue   = 128
	defaultVerificationRetries = 2
	defaultVerificationDelay   = 2 * time.Second
)

// NewVerificationWorker starts workers goroutines and returns a running worker.
// Call Stop to drain in-flight jobs.
func NewVerificationWorker(service *BackupService, workers int) *VerificationWorker {
	return newVerificationWorker(service, workers, defaultVerificationRetries, defaultVerificationDelay)
}

func newVerificationWorker(service *BackupService, workers, retries int, retryDelay time.Duration) *VerificationWorker {
	if workers < 1 {
		workers = 1
	}

	w := &VerificationWorker{
		service:    service,
		jobs:       make(chan VerificationJob, defaultVerificationQueue),
		retries:    retries,
		retryDelay: retryDelay,
	}

	w.wg.Add(workers)
	for i := 0; i < workers; i++ {
		go w.loop()
	}

	return w
}

// Enqueue schedules a backup for verification. It never blocks the caller; if
// the queue is full the job is dropped and false is returned so the handler can
// report back-pressure.
func (w *VerificationWorker) Enqueue(backupID string) bool {
	select {
	case w.jobs <- VerificationJob{BackupID: backupID}:
		return true
	default:
		return false
	}
}

func (w *VerificationWorker) loop() {
	defer w.wg.Done()
	for job := range w.jobs {
		w.process(job)
	}
}

func (w *VerificationWorker) process(job VerificationJob) {
	ctx := context.Background()

	var lastErr error
	for attempt := 0; attempt <= w.retries; attempt++ {
		lastErr = w.verifyOnce(ctx, job.BackupID)
		if lastErr == nil {
			return
		}
		if attempt < w.retries {
			time.Sleep(w.retryDelay)
		}
	}

	// Final failure: leave a notification so the user sees a bad backup even
	// without opening the history page.
	if err := w.service.createVerificationFailureNotification(job.BackupID, lastErr); err != nil {
		w.service.logf("failed to create verification failure notification: %v", err)
	}
}

// verifyOnce runs a single verification attempt and returns an error only for
// operational failures (the check itself failing is recorded, not returned).
func (w *VerificationWorker) verifyOnce(ctx context.Context, backupID string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = &panicError{value: r}
		}
	}()

	verification, err := w.service.createPendingVerification(backupID)
	if err != nil {
		return err
	}

	backup, err := w.service.backupRepo.GetBackup(backupID)
	if err != nil {
		return err
	}
	conn, err := w.service.connStorage.GetConnection(backup.ConnectionID)
	if err != nil {
		return err
	}

	if _, err := w.service.executeVerification(ctx, verification, backup, conn); err != nil {
		return err
	}
	return nil
}

// Stop closes the queue and waits for in-flight jobs to finish or ctx to end.
func (w *VerificationWorker) Stop(ctx context.Context) error {
	w.stopOnce.Do(func() { close(w.jobs) })

	done := make(chan struct{})
	go func() {
		w.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// panicError wraps a recovered panic so it can be returned as an error.
type panicError struct {
	value interface{}
}

func (e *panicError) Error() string {
	return "verification panicked"
}
