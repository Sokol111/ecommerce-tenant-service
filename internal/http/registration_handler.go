package http //nolint:revive // package name intentional

import (
	"context"
	"errors"

	"github.com/Sokol111/ecommerce-tenant-service/internal/application/registration"
	"github.com/Sokol111/ecommerce-tenant-service/internal/application/tenant"

	"github.com/Sokol111/ecommerce-tenant-service-api/gen/httpapi"
)

func (h *tenantHandler) RegisterTenant(ctx context.Context, req *httpapi.RegisterTenantRequest) (httpapi.RegisterTenantRes, error) {
	cmd := registration.RegisterCommand{
		Slug:      req.Slug,
		Name:      req.Name,
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	result, err := h.registrations.Register(ctx, cmd)
	if err != nil {
		if errors.Is(err, tenant.ErrInvalidTenantData) {
			return &httpapi.RegisterTenantBadRequest{
				Status: 400,
				Type:   *aboutBlankURL,
				Title:  err.Error(),
			}, nil
		}
		if errors.Is(err, tenant.ErrSlugAlreadyExists) || errors.Is(err, registration.ErrRegistrationAlreadyExists) {
			return &httpapi.RegisterTenantConflict{
				Status: 409,
				Type:   *aboutBlankURL,
				Title:  "Tenant with this slug already exists",
			}, nil
		}
		return nil, err
	}

	// Completed synchronously — return 201 with tenant
	if result.Tenant != nil {
		return toTenantResponse(result.Tenant), nil
	}

	// Deferred to worker — return 202 with status
	return toRegistrationStatusResponse(result.Registration), nil
}

func (h *tenantHandler) GetRegistrationStatus(ctx context.Context, params httpapi.GetRegistrationStatusParams) (httpapi.GetRegistrationStatusRes, error) {
	q := registration.GetStatusQuery{Slug: params.Slug}

	reg, err := h.registrations.GetStatus(ctx, q)
	if err != nil {
		if errors.Is(err, registration.ErrRegistrationNotFound) {
			return &httpapi.GetRegistrationStatusNotFound{
				Status: 404,
				Type:   *aboutBlankURL,
				Title:  "Registration not found",
			}, nil
		}
		return nil, err
	}

	return toRegistrationStatusResponse(reg), nil
}

func toRegistrationStatusResponse(reg *registration.Registration) *httpapi.RegistrationStatusResponse {
	resp := &httpapi.RegistrationStatusResponse{
		Slug:   reg.Slug,
		Status: httpapi.RegistrationStatusResponseStatus(reg.Status),
	}
	if reg.FailureReason != nil {
		resp.FailureReason = httpapi.NewOptString(*reg.FailureReason)
	}
	return resp
}
