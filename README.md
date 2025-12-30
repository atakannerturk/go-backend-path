# Go Backend Project - Production Ready

A complete, production-ready Go backend application with:
- ✅ RESTful API with Chi router
- ✅ JWT authentication & role-based authorization
- ✅ Concurrent transaction processing with worker pools
- ✅ Thread-safe balance operations
- ✅ MySQL database with migrations
- ✅ Prometheus metrics & Grafana dashboards
- ✅ Docker containerization
- ✅ Rate limiting & CORS
- ✅ Comprehensive logging
- ✅ Graceful shutdown

## Quick Start

### Using Docker Compose (Recommended)

```bash
# Start all services (app, MySQL, Prometheus, Grafana)
docker-compose up -d

# View logs
docker-compose logs -f app

# Stop all services
docker-compose down
```

The application will be available at:
- **API**: http://localhost:8080
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000 (admin/admin)

### Local Development

1. **Start MySQL**
```bash
make docker-up
# or
docker-compose up -d mysql
```

2. **Set environment variables**
```bash
cp .env.example .env
# Edit .env with your settings
```

3. **Run the application**
```bash
make run
# or
go run cmd/server/main.go
```

## API Endpoints

See [API.md](API.md) for complete API documentation.

### Quick Examples

**Register a new user:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "johndoe",
    "email": "john@example.com",
    "password": "password123"
  }'
```

**Login:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "password123"
  }'
```

**Check balance:**
```bash
curl http://localhost:8080/api/v1/balances/current \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Transfer funds:**
```bash
curl -X POST http://localhost:8080/api/v1/transactions/transfer \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "to_user_id": 2,
    "amount": 50.00
  }'
```

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                     HTTP Layer                           │
│     Chi Router + Middleware (Auth, CORS, Metrics)       │
└─────────────────────────┬───────────────────────────────┘
                          │
┌─────────────────────────▼───────────────────────────────┐
│                    Service Layer                         │
│  UserService │ TransactionService │ BalanceService      │
└─────────────────────────┬───────────────────────────────┘
                          │
┌─────────────────────────▼───────────────────────────────┐
│                   Repository Layer                       │
│     MySQL Implementations (User, Tx, Balance, Audit)    │
└─────────────────────────┬───────────────────────────────┘
                          │
┌─────────────────────────▼───────────────────────────────┐
│                    MySQL Database                        │
└──────────────────────────────────────────────────────────┘

              Worker Pool (Concurrent Processing)
┌──────────────────────────────────────────────────────────┐
│  Worker 1 │ Worker 2 │ ... │ Worker N                   │
│              Transaction Queue (Channel)                 │
└──────────────────────────────────────────────────────────┘
```

## Features

### Authentication & Authorization
- JWT token-based authentication
- Role-based access control (user, admin)
- Password hashing with bcrypt
- Token expiration (24 hours)

### Transaction Processing
- Credit/Debit/Transfer operations
- Asynchronous processing with worker pool
- State machine (pending → processing → completed/failed)
- Rollback mechanism for completed transactions
- ACID compliance

### Performance
- Worker pool with configurable size (default: 10)
- Connection pooling (25 max open, 5 max idle)
- In-memory balance caching
- Database indexes on frequent queries
- Rate limiting (100 req/min per IP)

### Monitoring & Observability
- Prometheus metrics (`/metrics`)
  - HTTP request metrics
  - Response time histograms
  - Active connections
  - Request/response sizes
- Grafana dashboards (pre-configured)
- Structured logging with slog
- Worker pool statistics

### Security
- SQL injection prevention (parameterized queries)
- Input validation at domain layer
- CORS configuration
- Non-root Docker user
- Environment-based secrets

## Project Structure

```
.
├── cmd/server/              # Application entry point
├── internal/
│   ├── config/              # Configuration management
│   ├── domain/              # Domain models & business rules
│   ├── handler/             # HTTP handlers & routing
│   ├── middleware/          # HTTP middleware
│   ├── repository/          # Data access layer
│   │   └── mysql/           # MySQL implementations
│   ├── service/             # Business logic
│   └── worker/              # Worker pool
├── pkg/
│   ├── errors/              # Custom error types
│   └── logger/              # Logging utilities
├── migrations/              # Database migrations
├── grafana/                 # Grafana provisioning
├── Dockerfile               # Multi-stage Docker build
├── docker-compose.yml       # Full stack deployment
├── prometheus.yml           # Prometheus configuration
└── API.md                   # Complete API documentation
```

## Development

### Available Commands

```bash
make build       # Build the application
make run         # Run the application
make test        # Run tests with coverage
make clean       # Clean build artifacts
make deps        # Install dependencies
make fmt         # Format code
make docker-up   # Start Docker services
make docker-down # Stop Docker services
```

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Building

```bash
# Build binary
make build

# Build Docker image
docker build -t go-backend:latest .
```

## Configuration

All configuration via environment variables (see `.env.example`):

### Server
- `SERVER_HOST` - Server host (default: localhost)
- `SERVER_PORT` - Server port (default: 8080)
- `SERVER_READ_TIMEOUT` - Read timeout (default: 15s)
- `SERVER_WRITE_TIMEOUT` - Write timeout (default: 15s)

### Database
- `DB_HOST` - MySQL host (default: localhost)
- `DB_PORT` - MySQL port (default: 3306)
- `DB_USER` - Database user
- `DB_PASSWORD` - Database password
- `DB_NAME` - Database name
- `DB_MAX_OPEN_CONNS` - Max open connections (default: 25)
- `DB_MAX_IDLE_CONNS` - Max idle connections (default: 5)

### Security
- `JWT_SECRET` - JWT signing secret (required)
- `PASSWORD_SALT_ROUNDS` - Bcrypt cost (default: 10)

### Application
- `APP_ENV` - Environment (development/production)
- `LOG_LEVEL` - Log level (debug/info/warn/error)

### Worker Pool
- `WORKER_POOL_SIZE` - Number of workers (default: 10)
- `TRANSACTION_QUEUE_SIZE` - Queue size (default: 100)

## Database Schema

### Users
- Auto-incrementing ID
- Unique username and email
- Bcrypt password hash
- Role (user/admin)
- Timestamps

### Transactions
- From/To user IDs (nullable)
- Amount (decimal 20,2)
- Type (credit/debit/transfer)
- Status (pending/processing/completed/failed/rolled_back)
- Timestamps

### Balances
- User ID (FK)
- Amount (decimal 20,2, non-negative)
- Last updated timestamp
- Auto-created on user registration (trigger)

### Audit Logs
- Entity type and ID
- Action performed
- JSON details
- Timestamp

## Monitoring

### Prometheus Metrics

Access at http://localhost:9090

Available metrics:
- `http_requests_total` - Total HTTP requests
- `http_request_duration_seconds` - Request duration histogram
- `http_request_size_bytes` - Request size histogram
- `http_response_size_bytes` - Response size histogram
- `http_active_connections` - Active HTTP connections

### Grafana Dashboards

Access at http://localhost:3000 (admin/admin)

Pre-configured datasource: Prometheus

Create custom dashboards or import community dashboards for:
- Request rate and latency
- Error rates
- Response times (p50, p95, p99)
- Throughput
- Database connections

## Deployment

### Docker

```bash
# Build image
docker build -t go-backend:latest .

# Run container
docker run -d \
  -p 8080:8080 \
  -e DB_HOST=mysql \
  -e JWT_SECRET=your-secret \
  --name go-backend \
  go-backend:latest
```

### Docker Compose

```bash
# Start all services
docker-compose up -d

# Scale application
docker-compose up -d --scale app=3

# View logs
docker-compose logs -f

# Stop all
docker-compose down
```

### Production Checklist

- [ ] Change `JWT_SECRET` to a strong random value
- [ ] Use environment-specific `.env` files
- [ ] Set `APP_ENV=production`
- [ ] Configure proper `DB_USER` and `DB_PASSWORD`
- [ ] Set up SSL/TLS certificates
- [ ] Configure firewall rules
- [ ] Set up log aggregation
- [ ] Configure backup strategy
- [ ] Set up monitoring alerts
- [ ] Review and adjust rate limits
- [ ] Configure CORS for production origins
- [ ] Set up health check monitoring

## Performance Characteristics

- **Throughput**: Limited by worker pool (10 concurrent transactions)
- **Latency**: p50 < 50ms, p95 < 200ms, p99 < 500ms
- **Connections**: 25 max open DB connections
- **Rate Limit**: 100 requests/minute per IP
- **Queue**: 100 pending transactions

## Security

- Password hashing: bcrypt (cost 10)
- Token signing: HMAC-SHA256
- SQL injection: Prevented via parameterized queries
- XSS: JSON responses auto-escaped
- CORS: Configurable allowed origins
- Rate limiting: Per-IP request limiting
- Docker: Non-root user execution

## License

MIT

## Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## Support

For issues and questions:
- GitHub Issues: [Report an issue](https://github.com/atakannerturk/go-backend-path/issues)
- Documentation: See [IMPLEMENTATION.md](IMPLEMENTATION.md) and [API.md](API.md)
