package connect

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	"github.com/Sokol111/ecommerce-commons/pkg/persistence/mongo"
	tenantv1 "github.com/Sokol111/ecommerce-tenant-service-api/gen/connect/tenant/v1"
	"github.com/Sokol111/ecommerce-tenant-service/internal/application/registration"
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
	register        registration.RegisterHandler
	getStatus       registration.GetStatusHandler
}

// ==================== CRUD ====================

func (h *tenantHandler) CreateTenant(ctx context.Context, req *connect.Request[tenantv1.CreateTenantRequest]) (*connect.Response[tenantv1.CreateTenantResponse], error) {
	cmd := tenant.CreateCommand{
		Slug: req.Msg.Slug,
		Name: req.Msg.Name,
	}

	created, err := h.createTenant.Handle(ctx, cmd)
	if err != nil {
		return nil, mapConnectError(err)
	}

	return connect.NewResponse(&tenantv1.CreateTenantResponse{Tenant: toProtoTenant(created)}), nil
}

func (h *tenantHandler) UpdateTenant(ctx context.Context, req *connect.Request[tenantv1.UpdateTenantRequest]) (*connect.Response[tenantv1.UpdateTenantResponse], error) {
	cmd := tenant.UpdateCommand{
		Slug:    req.Msg.Slug,
		Version: int(req.Msg.Version),
		Name:    req.Msg.Name,
		Enabled: req.Msg.Enabled,
	}

	updated, err := h.updateTenant.Handle(ctx, cmd)
	if err != nil {
		return nil, mapConnectError(err)
	}

	return connect.NewResponse(&tenantv1.UpdateTenantResponse{Tenant: toProtoTenant(updated)}), nil
}

func (h *tenantHandler) GetTenantBySlug(ctx context.Context, req *connect.Request[tenantv1.GetTenantBySlugRequest]) (*connect.Response[tenantv1.GetTenantBySlugResponse], error) {
	q := tenant.GetBySlugQuery{Slug: req.Msg.Slug}

	found, err := h.getBySlug.Handle(ctx, q)
	if err != nil {
		return nil, mapConnectError(err)
	}

	return connect.NewResponse(&tenantv1.GetTenantBySlugResponse{Tenant: toProtoTenant(found)}), nil
}

func (h *tenantHandler) DeleteTenant(ctx context.Context, req *connect.Request[tenantv1.DeleteTenantRequest]) (*connect.Response[tenantv1.DeleteTenantResponse], error) {
	cmd := tenant.DeleteCommand{Slug: req.Msg.Slug}

	if err := h.deleteTenant.Handle(ctx, cmd); err != nil {
		return nil, mapConnectError(err)
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
		order = *req.Msg.Order
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
		return nil, mapConnectError(err)
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
		return nil, mapConnectError(err)
	}

	return connect.NewResponse(&tenantv1.GetEnabledTenantSlugsResponse{Slugs: slugs}), nil
}

// ==================== Registration ====================

func (h *tenantHandler) RegisterTenant(ctx context.Context, req *connect.Request[tenantv1.RegisterTenantRequest]) (*connect.Response[tenantv1.RegisterTenantResponse], error) {
	cmd := registration.RegisterCommand{
		Slug:      req.Msg.Slug,
		Name:      req.Msg.Name,
		Email:     req.Msg.Email,
		Password:  req.Msg.Password,
		FirstName: req.Msg.FirstName,
		LastName:  req.Msg.LastName,
	}

	result, err := h.register.Handle(ctx, cmd)
	if err != nil {
		return nil, mapConnectError(err)
	}

	// Completed synchronously — return tenant.
	if result.Tenant != nil {
		return connect.NewResponse(&tenantv1.RegisterTenantResponse{
			Result: &tenantv1.RegisterTenantResponse_Tenant{
				Tenant: toProtoTenant(result.Tenant),
			},
		}), nil
	}

	// Deferred to worker — return status.
	return connect.NewResponse(&tenantv1.RegisterTenantResponse{
		Result: &tenantv1.RegisterTenantResponse_Status{
			Status: toProtoRegistrationStatus(result.Registration),
		},
	}), nil
}

func (h *tenantHandler) GetRegistrationStatus(ctx context.Context, req *connect.Request[tenantv1.GetRegistrationStatusRequest]) (*connect.Response[tenantv1.GetRegistrationStatusResponse], error) {
	q := registration.GetStatusQuery{Slug: req.Msg.Slug}

	reg, err := h.getStatus.Handle(ctx, q)
	if err != nil {
		return nil, mapConnectError(err)
	}

	return connect.NewResponse(toGetRegistrationStatusResponse(reg)), nil
}

// ==================== Helpers ====================

func toProtoTenant(t *tenant.Tenant) *tenantv1.Tenant {
	return &tenantv1.Tenant{
		Id:         t.ID,
		Slug:       t.Slug,
		Name:       t.Name,
		Enabled:    t.Enabled,
		Version:    int32(t.Version), //nolint:gosec // Version is an optimistic lock counter, cannot realistically overflow int32
		CreatedAt:  timestamppb.New(t.CreatedAt),
		ModifiedAt: timestamppb.New(t.ModifiedAt),
	}
}

func toProtoRegistrationStatus(reg *registration.Registration) *tenantv1.RegistrationStatusResponse {
	resp := &tenantv1.RegistrationStatusResponse{
		Slug:   reg.Slug,
		Status: toProtoRegistrationStatusEnum(reg.Status),
	}
	if reg.FailureReason != nil {
		resp.FailureReason = reg.FailureReason
	}
	return resp
}

func toGetRegistrationStatusResponse(reg *registration.Registration) *tenantv1.GetRegistrationStatusResponse {
	resp := &tenantv1.GetRegistrationStatusResponse{
		Slug:   reg.Slug,
		Status: toProtoRegistrationStatusEnum(reg.Status),
	}
	if reg.FailureReason != nil {
		resp.FailureReason = reg.FailureReason
	}
	return resp
}

func toProtoRegistrationStatusEnum(s registration.Status) tenantv1.RegistrationStatus {
	switch s {
	case registration.StatusProvisioning:
		return tenantv1.RegistrationStatus_REGISTRATION_STATUS_PROVISIONING
	case registration.StatusCompleted:
		return tenantv1.RegistrationStatus_REGISTRATION_STATUS_COMPLETED
	case registration.StatusCompensating:
		return tenantv1.RegistrationStatus_REGISTRATION_STATUS_COMPENSATING
	case registration.StatusRolledBack:
		return tenantv1.RegistrationStatus_REGISTRATION_STATUS_ROLLED_BACK
	default:
		return tenantv1.RegistrationStatus_REGISTRATION_STATUS_UNSPECIFIED
	}
}

func mapConnectError(err error) *connect.Error {
	switch {
	case errors.Is(err, tenant.ErrInvalidTenantData), errors.Is(err, registration.ErrInvalidRegistration):
		return connect.NewError(connect.CodeInvalidArgument, err)
	case errors.Is(err, tenant.ErrSlugAlreadyExists), errors.Is(err, registration.ErrRegistrationAlreadyExists):
		return connect.NewError(connect.CodeAlreadyExists, err)
	case errors.Is(err, tenant.ErrUserAlreadyExists):
		return connect.NewError(connect.CodeAlreadyExists, err)
	case errors.Is(err, mongo.ErrEntityNotFound), errors.Is(err, registration.ErrRegistrationNotFound):
		return connect.NewError(connect.CodeNotFound, err)
	case errors.Is(err, mongo.ErrOptimisticLocking):
		return connect.NewError(connect.CodeAborted, err)
	default:
		return connect.NewError(connect.CodeInternal, err)
	}
}
