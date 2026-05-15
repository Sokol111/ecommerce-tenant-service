package tenant

import (
	"context"

	commonsmongo "github.com/Sokol111/ecommerce-commons/pkg/persistence/mongo"
)

type ListQuery struct {
	Page    int
	Size    int
	Enabled *bool
	Sort    string
	Order   string
}

type Repository interface {
	Insert(ctx context.Context, tenant *Tenant) error
	FindByID(ctx context.Context, id string) (*Tenant, error)
	FindBySlug(ctx context.Context, slug string) (*Tenant, error)
	FindList(ctx context.Context, query ListQuery) (*commonsmongo.PageResult[Tenant], error)
	FindEnabledSlugs(ctx context.Context) ([]string, error)
	Update(ctx context.Context, tenant *Tenant) (*Tenant, error)
	Delete(ctx context.Context, id string) error
	Exists(ctx context.Context, slug string) (bool, error)
}
