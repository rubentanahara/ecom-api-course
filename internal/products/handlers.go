package products

import (
	"encoding/json"
	"net/http"
)

type handler struct {
	service Service
}

func NewHandler(service Service) *handler {
	return &handler{
		service: service,
	}
}

func (h *handler) ListProducts(writer http.ResponseWriter, request *http.Request) {
	// 1. Call the service to get the list of products
	// 2. Return Json in a HTTP response

	products := []string{"Product 1", "Product 2", "Product 3"}

	json.NewEncoder(writer).Encode(products)

}
