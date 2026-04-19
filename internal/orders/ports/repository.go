package ports

import (
	"context"

	"github.com/rubentanahara/ecomapi/internal/orders/domain"
)

type Repository interface {
	ListOrders(ctx context.Context) ([]domain.Order, error)
	CreateOrder(ctx context.Context, customerID int64) (domain.Order, error)
	FindProductByID(ctx context.Context, id int64) (domain.Product, error)
	CreateOrderItem(ctx context.Context, item domain.OrderItem) error
	WithTx(ctx context.Context, fn func(repo Repository) error) error
}
