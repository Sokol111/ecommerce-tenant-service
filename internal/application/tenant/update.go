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
	Version int
	Name    string
	Enabled bool
}

func (s *tenantService) Update(ctx context.Context, cmd UpdateCommand) (*Tenant, error) {
	t, err := s.repo.FindBySlug(ctx, cmd.Slug)
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

	msg := s.eventFactory.NewTenantUpdatedOutboxMessage(ctx, t)

	type updateResult struct {
		Tenant *Tenant
		Send   outbox.SendFunc
	}

	res, err := mongo.WithTransaction(ctx, s.txManager, func(txCtx context.Context) (*updateResult, error) {
		updated, updateErr := s.repo.Update(txCtx, t)
		if updateErr != nil {
			return nil, fmt.Errorf("failed to update tenant: %w", updateErr)
		}

		send, createOutboxErr := s.outbox.Create(txCtx, msg)
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
