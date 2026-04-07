package tenant

import (
	"context"
	"fmt"

	"github.com/Sokol111/ecommerce-tenant-service/internal/domain/tenant"
)

type GetTenantBySlugQuery struct {
	Slug string
}

type GetTenantBySlugQueryHandler interface {
	Handle(ctx context.Context, query GetTenantBySlugQuery) (*tenant.Tenant, error)
}

type getTenantBySlugHandler struct {
	repo tenant.Repository
}

func NewGetTenantBySlugHandler(repo tenant.Repository) GetTenantBySlugQueryHandler {
	return &getTenantBySlugHandler{repo: repo}
}

func (h *getTenantBySlugHandler) Handle(ctx context.Context, query GetTenantBySlugQuery) (*tenant.Tenant, error) {
	t, err := h.repo.FindBySlug(ctx, query.Slug)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}
	return t, nil
}
