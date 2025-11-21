# 🏦 Simple Banking API

A straightforward, flexible, and maintainable RESTful service for managing accounts and financial transactions. Built with **Go** using **Hexagonal Architecture** principles.

[![Go Version](https://img.shields.io/badge/Go-1.23-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

---

## ✨ Features

- ✅ **RESTful API** with proper HTTP semantics
- ✅ **Hexagonal Architecture** (Ports & Adapters)
- ✅ **Automatic Amount Normalization** (smart transaction sign conversion)
- ✅ **Idempotency Support** (required Idempotency-Key header prevents duplicate transactions)
- ✅ **Pagination Support** for transaction lists
- ✅ **SQLite Database** with migration system
- ✅ **Docker Support** for easy deployment
- ✅ **Comprehensive Validation** with clear error messages
- ✅ **Graceful Shutdown** for production readiness
- ✅ **Easy to Extend** - flexible design for future features

---

## 🚀 Quick Start

### Option 1: Using the Run Script (Easiest)

```bash
./run
```

### Option 2: Using Make

```bash
make run
```

### Option 3: Using Docker

```bash
# Build and run with Docker Compose
make docker-run

# Or manually
docker-compose up -d
```

### Option 4: Manual Build

```bash
# Install dependencies
go mod tidy

# Build
go build -o bin/simple-banking-api ./cmd/api

# Run
./bin/simple-banking-api
```

The API will be available at: **http://localhost:8080**

---

## 📋 API Endpoints

### Health Check
```
GET /health
```

### Accounts

| Method | Endpoint | Description | Status Code |
|--------|----------|-------------|-------------|
| POST | `/v1/accounts` | Create a new account | 201 Created |
| GET | `/v1/accounts/:accountId` | Get account by ID | 200 OK |

### Transactions

| Method | Endpoint | Description | Status Code |
|--------|----------|-------------|-------------|
| POST | `/v1/transactions` | Create a new transaction | 201 Created |
| GET | `/v1/accounts/:accountId/transactions` | Get account transactions (paginated) | 200 OK |

---

## 📝 API Usage Examples

### 1. Create an Account

**Request:**
```bash
curl -X POST http://localhost:8080/v1/accounts \
  -H "Content-Type: application/json" \
  -d '{
    "document_number": "12345678900"
  }'
```

**Response (201 Created):**
```json
{
  "account_id": 1,
  "document_number": "12345678900",
  "created_at": "2025-11-16T14:36:39Z"
}
```

---

### 2. Get Account Information

**Request:**
```bash
curl -X GET http://localhost:8080/v1/accounts/1
```

**Response (200 OK):**
```json
{
  "account_id": 1,
  "document_number": "12345678900",
  "created_at": "2025-11-16T14:36:39Z"
}
```

---

### 3. Create a Transaction

**Request:**
```bash
curl -X POST http://localhost:8080/v1/transactions \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: unique-key-123" \
  -d '{
    "account_id": 1,
    "operation_type_id": 4,
    "amount": 123.45
  }'
```

**Response (201 Created):**
```json
{
  "transaction_id": 1,
  "account_id": 1,
  "operation_type_id": 4,
  "amount": 123.45,
  "event_date": "2025-11-16T14:37:03Z"
}
```

**Note:** The `Idempotency-Key` header is **required** and prevents duplicate processing if the same request is sent multiple times with the same key.

---

### 4. Get Account Transactions (Paginated)

**Request:**
```bash
# Get first 50 transactions (default)
curl -X GET http://localhost:8080/v1/accounts/1/transactions

# Get with custom pagination
curl -X GET "http://localhost:8080/v1/accounts/1/transactions?limit=10&offset=20"
```

**Response (200 OK):**
```json
{
  "transactions": [
    {
      "transaction_id": 1,
      "account_id": 1,
      "operation_type_id": 4,
      "amount": 123.45,
      "event_date": "2025-11-16T14:37:03Z"
    },
    {
      "transaction_id": 2,
      "account_id": 1,
      "operation_type_id": 1,
      "amount": -50.00,
      "event_date": "2025-11-16T15:20:11Z"
    }
  ],
  "pagination": {
    "limit": 50,
    "offset": 0,
    "total": 2,
    "total_pages": 1
  }
}
```

**Query Parameters:**
- `limit` (optional): Number of items per page (default: 50, max: 100)
- `offset` (optional): Number of items to skip (default: 0)

---

## 💡 Automatic Amount Normalization

The API **automatically normalizes transaction amounts** based on the operation type:

| Operation Type ID | Description | Amount Conversion |
|-------------------|-------------|-------------------|
| 1 | Normal Purchase | Converts to **negative** |
| 2 | Purchase with Installments | Converts to **negative** |
| 3 | Withdrawal | Converts to **negative** |
| 4 | Credit Voucher | Converts to **positive** |

### Examples:

**Purchase (Type 1):**
```json
// You send:
{"account_id": 1, "operation_type_id": 1, "amount": 50.0}

// Stored as:
{"amount": -50.0}  ✅ Auto-converted to negative!
```

**Credit Voucher (Type 4):**
```json
// You send:
{"account_id": 1, "operation_type_id": 4, "amount": 100.0}

// Stored as:
{"amount": 100.0}  ✅ Stays positive!
```

**Even with wrong sign:**
```json
// You send (negative for credit):
{"account_id": 1, "operation_type_id": 4, "amount": -100.0}

// Stored as:
{"amount": 100.0}  ✅ Corrected to positive!
```

---

## 🧪 Running Tests

### Automated Test Suite

Run the comprehensive test suite that validates all endpoints and scenarios:

```bash
./run-local/test.sh
```

This will test:
- ✅ All success scenarios
- ✅ All validation scenarios
- ✅ Error handling
- ✅ Amount normalization
- ✅ Edge cases

Results are saved to `test-results.log`.

### Unit Tests

```bash
make test
```

---

## 🐳 Docker Commands

```bash
# Build Docker image
make docker-build

# Start container
make docker-run

# Stop container
make docker-stop

# View logs
make docker-logs

# Restart container
make docker-restart
```

---

## 🏗️ Architecture

This project follows **Hexagonal Architecture** (Ports & Adapters) with a clean separation of concerns:

### Architecture Layers

**Handler → Processor → Repository**

1. **Handlers** (HTTP Layer): Lightweight, focused on HTTP concerns
   - Decode JSON requests
   - Validate input
   - Delegate to processors
   - Return HTTP responses

2. **Processors** (Business Logic): Orchestrate domain operations
   - One processor per operation (divide to conquer)
   - Coordinate repository calls
   - Apply business rules
   - Handle domain logic

3. **Repositories** (Data Access): Interact with database
   - Implement port interfaces
   - Execute SQL queries
   - Map between database and domain models

### Benefits

- ✅ **Clear separation of concerns**: Each layer has a single responsibility
- ✅ **Easy to test**: Layers can be tested independently
- ✅ **Easy to extend**: Add new operations by creating new processor + handler
- ✅ **Flexible**: Can swap implementations (e.g., SQLite → PostgreSQL)
- ✅ **Maintainable**: "Divide to conquer" - one file per operation

```
simple-banking-api/
├── cmd/
│   └── api/
│       ├── main.go              # Application entry point
│       ├── config.go            # Configuration loading
│       └── server.go            # Application setup & Dependency Injection (DI)
├── internal/
│   ├── core/                    # Business Logic (Domain)
│   │   ├── domain/              # Domain entities & DTOs
│   │   ├── ports/               # Repository interfaces
│   │   └── services/
│   │       └── processors/      # Business logic processors
│   ├── adapters/
│   │   └── repository/          # Repository implementations
│   └── server/                  # HTTP Layer
│       ├── handlers/            # HTTP handlers (lightweight)
│       ├── middleware/          # Middleware (idempotency, etc)
│       └── router.go            # Route configuration
├── infra/
│   └── database/                # Infrastructure
│       ├── connection.go        # Database connection
│       └── migrations.go        # Schema migrations
├── docs/                        # PlantUML diagrams
├── run-local/                   # Test scripts
├── Dockerfile                   # Docker configuration
├── docker-compose.yml           # Docker Compose
├── Makefile                     # Build automation
└── README.md                    # This file
```

### Diagrams

View the PlantUML diagrams in the `docs/sequence-diagrams/` folder:
- `architecture.puml` - Overall system architecture
- `endpoint-create-account.puml` - Create account flow
- `endpoint-get-account.puml` - Get account flow
- `endpoint-create-transaction.puml` - Create transaction flow (with normalization)
- `endpoint-get-transactions.puml` - Get account transactions flow (with pagination)

### Current Implementation

**Implemented Endpoints:**
- ✅ POST `/v1/accounts` - Create account
- ✅ GET `/v1/accounts/:accountId` - Get account by ID
- ✅ POST `/v1/transactions` - Create transaction with automatic amount normalization
- ✅ GET `/v1/accounts/:accountId/transactions` - Get account transactions with pagination

**Processors:**
- `create_account_processor.go` - Handles account creation
- `get_account_processor.go` - Handles account retrieval
- `create_transaction_processor.go` - Handles transaction creation with amount normalization
- `get_transactions_processor.go` - Handles paginated transaction retrieval

**Handlers:**
- `create_account_handler.go` - Validates and delegates account creation
- `get_account_handler.go` - Validates and delegates account retrieval
- `create_transaction_handler.go` - Validates and delegates transaction creation
- `get_transactions_handler.go` - Validates and delegates paginated transaction retrieval

---

## 🛠️ Technology Stack

- **Language:** Go 1.23
- **Router:** Chi v5 (lightweight, fast, RESTful)
- **Database:** SQLite (modernc.org/sqlite - pure Go)
- **Architecture:** Hexagonal (Ports & Adapters)
- **Containerization:** Docker & Docker Compose

### Database Considerations

**SQLite is used for simplicity and portability:**
- ✅ Zero configuration, embedded database
- ✅ Perfect for development and testing
- ✅ Easy deployment (single file)

**Production Considerations:**
- ⚠️ **Write Lock Limitation**: SQLite uses database-level write locking, which can limit concurrent write operations
- ⚠️ **Scalability**: For high-concurrency production environments, consider PostgreSQL or MySQL
- ✅ **Easy Migration**: The hexagonal architecture makes switching databases straightforward (just implement a new repository adapter)

---

## 🔧 Configuration

Configuration via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_ADDRESS` | `:8080` | Server listen address |
| `DATABASE_PATH` | `./data/banking.db` | SQLite database file path |

---

## 📊 Database Schema

### Tables

**accounts**
- `id` (INTEGER, PK, AUTO_INCREMENT)
- `document_number` (TEXT, UNIQUE)
- `created_at` (DATETIME)

**transactions**
- `id` (INTEGER, PK, AUTO_INCREMENT)
- `account_id` (INTEGER, FK → accounts.id)
- `operation_type_id` (INTEGER, FK → operation_types.id)
- `amount` (REAL)
- `event_date` (DATETIME)
- `created_at` (DATETIME)

**operation_types** (Seeded Data)
- `id` (INTEGER, PK)
- `description` (TEXT)
- `created_at` (DATETIME)

**schema_migrations** (Version Control)
- `version` (INTEGER, PK)
- `description` (TEXT)
- `applied_at` (DATETIME)

---

## 🎯 Design Principles

### 1. **Maintainability**
- Clean separation of concerns
- Single Responsibility Principle
- Clear package structure

### 2. **Simplicity**
- No over-engineering
- Straightforward code flow
- Clear naming conventions

### 3. **Testability**
- Interface-based design
- Dependency injection
- Mock-friendly architecture

### 4. **Flexibility**
- Easy to add new features
- Migration-based schema evolution
- Pluggable components
---

## 🔮 Future Enhancements

The architecture is designed to easily accommodate:

- ✅ Add new columns (use migration system)
- ✅ Add database JOINs (SQL queries are modular)
- ✅ Add new endpoints (handler pattern)
- ✅ Add authentication/authorization
- ✅ Add pagination
- ✅ Add filtering and sorting
- ✅ Switch to PostgreSQL (implement new adapter)
- ✅ Add GraphQL layer
- ✅ Add event sourcing

### Example: Adding a New Column

1. Add migration in `infra/database/migrations.go`:
```go
{
    Version: 2,
    Description: "Add merchant_id to transactions",
    SQL: `
        ALTER TABLE transactions ADD COLUMN merchant_id INTEGER;
        CREATE INDEX idx_transactions_merchant_id ON transactions(merchant_id);
    `,
}
```

2. Update domain model in `internal/core/domain/transaction.go`
3. Update repository scanning in `internal/adapters/repository/transaction_repository.go`
4. Done! ✨

### Example: Adding a New Endpoint

1. Add Request/Response DTOs to `internal/core/domain/`
2. Create processor in `internal/core/services/processors/your_operation_processor.go`
3. Create handler in `internal/server/handlers/your_operation_handler.go`
4. Wire up in `internal/server/router.go` and `cmd/api/main.go`
5. Done! ✨

---

## 📖 Make Commands

```bash
make help              # Show all available commands
make build             # Build the application
make run               # Build and run
make test              # Run tests with coverage
make clean             # Clean build artifacts
make docker-build      # Build Docker image
make docker-run        # Start Docker container
make docker-stop       # Stop Docker container
make docker-logs       # View container logs
make lint              # Run linter
make format            # Format code
```

---

## 🤝 Contributing

This project is designed for a technical interview test practice and learning. Feel free to:
- Fork and experiment
- Add new features
- Improve existing code
- Share feedback

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 👨‍💻 Author

Built with ❤️ for a technical interview practice test.

**Focus Areas:**
- ✅ Clean Architecture
- ✅ RESTful API Design
- ✅ Go Best Practices
- ✅ Comprehensive Documentation

---

## 🙏 Acknowledgments

- Go community for excellent tooling
- Chi router for clean RESTful routing
- Hexagonal Architecture pattern for flexibility
