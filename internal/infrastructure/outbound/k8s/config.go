package k8s

// Config holds configuration for the K8s seeder job launcher.
type Config struct {
	Namespace   string `koanf:"namespace"`
	CronJobName string `koanf:"cronjob-name"`
}
