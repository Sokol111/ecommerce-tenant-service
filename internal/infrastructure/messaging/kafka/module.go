package kafka

import (
	"github.com/Sokol111/ecommerce-tenant-service-api/gen/events"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Options(
		events.Module(),
	)
}
