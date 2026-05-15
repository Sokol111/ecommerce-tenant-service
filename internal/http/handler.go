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
	tenants       tenant.TenantService
	registrations registration.RegistrationService
}

func newTenantHandler(
	tenants tenant.TenantService,
	registrations registration.RegistrationService,
) *tenantHandler {
	return &tenantHandler{
		tenants:       tenants,
		registrations: registrations,
	}
}
