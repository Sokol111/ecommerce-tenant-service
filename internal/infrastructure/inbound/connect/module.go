package connect

import (
	"net/http"

	"connectrpc.com/connect"
	"github.com/Sokol111/ecommerce-commons/pkg/security/validation"
	tenantv1connect "github.com/Sokol111/ecommerce-tenant-service-api/gen/connect/tenant/v1/tenantv1connect"
	"github.com/Sokol111/ecommerce-tenant-service/internal/application/registration"
	"github.com/Sokol111/ecommerce-tenant-service/internal/application/tenant"
	"go.uber.org/fx"
)

// Module provides the Connect gRPC/Connect-RPC server handler for tenant operations.
func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			newTenantHandler,
			provideProcedurePermissions,
		),
		fx.Invoke(registerConnectRoutes),
	)
}

// newTenantHandler wires domain/application handlers into a Connect handler.
func newTenantHandler(
	createTenant tenant.CreateTenantHandler,
	updateTenant tenant.UpdateTenantHandler,
	deleteTenant tenant.DeleteTenantHandler,
	getBySlug tenant.GetBySlugHandler,
	getList tenant.GetListHandler,
	getEnabledSlugs tenant.GetEnabledSlugsHandler,
	register registration.RegisterHandler,
	getStatus registration.GetStatusHandler,
) *tenantHandler {
	return &tenantHandler{
		createTenant:    createTenant,
		updateTenant:    updateTenant,
		deleteTenant:    deleteTenant,
		getBySlug:       getBySlug,
		getList:         getList,
		getEnabledSlugs: getEnabledSlugs,
		register:        register,
		getStatus:       getStatus,
	}
}

// registerConnectRoutes mounts the Connect handler under /tenant.v1.TenantService/*.
// The interceptor chain (auth, recovery, logging, etc.) is injected via FX.
func registerConnectRoutes(
	mux *http.ServeMux,
	handler *tenantHandler,
	interceptors []connect.Interceptor,
) {
	path, h := tenantv1connect.NewTenantServiceHandler(handler, connect.WithInterceptors(interceptors...))
	mux.Handle(path, h)
}

// provideProcedurePermissions maps each tenant RPC to required permission strings.
func provideProcedurePermissions() validation.ProcedurePermissions {
	return validation.ProcedurePermissions{
		tenantv1connect.TenantServiceCreateTenantProcedure:          {"tenants:write"},
		tenantv1connect.TenantServiceUpdateTenantProcedure:          {"tenants:write"},
		tenantv1connect.TenantServiceDeleteTenantProcedure:          {"tenants:delete"},
		tenantv1connect.TenantServiceGetTenantBySlugProcedure:       {"tenants:read"},
		tenantv1connect.TenantServiceGetTenantListProcedure:         {"tenants:read"},
		tenantv1connect.TenantServiceGetEnabledTenantSlugsProcedure: {"tenants:read"},
		tenantv1connect.TenantServiceRegisterTenantProcedure:        {"tenants:write"},
		tenantv1connect.TenantServiceGetRegistrationStatusProcedure: {"tenants:read"},
	}
}
