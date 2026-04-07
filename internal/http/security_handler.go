package http //nolint:revive // package name intentional

import (
	"context"

	"github.com/Sokol111/ecommerce-commons/pkg/security/token"
	"github.com/Sokol111/ecommerce-tenant-service-api/gen/httpapi"
)

type securityHandler struct {
	handler token.SecurityHandler
}

func newSecurityHandler(handler token.SecurityHandler) httpapi.SecurityHandler {
	return &securityHandler{handler: handler}
}

func (s *securityHandler) HandleBearerAuth(ctx context.Context, _ httpapi.OperationName, t httpapi.BearerAuth) (context.Context, error) {
	ctx, _, err := s.handler.HandleBearerAuth(ctx, t.Token, t.Roles)
	return ctx, err
}
