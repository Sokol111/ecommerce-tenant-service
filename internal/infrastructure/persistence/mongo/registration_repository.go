package mongo

import (
	"context"
	"errors"
	"fmt"
	"time"

	commonsmongo "github.com/Sokol111/ecommerce-commons/pkg/persistence/mongo"
	"github.com/Sokol111/ecommerce-tenant-service/internal/domain/registration"
	"go.mongodb.org/mongo-driver/v2/bson"
	mongodriver "go.mongodb.org/mongo-driver/v2/mongo"
)

type registrationRepository struct {
	*commonsmongo.GenericRepository[registration.Registration, registrationEntity]
}

func newRegistrationRepository(mongo commonsmongo.Mongo, mapper *registrationMapper) (registration.Repository, error) {
	genericRepo, err := commonsmongo.NewGenericRepository(mongo, "registration", mapper)
	if err != nil {
		return nil, err
	}
	return &registrationRepository{GenericRepository: genericRepo}, nil
}

func (r *registrationRepository) Insert(ctx context.Context, reg *registration.Registration) error {
	err := r.GenericRepository.Insert(ctx, reg)
	if err != nil {
		if mongodriver.IsDuplicateKeyError(err) {
			return registration.ErrRegistrationAlreadyExists
		}
		return err
	}
	return nil
}

func (r *registrationRepository) Update(ctx context.Context, reg *registration.Registration) error {
	updated, err := r.GenericRepository.Update(ctx, reg)
	if err != nil {
		return err
	}
	reg.Version = updated.Version
	return nil
}

func (r *registrationRepository) FindBySlug(ctx context.Context, slug string) (*registration.Registration, error) {
	result, err := r.FindOneByFilter(ctx, bson.D{{Key: "slug", Value: slug}})
	if err != nil {
		if errors.Is(err, commonsmongo.ErrEntityNotFound) {
			return nil, registration.ErrRegistrationNotFound
		}
		return nil, err
	}
	return result, nil
}

func (r *registrationRepository) FindActionable(ctx context.Context) ([]*registration.Registration, error) {
	now := time.Now().UTC()

	filter := bson.D{
		{Key: "status", Value: bson.M{
			"$in": []string{
				string(registration.StatusProvisioning),
				string(registration.StatusCompensating),
			},
		}},
		{Key: "$or", Value: bson.A{
			bson.D{{Key: "nextRetryAt", Value: nil}},
			bson.D{{Key: "nextRetryAt", Value: bson.M{"$exists": false}}},
			bson.D{{Key: "nextRetryAt", Value: bson.M{"$lte": now}}},
		}},
	}

	return r.FindAllWithFilter(ctx, filter, nil)
}

func (r *registrationRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	exists, err := r.ExistsWithFilter(ctx, bson.D{{Key: "slug", Value: slug}})
	if err != nil {
		return false, fmt.Errorf("failed to check registration existence: %w", err)
	}
	return exists, nil
}
