package connect

import (
	"context"

	"connectrpc.com/connect"
	tenantv1 "github.com/Sokol111/ecommerce-tenant-service-api/gen/go/tenant/v1"
	"github.com/Sokol111/ecommerce-tenant-service/internal/application/registration"
)

type registrationHandler struct {
	register  registration.RegisterHandler
	getStatus registration.GetStatusHandler
}

// ==================== Registration ====================

func (h *registrationHandler) RegisterTenant(ctx context.Context, req *connect.Request[tenantv1.RegisterTenantRequest]) (*connect.Response[tenantv1.RegisterTenantResponse], error) {
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
		return nil, err
	}

	return connect.NewResponse(toProtoRegistrationStatus(result.Registration)), nil
}

func (h *registrationHandler) GetRegistrationStatus(ctx context.Context, req *connect.Request[tenantv1.GetRegistrationStatusRequest]) (*connect.Response[tenantv1.GetRegistrationStatusResponse], error) {
	q := registration.GetStatusQuery{Slug: req.Msg.Slug}

	reg, err := h.getStatus.Handle(ctx, q)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(toGetRegistrationStatusResponse(reg)), nil
}

// ==================== Helpers ====================

func toProtoRegistrationStatus(reg *registration.Registration) *tenantv1.RegisterTenantResponse {
	resp := &tenantv1.RegisterTenantResponse{
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
