package products

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rubentanahara/ecomapi/internal/json"
	"github.com/rubentanahara/ecomapi/internal/products/domain"
)

type ProductService interface {
	ListProducts(ctx context.Context) ([]domain.Product, error)
	GetProduct(ctx context.Context, id int64) (domain.Product, error)
}

type handler struct {
	service ProductService
	logger  *slog.Logger
}

type HandlerOption func(*handler)

func WithLogger(logger *slog.Logger) HandlerOption {
	return func(h *handler) {
		h.logger = logger
	}
}

func NewHandler(service ProductService, opts ...HandlerOption) *handler {
	h := &handler{
		service: service,
		logger:  slog.Default(),
	}

	for _, opt := range opts {
		opt(h)
	}

	return h
}

func (h *handler) ListProducts(w http.ResponseWriter, r *http.Request) {
	products, err := h.service.ListProducts(r.Context())

	if err != nil {
		h.logger.Error("Error listing products", "error", err)
		json.Write(w, http.StatusInternalServerError, map[string]string{"error": "Failed to list products"})
		return
	}

	json.Write(w, http.StatusOK, products)
}

func (h *handler) GetProduct(w http.ResponseWriter, r *http.Request) {
	id, err := json.ParseInt64(chi.URLParam(r, "id"))
	if err != nil {
		h.logger.Warn("Invalid product ID", "id", chi.URLParam(r, "id"))
		json.Write(w, http.StatusBadRequest, map[string]string{"error": "Invalid product ID"})
		return
	}

	product, err := h.service.GetProduct(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			json.Write(w, http.StatusNotFound, map[string]string{"error": "Product not found"})
			return
		}

		h.logger.Error("Error getting product", "error", err, "id", id)
		json.Write(w, http.StatusInternalServerError, map[string]string{"error": "Failed to get product"})
		return
	}

	json.Write(w, http.StatusOK, product)
}
