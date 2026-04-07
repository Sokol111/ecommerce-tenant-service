package http //nolint:revive // package name intentional

import (
	"net/http"

	"go.uber.org/fx"

	"github.com/Sokol111/ecommerce-tenant-service-api/gen/httpapi"
)

func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			newTenantHandler,
			newHandler,
			httpapi.ProvideServer,
			newSecurityHandler,
		),
		fx.Invoke(registerOgenRoutes),
	)
}

func newHandler(tenantHandler *tenantHandler) httpapi.Handler {
	return tenantHandler
}

func registerOgenRoutes(mux *http.ServeMux, server *httpapi.Server) {
	mux.Handle("/", server)
}
