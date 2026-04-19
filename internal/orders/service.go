package orders

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/rubentanahara/ecomapi/internal/orders/domain"
	"github.com/rubentanahara/ecomapi/internal/orders/ports"
)

var (
	ErrProductNotFound = errors.New("product not found")
	ErrProductNoStock  = errors.New("product has not enough stock")
)

type Service struct {
	repo ports.Repository
}

func NewService(repo ports.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListOrders(ctx context.Context) ([]domain.Order, error) {
	orders, err := s.repo.ListOrders(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list orders: %w", err)
	}
	return orders, nil
}

func (s *Service) PlaceOrder(ctx context.Context, req CreateOrderRequest) (domain.Order, error) {
	if req.CustomerID == 0 {
		return domain.Order{}, fmt.Errorf("customer ID is required")
	}
	if len(req.Items) == 0 {
		return domain.Order{}, fmt.Errorf("at least one item is required")
	}

	var createdOrder domain.Order

	err := s.repo.WithTx(ctx, func(txRepo ports.Repository) error {
		order, err := txRepo.CreateOrder(ctx, req.CustomerID)
		if err != nil {
			return fmt.Errorf("failed to create order: %w", err)
		}

		for _, item := range req.Items {
			product, err := txRepo.FindProductByID(ctx, item.ProductID)
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					return fmt.Errorf("failed to find product %d: %w", item.ProductID, ErrProductNotFound)
				}
				return fmt.Errorf("failed to find product %d: %w", item.ProductID, err)
			}

			if product.Quantity < item.Quantity {
				return fmt.Errorf("product %d: %w", item.ProductID, ErrProductNoStock)
			}

			err = txRepo.CreateOrderItem(ctx, domain.OrderItem{
				OrderID:    order.ID,
				ProductID:  item.ProductID,
				Quantity:   item.Quantity,
				PriceCents: product.PriceInCents,
			})
			if err != nil {
				return fmt.Errorf("failed to create order item: %w", err)
			}
		}
		createdOrder = order
		return nil
	})

	if err != nil {
		return domain.Order{}, err
	}

	return createdOrder, nil
}
