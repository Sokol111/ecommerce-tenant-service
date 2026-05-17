package tenant

import (
	"context"
	"fmt"
)

type GetEnabledSlugsHandler interface {
	Handle(ctx context.Context) ([]string, error)
}

type getEnabledSlugsHandler struct {
	repo Repository
}

func NewGetEnabledSlugsHandler(repo Repository) GetEnabledSlugsHandler {
	return &getEnabledSlugsHandler{repo: repo}
}

func (h *getEnabledSlugsHandler) Handle(ctx context.Context) ([]string, error) {
	slugs, err := h.repo.FindEnabledSlugs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get enabled tenant slugs: %w", err)
	}
	return slugs, nil
}
