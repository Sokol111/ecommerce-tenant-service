package logto

import (
	"context"
	"net/url"

	"github.com/Sokol111/ecommerce-tenant-service/internal/application/tenant"
	"github.com/knadh/koanf/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"golang.org/x/oauth2/clientcredentials"
)

func Module() fx.Option {
	return fx.Provide(
		provideLogtoConfig,
		provideIdentityProvider,
	)
}

func provideLogtoConfig(k *koanf.Koanf) Config {
	var cfg Config
	if err := k.Unmarshal("logto", &cfg); err != nil {
		cfg = Config{}
	}
	return cfg
}

func provideIdentityProvider(cfg Config, log *zap.Logger) (tenant.IdentityProvider, error) {
	cc := &clientcredentials.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		TokenURL:     cfg.TokenURL,
		Scopes:       []string{"all"},
		EndpointParams: url.Values{
			"resource": {cfg.ManagementResource()},
		},
	}
	ts := cc.TokenSource(context.Background())
	return newLogtoClient(cfg, ts, log)
}
