package mongo

import "time"

type registrationEntity struct {
	ID             string     `bson:"_id"`
	Slug           string     `bson:"slug"`
	Status         string     `bson:"status"`
	Name           string     `bson:"name"`
	Email          string     `bson:"email"`
	FirstName      string     `bson:"firstName"`
	LastName       string     `bson:"lastName"`
	TenantID       *string    `bson:"tenantId,omitempty"`
	LogtoUserID    *string    `bson:"logtoUserId,omitempty"`
	TenantSet      bool       `bson:"tenantSet"`
	RoleAssigned   bool       `bson:"roleAssigned"`
	EventPublished bool       `bson:"eventPublished"`
	CatalogSeeded  bool       `bson:"catalogSeeded"`
	FailureReason  *string    `bson:"failureReason,omitempty"`
	RetryCount     int32      `bson:"retryCount"`
	NextRetryAt    *time.Time `bson:"nextRetryAt,omitempty"`
	CreatedAt      time.Time  `bson:"createdAt"`
	CompletedAt    *time.Time `bson:"completedAt,omitempty"`
	Version        int64      `bson:"version"`
}
