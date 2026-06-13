package http //nolint:revive // package name intentional

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/samber/lo"

	"github.com/Sokol111/ecommerce-commons/pkg/persistence/mongo"
	"github.com/Sokol111/ecommerce-tenant-service-api/gen/httpapi"
	"github.com/Sokol111/ecommerce-tenant-service/internal/application/tenant"
)

func toTenantResponse(t *tenant.Tenant) *httpapi.TenantResponse {
	return &httpapi.TenantResponse{
		ID:         uuid.MustParse(t.ID),
		Slug:       t.Slug,
		Name:       t.Name,
		Enabled:    t.Enabled,
		Version:    t.Version,
		CreatedAt:  t.CreatedAt,
		ModifiedAt: t.ModifiedAt,
	}
}

func (h *tenantHandler) CreateTenant(ctx context.Context, req *httpapi.CreateTenantRequest) (httpapi.CreateTenantRes, error) {
	cmd := tenant.CreateCommand{
		Slug: req.Slug,
		Name: req.Name,
	}

	created, err := h.createTenant.Handle(ctx, cmd)
	if err != nil {
		if errors.Is(err, tenant.ErrInvalidTenantData) {
			return &httpapi.CreateTenantBadRequest{
				Status: 400,
				Type:   *aboutBlankURL,
				Title:  err.Error(),
			}, nil
		}
		if errors.Is(err, tenant.ErrSlugAlreadyExists) {
			return &httpapi.CreateTenantConflict{
				Status: 409,
				Type:   *aboutBlankURL,
				Title:  "Tenant with this slug already exists",
			}, nil
		}
		return nil, err
	}

	return toTenantResponse(created), nil
}

func (h *tenantHandler) UpdateTenant(ctx context.Context, req *httpapi.UpdateTenantRequest) (httpapi.UpdateTenantRes, error) {
	cmd := tenant.UpdateCommand{
		Slug:    req.Slug,
		Version: req.Version,
		Name:    req.Name,
		Enabled: req.Enabled,
	}

	updated, err := h.updateTenant.Handle(ctx, cmd)
	if err != nil {
		if errors.Is(err, tenant.ErrInvalidTenantData) {
			return &httpapi.UpdateTenantBadRequest{
				Status: 400,
				Type:   *aboutBlankURL,
				Title:  err.Error(),
			}, nil
		}
		if errors.Is(err, mongo.ErrEntityNotFound) {
			return &httpapi.UpdateTenantNotFound{
				Status: 404,
				Type:   *aboutBlankURL,
				Title:  "Tenant not found",
			}, nil
		}
		if errors.Is(err, mongo.ErrOptimisticLocking) {
			return &httpapi.UpdateTenantPreconditionFailed{
				Status: 412,
				Type:   *aboutBlankURL,
				Title:  "Version mismatch",
			}, nil
		}
		return nil, err
	}

	return toTenantResponse(updated), nil
}

func (h *tenantHandler) GetTenantBySlug(ctx context.Context, params httpapi.GetTenantBySlugParams) (httpapi.GetTenantBySlugRes, error) {
	q := tenant.GetBySlugQuery{Slug: params.Slug}

	found, err := h.getBySlug.Handle(ctx, q)
	if errors.Is(err, mongo.ErrEntityNotFound) {
		return &httpapi.GetTenantBySlugNotFound{
			Status: 404,
			Type:   *aboutBlankURL,
			Title:  "Tenant not found",
		}, nil
	}
	if err != nil {
		return nil, err
	}

	return toTenantResponse(found), nil
}

func (h *tenantHandler) DeleteTenant(ctx context.Context, params httpapi.DeleteTenantParams) (httpapi.DeleteTenantRes, error) {
	cmd := tenant.DeleteCommand{Slug: params.Slug}

	err := h.deleteTenant.Handle(ctx, cmd)
	if errors.Is(err, mongo.ErrEntityNotFound) {
		return &httpapi.DeleteTenantNotFound{
			Status: 404,
			Type:   *aboutBlankURL,
			Title:  "Tenant not found",
		}, nil
	}
	if err != nil {
		return nil, err
	}

	return &httpapi.DeleteTenantNoContent{}, nil
}

func (h *tenantHandler) GetTenantList(ctx context.Context, params httpapi.GetTenantListParams) (httpapi.GetTenantListRes, error) {
	var enabled *bool
	if params.Enabled.IsSet() {
		e := params.Enabled.Value
		enabled = &e
	}

	q := tenant.GetListQuery{
		Page:    params.Page,
		Size:    params.Size,
		Enabled: enabled,
		Sort:    string(params.Sort.Or(httpapi.GetTenantListSortCreatedAt)),
		Order:   string(params.Order.Or(httpapi.GetTenantListOrderDesc)),
	}

	result, err := h.getList.Handle(ctx, q)
	if err != nil {
		return nil, err
	}

	return &httpapi.TenantListResponse{
		Items: lo.Map(result.Items, func(t *tenant.Tenant, _ int) httpapi.TenantResponse {
			return *toTenantResponse(t)
		}),
		Page:  result.Page,
		Size:  result.Size,
		Total: int(result.Total),
	}, nil
}

func (h *tenantHandler) GetEnabledTenantSlugs(ctx context.Context) (httpapi.GetEnabledTenantSlugsRes, error) {
	slugs, err := h.getEnabledSlugs.Handle(ctx)
	if err != nil {
		return nil, err
	}

	return &httpapi.TenantSlugListResponse{
		Slugs: slugs,
	}, nil
}
