package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rubentanahara/ecomapi/internal/products"
)

type application struct {
	configuration configuration
	// logger
	// db driver
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

	productsHandler := products.NewHandler(nil)
	router.Get("/products", productsHandler.ListProducts)

	return router
}

func (app *application) run(handler http.Handler) error {
	server := &http.Server{
		Addr:         app.configuration.address,
		Handler:      handler,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Starting server on %s ", app.configuration.address)

	return server.ListenAndServe()
}

type configuration struct {
	address        string
	databaseConfig databaseConfig
}

type databaseConfig struct {
	dsn string
}
