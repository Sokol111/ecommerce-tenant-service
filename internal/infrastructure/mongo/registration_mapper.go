package mongo

import (
	"time"

	"github.com/Sokol111/ecommerce-tenant-service/internal/application/registration"
)

type registrationMapper struct{}

func newRegistrationMapper() *registrationMapper {
	return &registrationMapper{}
}

func (m *registrationMapper) ToEntity(r *registration.Registration) *registrationEntity {
	return &registrationEntity{
		ID:             r.ID,
		Slug:           r.Slug,
		Status:         string(r.Status),
		Name:           r.Name,
		Email:          r.Email,
		FirstName:      r.FirstName,
		LastName:       r.LastName,
		TenantID:       r.TenantID,
		LogtoUserID:    r.LogtoUserID,
		TenantSet:      r.TenantSet,
		RoleAssigned:   r.RoleAssigned,
		EventPublished: r.EventPublished,
		FailureReason:  r.FailureReason,
		RetryCount:     r.RetryCount,
		NextRetryAt:    r.NextRetryAt,
		CreatedAt:      r.CreatedAt,
		CompletedAt:    r.CompletedAt,
		Version:        r.Version,
	}
}

func (m *registrationMapper) ToDomain(e *registrationEntity) *registration.Registration {
	var completedAt *time.Time
	if e.CompletedAt != nil {
		t := e.CompletedAt.UTC()
		completedAt = &t
	}

	var nextRetryAt *time.Time
	if e.NextRetryAt != nil {
		t := e.NextRetryAt.UTC()
		nextRetryAt = &t
	}

	return registration.Reconstruct(
		e.ID,
		e.Slug,
		registration.Status(e.Status),
		e.Name,
		e.Email,
		e.FirstName,
		e.LastName,
		e.TenantID,
		e.LogtoUserID,
		e.TenantSet,
		e.RoleAssigned,
		e.EventPublished,
		e.FailureReason,
		e.RetryCount,
		nextRetryAt,
		e.CreatedAt.UTC(),
		completedAt,
		e.Version,
	)
}

func (m *registrationMapper) GetID(e *registrationEntity) string {
	return e.ID
}

func (m *registrationMapper) GetVersion(e *registrationEntity) int {
	return e.Version
}

func (m *registrationMapper) SetVersion(e *registrationEntity, version int) {
	e.Version = version
}
