package registration

import (
	"context"

	"github.com/Sokol111/ecommerce-tenant-service/internal/application/tenant"
)

// RegistrationService provides tenant registration and status query operations.
type RegistrationService interface {
	Register(ctx context.Context, cmd RegisterCommand) (*RegisterResult, error)
	GetStatus(ctx context.Context, query GetStatusQuery) (*Registration, error)
}

type registrationService struct {
	tenantRepo tenant.Repository
	regRepo    Repository
	idp        tenant.IdentityProvider
	processor  *Processor
}

func NewRegistrationService(
	tenantRepo tenant.Repository,
	regRepo Repository,
	idp tenant.IdentityProvider,
	processor *Processor,
) RegistrationService {
	return &registrationService{
		tenantRepo: tenantRepo,
		regRepo:    regRepo,
		idp:        idp,
		processor:  processor,
	}
}
