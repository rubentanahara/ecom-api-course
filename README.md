# E-Commerce API (Go)

A high-performance, maintainable RESTful API for an e-commerce platform built with Go. This project implements a **Strict Ports and Adapters (Hexagonal Architecture)** combined with senior-level Go best practices to ensure a clean separation between business logic, external frameworks, and databases.

## 🚀 Technologies Used

- **Language:** Go 1.26
- **Routing:** [go-chi/chi](https://github.com/go-chi/chi) - lightweight, idiomatic, and composable router.
- **Logging:** `log/slog` - Structured, type-safe logging from the standard library.
- **Database:** PostgreSQL with `pgx/v5` driver.
- **Database Access:** [sqlc](https://sqlc.dev/) - generates type-safe Go code from SQL.
- **Migrations:** [goose](https://github.com/pressly/goose)
- **Live Reloading:** [air](https://github.com/air-verse/air)
- **API Client:** [Bruno](https://www.usebruno.com/) (Collections included in `/bruno`)

## 🏗 Architecture & Data Flow

The project follows **Clean Architecture** (Ports & Adapters) with a strict "Inside-Out" dependency rule:

### Layers
1.  **Domain (`internal/<feature>/domain`):** Pure business logic and entities. Contains custom domain errors (e.g., `ErrNotFound`) to decouple logic from implementation details.
2.  **Ports (`internal/<feature>/ports`):** Interfaces defining the requirements of the domain (e.g., `Repository`).
3.  **Services (`internal/<feature>/service.go`):** Orchestrates business logic. Adheres to **"Accept interfaces, return structs"** for maximum flexibility and IDE discoverability.
4.  **Adapters (`internal/<feature>/adapters`):** 
    -   **Postgres:** Implements domain ports using SQLC. Maps database-specific errors (like `pgx.ErrNoRows`) to domain-specific errors.
    -   **HTTP (Handlers):** Entry point for external requests. Uses the **Functional Options** pattern for configuration (e.g., injecting loggers).

### Request Flow
`HTTP Request` → `Middleware (Logger, RequestID, etc.)` → `Handler (Validation)` → `Service (Business Logic)` → `Repository (Data Access)` → `Database`

## 🛡️ Senior-Level Best Practices Applied

-   **Graceful Shutdown:** The server listens for `SIGINT`/`SIGTERM` and shuts down cleanly, allowing in-flight requests to finish.
-   **Structured Logging:** Every layer uses `slog` with request-scoped context, making debugging in production seamless.
-   **Explicit Dependency Injection:** No global state. All dependencies are injected via constructors.
-   **Error Decoupling:** Database errors never leak to the API consumer. The repository maps them to domain errors, which the handler then translates to appropriate HTTP status codes (404, 400, 500).
-   **Validation at the Edge:** Input validation is performed at the handler level to ensure the service layer only operates on valid data.

## ⚙️ Prerequisites

- Go 1.26+
- PostgreSQL instance running (default configured for `localhost:5432`)
- `make` utility

## 🛠 Setup & Installation

1. **Clone the repository:**
   ```bash
   git clone <repository-url>
   cd ecom-api-course
   ```

2. **Environment Variables:**
   Ensure you have a `.env` file in the root directory. Example:
   ```env
   GOOSE_DBSTRING=host=localhost port=5432 user=postgres password=postgres dbname=ecom sslmode=disable
   GOOSE_DRIVER=postgres
   GOOSE_MIGRATION_DIR=./internal/adapters/postgresql/migrations
   ```

3. **Install Dependencies:**
   ```bash
   go mod download
   ```

4. **Install Build Tools (Air & Goose):**
   ```bash
   go install github.com/air-verse/air@latest
   go install github.com/pressly/goose/v3/cmd/goose@latest
   ```

## 🗄️ Database Migrations

Use the provided `Makefile` commands to manage the database schema:

```bash
make migrate-status  # View current migration status
make migrate-up      # Apply all pending migrations
make migrate-down    # Rollback the last migration
```

## 🚀 Running the Application

### Development (Live Reload)
Starts the server on `http://localhost:8080` and reloads on file changes.
```bash
make watch
```

### Standard Run
Runs the application normally (stops any existing processes on port 8080 first).
```bash
make run
```

### Build Binary
Compiles the application into the `bin/` directory.
```bash
make build
```

## 📡 API Endpoints

| Method | Endpoint         | Description                   |
|--------|------------------|-------------------------------|
| `GET`  | `/health`        | API Health Check              |
| `GET`  | `/products`      | List all products             |
| `GET`  | `/products/{id}` | Get a specific product by ID  |
| `GET`  | `/orders`        | List all orders               |
| `POST` | `/orders`        | Place a new order             |

### Example: Place Order Payload
```json
{
  "customerId": 1,
  "items": [
    {
      "productId": 2,
      "quantity": 1
    }
  ]
}
```

