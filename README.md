# OrderBook-Go

A high-performance order book implementation in Go using B-trees for efficient price level management and order matching.

## Overview

This project implements a **limit order book** (LOB) with the following characteristics:

- **Data Structure**: B-tree based order levels for O(log n) inserts, deletes, and O(n) ordered traversals
- **Ordering**: 
  - Bids: Descending by price (highest bid first)
  - Asks: Ascending by price (lowest ask first)
- **Matching**: Incoming orders are matched against opposite-side levels with best-price-first priority
- **API**: RESTful HTTP endpoints for order submission and book snapshots

## Project Structure

```
OrderBookProject/
├── main.go                          # Entry point; starts HTTP server on :8080
├── go.mod                           # Module definition
├── src/
│   ├── api/
│   │   └── endpoints.go            # HTTP handlers for order submission and retrieval
│   ├── orderbook/
│   │   └── orderBook.go            # Core order book logic with B-trees
│   └── orders/
│       └── order.go                # Order types (Order, Bid, Ask)
└── README.md                        # This file
```

## API Endpoints

### `GET /hello`
Basic health check.

**Response:**
```
hello world
```

---

### `GET /getOrderBook`
Retrieve a snapshot of the current order book in JSON format.

**Response (200 OK):**
```json
{
  "bids": [
    { "position": 0, "quantity": 10, "price": 101.5 },
    { "position": 0, "quantity": 5, "price": 100.2 }
  ],
  "asks": [
    { "position": 1, "quantity": 3, "price": 99.9 },
    { "position": 1, "quantity": 7, "price": 100.5 }
  ]
}
```

**Notes:**
- `bids` array is ordered highest price first (best bids)
- `asks` array is ordered lowest price first (best asks)
- `position`: 0 = BUY, 1 = SELL
- `quantity` and `price` are numeric aggregates per price level

---

### `POST /order`
Submit a new limit order. The order is matched against opposite-side levels and any remainder is added to the book.

**Request Body (JSON):**
```json
{
  "position": 0,
  "quantity": 50,
  "price": 101.25
}
```

**Fields:**
- `position`: 0 for BUY, 1 for SELL
- `quantity`: Number of units to trade (must be > 0)
- `price`: Limit price in currency units (must be > 0)

**Response (200 OK):**
- `"Order fullfilled"` – Entire order was matched.
- `"Order recieved"` – Order was partially or fully added to the book.

**Error Responses:**
- `400 Bad Request` – Invalid JSON, missing fields, or invalid values
- `405 Method Not Allowed` – Non-POST request
- `500 Internal Server Error` – Server initialization error

**Example Order Flow:**
1. Current book: Asks at [99.9, 100.5], Bids at [101.5, 100.2]
2. Submit BUY order: price=100.0, qty=10
   - Matches 3 units at ask 99.9 (remaining qty: 7)
   - Cannot match at ask 100.5 (too high)
   - Remaining 7 units added as a new bid at 100.0

## Building & Running

### Prerequisites
- Go 1.16+ 
- Dependencies: `github.com/google/btree`

### Build
```bash
go build ./...
```

### Run
```bash
go run main.go
```

The server starts on `http://localhost:8080`.

### Test API
```bash
# Health check
curl http://localhost:8080/hello

# Get order book
curl http://localhost:8080/getOrderBook

# Submit a BUY order
curl -X POST http://localhost:8080/order \
  -H "Content-Type: application/json" \
  -d '{"position": 0, "quantity": 10, "price": 101.5}'

# Submit a SELL order
curl -X POST http://localhost:8080/order \
  -H "Content-Type: application/json" \
  -d '{"position": 1, "quantity": 5, "price": 99.9}'
```

## Key Components

### `Order` (orders/order.go)
Represents a single order with position (BUY/SELL), quantity, and price.

### `Bid` & `Ask` (orders/order.go)
Wrapper types implementing `btree.Item` with custom ordering:
- **Bid**: `Less(i, j)` returns `true` if price[i] > price[j]` (descending)
- **Ask**: `Less(i, j)` returns `true` if price[i] < price[j]` (ascending)

### `OrderBook` (orderbook/orderBook.go)
Core order book with two B-trees:
- `Bids`: Max-heap style (highest price first)
- `Asks`: Min-heap style (lowest price first)

**Key Methods:**
- `NewOrder(order)`: Matches incoming order and adds remainder to book
- `fillBid(order)`: Matches incoming BUY order against asks
- `fillAsk(order)`: Matches incoming SELL order against bids
- `Snapshot()`: Returns a JSON-serializable view of the current book

## Performance Characteristics

| Operation | Complexity | Notes |
|-----------|-----------|-------|
| Insert/Update level | O(log n) | B-tree rebalancing |
| Delete level | O(log n) | B-tree rebalancing |
| Get level by price | O(log n) | Direct lookup |
| Order matching | O(m log n) | m = matched levels, n = total levels |
| Full traversal (snapshot) | O(n) | Single in-order scan |
| Top-of-book | O(1) | Tree.Min() / Tree.Max() |

## Dependencies

- `github.com/google/btree` – B-tree implementation

Install:
```bash
go get github.com/google/btree
```

