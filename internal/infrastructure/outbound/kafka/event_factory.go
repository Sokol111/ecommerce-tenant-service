package kafka

import (
	"context"

	"github.com/Sokol111/ecommerce-commons/pkg/messaging/patterns/outbox"
	eventsv1 "github.com/Sokol111/ecommerce-tenant-service-api/gen/events/tenant/v1"
	"github.com/Sokol111/ecommerce-tenant-service-api/pkg/events"
	"github.com/Sokol111/ecommerce-tenant-service/internal/application/tenant"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type tenantEventFactory struct{}

func newTenantEventFactory() tenant.TenantEventFactory {
	return &tenantEventFactory{}
}

func (f *tenantEventFactory) NewTenantUpdatedOutboxMessage(_ context.Context, t *tenant.Tenant) outbox.Message {
	return outbox.Message{
		Event: &eventsv1.TenantUpdatedEvent{
			Id:         t.ID,
			Slug:       t.Slug,
			Name:       t.Name,
			Enabled:    t.Enabled,
			Version:    int64(t.Version),
			CreatedAt:  timestamppb.New(t.CreatedAt),
			ModifiedAt: timestamppb.New(t.ModifiedAt),
		},
		Key: t.Slug,
	}
}

func (f *tenantEventFactory) NewTenantDeletedOutboxMessage(_ context.Context, slug string) outbox.Message {
	return outbox.Message{
		Event: &eventsv1.TenantDeletedEvent{
			Slug: slug,
		},
		Topic: events.TopicFor(&eventsv1.TenantDeletedEvent{}),
		Key:   slug,
	}
}
