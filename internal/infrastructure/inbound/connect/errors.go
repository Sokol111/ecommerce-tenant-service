package connect

import (
	"errors"

	"connectrpc.com/connect"
	"github.com/Sokol111/ecommerce-commons/pkg/persistence/mongo"
	"github.com/Sokol111/ecommerce-tenant-service/internal/application/registration"
	"github.com/Sokol111/ecommerce-tenant-service/internal/application/tenant"
)

// mapConnectError maps domain errors to connect.Error codes.
func mapConnectError(err error) *connect.Error {
	switch {
	case errors.Is(err, tenant.ErrInvalidTenantData), errors.Is(err, registration.ErrInvalidRegistration):
		return connect.NewError(connect.CodeInvalidArgument, err)
	case errors.Is(err, tenant.ErrSlugAlreadyExists), errors.Is(err, registration.ErrRegistrationAlreadyExists):
		return connect.NewError(connect.CodeAlreadyExists, err)
	case errors.Is(err, tenant.ErrUserAlreadyExists):
		return connect.NewError(connect.CodeAlreadyExists, err)
	case errors.Is(err, mongo.ErrEntityNotFound), errors.Is(err, registration.ErrRegistrationNotFound):
		return connect.NewError(connect.CodeNotFound, err)
	case errors.Is(err, mongo.ErrOptimisticLocking):
		return connect.NewError(connect.CodeAborted, err)
	default:
		return connect.NewError(connect.CodeInternal, err)
	}
}
