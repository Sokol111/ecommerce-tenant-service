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

func (s *tenantService) Delete(ctx context.Context, cmd DeleteCommand) error {
	t, err := s.repo.FindBySlug(ctx, cmd.Slug)
	if err != nil {
		return fmt.Errorf("failed to get tenant: %w", err)
	}

	if t.Enabled {
		return ErrTenantNotDisabled
	}

	msg := s.eventFactory.NewTenantDeletedOutboxMessage(ctx, t.Slug)

	send, err := mongo.WithTransaction(ctx, s.txManager, func(txCtx context.Context) (outbox.SendFunc, error) {
		err = s.repo.Delete(txCtx, t.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to delete tenant: %w", err)
		}

		createdSend, createOutboxErr := s.outbox.Create(txCtx, msg)
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

	return nil
}
