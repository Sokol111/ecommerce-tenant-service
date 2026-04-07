package tenant

import (
	"context"
	"errors"
	"fmt"

	"github.com/Sokol111/ecommerce-commons/pkg/core/logger"
	"github.com/Sokol111/ecommerce-commons/pkg/messaging/patterns/outbox"
	"github.com/Sokol111/ecommerce-commons/pkg/persistence/mongo"
	"github.com/Sokol111/ecommerce-tenant-service/internal/domain/tenant"
	"github.com/Sokol111/ecommerce-tenant-service/internal/event"
	"go.uber.org/zap"
)

type UpdateTenantCommand struct {
	Slug    string
	Version int
	Name    string
	Enabled bool
}

type UpdateTenantCommandHandler interface {
	Handle(ctx context.Context, cmd UpdateTenantCommand) (*tenant.Tenant, error)
}

type updateTenantHandler struct {
	repo         tenant.Repository
	outbox       outbox.Outbox
	txManager    mongo.TxManager
	eventFactory event.TenantEventFactory
}

func NewUpdateTenantHandler(
	repo tenant.Repository,
	outbox outbox.Outbox,
	txManager mongo.TxManager,
	eventFactory event.TenantEventFactory,
) UpdateTenantCommandHandler {
	return &updateTenantHandler{
		repo:         repo,
		outbox:       outbox,
		txManager:    txManager,
		eventFactory: eventFactory,
	}
}

func (h *updateTenantHandler) Handle(ctx context.Context, cmd UpdateTenantCommand) (*tenant.Tenant, error) {
	t, err := h.repo.FindBySlug(ctx, cmd.Slug)
	if err != nil {
		if errors.Is(err, mongo.ErrEntityNotFound) {
			return nil, mongo.ErrEntityNotFound
		}
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	if t.Version != cmd.Version {
		return nil, mongo.ErrOptimisticLocking
	}

	if err := t.Update(cmd.Name, cmd.Enabled); err != nil {
		return nil, err
	}

	msg := h.eventFactory.NewTenantUpdatedOutboxMessage(ctx, t)

	type updateResult struct {
		Tenant *tenant.Tenant
		Send   outbox.SendFunc
	}

	res, err := mongo.WithTransaction(ctx, h.txManager, func(txCtx context.Context) (*updateResult, error) {
		updated, err := h.repo.Update(txCtx, t)
		if err != nil {
			return nil, fmt.Errorf("failed to update tenant: %w", err)
		}

		send, err := h.outbox.Create(txCtx, msg)
		if err != nil {
			return nil, fmt.Errorf("failed to create outbox: %w", err)
		}

		return &updateResult{Tenant: updated, Send: send}, nil
	})
	if err != nil {
		return nil, err
	}

	logger.Get(ctx).Debug("tenant updated", zap.String("slug", res.Tenant.Slug))

	_ = res.Send(ctx) //nolint:errcheck // best-effort send

	return res.Tenant, nil
}
