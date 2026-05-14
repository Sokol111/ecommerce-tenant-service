package logto

import (
	"net/url"

	"github.com/Sokol111/ecommerce-commons/pkg/security/token"
	"github.com/Sokol111/ecommerce-tenant-service/internal/domain/tenant"
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

func provideIdentityProvider(cfg Config, ccConfig token.ClientCredentialsConfig, log *zap.Logger) (tenant.IdentityProvider, error) {
	cc := &clientcredentials.Config{
		ClientID:     ccConfig.ClientID,
		ClientSecret: ccConfig.ClientSecret,
		TokenURL:     ccConfig.TokenURL,
		Scopes:       []string{"all"},
		EndpointParams: url.Values{
			"resource": {cfg.getResource()},
		},
	}
	ts := cc.TokenSource(nil)
	return newLogtoClient(cfg, ts, log)
}
