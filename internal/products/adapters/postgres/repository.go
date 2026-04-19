package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	repo "github.com/rubentanahara/ecomapi/internal/adapters/postgresql/sqlc"
	"github.com/rubentanahara/ecomapi/internal/products/domain"
	"github.com/rubentanahara/ecomapi/internal/products/ports"
)

type repository struct {
	q repo.Querier
}

func NewRepository(q repo.Querier) ports.Repository {
	return &repository{q: q}
}

func (r *repository) ListProducts(ctx context.Context) ([]domain.Product, error) {
	sqlcProducts, err := r.q.ListProducts(ctx)
	if err != nil {
		return nil, err
	}

	var products []domain.Product
	for _, p := range sqlcProducts {
		products = append(products, toDomainProduct(p))
	}
	return products, nil
}

func (r *repository) GetProduct(ctx context.Context, id int64) (domain.Product, error) {
	p, err := r.q.FindProductByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Product{}, domain.ErrNotFound
		}
		return domain.Product{}, err
	}
	return toDomainProduct(p), nil
}

func toDomainProduct(p repo.Product) domain.Product {
	return domain.Product{
		ID:           p.ID,
		Name:         p.Name,
		PriceInCents: p.PriceInCents,
		Quantity:     p.Quantity,
		CreatedAt:    p.CreatedAt.Time,
	}
}
