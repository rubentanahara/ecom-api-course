package domain

import "time"

type Order struct {
	ID         int64       `json:"id"`
	CustomerID int64       `json:"customer_id"`
	CreatedAt  time.Time   `json:"created_at"`
	Items      []OrderItem `json:"items,omitempty"`
}

type OrderItem struct {
	ID         int64 `json:"id"`
	OrderID    int64 `json:"order_id"`
	ProductID  int64 `json:"product_id"`
	Quantity   int32 `json:"quantity"`
	PriceCents int32 `json:"price_cents"`
}

// Product in orders domain context
type Product struct {
	ID           int64 `json:"id"`
	PriceInCents int32 `json:"price_in_cents"`
	Quantity     int32 `json:"quantity"`
}
