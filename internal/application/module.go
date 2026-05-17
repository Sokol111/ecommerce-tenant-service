package application

import (
	"github.com/Sokol111/ecommerce-commons/pkg/core/worker"
	"github.com/Sokol111/ecommerce-tenant-service/internal/application/registration"
	"github.com/Sokol111/ecommerce-tenant-service/internal/application/tenant"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			tenant.NewCreateTenantHandler,
			tenant.NewUpdateTenantHandler,
			tenant.NewDeleteTenantHandler,
			tenant.NewGetBySlugHandler,
			tenant.NewGetListHandler,
			tenant.NewGetEnabledSlugsHandler,
		),
		fx.Provide(
			registration.NewProcessor,
			registration.NewRegisterHandler,
			registration.NewGetStatusHandler,
			registration.NewWorker,
		),
		fx.Invoke(worker.RunWorker[*registration.Worker]("registration-worker", worker.WithReady())),
	)
}
