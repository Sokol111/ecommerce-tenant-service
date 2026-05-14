package tenant

import (
	"context"

	"github.com/Sokol111/ecommerce-tenant-service/internal/domain/registration"
)

type GetRegistrationStatusQuery struct {
	Slug string
}

type GetRegistrationStatusQueryHandler interface {
	Handle(ctx context.Context, query GetRegistrationStatusQuery) (*registration.Registration, error)
}

type getRegistrationStatusHandler struct {
	regRepo registration.Repository
}

func NewGetRegistrationStatusHandler(regRepo registration.Repository) GetRegistrationStatusQueryHandler {
	return &getRegistrationStatusHandler{regRepo: regRepo}
}

func (h *getRegistrationStatusHandler) Handle(ctx context.Context, query GetRegistrationStatusQuery) (*registration.Registration, error) {
	return h.regRepo.FindBySlug(ctx, query.Slug)
}
