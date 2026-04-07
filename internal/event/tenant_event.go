package event

import (
	"context"

	"github.com/Sokol111/ecommerce-commons/pkg/messaging/patterns/outbox"
	"github.com/Sokol111/ecommerce-tenant-service-api/gen/events"
	"github.com/Sokol111/ecommerce-tenant-service/internal/domain/tenant"
)

type TenantEventFactory interface {
	NewTenantUpdatedOutboxMessage(ctx context.Context, t *tenant.Tenant) outbox.Message
	NewTenantDeletedOutboxMessage(ctx context.Context, slug string) outbox.Message
}

type tenantEventFactory struct{}

func newTenantEventFactory() TenantEventFactory {
	return &tenantEventFactory{}
}

func (f *tenantEventFactory) NewTenantUpdatedOutboxMessage(_ context.Context, t *tenant.Tenant) outbox.Message {
	return outbox.Message{
		Event: &events.TenantUpdatedEvent{
			Payload: events.TenantUpdatedPayload{
				ID:         t.ID,
				Slug:       t.Slug,
				Name:       t.Name,
				Enabled:    t.Enabled,
				Version:    t.Version,
				CreatedAt:  t.CreatedAt,
				ModifiedAt: t.ModifiedAt,
			},
		},
		Key: t.Slug,
	}
}

func (f *tenantEventFactory) NewTenantDeletedOutboxMessage(_ context.Context, slug string) outbox.Message {
	return outbox.Message{
		Event: &events.TenantDeletedEvent{
			Payload: events.TenantDeletedPayload{
				Slug: slug,
			},
		},
		Key: slug,
	}
}
