package tenant

import (
	"context"
	"errors"
	"fmt"

	"github.com/Sokol111/ecommerce-commons/pkg/core/logger"
	"github.com/Sokol111/ecommerce-commons/pkg/messaging/patterns/outbox"
	"github.com/Sokol111/ecommerce-commons/pkg/persistence/mongo"
	"go.uber.org/zap"
)

type UpdateCommand struct {
	Slug    string
	Version int64
	Name    string
	Enabled bool
}

type UpdateTenantHandler interface {
	Handle(ctx context.Context, cmd UpdateCommand) (*Tenant, error)
}

type updateTenantHandler struct {
	repo         Repository
	outbox       outbox.Outbox
	txManager    mongo.TxManager
	eventFactory TenantEventFactory
}

func NewUpdateTenantHandler(
	repo Repository,
	outbox outbox.Outbox,
	txManager mongo.TxManager,
	eventFactory TenantEventFactory,
) UpdateTenantHandler {
	return &updateTenantHandler{
		repo:         repo,
		outbox:       outbox,
		txManager:    txManager,
		eventFactory: eventFactory,
	}
}

func (h *updateTenantHandler) Handle(ctx context.Context, cmd UpdateCommand) (*Tenant, error) {
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

	err = t.Update(cmd.Name, cmd.Enabled)
	if err != nil {
		return nil, err
	}

	msg := h.eventFactory.NewTenantUpdatedOutboxMessage(ctx, t)

	type updateResult struct {
		Tenant *Tenant
		Send   outbox.SendFunc
	}

	res, err := mongo.WithTransaction(ctx, h.txManager, func(txCtx context.Context) (*updateResult, error) {
		updated, updateErr := h.repo.Update(txCtx, t)
		if updateErr != nil {
			return nil, fmt.Errorf("failed to update tenant: %w", updateErr)
		}

		send, createOutboxErr := h.outbox.Create(txCtx, msg)
		if createOutboxErr != nil {
			return nil, fmt.Errorf("failed to create outbox: %w", createOutboxErr)
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
