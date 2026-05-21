package mongo

import "github.com/Sokol111/ecommerce-tenant-service/internal/application/tenant"

type tenantMapper struct{}

func newTenantMapper() *tenantMapper {
	return &tenantMapper{}
}

func (m *tenantMapper) ToEntity(t *tenant.Tenant) *tenantEntity {
	return &tenantEntity{
		ID:         t.ID,
		Slug:       t.Slug,
		Version:    t.Version,
		Name:       t.Name,
		Enabled:    t.Enabled,
		CreatedAt:  t.CreatedAt,
		ModifiedAt: t.ModifiedAt,
	}
}

func (m *tenantMapper) ToDomain(e *tenantEntity) *tenant.Tenant {
	return tenant.Reconstruct(
		e.ID,
		e.Slug,
		e.Version,
		e.Name,
		e.Enabled,
		e.CreatedAt.UTC(),
		e.ModifiedAt.UTC(),
	)
}

func (m *tenantMapper) GetID(e *tenantEntity) string {
	return e.ID
}

func (m *tenantMapper) GetVersion(e *tenantEntity) int {
	return e.Version
}

func (m *tenantMapper) SetVersion(e *tenantEntity, version int) {
	e.Version = version
}
