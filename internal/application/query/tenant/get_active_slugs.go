package tenant

import (
	"context"
	"fmt"

	"github.com/Sokol111/ecommerce-tenant-service/internal/domain/tenant"
)

type GetEnabledSlugsQueryHandler interface {
	Handle(ctx context.Context) ([]string, error)
}

type getEnabledSlugsHandler struct {
	repo tenant.Repository
}

func NewGetEnabledSlugsHandler(repo tenant.Repository) GetEnabledSlugsQueryHandler {
	return &getEnabledSlugsHandler{repo: repo}
}

func (h *getEnabledSlugsHandler) Handle(ctx context.Context) ([]string, error) {
	slugs, err := h.repo.FindEnabledSlugs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get enabled tenant slugs: %w", err)
	}
	return slugs, nil
}
