# API Documentation

Base URL: `http://localhost:8080/api/v1`

## Authentication

All endpoints except `/auth/register` and `/auth/login` require a JWT token in the Authorization header:

```
Authorization: Bearer <token>
```

## Endpoints

### Health Check

#### GET /health
Get server health status.

**Response:**
```json
{
  "data": {
    "status": "ok",
    "version": "1.0.0"
  }
}
```

### Metrics

#### GET /metrics
Prometheus metrics endpoint (for monitoring).

---

## Authentication Endpoints

### POST /api/v1/auth/register
Register a new user.

**Request Body:**
```json
{
  "username": "johndoe",
  "email": "john@example.com",
  "password": "password123",
  "role": "user"
}
```

**Response (201):**
```json
{
  "data": {
    "user": {
      "id": 1,
      "username": "johndoe",
      "email": "john@example.com",
      "role": "user",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    },
    "token": "eyJhbGciOiJIUzI1NiIs..."
  }
}
```

### POST /api/v1/auth/login
Authenticate a user.

**Request Body:**
```json
{
  "email": "john@example.com",
  "password": "password123"
}
```

**Response (200):**
```json
{
  "data": {
    "user": {
      "id": 1,
      "username": "johndoe",
      "email": "john@example.com",
      "role": "user",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    },
    "token": "eyJhbGciOiJIUzI1NiIs..."
  }
}
```

### GET /api/v1/auth/me
Get current user information (requires authentication).

**Response (200):**
```json
{
  "data": {
    "id": 1,
    "username": "johndoe",
    "email": "john@example.com",
    "role": "user",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

---

## User Management Endpoints

### GET /api/v1/users
List all users (requires authentication).

**Query Parameters:**
- `limit` (optional): Number of users to return (default: 20, max: 100)
- `offset` (optional): Number of users to skip (default: 0)

**Response (200):**
```json
{
  "data": {
    "users": [
      {
        "id": 1,
        "username": "johndoe",
        "email": "john@example.com",
        "role": "user",
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z"
      }
    ],
    "limit": 20,
    "offset": 0
  }
}
```

### GET /api/v1/users/{id}
Get user by ID (requires authentication).

**Response (200):**
```json
{
  "data": {
    "id": 1,
    "username": "johndoe",
    "email": "john@example.com",
    "role": "user",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

### DELETE /api/v1/users/{id}
Delete user by ID (requires admin role).

**Response (200):**
```json
{
  "data": {
    "message": "User deleted successfully"
  }
}
```

---

## Transaction Endpoints

### POST /api/v1/transactions/credit
Credit funds to a user account (requires authentication).

**Request Body:**
```json
{
  "user_id": 1,
  "amount": 100.50
}
```

**Response (201):**
```json
{
  "data": {
    "id": 1,
    "to_user_id": 1,
    "amount": 100.50,
    "type": "credit",
    "status": "pending",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

### POST /api/v1/transactions/debit
Debit funds from a user account (requires authentication).

**Request Body:**
```json
{
  "user_id": 1,
  "amount": 50.25
}
```

**Response (201):**
```json
{
  "data": {
    "id": 2,
    "from_user_id": 1,
    "amount": 50.25,
    "type": "debit",
    "status": "pending",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

### POST /api/v1/transactions/transfer
Transfer funds between accounts (requires authentication).

**Request Body:**
```json
{
  "to_user_id": 2,
  "amount": 25.00
}
```

Note: `from_user_id` is automatically set to the authenticated user.

**Response (201):**
```json
{
  "data": {
    "id": 3,
    "from_user_id": 1,
    "to_user_id": 2,
    "amount": 25.00,
    "type": "transfer",
    "status": "pending",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

### GET /api/v1/transactions/history
Get current user's transaction history (requires authentication).

**Query Parameters:**
- `limit` (optional): Number of transactions to return (default: 20, max: 100)
- `offset` (optional): Number of transactions to skip (default: 0)

**Response (200):**
```json
{
  "data": {
    "transactions": [
      {
        "id": 1,
        "from_user_id": 1,
        "to_user_id": 2,
        "amount": 25.00,
        "type": "transfer",
        "status": "completed",
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z"
      }
    ],
    "limit": 20,
    "offset": 0
  }
}
```

### GET /api/v1/transactions/{id}
Get transaction by ID (requires authentication).

**Response (200):**
```json
{
  "data": {
    "id": 1,
    "from_user_id": 1,
    "to_user_id": 2,
    "amount": 25.00,
    "type": "transfer",
    "status": "completed",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

### GET /api/v1/transactions
List all transactions (requires admin role).

**Query Parameters:**
- `limit` (optional): Number of transactions to return (default: 20, max: 100)
- `offset` (optional): Number of transactions to skip (default: 0)

**Response (200):**
```json
{
  "data": {
    "transactions": [...],
    "limit": 20,
    "offset": 0
  }
}
```

### POST /api/v1/transactions/{id}/rollback
Rollback a completed transaction (requires admin role).

**Response (200):**
```json
{
  "data": {
    "message": "Transaction rolled back successfully"
  }
}
```

---

## Balance Endpoints

### GET /api/v1/balances/current
Get current user's balance (requires authentication).

**Response (200):**
```json
{
  "data": {
    "user_id": 1,
    "amount": 125.25,
    "last_updated_at": "2024-01-01T00:00:00Z"
  }
}
```

---

## Error Responses

All error responses follow this format:

```json
{
  "error": "Bad Request",
  "message": "Detailed error message",
  "code": "ERROR_CODE"
}
```

### Common Error Codes

- `400 Bad Request`: Invalid request body or parameters
- `401 Unauthorized`: Missing or invalid authentication token
- `403 Forbidden`: Insufficient permissions
- `404 Not Found`: Resource not found
- `500 Internal Server Error`: Server error

### Example Error Response

```json
{
  "error": "Unauthorized",
  "message": "Invalid or expired token",
  "code": "INVALID_TOKEN"
}
```

---

## Transaction States

Transactions go through the following states:

1. `pending` - Transaction created, waiting to be processed
2. `processing` - Worker is processing the transaction
3. `completed` - Transaction successfully completed
4. `failed` - Transaction failed (e.g., insufficient funds)
5. `rolled_back` - Transaction was rolled back by admin

---

## Rate Limiting

The API implements rate limiting of 100 requests per minute per IP address.

If you exceed this limit, you'll receive a `429 Too Many Requests` response.

---

## Example Usage

### Register and Login

```bash
# Register
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "johndoe",
    "email": "john@example.com",
    "password": "password123",
    "role": "user"
  }'

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "password123"
  }'
```

### Make a Transfer

```bash
# Transfer (using token from login)
curl -X POST http://localhost:8080/api/v1/transactions/transfer \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -d '{
    "to_user_id": 2,
    "amount": 50.00
  }'
```

### Check Balance

```bash
# Get current balance
curl -X GET http://localhost:8080/api/v1/balances/current \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```
