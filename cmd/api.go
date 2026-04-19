package main

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5"
	repo "github.com/rubentanahara/ecomapi/internal/adapters/postgresql/sqlc"
	"github.com/rubentanahara/ecomapi/internal/orders"
	ordersPostgres "github.com/rubentanahara/ecomapi/internal/orders/adapters/postgres"
	"github.com/rubentanahara/ecomapi/internal/products"
	productsPostgres "github.com/rubentanahara/ecomapi/internal/products/adapters/postgres"
)

type application struct {
	configuration configuration
	database      *pgx.Conn
}

func (app *application) mount() http.Handler {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Use(middleware.Timeout(60 * time.Second))

	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	productRepo := productsPostgres.NewRepository(repo.New(app.database))
	productService := products.NewService(productRepo)
	productsHandler := products.NewHandler(productService, products.WithLogger(slog.Default()))
	router.Get("/products", productsHandler.ListProducts)
	router.Get("/products/{id}", productsHandler.GetProduct)

	ordersRepo := ordersPostgres.NewRepository(repo.New(app.database), app.database)
	ordersService := orders.NewService(ordersRepo)
	ordersHandler := orders.NewHandler(ordersService, orders.WithLogger(slog.Default()))
	router.Post("/orders", ordersHandler.PlaceOrder)
	router.Get("/orders", ordersHandler.ListOrders)

	return router
}

func (app *application) run(ctx context.Context, handler http.Handler) error {
	server := &http.Server{
		Addr:         app.configuration.address,
		Handler:      handler,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	errChan := make(chan error, 1)

	go func() {
		slog.Info("Starting server", "address", app.configuration.address)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		slog.Info("Shutting down server gracefully...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		return server.Shutdown(shutdownCtx)
	}
}

type configuration struct {
	address        string
	databaseConfig databaseConfig
}

type databaseConfig struct {
	dsn string
}
