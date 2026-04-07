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

type CreateTenantCommand struct {
	Slug string
	Name string
}

type CreateTenantCommandHandler interface {
	Handle(ctx context.Context, cmd CreateTenantCommand) (*tenant.Tenant, error)
}

type createTenantHandler struct {
	repo         tenant.Repository
	outbox       outbox.Outbox
	txManager    mongo.TxManager
	eventFactory event.TenantEventFactory
}

func NewCreateTenantHandler(
	repo tenant.Repository,
	outbox outbox.Outbox,
	txManager mongo.TxManager,
	eventFactory event.TenantEventFactory,
) CreateTenantCommandHandler {
	return &createTenantHandler{
		repo:         repo,
		outbox:       outbox,
		txManager:    txManager,
		eventFactory: eventFactory,
	}
}

func (h *createTenantHandler) Handle(ctx context.Context, cmd CreateTenantCommand) (*tenant.Tenant, error) {
	exists, err := h.repo.Exists(ctx, cmd.Slug)
	if err != nil {
		return nil, fmt.Errorf("failed to check tenant existence: %w", err)
	}
	if exists {
		return nil, tenant.ErrSlugAlreadyExists
	}

	t, err := tenant.NewTenant(cmd.Slug, cmd.Name)
	if err != nil {
		return nil, err
	}

	msg := h.eventFactory.NewTenantUpdatedOutboxMessage(ctx, t)

	type createResult struct {
		Tenant *tenant.Tenant
		Send   outbox.SendFunc
	}

	res, err := mongo.WithTransaction(ctx, h.txManager, func(txCtx context.Context) (*createResult, error) {
		err = h.repo.Insert(txCtx, t)
		if err != nil {
			return nil, fmt.Errorf("failed to insert tenant: %w", err)
		}

		send, createOutboxErr := h.outbox.Create(txCtx, msg)
		if createOutboxErr != nil {
			return nil, fmt.Errorf("failed to create outbox: %w", createOutboxErr)
		}

		return &createResult{Tenant: t, Send: send}, nil
	})
	if err != nil {
		return nil, err
	}

	logger.Get(ctx).Debug("tenant created", zap.String("slug", res.Tenant.Slug))

	_ = res.Send(ctx) //nolint:errcheck // best-effort send

	return res.Tenant, nil
}
