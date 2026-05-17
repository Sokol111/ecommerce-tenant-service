package http //nolint:revive // package name intentional

import (
	"net/url"

	"github.com/Sokol111/ecommerce-tenant-service/internal/application/registration"
	"github.com/Sokol111/ecommerce-tenant-service/internal/application/tenant"
)

var aboutBlankURL = mustParseURL("about:blank")

func mustParseURL(rawURL string) *url.URL {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}

	return parsedURL
}

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
