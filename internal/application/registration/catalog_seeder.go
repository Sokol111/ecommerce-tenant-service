package registration

import "context"

// CatalogSeeder triggers initial catalog data seeding for a newly registered tenant.
type CatalogSeeder interface {
	SeedTenant(ctx context.Context, tenantSlug string) error
}
