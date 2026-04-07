package mongo

import (
	"context"
	"fmt"

	commonsmongo "github.com/Sokol111/ecommerce-commons/pkg/persistence/mongo"
	"github.com/Sokol111/ecommerce-tenant-service/internal/domain/tenant"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type tenantRepository struct {
	*commonsmongo.GenericRepository[tenant.Tenant, tenantEntity]
}

func newTenantRepository(mongo commonsmongo.Mongo, mapper *tenantMapper) (tenant.Repository, error) {
	genericRepo, err := commonsmongo.NewGenericRepository(
		mongo.GetCollection("tenant"),
		mapper,
	)
	if err != nil {
		return nil, err
	}

	return &tenantRepository{GenericRepository: genericRepo}, nil
}

func (r *tenantRepository) FindByID(ctx context.Context, id string) (*tenant.Tenant, error) {
	return r.GenericRepository.FindByID(ctx, id)
}

func (r *tenantRepository) FindBySlug(ctx context.Context, slug string) (*tenant.Tenant, error) {
	filter := bson.D{{Key: "slug", Value: slug}}
	items, err := r.FindAllWithFilter(ctx, filter, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to find tenant by slug: %w", err)
	}
	if len(items) == 0 {
		return nil, commonsmongo.ErrEntityNotFound
	}
	return items[0], nil
}

func (r *tenantRepository) FindList(ctx context.Context, query tenant.ListQuery) (*commonsmongo.PageResult[tenant.Tenant], error) {
	filter := bson.D{}
	if query.Enabled != nil {
		filter = append(filter, bson.E{Key: "enabled", Value: *query.Enabled})
	}

	var sortBson bson.D
	if query.Sort != "" {
		sortOrder := 1
		if query.Order == "desc" {
			sortOrder = -1
		}
		sortBson = bson.D{{Key: query.Sort, Value: sortOrder}}
	}

	opts := commonsmongo.QueryOptions{
		Filter: filter,
		Page:   query.Page,
		Size:   query.Size,
		Sort:   sortBson,
	}

	return r.FindWithOptions(ctx, opts)
}

func (r *tenantRepository) FindEnabledSlugs(ctx context.Context) ([]string, error) {
	tenants, err := r.FindAllWithFilter(ctx, bson.D{{Key: "enabled", Value: true}}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to find enabled tenants: %w", err)
	}

	slugs := make([]string, len(tenants))
	for i, t := range tenants {
		slugs[i] = t.Slug
	}
	return slugs, nil
}

func (r *tenantRepository) Exists(ctx context.Context, slug string) (bool, error) {
	filter := bson.D{{Key: "slug", Value: slug}}
	items, err := r.FindAllWithFilter(ctx, filter, nil)
	if err != nil {
		return false, fmt.Errorf("failed to check tenant existence: %w", err)
	}
	return len(items) > 0, nil
}

func (r *tenantRepository) Delete(ctx context.Context, id string) error {
	return r.GenericRepository.Delete(ctx, id)
}
