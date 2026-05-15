package registration

import (
	"context"
	"time"

	"go.uber.org/zap"
)

const pollInterval = 10 * time.Second

// Worker polls for incomplete registrations and drives them to completion or rollback.
type Worker struct {
	regRepo   Repository
	processor *Processor
	log       *zap.Logger
}

func NewWorker(
	regRepo Repository,
	processor *Processor,
	log *zap.Logger,
) *Worker {
	return &Worker{
		regRepo:   regRepo,
		processor: processor,
		log:       log.Named("registration-worker"),
	}
}

// Run implements the worker.runnable interface.
func (w *Worker) Run(ctx context.Context) error {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			w.processActionable(ctx)
		}
	}
}

func (w *Worker) processActionable(ctx context.Context) {
	regs, err := w.regRepo.FindActionable(ctx)
	if err != nil {
		w.log.Error("failed to find actionable registrations", zap.Error(err))
		return
	}

	for _, reg := range regs {
		if ctx.Err() != nil {
			return
		}
		w.processOne(ctx, reg)
	}
}

func (w *Worker) processOne(ctx context.Context, reg *Registration) {
	log := w.log.With(zap.String("slug", reg.Slug), zap.String("status", string(reg.Status)))

	switch reg.Status {
	case StatusProvisioning:
		log.Debug("resuming registration saga")
		if _, err := w.processor.Process(ctx, reg); err != nil {
			log.Warn("saga step failed, will retry", zap.Error(err))
		}

	case StatusCompensating:
		log.Debug("compensating registration saga")
		if err := w.processor.Compensate(ctx, reg); err != nil {
			log.Warn("compensation step failed, will retry", zap.Error(err))
		}
	}

	if reg.RetryCount > 10 {
		log.Error("CRITICAL: registration stuck after many retries",
			zap.Int("retryCount", reg.RetryCount))
	}
}
