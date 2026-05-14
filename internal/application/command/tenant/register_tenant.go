package tenant

import (
	"context"
	"fmt"

	"github.com/Sokol111/ecommerce-commons/pkg/core/logger"
	"github.com/Sokol111/ecommerce-tenant-service/internal/domain/registration"
	"github.com/Sokol111/ecommerce-tenant-service/internal/domain/tenant"
	"go.uber.org/zap"
)

type RegisterTenantCommand struct {
	Slug      string
	Name      string
	Email     string
	Password  string
	FirstName string
	LastName  string
}

// RegisterTenantResult holds the result of a registration attempt.
type RegisterTenantResult struct {
	Tenant       *tenant.Tenant             // non-nil when completed synchronously
	Registration *registration.Registration // always set
}

type RegisterTenantCommandHandler interface {
	Handle(ctx context.Context, cmd RegisterTenantCommand) (*RegisterTenantResult, error)
}

type registerTenantHandler struct {
	tenantRepo tenant.Repository
	regRepo    registration.Repository
	idp        tenant.IdentityProvider
	processor  *SagaProcessor
}

func NewRegisterTenantHandler(
	tenantRepo tenant.Repository,
	regRepo registration.Repository,
	idp tenant.IdentityProvider,
	processor *SagaProcessor,
) RegisterTenantCommandHandler {
	return &registerTenantHandler{
		tenantRepo: tenantRepo,
		regRepo:    regRepo,
		idp:        idp,
		processor:  processor,
	}
}

func (h *registerTenantHandler) Handle(ctx context.Context, cmd RegisterTenantCommand) (*RegisterTenantResult, error) {
	log := logger.Get(ctx)

	if err := h.checkSlugAvailability(ctx, cmd.Slug); err != nil {
		return nil, err
	}

	// Create user in Logto immediately (password used and discarded)
	logtoUserID, err := h.idp.CreateUser(ctx, tenant.CreateUserParams{
		Email:     cmd.Email,
		Password:  cmd.Password,
		FirstName: cmd.FirstName,
		LastName:  cmd.LastName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Create registration (user already exists in Logto)
	reg, err := registration.New(cmd.Slug, cmd.Name, cmd.Email, cmd.FirstName, cmd.LastName, logtoUserID)
	if err != nil {
		h.compensateUser(ctx, log, logtoUserID)
		return nil, err
	}

	if err := h.regRepo.Insert(ctx, reg); err != nil {
		h.compensateUser(ctx, log, logtoUserID)
		return nil, fmt.Errorf("failed to create registration: %w", err)
	}

	log.Debug("registration created, attempting inline processing", zap.String("slug", cmd.Slug))

	// Try fast path (inline saga)
	t, err := h.processor.Process(ctx, reg)
	if err != nil {
		// Saga didn't complete inline — worker will pick it up
		log.Warn("registration deferred to worker",
			zap.String("slug", cmd.Slug),
			zap.String("status", string(reg.Status)),
			zap.Error(err))
		return &RegisterTenantResult{Registration: reg}, nil
	}

	return &RegisterTenantResult{Tenant: t, Registration: reg}, nil
}

func (h *registerTenantHandler) checkSlugAvailability(ctx context.Context, slug string) error {
	exists, err := h.tenantRepo.Exists(ctx, slug)
	if err != nil {
		return fmt.Errorf("failed to check tenant existence: %w", err)
	}
	if exists {
		return tenant.ErrSlugAlreadyExists
	}

	regExists, err := h.regRepo.ExistsBySlug(ctx, slug)
	if err != nil {
		return fmt.Errorf("failed to check registration existence: %w", err)
	}
	if regExists {
		return tenant.ErrSlugAlreadyExists
	}

	return nil
}

func (h *registerTenantHandler) compensateUser(ctx context.Context, log *zap.Logger, logtoUserID string) {
	if err := h.idp.DeleteUser(ctx, logtoUserID); err != nil {
		log.Error("failed to delete orphaned Logto user, manual cleanup required",
			zap.String("logtoUserID", logtoUserID),
			zap.Error(err))
	}
}
