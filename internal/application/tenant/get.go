package tenant

import (
	"context"
	"fmt"
)

type GetBySlugQuery struct {
	Slug string
}

type GetBySlugHandler interface {
	Handle(ctx context.Context, query GetBySlugQuery) (*Tenant, error)
}

type getBySlugHandler struct {
	repo Repository
}

func NewGetBySlugHandler(repo Repository) GetBySlugHandler {
	return &getBySlugHandler{repo: repo}
}

func (h *getBySlugHandler) Handle(ctx context.Context, query GetBySlugQuery) (*Tenant, error) {
	t, err := h.repo.FindBySlug(ctx, query.Slug)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}
	return t, nil
}
