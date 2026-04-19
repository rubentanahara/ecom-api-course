package products

import (
	"context"

	"github.com/rubentanahara/ecomapi/internal/products/domain"
	"github.com/rubentanahara/ecomapi/internal/products/ports"
)

type Service struct {
	repo ports.Repository
}

func NewService(repo ports.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListProducts(ctx context.Context) ([]domain.Product, error) {
	return s.repo.ListProducts(ctx)
}

func (s *Service) GetProduct(ctx context.Context, id int64) (domain.Product, error) {
	return s.repo.GetProduct(ctx, id)
}
