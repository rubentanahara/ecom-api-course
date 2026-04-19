package orders

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/rubentanahara/ecomapi/internal/json"
	"github.com/rubentanahara/ecomapi/internal/orders/domain"
)

type OrderService interface {
	PlaceOrder(ctx context.Context, req CreateOrderRequest) (domain.Order, error)
	ListOrders(ctx context.Context) ([]domain.Order, error)
}

type handler struct {
	service OrderService
	logger  *slog.Logger
}

type HandlerOption func(*handler)

func WithLogger(logger *slog.Logger) HandlerOption {
	return func(h *handler) {
		h.logger = logger
	}
}

func NewHandler(service OrderService, opts ...HandlerOption) *handler {
	h := &handler{
		service: service,
		logger:  slog.Default(),
	}

	for _, opt := range opts {
		opt(h)
	}

	return h
}

func (h *handler) ListOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := h.service.ListOrders(r.Context())
	if err != nil {
		h.logger.Error("Failed to list orders", "error", err)
		json.Write(w, http.StatusInternalServerError, map[string]string{"error": "Failed to list orders"})
		return
	}

	json.Write(w, http.StatusOK, orders)
}

func (h *handler) PlaceOrder(w http.ResponseWriter, r *http.Request) {
	var req CreateOrderRequest
	if err := json.Read(r, &req); err != nil {
		h.logger.Warn("Failed to read request body", "error", err)
		json.Write(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	// Basic validation
	if req.CustomerID == 0 {
		json.Write(w, http.StatusBadRequest, map[string]string{"error": "customer ID is required"})
		return
	}
	if len(req.Items) == 0 {
		json.Write(w, http.StatusBadRequest, map[string]string{"error": "at least one item is required"})
		return
	}

	createdOrder, err := h.service.PlaceOrder(r.Context(), req)
	if err != nil {
		if errors.Is(err, ErrProductNotFound) {
			json.Write(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}

		if errors.Is(err, ErrProductNoStock) {
			json.Write(w, http.StatusConflict, map[string]string{"error": err.Error()})
			return
		}

		h.logger.Error("Failed to place order", "error", err)
		json.Write(w, http.StatusInternalServerError, map[string]string{"error": "Failed to place order"})
		return
	}

	json.Write(w, http.StatusCreated, createdOrder)
}
