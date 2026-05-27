package k8s

import (
	"fmt"

	"github.com/Sokol111/ecommerce-tenant-service/internal/application/registration"
	"github.com/knadh/koanf/v2"
	"go.uber.org/fx"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func Module() fx.Option {
	return fx.Provide(
		provideConfig,
		provideKubeClient,
		provideSeeder,
	)
}

func provideConfig(k *koanf.Koanf) Config {
	var cfg Config
	if err := k.Unmarshal("k8s", &cfg); err != nil {
		cfg = Config{}
	}
	return cfg
}

func provideKubeClient() (kubernetes.Interface, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create in-cluster k8s config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create k8s clientset: %w", err)
	}

	return clientset, nil
}

func provideSeeder(client kubernetes.Interface, cfg Config) registration.CatalogSeeder {
	return NewSeederJobLauncher(client, cfg)
}
