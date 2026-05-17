package tenant

import (
	"context"
	"fmt"
)

type GetListQuery struct {
	Page    int
	Size    int
	Enabled *bool
	Sort    string
	Order   string
}

type ListResult struct {
	Items []*Tenant
	Page  int
	Size  int
	Total int64
}

type GetListHandler interface {
	Handle(ctx context.Context, query GetListQuery) (*ListResult, error)
}

type getListHandler struct {
	repo Repository
}

func NewGetListHandler(repo Repository) GetListHandler {
	return &getListHandler{repo: repo}
}

func (h *getListHandler) Handle(ctx context.Context, query GetListQuery) (*ListResult, error) {
	listQuery := ListQuery{
		Page:    query.Page,
		Size:    query.Size,
		Enabled: query.Enabled,
		Sort:    query.Sort,
		Order:   query.Order,
	}

	result, err := h.repo.FindList(ctx, listQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenants list: %w", err)
	}

	return &ListResult{
		Items: result.Items,
		Page:  result.Page,
		Size:  result.Size,
		Total: result.Total,
	}, nil
}
