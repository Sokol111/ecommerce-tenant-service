package main

import (
	"context"

	commons_core "github.com/Sokol111/ecommerce-commons/pkg/core"
	commons_http "github.com/Sokol111/ecommerce-commons/pkg/http"
	commons_messaging "github.com/Sokol111/ecommerce-commons/pkg/messaging"
	commons_observability "github.com/Sokol111/ecommerce-commons/pkg/observability"
	commons_persistence "github.com/Sokol111/ecommerce-commons/pkg/persistence"
	commons_pprof "github.com/Sokol111/ecommerce-commons/pkg/pprof"
	commons_security "github.com/Sokol111/ecommerce-commons/pkg/security"
	commons_swaggerui "github.com/Sokol111/ecommerce-commons/pkg/swaggerui"
	"github.com/Sokol111/ecommerce-tenant-service-api/gen/httpapi"
	"github.com/Sokol111/ecommerce-tenant-service/internal/application"
	"github.com/Sokol111/ecommerce-tenant-service/internal/event"
	internalhttp "github.com/Sokol111/ecommerce-tenant-service/internal/http"
	"github.com/Sokol111/ecommerce-tenant-service/internal/infrastructure/logto"
	"github.com/Sokol111/ecommerce-tenant-service/internal/infrastructure/messaging/kafka"
	internalmongo "github.com/Sokol111/ecommerce-tenant-service/internal/infrastructure/persistence/mongo"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var AppModules = fx.Options(
	// Commons
	commons_core.NewCoreModule(),
	commons_persistence.NewPersistenceModule(),
	commons_http.NewHTTPModule(),
	commons_observability.NewObservabilityModule(),
	commons_messaging.NewMessagingModule(),
	commons_security.NewSecurityModule(),
	commons_pprof.NewPprofModule(),
	commons_swaggerui.NewSwaggerModule(commons_swaggerui.SwaggerConfig{OpenAPIContent: httpapi.OpenAPIDoc}),

	// Domain & Application
	internalmongo.Module(),
	event.Module(),
	application.Module(),
	kafka.Module(),
	logto.Module(),

	// HTTP
	internalhttp.Module(),
)

func main() {
	app := fx.New(
		AppModules,
		fx.Invoke(func(lc fx.Lifecycle, log *zap.Logger) {
			lc.Append(fx.Hook{
				OnStop: func(ctx context.Context) error {
					log.Info("Application stopping...")
					return nil
				},
			})
		}),
	)
	app.Run()
}
