package ports

import (
	"context"

	"github.com/rubentanahara/ecomapi/internal/products/domain"
)

type Repository interface {
	ListProducts(ctx context.Context) ([]domain.Product, error)
	GetProduct(ctx context.Context, id int64) (domain.Product, error)
}
