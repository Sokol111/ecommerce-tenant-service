package connect

import (
	"context"

	"connectrpc.com/connect"
	tenantv1 "github.com/Sokol111/ecommerce-tenant-service-api/gen/go/tenant/v1"
	"github.com/Sokol111/ecommerce-tenant-service/internal/application/tenant"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type tenantHandler struct {
	createTenant    tenant.CreateTenantHandler
	updateTenant    tenant.UpdateTenantHandler
	deleteTenant    tenant.DeleteTenantHandler
	getBySlug       tenant.GetBySlugHandler
	getList         tenant.GetListHandler
	getEnabledSlugs tenant.GetEnabledSlugsHandler
}

// ==================== CRUD ====================

func (h *tenantHandler) CreateTenant(ctx context.Context, req *connect.Request[tenantv1.CreateTenantRequest]) (*connect.Response[tenantv1.CreateTenantResponse], error) {
	cmd := tenant.CreateCommand{
		Slug: req.Msg.Slug,
		Name: req.Msg.Name,
	}

	created, err := h.createTenant.Handle(ctx, cmd)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&tenantv1.CreateTenantResponse{Tenant: toProtoTenant(created)}), nil
}

func (h *tenantHandler) UpdateTenant(ctx context.Context, req *connect.Request[tenantv1.UpdateTenantRequest]) (*connect.Response[tenantv1.UpdateTenantResponse], error) {
	cmd := tenant.UpdateCommand{
		Slug:    req.Msg.Slug,
		Version: req.Msg.Version,
		Name:    req.Msg.Name,
		Enabled: req.Msg.Enabled,
	}

	updated, err := h.updateTenant.Handle(ctx, cmd)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&tenantv1.UpdateTenantResponse{Tenant: toProtoTenant(updated)}), nil
}

func (h *tenantHandler) GetTenantBySlug(ctx context.Context, req *connect.Request[tenantv1.GetTenantBySlugRequest]) (*connect.Response[tenantv1.GetTenantBySlugResponse], error) {
	q := tenant.GetBySlugQuery{Slug: req.Msg.Slug}

	found, err := h.getBySlug.Handle(ctx, q)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&tenantv1.GetTenantBySlugResponse{Tenant: toProtoTenant(found)}), nil
}

func (h *tenantHandler) DeleteTenant(ctx context.Context, req *connect.Request[tenantv1.DeleteTenantRequest]) (*connect.Response[tenantv1.DeleteTenantResponse], error) {
	cmd := tenant.DeleteCommand{Slug: req.Msg.Slug}

	if err := h.deleteTenant.Handle(ctx, cmd); err != nil {
		return nil, err
	}

	return connect.NewResponse(&tenantv1.DeleteTenantResponse{}), nil
}

func (h *tenantHandler) GetTenantList(ctx context.Context, req *connect.Request[tenantv1.GetTenantListRequest]) (*connect.Response[tenantv1.GetTenantListResponse], error) {
	var enabled *bool
	if req.Msg.Enabled != nil {
		e := *req.Msg.Enabled
		enabled = &e
	}

	var sort, order string
	if req.Msg.Sort != nil {
		sort = *req.Msg.Sort
	}
	if req.Msg.Order != nil {
		switch *req.Msg.Order {
		case tenantv1.SortOrder_SORT_ORDER_DESC:
			order = "desc"
		default:
			order = "asc"
		}
	}

	q := tenant.GetListQuery{
		Page:    int(req.Msg.Page),
		Size:    int(req.Msg.Size),
		Enabled: enabled,
		Sort:    sort,
		Order:   order,
	}

	result, err := h.getList.Handle(ctx, q)
	if err != nil {
		return nil, err
	}

	resp := &tenantv1.GetTenantListResponse{
		Page:  int32(result.Page), //nolint:gosec // Page originates from int32 proto field, cannot overflow
		Size:  int32(result.Size), //nolint:gosec // Size originates from int32 proto field, cannot overflow
		Total: result.Total,
	}
	for _, t := range result.Items {
		resp.Items = append(resp.Items, toProtoTenant(t))
	}

	return connect.NewResponse(resp), nil
}

func (h *tenantHandler) GetEnabledTenantSlugs(ctx context.Context, _ *connect.Request[tenantv1.GetEnabledTenantSlugsRequest]) (*connect.Response[tenantv1.GetEnabledTenantSlugsResponse], error) {
	slugs, err := h.getEnabledSlugs.Handle(ctx)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&tenantv1.GetEnabledTenantSlugsResponse{Slugs: slugs}), nil
}

// ==================== Helpers ====================

func toProtoTenant(t *tenant.Tenant) *tenantv1.Tenant {
	return &tenantv1.Tenant{
		Id:         t.ID,
		Slug:       t.Slug,
		Name:       t.Name,
		Enabled:    t.Enabled,
		Version:    int64(t.Version),
		CreatedAt:  timestamppb.New(t.CreatedAt),
		ModifiedAt: timestamppb.New(t.ModifiedAt),
	}
}
