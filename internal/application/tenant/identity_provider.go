package tenant

import "context"

// CreateUserParams holds the parameters needed to create a user in the identity provider.
type CreateUserParams struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
}

// IdentityProvider defines the port for interacting with the external identity provider.
type IdentityProvider interface {
	CreateUser(ctx context.Context, params CreateUserParams) (userID string, err error)
	SetUserTenant(ctx context.Context, userID string, tenantSlug string) error
	AssignRole(ctx context.Context, userID string, roleName string) error
	DeleteUser(ctx context.Context, userID string) error
}
