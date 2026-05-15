package tenant

import (
	"context"
	"fmt"
)

func (s *tenantService) GetEnabledSlugs(ctx context.Context) ([]string, error) {
	slugs, err := s.repo.FindEnabledSlugs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get enabled tenant slugs: %w", err)
	}
	return slugs, nil
}
