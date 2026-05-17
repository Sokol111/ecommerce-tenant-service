package registration

import (
	"context"
)

type GetStatusQuery struct {
	Slug string
}

type GetStatusHandler interface {
	Handle(ctx context.Context, query GetStatusQuery) (*Registration, error)
}

type getStatusHandler struct {
	regRepo Repository
}

func NewGetStatusHandler(regRepo Repository) GetStatusHandler {
	return &getStatusHandler{regRepo: regRepo}
}

func (h *getStatusHandler) Handle(ctx context.Context, query GetStatusQuery) (*Registration, error) {
	return h.regRepo.FindBySlug(ctx, query.Slug)
}
