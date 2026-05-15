package registration

import (
	"context"
	"fmt"

	"github.com/Sokol111/ecommerce-commons/pkg/core/logger"
	"github.com/Sokol111/ecommerce-tenant-service/internal/application/tenant"
	"go.uber.org/zap"
)

type RegisterCommand struct {
	Slug      string
	Name      string
	Email     string
	Password  string
	FirstName string
	LastName  string
}

type RegisterResult struct {
	Tenant       *tenant.Tenant
	Registration *Registration
}

func (s *registrationService) Register(ctx context.Context, cmd RegisterCommand) (*RegisterResult, error) {
	log := logger.Get(ctx)

	if err := s.checkSlugAvailability(ctx, cmd.Slug); err != nil {
		return nil, err
	}

	// Create user in Logto immediately (password used and discarded)
	logtoUserID, err := s.idp.CreateUser(ctx, tenant.CreateUserParams{
		Email:     cmd.Email,
		Password:  cmd.Password,
		FirstName: cmd.FirstName,
		LastName:  cmd.LastName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Create registration (user already exists in Logto)
	reg, err := New(cmd.Slug, cmd.Name, cmd.Email, cmd.FirstName, cmd.LastName, logtoUserID)
	if err != nil {
		s.compensateUser(ctx, log, logtoUserID)
		return nil, err
	}

	if err := s.regRepo.Insert(ctx, reg); err != nil {
		s.compensateUser(ctx, log, logtoUserID)
		return nil, fmt.Errorf("failed to create registration: %w", err)
	}

	log.Debug("registration created, attempting inline processing", zap.String("slug", cmd.Slug))

	// Try fast path (inline saga)
	t, err := s.processor.Process(ctx, reg)
	if err != nil {
		// Saga didn't complete inline — worker will pick it up
		log.Warn("registration deferred to worker",
			zap.String("slug", cmd.Slug),
			zap.String("status", string(reg.Status)),
			zap.Error(err))
		return &RegisterResult{Registration: reg}, nil
	}

	return &RegisterResult{Tenant: t, Registration: reg}, nil
}

func (s *registrationService) checkSlugAvailability(ctx context.Context, slug string) error {
	exists, err := s.tenantRepo.Exists(ctx, slug)
	if err != nil {
		return fmt.Errorf("failed to check tenant existence: %w", err)
	}
	if exists {
		return tenant.ErrSlugAlreadyExists
	}

	regExists, err := s.regRepo.ExistsBySlug(ctx, slug)
	if err != nil {
		return fmt.Errorf("failed to check registration existence: %w", err)
	}
	if regExists {
		return tenant.ErrSlugAlreadyExists
	}

	return nil
}

func (s *registrationService) compensateUser(ctx context.Context, log *zap.Logger, logtoUserID string) {
	if err := s.idp.DeleteUser(ctx, logtoUserID); err != nil {
		log.Error("failed to delete orphaned Logto user, manual cleanup required",
			zap.String("logtoUserID", logtoUserID),
			zap.Error(err))
	}
}
