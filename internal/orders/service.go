package orders

import (
	"context"
	"fmt"
	"sort"

	"github.com/rubentanahara/ecomapi/internal/orders/domain"
	"github.com/rubentanahara/ecomapi/internal/orders/ports"
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
	items, err := normalizeItems(req)
	if err != nil {
		return domain.Order{}, err
	}

	var createdOrder domain.Order

	err = s.repo.WithTx(ctx, func(txRepo ports.Repository) error {
		order, err := txRepo.CreateOrder(ctx, req.CustomerID)
		if err != nil {
			return fmt.Errorf("failed to create order: %w", err)
		}

		for _, item := range items {
			price, err := txRepo.ReserveStock(ctx, item.ProductID, item.Quantity)
			if err != nil {
				return fmt.Errorf("product %d: %w", item.ProductID, err)
			}

			created, err := txRepo.CreateOrderItem(ctx, domain.OrderItem{
				OrderID:    order.ID,
				ProductID:  item.ProductID,
				Quantity:   item.Quantity,
				PriceCents: price,
			})
			if err != nil {
				return fmt.Errorf("failed to create order item: %w", err)
			}
			order.Items = append(order.Items, created)
		}
		createdOrder = order
		return nil
	})
	if err != nil {
		return domain.Order{}, err
	}

	return createdOrder, nil
}

func normalizeItems(req CreateOrderRequest) ([]OrderItemRequest, error) {
	if req.CustomerID == 0 {
		return nil, fmt.Errorf("%w: customer ID is required", domain.ErrInvalidOrder)
	}
	if len(req.Items) == 0 {
		return nil, fmt.Errorf("%w: at least one item is required", domain.ErrInvalidOrder)
	}

	quantityByProduct := make(map[int64]int32, len(req.Items))
	for _, item := range req.Items {
		if item.Quantity <= 0 {
			return nil, fmt.Errorf("%w: quantity for product %d must be positive", domain.ErrInvalidOrder, item.ProductID)
		}
		quantityByProduct[item.ProductID] += item.Quantity
	}

	items := make([]OrderItemRequest, 0, len(quantityByProduct))
	for productID, quantity := range quantityByProduct {
		items = append(items, OrderItemRequest{ProductID: productID, Quantity: quantity})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ProductID < items[j].ProductID })
	return items, nil
}
