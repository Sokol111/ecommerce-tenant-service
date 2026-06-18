package main

import (
	"context"

	commons_core "github.com/Sokol111/ecommerce-commons/pkg/core"
	commons_http "github.com/Sokol111/ecommerce-commons/pkg/http"
	commons_messaging "github.com/Sokol111/ecommerce-commons/pkg/messaging"
	commons_observability "github.com/Sokol111/ecommerce-commons/pkg/observability"
	commons_persistence "github.com/Sokol111/ecommerce-commons/pkg/persistence"
	commons_validation "github.com/Sokol111/ecommerce-commons/pkg/security/validation"
	"github.com/Sokol111/ecommerce-tenant-service/internal/application"
	internalconnect "github.com/Sokol111/ecommerce-tenant-service/internal/infrastructure/inbound/connect"
	internalk8s "github.com/Sokol111/ecommerce-tenant-service/internal/infrastructure/outbound/k8s"
	"github.com/Sokol111/ecommerce-tenant-service/internal/infrastructure/outbound/kafka"
	"github.com/Sokol111/ecommerce-tenant-service/internal/infrastructure/outbound/logto"
	internalmongo "github.com/Sokol111/ecommerce-tenant-service/internal/infrastructure/outbound/mongo"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var AppModules = fx.Options(
	// Commons
	commons_core.NewCoreModule(),
	commons_persistence.NewPersistenceModule(commons_persistence.WithMigrations()),
	commons_http.NewHTTPModule(commons_http.WithH2C()),
	commons_observability.NewObservabilityModule(),
	commons_messaging.NewMessagingModule(),
	commons_validation.NewModule(),

	// Domain & Application
	internalmongo.Module(),
	application.Module(),
	kafka.Module(),
	logto.Module(),
	internalk8s.Module(),

	// Connect (gRPC/Connect-RPC — new)
	internalconnect.Module(),
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
