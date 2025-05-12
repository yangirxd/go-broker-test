# Go Broker Test

This project is a Go-based application for managing trades and account statistics. It includes a server and worker component, along with a SQLite database for data storage.

## Features
- **Trades Management**: Add and process trades.
- **Account Statistics**: Fetch account statistics.
- **Health Check**: Verify server health.

## Prerequisites
- Go 1.18 or later
- SQLite

## Setup

1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd go-broker-test
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Run the server:
   ```bash
   go run cmd/server/main.go
   ```

4. Run the worker:
   ```bash
   go run cmd/worker/main.go
   ```

## API Endpoints

### 1. Add Trade
**POST** `/trades`

**Request Body:**
```json
{
  "account": "ACC1",
  "symbol": "EURUSD",
  "volume": 1.5,
  "open": 1.2345,
  "close": 1.2350,
  "side": "buy"
}
```

**Response:**
- `204 No Content` on success
- `400 Bad Request` for invalid input
- `500 Internal Server Error` for database errors

### 2. Fetch Account Statistics
**GET** `/stats/{account}`

**Response:**
```json
{
  "account": "ACC1",
  "trades": 5,
  "profit": 1000.50
}
```

**Error Response:**
- `400 Bad Request` if account is missing
- `500 Internal Server Error` for database errors

### 3. Health Check
**GET** `/healthz`

**Response:**
- `200 OK` if the server is healthy
- `500 Internal Server Error` if the database connection fails

## Testing

Run all tests:
```bash
make test
```

Generate a coverage report:
```bash
make coverage
```

## Example cURL Requests

### Add a Trade
```bash
curl -X POST http://localhost:8080/trades \
-H "Content-Type: application/json" \
-d '{
  "account": "ACC1",
  "symbol": "EURUSD",
  "volume": 1.5,
  "open": 1.2345,
  "close": 1.2350,
  "side": "buy"
}'
```

### Fetch Account Statistics
```bash
curl -X GET http://localhost:8080/stats/ACC1
```

### Health Check
```bash
curl -X GET http://localhost:8080/healthz
```
