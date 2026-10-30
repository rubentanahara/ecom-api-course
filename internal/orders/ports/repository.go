package ports

import (
	"context"

	"github.com/rubentanahara/ecomapi/internal/orders/domain"
)

type Repository interface {
	ListOrders(ctx context.Context) ([]domain.Order, error)
	CreateOrder(ctx context.Context, customerID int64) (domain.Order, error)
	ReserveStock(ctx context.Context, productID int64, quantity int32) (priceCents int32, err error)
	CreateOrderItem(ctx context.Context, item domain.OrderItem) (domain.OrderItem, error)
	WithTx(ctx context.Context, fn func(repo Repository) error) error
}
