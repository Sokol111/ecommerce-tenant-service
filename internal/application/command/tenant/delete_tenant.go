package tenant

import (
	"context"
	"fmt"

	"github.com/Sokol111/ecommerce-commons/pkg/core/logger"
	"github.com/Sokol111/ecommerce-commons/pkg/messaging/patterns/outbox"
	"github.com/Sokol111/ecommerce-commons/pkg/persistence/mongo"
	"github.com/Sokol111/ecommerce-tenant-service/internal/domain/tenant"
	"github.com/Sokol111/ecommerce-tenant-service/internal/event"
	"go.uber.org/zap"
)

type DeleteTenantCommand struct {
	Slug string
}

type DeleteTenantCommandHandler interface {
	Handle(ctx context.Context, cmd DeleteTenantCommand) error
}

type deleteTenantHandler struct {
	repo         tenant.Repository
	outbox       outbox.Outbox
	txManager    mongo.TxManager
	eventFactory event.TenantEventFactory
}

func NewDeleteTenantHandler(
	repo tenant.Repository,
	outbox outbox.Outbox,
	txManager mongo.TxManager,
	eventFactory event.TenantEventFactory,
) DeleteTenantCommandHandler {
	return &deleteTenantHandler{
		repo:         repo,
		outbox:       outbox,
		txManager:    txManager,
		eventFactory: eventFactory,
	}
}

func (h *deleteTenantHandler) Handle(ctx context.Context, cmd DeleteTenantCommand) error {
	t, err := h.repo.FindBySlug(ctx, cmd.Slug)
	if err != nil {
		return fmt.Errorf("failed to get tenant: %w", err)
	}

	if t.Enabled {
		return tenant.ErrTenantNotDisabled
	}

	msg := h.eventFactory.NewTenantDeletedOutboxMessage(ctx, t.Slug)

	send, err := mongo.WithTransaction(ctx, h.txManager, func(txCtx context.Context) (outbox.SendFunc, error) {
		if err := h.repo.Delete(txCtx, t.ID); err != nil {
			return nil, fmt.Errorf("failed to delete tenant: %w", err)
		}

		send, err := h.outbox.Create(txCtx, msg)
		if err != nil {
			return nil, fmt.Errorf("failed to create outbox: %w", err)
		}

		return send, nil
	})
	if err != nil {
		return err
	}

	logger.Get(ctx).Debug("tenant deleted", zap.String("slug", cmd.Slug))

	_ = send(ctx) //nolint:errcheck // best-effort send

	return nil
}
