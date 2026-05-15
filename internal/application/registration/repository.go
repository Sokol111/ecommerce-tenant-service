package registration

import "context"

type Repository interface {
	Insert(ctx context.Context, reg *Registration) error
	Update(ctx context.Context, reg *Registration) error
	FindBySlug(ctx context.Context, slug string) (*Registration, error)
	FindActionable(ctx context.Context) ([]*Registration, error)
	ExistsBySlug(ctx context.Context, slug string) (bool, error)
}
