package tenant

import (
	"context"
	"fmt"

	"github.com/Sokol111/ecommerce-commons/pkg/core/logger"
	"github.com/Sokol111/ecommerce-commons/pkg/messaging/patterns/outbox"
	"github.com/Sokol111/ecommerce-commons/pkg/persistence/mongo"
	"go.uber.org/zap"
)

type CreateCommand struct {
	Slug string
	Name string
}

func (s *tenantService) Create(ctx context.Context, cmd CreateCommand) (*Tenant, error) {
	exists, err := s.repo.Exists(ctx, cmd.Slug)
	if err != nil {
		return nil, fmt.Errorf("failed to check tenant existence: %w", err)
	}
	if exists {
		return nil, ErrSlugAlreadyExists
	}

	t, err := NewTenant(cmd.Slug, cmd.Name)
	if err != nil {
		return nil, err
	}

	msg := s.eventFactory.NewTenantUpdatedOutboxMessage(ctx, t)

	type createResult struct {
		Tenant *Tenant
		Send   outbox.SendFunc
	}

	res, err := mongo.WithTransaction(ctx, s.txManager, func(txCtx context.Context) (*createResult, error) {
		err = s.repo.Insert(txCtx, t)
		if err != nil {
			return nil, fmt.Errorf("failed to insert tenant: %w", err)
		}

		send, createOutboxErr := s.outbox.Create(txCtx, msg)
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
