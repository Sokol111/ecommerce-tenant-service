package registration

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusProvisioning Status = "provisioning"
	StatusCompleted    Status = "completed"
	StatusCompensating Status = "compensating"
	StatusRolledBack   Status = "rolled_back"
)

type Registration struct {
	ID        string
	Slug      string
	Status    Status
	Name      string
	Email     string
	FirstName string
	LastName  string

	TenantID       *string
	LogtoUserID    *string
	TenantSet      bool
	RoleAssigned   bool
	EventPublished bool
	CatalogSeeded  bool

	FailureReason *string
	RetryCount    int32
	NextRetryAt   *time.Time

	CreatedAt   time.Time
	CompletedAt *time.Time
	Version     int64
}

func New(slug, name, email, firstName, lastName, logtoUserID string) (*Registration, error) {
	if slug == "" || name == "" || email == "" || firstName == "" || lastName == "" {
		return nil, fmt.Errorf("%w: all fields are required", ErrInvalidRegistration)
	}
	if logtoUserID == "" {
		return nil, fmt.Errorf("%w: logto user ID is required", ErrInvalidRegistration)
	}

	now := time.Now().UTC()
	return &Registration{
		ID:          uuid.New().String(),
		Slug:        slug,
		Status:      StatusProvisioning,
		Name:        name,
		Email:       email,
		FirstName:   firstName,
		LastName:    lastName,
		LogtoUserID: &logtoUserID,
		CreatedAt:   now,
		Version:     1,
	}, nil
}

func Reconstruct(
	id, slug string, status Status,
	name, email, firstName, lastName string,
	tenantID, logtoUserID *string,
	tenantSet, roleAssigned, eventPublished, catalogSeeded bool,
	failureReason *string,
	retryCount int32, nextRetryAt *time.Time,
	createdAt time.Time, completedAt *time.Time,
	version int64,
) *Registration {
	return &Registration{
		ID:             id,
		Slug:           slug,
		Status:         status,
		Name:           name,
		Email:          email,
		FirstName:      firstName,
		LastName:       lastName,
		TenantID:       tenantID,
		LogtoUserID:    logtoUserID,
		TenantSet:      tenantSet,
		RoleAssigned:   roleAssigned,
		EventPublished: eventPublished,
		CatalogSeeded:  catalogSeeded,
		FailureReason:  failureReason,
		RetryCount:     retryCount,
		NextRetryAt:    nextRetryAt,
		CreatedAt:      createdAt,
		CompletedAt:    completedAt,
		Version:        version,
	}
}

func (r *Registration) SetTenantID(id string) {
	r.TenantID = &id
}

func (r *Registration) SetTenantOnUser() {
	r.TenantSet = true
}

func (r *Registration) SetRoleAssigned() {
	r.RoleAssigned = true
}

func (r *Registration) SetEventPublished() {
	r.EventPublished = true
}

func (r *Registration) SetCatalogSeeded() {
	r.CatalogSeeded = true
}

func (r *Registration) MarkCompleted() {
	r.Status = StatusCompleted
	now := time.Now().UTC()
	r.CompletedAt = &now
}

func (r *Registration) MarkCompensating(reason string) {
	r.Status = StatusCompensating
	r.FailureReason = &reason
}

func (r *Registration) ClearLogtoUser() {
	r.LogtoUserID = nil
	r.TenantSet = false
	r.RoleAssigned = false
}

func (r *Registration) ClearTenant() {
	r.TenantID = nil
}

func (r *Registration) MarkRolledBack() {
	r.Status = StatusRolledBack
	now := time.Now().UTC()
	r.CompletedAt = &now
}

func (r *Registration) ScheduleRetry() {
	r.RetryCount++
	next := r.nextRetryDelay()
	nextAt := time.Now().UTC().Add(next)
	r.NextRetryAt = &nextAt
}

func (r *Registration) nextRetryDelay() time.Duration {
	base := 30 * time.Second
	delay := base
	for i := 0; i < int(r.RetryCount-1) && i < 5; i++ {
		delay *= 2
	}
	maxDelay := 15 * time.Minute
	if delay > maxDelay {
		return maxDelay
	}
	return delay
}

func (r *Registration) IsTerminal() bool {
	return r.Status == StatusCompleted || r.Status == StatusRolledBack
}
