package tenant

import (
	"context"
	"fmt"

	"github.com/Sokol111/ecommerce-tenant-service/internal/domain/tenant"
)

type GetTenantListQuery struct {
	Page    int
	Size    int
	Enabled *bool
	Sort    string
	Order   string
}

type ListTenantsResult struct {
	Items []*tenant.Tenant
	Page  int
	Size  int
	Total int64
}

type GetTenantListQueryHandler interface {
	Handle(ctx context.Context, query GetTenantListQuery) (*ListTenantsResult, error)
}

type getTenantListHandler struct {
	repo tenant.Repository
}

func NewGetTenantListHandler(repo tenant.Repository) GetTenantListQueryHandler {
	return &getTenantListHandler{repo: repo}
}

func (h *getTenantListHandler) Handle(ctx context.Context, query GetTenantListQuery) (*ListTenantsResult, error) {
	listQuery := tenant.ListQuery{
		Page:    query.Page,
		Size:    query.Size,
		Enabled: query.Enabled,
		Sort:    query.Sort,
		Order:   query.Order,
	}

	result, err := h.repo.FindList(ctx, listQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenants list: %w", err)
	}

	return &ListTenantsResult{
		Items: result.Items,
		Page:  result.Page,
		Size:  result.Size,
		Total: result.Total,
	}, nil
}
