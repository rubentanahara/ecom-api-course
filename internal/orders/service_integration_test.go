package orders_test

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	repo "github.com/rubentanahara/ecomapi/internal/adapters/postgresql/sqlc"
	"github.com/rubentanahara/ecomapi/internal/orders"
	ordersPostgres "github.com/rubentanahara/ecomapi/internal/orders/adapters/postgres"
	"github.com/rubentanahara/ecomapi/internal/orders/domain"
)

const (
	concurrentBuyers = 20
	initialStock     = 5
	productPrice     = 1000
)

type fixture struct {
	service    *orders.Service
	pool       *pgxpool.Pool
	productID  int64
	customerID int64
}

func newFixture(t *testing.T) fixture {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}

	var productID int64
	err = pool.QueryRow(ctx,
		"INSERT INTO products (name, price_in_cents, quantity) VALUES ('oversell-test', $1, $2) RETURNING id",
		productPrice, initialStock).Scan(&productID)
	if err != nil {
		t.Fatal(err)
	}

	customerID := time.Now().UnixNano()
	t.Cleanup(func() {
		pool.Exec(ctx, "DELETE FROM order_items WHERE order_id IN (SELECT id FROM orders WHERE customer_id = $1)", customerID)
		pool.Exec(ctx, "DELETE FROM orders WHERE customer_id = $1", customerID)
		pool.Exec(ctx, "DELETE FROM products WHERE id = $1", productID)
		pool.Close()
	})

	repository := ordersPostgres.NewRepository(repo.New(pool), pool)
	return fixture{service: orders.NewService(repository), pool: pool, productID: productID, customerID: customerID}
}

func (f fixture) stock(t *testing.T) int32 {
	t.Helper()
	var quantity int32
	err := f.pool.QueryRow(context.Background(), "SELECT quantity FROM products WHERE id = $1", f.productID).Scan(&quantity)
	if err != nil {
		t.Fatal(err)
	}
	return quantity
}

func (f fixture) order(items ...orders.OrderItemRequest) orders.CreateOrderRequest {
	return orders.CreateOrderRequest{CustomerID: f.customerID, Items: items}
}

func TestConcurrentOrdersCannotOversell(t *testing.T) {
	f := newFixture(t)

	var wg sync.WaitGroup
	results := make(chan error, concurrentBuyers)
	for range concurrentBuyers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := f.service.PlaceOrder(context.Background(), f.order(orders.OrderItemRequest{ProductID: f.productID, Quantity: 1}))
			results <- err
		}()
	}
	wg.Wait()
	close(results)

	succeeded, outOfStock := 0, 0
	for err := range results {
		switch {
		case err == nil:
			succeeded++
		case errors.Is(err, domain.ErrProductNoStock):
			outOfStock++
		default:
			t.Fatalf("unexpected error: %v", err)
		}
	}

	if succeeded != initialStock || outOfStock != concurrentBuyers-initialStock {
		t.Fatalf("succeeded=%d outOfStock=%d, want %d and %d", succeeded, outOfStock, initialStock, concurrentBuyers-initialStock)
	}
	if got := f.stock(t); got != 0 {
		t.Fatalf("stock = %d, want 0", got)
	}
}

func TestDuplicateLinesAreSummedAgainstStock(t *testing.T) {
	f := newFixture(t)

	_, err := f.service.PlaceOrder(context.Background(), f.order(
		orders.OrderItemRequest{ProductID: f.productID, Quantity: 3},
		orders.OrderItemRequest{ProductID: f.productID, Quantity: 3},
	))

	if !errors.Is(err, domain.ErrProductNoStock) {
		t.Fatalf("err = %v, want ErrProductNoStock", err)
	}
	if got := f.stock(t); got != initialStock {
		t.Fatalf("stock = %d, want %d (rolled back)", got, initialStock)
	}
}

func TestNonPositiveQuantityIsRejected(t *testing.T) {
	f := newFixture(t)

	for _, quantity := range []int32{0, -1} {
		_, err := f.service.PlaceOrder(context.Background(), f.order(orders.OrderItemRequest{ProductID: f.productID, Quantity: quantity}))
		if !errors.Is(err, domain.ErrInvalidOrder) {
			t.Fatalf("quantity %d: err = %v, want ErrInvalidOrder", quantity, err)
		}
	}
	if got := f.stock(t); got != initialStock {
		t.Fatalf("stock = %d, want %d", got, initialStock)
	}
}

func TestPlacedOrderReturnsItemsWithSnapshotPrice(t *testing.T) {
	f := newFixture(t)

	order, err := f.service.PlaceOrder(context.Background(), f.order(orders.OrderItemRequest{ProductID: f.productID, Quantity: 2}))
	if err != nil {
		t.Fatal(err)
	}

	if len(order.Items) != 1 || order.Items[0].PriceCents != productPrice || order.Items[0].Quantity != 2 {
		t.Fatalf("items = %+v", order.Items)
	}
	if got := f.stock(t); got != initialStock-2 {
		t.Fatalf("stock = %d, want %d", got, initialStock-2)
	}

	listed, err := f.service.ListOrders(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for _, o := range listed {
		if o.ID == order.ID && len(o.Items) == 1 {
			return
		}
	}
	t.Fatalf("order %d not listed with its items", order.ID)
}
