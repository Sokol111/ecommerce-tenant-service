package tenant

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Tenant struct {
	ID         string
	Slug       string
	Version    int
	Name       string
	Enabled    bool
	CreatedAt  time.Time
	ModifiedAt time.Time
}

func NewTenant(slug, name string) (*Tenant, error) {
	if err := validate(slug, name); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	return &Tenant{
		ID:         uuid.New().String(),
		Slug:       slug,
		Version:    1,
		Name:       name,
		Enabled:    true,
		CreatedAt:  now,
		ModifiedAt: now,
	}, nil
}

func Reconstruct(id, slug string, version int, name string, enabled bool, createdAt, modifiedAt time.Time) *Tenant {
	return &Tenant{
		ID:         id,
		Slug:       slug,
		Version:    version,
		Name:       name,
		Enabled:    enabled,
		CreatedAt:  createdAt,
		ModifiedAt: modifiedAt,
	}
}

func (t *Tenant) Update(name string, enabled bool) error {
	if name == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidTenantData)
	}
	if len(name) > 200 {
		return fmt.Errorf("%w: name is too long (max 200 characters)", ErrInvalidTenantData)
	}

	t.Name = name
	t.Enabled = enabled
	t.ModifiedAt = time.Now().UTC()
	return nil
}

func validate(slug, name string) error {
	if slug == "" {
		return fmt.Errorf("%w: slug is required", ErrInvalidTenantData)
	}
	if len(slug) < 2 || len(slug) > 63 {
		return fmt.Errorf("%w: slug must be between 2 and 63 characters", ErrInvalidTenantData)
	}
	if name == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidTenantData)
	}
	if len(name) > 200 {
		return fmt.Errorf("%w: name is too long (max 200 characters)", ErrInvalidTenantData)
	}
	return nil
}
