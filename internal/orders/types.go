package orders

type OrderItemRequest struct {
	ProductID int64 `json:"productId"`
	Quantity  int32 `json:"quantity"`
}

type CreateOrderRequest struct {
	CustomerID int64              `json:"customerId"`
	Items      []OrderItemRequest `json:"items"`
}
