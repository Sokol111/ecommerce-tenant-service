package logto

type Config struct {
	BaseURL  string `koanf:"base-url"`
	Resource string `koanf:"resource"`
}

func (c Config) getResource() string {
	if c.Resource != "" {
		return c.Resource
	}
	return "https://default.logto.app/api"
}
