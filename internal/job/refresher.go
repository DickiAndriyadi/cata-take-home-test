package job

import (
	"context"
	"time"

	"cata-take-home-test/internal/api"

	"go.uber.org/zap"
)

type Refresher struct {
	handler  *api.Handler
	interval time.Duration
	logger   *zap.Logger
}

func NewRefresher(handler *api.Handler, interval time.Duration, logger *zap.Logger) *Refresher {
	return &Refresher{
		handler:  handler,
		interval: interval,
		logger:   logger,
	}
}

// Start runs the background refresh job every interval until context cancellation.
func (r *Refresher) Start(ctx context.Context) {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	r.syncWithRetry(ctx)

	for {
		select {
		case <-ctx.Done():
			r.logger.Info("Refresher stopped due to context cancellation")
			return
		case <-ticker.C:
			r.syncWithRetry(ctx)
		}
	}
}

// syncWithRetry retries sync with exponential backoff
func (r *Refresher) syncWithRetry(ctx context.Context) {
	maxAttempts := 5
	initialBackoff := 1 * time.Second

	for attempt := 0; attempt < maxAttempts; attempt++ {
		err := r.handler.Sync(ctx)
		if err == nil {
			r.logger.Info("Background sync successful")
			return
		}

		backoffDuration := initialBackoff << attempt // exponential backoff
		r.logger.Warn("Sync failed, retrying later",
			zap.Int("attempt", attempt+1),
			zap.Duration("retry_in", backoffDuration),
			zap.Error(err),
		)

		select {
		case <-ctx.Done():
			r.logger.Info("Refresher stopped during backoff")
			return
		case <-time.After(backoffDuration):
			// continue retry
		}
	}
	r.logger.Error("Background sync failed after max retry attempts")
}
