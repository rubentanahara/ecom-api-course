package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	repo "github.com/rubentanahara/ecomapi/internal/adapters/postgresql/sqlc"
	"github.com/rubentanahara/ecomapi/internal/orders/domain"
	"github.com/rubentanahara/ecomapi/internal/orders/ports"
)

type repository struct {
	q  *repo.Queries
	db *pgx.Conn
}

func NewRepository(q *repo.Queries, db *pgx.Conn) ports.Repository {
	return &repository{q: q, db: db}
}

func (r *repository) ListOrders(ctx context.Context) ([]domain.Order, error) {
	sqlcOrders, err := r.q.ListOrders(ctx)
	if err != nil {
		return nil, err
	}
	var orders []domain.Order
	for _, o := range sqlcOrders {
		orders = append(orders, domain.Order{
			ID:         o.ID,
			CustomerID: o.CustomerID,
			CreatedAt:  o.CreatedAt.Time,
		})
	}
	return orders, nil
}

func (r *repository) CreateOrder(ctx context.Context, customerID int64) (domain.Order, error) {
	o, err := r.q.CreateOrder(ctx, customerID)
	if err != nil {
		return domain.Order{}, err
	}
	return domain.Order{
		ID:         o.ID,
		CustomerID: o.CustomerID,
		CreatedAt:  o.CreatedAt.Time,
	}, nil
}

func (r *repository) FindProductByID(ctx context.Context, id int64) (domain.Product, error) {
	p, err := r.q.FindProductByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Product{}, pgx.ErrNoRows
		}
		return domain.Product{}, err
	}
	return domain.Product{
		ID:           p.ID,
		PriceInCents: p.PriceInCents,
		Quantity:     p.Quantity,
	}, nil
}

func (r *repository) CreateOrderItem(ctx context.Context, item domain.OrderItem) error {
	_, err := r.q.CreateOrderItem(ctx, repo.CreateOrderItemParams{
		OrderID:    item.OrderID,
		ProductID:  item.ProductID,
		Quantity:   item.Quantity,
		PriceCents: item.PriceCents,
	})
	return err
}

func (r *repository) WithTx(ctx context.Context, fn func(repo ports.Repository) error) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	txRepo := &repository{
		q:  r.q.WithTx(tx),
		db: r.db,
	}

	if err := fn(txRepo); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
