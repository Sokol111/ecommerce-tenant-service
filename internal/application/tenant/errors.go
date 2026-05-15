package tenant

import "errors"

var (
	ErrInvalidTenantData = errors.New("invalid tenant data")
	ErrSlugAlreadyExists = errors.New("tenant with this slug already exists")
	ErrTenantNotDisabled = errors.New("tenant must be disabled before deletion")
)
