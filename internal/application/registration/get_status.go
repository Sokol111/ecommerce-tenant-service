package registration

import (
	"context"
)

type GetStatusQuery struct {
	Slug string
}

func (s *registrationService) GetStatus(ctx context.Context, query GetStatusQuery) (*Registration, error) {
	return s.regRepo.FindBySlug(ctx, query.Slug)
}
