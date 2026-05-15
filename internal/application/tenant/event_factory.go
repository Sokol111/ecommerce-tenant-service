package tenant

import (
	"context"

	"github.com/Sokol111/ecommerce-commons/pkg/messaging/patterns/outbox"
)

// TenantEventFactory defines the port for creating tenant event outbox messages.
type TenantEventFactory interface {
	NewTenantUpdatedOutboxMessage(ctx context.Context, t *Tenant) outbox.Message
	NewTenantDeletedOutboxMessage(ctx context.Context, slug string) outbox.Message
}
