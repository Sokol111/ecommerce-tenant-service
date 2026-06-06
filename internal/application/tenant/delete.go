package tenant

import (
	"context"
	"fmt"

	"github.com/Sokol111/ecommerce-commons/pkg/core/logger"
	"github.com/Sokol111/ecommerce-commons/pkg/messaging/patterns/outbox"
	"github.com/Sokol111/ecommerce-commons/pkg/persistence/mongo"
	"go.uber.org/zap"
)

type DeleteCommand struct {
	Slug string
}

type DeleteTenantHandler interface {
	Handle(ctx context.Context, cmd DeleteCommand) error
}

type deleteTenantHandler struct {
	repo         Repository
	outbox       outbox.Outbox
	txManager    mongo.TxManager
	eventFactory TenantEventFactory
	idp          IdentityProvider
}

func NewDeleteTenantHandler(
	repo Repository,
	outbox outbox.Outbox,
	txManager mongo.TxManager,
	eventFactory TenantEventFactory,
	idp IdentityProvider,
) DeleteTenantHandler {
	return &deleteTenantHandler{
		repo:         repo,
		outbox:       outbox,
		txManager:    txManager,
		eventFactory: eventFactory,
		idp:          idp,
	}
}

func (h *deleteTenantHandler) Handle(ctx context.Context, cmd DeleteCommand) error {
	t, err := h.repo.FindBySlug(ctx, cmd.Slug)
	if err != nil {
		return fmt.Errorf("failed to get tenant: %w", err)
	}

	if t.Enabled {
		return ErrTenantNotDisabled
	}

	msg := h.eventFactory.NewTenantDeletedOutboxMessage(ctx, t.Slug)

	send, err := mongo.WithTransaction(ctx, h.txManager, func(txCtx context.Context) (outbox.SendFunc, error) {
		err = h.repo.Delete(txCtx, t.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to delete tenant: %w", err)
		}

		createdSend, createOutboxErr := h.outbox.Create(txCtx, msg)
		if createOutboxErr != nil {
			return nil, fmt.Errorf("failed to create outbox: %w", createOutboxErr)
		}

		return createdSend, nil
	})
	if err != nil {
		return err
	}

	logger.Get(ctx).Debug("tenant deleted", zap.String("slug", cmd.Slug))

	_ = send(ctx) //nolint:errcheck // best-effort send

	if err := h.idp.DeleteUser(ctx, t.OwnerUserID); err != nil {
		logger.Get(ctx).Warn("failed to delete owner user from identity provider",
			zap.String("slug", cmd.Slug), zap.String("userID", t.OwnerUserID), zap.Error(err))
	}

	return nil
}
