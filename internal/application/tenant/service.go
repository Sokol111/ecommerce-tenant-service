package tenant

import (
	"context"

	"github.com/Sokol111/ecommerce-commons/pkg/messaging/patterns/outbox"
	"github.com/Sokol111/ecommerce-commons/pkg/persistence/mongo"
)

// TenantService provides all tenant CRUD and query operations.
type TenantService interface {
	Create(ctx context.Context, cmd CreateCommand) (*Tenant, error)
	Update(ctx context.Context, cmd UpdateCommand) (*Tenant, error)
	Delete(ctx context.Context, cmd DeleteCommand) error
	GetBySlug(ctx context.Context, query GetBySlugQuery) (*Tenant, error)
	GetList(ctx context.Context, query GetListQuery) (*ListResult, error)
	GetEnabledSlugs(ctx context.Context) ([]string, error)
}

type tenantService struct {
	repo         Repository
	outbox       outbox.Outbox
	txManager    mongo.TxManager
	eventFactory TenantEventFactory
}

func NewTenantService(
	repo Repository,
	outbox outbox.Outbox,
	txManager mongo.TxManager,
	eventFactory TenantEventFactory,
) TenantService {
	return &tenantService{
		repo:         repo,
		outbox:       outbox,
		txManager:    txManager,
		eventFactory: eventFactory,
	}
}
