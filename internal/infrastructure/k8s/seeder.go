package k8s

import (
	"context"
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// SeederJobLauncher creates K8s Jobs from a CronJob template to seed catalog data for new tenants.
type SeederJobLauncher struct {
	client      kubernetes.Interface
	namespace   string
	cronJobName string
}

func NewSeederJobLauncher(client kubernetes.Interface, cfg Config) *SeederJobLauncher {
	return &SeederJobLauncher{
		client:      client,
		namespace:   cfg.Namespace,
		cronJobName: cfg.CronJobName,
	}
}

func (l *SeederJobLauncher) SeedTenant(ctx context.Context, tenantSlug string) error {
	cronJob, err := l.client.BatchV1().CronJobs(l.namespace).Get(ctx, l.cronJobName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get cronjob template %q: %w", l.cronJobName, err)
	}

	job := l.buildJobFromTemplate(cronJob, tenantSlug)

	_, err = l.client.BatchV1().Jobs(l.namespace).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create seeder job for tenant %q: %w", tenantSlug, err)
	}

	return nil
}

func (l *SeederJobLauncher) buildJobFromTemplate(cronJob *batchv1.CronJob, tenantSlug string) *batchv1.Job {
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("seeder-%s-", tenantSlug),
			Namespace:    l.namespace,
			Labels:       cronJob.Spec.JobTemplate.Labels,
			Annotations: map[string]string{
				"cronjob.kubernetes.io/instantiate": "manual",
			},
		},
		Spec: *cronJob.Spec.JobTemplate.Spec.DeepCopy(),
	}

	// Add tenant-slug label
	if job.Labels == nil {
		job.Labels = make(map[string]string)
	}
	job.Labels["tenant-slug"] = tenantSlug

	// Inject TENANT_SLUG env var into the first container
	if len(job.Spec.Template.Spec.Containers) > 0 {
		job.Spec.Template.Spec.Containers[0].Env = append(
			job.Spec.Template.Spec.Containers[0].Env,
			corev1.EnvVar{Name: "TENANT_SLUG", Value: tenantSlug},
		)
	}

	return job
}
