package tenant

import (
	"context"
	"errors"
	"fmt"

	"github.com/Sokol111/ecommerce-commons/pkg/core/logger"
	"github.com/Sokol111/ecommerce-commons/pkg/messaging/patterns/outbox"
	"github.com/Sokol111/ecommerce-tenant-service/internal/domain/registration"
	"github.com/Sokol111/ecommerce-tenant-service/internal/domain/tenant"
	"github.com/Sokol111/ecommerce-tenant-service/internal/event"
	"github.com/Sokol111/ecommerce-tenant-service/internal/infrastructure/logto"
	"go.uber.org/zap"
)

// SagaProcessor executes and compensates registration saga steps.
type SagaProcessor struct {
	tenantRepo   tenant.Repository
	regRepo      registration.Repository
	outbox       outbox.Outbox
	eventFactory event.TenantEventFactory
	idp          tenant.IdentityProvider
}

func NewSagaProcessor(
	tenantRepo tenant.Repository,
	regRepo registration.Repository,
	outbox outbox.Outbox,
	eventFactory event.TenantEventFactory,
	idp tenant.IdentityProvider,
) *SagaProcessor {
	return &SagaProcessor{
		tenantRepo:   tenantRepo,
		regRepo:      regRepo,
		outbox:       outbox,
		eventFactory: eventFactory,
		idp:          idp,
	}
}

// Process executes the registration saga steps forward. Returns the tenant on success.
func (p *SagaProcessor) Process(ctx context.Context, reg *registration.Registration) (*tenant.Tenant, error) {
	log := logger.Get(ctx)

	if err := p.stepCreateTenant(ctx, reg, log); err != nil {
		return nil, err
	}

	if err := p.stepSetTenantOnUser(ctx, reg, log); err != nil {
		return nil, err
	}

	if err := p.stepAssignRole(ctx, reg, log); err != nil {
		return nil, err
	}

	if err := p.stepPublishEvent(ctx, reg, log); err != nil {
		return nil, err
	}

	reg.MarkCompleted()
	if err := p.regRepo.Update(ctx, reg); err != nil {
		return nil, fmt.Errorf("failed to mark completed: %w", err)
	}

	log.Info("registration saga completed", zap.String("slug", reg.Slug))

	t, err := p.tenantRepo.FindByID(ctx, *reg.TenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve tenant after completion: %w", err)
	}
	return t, nil
}

func (p *SagaProcessor) stepCreateTenant(ctx context.Context, reg *registration.Registration, log *zap.Logger) error {
	if reg.TenantID != nil {
		return nil
	}

	t, err := p.createTenant(ctx, reg)
	if err != nil {
		p.handleStepFailure(ctx, reg, err, log)
		return err
	}
	reg.SetTenantID(t.ID)
	if err := p.regRepo.Update(ctx, reg); err != nil {
		return fmt.Errorf("failed to persist tenant ID: %w", err)
	}
	return nil
}

func (p *SagaProcessor) stepSetTenantOnUser(ctx context.Context, reg *registration.Registration, log *zap.Logger) error {
	if reg.TenantSet {
		return nil
	}

	if err := p.idp.SetUserTenant(ctx, *reg.LogtoUserID, reg.Slug); err != nil {
		p.handleStepFailure(ctx, reg, err, log)
		return err
	}
	reg.SetTenantOnUser()
	if err := p.regRepo.Update(ctx, reg); err != nil {
		return fmt.Errorf("failed to persist tenant-set: %w", err)
	}
	return nil
}

func (p *SagaProcessor) stepAssignRole(ctx context.Context, reg *registration.Registration, log *zap.Logger) error {
	if reg.RoleAssigned {
		return nil
	}

	if err := p.idp.AssignRole(ctx, *reg.LogtoUserID, "super_admin"); err != nil {
		p.handleStepFailure(ctx, reg, err, log)
		return err
	}
	reg.SetRoleAssigned()
	if err := p.regRepo.Update(ctx, reg); err != nil {
		return fmt.Errorf("failed to persist role-assigned: %w", err)
	}
	return nil
}

func (p *SagaProcessor) stepPublishEvent(ctx context.Context, reg *registration.Registration, log *zap.Logger) error {
	if reg.EventPublished {
		return nil
	}

	if err := p.publishTenantEvent(ctx, reg); err != nil {
		p.handleStepFailure(ctx, reg, err, log)
		return err
	}
	reg.SetEventPublished()
	return nil
}

// Compensate reverses the registration saga steps (cleanup).
func (p *SagaProcessor) Compensate(ctx context.Context, reg *registration.Registration) error {
	log := logger.Get(ctx)

	// Reverse order: delete user first, then tenant
	if reg.LogtoUserID != nil {
		if err := p.idp.DeleteUser(ctx, *reg.LogtoUserID); err != nil {
			log.Error("compensation: failed to delete user, scheduling retry",
				zap.String("slug", reg.Slug), zap.Error(err))
			reg.ScheduleRetry()
			_ = p.regRepo.Update(ctx, reg)
			return err
		}
		reg.ClearLogtoUser()
		if err := p.regRepo.Update(ctx, reg); err != nil {
			return fmt.Errorf("failed to persist user deletion: %w", err)
		}
	}

	if reg.TenantID != nil {
		if err := p.tenantRepo.Delete(ctx, *reg.TenantID); err != nil {
			log.Error("compensation: failed to delete tenant, scheduling retry",
				zap.String("slug", reg.Slug), zap.Error(err))
			reg.ScheduleRetry()
			_ = p.regRepo.Update(ctx, reg)
			return err
		}
		reg.ClearTenant()
		if err := p.regRepo.Update(ctx, reg); err != nil {
			return fmt.Errorf("failed to persist tenant deletion: %w", err)
		}
	}

	reg.MarkRolledBack()
	if err := p.regRepo.Update(ctx, reg); err != nil {
		return fmt.Errorf("failed to mark rolled back: %w", err)
	}

	log.Info("registration saga rolled back", zap.String("slug", reg.Slug))
	return nil
}

func (p *SagaProcessor) createTenant(ctx context.Context, reg *registration.Registration) (*tenant.Tenant, error) {
	t, err := tenant.NewTenant(reg.Slug, reg.Name)
	if err != nil {
		return nil, err
	}

	if err := p.tenantRepo.Insert(ctx, t); err != nil {
		return nil, fmt.Errorf("failed to insert tenant: %w", err)
	}

	return t, nil
}

func (p *SagaProcessor) publishTenantEvent(ctx context.Context, reg *registration.Registration) error {
	t, err := p.tenantRepo.FindBySlug(ctx, reg.Slug)
	if err != nil {
		return fmt.Errorf("failed to load tenant for event: %w", err)
	}

	msg := p.eventFactory.NewTenantUpdatedOutboxMessage(ctx, t)
	send, err := p.outbox.Create(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to create outbox: %w", err)
	}

	_ = send(ctx) //nolint:errcheck // best-effort send
	return nil
}

func (p *SagaProcessor) handleStepFailure(ctx context.Context, reg *registration.Registration, err error, log *zap.Logger) {
	if isPermanentError(err) {
		log.Warn("permanent error in registration saga, compensating",
			zap.String("slug", reg.Slug), zap.Error(err))
		reg.MarkCompensating(err.Error())
	} else {
		log.Warn("transient error in registration saga, scheduling retry",
			zap.String("slug", reg.Slug), zap.Error(err))
		reg.ScheduleRetry()
	}
	_ = p.regRepo.Update(ctx, reg)
}

func isPermanentError(err error) bool {
	return errors.Is(err, logto.ErrUserAlreadyExists) ||
		errors.Is(err, tenant.ErrSlugAlreadyExists) ||
		errors.Is(err, tenant.ErrInvalidTenantData)
}
