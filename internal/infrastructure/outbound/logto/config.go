package logto

type Config struct {
	BaseURL      string `koanf:"base-url"`
	Resource     string `koanf:"resource"`
	ClientID     string `koanf:"client-id"`
	ClientSecret string `koanf:"client-secret"`
	TokenURL     string `koanf:"token-url"`
}

// ManagementResource returns the Logto Management API resource indicator.
func (c Config) ManagementResource() string {
	if c.Resource != "" {
		return c.Resource
	}
	return "https://default.logto.app/api"
}
