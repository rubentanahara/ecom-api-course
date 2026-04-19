package domain

import (
	"errors"
	"time"
)

var ErrNotFound = errors.New("product not found")

type Product struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	PriceInCents int32     `json:"price_in_cents"`
	Quantity     int32     `json:"quantity"`
	CreatedAt    time.Time `json:"created_at"`
}
