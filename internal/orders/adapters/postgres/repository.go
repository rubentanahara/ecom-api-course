package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	repo "github.com/rubentanahara/ecomapi/internal/adapters/postgresql/sqlc"
	"github.com/rubentanahara/ecomapi/internal/orders/domain"
	"github.com/rubentanahara/ecomapi/internal/orders/ports"
)

type repository struct {
	q  *repo.Queries
	db *pgxpool.Pool
}

func NewRepository(q *repo.Queries, db *pgxpool.Pool) ports.Repository {
	return &repository{q: q, db: db}
}

func (r *repository) ListOrders(ctx context.Context) ([]domain.Order, error) {
	sqlcOrders, err := r.q.ListOrders(ctx)
	if err != nil {
		return nil, err
	}
	sqlcItems, err := r.q.ListOrderItems(ctx)
	if err != nil {
		return nil, err
	}

	itemsByOrder := make(map[int64][]domain.OrderItem)
	for _, i := range sqlcItems {
		itemsByOrder[i.OrderID] = append(itemsByOrder[i.OrderID], toDomainOrderItem(i))
	}

	orders := make([]domain.Order, 0, len(sqlcOrders))
	for _, o := range sqlcOrders {
		order := toDomainOrder(o)
		order.Items = itemsByOrder[o.ID]
		orders = append(orders, order)
	}
	return orders, nil
}

func (r *repository) CreateOrder(ctx context.Context, customerID int64) (domain.Order, error) {
	o, err := r.q.CreateOrder(ctx, customerID)
	if err != nil {
		return domain.Order{}, err
	}
	return toDomainOrder(o), nil
}

func (r *repository) ReserveStock(ctx context.Context, productID int64, quantity int32) (int32, error) {
	price, err := r.q.DecrementProductStock(ctx, repo.DecrementProductStockParams{ID: productID, Quantity: quantity})
	if err == nil {
		return price, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return 0, err
	}

	_, err = r.q.FindProductByID(ctx, productID)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, domain.ErrProductNotFound
	}
	if err != nil {
		return 0, err
	}
	return 0, domain.ErrProductNoStock
}

func (r *repository) CreateOrderItem(ctx context.Context, item domain.OrderItem) (domain.OrderItem, error) {
	created, err := r.q.CreateOrderItem(ctx, repo.CreateOrderItemParams{
		OrderID:    item.OrderID,
		ProductID:  item.ProductID,
		Quantity:   item.Quantity,
		PriceCents: item.PriceCents,
	})
	if err != nil {
		return domain.OrderItem{}, err
	}
	return toDomainOrderItem(created), nil
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

func toDomainOrder(o repo.Order) domain.Order {
	return domain.Order{
		ID:         o.ID,
		CustomerID: o.CustomerID,
		CreatedAt:  o.CreatedAt.Time,
	}
}

func toDomainOrderItem(i repo.OrderItem) domain.OrderItem {
	return domain.OrderItem{
		ID:         i.ID,
		OrderID:    i.OrderID,
		ProductID:  i.ProductID,
		Quantity:   i.Quantity,
		PriceCents: i.PriceCents,
	}
}
