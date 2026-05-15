package registration

import "errors"

var (
	ErrInvalidRegistration       = errors.New("invalid registration data")
	ErrRegistrationNotFound      = errors.New("registration not found")
	ErrRegistrationAlreadyExists = errors.New("registration for this slug already exists")
)
