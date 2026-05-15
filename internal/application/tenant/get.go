package tenant

import (
	"context"
	"fmt"
)

type GetBySlugQuery struct {
	Slug string
}

func (s *tenantService) GetBySlug(ctx context.Context, query GetBySlugQuery) (*Tenant, error) {
	t, err := s.repo.FindBySlug(ctx, query.Slug)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}
	return t, nil
}
