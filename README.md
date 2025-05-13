
# Go Broker Test

Этот проект — пример брокерского сервиса на Go для управления сделками и статистикой аккаунтов. Включает серверную часть и воркер, использует SQLite для хранения данных.

## Возможности
- **Управление сделками**: добавление и обработка сделок через REST API
- **Статистика аккаунтов**: получение статистики по аккаунтам
- **Проверка состояния**: endpoint для healthcheck

## Требования
- Go 1.18 или новее
- SQLite

## Быстрый старт

1. Клонируйте репозиторий:
   ```bash
   git clone <repository-url>
   cd go-broker-test
   ```

2. Установите зависимости:
   ```bash
   go mod tidy
   ```

3. Запустите сервер:
   ```bash
   go run cmd/server/main.go
   ```

4. Запустите воркер:
   ```bash
   go run cmd/worker/main.go
   ```

## API эндпоинты

### 1. Добавить сделку
**POST** `/trades`

Тело запроса:
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

Ответы:
- `204 No Content` — успех
- `400 Bad Request` — невалидный ввод
- `500 Internal Server Error` — ошибка базы данных

### 2. Получить статистику аккаунта
**GET** `/stats/{account}`

Ответ:
```json
{
  "account": "ACC1",
  "trades": 5,
  "profit": 1000.50
}
```

Ошибки:
- `400 Bad Request` — не указан аккаунт
- `500 Internal Server Error` — ошибка базы данных

### 3. Проверка состояния
**GET** `/healthz`

Ответы:
- `200 OK` — сервер работает
- `500 Internal Server Error` — проблемы с базой данных

## Тестирование

Запуск всех тестов:
```bash
go test ./...
```

Покрытие:
```bash
go test ./... -cover
```

## Примеры cURL-запросов

### 1. Добавить сделку (валидный)
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
# Ожидается: 204 No Content
```

### 2. Добавить сделку (невалидный JSON)
```bash
curl -X POST http://localhost:8080/trades \
  -H "Content-Type: application/json" \
  -d '{invalid json}'
# Ожидается: 400 Bad Request, тело: Invalid JSON payload
```

### 3. Добавить сделку (пустое поле account)
```bash
curl -X POST http://localhost:8080/trades \
  -H "Content-Type: application/json" \
  -d '{
    "account": "",
    "symbol": "EURUSD",
    "volume": 1.5,
    "open": 1.2345,
    "close": 1.2350,
    "side": "buy"
  }'
# Ожидается: 400 Bad Request, тело: account must not be empty
```

### 4. Добавить сделку (неверный HTTP-метод)
```bash
curl -X GET http://localhost:8080/trades
# Ожидается: 405 Method Not Allowed
```

### 5. Получить статистику (существующий аккаунт)
```bash
curl -X GET http://localhost:8080/stats/ACC1
# Ожидается: 200 OK, JSON с данными аккаунта
```

### 6. Получить статистику (несуществующий аккаунт)
```bash
curl -X GET http://localhost:8080/stats/UNKNOWN
# Ожидается: 200 OK, JSON с нулевой статистикой
```

### 7. Получить статистику (без аккаунта)
```bash
curl -X GET http://localhost:8080/stats/
# Ожидается: 400 Bad Request, тело: Account is required
```

### 8. Получить статистику (неверный HTTP-метод)
```bash
curl -X POST http://localhost:8080/stats/ACC1
# Ожидается: 405 Method Not Allowed
```

### 9. Healthcheck (валидный)
```bash
curl -X GET http://localhost:8080/healthz
# Ожидается: 200 OK, тело: OK
```

### 10. Healthcheck (неверный HTTP-метод)
```bash
curl -X POST http://localhost:8080/healthz
# Ожидается: 405 Method Not Allowed
```
