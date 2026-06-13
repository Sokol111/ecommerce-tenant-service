package tenant

import "errors"

var (
	ErrInvalidTenantData = errors.New("invalid tenant data")
	ErrSlugAlreadyExists = errors.New("tenant with this slug already exists")
	ErrUserAlreadyExists = errors.New("user already exists in identity provider")
)
