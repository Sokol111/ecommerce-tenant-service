package application

import (
	command "github.com/Sokol111/ecommerce-tenant-service/internal/application/command/tenant"
	query "github.com/Sokol111/ecommerce-tenant-service/internal/application/query/tenant"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			command.NewCreateTenantHandler,
			command.NewUpdateTenantHandler,
			command.NewDeleteTenantHandler,
		),
		fx.Provide(
			query.NewGetTenantBySlugHandler,
			query.NewGetTenantListHandler,
			query.NewGetEnabledSlugsHandler,
		),
	)
}
