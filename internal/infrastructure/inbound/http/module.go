package http //nolint:revive // package name intentional

import (
	"go.uber.org/fx"

	"github.com/Sokol111/ecommerce-tenant-service-api/gen/httpapi"
)

func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			newTenantHandler,
			newHandler,
		),
	)
}

func newHandler(tenantHandler *tenantHandler) httpapi.Handler {
	return tenantHandler
}
