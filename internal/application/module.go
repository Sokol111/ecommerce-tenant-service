package application

import (
	"github.com/Sokol111/ecommerce-commons/pkg/core/worker"
	command "github.com/Sokol111/ecommerce-tenant-service/internal/application/command/tenant"
	query "github.com/Sokol111/ecommerce-tenant-service/internal/application/query/tenant"
	appworker "github.com/Sokol111/ecommerce-tenant-service/internal/application/worker"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			command.NewCreateTenantHandler,
			command.NewUpdateTenantHandler,
			command.NewDeleteTenantHandler,
			command.NewRegisterTenantHandler,
			command.NewSagaProcessor,
		),
		fx.Provide(
			query.NewGetTenantBySlugHandler,
			query.NewGetTenantListHandler,
			query.NewGetEnabledSlugsHandler,
			query.NewGetRegistrationStatusHandler,
		),
		fx.Provide(appworker.NewRegistrationWorker),
		fx.Invoke(worker.RunWorker[*appworker.RegistrationWorker]("registration-worker", worker.WithReady())),
	)
}
