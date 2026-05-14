package worker

import (
	"context"
	"time"

	command "github.com/Sokol111/ecommerce-tenant-service/internal/application/command/tenant"
	"github.com/Sokol111/ecommerce-tenant-service/internal/domain/registration"
	"go.uber.org/zap"
)

const pollInterval = 10 * time.Second

// RegistrationWorker polls for incomplete registrations and drives them to completion or rollback.
type RegistrationWorker struct {
	regRepo   registration.Repository
	processor *command.SagaProcessor
	log       *zap.Logger
}

func NewRegistrationWorker(
	regRepo registration.Repository,
	processor *command.SagaProcessor,
	log *zap.Logger,
) *RegistrationWorker {
	return &RegistrationWorker{
		regRepo:   regRepo,
		processor: processor,
		log:       log.Named("registration-worker"),
	}
}

// Run implements the worker.runnable interface.
func (w *RegistrationWorker) Run(ctx context.Context) error {
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

func (w *RegistrationWorker) processActionable(ctx context.Context) {
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

func (w *RegistrationWorker) processOne(ctx context.Context, reg *registration.Registration) {
	log := w.log.With(zap.String("slug", reg.Slug), zap.String("status", string(reg.Status)))

	switch reg.Status {
	case registration.StatusProvisioning:
		log.Debug("resuming registration saga")
		if _, err := w.processor.Process(ctx, reg); err != nil {
			log.Warn("saga step failed, will retry", zap.Error(err))
		}

	case registration.StatusCompensating:
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
