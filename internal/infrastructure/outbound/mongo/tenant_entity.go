package mongo

import "time"

type tenantEntity struct {
	ID          string    `bson:"_id"`
	Slug        string    `bson:"slug"`
	Version     int64     `bson:"version"`
	Name        string    `bson:"name"`
	Enabled     bool      `bson:"enabled"`
	OwnerUserID string    `bson:"ownerUserId"`
	CreatedAt   time.Time `bson:"createdAt"`
	ModifiedAt  time.Time `bson:"modifiedAt"`
}
