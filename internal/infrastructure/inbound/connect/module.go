package connect

import (
	"net/http"

	"connectrpc.com/connect"
	"github.com/Sokol111/ecommerce-commons/pkg/security/validation"
	tenantv1connect "github.com/Sokol111/ecommerce-tenant-service-api/gen/go/tenant/v1/tenantv1connect"
	"github.com/Sokol111/ecommerce-tenant-service/internal/application/registration"
	"github.com/Sokol111/ecommerce-tenant-service/internal/application/tenant"
	"go.uber.org/fx"
)

// Module provides the Connect gRPC/Connect-RPC server handlers for tenant operations.
func Module() fx.Option {
	return fx.Options(
		ErrorMappingModule(25),
		fx.Provide(
			newTenantHandler,
			newRegistrationHandler,
			provideProcedurePermissions,
		),
		fx.Invoke(registerConnectRoutes),
		fx.Invoke(registerRegistrationRoutes),
	)
}

// newTenantHandler wires domain/application handlers into a Connect handler for CRUD operations.
func newTenantHandler(
	createTenant tenant.CreateTenantHandler,
	updateTenant tenant.UpdateTenantHandler,
	deleteTenant tenant.DeleteTenantHandler,
	getBySlug tenant.GetBySlugHandler,
	getList tenant.GetListHandler,
	getEnabledSlugs tenant.GetEnabledSlugsHandler,
) *tenantHandler {
	return &tenantHandler{
		createTenant:    createTenant,
		updateTenant:    updateTenant,
		deleteTenant:    deleteTenant,
		getBySlug:       getBySlug,
		getList:         getList,
		getEnabledSlugs: getEnabledSlugs,
	}
}

// newRegistrationHandler wires registration domain/application handlers into a Connect handler.
func newRegistrationHandler(
	register registration.RegisterHandler,
	getStatus registration.GetStatusHandler,
) *registrationHandler {
	return &registrationHandler{
		register:  register,
		getStatus: getStatus,
	}
}

// registerConnectRoutes mounts the TenantService handler under /tenant.v1.TenantService/*.
func registerConnectRoutes(
	mux *http.ServeMux,
	handler *tenantHandler,
	interceptors []connect.Interceptor,
) {
	path, h := tenantv1connect.NewTenantServiceHandler(handler, connect.WithInterceptors(interceptors...))
	mux.Handle(path, h)
}

// registerRegistrationRoutes mounts the TenantRegistrationService handler under /tenant.v1.TenantRegistrationService/*.
func registerRegistrationRoutes(
	mux *http.ServeMux,
	handler *registrationHandler,
	interceptors []connect.Interceptor,
) {
	path, h := tenantv1connect.NewTenantRegistrationServiceHandler(handler, connect.WithInterceptors(interceptors...))
	mux.Handle(path, h)
}

// provideProcedurePermissions maps each RPC to required permission strings.
func provideProcedurePermissions() validation.ProcedurePermissions {
	return validation.ProcedurePermissions{
		// TenantService
		tenantv1connect.TenantServiceCreateTenantProcedure:          {"tenants:write"},
		tenantv1connect.TenantServiceUpdateTenantProcedure:          {"tenants:write"},
		tenantv1connect.TenantServiceDeleteTenantProcedure:          {"tenants:delete"},
		tenantv1connect.TenantServiceGetTenantBySlugProcedure:       {"tenants:read"},
		tenantv1connect.TenantServiceGetTenantListProcedure:         {"tenants:read"},
		tenantv1connect.TenantServiceGetEnabledTenantSlugsProcedure: {"tenants:read"},
		// TenantRegistrationService
		tenantv1connect.TenantRegistrationServiceRegisterTenantProcedure:        {"tenants:write"},
		tenantv1connect.TenantRegistrationServiceGetRegistrationStatusProcedure: {"tenants:read"},
	}
}
